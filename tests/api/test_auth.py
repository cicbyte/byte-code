"""认证链路：登录 / 首登强制改密 / agent key 换 token / 认证头三形态。"""

import pytest

from apiclient import Client, expect_biz

pytestmark = pytest.mark.order(1)


def test_admin_login_after_onboarding(backend, data):
    """backend fixture 已完成首登改密：新密码可登录且不再强制改密。"""
    users = data["users"]
    c = Client(backend.base_url)
    d = c.post("/api/login", {"username": users["admin"]["username"], "password": users["admin"]["password"]})
    assert d["token"]
    assert not d.get("mustChangePassword")


def test_wrong_password_rejected(backend, data):
    users = data["users"]
    c = Client(backend.base_url)
    expect_biz(
        lambda: c.post("/api/login", {"username": users["admin"]["username"], "password": "wrong-pass"}),
        contains="用户名或密码错误",
    )


def test_unauthenticated_rejected(backend):
    r = Client(backend.base_url).raw_get("/api/v1/projects")
    assert r.status_code == 401


def test_token_header_forms_equivalent(backend, data):
    """token 头与 Authorization Bearer 等效（extractToken 三段顺序）。"""
    users = data["users"]
    token = Client(backend.base_url).post(
        "/api/login", {"username": users["admin"]["username"], "password": users["admin"]["password"]}
    )["token"]
    a = Client(backend.base_url, token=token).get("/api/admin_info")
    b = Client(backend.base_url, bearer=token).get("/api/admin_info")
    assert a["username"] == b["username"] == users["admin"]["username"]


def test_agent_ai_login_swaps_key_for_token(backend, project_env, data):
    """bc_ key 除直连外也可经 /auth/ai/login 换人类同构 token。"""
    r = Client(backend.base_url).post("/api/auth/ai/login", {"apiKey": project_env.agent_key})
    assert r["token"]
