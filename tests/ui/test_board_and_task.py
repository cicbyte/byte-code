"""UI 看板与任务详情：五态列渲染、卡片锚点、详情抽屉。"""

import pytest

from ui import UiBase

pytestmark = [pytest.mark.ui, pytest.mark.order(2)]


@pytest.mark.dependency()
def test_board_columns_render(logged_in, frontend, ui_world, data):
    page = UiBase(logged_in, frontend)
    page.goto(data["ui"]["routes"]["board"].format(pid=ui_world.pid))
    # 已完成列存在且含 done 任务卡（data-test-id 锚点）
    col = page.wait_ele(f'board.col-done')
    assert col
    card = page.wait_ele(f'board.card-{ui_world.done_task}')
    assert card


def test_board_blocked_column(logged_in, frontend, ui_world):
    page = UiBase(logged_in, frontend)
    page.goto(f"/project/{ui_world.pid}/board")
    assert page.wait_ele("board.col-blocked")
    assert page.wait_ele(f'board.card-{ui_world.blocked_task}')


@pytest.mark.dependency(depends=["test_board_columns_render"])
def test_task_detail_drawer(logged_in, frontend, ui_world, data):
    """任务列表 → 详情抽屉：标题锚点 + 描述卡。"""
    page = UiBase(logged_in, frontend)
    page.goto(data["ui"]["routes"]["tasks"].format(pid=ui_world.pid))
    page.click(f'tasks.detail-btn-{ui_world.done_task}')
    title = page.text_of("task-detail.title")
    assert title  # 抽屉打开且标题可读
