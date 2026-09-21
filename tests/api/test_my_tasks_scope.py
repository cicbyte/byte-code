"""my-tasks scope=all：仪表盘任务卡钻取口径（可见范围全部 vs 指派给我）。"""

import pytest

pytestmark = pytest.mark.order(4)


@pytest.fixture(scope="module")
def ms_env(project_env, admin):
    """pid 内一个 agent 执行中的任务（admin 未被指派）+ pid2 一个任务（member 不可见）。"""
    pid = project_env.pid
    tid = project_env.member.post(
        f"/api/v1/projects/{pid}/tasks", {"title": "scope口征-agent执行中", "description": "", "type": "chore", "priority": 3}
    )["id"]
    project_env.agent.post(f"/api/v1/tasks/{tid}/claim")
    other = admin.post(
        f"/api/v1/projects/{project_env.pid2}/tasks", {"title": "scope口征-别项目", "description": "", "type": "chore", "priority": 3}
    )["id"]

    class NS:
        pass

    ns = NS()
    ns.pid = pid
    ns.tid = tid
    ns.other = other
    return ns


def test_admin_scope_all_sees_agent_tasks(admin, backend, ms_env):
    """scope=all：admin 看得到 agent 名下的进行中任务；mine 缺省看不到。"""
    # 空结果序列化为 null（nil 切片既有行为），与 [] 同义
    mine = admin.get("/api/v1/my-tasks", params={"status": "in_progress"})
    assert ms_env.tid not in {t["id"] for t in (mine["list"] or [])}
    allres = admin.get("/api/v1/my-tasks", params={"scope": "all", "status": "in_progress"})
    assert ms_env.tid in {t["id"] for t in (allres["list"] or [])}


def test_member_scope_all_limited_to_own_projects(project_env, ms_env):
    """成员 scope=all 仍限所在项目：pid2 任务不可见（防跨项目泄露）。"""
    res = project_env.member.get("/api/v1/my-tasks", params={"scope": "all", "size": 200})
    ids = {t["id"] for t in res["list"]}
    assert ms_env.tid in ids, "member 在 pid 内，应可见"
    assert ms_env.other not in ids, "pid2 任务对非成员不可见"


def test_scope_validation(project_env, ms_env):
    from apiclient import expect_biz

    expect_biz(
        lambda: project_env.member.get("/api/v1/my-tasks", params={"scope": "world"}),
        contains="视角不合法",
    )
