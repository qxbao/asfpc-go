"""Embedded Profile model for storing vector embeddings of user profiles"""

from datetime import datetime
from typing import TYPE_CHECKING

from pgvector.sqlalchemy import Vector
from sqlalchemy import (
  DateTime,
  ForeignKey,
  Integer,
  UniqueConstraint,
)
from sqlalchemy.orm import Mapped, mapped_column, relationship

from .base import Base

if TYPE_CHECKING:
  from .category import Category
  from .profile import UserProfile


class EmbeddedProfile(Base):
  """Embedded Profile model for storing vector embeddings of user profiles per category"""

  __tablename__ = "embedded_profile"

  id: Mapped[int] = mapped_column(Integer, primary_key=True)
  pid: Mapped[int] = mapped_column(
    ForeignKey("user_profile.id", ondelete="CASCADE", onupdate="CASCADE"),
    nullable=False,
  )
  cid: Mapped[int] = mapped_column(
    ForeignKey("category.id", ondelete="SET NULL", onupdate="CASCADE"),
    nullable=False,
  )
  embedding: Mapped[list[float] | None] = mapped_column(
    Vector(1024), nullable=True, comment="1024-dimensional BGE-M3 embedding"
  )
  created_at: Mapped[datetime] = mapped_column(DateTime, default=datetime.now)
  profile: Mapped["UserProfile"] = relationship(
    back_populates="emb_profiles", foreign_keys=[pid]
  )
  category: Mapped["Category"] = relationship(foreign_keys=[cid])
  __table_args__ = (UniqueConstraint("pid", "cid", name="embedded_profile_pid_key"),)
