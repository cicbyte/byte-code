"""UI 发布页（#551/#552/#554）：列表渲染、时间轴导航、文件 chip、新建弹窗闭环、直链弹窗。"""

import pytest

from ui import UiBase

pytestmark = [pytest.mark.ui, pytest.mark.order(5)]


@pytest.fixture(scope="module")
def ui_release(admin, project_env):
    """API 造两个带文件的发布（UI 只验形，造数走 API）。"""
    pid = admin.post("/api/v1/projects", {"name": "集成测试-发布UI", "code": "RELU"})["id"]
    rid = admin.post(
        f"/api/v1/projects/{pid}/releases",
        {
            "version": "v1.0.0",
            "title": "UI 验收",
            # 足量 markdown 撑破 180px 折叠高度，锁定「按渲染溢出出开关」逻辑
            "notes": "## 标题\n\n" + "\n".join(f"- 条目 {i} 有一段说明文字用于撑高内容" for i in range(1, 15)),
        },
    )["id"]
    rid2 = admin.post(
        f"/api/v1/projects/{pid}/releases", {"version": "v0.9.0", "title": "旧版本", "notes": "旧"}
    )["id"]
    admin.http.post(
        f"/api/v1/releases/{rid}/files",
        files={"file": ("app.zip", b"ui-bytes", "application/zip")},
    )
    # 第二个文件专供行内删除闭环（删了不影响后续直链流程）
    admin.http.post(
        f"/api/v1/releases/{rid}/files",
        files={"file": ("legacy.zip", b"old-bytes", "application/zip")},
    )
    files = admin.get(f"/api/v1/releases/{rid}/files")["list"]

    class NS:
        pass

    ns = NS()
    ns.pid = pid
    ns.rid = rid
    ns.rid2 = rid2
    ns.fids = {f["fileName"]: f["id"] for f in files}
    return ns


def test_releases_page_renders_create_and_share(logged_in, frontend, ui_release, data):
    page = UiBase(logged_in, frontend)
    page.goto(f"/project/{ui_release.pid}/releases")
    assert page.wait_ele(f"releases.row-{ui_release.rid}", timeout=30)
    # 文件 chip 在列（独立文件体系回填）
    assert "app.zip" in (logged_in.ele("tag:body").text or "")

    # 版本时间轴：容器 + 每版本一项，点击旧版本项 → URL 带 ?tag= 且该项高亮
    assert page.wait_ele("releases.timeline", timeout=15)
    tl_old = f"releases.timeline-item-{ui_release.rid2}"
    assert page.wait_ele(tl_old, timeout=15)
    page.click(tl_old)
    assert "tag=v0.9.0" in (logged_in.url or "")
    assert "active" in (page.ele(tl_old).attr("class") or "")

    # 长说明折叠出现「展开完整说明」，点击展开后变「收起」（元素下钻定位后代）
    card = page.ele(f"releases.row-{ui_release.rid}")
    toggle = card.ele(".rel-notes-toggle", timeout=8)
    assert toggle is not None and "展开完整说明" in (toggle.text or "")
    toggle.click()
    assert "收起" in (card.ele(".rel-notes-toggle", timeout=5).text or "")

    # 行内快捷操作：hover 文件行显现 直链/删除，行内删除 legacy.zip 闭环
    legacy = ui_release.fids["legacy.zip"]
    page.ele(f"releases.file-{legacy}").hover()
    # JS 点击：DP 二次定位的滚动会丢失 :hover（ops 回到 display:none 无矩形），真实用户无此问题
    page.ele(f"releases.row-del-{legacy}").click(by_js=True)
    dlg = logged_in.ele("css:.n-dialog")  # 限定对话框内找「删除」，避免命中卡头同名按钮
    confirm = [b for b in dlg.eles("css:.n-dialog__action .n-button") if (b.text or "").strip() == "删除"]
    assert confirm, "确认对话框未出现删除按钮"
    confirm[0].click()
    assert logged_in.wait.ele_deleted("css:.n-dialog", timeout=8)  # 等对话框关闭过渡完成
    legacy_sel = page.sel(f"releases.file-{legacy}")
    assert logged_in.wait.ele_deleted(legacy_sel, timeout=8)
    assert "legacy.zip" not in (logged_in.ele("tag:body").text or "")

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
