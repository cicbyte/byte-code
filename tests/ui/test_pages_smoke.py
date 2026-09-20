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


def test_docs_crlf_no_false_dirty(logged_in, frontend, ui_world, backend, admin, data):
    """CRLF 文档打开后未编辑直接切换，不弹「未保存」（#558：编辑器把 CRLF
    规范化为 LF，换行差异不计为脏）。
    注：真编辑仍拦截的正向对照已在真实浏览器人工验证（headless DP 模拟 CM6
    键入不可行），自动用例只锁本任务回归方向——不误报。"""
    import time as _time

    pid = ui_world.pid
    docs_dir = backend.workdir / "resource" / "projects" / str(pid) / "docs"
    docs_dir.mkdir(parents=True, exist_ok=True)
    (docs_dir / "crlf-itest.md").write_bytes("# CRLF 文档\r\n第一行\r\n".encode("utf-8"))
    (docs_dir / "lf-itest.md").write_bytes("# LF 文档\n第一行\n".encode("utf-8"))
    admin.get(f"/api/v1/projects/{pid}/docs/tree?space=work")  # 触发目录重扫

    page = UiBase(logged_in, frontend)
    page.goto(data["ui"]["routes"]["docs"].format(pid=pid))
    assert page.wait_ele("project-docs.search-input", timeout=30)

    def wait_heading(name: str, timeout_s: float = 10.0) -> bool:
        """等卡片标题真正切到目标文件（openFile 异步，.cm-content 会命中上一个文件）。
        页面有两张卡（目录 + 文档），querySelector 会命中「目录」，须遍历匹配。"""
        deadline = _time.time() + timeout_s
        while _time.time() < deadline:
            for el in logged_in.eles("css:.n-card-header__main"):
                if name in (el.text or ""):
                    return True
            _time.sleep(0.3)
        return False

    def open_doc(name: str):
        logged_in.ele(f"text:{name}").click()
        assert wait_heading(name), f"打开 {name} 超时"

    open_doc("crlf-itest.md")
    assert logged_in.wait.ele_displayed("css:.cm-content", timeout=15)

    # 未编辑切到 LF 文档：不得弹未保存确认
    open_doc("lf-itest.md")
    _time.sleep(1.0)
    assert not logged_in.eles("css:.n-dialog"), "CRLF 文档未编辑切换被误报为未保存"
