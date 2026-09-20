"""反馈已发送视图 + QA 库（去重 upsert/命中计数/检索/归档）+ 全局记忆提案审核。"""

import pytest

from apiclient import BizError

pytestmark = pytest.mark.order(2)


# ---------- 反馈已发送 ----------


def test_feedback_sent_view(admin, project_env):
    """#447：agent 的已发送反馈视图（自建反馈，跨模块无依赖）。"""
    # 能力按「来源项目」校验：来源必须是 agent 接入的项目（pid）
    p = admin.post("/api/v1/projects", {"name": "集成测试-发送方", "code": "SENT"})["id"]
    admin.post(f"/api/v1/projects/{p}/relations", {"relatedProjectId": project_env.pid})
    project_env.agent.post(
        f"/api/v1/projects/{p}/feedbacks",
        {"title": "已发送视图验证", "content": "sent", "sourceProjectId": project_env.pid},
    )
    rows = project_env.agent.get("/api/v1/feedbacks/sent")["list"]
    hit = next(r for r in rows if r["title"] == "已发送视图验证")
    assert hit.get("targetProjectId") or hit.get("targetProjectName")  # 回填目标


# ---------- QA 库 ----------


@pytest.fixture(scope="module")
def qa_world(admin, project_env):
    q = admin.post(
        f"/api/v1/projects/{project_env.pid}/qas",
        {"question": "集成测试如何构造临时后端？", "answer": "用 helpers/booted.py 编排", "tags": "测试,基建"},
    )

    class NS:
        pass

    ns = NS()
    ns.qid = q["id"]
    ns.updated_first = q["updated"]
    return ns


def test_qa_dedup_upsert(admin, project_env, qa_world):
    """同问题再答=更新而非新建。"""
    r = admin.post(
        f"/api/v1/projects/{project_env.pid}/qas",
        {"question": "集成测试如何构造临时后端？", "answer": "更新后的答案"},
    )
    assert r["updated"] is True and r["id"] == qa_world.qid
    assert qa_world.updated_first is False


def test_qa_keyword_search_and_hits(admin, project_env, qa_world):
    rows = admin.get(
        f"/api/v1/projects/{project_env.pid}/qas", params={"keyword": "临时后端"}
    )["list"]
    assert any(x["id"] == qa_world.qid for x in rows)
    # 命中计数：两次 hit → hits=2
    for _ in range(2):
        admin.post(f"/api/v1/projects/{project_env.pid}/qas/{qa_world.qid}/hit")
    rows2 = admin.get(f"/api/v1/projects/{project_env.pid}/qas", params={"keyword": "临时后端"})["list"]
    hit = next(x for x in rows2 if x["id"] == qa_world.qid)
    assert hit["hits"] == 2


def test_qa_archive_hides_from_default(admin, project_env, qa_world):
    admin.post(f"/api/v1/projects/{project_env.pid}/qas/{qa_world.qid}/archive")
    rows = admin.get(f"/api/v1/projects/{project_env.pid}/qas", params={"keyword": "临时后端"})["list"]
    assert all(x["id"] != qa_world.qid for x in rows)


# ---------- 全局记忆提案 ----------


def test_global_memory_proposal_lifecycle(admin, project_env, data):
    """提案 → 未决防重 → 采纳落正式记忆（全员可读）→ 占用后再提案被拒。"""
    key = "itest-global-recipe"
    p = project_env.member.post(
        "/api/v1/global-memories/propose",
        {"key": key, "value": "全局通用约定", "note": "集成测试提案"},
    )
    assert p["id"]
    # 未决防重
    with pytest.raises(BizError):
        project_env.member.post(
            "/api/v1/global-memories/propose", {"key": key, "value": "again", "note": ""}
        )
    # 管理员列表可见
    rows = admin.get("/api/v1/global-memories/proposals")["list"]
    assert any(x["id"] == p["id"] for x in rows)
    # 采纳
    admin.post(f"/api/v1/global-memories/proposals/{p['id']}/review", {"decision": "approved"})
    got = project_env.member.get(f"/api/v1/global-memories/{key}")
    assert "全局通用约定" in str(got)
    # 正式占用后同名提案被拒
    with pytest.raises(BizError):
        project_env.member.post(
            "/api/v1/global-memories/propose", {"key": key, "value": "x", "note": ""}
        )
    # 重复审核被拒
    with pytest.raises(BizError):
        admin.post(f"/api/v1/global-memories/proposals/{p['id']}/review", {"decision": "approved"})


def test_global_memory_reject_requires_reason(admin, project_env):
    p = project_env.member.post(
        "/api/v1/global-memories/propose",
        {"key": "itest-global-rejected", "value": "不会被采纳", "note": ""},
    )
    with pytest.raises(BizError):
        admin.post(f"/api/v1/global-memories/proposals/{p['id']}/review", {"decision": "rejected"})
    admin.post(
        f"/api/v1/global-memories/proposals/{p['id']}/review",
        {"decision": "rejected", "reason": "价值不足"},
    )
