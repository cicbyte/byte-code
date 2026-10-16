"""agent 注册来源 IP 审计：注册落库 + 管理列表回显。"""

import pytest

from apiclient import Client

pytestmark = pytest.mark.order(4)


def test_register_ip_recorded_and_listed(admin, backend, data):
    """自助注册的 agent 带来源 IP；管理端创建的为空。"""
    name = f'{data["users"]["agent"]["username"]}-ip-audit'
    agent_id, _ = Client.register_agent(backend.base_url, name)
    assert agent_id > 0

    rows = admin.get("/api/v1/ai-users", params={"page": 1, "size": 100})["list"]
    row = next(x for x in rows if x["id"] == agent_id)
    assert row["registerIp"] in ("127.0.0.1", "::1"), f"自助注册应记录来源 IP：{row['registerIp']!r}"

    admin_created = admin.post("/api/v1/ai-users", {"username": f"{name}-admin", "realName": "", "capabilities": ""})
    row2 = next(x for x in admin.get("/api/v1/ai-users", params={"page": 1, "size": 100})["list"] if x["id"] == admin_created["id"])
    assert row2["registerIp"] == ""
