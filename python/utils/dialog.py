import tkinter as tk
import asyncio
import threading
import platform
import subprocess
import os
from tkinter import messagebox

class DialogUtil:
  @staticmethod
  async def confirmation(
    title: str = "Confirmation",
    message: str = "Are you sure?"
  ) -> bool:
        
    if platform.system() == "Darwin":
        try:
            return await DialogUtil._macos_confirmation(title, message)
        except Exception:
            pass
    
    loop = asyncio.get_event_loop()
    future = loop.create_future()
    
    def show_dialog():
        try:
            root = tk.Tk()
            root.withdraw()
            
            # macOS-specific window handling
            if platform.system() == "Darwin":
                # Bring Python app to foreground on macOS
                try:
                    subprocess.run([
                        'osascript', '-e',
                        'tell application "System Events" to set frontmost of first process whose unix id is {} to true'.format(os.getpid())
                    ], check=False, capture_output=True)
                except Exception:
                    pass
                
                # Configure for macOS
                root.lift()
                root.call('wm', 'attributes', '.', '-topmost', True)
                root.after_idle(root.attributes, '-topmost', False)
            else:
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
                root.destroy() # type: ignore
            except Exception:
                pass
    
    thread = threading.Thread(target=show_dialog, daemon=True)
    thread.start()
    
    try:
        result = await asyncio.wait_for(future, timeout=30.0)
        return result
    except asyncio.TimeoutError:
        return False

  @staticmethod
  async def _macos_confirmation(title: str, message: str) -> bool:
    loop = asyncio.get_event_loop()
    
    def run_native_dialog():
        try:
            script = f'''
            display dialog "{message}" with title "{title}" buttons {{"No", "Yes"}} default button "Yes" with icon caution
            '''
            
            result = subprocess.run([
                'osascript', '-e', script
            ], capture_output=True, text=True, timeout=30)
            
            return result.returncode == 0 and "Yes" in result.stdout
        except Exception:
            return False
    
    return await loop.run_in_executor(None, run_native_dialog)