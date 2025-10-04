"""Service for managing profiles."""

import logging

from sqlalchemy import and_, select
from sqlalchemy.orm import selectinload

from database.database import Database
from database.models.emb_profile import EmbeddedProfile
from database.models.profile import UserProfile


class ProfileService:
  def __init__(self):
    self.logger = logging.getLogger("ProfileService")

  async def get_training_profiles(self) -> list[UserProfile]:
    try:
      async with Database.get_session() as conn:
        query = (
          select(UserProfile)
          .where(and_(UserProfile.is_analyzed, UserProfile.emb_profile.has()))
          .options(selectinload(UserProfile.emb_profile))
        )
        res = await conn.execute(query)
        return list(res.scalars().all())
    except Exception:
      self.logger.exception("Exception occurred in get_training_profiles")
      return []

  async def get_profile_by_id(self, profile_id: int, with_embed: bool = False):
    try:
      async with Database.get_session() as conn:
        query = select(UserProfile).where(UserProfile.id == profile_id)
        if with_embed:
          query = query.options(selectinload(UserProfile.emb_profile))
        res = await conn.execute(query)
        return res.scalar_one_or_none()
    except Exception:
      self.logger.exception("Exception occurred in get_profile_by_id")
      return None

  async def insert_profile_embedding(self, profile_id: int, embedding: list[float]) -> bool:
    try:
      async with Database.get_session() as conn:
        embedded_profile = EmbeddedProfile(
          pid=profile_id,
          embedding=embedding,
        )
        await conn.merge(embedded_profile)
        await conn.commit()
        return True
    except Exception:
      self.logger.exception("Exception occurred in insert_profile_embedding")
      return False
