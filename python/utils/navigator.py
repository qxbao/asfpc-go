import sys
from typing import Any

from database.services.account import AccountService
from browser.account import AccountAutomationService
from browser.group import GroupAutomationService
from database.services.group import GroupService
from utils.dialog import DialogUtil


class TaskNavigator:
  def __init__(self, config: dict[str, Any]):
    self.config = config
    
  async def login(self) -> None:
    user_id = self.config.get("uid", None)
    if not user_id:
      raise ValueError("--uid is required for login task")
    account_service = AccountService()
    account = await account_service.get_account_by_id(int(user_id))
    if not account:
      raise ValueError(f"Account with id {user_id} not found")
    account_automation_service = AccountAutomationService()
    is_ok = await account_automation_service.login_account(account)
    if not is_ok:
      raise ValueError(f"Failed to login account with id {user_id}")
    
    is_blocked = await DialogUtil.confirmation("Account Status", "Is the account blocked?")
    account.is_block = is_blocked
    await account_service.update_account(account)
  
  async def join_group(self) -> None:
    group_id = self.config.get("group_id", None)
    if not group_id:
      raise ValueError("--group_id is required for join_group task")
    gs = GroupService()
    group = await gs.get_group_by_id(int(group_id), include_account=True)
    if not group:
      raise ValueError(f"Group with id {group_id} not found")
    gas = GroupAutomationService()
    is_ok = await gas.join_group(group)
    group.is_joined = is_ok
    await gs.update_group(group)
    if not is_ok:
      sys.exit(1)
    else:
      sys.exit(0)