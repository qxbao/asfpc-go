"""Service for managing groups."""
import logging
from sqlalchemy import select
from sqlalchemy.orm import selectinload
from database.database import Database
from database.models.account import Account
from database.models.group import Group

class GroupService:
  def __init__(self):
    self.__session = Database.get_session()
    self.logger = logging.getLogger("GroupService")

  async def get_group_by_id(self, group_id: int, include_account: bool = False) -> Group | None:
    try:
      async with self.__session as conn:
        if include_account:
          return await conn.get(Group, group_id, options=[selectinload(Group.account, Account.proxy)])
      return await self.__session.get(Group, group_id)
    except Exception as e:
      self.logger.exception(e)
      return None

  async def get_group_by_gid(self, group_gid: str, lazyload: bool = False) -> Group | None:
    try:
      async with self.__session as conn:
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
    async with self.__session as conn:
      merged = await conn.merge(group)
      await conn.commit()
      await conn.refresh(merged)

  async def link_group(self,
      account: Account,
      group_id: str,
      group_name: str,
      is_joined: bool=False) -> Group:
    try:
      async with self.__session as conn:
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
        self.__session.add(group)
        await self.__session.commit()
        await self.__session.refresh(group)
        return group
    except Exception as e:
      self.logger.exception(e)
      await self.__session.rollback()
      raise RuntimeError(e)