from datetime import datetime

from sqlalchemy import Column, DateTime, Float, ForeignKey, Integer, Table

from .base import Base

user_profile_category_table = Table(
    "user_profile_category",
    Base.metadata,
    Column("user_profile_id", Integer, ForeignKey("user_profile.id", ondelete="CASCADE"), primary_key=True),
    Column("category_id", Integer, ForeignKey("category.id", ondelete="CASCADE"), primary_key=True),
    Column("gemini_score", Float, nullable=True),
    Column("model_score", Float, nullable=True),
    Column("created_at", DateTime, nullable=False, default=datetime.now),
)
