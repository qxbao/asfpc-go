"""Service for managing groups."""
import logging
from typing import List
from sqlalchemy import and_, select
from sqlalchemy.orm import selectinload
from database.database import Database
from database.models.profile import UserProfile

class ProfileService:
  def __init__(self):
    self.logger = logging.getLogger("ProfileService")

  async def get_training_profiles(self) -> List[UserProfile]:
    try:
      async with Database.get_session() as conn:
        query = select(UserProfile)\
          .where(and_(UserProfile.is_analyzed,
                      UserProfile.emb_profile))\
          .options(selectinload(UserProfile.emb_profile))
        res = await conn.execute(query)
        return res.scalars().all()
    except Exception as e:
      self.logger.exception(e)
      return []
    
  async def get_profile_by_id(self, id: int, with_embed: bool = False):
    try:
      async with Database.get_session() as conn:
        query = select(UserProfile).where(UserProfile.id == id)
        if with_embed:
          query = query.options(selectinload(UserProfile.emb_profile))
        res = await conn.execute(query)
        return res.scalar_one_or_none()
    except Exception as e:
      self.logger.exception(e)
      return None