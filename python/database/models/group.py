from datetime import datetime
from typing import TYPE_CHECKING

from pydantic import BaseModel
from sqlalchemy import DateTime, ForeignKey, UniqueConstraint
from sqlalchemy.orm import Mapped, mapped_column, relationship

from .base import Base

if TYPE_CHECKING:
  from .account import Account
  from .category import Category
  from .post import Post


class GroupSchema(BaseModel):
  """Schema for Group model."""

  id: int
  group_id: str
  group_name: str
  is_joined: bool
  model_config = {"from_attributes": True}


class Group(Base):
  __tablename__ = "group"
  id: Mapped[int] = mapped_column(primary_key=True)
  group_id: Mapped[str] = mapped_column(nullable=False)
  group_name: Mapped[str] = mapped_column(nullable=False)
  is_joined: Mapped[bool] = mapped_column(default=False)
  scanned_at: Mapped[datetime | None] = mapped_column(DateTime, nullable=True)
  account_id: Mapped[int | None] = mapped_column(
    ForeignKey("account.id"), nullable=True
  )
  account: Mapped["Account | None"] = relationship(
    back_populates="groups", lazy="selectin"
  )
  posts: Mapped[list["Post"]] = relationship(back_populates="group")
  categories: Mapped[list["Category"]] = relationship(
    secondary="group_category", back_populates="groups"
  )
  __table_args__ = (
    UniqueConstraint("group_id", "account_id", name="uq_group_account"),
  )

  def to_schema(self) -> GroupSchema:
    """Convert the Group object to GroupSchema"""
    return GroupSchema.model_validate(self)

  def to_json(self) -> dict:
    """Convert the Group object to a JSON-serializable dictionary."""
    return GroupSchema.model_validate(self).model_dump()
