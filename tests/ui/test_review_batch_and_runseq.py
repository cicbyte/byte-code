"""项目待审核批量操作（#546）+ 执行记录用例序号列（#545）。"""

import time

import pytest

from ui import UiBase

pytestmark = [pytest.mark.ui, pytest.mark.order(6)]


@pytest.fixture(scope="module")
def review_batch_env(admin, project_env):
    """自造两条待审任务（不依赖 task_chain——review_env 会把那条审掉）。"""
    pid = project_env.pid
    tids = []
    for i in (1, 2):
        tid = admin.post(
            f"/api/v1/projects/{pid}/tasks",
            {"title": f"批量审核验收-{i}", "type": "chore", "priority": 3},
        )["id"]
        project_env.agent.post(f"/api/v1/tasks/{tid}/claim")
        project_env.agent.post(f"/api/v1/tasks/{tid}/complete", {"note": "批量验收"})
        tids.append(tid)
    return {"pid": pid, "tids": tids}


def test_project_reviews_batch_approve(logged_in, frontend, review_batch_env, data):
    page = UiBase(logged_in, frontend)
    env = review_batch_env
    page.goto(f"/project/{env['pid']}/reviews")
    # 两条待审行都在
    import time as _t
    _t.sleep(2)
    body = logged_in.ele("tag:body").text or ""
    print("ANCHORS>>>", logged_in.run_js("""
      return [...document.querySelectorAll('[data-test-id]')].map(e => e.getAttribute('data-test-id')).slice(0, 20);
    """))
    for tid in env["tids"]:
        assert page.wait_ele(f"reviews.item-{tid}", timeout=20)
    # 勾两条 → 批量通过
    for tid in env["tids"]:
        page.click(f"project-reviews.check-{tid}")
    time.sleep(0.4)
    page.click("project-reviews.batch-approve-btn")
    # 批量通过无确认弹窗直接执行；行消失
    for tid in env["tids"]:
        sel = page.sel(f"reviews.item-{tid}")
        assert logged_in.wait.ele_deleted(sel, timeout=10), f"任务 {tid} 未出队"


def test_run_detail_case_sequence(logged_in, frontend, ui_runs, ui_world, data):
    page = UiBase(logged_in, frontend)
    run_id = ui_runs[0]
    page.goto(f"/project/{ui_world.pid}/test-runs/{run_id}")
    assert page.wait_ele("runs.detail.stats", timeout=30)
    # 序号列：表头 # 与行序号 1 存在
    nums = logged_in.eles("css:.num-col")
    texts = [n.text.strip() for n in nums if n.text.strip()]
    assert "#" in texts, f"缺少序号表头: {texts[:6]}"
    assert "1" in texts, f"缺少行序号: {texts[:6]}"
