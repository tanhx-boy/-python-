# i18n.py
import gettext
import os
import sys

# 确定语言文件所在目录
# 假设此脚本在 file_checker/ 目录下
LOCALE_DIR = os.path.join(os.path.dirname(os.path.abspath(__file__)), 'locales')

# 默认语言
DEFAULT_LANG = 'zh_CN'

# 全局翻译对象
_ = None

def setup_i18n(lang=None):
    """
    初始化国际化设置
    :param lang: 语言代码, e.g., 'zh_CN' or 'en'
    """
    global _
    if lang is None:
        lang = DEFAULT_LANG
    
    try:
        # 尝试加载指定语言的翻译
        trans = gettext.translation('base', LOCALE_DIR, languages=[lang], fallback=True)
        _ = trans.gettext
        print(f"Language set to: {lang}")
    except FileNotFoundError:
        print(f"Warning: Language file for '{lang}' not found. Falling back to '{DEFAULT_LANG}'.")
        trans = gettext.translation('base', LOCALE_DIR, languages=[DEFAULT_LANG], fallback=True)
        _ = trans.gettext
    except Exception as e:
        print(f"Error setting up i18n: {e}. Using default strings.")
        # 如果gettext失败，提供一个备用的_函数
        _ = lambda s: s

# 初始调用
setup_i18n()

# 提供一个函数来动态切换语言
def change_language(lang):
    setup_i18n(lang)
    # 在实际应用中，你需要一个机制来通知GUI更新所有文本
    # 例如，通过一个自定义事件或回调函数
    print(f"Language changed to {lang}. Please restart or refresh the UI.")
    # For simplicity, we just re-initialize. A full app would have a refresh mechanism.

