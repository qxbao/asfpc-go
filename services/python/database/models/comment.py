from typing import TYPE_CHECKING
from sqlalchemy import Integer, String, DateTime, ForeignKey
from sqlalchemy.orm import Mapped, mapped_column, relationship
from datetime import datetime
from .base import Base

if TYPE_CHECKING:
  from .post import Post
  from .profile import UserProfile

class Comment(Base):
  __tablename__ = "comment"

  id: Mapped[int] = mapped_column(Integer, primary_key=True)
  comment_id: Mapped[str] = mapped_column(String, unique=True, nullable=False)
  author_id: Mapped[int] = mapped_column(ForeignKey("user_profile.id"), nullable=False)
  author: Mapped["UserProfile"] = relationship(back_populates="comments")
  content: Mapped[str] = mapped_column(String, nullable=False)
  is_analyzed: Mapped[bool] = mapped_column(default=False)
  created_at: Mapped[datetime] = mapped_column(DateTime, default=datetime.now)
  inserted_at: Mapped[datetime] = mapped_column(DateTime, default=datetime.now)
  post_id: Mapped[int] = mapped_column(ForeignKey("post.id"), nullable=False)
  post: Mapped["Post"] = relationship(back_populates="comments")