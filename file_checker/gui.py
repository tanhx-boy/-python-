# gui.py
import tkinter as tk
from tkinter import ttk, filedialog, messagebox, Listbox, Scrollbar
import threading
import queue
import os
import sys
from i18n import _
import algorithms

# --- 主题和样式 ---
class ThemedStyle:
    LIGHT = {
        "bg": "#f0f0f0",
        "fg": "#000000",
        "select_bg": "#0078d7",
        "select_fg": "#ffffff",
        "tree_bg": "#ffffff",
        "tree_fg": "#000000",
        "tree_select_bg": "#cce8ff",
        "tree_select_fg": "#000000",
        "button_bg": "#e1e1e1",
        "button_fg": "#000000",
        "frame_bg": "#f0f0f0"
    }
    DARK = {
        "bg": "#2b2b2b",
        "fg": "#ffffff",
        "select_bg": "#005a9e",
        "select_fg": "#ffffff",
        "tree_bg": "#3c3c3c",
        "tree_fg": "#ffffff",
        "tree_select_bg": "#0078d7",
        "tree_select_fg": "#ffffff",
        "button_bg": "#454545",
        "button_fg": "#ffffff",
        "frame_bg": "#2b2b2b"
    }

class FileCheckerApp:
    def __init__(self, root):
        self.root = root
        self.files_to_process = []
        self.results = {}  # {filepath: {alg1: val1, alg2: val2}}
        self.is_calculating = False
        self.theme = ThemedStyle.LIGHT
        self.transparency = 1.0

        self.setup_ui()
        self.apply_theme()
        self.setup_drag_drop()

    def setup_ui(self):
        self.root.title(_("File Integrity Checker"))
        self.root.geometry("1000x700")
        
        # --- 菜单栏 ---
        menubar = tk.Menu(self.root)
        self.root.config(menu=menubar)
        
        file_menu = tk.Menu(menubar, tearoff=0)
        menubar.add_cascade(label=_("File"), menu=file_menu)
        file_menu.add_command(label=_("Add Files"), command=self.add_files)
        file_menu.add_command(label=_("Export Results"), command=self.export_results)
        file_menu.add_separator()
        file_menu.add_command(label=_("Exit"), command=self.root.quit)
        
        help_menu = tk.Menu(menubar, tearoff=0)
        menubar.add_cascade(label=_("Help"), menu=help_menu)
        help_menu.add_command(label=_("About"), command=self.show_about)

        # --- 主框架 ---
        main_frame = ttk.Frame(self.root, padding="10")
        main_frame.pack(fill=tk.BOTH, expand=True)

        # --- 左侧：文件列表和控制 ---
        left_frame = ttk.Frame(main_frame, padding="5")
        left_frame.pack(side=tk.LEFT, fill=tk.BOTH, expand=True)

        # 文件列表 (Treeview)
        columns = ("file", "status")
        self.tree = ttk.Treeview(left_frame, columns=columns, show="headings")
        self.tree.heading("file", text=_("File Name"))
        self.tree.heading("status", text=_("Status"))
        self.tree.column("file", width=300)
        self.tree.column("status", width=100)
        
        tree_scrollbar = Scrollbar(left_frame, orient=tk.VERTICAL, command=self.tree.yview)
        self.tree.configure(yscrollcommand=tree_scrollbar.set)
        
        self.tree.pack(side=tk.LEFT, fill=tk.BOTH, expand=True)
        tree_scrollbar.pack(side=tk.RIGHT, fill=tk.Y)

        # 算法选择
        algo_frame = ttk.LabelFrame(left_frame, text=_("Select Algorithm(s)"), padding="5")
        algo_frame.pack(fill=tk.X, pady=10)
        
        self.algo_vars = {}
        algo_list = list(algorithms.ALGORITHMS.keys())
        algo_list.sort()
        
        # 创建一个带滚动条的框架来放置复选框
        canvas = tk.Canvas(algo_frame)
        scrollbar = ttk.Scrollbar(algo_frame, orient="vertical", command=canvas.yview)
        scrollable_frame = ttk.Frame(canvas)

        scrollable_frame.bind(
            "<Configure>",
            lambda e: canvas.configure(scrollregion=canvas.bbox("all"))
        )
        canvas.create_window((0, 0), window=scrollable_frame, anchor="nw")
        canvas.configure(yscrollcommand=scrollbar.set)
        
        for algo in algo_list:
            var = tk.BooleanVar()
            cb = ttk.Checkbutton(scrollable_frame, text=algo, variable=var)
            cb.pack(anchor=tk.W, padx=5)
            self.algo_vars[algo] = var
        
        canvas.pack(side="left", fill="both", expand=True)
        scrollbar.pack(side="right", fill="y")

        # 控制按钮
        control_frame = ttk.Frame(left_frame)
        control_frame.pack(fill=tk.X, pady=5)
        
        self.calc_button = ttk.Button(control_frame, text=_("Calculate"), command=self.start_calculation)
        self.calc_button.pack(side=tk.LEFT, expand=True, fill=tk.X, padx=2)
        
        self.clear_button = ttk.Button(control_frame, text=_("Clear All"), command=self.clear_all)
        self.clear_button.pack(side=tk.LEFT, expand=True, fill=tk.X, padx=2)

        # --- 右侧：结果显示 ---
        right_frame = ttk.Frame(main_frame, padding="5")
        right_frame.pack(side=tk.RIGHT, fill=tk.BOTH, expand=True)
        
        self.result_text = tk.Text(right_frame, wrap=tk.WORD, state=tk.DISABLED, height=10)
        self.result_text.pack(fill=tk.BOTH, expand=True)
        
        result_scrollbar = Scrollbar(right_frame, orient=tk.VERTICAL, command=self.result_text.yview)
        result_scrollbar.pack(side=tk.RIGHT, fill=tk.Y)
        self.result_text.configure(yscrollcommand=result_scrollbar.set)

        # --- 底部：设置 ---
        settings_frame = ttk.Frame(self.root, padding="5")
        settings_frame.pack(side=tk.BOTTOM, fill=tk.X)
        
        ttk.Label(settings_frame, text=_("Theme")).pack(side=tk.LEFT, padx=5)
        self.theme_var = tk.StringVar(value="Light")
        theme_combo = ttk.Combobox(settings_frame, textvariable=self.theme_var, values=[_("Light"), _("Dark")], state="readonly")
        theme_combo.pack(side=tk.LEFT, padx=5)
        theme_combo.bind("<<ComboboxSelected>>", self.on_theme_change)
        
        ttk.Label(settings_frame, text=_("Transparency")).pack(side=tk.LEFT, padx=5)
        self.trans_slider = ttk.Scale(settings_frame, from_=0.5, to=1.0, orient=tk.HORIZONTAL)
        self.trans_slider.set(1.0)
        self.trans_slider.pack(side=tk.LEFT, padx=5, fill=tk.X, expand=True)
        self.trans_slider.bind("<ButtonRelease-1>", self.on_transparency_change)
        
        ttk.Label(settings_frame, text=_("Language")).pack(side=tk.LEFT, padx=5)
        self.lang_var = tk.StringVar(value="简体中文")
        lang_combo = ttk.Combobox(settings_frame, textvariable=self.lang_var, values=[_("Simplified Chinese"), _("English")], state="readonly")
        lang_combo.pack(side=tk.LEFT, padx=5)
        lang_combo.bind("<<ComboboxSelected>>", self.on_lang_change)

        # --- 状态队列和后台处理 ---
        self.status_queue = queue.Queue()
        self.root.after(100, self.process_queue)

    def setup_drag_drop(self):
        # Windows
        try:
            import ctypes
            ctypes.windll.user32.DragAcceptFiles(self.root.winfo_id(), 1)
            self.root.bind('<<DropFiles>>', self.on_drop_files)
        except Exception as e:
            print(f"Drag and drop setup for Windows failed: {e}")

        # Linux/Mac (using Tkinter DnD wrapper would be ideal, but is third-party)
        # As a fallback, we'll just rely on the "Add Files" button.
        # A more advanced solution would involve binding to <Button-1>, <B1-Motion>, etc.
        # but that's complex without a library. For now, we'll just log it.
        print("Drag and drop on Linux/Mac may require a third-party library like tkinterdnd2.")
        print("Using the 'Add Files' button is the recommended method on non-Windows systems.")

    def on_drop_files(self, event):
        # The event.data is a string of file paths, possibly with curly braces
        # and null characters. This is a bit tricky to parse reliably.
        # A simpler approach for this example is to just use filedialog.
        # But let's try to parse it.
        filepaths = self.root.tk.splitlist(event.data)
        for f in filepaths:
            if os.path.isfile(f):
                self.add_file_to_list(f)
    
    def add_files(self):
        filepaths = filedialog.askopenfilenames(title=_("Select Files"))
        if not filepaths:
            return
        for f in filepaths:
            self.add_file_to_list(f)

    def add_file_to_list(self, filepath):
        if filepath not in self.files_to_process:
            self.files_to_process.append(filepath)
            filename = os.path.basename(filepath)
            self.tree.insert("", tk.END, values=(filename, _("Ready")))

    def clear_all(self):
        self.files_to_process.clear()
        self.results.clear()
        for item in self.tree.get_children():
            self.tree.delete(item)
        self.update_result_text("")

    def start_calculation(self):
        if self.is_calculating:
            messagebox.showwarning(_("Warning"), _("Calculation is already in progress."))
            return
            
        selected_algos = [algo for algo, var in self.algo_vars.items() if var.get()]
        if not selected_algos:
            messagebox.showerror(_("Error"), _("Please select at least one algorithm."))
            return
            
        if not self.files_to_process:
            messagebox.showerror(_("Error"), _("No files to process."))
            return

        self.is_calculating = True
        self.calc_button.config(state=tk.DISABLED)
        self.clear_button.config(state=tk.DISABLED)
        
        # 启动后台线程
        thread = threading.Thread(target=self.calculation_worker, args=(self.files_to_process, selected_algos), daemon=True)
        thread.start()

    def calculation_worker(self, files, algos):
        """在后台线程中运行的计算任务"""
        for i, filepath in enumerate(files):
            filename = os.path.basename(filepath)
            # 更新状态为 "Calculating..."
            self.status_queue.put(("update_status", filename, _("Calculating...")))
            
            file_results = {}
            for algo in algos:
                hash_value = algorithms.calculate_hash(filepath, algo)
                file_results[algo] = hash_value
            
            self.results[filepath] = file_results
            
            # 更新状态为 "Complete"
            self.status_queue.put(("update_status", filename, _("Complete")))
            
            # 实时更新结果显示
            self.status_queue.put(("update_result_text", filepath, file_results))

        # 计算完成
        self.status_queue.put(("calculation_finished",))

    def process_queue(self):
        """从后台线程处理消息队列"""
        try:
            while True:
                msg = self.status_queue.get_nowait()
                msg_type = msg[0]
                
                if msg_type == "update_status":
                    _, filename, status = msg
                    for item in self.tree.get_children():
                        if self.tree.item(item, "values")[0] == filename:
                            self.tree.item(item, values=(filename, status))
                            break
                
                elif msg_type == "update_result_text":
                    _, filepath, file_results = msg
                    self.update_result_text_for_file(filepath, file_results)

                elif msg_type == "calculation_finished":
                    self.is_calculating = False
                    self.calc_button.config(state=tk.NORMAL)
                    self.clear_button.config(state=tk.NORMAL)
                    messagebox.showinfo(_("Info"), _("Calculation finished."))

        except queue.Empty:
            pass
        finally:
            self.root.after(100, self.process_queue)
            
    def update_result_text_for_file(self, filepath, file_results):
        self.result_text.config(state=tk.NORMAL)
        filename = os.path.basename(filepath)
        
        # 清空旧的该文件结果
        lines = self.result_text.get("1.0", tk.END).strip().split('\n\n')
        new_lines = [block for block in lines if not block.startswith(f"{filename}--")]
        
        # 添加新结果
        result_block = f"{filename}--"
        for algo, value in file_results.items():
            result_block += f" {algo}:{value}"
        
        new_lines.append(result_block)
        self.result_text.delete("1.0", tk.END)
        self.result_text.insert(tk.END, "\n\n".join(new_lines))
        self.result_text.config(state=tk.DISABLED)

    def update_result_text(self, text):
        self.result_text.config(state=tk.NORMAL)
        self.result_text.delete("1.0", tk.END)
        self.result_text.insert(tk.END, text)
        self.result_text.config(state=tk.DISABLED)

    def export_results(self):
        if not self.results:
            messagebox.showwarning(_("Warning"), _("No results to export."))
            return

        filepath = filedialog.asksaveasfilename(
            title=_("Export Results"),
            defaultextension=".txt",
            filetypes=[(_("Text Files"), "*.txt"), (_("All Files"), "*.*")]
        )
        if not filepath:
            return
            
        try:
            with open(filepath, 'w', encoding='utf-8') as f:
                for filepath, file_results in self.results.items():
                    filename = os.path.basename(filepath)
                    # 格式: 文件名--算法1:值1 算法2:值2
                    line = f"{filename}--"
                    for algo, value in file_results.items():
                        line += f" {algo}:{value}"
                    f.write(line + '\n')
            messagebox.showinfo(_("Success"), _("Export successful"))
        except IOError as e:
            messagebox.showerror(_("Error"), f"{_('Export failed')}: {e}")

    def show_about(self):
        messagebox.showinfo(
            _("About File Checker"),
            _("File Integrity Checker\n\nA modern file integrity checker built with Python.\n\nSupports multiple hashing algorithms.")
        )

    def apply_theme(self):
        style = ttk.Style()
        style.theme_use('clam') # A good base theme
        
        # 配置颜色
        bg = self.theme["bg"]
        fg = self.theme["fg"]
        select_bg = self.theme["select_bg"]
        select_fg = self.theme["select_fg"]
        
        style.configure(".", background=bg, foreground=fg, font=("Segoe UI", 10))
        style.configure("TFrame", background=bg)
        style.configure("TLabel", background=bg, foreground=fg)
        style.configure("TButton", background=self.theme["button_bg"], foreground=self.theme["button_fg"])
        style.map("TButton", background=[("active", select_bg)])
        style.configure("TCheckbutton", background=bg, foreground=fg)
        style.configure("Treeview", background=self.theme["tree_bg"], foreground=self.theme["tree_fg"], fieldbackground=self.theme["tree_bg"])
        style.map("Treeview", background=[("selected", self.theme["tree_select_bg"])], foreground=[("selected", self.theme["tree_select_fg"])])
        style.configure("Treeview.Heading", background=self.theme["button_bg"], foreground=self.theme["button_fg"])
        style.configure("TLabelframe", background=bg, bordercolor=fg)
        style.configure("TLabelframe.Label", background=bg, foreground=fg, font=("Segoe UI", 10, "bold"))
        style.configure("TCombobox", fieldbackground=self.theme["button_bg"], background=self.theme["button_bg"], foreground=self.theme["button_fg"])
        
        # Text widget needs separate configuration
        self.result_text.config(bg=self.theme["tree_bg"], fg=self.theme["tree_fg"], insertbackground=self.theme["tree_fg"])

    def on_theme_change(self, event=None):
        selected = self.theme_var.get()
        if selected == _("Dark"):
            self.theme = ThemedStyle.DARK
        else:
            self.theme = ThemedStyle.LIGHT
        self.apply_theme()
        self.set_window_transparency() # Re-apply transparency with new colors

    def on_transparency_change(self, event=None):
        self.transparency = self.trans_slider.get()
        self.set_window_transparency()

    def set_window_transparency(self):
        # 仅在Windows上有效
        if sys.platform == "win32":
            try:
                # 0.5 (50%) to 1.0 (100%)
                alpha = int(self.transparency * 255)
                # GWL_EXSTYLE = -20, WS_EX_LAYERED = 0x80000
                # LWA_ALPHA = 2
                hwnd = self.root.winfo_id()
                ctypes.windll.user32.SetWindowLongW(hwnd, -20, ctypes.windll.user32.GetWindowLongW(hwnd, -20) | 0x80000)
                ctypes.windll.user32.SetLayeredWindowAttributes(hwnd, 0, alpha, 2)
            except Exception as e:
                print(f"Failed to set transparency: {e}")
    
    def on_lang_change(self, event=None):
        selected_lang_name = self.lang_var.get()
        if selected_lang_name == _("Simplified Chinese"):
            lang_code = "zh_CN"
        else:
            lang_code = "en"
        
        # 重新初始化i18n
        from i18n import setup_i18n, change_language
        change_language(lang_code)
        
        # 刷新整个UI的文本
        # 这是一个简化的实现，一个完整的应用会有一个更优雅的刷新机制
        messagebox.showinfo(_("Info"), _("Language will be applied after restarting the application."))
        # 实际操作中，可以写一个函数来递归更新所有组件的文本
        # self.refresh_ui_text()
        # self.root.update()


if __name__ == '__main__':
    root = tk.Tk()
    app = FileCheckerApp(root)
    root.mainloop()

