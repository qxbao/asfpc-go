from typing import Optional
from zendriver.core.config import PathLike
import zendriver
import logging
from database.models.proxy import Proxy
from zendriver.cdp.network import Cookie, CookieParam

class BrowserAutomationService:
  proxy: Optional[Proxy]
  user_data_dir: Optional[PathLike]
  logger = logging.getLogger("BrowserAutomationService")

  def __init__(self, proxy: Optional[Proxy] = None, user_data_dir: Optional[PathLike] = None):
    self.proxy = proxy
    self.user_data_dir = user_data_dir

  async def get_browser(self, browser_args: list[str] = [], **kwargs) -> zendriver.Browser:
    if self.proxy:
      proxy_arg = f"--proxy-server=http://{self.proxy.host}:{self.proxy.port}"
      browser_args.append(proxy_arg)
    browser = await zendriver.Browser.create(
      headless=False,
      user_data_dir=self.user_data_dir,
      args=browser_args,
      **kwargs
    )
    await browser.start()
    if self.proxy:
      await self.__config_proxy(browser, self.proxy.username, self.proxy.password)
    return browser
  
  async def __config_proxy(self, browser: zendriver.Browser, username: str, password: str):
    tab = browser.main_tab
    self.logger.info(f"username={username}")

    def req_paused(event: zendriver.cdp.fetch.RequestPaused):
        try:
            tab.feed_cdp(
                zendriver.cdp.fetch.continue_request(event.request_id, url=event.request.url)
            )
        except Exception as e:
            self.logger.warning(f"Failed to continue request {event.request_id}: {e}")

    def auth_challenge_handler(event: zendriver.cdp.fetch.AuthRequired):
        self.logger.info(f"auth_challenge_handler: {username}")
        tab.feed_cdp(
            zendriver.cdp.fetch.continue_with_auth(
                request_id=event.request_id,
                auth_challenge_response=zendriver.cdp.fetch.AuthChallengeResponse(
                    response="ProvideCredentials",
                    username=username,
                    password=password,
                ),
            )
        )

    tab.add_handler(zendriver.cdp.fetch.RequestPaused, req_paused)
    tab.add_handler(zendriver.cdp.fetch.AuthRequired, auth_challenge_handler)
    await tab.send(zendriver.cdp.fetch.enable(handle_auth_requests=True))
  
  @staticmethod
  def cookie_param_converter(cookies: list[Cookie]) -> list[CookieParam]:
    """Convert Zendriver list[Cookie] to list[CookieParam]

    Args:
        cookies (list[Cookie]): The list of Zendriver cookies.

    Returns:
        list[CookieParam]: A list representation of the cookies.
    """
    return [CookieParam.from_json(cookie.to_json()) for cookie in cookies]