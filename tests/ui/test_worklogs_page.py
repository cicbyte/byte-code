"""工作日志 UI：手写发布、时间轴分组渲染、任务草稿生成闭环。"""

import pytest

from ui import UiBase

pytestmark = [pytest.mark.ui, pytest.mark.order(5)]


@pytest.fixture(scope="module")
def ui_wl(admin, project_env):
    """API 造数：一条手写日志 + 一个已完成（待审）任务。"""
    pid = project_env.pid
    wid = admin.post(
        f"/api/v1/projects/{pid}/worklogs",
        {"content": "UI 验收基线日志", "source": "manual"},
    )["id"]
    tid = admin.post(
        f"/api/v1/projects/{pid}/tasks",
        {"title": "草稿源任务-UI", "description": "", "type": "chore", "priority": 3},
    )["id"]
    project_env.agent.post(f"/api/v1/tasks/{tid}/claim")
    project_env.agent.post(f"/api/v1/tasks/{tid}/complete", {"artifacts": "ui-artifact-说明"})

    class NS:
        pass

    ns = NS()
    ns.pid = pid
    ns.wid = wid
    ns.tid = tid
    return ns


def test_worklog_write_and_group(logged_in, frontend, ui_wl):
    page = UiBase(logged_in, frontend)
    page.goto(f"/project/{ui_wl.pid}/worklogs")
    assert page.wait_ele(f"project-worklog.item-{ui_wl.wid}", timeout=30)

    # 日期分组头渲染（今天造的数）
    import datetime

    today = datetime.date.today().isoformat()
    assert page.wait_ele(f"project-worklog.group-{today}", timeout=10)

    # 手写发布闭环
    page.click("project-worklog.add-btn")
    assert page.wait_ele("project-worklog.editor-modal", timeout=10)
    page.input("project-worklog.content-input", "UI 手写：发布链路验收完成")
    page.click("project-worklog.submit-btn")
    assert logged_in.wait.ele_displayed("text:UI 手写：发布链路验收完成", timeout=10)

    # 编辑闭环：改内容后正文更新
    page.click(f"project-worklog.edit-{ui_wl.wid}")
    page.input("project-worklog.content-input", "UI 验收基线日志（已编辑）")
    page.click("project-worklog.submit-btn")
    assert logged_in.wait.ele_displayed("text:已编辑", timeout=10)


def test_worklog_draft_from_tasks(logged_in, frontend, ui_wl):
    page = UiBase(logged_in, frontend)
    page.goto(f"/project/{ui_wl.pid}/worklogs")
    assert page.wait_ele("project-worklog.gen-btn", timeout=30)

    # 草稿生成：默认最近 7 天 → 汇总进编辑器（含任务标题与产物）
    page.click("project-worklog.gen-btn")
    assert page.wait_ele("project-worklog.gen-modal", timeout=10)
    page.click("project-worklog.gen-confirm")
    assert page.wait_ele("project-worklog.editor-modal", timeout=10)
    draft = page._input_node("project-worklog.content-input").attr("value") or ""
    assert "草稿源任务-UI" in draft
    page.click("project-worklog.submit-btn")
    assert logged_in.wait.ele_displayed("text:草稿源任务-UI", timeout=10)
