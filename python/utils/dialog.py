import asyncio
import contextlib
import logging
import os
import platform
import subprocess
import threading
import tkinter as tk
from tkinter import messagebox


class DialogUtil:
  @staticmethod
  async def confirmation(
    title: str = "Confirmation", message: str = "Are you sure?"
  ) -> bool:
    if platform.system() == "Darwin":
      try:
        return await DialogUtil._macos_confirmation(title, message)
      except Exception:
        logging.getLogger("DialogUtil").exception("macOS native dialog failed, falling back to Tkinter")

    loop = asyncio.get_event_loop()
    future = loop.create_future()

    def show_dialog():
      try:
        root = tk.Tk()
        root.withdraw()

        if platform.system() == "Darwin":
          with contextlib.suppress(Exception):
            subprocess.run(  # noqa: S603
              [
                "/usr/bin/osascript",
                "-e",
                f'tell application "System Events" to set frontmost of first process whose unix id is {os.getpid()} to true',
              ],
              check=False,
              capture_output=True,
            )

          # Configure for macOS
          root.lift()
          root.call("wm", "attributes", ".", "-topmost", True)
          root.after_idle(root.attributes, "-topmost", False)
        else:
          root.lift()
          root.attributes("-topmost", True)
          root.attributes("-alpha", 0.0)  # Make invisible initially
          root.update()
          root.attributes("-alpha", 1.0)  # Make visible
          root.focus_force()

        result = messagebox.askyesno(title, message, parent=root)
        loop.call_soon_threadsafe(future.set_result, result)
      except Exception as e:  # noqa: BLE001
        loop.call_soon_threadsafe(future.set_exception, e)
      finally:
        with contextlib.suppress(Exception):
          root.destroy()  # type: ignore[call-arg]

    thread = threading.Thread(target=show_dialog, daemon=True)
    thread.start()

    try:
      result = await asyncio.wait_for(future, timeout=30.0)
    except TimeoutError:
      return False
    else:
      return result

  @staticmethod
  async def _macos_confirmation(title: str, message: str) -> bool:
    loop = asyncio.get_event_loop()

    def run_native_dialog():
      try:
          script = f"""
              display dialog "{message}" with title "{title}" buttons {{"No", "Yes"}} default button "Yes" with icon caution
              """

          result = subprocess.run(  # noqa: S603
            ["/usr/bin/osascript", "-e", script], check=False, capture_output=True, text=True, timeout=30
          )
      except subprocess.SubprocessError:
          return False
      else:
          return result.returncode == 0 and "Yes" in result.stdout

    return await loop.run_in_executor(None, run_native_dialog)
