import logging

from sqlalchemy import select

from database.database import Database
from database.models.prompt import Prompt


class PromptService:
  def __init__(self):
    self.logger = logging.getLogger("PromptService")

  async def get_prompt(self, name: str) -> str | None:
    try:
      async with Database.get_session() as conn:
        query = select(Prompt).where(Prompt.service_name == name).order_by(Prompt.version.desc()).limit(1)
        res = await conn.execute(query)
        prompt = res.scalar_one_or_none()
        return prompt.content if prompt else None
    except Exception:
      self.logger.exception("Exception occurred in get_prompt")
      return None

  def inject_prompt(self, template: str, *args) -> str:
    for i, arg in enumerate(args):
      if arg is None or arg == "":
        template = template.replace(f"INSERT_{i+1}", "(null)", 1)
      else:
        template = template.replace(f"INSERT_{i+1}", arg, 1)
    return template

  async def get_prompt_by_key_and_category(self, key: str, category_id: int) -> str | None:
    try:
      async with Database.get_session() as conn:
        query = select(Prompt).where(
          Prompt.service_name == key,
          Prompt.category_id == category_id
        ).order_by(Prompt.version.desc()).limit(1)
        res = await conn.execute(query)
        prompt = res.scalar_one_or_none()
        return prompt.content if prompt else None
    except Exception:
      self.logger.exception("Exception occurred in get_prompt_by_key_and_category")
      return None
