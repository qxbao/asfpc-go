from datetime import datetime
from typing import Optional
from sqlalchemy import Integer, String, Float, SmallInteger, DateTime, Text
from sqlalchemy.orm import mapped_column, Mapped
from .base import Base


class Request(Base):
  __tablename__ = "request"

  id: Mapped[int] = mapped_column(Integer, primary_key=True)
  progress: Mapped[float] = mapped_column(Float, nullable=False, default=0.0)
  status: Mapped[int] = mapped_column(SmallInteger, nullable=False, default=0)
  description: Mapped[Optional[str]] = mapped_column(String(50), nullable=True)
  created_at: Mapped[datetime] = mapped_column(DateTime, default=datetime.now)
  updated_at: Mapped[datetime] = mapped_column(DateTime, default=datetime.now, onupdate=datetime.now)
  error_message: Mapped[Optional[str]] = mapped_column(Text, nullable=True)