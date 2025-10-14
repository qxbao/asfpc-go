from datetime import datetime

from sqlalchemy import Column, DateTime, ForeignKey, Integer, Table

from .base import Base

group_category_table = Table(
    "group_category",
    Base.metadata,
    Column("group_id", Integer, ForeignKey("group.id", ondelete="CASCADE"), primary_key=True),
    Column("category_id", Integer, ForeignKey("category.id", ondelete="CASCADE"), primary_key=True),
    Column("created_at", DateTime, nullable=False, default=datetime.now),
)
