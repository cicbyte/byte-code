"""任务全链：创建（checklist）→ 认领 → 完成 → 审核（order 依赖 task_chain/review_env）。

门禁负路径：能力集（只读 agent）、会话项目约束。
"""

import json

import pytest

from apiclient import Client, expect_biz

pytestmark = pytest.mark.order(3)


def test_task_created_with_checklist(admin, project_env, task_chain, data):
    detail = admin.get(f"/api/v1/tasks/{task_chain.task_id}")
    t = data["tasks"]["tasks"][0]
    assert detail["title"] == t["title"]
    items = json.loads(detail.get("checklist") or "[]")  # 平台返回 JSON 字符串
    assert [c["text"] for c in items] == [c["text"] for c in t["checklist"]]


def test_task_flow_status_lifecycle(admin, project_env, review_env):
    """task_chain + review_env 串联后的终态：done。"""
    detail = admin.get(f"/api/v1/tasks/{review_env.task_id}")
    assert detail["status"] == "done"


def test_project_member_can_claim(admin, project_env, data):
    """正路径：项目成员可认领未指派 open 任务。"""
    t = data["tasks"]["tasks"][1]
    tid = admin.post(
        f"/api/v1/projects/{project_env.pid}/tasks",
        {"title": t["title"], "type": t["type"], "priority": t["priority"]},
    )["id"]
    project_env.member.post(f"/api/v1/tasks/{tid}/claim")
    detail = admin.get(f"/api/v1/tasks/{tid}")
    assert detail["status"] == "in_progress"


def test_outsider_cannot_claim(admin, backend, project_env, data):
    """非项目成员（viewer 未被加入）越权认领被项目归属校验拒绝。"""
    v = data["users"]["viewer"]
    admin.post(
        "/api/v1/admin/users",
        {"username": v["username"], "password": v["password"], "realName": v["real_name"]},
    )
    outsider = Client.login(backend.base_url, v["username"], v["password"])
    tid = admin.post(
        f"/api/v1/projects/{project_env.pid}/tasks",
        {"title": "集成测试-越权认领", "type": "feature", "priority": 3},
    )["id"]
    expect_biz(lambda: outsider.post(f"/api/v1/tasks/{tid}/claim"), contains="无权限访问该项目资源")


def test_readonly_agent_blocked_on_claim(admin, project_env):
    """只读能力 agent（tasks_read,docs_read）认领被拒（能力门禁文案）。"""
    tid = admin.post(
        f"/api/v1/projects/{project_env.pid}/tasks",
        {"title": "集成测试-能力门禁", "type": "feature", "priority": 3},
    )["id"]
    expect_biz(lambda: project_env.ro_agent.post(f"/api/v1/tasks/{tid}/claim"), contains="写任务")


def test_agent_can_comment_and_admin_sees(project_env, admin, task_chain, data):
    """agent 评论入流，admin 可读（IM 同流）。"""
    detail = admin.get(f"/api/v1/tasks/{task_chain.task_id}")
    assert detail["status"] == "done"
    # done 任务评论仍开放
    tid2 = admin.post(
        f"/api/v1/projects/{project_env.pid}/tasks",
        {"title": "集成测试-评论载体", "type": "chore", "priority": 3},
    )["id"]
    project_env.agent.post(f"/api/v1/tasks/{tid2}/comments", {"content": data["tasks"]["comment"]})
    rows = admin.get(f"/api/v1/tasks/{tid2}/comments")
    assert any(data["tasks"]["comment"] in (c.get("content") or "") for c in rows["list"])
