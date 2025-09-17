from typing import TYPE_CHECKING

from sqlalchemy import Boolean, ForeignKey, Integer, String
from sqlalchemy.orm import Mapped, mapped_column, relationship
from .base import Base


if TYPE_CHECKING:
  from .profile import UserProfile


class Image(Base):
  __tablename__ = "image"

  id: Mapped[int] = mapped_column(Integer, primary_key=True)
  path: Mapped[str] = mapped_column(String, nullable=False)
  is_analyzed: Mapped[bool] = mapped_column(Boolean, default=False)
  belong_to_id: Mapped[int] = mapped_column(ForeignKey("user_profile.id"), nullable=False)
  belong_to: Mapped["UserProfile"] = relationship(back_populates="images")
