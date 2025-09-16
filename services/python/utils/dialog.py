import tkinter as tk
import asyncio
import threading
from tkinter import messagebox

class DialogUtil:
  @staticmethod
  async def confirmation(
    title: str = "Confirmation",
    message: str = "Are you sure?"
  ) -> bool:
        
    loop = asyncio.get_event_loop()
    future = loop.create_future()
    
    def show_dialog():
        try:
            root = tk.Tk()
            root.withdraw()
            root.lift()
            root.attributes("-topmost", True)
            root.attributes("-alpha", 0.0)  # Make invisible initially
            root.update()
            root.attributes("-alpha", 1.0)  # Make visible
            root.focus_force()
            
            result = messagebox.askyesno(title, message, parent=root)
            loop.call_soon_threadsafe(future.set_result, result)
        except Exception as e:
            loop.call_soon_threadsafe(future.set_exception, e)
        finally:
            try:
                root.destroy()
            except Exception:
                pass
    
    thread = threading.Thread(target=show_dialog, daemon=True)
    thread.start()
    
    try:
        result = await asyncio.wait_for(future, timeout=30.0)
        return result
    except asyncio.TimeoutError:
        return False