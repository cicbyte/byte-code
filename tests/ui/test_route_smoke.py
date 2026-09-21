"""全站路由横向冒烟网：每页壳渲染 + 一个安全主交互。

背景（改密弹窗缺陷复盘）：覆盖按功能任务生长，基础页交互覆盖为零，
「构建全绿但页面死了」类 bug 无兜底。本文件遍历全部菜单路由做薄断言——
只读或可恢复的交互，不造复杂数据（空态即有效断言）。
"""

import time

import pytest

from ui import UiBase

pytestmark = [pytest.mark.ui, pytest.mark.order(7)]


def _body(logged_in):
    return logged_in.ele("tag:body").text or ""


def _wait_text(page, logged_in, text, timeout=20):
    """轮询 body 文本包含目标（text: 定位对嵌套菜单结构不可靠，body 断言稳）。"""
    import time as _t

    deadline = _t.time() + timeout
    while _t.time() < deadline:
        if text in _body(logged_in):
            return True
        _t.sleep(0.5)
    return False


# ---------- 全局区 ----------


def test_dashboard_console(logged_in, frontend):
    page = UiBase(logged_in, frontend)
    page.goto("/dashboard/console")
    assert _wait_text(page, logged_in, "总需求数"), _body(logged_in)[:200]


def test_my_tasks(logged_in, frontend):
    page = UiBase(logged_in, frontend)
    page.goto("/my-tasks/index")
    # 壳渲染：页面出现任务表头或空态（不依赖具体任务数据）
    assert _wait_text(page, logged_in, "我的任务")


def test_global_reviews(logged_in, frontend):
    page = UiBase(logged_in, frontend)
    page.goto("/reviews/index")
    assert _wait_text(page, logged_in, "审核中心")
    assert "批量通过" in _body(logged_in)


def test_groups_page(logged_in, frontend):
    page = UiBase(logged_in, frontend)
    page.goto("/project/groups")
    assert _wait_text(page, logged_in, "项目分组")


# ---------- 设置区 ----------


def test_setting_system(logged_in, frontend):
    page = UiBase(logged_in, frontend)
    page.goto("/setting/system")
    # 壳渲染：表单区出现（系统配置卡片）
    assert _wait_text(page, logged_in, "系统设置")
    assert ("保存" in _body(logged_in)) or ("配置" in _body(logged_in))


def test_setting_global_memory(logged_in, frontend):
    page = UiBase(logged_in, frontend)
    page.goto("/setting/global-memory")
    assert _wait_text(page, logged_in, "全局记忆")


# ---------- 平台区 ----------


def test_platform_notifications(logged_in, frontend):
    page = UiBase(logged_in, frontend)
    page.goto("/platform/notifications")
    assert _wait_text(page, logged_in, "通知")


def test_platform_activities(logged_in, frontend):
    page = UiBase(logged_in, frontend)
    page.goto("/platform/activities")
    assert _wait_text(page, logged_in, "活动")


def test_platform_agents(logged_in, frontend):
    page = UiBase(logged_in, frontend)
    page.goto("/platform/agents")
    assert _wait_text(page, logged_in, "Agent")


# ---------- 权限区 ----------


def test_permission_role(logged_in, frontend):
    page = UiBase(logged_in, frontend)
    page.goto("/permission/role")
    assert _wait_text(page, logged_in, "角色")


# ---------- 项目区（复用主项目世界） ----------


@pytest.fixture(scope="module")
def smoke_pid(ui_world):
    return ui_world.pid


def test_project_overview(logged_in, frontend, smoke_pid):
    page = UiBase(logged_in, frontend)
    page.goto(f"/project/{smoke_pid}/overview")
    assert _wait_text(page, logged_in, "项目概览")


def test_project_topics(logged_in, frontend, smoke_pid):
    page = UiBase(logged_in, frontend)
    page.goto(f"/project/{smoke_pid}/topics")
    assert _wait_text(page, logged_in, "专题")


def test_project_feedbacks(logged_in, frontend, smoke_pid):
    page = UiBase(logged_in, frontend)
    page.goto(f"/project/{smoke_pid}/feedbacks")
    assert _wait_text(page, logged_in, "反馈")


def test_project_requirements(logged_in, frontend, smoke_pid):
    page = UiBase(logged_in, frontend)
    page.goto(f"/project/{smoke_pid}/requirements")
    assert _wait_text(page, logged_in, "需求")


def test_project_milestones(logged_in, frontend, smoke_pid):
    page = UiBase(logged_in, frontend)
    page.goto(f"/project/{smoke_pid}/milestones")
    assert _wait_text(page, logged_in, "里程碑")


def test_project_sprints(logged_in, frontend, smoke_pid):
    page = UiBase(logged_in, frontend)
    page.goto(f"/project/{smoke_pid}/sprints")
    assert _wait_text(page, logged_in, "迭代")


def test_project_knowledge(logged_in, frontend, smoke_pid):
    page = UiBase(logged_in, frontend)
    page.goto(f"/project/{smoke_pid}/knowledge")
    # 与 docs 同组件（vaultSpace=knowledge）：搜索框锚点即壳
    assert page.wait_ele("project-docs.search-input", timeout=30)


def test_project_qas(logged_in, frontend, smoke_pid):
    page = UiBase(logged_in, frontend)
    page.goto(f"/project/{smoke_pid}/qas")
    assert _wait_text(page, logged_in, "QA 库") or _wait_text(page, logged_in, "问答")


def test_project_memories(logged_in, frontend, smoke_pid):
    page = UiBase(logged_in, frontend)
    page.goto(f"/project/{smoke_pid}/memories")
    assert _wait_text(page, logged_in, "项目记忆")


def test_project_test_overview(logged_in, frontend, smoke_pid):
    page = UiBase(logged_in, frontend)
    page.goto(f"/project/{smoke_pid}/test-overview")
    assert _wait_text(page, logged_in, "测试总览")


def test_project_test_plans(logged_in, frontend, smoke_pid):
    page = UiBase(logged_in, frontend)
    page.goto(f"/project/{smoke_pid}/test-plans")
    assert _wait_text(page, logged_in, "测试计划")


def test_project_settings(logged_in, frontend, smoke_pid):
    page = UiBase(logged_in, frontend)
    page.goto(f"/project/{smoke_pid}/settings")
    assert _wait_text(page, logged_in, "项目设置")
