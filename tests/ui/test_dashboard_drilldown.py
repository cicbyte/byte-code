"""仪表盘任务卡钻取 UI：进行中卡 → 任务总览（scope=all + 状态预筛）。"""

import pytest

from ui import UiBase

pytestmark = [pytest.mark.ui, pytest.mark.order(5)]


@pytest.fixture(scope="module")
def ui_ds(project_env):
    pid = project_env.pid
    tid = project_env.member.post(
        f"/api/v1/projects/{pid}/tasks", {"title": "仪表盘钻取-agent执行中", "description": "", "type": "chore", "priority": 3}
    )["id"]
    project_env.agent.post(f"/api/v1/tasks/{tid}/claim")

    class NS:
        pass

    ns = NS()
    ns.pid = pid
    ns.tid = tid
    return ns


def test_dashboard_inprogress_drilldown(logged_in, frontend, ui_ds):
    page = UiBase(logged_in, frontend)
    page.goto("/dashboard/console")
    assert page.wait_ele("dashboard.inprogress-card", timeout=30)
    page.click("dashboard.inprogress-card")

    # 落地任务总览：URL 带 scope=all&status=in_progress，标题切换，我的效率卡隐藏
    import time as _t

    deadline = _t.time() + 15
    while _t.time() < deadline:
        if "scope=all" in logged_in.url and "status=in_progress" in logged_in.url:
            break
        _t.sleep(0.3)
    assert "scope=all" in logged_in.url and "status=in_progress" in logged_in.url, logged_in.url

    body = logged_in.ele("tag:body").text or ""
    assert "任务总览" in body, "scope=all 应切换为任务总览标题"
    assert "我的效率" not in body, "总览模式不应渲染我的效率统计卡"
    assert logged_in.wait.ele_displayed("text:仪表盘钻取-agent执行中", timeout=10), (
        "agent 名下的进行中任务应出现（可见范围口径）"
    )
