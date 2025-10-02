"""Account model"""

from datetime import datetime
from pathlib import Path
from typing import TYPE_CHECKING

import sqlalchemy
from pydantic import BaseModel
from sqlalchemy import (
  Boolean,
  DateTime,
  Dialect,
  ForeignKey,
  Integer,
  String,
  TypeDecorator,
)
from sqlalchemy.orm import Mapped, mapped_column, relationship
from zendriver.cdp.network import Cookie

from .base import Base

if TYPE_CHECKING:
  from .group import Group
  from .profile import UserProfile
  from .proxy import Proxy


class AccountSchema(BaseModel):
  """Schema for Account model."""

  id: int
  username: str
  email: str
  is_block: bool
  ua: str
  created_at: datetime
  updated_at: datetime
  model_config = {"from_attributes": True}


class CookieType(TypeDecorator):
  """Custom SQLAlchemy type for handling CookieParam objects."""

  impl = sqlalchemy.types.JSON

  def process_bind_param(
    self, value: list[Cookie] | None, _: Dialect
  ) -> list[dict] | None:
    return [cookie.to_json() for cookie in value] if value else None

  def process_result_value(
    self, value: dict | None, _: Dialect
  ) -> list[Cookie] | None:
    return [Cookie.from_json(cookie) for cookie in value] if value else None


class Account(Base):
  """Account model for the application."""

  __tablename__ = "account"

  id: Mapped[int] = mapped_column(Integer, primary_key=True)
  username: Mapped[str] = mapped_column(String, unique=True, nullable=False)
  email: Mapped[str] = mapped_column(String, nullable=False)
  password: Mapped[str] = mapped_column(String, nullable=False)
  is_block: Mapped[bool] = mapped_column(Boolean, default=False)
  ua: Mapped[str] = mapped_column(String, nullable=False)
  created_at: Mapped[datetime] = mapped_column(DateTime, default=datetime.now)
  updated_at: Mapped[datetime] = mapped_column(
    DateTime, default=datetime.now, onupdate=datetime.now
  )
  cookies: Mapped[list[Cookie]] = mapped_column(CookieType, nullable=True, default=None)
  access_token: Mapped[str] = mapped_column(String, default=None, nullable=True)
  proxy_id: Mapped[int | None] = mapped_column(
    ForeignKey("proxy.id"), nullable=True, default=None
  )
  proxy: Mapped["Proxy | None"] = relationship(back_populates="accounts")
  scraped_profiles: Mapped[set["UserProfile"]] = relationship(
    back_populates="scraped_by", lazy="selectin"
  )
  groups: Mapped[set["Group"]] = relationship(back_populates="account", lazy="selectin")

  def to_schema(self) -> AccountSchema:
    """Convert the Account object to an AccountSchema."""
    return AccountSchema.model_validate(self)

  def to_json(self) -> dict:
    """Convert the Account object to a JSON serializable dictionary."""
    return AccountSchema.model_validate(self).model_dump()

  def get_user_data_dir(self) -> str:
    """
    Get the user data directory for the account.

    Returns:
        str: The user data directory path.

    """
    user_data_dir = Path(
      Path().cwd() / "resources" / "user_data_dir" / str(self.id)
    )
    if not user_data_dir.exists():
      user_data_dir.mkdir(parents=True)
    return user_data_dir
