import logging

import zendriver
from zendriver.cdp.network import Cookie, CookieParam
from zendriver.core.config import PathLike

from database.models.proxy import Proxy


class BrowserAutomationService:
  proxy: Proxy | None
  user_data_dir: PathLike | None
  logger = logging.getLogger("BrowserAutomationService")

  def __init__(self, proxy: Proxy | None = None, user_data_dir: PathLike | None = None):
    self.proxy = proxy
    self.user_data_dir = user_data_dir

  async def get_browser(
    self, browser_args: list[str] | None = None, **kwargs
  ) -> zendriver.Browser:
    if browser_args is None:
      browser_args = []
    if self.proxy:
      proxy_arg = f"--proxy-server=http://{self.proxy.host}:{self.proxy.port}"
      browser_args.append(proxy_arg)
    browser = await zendriver.Browser.create(
      headless=False, user_data_dir=self.user_data_dir, args=browser_args, **kwargs
    )
    await browser.start()
    if self.proxy:
      await self.__config_proxy(browser, self.proxy.username, self.proxy.password)
    return browser

  async def __config_proxy(
    self, browser: zendriver.Browser, username: str, password: str
  ):
    tab = browser.main_tab
    self.logger.info("username=%s", username)

    def req_paused(event: zendriver.cdp.fetch.RequestPaused):
      try:
        tab.feed_cdp(
          zendriver.cdp.fetch.continue_request(event.request_id, url=event.request.url)
        )
      except RuntimeError:
        self.logger.warning("Failed to continue request %s", event.request_id)

    def auth_challenge_handler(event: zendriver.cdp.fetch.AuthRequired):
      self.logger.info("auth_challenge_handler: %s", username)
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
    """
    Convert Zendriver list[Cookie] to list[CookieParam]

    Args:
        cookies (list[Cookie]): The list of Zendriver cookies.

    Returns:
        list[CookieParam]: A list representation of the cookies.

    """
    return [CookieParam.from_json(cookie.to_json()) for cookie in cookies]
