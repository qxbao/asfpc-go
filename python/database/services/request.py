import logging
from database.models import Request

from database.database import Database

class RequestService:
  def __init__(self):
    self.logger = logging.getLogger("RequestService")

  async def update_request(self, request_id: int, **kwargs) -> bool:
    try:
      async with Database.get_session() as conn:
        request = await conn.get(Request, request_id)
        if not request:
          self.logger.error("Request with ID %d not found", request_id)
          return False
        for key, value in kwargs.items():
          setattr(request, key, value)
        await conn.commit()
        await conn.refresh(request)
        return True
    except Exception as e:
      self.logger.exception("Failed to update request: %s", str(e))
      return False