import logging
from typing import TYPE_CHECKING

from browser.browser import BrowserAutomationService
from browser.facebook import FacebookAutomationService
from database.models import Account, Group

if TYPE_CHECKING:
  from zendriver import Browser

class AccountAutomationService:
  def __init__(self):
    self.logger = logging.getLogger("AccountAutomationService")

  async def join_group(self, account: Account, group: Group) -> bool:
    """
    Join a Facebook group using the provided account.

    Args:
        account (Account): The Account object to use for joining the group.
        group (Group): The Group object representing the group to join.
        group_name (str): The name of the group to join.

    Returns:
        bool: True if the join request was successful, False otherwise.

    """
    try:
      browser = await BrowserAutomationService(
        proxy=account.proxy, user_data_dir=account.get_user_data_dir()
      ).get_browser()
      await browser.cookies.set_all(account.cookies)
      return await FacebookAutomationService.join_group(group, browser)
    except Exception as e:
      self.logger.exception(
        "Error joining group %s with account %s", group.group_id, account.id
      )
      raise RuntimeError(e) from e

  async def login_account(self, account: Account) -> bool:
    """
    Log in with an account and save its cookies.

    Args:
        account (Account): The Account object to log in.

    Returns:
        bool: True if the login was successful, False otherwise.

    """
    try:
      browser: Browser = await BrowserAutomationService(
        proxy=account.proxy, user_data_dir=account.get_user_data_dir()
      ).get_browser()
      await browser.main_tab.get(
        f"{FacebookAutomationService.url.get('login', 'https://www.facebook.com')}?username={account.username}&password={account.password}"
      )
      while True:
        if len(browser.tabs) < 1:
          break
        account.cookies = await browser.cookies.get_all()
        await browser.main_tab.sleep(1)
    except Exception:
      self.logger.exception("Error logging in account %s", account.id)
      return False
    else:
      return True
