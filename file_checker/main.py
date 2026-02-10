# main.py
import sys
from gui import FileCheckerApp
import tkinter as tk

def main():
    # 在Windows上设置DPI感知，使UI在高分屏上更清晰
    if sys.platform == "win32":
        try:
            from ctypes import windll
            windll.shcore.SetProcessDpiAwareness(1)
        except Exception:
            pass # Fallback for older Windows versions

    root = tk.Tk()
    app = FileCheckerApp(root)
    root.mainloop()

if __name__ == "__main__":
    main()
