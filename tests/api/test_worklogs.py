"""工作日志：能力门禁、编辑/删除权限、草稿汇总（任务+发布）、范围校验。"""

import pytest

from apiclient import expect_biz

pytestmark = pytest.mark.order(4)


@pytest.fixture(scope="module")
def wl_env(project_env):
    """主项目 + 一条 member 手写日志。"""

    class NS:
        pass

    ns = NS()
    ns.pid = project_env.pid
    ns.wid = project_env.member.post(
        f"/api/v1/projects/{ns.pid}/worklogs",
        {"content": "完成缓存层改造，压测 P99 从 210ms 降到 45ms", "source": "manual"},
    )["id"]
    return ns


def test_member_create_and_list(project_env, wl_env):
    res = project_env.member.get(f"/api/v1/projects/{project_env.pid}/worklogs")
    row = next(x for x in res["list"] if x["id"] == wl_env.wid)
    assert row["source"] == "manual" and row["authorType"] == "human"
    assert row["authorName"] and row["createdAt"]


def test_agent_write_capability_gate(project_env, wl_env):
    """全能 agent（空能力集=全量）可写且署名 ai；收窄 agent 缺 worklog 能力拒。"""
    aid = project_env.agent.post(
        f"/api/v1/projects/{project_env.pid}/worklogs",
        {"content": "agent 视角：本轮自动回归 111/111"},
    )["id"]
    assert aid > 0
    res = project_env.member.get(f"/api/v1/projects/{project_env.pid}/worklogs")
    row = next(x for x in res["list"] if x["id"] == aid)
    assert row["authorType"] == "ai"

    expect_biz(
        lambda: project_env.ro_agent.post(
            f"/api/v1/projects/{project_env.pid}/worklogs", {"content": "不该出现"}
        ),
        contains="未被授予",
    )


def test_edit_delete_gate(admin, project_env, wl_env):
    """非作者非 maintainer 编辑/删除拒；owner（maintainer 档）放行。"""
    expect_biz(
        lambda: project_env.ro_agent.put(f"/api/v1/worklogs/{wl_env.wid}", {"content": "越权改"}),
        contains="无权限",
    )
    expect_biz(
        lambda: project_env.agent.delete(f"/api/v1/worklogs/{wl_env.wid}"),
        contains="无权限",
    )
    project_env.member.put(
        f"/api/v1/worklogs/{wl_env.wid}", {"content": "更新：缓存层改造验收通过"}
    )
    rows = project_env.member.get(f"/api/v1/projects/{project_env.pid}/worklogs")["list"]
    row = next(x for x in rows if x["id"] == wl_env.wid)
    assert "验收通过" in row["content"]

    admin.delete(f"/api/v1/worklogs/{wl_env.wid}")
    rows = project_env.member.get(f"/api/v1/projects/{project_env.pid}/worklogs")["list"]
    assert all(x["id"] != wl_env.wid for x in rows)


def test_draft_aggregates_tasks_and_releases(admin, project_env):
    """草稿汇总：期内已完成任务（标题+产物）与发布（版本+说明）进文本，不落库。"""
    pid = project_env.pid
    tid = project_env.member.post(
        f"/api/v1/projects/{pid}/tasks",
        {"title": "worklog 草稿源任务", "description": "", "type": "chore", "priority": 3},
    )["id"]
    project_env.agent.post(f"/api/v1/tasks/{tid}/claim")
    project_env.agent.post(
        f"/api/v1/tasks/{tid}/complete", {"artifacts": "wl-artifact-回归产物说明"}
    )
    admin.post(
        f"/api/v1/projects/{pid}/releases",
        {"version": "v0.0.1-wl", "title": "草稿源发布", "notes": "草稿用版本", "channel": "stable"},
    )

    import datetime

    today = datetime.date.today().isoformat()
    res = project_env.member.get(
        f"/api/v1/projects/{pid}/worklogs/draft", params={"from": today, "to": today}
    )
    assert res["taskCount"] >= 1 and res["releaseCount"] >= 1
    assert "worklog 草稿源任务" in res["content"]
    assert "wl-artifact-回归产物说明" in res["content"]
    assert "v0.0.1-wl" in res["content"] and "草稿用版本" in res["content"]

    # 草稿是只读汇总：列表里不多出条目
    total = project_env.member.get(f"/api/v1/projects/{pid}/worklogs")["total"]
    assert total >= 1


def test_draft_range_validation(project_env):
    """from>to / 非法格式 / 超 92 天：明确拒绝。"""
    pid = project_env.pid
    expect_biz(
        lambda: project_env.member.get(
            f"/api/v1/projects/{pid}/worklogs/draft", params={"from": "2026-09-21", "to": "2026-09-20"}
        ),
        contains="不能早于",
    )
    expect_biz(
        lambda: project_env.member.get(
            f"/api/v1/projects/{pid}/worklogs/draft", params={"from": "2026/09/01", "to": "2026-09-21"}
        ),
        contains="格式",
    )
    expect_biz(
        lambda: project_env.member.get(
            f"/api/v1/projects/{pid}/worklogs/draft", params={"from": "2026-01-01", "to": "2026-09-21"}
        ),
        contains="92 天",
    )


def test_detail_endpoint(project_env):
    """轻量单条详情：完整正文 + 署名；不存在明确拒绝。"""
    pid = project_env.pid
    wid = project_env.member.post(
        f"/api/v1/projects/{pid}/worklogs",
        {"content": "detail 端点验收：缓存层改造完整正文"},
    )["id"]
    d = project_env.member.get(f"/api/v1/worklogs/{wid}")
    assert d["id"] == wid and d["authorName"] and d["createdAt"]
    assert "缓存层" in d["content"]
    assert d["source"] in ("manual", "tasks")
    expect_biz(
        lambda: project_env.member.get("/api/v1/worklogs/999999"),
        contains="不存在",
    )


def test_session_pack_injects_recent_three(project_env):
    """开工包只注入最近 3 条摘录（倒序 + 截断），完整历史不进包。"""
    pid = project_env.pid
    # 补到 4+ 条：最新一条带 markdown 首行（验摘录取首段）
    for i in range(3):
        project_env.member.post(
            f"/api/v1/projects/{pid}/worklogs", {"content": f"开工包注入序号 {i}"}
        )
    project_env.member.post(
        f"/api/v1/projects/{pid}/worklogs",
        {"content": "### 压平后的首行摘录\n\n- 第二行不进包"},
    )
    pack = project_env.agent.post("/api/v1/agent/sessions", {"projectId": pid})
    wls = pack["worklogs"]
    assert len(wls) == 3, f"应只带最近 3 条，实际 {len(wls)}"
    assert wls[0]["excerpt"].startswith("压平后的首行摘录")
    assert "第二行" not in wls[0]["excerpt"]
    assert wls[0]["authorType"] == "human" and wls[0]["author"]
