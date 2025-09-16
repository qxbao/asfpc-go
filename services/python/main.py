""" "
Main module for the Python service.
"""
import asyncio
import logging
import sys
from dotenv import load_dotenv
import os
import argparse
from database.database import Database
from utils.navigator import TaskNavigator

load_dotenv(dotenv_path=os.path.join(os.path.dirname(__file__), "..", "..", ".env"))

class MainProcess:
  def __init__(self):
    """
    Should not be called directly. Use new() instead.
    """
    parser = argparse.ArgumentParser()
    parser.add_argument("--task", type=str, required=True, help="Task to perform")
    args, unknown = parser.parse_known_args()
    self.config = vars(args)
    for i in range(0, len(unknown)):
      if unknown[i].startswith("--"):
        key = unknown[i].lstrip("-")
        if i + 1 < len(unknown) and not unknown[i+1].startswith("--"):
          self.config[key] = unknown[i+1]
        elif "=" in unknown[i]:
          k, v = key.split("=", 1)
          self.config[k] = v
        else:
          self.config[key] = True
      
    self.logger = logging.getLogger("MainProcess")

  @staticmethod
  async def new() -> "MainProcess":
    await Database.init(
        username=os.getenv("POSTGRE_USER", "postgres"),
        password=os.getenv("POSTGRE_PASSWORD", "password"),
        host=f'{os.getenv("POSTGRE_HOST", "localhost")}:{os.getenv("POSTGRE_PORT", "5432")}',
        db=os.getenv("POSTGRE_DBNAME", "asfpc"),
    )
    return MainProcess()
  async def run(self):
    task_navigator = TaskNavigator(self.config)
    task = self.config.get("task")
    if task == "login":
      await task_navigator.login()
      sys.exit(0)
    elif task == "joingroup":
      await task_navigator.join_group()
      sys.exit(0)
    else:
      self.logger.error(f"Unknown task: {task}")
      sys.exit(1)

async def execute():
    p = await MainProcess.new()
    await p.run()

if __name__ == "__main__":
    logging.basicConfig(
      level=logging.INFO,
      format="%(asctime)s [%(levelname)s] %(message)s",
      stream=sys.stderr,
    )
    asyncio.run(execute())