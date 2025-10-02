from datetime import datetime

from sqlalchemy import DateTime, Float, Integer, SmallInteger, String, Text
from sqlalchemy.orm import Mapped, mapped_column

from .base import Base


class Request(Base):
  __tablename__ = "request"

  id: Mapped[int] = mapped_column(Integer, primary_key=True)
  progress: Mapped[float] = mapped_column(Float, nullable=False, default=0.0)
  status: Mapped[int] = mapped_column(SmallInteger, nullable=False, default=0)
  description: Mapped[str | None] = mapped_column(String(50), nullable=True)
  created_at: Mapped[datetime] = mapped_column(DateTime, default=datetime.now)
  updated_at: Mapped[datetime] = mapped_column(
    DateTime, default=datetime.now, onupdate=datetime.now
  )
  error_message: Mapped[str | None] = mapped_column(Text, nullable=True)
