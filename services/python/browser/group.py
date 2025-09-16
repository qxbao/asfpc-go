import logging
from browser.browser import BrowserAutomationService
from browser.facebook import FacebookAutomationService
from database.models.group import Group

class GroupAutomationService:
  def __init__(self):
    self.logger = logging.getLogger("GroupService")
    
  async def join_group(self, group: Group) -> bool:
    """Join a Facebook group using the provided account.

    Args:
        group_id (str): The ID of the group to join.
    Returns:
        bool: True if the join request was successful, False otherwise.
    """
    try:
      browser = await BrowserAutomationService(
        proxy=group.account.proxy,
        user_data_dir=group.account.get_user_data_dir()
      ).get_browser()
      await browser.cookies.set_all(
        BrowserAutomationService.cookie_param_converter(group.account.cookies)
      )
      return await FacebookAutomationService.join_group(group, browser)
    except Exception as e:
      self.logger.error(f"Error joining group {group.group_id} with account {group.account.id}: {e}")
      return False