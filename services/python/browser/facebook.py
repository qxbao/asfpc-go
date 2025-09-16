from zendriver import Browser

from database.models.group import Group
from utils.dialog import DialogUtil

class FacebookAutomationService:
  url: dict[str, str] = {
    "login": "https://www.facebook.com/",
    "group": "https://www.facebook.com/groups/",
  }

  @staticmethod
  async def join_group(group: Group, browser: Browser) -> bool:
    await browser.get(FacebookAutomationService.url["group"] + group.group_id)
    while True:
      if len(browser.tabs) < 1:
        break
      await browser.main_tab.sleep(1)
    is_joined = await DialogUtil.confirmation(
      "Status Confirmation",
      "Did you join the group successfully?")
    return is_joined