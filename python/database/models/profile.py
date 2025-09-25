from datetime import datetime
from typing import TYPE_CHECKING, List, Optional
from sqlalchemy import VARCHAR, ForeignKey, Integer, String, Text, DateTime
from sqlalchemy.orm import mapped_column, Mapped, relationship

from .base import Base

if TYPE_CHECKING:
    from .emb_profile import EmbeddedProfile
    from .account import Account
    from .image import Image
    from .comment import Comment


class UserProfile(Base):
    """User Profile model for storing scraped Facebook profile data"""

    __tablename__ = "user_profile"

    id: Mapped[int] = mapped_column(Integer, primary_key=True)
    facebook_id: Mapped[str] = mapped_column(String, unique=True, nullable=False)
    name: Mapped[Optional[str]] = mapped_column(String, nullable=True)
    bio: Mapped[Optional[str]] = mapped_column(Text, nullable=True)
    location: Mapped[Optional[str]] = mapped_column(String, nullable=True)
    work: Mapped[Optional[str]] = mapped_column(Text, nullable=True)
    education: Mapped[Optional[str]] = mapped_column(Text, nullable=True)
    relationship_status: Mapped[Optional[str]] = mapped_column(String, nullable=True)
    profile_url: Mapped[str] = mapped_column(String, nullable=False)
    locale: Mapped[str] = mapped_column(
        VARCHAR(16), nullable=False, default="NOT_SPECIFIED"
    )
    gender: Mapped[Optional[str]] = mapped_column(VARCHAR(16), nullable=True)
    birthday: Mapped[Optional[str]] = mapped_column(VARCHAR(10), nullable=True)
    email: Mapped[Optional[str]] = mapped_column(VARCHAR(100), nullable=True)
    is_scanned: Mapped[bool] = mapped_column(nullable=False, default=False)
    is_analyzed: Mapped[bool] = mapped_column(nullable=False, default=False)
    gemini_score: Mapped[Optional[float]] = mapped_column(nullable=True)
    phone: Mapped[Optional[str]] = mapped_column(VARCHAR(12), nullable=True)
    created_at: Mapped[datetime] = mapped_column(DateTime, default=datetime.now)
    updated_at: Mapped[datetime] = mapped_column(
        DateTime, default=datetime.now, onupdate=datetime.now
    )
    scraped_by_id: Mapped[int] = mapped_column(ForeignKey("account.id"), nullable=False)
    scraped_by: Mapped["Account"] = relationship(back_populates="scraped_profiles")
    images: Mapped[List["Image"]] = relationship(back_populates="belong_to")
    emb_profile: Mapped[Optional["EmbeddedProfile"]] = relationship(back_populates="profile", lazy="selectin", uselist=False)
    comments: Mapped[List["Comment"]] = relationship(back_populates="author")
    
    def to_df(self) -> dict:
        """Convert profile data to a DataFrame"""
        return {
            "embedding": [self.emb_profile.embedding if self.emb_profile else None],
            "gender": self.gender,
            "relationship_status": self.relationship_status,
            "locale": self.locale,
            "birthday": self.birthday,
            "gemini_score": self.gemini_score,
        }