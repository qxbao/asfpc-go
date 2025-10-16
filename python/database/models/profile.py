from datetime import datetime
from typing import TYPE_CHECKING

from sqlalchemy import VARCHAR, DateTime, ForeignKey, Integer, String, Text
from sqlalchemy.orm import Mapped, mapped_column, relationship

from .base import Base

if TYPE_CHECKING:
  from .account import Account
  from .category import Category
  from .comment import Comment
  from .emb_profile import EmbeddedProfile


class UserProfile(Base):
  """User Profile model for storing scraped Facebook profile data"""

  __tablename__ = "user_profile"

  id: Mapped[int] = mapped_column(Integer, primary_key=True)
  facebook_id: Mapped[str] = mapped_column(String, unique=True, nullable=False)
  name: Mapped[str | None] = mapped_column(String, nullable=True)
  bio: Mapped[str | None] = mapped_column(Text, nullable=True)
  location: Mapped[str | None] = mapped_column(String, nullable=True)
  work: Mapped[str | None] = mapped_column(Text, nullable=True)
  hometown: Mapped[str | None] = mapped_column(String, nullable=True)
  education: Mapped[str | None] = mapped_column(Text, nullable=True)
  relationship_status: Mapped[str | None] = mapped_column(String, nullable=True)
  profile_url: Mapped[str] = mapped_column(String, nullable=False)
  locale: Mapped[str] = mapped_column(
    VARCHAR(16), nullable=False, default="NOT_SPECIFIED"
  )
  gender: Mapped[str | None] = mapped_column(VARCHAR(16), nullable=True)
  birthday: Mapped[str | None] = mapped_column(VARCHAR(10), nullable=True)
  email: Mapped[str | None] = mapped_column(VARCHAR(100), nullable=True)
  is_scanned: Mapped[bool] = mapped_column(nullable=False, default=False)
  is_analyzed: Mapped[bool] = mapped_column(nullable=False, default=False)
  phone: Mapped[str | None] = mapped_column(VARCHAR(12), nullable=True)
  created_at: Mapped[datetime] = mapped_column(DateTime, default=datetime.now)
  updated_at: Mapped[datetime] = mapped_column(
    DateTime, default=datetime.now, onupdate=datetime.now
  )
  scraped_by_id: Mapped[int] = mapped_column(ForeignKey("account.id"), nullable=False)
  scraped_by: Mapped["Account"] = relationship(back_populates="scraped_profiles")
  emb_profiles: Mapped[list["EmbeddedProfile"]] = relationship(
    back_populates="profile", lazy="selectin", foreign_keys="EmbeddedProfile.pid"
  )
  comments: Mapped[list["Comment"]] = relationship(back_populates="author")
  categories: Mapped[list["Category"]] = relationship(
    secondary="user_profile_category", back_populates="user_profiles"
  )

  def to_df(self, category_id: int) -> dict:
    """
    Convert profile data to a DataFrame for a specific category

    Args:
        category_id: The category ID to get the embedding for

    Returns:
        Dictionary with profile features including category-specific embedding

    """
    embedding = None
    for emb in self.emb_profiles:
      if emb.cid == category_id:
        embedding = emb.embedding
        break

    return {
      "embedding": [embedding] if embedding is not None else [None],
      "gender": self.gender,
      "relationship_status": self.relationship_status,
      "locale": self.locale,
      "birthday": self.birthday,
    }
