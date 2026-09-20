"""UI 发布页（#551）：列表渲染、新建发布弹窗闭环、文件 chip 可见。"""

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
        "/api/v1/attachments/upload",
        data={"entityType": "release", "entityId": str(rid)},
        files={"file": ("app.zip", b"ui-bytes", "application/zip")},
    )

    class NS:
        pass

    ns = NS()
    ns.pid = pid
    ns.rid = rid
    return ns


def test_releases_page_renders_and_create(logged_in, frontend, ui_release, data):
    page = UiBase(logged_in, frontend)
    page.goto(f"/project/{ui_release.pid}/releases")
    assert page.wait_ele(f"releases.row-{ui_release.rid}", timeout=30)
    # 文件 chip 在列（文件名/大小由附件通道回填）
    assert "app.zip" in (logged_in.ele("tag:body").text or "") or page.eles(
        f"@data-test-id^=releases.file-"
    )

    # 新建发布弹窗闭环（新行 id 由后端分配，用版本号文本断言）
    page.click("releases.create-btn")
    page.input("releases.version-input", "v2.0.0")
    logged_in.ele("text:确定").click()
    assert logged_in.wait.ele_displayed("text:v2.0.0", timeout=15)
