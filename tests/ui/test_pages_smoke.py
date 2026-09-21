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


def test_docs_batch_move(logged_in, frontend, ui_world, backend, admin, data):
    """批量归类移动：勾选两个文档 → 默认目录预选所在父目录 → 填新子目录移入。"""
    import time
    page = UiBase(logged_in, frontend)
    pid = ui_world.pid
    docs_dir = backend.workdir / "resource" / "projects" / str(pid) / "docs" / "dev-docs"
    docs_dir.mkdir(parents=True, exist_ok=True)
    for name in ("ops-a.md", "ops-b.md"):
        (docs_dir / name).write_bytes(f"# {name}\n内容\n".encode("utf-8"))
    admin.get(f"/api/v1/projects/{pid}/docs/tree?space=work")

    page.goto(data["ui"]["routes"]["docs"].format(pid=pid))
    assert page.wait_ele("project-docs.search-input", timeout=30)

    def check_row(path):
        """勾选树节点：node-{路径} 锚点定位行，行内下钻 checkbox，
        坐标点击（DP 元素引用在点击多步间会因 DOM 替换失效）。"""
        row = page.ele(f"project-docs.node-{path}")
        cb = row.ele("css:.n-tree-node-checkbox", timeout=5)
        assert cb is not None, f"未找到 {path} 的勾选框"
        # actions.move(x,y) 是相对偏移语义（第二次会飞），move_to 现取坐标最稳
        logged_in.actions.move_to(cb)
        logged_in.actions.click()

    # 目录默认折叠：点树行锚点展开 dev-docs（node-{路径} 由 treeNodeProps 挂载）
    page.ele("project-docs.node-dev-docs").click()
    assert logged_in.wait.ele_displayed("text:ops-a.md", timeout=10)
    check_row("dev-docs/ops-a.md")
    time.sleep(0.5)
    check_row("dev-docs/ops-b.md")
    assert page.wait_ele("project-docs.batch-bar", timeout=8)
    page.click("project-docs.batch-move-btn")
    assert page.wait_ele("project-docs.batch-subdir-input", timeout=8)
    # 默认目标目录 = 选中项公共父目录（dev-docs）
    sel_text = page.ele("project-docs.batch-target-select").text or ""
    assert "dev-docs" in sel_text, f"目标目录未预选 dev-docs: {sel_text}"
    page.input("project-docs.batch-subdir-input", "ops")
    # 弹窗内精确匹配「移动」（text: 模糊会先命中批量条的「移动到…」）
    dlg = logged_in.ele("css:.n-dialog")
    btns = [b for b in dlg.eles("css:button") if (b.text or "").strip() == "移动"]
    assert btns, "批量移动弹窗缺少「移动」按钮"
    btns[0].click()
    assert logged_in.wait.ele_displayed("text:已移动 2 项", timeout=10)
    # API 级验证落点：dev-docs/ops/ 下两文件
    tree = admin.get(f"/api/v1/projects/{pid}/docs/tree?space=work")["tree"]

    def paths_of(nodes, acc):
        for n in nodes:
            acc.append(n["path"])
            paths_of(n.get("children") or [], acc)
        return acc

    all_paths = paths_of(tree, [])
    assert "dev-docs/ops/ops-a.md" in all_paths and "dev-docs/ops/ops-b.md" in all_paths, all_paths


def test_docs_crlf_no_false_dirty(logged_in, frontend, ui_world, backend, admin, data):
    """CRLF 文档打开后未编辑直接切换，不弹「未保存」（编辑器把 CRLF
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
