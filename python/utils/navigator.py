import asyncio
import json
import logging
import sys
from typing import Any

import pandas as pd
from database.services.account import AccountService
from browser.account import AccountAutomationService
from browser.group import GroupAutomationService
from database.services.group import GroupService
from database.services.profile import ProfileService
from ml.model import PotentialCustomerScoringModel
from database.services.request import RequestService
from utils.dialog import DialogUtil
import re


class TaskNavigator:
  def __init__(self, config: dict[str, Any]):
    self.config = config
    self.logger = logging.getLogger("TaskNavigator")
    
  async def login(self) -> None:
    user_id = self.config.get("uid", None)
    if not user_id:
      raise ValueError("--uid is required for login task")
    account_service = AccountService()
    account = await account_service.get_account_by_id(int(user_id))
    if not account:
      raise ValueError(f"Account with id {user_id} not found")
    account_automation_service = AccountAutomationService()
    is_ok = await account_automation_service.login_account(account)
    if not is_ok:
      raise ValueError(f"Failed to login account with id {user_id}")
    
    is_blocked = await DialogUtil.confirmation("Account Status", "Is the account blocked?")
    account.is_block = is_blocked
    await account_service.update_account(account)
  
  async def join_group(self) -> None:
    group_id = self.config.get("group_id", None)
    if not group_id:
      raise ValueError("--group_id is required for join_group task")
    gs = GroupService()
    group = await gs.get_group_by_id(int(group_id), include_account=True)
    if not group:
      raise ValueError(f"Group with id {group_id} not found")
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
    if request_id is not None:
      try:
        request_id = int(request_id)
      except ValueError:
        raise ValueError("--request-id must be an integer")
    else:
      raise ValueError("--request-id is required for train-model task")
    
    self.logger.info(f"Training model: {model_name}")
    ps = ProfileService()
    profiles = await ps.get_training_profiles()
    if not profiles:
      raise ValueError("No profiles available for training")
    self.logger.info(f"Found {len(profiles)} profiles for training")
    input_df = pd.DataFrame([p.to_df() for p in profiles])
    rs = RequestService()
    await rs.update_request(request_id, status=1, description="Preparing data for training...", progress=0.0)
    model = PotentialCustomerScoringModel(
      request_id=request_id
    )
    auto_tune = self.config.get("auto-tune")
    if not auto_tune or auto_tune == "False":
      auto_tune = False
    else:
      auto_tune = True
    model.load_data(input_df)
    await rs.update_request(request_id, status=1, progress=0.1, description="Training in progress...")
    model.train(auto_tune=auto_tune)
    await rs.update_request(request_id, status=1, progress=0.95, description="Finalizing training...")
    self.logger.info("Model trained successfully")
    test_results = model.test()
    self.logger.info(f"Test result: {test_results}")
    await rs.update_request(request_id, status=1, progress=0.99, description="Saving model...")
    model.save_model(model_name)
    self.logger.info(f"Model saved as: {model_name}")
    
  async def predict(self) -> None:
    model_name = self.config.get("model-name", None)
    if not model_name:
      raise Exception("Argument --model-name is required")
    targets = self.config.get("targets", None)
    if not targets:
      raise Exception("Argument --targets is required")
    match = re.match(r'^([0-9]+[,]?)*[0-9]+$', targets)
    if not match:
      raise Exception("Argument --targets format is invalid. Example: 1,2,3")
    id_set = set(int(x) for x in targets.split(","))
    id_list = list(id_set)
    model = PotentialCustomerScoringModel()
    model.load_model(model_name)
    ps = ProfileService()
    sem = asyncio.Semaphore(10)
    async def get_score(profile_id: int):
      profile = await ps.get_profile_by_id(profile_id, True)
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
    
    tasks = [
      semaphore_proc(sem, i) for i in range(len(id_list))
    ]
    
    result = await asyncio.gather(*tasks)
    res_obj = {}
    for i in range(len(id_list)):
      res_obj[str(id_list[i])] = result[i]
    print(json.dumps(res_obj))