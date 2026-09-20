"""UI 冒烟：项目列表（创建入口/卡片）、成员管理（角色/agent 行）、文档树（API 层写入的文件）。"""

import pytest

from ui import UiBase

pytestmark = [pytest.mark.ui, pytest.mark.order(5)]


def test_project_list_renders_cards_and_create(logged_in, frontend, ui_world, data):
    page = UiBase(logged_in, frontend)
    page.goto(data["ui"]["routes"]["project_list"])
    assert page.wait_ele("project-list.create-btn", timeout=20)
    body = logged_in.ele("tag:body").text
    assert "集成测试-主项目" in body


def test_members_page_shows_roles_and_agent(logged_in, frontend, ui_world, data):
    page = UiBase(logged_in, frontend)
    page.goto(data["ui"]["routes"]["members"].format(pid=ui_world.pid))
    # 成员表渲染：agent 行（用户名在列）+ 管理员
    assert logged_in.wait.ele_displayed("text:itest-agent", timeout=30)
    body = logged_in.ele("tag:body").text
    assert "admin" in body


def test_docs_page_renders(logged_in, frontend, ui_world, data):
    """文档页壳渲染冒烟（树内容正确性由 api/test_docs_vault 覆盖；
    该页 body 文本提取对 DrissionPage 不友好，用页面锚点断言）。"""
    page = UiBase(logged_in, frontend)
    page.goto(data["ui"]["routes"]["docs"].format(pid=ui_world.pid))
    assert page.wait_ele("project-docs.search-input", timeout=30)
