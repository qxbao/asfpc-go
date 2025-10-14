from datetime import datetime
from typing import TYPE_CHECKING

from sqlalchemy import Column, DateTime, Integer, String
from sqlalchemy.orm import Mapped, relationship

from .base import Base

if TYPE_CHECKING:
    from .group import Group
    from .profile import UserProfile
    from .prompt import Prompt


class Category(Base):
    __tablename__ = "category"

    id = Column(Integer, primary_key=True, autoincrement=True)
    name = Column(String, nullable=False)
    description = Column(String)
    created_at = Column(DateTime, nullable=False, default=datetime.now)
    updated_at = Column(DateTime, nullable=False, default=datetime.now, onupdate=datetime.now)

    # Relationships
    prompts: Mapped[list["Prompt"]] = relationship(back_populates="category")
    groups: Mapped[list["Group"]] = relationship(
        secondary="group_category", back_populates="categories"
    )
    user_profiles: Mapped[list["UserProfile"]] = relationship(
        secondary="user_profile_category", back_populates="categories"
    )

    def __repr__(self):
        return f"<Category(id={self.id}, name='{self.name}')>"
