"""Embedded Profile model for storing vector embeddings of user profiles"""

from datetime import datetime
from typing import TYPE_CHECKING

from sqlalchemy import (
  TEXT,
  DateTime,
  ForeignKey,
  Integer,
  TypeDecorator,
  UniqueConstraint,
)
from sqlalchemy.orm import Mapped, mapped_column, relationship

from .base import Base

if TYPE_CHECKING:
  from .profile import UserProfile


class VectorType(TypeDecorator):
  """Custom SQLAlchemy type for handling vector(768) PostgreSQL type."""

  impl = TEXT
  cache_ok = True

  def process_bind_param(self, value: list[float] | None, _) -> str | None:
    """Convert Python list to PostgreSQL vector format."""
    if value is None:
      return None
    # Convert list of floats to PostgreSQL vector format: [1.0,2.0,3.0]
    return "[" + ",".join(str(float(x)) for x in value) + "]"

  def process_result_value(self, value: str | None, _) -> list[float] | None:
    """Convert PostgreSQL vector format to Python list."""
    if value is None:
      return None
    # Parse PostgreSQL vector format [1.0,2.0,3.0] to list
    if value.startswith("[") and value.endswith("]"):
      inner = value[1:-1]
      if not inner.strip():
        return []
      return [float(x.strip()) for x in inner.split(",")]
    return None


class EmbeddedProfile(Base):
  """Embedded Profile model for storing vector embeddings of user profiles"""

  __tablename__ = "embedded_profile"

  id: Mapped[int] = mapped_column(Integer, primary_key=True)
  pid: Mapped[int] = mapped_column(
    ForeignKey("user_profile.id", ondelete="CASCADE", onupdate="CASCADE"),
    unique=True,
    nullable=False,
  )
  embedding: Mapped[list[float] | None] = mapped_column(
    VectorType, nullable=True, comment="768-dimensional vector embedding"
  )
  created_at: Mapped[datetime] = mapped_column(DateTime, default=datetime.now)
  profile: Mapped["UserProfile"] = relationship(
    back_populates="emb_profile", uselist=False
  )
  __table_args__ = (UniqueConstraint("pid", name="embedded_profile_pid_key"),)
