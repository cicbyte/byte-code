"""UI 执行记录页：趋势图容器、记录行、Flaky 面板。"""

import pytest

from ui import UiBase

pytestmark = [pytest.mark.ui, pytest.mark.order(3)]


@pytest.mark.dependency()
def test_runs_page_renders(logged_in, frontend, ui_world, ui_runs, data):
    page = UiBase(logged_in, frontend)
    page.goto(data["ui"]["routes"]["runs"].format(pid=ui_world.pid))
    # 趋势图容器 + 最新一条执行记录行（锚点带 run id）
    assert page.wait_ele("runs.trend-chart", timeout=15)
    assert page.wait_ele(f'runs.row-{ui_runs[-1]}', timeout=15)


def test_flaky_tab_lists_flip(logged_in, frontend, ui_world, ui_runs, data):
    page = UiBase(logged_in, frontend)
    page.goto(f"/project/{ui_world.pid}/test-runs")
    # n-tabs 的按钮与 pane 分离渲染：按文本点 tab（pane 锚点只标在内容容器）
    logged_in.ele("text:Flaky 用例").click()
    logged_in.wait.doc_loaded()
    body = logged_in.ele("tag:body").text
    assert "test_flip" in body
