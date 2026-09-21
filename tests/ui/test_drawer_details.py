"""详情抽屉化交互（#590）：审核中心/QA/专题/执行记录四页抽屉闭环。"""

import time

import pytest

from ui import UiBase

pytestmark = [pytest.mark.ui, pytest.mark.order(6)]


def _wait_drawer(page, timeout=10):
    """抽屉出现轮询（UiBase.eles 会包锚点，须走原生 DP 页面）。"""
    deadline = time.time() + timeout
    while time.time() < deadline:
        if page.page.eles("css:.n-drawer-content"):
            return True
        time.sleep(0.3)
    return False


@pytest.fixture(scope="module")
def dw_env(project_env, admin):
    pid = project_env.pid
    # 待审任务（进审核中心）
    rtid = project_env.member.post(
        f"/api/v1/projects/{pid}/tasks",
        {"title": "抽屉验收-待审任务", "description": "抽屉验收用的任务描述正文", "type": "chore", "priority": 3},
    )["id"]
    project_env.agent.post(f"/api/v1/tasks/{rtid}/claim")
    project_env.agent.post(f"/api/v1/tasks/{rtid}/complete", {"artifacts": ""})
    # QA 条目
    admin.post(f"/api/v1/projects/{pid}/qas", {"question": "抽屉验收-怎么备份数据库", "answer": "见官方备份文档，先停写再拷贝。", "tags": "部署"})
    # 专题（带阶段）
    tp = admin.post(
        f"/api/v1/projects/{pid}/topics",
        {"title": "抽屉验收-专题", "goal": "验证专题抽屉交互", "acceptance": "抽屉可见阶段清单", "assigneeId": 0},
    )["id"]
    admin.post(f"/api/v1/projects/{pid}/topics/{tp}/phases/add", {"title": "阶段一"})

    class NS:
        pass

    ns = NS()
    ns.pid = pid
    ns.rtid = rtid
    ns.tp = tp
    return ns


def test_global_review_drawer_approve(logged_in, frontend, dw_env):
    page = UiBase(logged_in, frontend)
    page.goto("/reviews/index")
    assert page.wait_ele(f"reviews.item-{dw_env.rtid}", timeout=30)
    page.click(f"reviews.item-{dw_env.rtid}")
    assert _wait_drawer(page)
    body = logged_in.ele("tag:body").text or ""
    assert "抽屉验收用的任务描述正文" in body, "抽屉应展示完整描述"
    # footer 通过：抽屉内按文本精确过滤（避开表格同名按钮）
    for b in page.page.eles("css:.n-drawer .n-button"):
        if (b.text or "").strip() == "通过":
            b.click()
            break
    else:
        raise AssertionError("抽屉 footer 未找到「通过」")
    deadline = time.time() + 10
    while time.time() < deadline:
        if not page.eles(f"@data-test-id=reviews.item-{dw_env.rtid}"):
            break
        time.sleep(0.3)
    assert not page.eles(f"@data-test-id=reviews.item-{dw_env.rtid}"), "通过后行应出队"


def test_qa_drawer(logged_in, frontend, dw_env):
    page = UiBase(logged_in, frontend)
    page.goto(f"/project/{dw_env.pid}/qas")
    assert logged_in.wait.ele_displayed("text:抽屉验收-怎么备份数据库", timeout=30)
    logged_in.ele("text:抽屉验收-怎么备份数据库").click()
    assert _wait_drawer(page)
    assert logged_in.wait.ele_displayed("text:先停写再拷贝", timeout=10), "抽屉应展示答案"


def test_topics_drawer(logged_in, frontend, dw_env):
    page = UiBase(logged_in, frontend)
    page.goto(f"/project/{dw_env.pid}/topics?status=all")
    assert logged_in.wait.ele_displayed("text:抽屉验收-专题", timeout=30)
    logged_in.ele("text:抽屉验收-专题").click()
    assert _wait_drawer(page)
    body = logged_in.ele("tag:body").text or ""
    assert "验证专题抽屉交互" in body and "阶段清单" in body and "阶段一" in body


def test_run_case_drawer(logged_in, frontend, dw_env):
    """执行记录失败用例：查看按钮开抽屉（无执行记录则跳过——全量套件内必有）。"""
    page = UiBase(logged_in, frontend)
    page.goto(f"/project/{dw_env.pid}/test-runs")
    time.sleep(2)
    rows = page.page.eles("css:[data-test-id^='runs.detail-btn-']")
    if not rows:
        pytest.skip("项目内暂无执行记录")
    rows[0].click()
    assert page.wait_ele("runs.detail.stats", timeout=30)
    views = page.page.eles("css:[data-test-id^='runs.detail-view-']")
    if not views:
        pytest.skip("该次执行无失败/错误用例")
    views[0].click()
    assert _wait_drawer(page)
    assert logged_in.wait.ele_displayed("text:失败信息", timeout=10)
