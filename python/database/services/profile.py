"""Service for managing profiles."""

import logging

from sqlalchemy import and_, select
from sqlalchemy.orm import contains_eager, selectinload

from database.database import Database
from database.models.emb_profile import EmbeddedProfile
from database.models.profile import UserProfile
from database.models.user_profile_category import user_profile_category_table


class ProfileService:
  def __init__(self):
    self.logger = logging.getLogger("ProfileService")

  async def get_training_profiles(self, category_id: int | None = None) -> list[tuple[UserProfile, float]]:
    """
    Get profiles for training with their gemini scores.

    Args:
        category_id: Optional category ID to filter profiles by

    Returns:
        List of tuples (UserProfile, gemini_score)

    """
    try:
      async with Database.get_session() as conn:
        if category_id is None:
          err_msg = "category_id is required for get_training_profiles"
          raise Exception(err_msg)  # noqa: TRY301
        query = (
          select(UserProfile, user_profile_category_table.c.gemini_score)
          .join(
            user_profile_category_table,
            and_(
              UserProfile.id == user_profile_category_table.c.user_profile_id,
              user_profile_category_table.c.category_id == category_id,
            ),
          )
          .join(
            EmbeddedProfile,
            and_(
              UserProfile.id == EmbeddedProfile.pid,
              EmbeddedProfile.cid == category_id,
            ),
          )
          .where(
            and_(
              UserProfile.is_analyzed,
              user_profile_category_table.c.gemini_score.isnot(None),
            )
          )
          .options(selectinload(UserProfile.emb_profiles))
        )

        res = await conn.execute(query)
        return [(row[0], float(row[1])) for row in res.all()]
    except Exception:
      self.logger.exception("Exception occurred in get_training_profiles")
      return []

  async def get_profile_by_id(self, profile_id: int, category_id: int | None = None, with_embed: bool = False):
    """
    Get profile by ID, optionally with embeddings filtered by category.

    Args:
        profile_id: The profile ID to fetch
        category_id: Optional category ID to filter embeddings
        with_embed: Whether to load embeddings for the specified category

    Returns:
        UserProfile or None if not found

    """
    try:
      async with Database.get_session() as conn:
        query = select(UserProfile).where(UserProfile.id == profile_id)

        if with_embed and category_id is not None:
          query = (
            query
            .outerjoin(
              EmbeddedProfile,
              and_(
                UserProfile.id == EmbeddedProfile.pid,
                EmbeddedProfile.cid == category_id,
              ),
            )
            .options(contains_eager(UserProfile.emb_profiles))
          )
        elif with_embed:
          query = query.options(selectinload(UserProfile.emb_profiles))
        res = await conn.execute(query)
        profile = res.unique().scalar_one_or_none() if with_embed and category_id is not None else res.scalar_one_or_none()
        if profile and with_embed and category_id is not None:
          profile.emb_profiles = [
            emb for emb in profile.emb_profiles if emb.cid == category_id
          ]
        return profile
    except Exception:
      self.logger.exception("Exception occurred in get_profile_by_id")
      return None

  async def get_profile_with_category_score(
    self, profile_id: int, category_id: int
  ) -> tuple[UserProfile, float | None] | None:
    """
    Get profile by ID with its category-specific model score and embedding.

    Args:
        profile_id: The profile ID to fetch
        category_id: The category ID to get score for

    Returns:
        Tuple of (UserProfile, model_score) or None if not found
        The UserProfile will have emb_profiles filtered to only include the specified category

    """
    try:
      async with Database.get_session() as conn:
        query = (
          select(UserProfile, user_profile_category_table.c.model_score)
          .outerjoin(
            user_profile_category_table,
            and_(
              UserProfile.id == user_profile_category_table.c.user_profile_id,
              user_profile_category_table.c.category_id == category_id,
            ),
          )
          .outerjoin(
            EmbeddedProfile,
            and_(
              UserProfile.id == EmbeddedProfile.pid,
              EmbeddedProfile.cid == category_id,
            ),
          )
          .where(UserProfile.id == profile_id)
          .options(contains_eager(UserProfile.emb_profiles))
        )
        res = await conn.execute(query)
        row = res.unique().one_or_none()
        if row is None:
          return None

        profile = row[0]
        model_score = float(row[1]) if row[1] is not None else None

        # Filter emb_profiles to only include the requested category
        profile.emb_profiles = [
          emb for emb in profile.emb_profiles if emb.cid == category_id
        ]

        return (profile, model_score)
    except Exception:
      self.logger.exception("Exception occurred in get_profile_with_category_score")
      return None

  async def insert_profile_embedding(self, profile_id: int, category_id: int, embedding: list[float]) -> bool:
    try:
      async with Database.get_session() as conn:
        embedded_profile = EmbeddedProfile(
          pid=profile_id,
          cid=category_id,
          embedding=embedding,
        )
        await conn.merge(embedded_profile)
        await conn.commit()
        return True
    except Exception:
      self.logger.exception("Exception occurred in insert_profile_embedding")
      return False
