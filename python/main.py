""" "
Main module for the Python service.
"""
import asyncio
import datetime
import logging
import sys
from dotenv import load_dotenv
import os
import argparse

import pandas as pd
from database.database import Database
from utils.navigator import TaskNavigator

load_dotenv(dotenv_path=os.path.join(os.path.dirname(__file__), "..", ".env"))
# Ensure logs directory exists
logs_dir = os.path.join(os.path.dirname(__file__), "logs")
os.makedirs(logs_dir, exist_ok=True)

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
    main = MainProcess()
    silent = main.config.get("silent", False)
    no_log = main.config.get("no-log", False)
    pd.set_option('future.no_silent_downcasting', True)
    if not no_log:
      if silent:
        logging.basicConfig(
          level=logging.INFO,
          format="%(asctime)s [%(levelname)s] %(message)s",
          filemode="w",
          filename=os.path.join(logs_dir, f"{datetime.datetime.now().isoformat().replace(':', '-')}.log"),
        )
      else:
        logging.basicConfig(
          level=logging.INFO,
          format="%(asctime)s [%(levelname)s] %(message)s",
          stream=sys.stderr,
        )
    
    await Database.init(
        username=os.getenv("POSTGRE_USER", "postgres"),
        password=os.getenv("POSTGRE_PASSWORD", "password"),
        host=f'{os.getenv("POSTGRE_HOST", "localhost")}:{os.getenv("POSTGRE_PORT", "5432")}',
        db=os.getenv("POSTGRE_DBNAME", "asfpc"),
    )
    return main
  async def run(self):
    task_navigator = TaskNavigator(self.config)
    task = self.config.get("task")
    if task == "login":
      await task_navigator.login()
      sys.exit(0)
    elif task == "joingroup":
      await task_navigator.join_group()
      sys.exit(0)
    elif task == "train-model":
      await task_navigator.train_model()
      sys.exit(0)
    elif task == "predict":
      await task_navigator.predict()
    else:
      self.logger.error(f"Unknown task: {task}")
      sys.exit(1)

async def execute():
    p = await MainProcess.new()
    await p.run()

if __name__ == "__main__":
    asyncio.run(execute())