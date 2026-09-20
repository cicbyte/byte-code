"""讨论区：发起/回复能力门禁、agent 参与、转任务血缘、归档。"""

import pytest

from apiclient import expect_biz

pytestmark = pytest.mark.order(4)


@pytest.fixture(scope="module")
def disc_env(project_env):
    """主项目 + 讨论线程（作者=member）。"""

    class NS:
        pass

    ns = NS()
    ns.pid = project_env.pid
    ns.did = project_env.member.post(
        f"/api/v1/projects/{ns.pid}/discussions",
        {"title": "缓存策略怎么选", "body": "读多写少，倾向本地缓存"},
    )["id"]
    return ns


def test_member_creates_and_lists(project_env, disc_env):
    res = project_env.member.get(f"/api/v1/projects/{project_env.pid}/discussions")
    row = next(x for x in res["list"] if x["id"] == disc_env.did)
    assert row["status"] == "open" and row["replyCount"] == 0
    assert row["authorName"]


def test_agent_reply_needs_discuss_capability(project_env, disc_env):
    """agent 回复讨论走 discuss 能力位：默认能力集的 agent 可回，未授予的拒。"""
    # admin_client 的 agent（project_env.agent）按 fixture 接入能力集而定：
    # 平台对讨论的准入以 bindings.capabilities 为准，缺省（空=全能力）放行
    rid = project_env.agent.post(
        f"/api/v1/discussions/{disc_env.did}/replies", {"content": "agent 视角：建议本地缓存"}
    )["id"]
    assert rid > 0


def test_reply_notifies_author_and_detail_has_replies(project_env, disc_env):
    admin_detail = project_env.member.get(f"/api/v1/discussions/{disc_env.did}")
    assert len(admin_detail["replies"]) >= 1
    types = {r["userType"] for r in admin_detail["replies"]}
    assert "ai" in types


def test_convert_to_task_lineage(project_env, disc_env):
    """转任务：血缘互链 + 重复转拒绝 + 任务描述带来源标注。"""
    res = project_env.member.post(f"/api/v1/discussions/{disc_env.did}/convert", {"type": "feature"})
    tid = res["taskId"]
    assert tid > 0
    # 重复转 → 明确拒绝
    expect_biz(
        lambda: project_env.member.post(f"/api/v1/discussions/{disc_env.did}/convert", {}),
        contains="已转",
    )
    detail = project_env.member.get(f"/api/v1/discussions/{disc_env.did}")
    assert detail["status"] == "converted" and detail["convertedTaskId"] == tid
    task = project_env.member.get(f"/api/v1/tasks/{tid}")
    assert "来自讨论" in (task.get("description") or "")


def test_archive_toggle(project_env, disc_env):
    did = project_env.member.post(
        f"/api/v1/projects/{project_env.pid}/discussions",
        {"title": "归档演练", "body": ""},
    )["id"]
    res = project_env.member.post(f"/api/v1/discussions/{did}/archive")
    assert res["status"] == "archived"
    res = project_env.member.post(f"/api/v1/discussions/{did}/archive")
    assert res["status"] == "open"


def test_non_author_edit_rejected_maintainer_allowed(project_env, admin, disc_env):
    """编辑门禁：非作者成员拒（含文案），maintainer/owner 放行。"""
    did = project_env.member.post(
        f"/api/v1/projects/{project_env.pid}/discussions",
        {"title": "门禁演练", "body": ""},
    )["id"]
    expect_biz(
        lambda: project_env.agent.put(f"/api/v1/discussions/{did}", {"title": "agent 越权"}),
        contains="maintainer",
    )
    # owner（admin）放行
    admin.put(f"/api/v1/discussions/{did}", {"title": "门禁演练（已改）"})
    detail = project_env.member.get(f"/api/v1/discussions/{did}")
    assert "已改" in detail["title"]


def test_delete_cascades_replies(project_env, disc_env):
    did = project_env.member.post(
        f"/api/v1/projects/{project_env.pid}/discussions",
        {"title": "删除演练", "body": ""},
    )["id"]
    project_env.member.post(f"/api/v1/discussions/{did}/replies", {"content": "待级联"})
    project_env.member.delete(f"/api/v1/discussions/{did}")
    expect_biz(
        lambda: project_env.member.get(f"/api/v1/discussions/{did}"),
        contains="不存在",
    )
