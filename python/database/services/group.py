"""Service for managing groups."""
import logging
from sqlalchemy import select
from sqlalchemy.orm import selectinload
from database.database import Database
from database.models.account import Account
from database.models.group import Group

class GroupService:
  def __init__(self):
    self.logger = logging.getLogger("GroupService")

  async def get_group_by_id(self, group_id: int, include_account: bool = False) -> Group | None:
    try:
      async with Database.get_session() as conn:
        if include_account:
          return await conn.get(Group, group_id, options=[selectinload(Group.account, Account.proxy)])
      return await Database.get_session().get(Group, group_id)
    except Exception as e:
      self.logger.exception(e)
      return None

  async def get_group_by_gid(self, group_gid: str, lazyload: bool = False) -> Group | None:
    try:
      async with Database.get_session() as conn:
        query = select(Group).where(Group.group_id == group_gid)
        if lazyload:
          query = query.options(selectinload(Group.account))
        res = await conn.execute(
          query
        )
        return res.scalar_one_or_none()
    except Exception as e:
      self.logger.exception(e)
      raise RuntimeError("Failed to retrieve group by GID: " + str(e))

  async def update_group(self, group: Group) -> None:
    async with Database.get_session() as conn:
      merged = await conn.merge(group)
      await conn.commit()
      await conn.refresh(merged)

  async def link_group(self,
      account: Account,
      group_id: str,
      group_name: str,
      is_joined: bool=False) -> Group:
    try:
      async with Database.get_session() as conn:
        account = await conn.merge(account)
        group = (await conn.execute(
          select(Group).where(
            Group.group_id == group_id,
            Group.group_name == group_name,
          )
        )).scalar_one_or_none()
        if not group:
          group = Group(
            group_name=group_name,
            group_id=group_id,
            is_joined=is_joined
          )
        group.account = account
        conn.add(group)
        await conn.commit()
        await conn.refresh(group)
        return group
    except Exception as e:
      self.logger.exception(e)
      raise RuntimeError(e)