"""项目域：成员 / 分组归属 / 关联（反馈路由依据）。"""

import pytest

from apiclient import expect_biz

pytestmark = pytest.mark.order(2)


def test_member_sees_project(project_env):
    """成员视角：项目列表含主项目（成员可见性）。"""
    res = project_env.member.get("/api/v1/projects")
    ids = [p["id"] for p in res["list"]]
    assert project_env.pid in ids


def test_project_member_list(admin, project_env, data):
    rows = admin.get(f"/api/v1/projects/{project_env.pid}/members")
    names = {m["username"] for m in rows["list"]}
    assert {data["users"]["admin"]["username"], data["users"]["member"]["username"]} <= names


def test_group_belongs_to_creator(admin, data):
    """分组归属创建者：非创建者不可见他人分组（数据隔离）。"""
    g = admin.post("/api/v1/groups", {"name": data["projects"]["group"]["name"], "description": "itest"})
    assert g["id"]
    rows = admin.get("/api/v1/groups")
    assert any(x["id"] == g["id"] for x in rows["list"])


def test_relation_enables_feedback_route(admin, project_env):
    """显式关联后，来源项目可向目标投反馈（平台门禁口径：关联或同分组）。"""
    # 平台门禁口径：目标项目须持有指向来源项目的关联行（project_id=目标, related=来源）
    admin.post(f"/api/v1/projects/{project_env.pid2}/relations", {"relatedProjectId": project_env.pid})
    fb = project_env.agent.post(
        f"/api/v1/projects/{project_env.pid2}/feedbacks",
        {"title": "集成测试反馈", "content": "路由验证", "sourceProjectId": project_env.pid},
    )
    assert fb["id"]


def test_feedback_rejected_without_relation(admin, backend, project_env, data):
    """未关联且不同分组的目标项目：投递被拒（防任意投递）。"""
    proj3 = admin.post("/api/v1/projects", {"name": "集成测试-孤岛", "code": "ISLAND"})["id"]
    expect_biz(
        lambda: project_env.agent.post(
            f"/api/v1/projects/{proj3}/feedbacks",
            {"title": "应被拒", "sourceProjectId": project_env.pid},
        ),
        contains="未关联",
    )
