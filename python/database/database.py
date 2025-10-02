# noqa: INP001
import logging
from urllib.parse import quote

from sqlalchemy.ext.asyncio import (
  AsyncEngine,
  AsyncSession,
  async_sessionmaker,
  create_async_engine,
)


class Database:
  """
  Initialize the database connection for later use.
  """

  __engine: AsyncEngine | None = None
  __session: async_sessionmaker[AsyncSession] | None = None
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
      Database.cn_str = connection_str
      connection_str_hidden: str = (
        f"postgresql+asyncpg://{username}:{'*' * len(password)}@{host}/{db}"
      )
      Database.logger.info(
        "Initializing database connection to %s", connection_str_hidden
      )
      Database.__engine = create_async_engine(connection_str)
      Database.__session = async_sessionmaker(bind=Database.__engine)
    except Exception as e:  # noqa: BLE001
      Database.logger.error("Error initializing database: %s", str(e))
      return False
    else:
      return True

  @staticmethod
  def get_engine() -> AsyncEngine:
    """
    Get the database engine.

    Raises:
      RuntimeError: If the database engine is not initialized.

    Returns:
      AsyncEngine: The database engine.

    """
    if not Database.__engine:
      Database.logger.exception("Database engine is not initialized.")
      msg = "Database engine is not initialized."
      raise RuntimeError(msg)
    return Database.__engine

  @staticmethod
  def get_session() -> AsyncSession:
    """
    Get the database session. This session is bound to current event loop.

    Raises:
      RuntimeError: If the database session is not initialized or is closed.

    Returns:
      AsyncSession: The database session.

    """
    if not Database.__session:
      Database.logger.exception("Database session is not initialized.")
      msg = "Database session is not initialized."
      raise RuntimeError(msg)
    return Database.__session()

  @staticmethod
  def get_isolated_session() -> AsyncSession:
    """
    Get an isolated database session for use in background threads.

    This creates a new engine and session that won't be bound to the main event loop,
    allowing safe use in background threads with their own event loops.

    Important: The caller is responsible for properly closing the session and disposing
    the engine when done to prevent resource leaks.

    Returns:
        AsyncSession: A new isolated database session.

    Raises:
        RuntimeError: If the database connection string is not available.

    """
    if not hasattr(Database, "cn_str") or not Database.cn_str:
      Database.logger.exception("Database connection string is not available.")
      msg = "Database connection string is not available. Call Database.init() first."
      raise RuntimeError(msg)

    try:
      # Create a new engine specifically for this isolated session
      isolated_engine = create_async_engine(Database.cn_str, echo=False, future=True)
      isolated_sessionmaker = async_sessionmaker(
        bind=isolated_engine, expire_on_commit=False
      )

      Database.logger.debug("Created isolated database session for background thread")
      return isolated_sessionmaker()

    except Exception as e:
      Database.logger.error("Error creating isolated database session: %s", str(e))
      msg = f"Failed to create isolated database session: {e!s}"
      raise RuntimeError(msg) from e

  @staticmethod
  async def close():
    """Close the database connection."""
    if Database.__session:
      async with Database.__session() as session:
        await session.close()
    if Database.__engine:
      await Database.__engine.dispose()
