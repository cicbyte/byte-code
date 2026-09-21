"""任务批量关闭：done→closed 清账、混合批次失败明细、agent 拒绝。"""

import pytest

from apiclient import expect_biz

pytestmark = pytest.mark.order(4)


def _make_done(project_env, admin, pid, title):
    """造一个 done 任务：member 建 → agent 认领完成 → admin 批量审过。"""
    tid = project_env.member.post(
        f"/api/v1/projects/{pid}/tasks", {"title": title, "description": "", "type": "chore", "priority": 3}
    )["id"]
    project_env.agent.post(f"/api/v1/tasks/{tid}/claim")
    project_env.agent.post(f"/api/v1/tasks/{tid}/complete", {"artifacts": ""})
    admin.post("/api/v1/reviews/batch", {"ids": [tid], "status": "approved"})
    return tid


@pytest.fixture(scope="module")
def bc_env(project_env, admin):
    pid = project_env.pid
    ids = [_make_done(project_env, admin, pid, f"批量关闭源任务-{i}") for i in range(2)]
    open_id = project_env.member.post(
        f"/api/v1/projects/{pid}/tasks", {"title": "批量关闭-未完成任务", "description": "", "type": "chore", "priority": 3}
    )["id"]

    class NS:
        pass

    ns = NS()
    ns.pid = pid
    ns.done_ids = ids
    ns.open_id = open_id
    return ns


def test_batch_close_done(project_env, bc_env):
    res = project_env.member.post(
        f"/api/v1/projects/{bc_env.pid}/tasks/batch-close", {"ids": bc_env.done_ids}
    )
    assert res["succeeded"] == 2 and res["failed"] == []
    for tid in bc_env.done_ids:
        task = project_env.member.get(f"/api/v1/tasks/{tid}")
        assert task["status"] == "closed"


def test_batch_close_mixed_and_missing(project_env, admin, bc_env):
    """混合批次：done 关闭成功，open/不存在逐条报因，不阻断整批。"""
    done_id = _make_done(project_env, admin, bc_env.pid, "批量关闭-混合-done")
    res = project_env.member.post(
        f"/api/v1/projects/{bc_env.pid}/tasks/batch-close",
        {"ids": [done_id, bc_env.open_id, 999999]},
    )
    assert res["succeeded"] == 1
    errors = {f["id"]: f["error"] for f in res["failed"]}
    assert "仅已完成" in errors[bc_env.open_id]
    assert 999999 in errors and errors[999999] == "任务不存在"


def test_batch_close_rejects_agent(project_env, bc_env):
    expect_biz(
        lambda: project_env.agent.post(
            f"/api/v1/projects/{bc_env.pid}/tasks/batch-close", {"ids": bc_env.done_ids}
        ),
        contains="仅限人类",
    )


def test_batch_close_already_closed(project_env, bc_env):
    res = project_env.member.post(
        f"/api/v1/projects/{bc_env.pid}/tasks/batch-close", {"ids": bc_env.done_ids}
    )
    assert res["succeeded"] == 0
    assert all(f["error"] == "已关闭" for f in res["failed"])


def test_batch_close_cross_project_guard(project_env, admin, bc_env):
    """另一项目的任务 id 不能经由本项目端点关闭（member 非 pid2 成员，任务由 admin 建）。"""
    other_tid = admin.post(
        f"/api/v1/projects/{project_env.pid2}/tasks",
        {"title": "别项目的任务", "description": "", "type": "chore", "priority": 3},
    )["id"]
    res = project_env.member.post(
        f"/api/v1/projects/{bc_env.pid}/tasks/batch-close", {"ids": [other_tid]}
    )
    assert res["succeeded"] == 0
    assert any("不属于该项目" in f["error"] for f in res["failed"])
