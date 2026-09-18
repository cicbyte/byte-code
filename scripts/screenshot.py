#!/usr/bin/env python3
"""README 配图截图脚本。

对运行中的 ByteCode（默认 http://127.0.0.1:8001）逐页截图到 docs/images/。
前置：后端(8000) 与前端(8001) 均在运行；pip install playwright 且已
python -m playwright install chromium。

用法：
    python scripts/screenshot.py [--base http://127.0.0.1:8001] [--out docs/images] \
        [--project 1] [--user admin] [--password <密码>]

配图对应 README「界面一览」：任务详情 / 审核台 / 项目记忆 / 测试执行。
dashboard / board / knowledge / members 为备用素材页。
"""
import argparse
import pathlib
import sys

from playwright.sync_api import sync_playwright

# (输出文件名, 路径后缀, 拍后动作)——动作用于构造更饱满的画面（开抽屉/展开组）
# 路径中的 {p} 是项目 id 占位


def pages(project: int) -> list:
    return [
        ("dashboard.png", "/dashboard/console", None),
        ("board.png", f"/project/{project}/board", None),
        ("task-detail.png", f"/project/{project}/tasks", "open_task"),
        ("reviews.png", f"/project/{project}/reviews", None),
        ("knowledge.png", f"/project/{project}/knowledge", None),
        ("memories.png", f"/project/{project}/memories", None),
        ("members.png", f"/project/{project}/members", None),
        ("test-runs.png", f"/project/{project}/test-runs", "wait_chart"),
    ]


def login(page, base, user, password):
    page.goto(f"{base}/login", wait_until="domcontentloaded")
    page.wait_for_timeout(1500)
    # 登录页：placeholder 定位（Naive UI input）
    page.locator('input[placeholder="请输入用户名"]').fill(user)
    page.locator('input[placeholder="请输入密码"]').fill(password)
    page.get_by_role("button", name="登录").click()
    page.wait_for_timeout(2500)


def open_task_drawer(page):
    """任务列表打开第一条任务详情抽屉，展示锚点/清单/评论区富界面"""
    btns = page.locator("tbody button")
    if btns.count() > 0:
        btns.first.click()
        page.wait_for_timeout(1500)


def wait_chart(page):
    """执行记录页：等 echarts 趋势图动画收尾再拍"""
    page.wait_for_timeout(1500)


def main():
    ap = argparse.ArgumentParser()
    ap.add_argument("--base", default="http://127.0.0.1:8001")
    ap.add_argument("--out", default="docs/images")
    ap.add_argument("--project", type=int, default=1, help="截图的项目 id")
    ap.add_argument("--user", default="admin")
    ap.add_argument("--password", required=True, help="登录密码（不走默认值，避免明文密码入库）")
    args = ap.parse_args()

    out = pathlib.Path(args.out)
    out.mkdir(parents=True, exist_ok=True)

    with sync_playwright() as p:
        browser = p.chromium.launch()
        ctx = browser.new_context(
            viewport={"width": 1440, "height": 900},
            device_scale_factor=2,  # 2x 高清，README 里缩小显示更锐利
            locale="zh-CN",
        )
        page = ctx.new_page()

        login(page, args.base, args.user, args.password)
        if "/login" in page.url:
            print("登录失败：请核对账号密码与登录页选择器", file=sys.stderr)
            sys.exit(1)

        for name, path, action in pages(args.project):
            page.goto(f"{args.base}{path}", wait_until="domcontentloaded")
            page.wait_for_timeout(2000)  # 等数据渲染稳定
            if action == "open_task":
                open_task_drawer(page)
            elif action == "wait_chart":
                wait_chart(page)
            target = out / name
            page.screenshot(path=str(target), full_page=False)
            print(f"✓ {target}")

        browser.close()


if __name__ == "__main__":
    main()
