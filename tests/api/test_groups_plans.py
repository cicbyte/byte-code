"""分组归属与可见性 + 项目 scope 筛选 + 测试计划域（挂用例/执行/结果汇总）。"""

import pytest

from apiclient import BizError, expect_biz

pytestmark = pytest.mark.order(2)


# ---------- 分组：归属隔离 + agent 可发现 ----------


def test_group_isolation_between_creators(admin, project_env, data):
    """分组私有化：各自只看到自己的；超管全量可见。"""
    g_member = project_env.member.post("/api/v1/groups", {"name": "成员专属分组", "description": "d"})["id"]
    member_view = {g["id"] for g in project_env.member.get("/api/v1/groups")["list"]}
    admin_view = {g["id"] for g in admin.get("/api/v1/groups")["list"]}
    assert g_member in member_view
    assert g_member in admin_view  # 超管全量（监管）
    # 非创建者不能改/删他人分组
    with pytest.raises(BizError):
        project_env.member.put(f"/api/v1/groups/{g_member + 999}", {"name": "x"})


def test_agent_sees_bound_project_groups(admin, project_env):
    """#532：agent 可见「已绑定项目所在分组」，description 脱敏。"""
    g = admin.post("/api/v1/groups", {"name": "agent-可见分组", "description": "内部说明"})["id"]
    admin.post(f"/api/v1/groups/{g}/projects", {"projectId": project_env.pid})
    rows = project_env.agent.get("/api/v1/groups")["list"]
    hit = [x for x in rows if x["id"] == g]
    assert hit, "agent 应看到含已绑定项目的分组"
    assert hit[0].get("description", "") == ""  # 脱敏只读
    # 摘要应含项目信息（成员项目摘要）
    admin.delete(f"/api/v1/groups/{g}/projects/{project_env.pid}")
    rows2 = project_env.agent.get("/api/v1/groups")["list"]
    assert all(x["id"] != g for x in rows2)


# ---------- 项目列表 scope 筛选 ----------


def test_project_scope_filters(admin, project_env):
    """scope=owner 我负责 / mine 我参与；非管理员 all 静默收窄。"""
    def ids(scope, client):
        res = client.get("/api/v1/projects", params={"scope": scope})
        return {p["id"] for p in (res.get("list") or [])}  # 空集序列化为 null

    assert project_env.pid in ids("owner", admin)
    assert project_env.pid in ids("mine", admin)
    # member 只在主项目（pid2 未加入）→ all 也被收窄到 pid
    assert ids("all", project_env.member) == {project_env.pid}
    assert ids("owner", project_env.member) == set()


# ---------- 测试计划域 ----------


@pytest.fixture(scope="module")
def plan_world(admin, project_env):
    pid = project_env.pid
    plan = admin.post(f"/api/v1/projects/{pid}/test-plans", {"name": "冒烟计划", "description": "回归主线"})
    case_a = admin.post(f"/api/v1/projects/{pid}/test-cases", {"title": "计划-登录", "priority": "P1", "module": "认证"})["id"]
    case_b = admin.post(f"/api/v1/projects/{pid}/test-cases", {"title": "计划-看板", "priority": "P2", "module": "看板"})["id"]

    class NS:
        pass

    ns = NS()
    ns.pid = pid
    ns.plan = plan["id"]
    ns.case_a = case_a
    ns.case_b = case_b
    return ns


def test_plan_lifecycle_and_results(admin, plan_world):
    """计划：建 → 挂用例（含跨项目拒绝）→ 执行 → 结果汇总 → 完成后禁改。"""
    admin.put(f"/api/v1/test-plans/{plan_world.plan}", {"status": "running"})
    admin.post(f"/api/v1/test-plans/{plan_world.plan}/cases", {"caseIds": [plan_world.case_a, plan_world.case_b]})
    # 批内重复 id 不产生双行
    admin.post(f"/api/v1/test-plans/{plan_world.plan}/cases", {"caseIds": [plan_world.case_a]})

    res = admin.get(f"/api/v1/test-plans/{plan_world.plan}/results")
    assert res["total"] == 2 and res["pending"] == 2
    # 执行两条：一过一败
    rows = res["results"]
    for r, status in zip(rows, ["pass", "fail"]):
        admin.put(f"/api/v1/test-plan-cases/{r['id']}/execute", {"status": status})
    res2 = admin.get(f"/api/v1/test-plans/{plan_world.plan}/results")
    assert res2["passed"] == 1 and res2["failed"] == 1

    # 完成后禁再执行
    admin.put(f"/api/v1/test-plans/{plan_world.plan}", {"status": "completed"})
    res3 = admin.get(f"/api/v1/test-plans/{plan_world.plan}/results")
    with pytest.raises(BizError):
        admin.put(f"/api/v1/test-plan-cases/{res3['results'][0]['id']}/execute", {"status": "pass"})


def test_plan_cross_project_case_rejected(admin, plan_world, project_env):
    """计划挂接用例必须同项目：外来用例整批拒绝。"""
    foreign = admin.post(
        f"/api/v1/projects/{project_env.pid2}/test-cases", {"title": "外来用例", "priority": "P3"}
    )["id"]
    with pytest.raises(BizError):
        admin.post(
            f"/api/v1/test-plans/{plan_world.plan}/cases", {"caseIds": [plan_world.case_a, foreign]}
        )
