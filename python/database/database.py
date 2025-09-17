"""
Database module
"""
import logging
from typing import Optional
from urllib.parse import quote
from sqlalchemy.ext.asyncio import (
  create_async_engine,
  async_sessionmaker,
  AsyncSession,
  AsyncEngine,
)

class Database:
  """
  Initialize the database connection for later use.
  """
  __engine: Optional[AsyncEngine] = None
  __session: Optional[async_sessionmaker[AsyncSession]] = None
  logger = logging.getLogger("Database")

  @staticmethod
  async def init(username: str, password: str, host: str, db: str = "asfpc") -> bool:
    """
    Initialize the database connection. You need to call this once only.

    Args:
        username (str): database username.
        password (str): database password.
        host (str): database host in <host>:<port> pattern.
        db (str, optional): database name. Defaults to "asfpc".

    Returns:
        bool: True if the database connection was successful, False otherwise.
    """
    try:
      connection_str: str = (
        f"postgresql+asyncpg://{username}:{quote(password)}@{host}/{db}"
      )
      connection_str_hidden: str = (
        f"postgresql+asyncpg://{username}:{'*' * len(password)}@{host}/{db}"
      )
      Database.logger.info(
        "Initializing database connection to %s",
        connection_str_hidden
      )
      Database.__engine = create_async_engine(connection_str)
      Database.__session = async_sessionmaker(bind=Database.__engine)
      return True
    except Exception as e:
      Database.logger.error("Error initializing database: %s", str(e))
      return False

  @staticmethod
  def get_engine() -> AsyncEngine:
    """Get the database engine.

    Raises:
      RuntimeError: If the database engine is not initialized.

    Returns:
      AsyncEngine: The database engine.
    """
    if not Database.__engine:
      Database.logger.exception("Database engine is not initialized.")
      raise RuntimeError("Database engine is not initialized.")
    return Database.__engine

  @staticmethod
  def get_session() -> AsyncSession:
    """Get the database session.

    Raises:
      RuntimeError: If the database session is not initialized or is closed.

    Returns:
      AsyncSession: The database session.
    """
    if not Database.__session:
      Database.logger.exception("Database session is not initialized.")
      raise RuntimeError("Database session is not initialized.")
    return Database.__session()

  @staticmethod
  async def close():
    """Close the database connection.
    """
    if Database.__session:
      async with Database.__session() as session:
        await session.close()
    if Database.__engine:
      await Database.__engine.dispose()
