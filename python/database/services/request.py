import logging

from sqlalchemy.ext.asyncio import AsyncSession

from database.database import Database
from database.models import Request


class RequestService:
  def __init__(self):
    self.logger = logging.getLogger("RequestService")

  async def update_request(
    self, request_id: int, session: AsyncSession | None = None, **kwargs
  ) -> bool:
    try:
      if not session:
        session = Database.get_session()
      async with session as conn:
        request = await conn.get(Request, request_id)
        if not request:
          self.logger.error("Request with ID %d not found", request_id)
          return False
        for key, value in kwargs.items():
          setattr(request, key, value)
        await conn.commit()
        await conn.refresh(request)
    except Exception:
      self.logger.exception("Failed to update request: %s", request_id)
      return False
    else:
      return True
