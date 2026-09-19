"""UI 测试区三页：总览（趋势/统计/Flaky）、执行记录列表、run 详情页（#535 拆分）。"""

import pytest

from ui import UiBase

pytestmark = [pytest.mark.ui, pytest.mark.order(3)]


@pytest.mark.dependency()
def test_overview_page_renders(logged_in, frontend, ui_world, ui_runs, data):
    page = UiBase(logged_in, frontend)
    page.goto(data["ui"]["routes"]["overview"].format(pid=ui_world.pid))
    # 统计卡 + 趋势图容器（从执行记录页迁出）
    assert page.wait_ele("overview.stat-cards", timeout=15)
    assert page.wait_ele("overview.trend-chart", timeout=15)
    # Flaky 面板在总览页，且抖动用例在列
    assert page.wait_ele("overview.flaky-panel", timeout=10)
    assert "test_flip" in logged_in.ele("tag:body").text


@pytest.mark.dependency()
def test_runs_page_lists_rows(logged_in, frontend, ui_world, ui_runs, data):
    page = UiBase(logged_in, frontend)
    page.goto(data["ui"]["routes"]["runs"].format(pid=ui_world.pid))
    # 纯列表页：最新一条执行记录行在（趋势/Flaky 已迁总览）
    assert page.wait_ele(f"runs.row-{ui_runs[-1]}", timeout=15)


@pytest.mark.dependency(depends=["test_runs_page_lists_rows"])
def test_run_detail_page(logged_in, frontend, ui_world, ui_runs, data):
    page = UiBase(logged_in, frontend)
    page.goto(data["ui"]["routes"]["runs"].format(pid=ui_world.pid))
    page.click(f"runs.detail-btn-{ui_runs[-1]}")
    # 路由跳详情页：URL 带 runId，统计区与标题渲染
    logged_in.wait.url_change(f"/test-runs/{ui_runs[-1]}", timeout=10)
    assert page.wait_ele("runs.detail.stats", timeout=15)
    body = logged_in.ele("tag:body").text
    assert f"执行记录 #{ui_runs[-1]}" in body
    # 返回列表
    page.click("runs.detail.back")
    logged_in.wait.url_change(f"/test-runs/{ui_runs[-1]}", exclude=True, timeout=10)
