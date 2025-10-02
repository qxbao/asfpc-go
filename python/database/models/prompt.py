from datetime import datetime

from sqlalchemy import DateTime, Integer, String, UniqueConstraint
from sqlalchemy.orm import Mapped, mapped_column

from .base import Base


class Prompt(Base):
  """Model for storing prompts"""

  __tablename__ = "prompt"
  id: Mapped[int] = mapped_column(Integer, primary_key=True)
  content: Mapped[str] = mapped_column(String, nullable=False)
  service_name: Mapped[str] = mapped_column(String, nullable=False)
  version: Mapped[int] = mapped_column(Integer, nullable=False)
  created_by: Mapped[str] = mapped_column(String, nullable=False)
  created_at: Mapped[datetime] = mapped_column(DateTime, default=datetime.now)
  __table_args__: tuple = (
    UniqueConstraint("service_name", "version", name="uq_service_name_version"),
  )
