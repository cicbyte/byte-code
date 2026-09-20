"""UI 发布页（#551/#552）：列表渲染、文件 chip、新建弹窗闭环、直链弹窗。"""

import pytest

from ui import UiBase

pytestmark = [pytest.mark.ui, pytest.mark.order(5)]


@pytest.fixture(scope="module")
def ui_release(admin, project_env):
    """API 造一个带文件的发布（UI 只验形，造数走 API）。"""
    pid = admin.post("/api/v1/projects", {"name": "集成测试-发布UI", "code": "RELU"})["id"]
    rid = admin.post(
        f"/api/v1/projects/{pid}/releases", {"version": "v1.0.0", "title": "UI 验收", "notes": "说明"}
    )["id"]
    admin.http.post(
        f"/api/v1/releases/{rid}/files",
        files={"file": ("app.zip", b"ui-bytes", "application/zip")},
    )

    class NS:
        pass

    ns = NS()
    ns.pid = pid
    ns.rid = rid
    return ns


def test_releases_page_renders_create_and_share(logged_in, frontend, ui_release, data):
    page = UiBase(logged_in, frontend)
    page.goto(f"/project/{ui_release.pid}/releases")
    assert page.wait_ele(f"releases.row-{ui_release.rid}", timeout=30)
    # 文件 chip 在列（独立文件体系回填）
    assert "app.zip" in (logged_in.ele("tag:body").text or "")

    # 新建发布弹窗闭环（新行 id 由后端分配，用版本号文本断言）
    page.click("releases.create-btn")
    page.input("releases.version-input", "v2.0.0")
    logged_in.ele("text:确定").click()
    assert logged_in.wait.ele_displayed("text:v2.0.0", timeout=15)

    # 文件管理弹窗 → 生成直链 → 直链弹窗出现
    page.click(f"releases.files-btn-{ui_release.rid}")
    page.wait_ele("releases.files-modal", timeout=15)
    # 该发布只有一个文件：按按钮文本点「生成直链」（前缀定位器在 DP 上不稳）
    assert logged_in.wait.ele_displayed("text:生成直链", timeout=15)
    logged_in.ele("text:生成直链").click()
    assert page.wait_ele("releases.link-modal", timeout=15)
