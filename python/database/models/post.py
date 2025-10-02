from datetime import datetime
from typing import TYPE_CHECKING

from sqlalchemy import DateTime, ForeignKey, Integer, String
from sqlalchemy.orm import Mapped, mapped_column, relationship

from .base import Base

if TYPE_CHECKING:
  from .comment import Comment
  from .group import Group


class Post(Base):
  __tablename__ = "post"
  id: Mapped[int] = mapped_column(Integer, primary_key=True)
  post_id: Mapped[str] = mapped_column(String, unique=True, nullable=False)
  content: Mapped[str] = mapped_column(String, nullable=False)
  created_at: Mapped[datetime] = mapped_column(DateTime, default=datetime.now)
  inserted_at: Mapped[datetime] = mapped_column(DateTime, default=datetime.now)
  group_id: Mapped[int] = mapped_column(ForeignKey("group.id"), nullable=False)
  group: Mapped["Group"] = relationship(back_populates="posts")
  is_analyzed: Mapped[bool] = mapped_column(default=False)
  comments: Mapped[list["Comment"]] = relationship(back_populates="post", cascade="all")
