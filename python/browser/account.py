import logging
from zendriver import Browser
from browser.browser import BrowserAutomationService
from browser.facebook import FacebookAutomationService
from database.models import Account, Group

class AccountAutomationService:
  def __init__(self):
    self.logger = logging.getLogger("AccountAutomationService")
    
  async def join_group(self, account: Account, group: Group) -> bool:
    """Join a Facebook group using the provided account.

    Args:
        account (Account): The Account object to use for joining the group.
        group (Group): The Group object representing the group to join.
        group_name (str): The name of the group to join.

    Returns:
        bool: True if the join request was successful, False otherwise.
    """
    try:
      browser = await BrowserAutomationService(
        proxy=account.proxy,
        user_data_dir=account.get_user_data_dir()
      ).get_browser()
      await browser.cookies.set_all(
        BrowserAutomationService.cookie_param_converter(account.cookies)
      )
      return await FacebookAutomationService.join_group(group, browser)
    except Exception as e:
      self.logger.error(f"Error joining group {group.group_id} with account {account.id}: {e}")
      raise e
  
  async def login_account(self, account: Account) -> bool:  # noqa: PLR6301
    """Log in with an account and save its cookies.

    Args:
        account (Account): The Account object to log in.

    Returns:
        bool: True if the login was successful, False otherwise.
    """
    try:
      browser: Browser = await BrowserAutomationService(
        proxy=account.proxy,
        user_data_dir=account.get_user_data_dir()
      ).get_browser()
      await browser.main_tab.get(f"{FacebookAutomationService.url.get("login", "https://www.facebook.com")}?username={account.username}&password={account.password}")
      while True:
        if len(browser.tabs) < 1:
          break
        account.cookies = await browser.cookies.get_all() # type: ignore
        await browser.main_tab.sleep(1)
      return True
    except Exception as e:
      self.logger.error(f"Error logging in account {account.id}: {e}")
      return False