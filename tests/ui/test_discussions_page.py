"""讨论区 UI 冒烟：创建、回复、转任务闭环。"""

import pytest

from ui import UiBase

pytestmark = [pytest.mark.ui, pytest.mark.order(5)]


@pytest.fixture(scope="module")
def ui_disc(admin, project_env):
    pid = project_env.pid
    did = admin.post(
        f"/api/v1/projects/{pid}/discussions",
        {"title": "UI 验收讨论", "body": "API 造的讨论正文"},
    )["id"]

    class NS:
        pass

    ns = NS()
    ns.pid = pid
    ns.did = did
    return ns


def test_discussions_create_reply_convert(logged_in, frontend, ui_disc):
    page = UiBase(logged_in, frontend)
    page.goto(f"/project/{ui_disc.pid}/discussions")
    assert page.wait_ele("discussions.create-btn", timeout=30)

    # 打开 API 造的线程 → 详情渲染
    page.click(f"discussions.row-{ui_disc.did}")
    assert page.wait_ele("discussions.detail-title", timeout=15)
    assert "UI 验收讨论" in page.text_of("discussions.detail-title")

    # 回复闭环
    page.input("discussions.reply-input", "UI 回复：赞成")
    page.click("discussions.reply-btn")
    assert logged_in.wait.ele_displayed("text:UI 回复：赞成", timeout=10)

    # 转任务闭环：默认标题同讨论标题 → 创建 → 详情出现已转链接
    page.click("discussions.convert-btn")
    page.wait_ele("discussions.convert-title", timeout=10)
    logged_in.ele("text:创建任务").click()
    assert logged_in.wait.ele_displayed("text:已转任务", timeout=10)
    assert page.wait_ele("discussions.task-link", timeout=10)

    # 新建讨论弹窗闭环（标题必填）
    page.click("discussions.create-btn")
    page.input("discussions.title-input", "冒烟新建的讨论")
    logged_in.ele("text:发布").click()
    assert logged_in.wait.ele_displayed("text:冒烟新建的讨论", timeout=10)
