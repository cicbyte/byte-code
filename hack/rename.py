"""将项目中所有 byte-code 引用替换为 byte-code"""

import os
import re

OLD = "byte-code"
NEW = "byte-code"
ROOT = os.path.dirname(os.path.dirname(os.path.abspath(__file__)))

# 跳过的目录
SKIP_DIRS = {".git", "node_modules", "vendor", ".idea", "__pycache__"}


def should_process(path: str) -> bool:
    parts = path.replace("\\", "/").split("/")
    return not any(p in SKIP_DIRS for p in parts)


def replace_in_file(filepath: str) -> bool:
    try:
        with open(filepath, "r", encoding="utf-8") as f:
            content = f.read()
    except (UnicodeDecodeError, PermissionError):
        return False

    if OLD not in content:
        return False

    new_content = content.replace(OLD, NEW)
    with open(filepath, "w", encoding="utf-8") as f:
        f.write(new_content)
    return True


def main():
    changed = 0
    for dirpath, dirnames, filenames in os.walk(ROOT):
        dirnames[:] = [d for d in dirnames if should_process(os.path.join(dirpath, d))]
        for name in filenames:
            filepath = os.path.join(dirpath, name)
            if not should_process(filepath):
                continue
            if replace_in_file(filepath):
                rel = os.path.relpath(filepath, ROOT)
                print(f"  {rel}")
                changed += 1

    print(f"\n共替换 {changed} 个文件: '{OLD}' -> '{NEW}'")


if __name__ == "__main__":
    main()
