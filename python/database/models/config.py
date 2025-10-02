from sqlalchemy import Integer, String
from sqlalchemy.orm import Mapped, mapped_column

from .base import Base


class Config(Base):
  __tablename__ = "config"

  id: Mapped[int] = mapped_column(Integer, primary_key=True, index=True)
  key: Mapped[str] = mapped_column(String, unique=True, index=True, nullable=False)
  value: Mapped[str] = mapped_column(String, nullable=False)
