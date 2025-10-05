import asyncio
import json
import logging
import re
import sys
from typing import Any

import pandas as pd

from browser.account import AccountAutomationService
from browser.group import GroupAutomationService
from database.services.account import AccountService
from database.services.group import GroupService
from database.services.profile import ProfileService
from database.services.prompt import PromptService
from database.services.request import RequestService
from ml import BGEM3EmbedModel, PotentialCustomerScoringModel
from utils.dialog import DialogUtil


class TaskNavigator:
  def __init__(self, config: dict[str, Any]):
    self.config = config
    self.logger = logging.getLogger("TaskNavigator")

  async def login(self) -> None:
    user_id = self.config.get("uid", None)
    if not user_id:
      err_msg = "--uid is required for login task"
      raise ValueError(err_msg)
    account_service = AccountService()
    account = await account_service.get_account_by_id(int(user_id))
    if not account:
      err_msg = f"Account with id {user_id} not found"
      raise ValueError(err_msg)
    account_automation_service = AccountAutomationService()
    is_ok = await account_automation_service.login_account(account)
    if not is_ok:
      err_msg = f"Failed to login account with id {user_id}"
      raise ValueError(err_msg)

    is_blocked = await DialogUtil.confirmation(
      "Account Status", "Is the account blocked?"
    )
    account.is_block = is_blocked
    await account_service.update_account(account)

  async def join_group(self) -> None:
    group_id = self.config.get("group_id", None)
    if not group_id:
      err_msg = "--group_id is required for join_group task"
      raise ValueError(err_msg)
    gs = GroupService()
    group = await gs.get_group_by_id(int(group_id), include_account=True)
    if not group:
      err_msg = f"Group with id {group_id} not found"
      raise ValueError(err_msg)
    gas = GroupAutomationService()
    is_ok = await gas.join_group(group)
    group.is_joined = is_ok
    await gs.update_group(group)
    if not is_ok:
      sys.exit(1)
    else:
      sys.exit(0)

  async def train_model(self) -> None:
    model_name = self.config.get("model-name", "ModelX")
    request_id = self.config.get("request-id", None)
    trials = self.config.get("trials", None)
    if request_id is not None:
      try:
        request_id = int(request_id)
      except ValueError as err:
        err_msg = "--request-id must be an integer"
        raise ValueError(err_msg) from err
    else:
      err_msg = "--request-id is required for train-model task"
      raise ValueError(err_msg)

    self.logger.info("Training model: %s", model_name)
    ps = ProfileService()
    profiles = await ps.get_training_profiles()
    if not profiles:
      err_msg = "No profiles available for training"
      raise ValueError(err_msg)
    self.logger.info("Found %d profiles for training", len(profiles))
    input_df = pd.DataFrame([p.to_df() for p in profiles])
    rs = RequestService()
    await rs.update_request(
      request_id, status=1, description="Preparing data for training...", progress=0.0
    )
    model = PotentialCustomerScoringModel(
      request_id=request_id,
      model_name=model_name
    )
    self.logger.info("Loading trial: %s", str(trials))
    if trials is not None:
      model.trials = int(trials)
    auto_tune = self.config.get("auto-tune")
    auto_tune = not (not auto_tune or auto_tune == "False")
    model.load_data(input_df)
    await rs.update_request(
      request_id, status=1, progress=0.1, description="Training in progress..."
    )
    model.train(auto_tune=auto_tune)
    await rs.update_request(
      request_id, status=1, progress=0.95, description="Finalizing training..."
    )
    self.logger.info("Model trained successfully")
    test_results = model.test()
    self.logger.info("Test result: %s", test_results)
    await rs.update_request(
      request_id, status=1, progress=0.99, description="Saving model..."
    )
    model.save_model()
    self.logger.info("Model saved as: %s", model_name)

  async def predict(self) -> None:
    model_name = self.config.get("model-name", None)
    if not model_name:
      err_msg = "Argument --model-name is required"
      raise Exception(err_msg)
    targets = self.config.get("targets", None)
    if not targets:
      err_msg = "Argument --targets is required"
      raise Exception(err_msg)
    match = re.match(r"^([0-9]+[,]?)*[0-9]+$", targets)
    if not match:
      err_msg = "Argument --targets format is invalid. Example: 1,2,3"
      raise Exception(err_msg)

    id_set = {int(x) for x in targets.split(",")}
    id_list = list(id_set)
    model = PotentialCustomerScoringModel(
      model_name=model_name,
    )
    model.load_model(model_name)
    ps = ProfileService()
    sem = asyncio.Semaphore(10)

    async def get_score(profile_id: int):
      profile = await ps.get_profile_by_id(profile_id, with_embed=True)
      if not profile:
        return None
      input_df = pd.DataFrame(profile.to_df())
      score = model.predict(input_df)
      if isinstance(score, list):
        score = score[0]
      return score

    async def semaphore_proc(sem, i):
      async with sem:
        return await get_score(id_list[i])

    tasks = [semaphore_proc(sem, i) for i in range(len(id_list))]

    result = await asyncio.gather(*tasks)
    res_obj = {}
    for i in range(len(id_list)):
      res_obj[str(id_list[i])] = result[i]
    print(json.dumps(res_obj))  # noqa: T201

  async def embed_profiles(self) -> None:
    targets = self.config.get("targets", None)
    if not targets:
      err_msg = "Argument --targets is required"
      raise Exception(err_msg)
    match = re.match(r"^([0-9]+[,]?)*[0-9]+$", targets)
    if not match:
      err_msg = "Argument --targets format is invalid. Example: 1,2,3"
      raise Exception(err_msg)

    id_set = {int(x) for x in targets.split(",")}
    id_list = list(id_set)
    model = BGEM3EmbedModel()
    profile_service = ProfileService()
    prompt_service = PromptService()
    template = await prompt_service.get_prompt("self-embedding")
    if not template:
      err_msg = "Prompt 'self-embedding' not found in database"
      raise Exception(err_msg)

    sem = asyncio.Semaphore(10)

    async def get_embedding(profile_id: int) -> list[float] | None:
      profile = await profile_service.get_profile_by_id(profile_id)
      if not profile:
        return None
      my_temp = template
      final_str = prompt_service.inject_prompt(
        my_temp,
        profile.location,
        profile.work,
        profile.bio,
        profile.education,
        profile.relationship_status,
        profile.hometown,
        profile.locale,
        profile.gender,
        profile.birthday,
      )
      embedding = model.embed([final_str])
      if isinstance(embedding, list) and len(embedding) > 0 and isinstance(embedding[0], list):
        return embedding[0]
      if isinstance(embedding, list) and all(isinstance(x, float) for x in embedding):
        return embedding # type: ignore[return-value]
      return None

    async def semaphore_proc(sem, i):
      async with sem:
        return await get_embedding(id_list[i])
    tasks = [semaphore_proc(sem, i) for i in range(len(id_list))]
    results = await asyncio.gather(*tasks)
    success_count = 0
    for pid, emb in zip(id_list, results, strict=True):
      if emb is None:
        continue
      if await profile_service.insert_profile_embedding(pid, emb):
        success_count += 1
    print(f"Embeddings generated for {success_count}/{len(id_list)} profiles")  # noqa: T201
