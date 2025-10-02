"""Account service module"""

import logging

from sqlalchemy import select

from database.database import Database
from database.models import Account


class AccountService:
  """
  Service for managing Account entities.
  """

  def __init__(self):
    self.logger = logging.getLogger("AccountService")

  async def add_account(self, username, password, **kwargs) -> Account:
    """
    Add a new account.

    Args:
        username (str): The username for the account.
        password (str): The password for the account.
        **kwargs: Additional keyword arguments for the account.

    Returns:
        Account: The created Account object.

    """
    async with Database.get_session() as conn:
      account = Account(username=username, password=password, **kwargs)
      conn.add(account)
      await conn.commit()
      await conn.refresh(account)
      conn.expunge(account)
      return account

  async def get_all_account(self) -> list[Account]:
    """
    Get all accounts.

    Returns:
        List[Account]: A list of all Account objects.

    """
    async with Database.get_session() as conn:
      result = await conn.execute(select(Account))
      return list(result.scalars().all())

  async def get_ok_account(self) -> Account | None:
    """
    Get a working account.

    Returns:
        Account | None: The valid Account object or None if not found.

    """
    async with Database.get_session() as conn:
      result = await conn.execute(
        select(Account).where(
          Account.is_block.is_not(False),
          Account.access_token.is_not(None),
          Account.cookies.is_not(None),
        )
      )
      return result.scalar_one_or_none()

  async def get_account_by_id(self, account_id: int) -> Account | None:
    """
    Get an account by its ID.

    Args:
        account_id (int): The ID of the account.

    Returns:
        Account | None: The Account object with the given ID or None if not found.

    """
    return await Database.get_session().get(Account, account_id)

  async def update_account(self, account: Account) -> None:
    """
    Update an existing account.

    Args:
        account (Account): The Account object to update.

    """
    async with Database.get_session() as conn:
      merged_account = await conn.merge(account)
      await conn.commit()
      await conn.refresh(merged_account)
