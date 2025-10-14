"""Config service for managing configuration values."""

import logging

from sqlalchemy import select

from database.database import Database
from database.models.config import Config


class ConfigService:
  """Service for managing configuration values."""

  def __init__(self):
    self.logger = logging.getLogger(__name__)

  async def get_config(self, key: str) -> str | None:
    """Get a configuration value by key."""
    async with Database.get_session() as session:
      result = await session.execute(
        select(Config).where(Config.key == key)
      )
      config = result.scalar_one_or_none()
      return config.value if config else None

  async def get_ml_model_path(self, category_id: int) -> str | None:
    """Get the ML model path for a specific category."""
    return await self.get_config(f"ml_model_path_category_{category_id}")

  async def get_embedding_model_path(self, category_id: int) -> str | None:
    """Get the embedding model path for a specific category."""
    return await self.get_config(f"embedding_model_category_{category_id}")
