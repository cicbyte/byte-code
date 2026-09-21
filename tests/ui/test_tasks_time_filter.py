"""任务列表日期筛选 UI：daterange 选今天 → 只剩今日完成任务。"""

import time

import pytest

from ui import UiBase

pytestmark = [pytest.mark.ui, pytest.mark.order(5)]


@pytest.fixture(scope="module")
def ui_tf(project_env, admin):
    pid = project_env.pid
    tid = project_env.member.post(
        f"/api/v1/projects/{pid}/tasks", {"title": "UI日期-今日完成", "description": "", "type": "chore", "priority": 3}
    )["id"]
    project_env.agent.post(f"/api/v1/tasks/{tid}/claim")
    project_env.agent.post(f"/api/v1/tasks/{tid}/complete", {"artifacts": ""})
    admin.post("/api/v1/reviews/batch", {"ids": [tid], "status": "approved"})
    open_id = project_env.member.post(
        f"/api/v1/projects/{pid}/tasks", {"title": "UI日期-未完成", "description": "", "type": "chore", "priority": 3}
    )["id"]

    class NS:
        pass

    ns = NS()
    ns.pid = pid
    ns.done_id = tid
    ns.open_id = open_id
    return ns


def _row_gone(page, rid, timeout=8):
    deadline = time.time() + timeout
    while time.time() < deadline:
        if not page.eles(f"@data-test-id=tasks.row-{rid}"):
            return True
        time.sleep(0.3)
    return False


def test_tasks_date_filter(logged_in, frontend, ui_tf):
    page = UiBase(logged_in, frontend)
    page.goto(f"/project/{ui_tf.pid}/tasks")
    assert page.wait_ele(f"tasks.row-{ui_tf.open_id}", timeout=30)

    # 打开日期面板：daterange 点两次「今天」（起=止=今天）
    page.click("tasks.date-range")
    deadline = time.time() + 10
    while time.time() < deadline:
        cells = logged_in.eles("css:.n-date-panel-date--current")
        if cells:
            break
        time.sleep(0.3)
    cells[0].click()
    time.sleep(0.4)
    # 起点选中后面板保持打开，重新定位当前日再点一次作终点
    cells2 = logged_in.eles("css:.n-date-panel-date--current")
    cells2[0].click()

    # 触发重查后：未完成任务行消失（completed_at 空），今日完成任务行保留
    assert _row_gone(page, ui_tf.open_id), "日期筛选后未完成任务行应消失"
    assert page.wait_ele(f"tasks.row-{ui_tf.done_id}", timeout=10)
    # 行内展示完成时刻
    body = logged_in.ele("tag:body").text or ""
    assert "完成 " in body, "更新时间列应展示完成时刻次行"
