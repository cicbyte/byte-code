"""UI 用例页：执行统计列 + 历史记录弹窗。

API 造数（用例 + 多次 run 上报，含 pre-sync 无映射旧行与显式映射行），
UI 验形（统计列计数 / 弹窗汇总与明细 / 失败信息展开）。
"""

import pytest

from ui import UiBase

pytestmark = [pytest.mark.ui, pytest.mark.order(4)]


@pytest.fixture(scope="module")
def cases_stats_world(admin, project_env):
    """统计世界：sync 用例（外部键：run1 无映射 pass + run2 显式映射 fail）
    + 手工用例（无外部键，run3 显式映射 error）。"""
    pid = project_env.pid
    base = f"/api/v1/projects/{pid}"
    sync_case = admin.post(
        f"{base}/test-cases",
        {
            "title": "统计-登录成功",
            "module": "认证",
            "priority": "P1",
            "category": "functional",
            "externalKey": "ui/test_stats.py::test_login",
        },
    )["id"]
    manual_case = admin.post(
        f"{base}/test-cases",
        {"title": "统计-手工冒烟", "module": "认证", "priority": "P2", "category": "functional"},
    )["id"]

    def report(cases):
        return project_env.agent.post(
            f"{base}/test-runs", {"source": "pytest", "branch": "feat-stats", "cases": cases}
        )["id"]

    # run1 模拟 pre-sync 旧行：无映射仅外部键（pass）
    report([{"externalKey": "ui/test_stats.py::test_login", "status": "pass", "durationMs": 300}])
    # run2 显式映射（fail，带 traceback）
    report(
        [
            {
                "externalKey": "ui/test_stats.py::test_login",
                "status": "fail",
                "durationMs": 120,
                "message": "AssertionError: expected 200",
                "testCaseId": sync_case,
            }
        ]
    )
    # run3 手工用例显式映射（error）
    report(
        [
            {
                "externalKey": "ui/test_stats.py::test_manual",
                "status": "error",
                "durationMs": 50,
                "message": "fixture boom",
                "testCaseId": manual_case,
            }
        ]
    )

    class NS:
        pass

    ns = NS()
    ns.pid = pid
    ns.sync_case = sync_case
    ns.manual_case = manual_case
    return ns


@pytest.mark.dependency()
def test_stats_column_counts(logged_in, frontend, cases_stats_world, data):
    page = UiBase(logged_in, frontend)
    page.goto(data["ui"]["routes"]["cases"].format(pid=cases_stats_world.pid))
    td = page.wait_ele(f"test-cases.stats-{cases_stats_world.sync_case}", timeout=30).parent("tag:td")
    row = td.text
    assert "2 次" in row and "1 过" in row and "1 败" in row, row
    # 手工用例：1 次（error 并入失败口径）
    mtd = page.wait_ele(f"test-cases.stats-{cases_stats_world.manual_case}").parent("tag:td")
    mrow = mtd.text
    assert "1 次" in mrow and "1 败" in mrow, mrow


@pytest.mark.dependency(depends=["test_stats_column_counts"])
def test_history_dialog_summary_and_rows(logged_in, frontend, cases_stats_world, data):
    page = UiBase(logged_in, frontend)
    page.goto(data["ui"]["routes"]["cases"].format(pid=cases_stats_world.pid))
    page.click(f"test-cases.stats-{cases_stats_world.sync_case}")
    page.wait_ele("test-cases.history-drawer", timeout=10)
    summary = page.text_of("test-cases.history-summary")
    assert "共 2 次" in summary and "1 成功" in summary and "1 失败" in summary, summary
    assert "通过率 50%" in summary, summary
    body = logged_in.ele("tag:body").text
    # 明细行带 run 上下文（分支）；失败信息默认折叠
    assert "feat-stats" in body
    assert "AssertionError" not in body


@pytest.mark.dependency(depends=["test_history_dialog_summary_and_rows"])
def test_history_dialog_expand_message(logged_in, frontend, cases_stats_world, data):
    page = UiBase(logged_in, frontend)
    page.goto(data["ui"]["routes"]["cases"].format(pid=cases_stats_world.pid))
    page.click(f"test-cases.stats-{cases_stats_world.sync_case}")
    page.wait_ele("test-cases.history-drawer", timeout=10)
    # 最近一行（run2 fail）点击展开 traceback
    page.ele("test-cases.history-drawer").ele("tag:tbody").ele("tag:tr").click()
    assert "AssertionError: expected 200" in logged_in.ele("tag:body").text
