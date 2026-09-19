"""UI 层 fixtures：browser（DrissionPage）+ 登录态 + 世界数据（复用 API 层状态链）。

世界由 API 层 fixture 造（project_env / task_chain / review_env 皆 session 级），
UI 只验证渲染与交互——API 造数、UI 验形，两层共享同一个临时后端。
"""

import sys
from pathlib import Path

import pytest

sys.path.insert(0, str(Path(__file__).resolve().parents[1] / "helpers"))

from apiclient import Client  # noqa: E402
from ui import LoginPage  # noqa: E402


@pytest.fixture(scope="module")
def browser(frontend, pytestconfig):
    from DrissionPage import ChromiumOptions, ChromiumPage

    co = ChromiumOptions()
    co.auto_port()  # 独立端口，不抢占日常浏览器实例
    if not pytestconfig.getoption("--ui-headed"):
        co.headless()
    co.set_argument("--window-size", "1440,900")
    page = ChromiumPage(co)
    yield page
    page.quit()


@pytest.fixture(scope="module")
def logged_in(browser, frontend, backend, data):
    """登录态页面：真实走一遍登录表单（本身就是第一条 UI 验证路径）。"""
    lp = LoginPage(browser, frontend)
    lp.goto(data["ui"]["routes"]["login"])
    users = data["users"]
    lp.login(users["admin"]["username"], users["admin"]["password"])
    lp.expect_login_ok()
    return browser


@pytest.fixture(scope="module")
def ui_world(admin, backend, project_env, review_env):
    """看板/任务页数据世界（复用 API 层已建状态）+ 补一组 blocked 任务。

    返回 pid 与若干任务 id（open/in_progress/done 已由 API 层链路产生）。
    """

    class NS:
        pass

    ns = NS()
    ns.pid = project_env.pid
    ns.done_task = review_env.task_id
    # blocked 态（agent 上报阻塞）与 open 态各补一件，看板列更饱满
    ns.blocked_task = admin.post(
        f"/api/v1/projects/{ns.pid}/tasks", {"title": "集成测试-被阻塞", "type": "feature", "priority": 3}
    )["id"]
    project_env.agent.post(f"/api/v1/tasks/{ns.blocked_task}/claim")
    project_env.agent.post(f"/api/v1/tasks/{ns.blocked_task}/block", {"reason": "等外部依赖"})
    return ns


@pytest.fixture(scope="module")
def ui_runs(project_env, data):
    """执行记录页数据：按 testruns.yaml 三连报（复用同一构造器语义）。"""
    cfg = data["testruns"]
    run_ids = []
    for spec in cfg["runs"]:
        cases = [
            {
                "externalKey": c["key"],
                "title": c["key"].rsplit("::", 1)[-1],
                "status": c["status"],
                "durationMs": c["duration"],
                "message": c.get("message", ""),
            }
            for c in spec["cases"]
        ]
        rid = project_env.agent.post(
            f"/api/v1/projects/{project_env.pid}/test-runs",
            {"source": "pytest", "branch": cfg["branch"], "env": cfg["env"], "cases": cases},
        )["id"]
        run_ids.append(rid)
    return run_ids
