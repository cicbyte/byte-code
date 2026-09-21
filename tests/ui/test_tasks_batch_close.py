"""任务列表批量关闭 UI：勾选已完成行 → 批量关闭 → 状态转 closed。"""

import time

import pytest

from ui import UiBase

pytestmark = [pytest.mark.ui, pytest.mark.order(5)]


def _make_done(project_env, admin, pid, title):
    tid = project_env.member.post(
        f"/api/v1/projects/{pid}/tasks", {"title": title, "description": "", "type": "chore", "priority": 3}
    )["id"]
    project_env.agent.post(f"/api/v1/tasks/{tid}/claim")
    project_env.agent.post(f"/api/v1/tasks/{tid}/complete", {"artifacts": ""})
    admin.post("/api/v1/reviews/batch", {"ids": [tid], "status": "approved"})
    return tid


@pytest.fixture(scope="module")
def ui_bc(project_env, admin):
    pid = project_env.pid
    ids = [_make_done(project_env, admin, pid, f"UI 批量关闭-{i}") for i in range(2)]

    class NS:
        pass

    ns = NS()
    ns.pid = pid
    ns.ids = ids
    return ns


def _click_dialog_btn(page, text, timeout=10):
    """n-dialog 动作区按 strip 文本精确过滤（UiBase.eles 会包锚点，这里要原生 DP 页面）。"""
    deadline = time.time() + timeout
    while time.time() < deadline:
        for b in page.page.eles("css:.n-dialog__action .n-button"):
            if (b.text or "").strip() == text:
                b.click()
                return True
        time.sleep(0.3)
    return False


def test_tasks_batch_close(logged_in, frontend, ui_bc):
    page = UiBase(logged_in, frontend)
    page.goto(f"/project/{ui_bc.pid}/tasks")
    assert page.wait_ele(f"tasks.row-{ui_bc.ids[0]}", timeout=30)

    # 勾选两个已完成任务 → 批量条出现
    page.click(f"tasks.check-{ui_bc.ids[0]}")
    page.click(f"tasks.check-{ui_bc.ids[1]}")
    assert page.wait_ele("tasks.batch-close-btn", timeout=10)
    btn_text = page.text_of("tasks.batch-close-btn")
    assert "2" in btn_text, f"批量条计数不符：{btn_text!r}"

    # 确认弹窗 → 关闭
    page.click("tasks.batch-close-btn")
    assert _click_dialog_btn(page, "关闭")
    assert logged_in.wait.ele_displayed("text:已关闭 2 个任务", timeout=10)

    # 状态落 closed：行内状态 tag 变「已关闭」，勾选框随 done 消失
    assert logged_in.wait.ele_displayed("text:已关闭", timeout=10)
    deadline = time.time() + 10
    while time.time() < deadline:
        gone = not page.eles(f"@data-test-id=tasks.check-{ui_bc.ids[0]}")
        if gone:
            break
        time.sleep(0.3)
    assert gone, "关闭后 done 行勾选框应消失"
