from datetime import datetime
from sqlalchemy import Integer, String, Boolean, DateTime, UniqueConstraint
from sqlalchemy.orm import mapped_column, relationship, Mapped
from .base import Base

from typing import TYPE_CHECKING
if TYPE_CHECKING:
  from .account import Account

class Proxy(Base):
  __tablename__ = "proxy"

  id: Mapped[int] = mapped_column(Integer, primary_key=True)
  ip: Mapped[str] = mapped_column(String, nullable=False)
  port: Mapped[str] = mapped_column(String, nullable=False)
  username: Mapped[str] = mapped_column(String, nullable=False)
  password: Mapped[str] = mapped_column(String, nullable=False)
  is_active: Mapped[bool] = mapped_column(Boolean, default=True)
  created_at: Mapped[datetime] = mapped_column(DateTime, default=datetime.now)
  updated_at: Mapped[datetime] = mapped_column(DateTime, default=datetime.now,
                                               onupdate=datetime.now)
  accounts: Mapped[list["Account"]] = relationship(back_populates="proxy")
  __table_args__: tuple = (UniqueConstraint("ip", "port", "username", name="uq_ip_port_username"),)

  def get_proxy_url(self) -> str:
    return f"http://{self.username}:{self.password}@{self.ip}:{self.port}"