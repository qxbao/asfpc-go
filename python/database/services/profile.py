"""Service for managing groups."""
import logging
from typing import List
from sqlalchemy import and_, select
from sqlalchemy.orm import selectinload
from database.database import Database
from database.models.profile import UserProfile

class ProfileService:
  def __init__(self):
    self.__session = Database.get_session()
    self.logger = logging.getLogger("ProfileService")

  async def get_training_profiles(self) -> List[UserProfile]:
    try:
      async with self.__session as conn:
        query = select(UserProfile)\
          .where(and_(UserProfile.is_analyzed,
                      UserProfile.emb_profile))\
          .options(selectinload(UserProfile.emb_profile))
        res = await conn.execute(query)
        return res.scalars().all()
    except Exception as e:
      self.logger.exception(e)
      return []