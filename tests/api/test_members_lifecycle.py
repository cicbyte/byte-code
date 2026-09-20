"""成员域全流程：用户搜索、三档角色边界、自助退出、项目归档只读、移交邀请制。

统一在专属项目 proj4 上操作，不污染共享世界（project_env 的 pid/pid2）。
"""

import pytest

from apiclient import BizError, Client, expect_biz

pytestmark = pytest.mark.order(2)


@pytest.fixture(scope="module")
def life_world(admin, backend, project_env, data):
    """专属项目 + 三类账号：owner(admin) / maintainer / member / 门外汉。"""
    pid = admin.post("/api/v1/projects", {"name": "集成测试-生命周期", "code": "LIFE"})["id"]
    mt = admin.post(
        "/api/v1/admin/users",
        {"username": "itest-mt", "password": "Mt@123456", "realName": "维护者"},
    )["id"]
    mb = admin.post(
        "/api/v1/admin/users",
        {"username": "itest-mb", "password": "Mb@123456", "realName": "普通成员"},
    )["id"]
    admin.post(f"/api/v1/projects/{pid}/members", {"userId": mt, "role": "maintainer"})
    admin.post(f"/api/v1/projects/{pid}/members", {"userId": mb, "role": "member"})
    admin.post(
        "/api/v1/admin/users",
        {"username": "itest-stranger", "password": "Str@123456", "realName": "陌生人"},
    )
    stranger = Client.login(backend.base_url, "itest-stranger", "Str@123456")
    member = Client.login(backend.base_url, "itest-mb", "Mb@123456")
    mt_c = Client.login(backend.base_url, "itest-mt", "Mt@123456")

    class NS:
        pass

    ns = NS()
    ns.pid = pid
    ns.mt_id = mt
    ns.mb_id = mb
    ns.mt = mt_c
    ns.member = member
    ns.stranger = stranger
    return ns


def test_user_search_finds_human_only(admin, life_world, data):
    rows = admin.get("/api/v1/users/search", params={"q": "itest-m"})["list"]
    names = {r["username"] for r in rows}
    assert {"itest-mt", "itest-mb"} <= names
    # agent 账号不出现在人类搜索（选择器只供人类成员）
    assert "itest-agent" not in names


def test_user_search_rejects_agent(project_env):
    """用户目录不给 agent：控制器层拒绝（业务错误而非 HTTP 403）。"""
    with pytest.raises(BizError):
        project_env.agent.get("/api/v1/users/search", params={"q": "itest"})


def test_maintainer_can_manage_members_but_not_owner(life_world, admin, backend):
    """三档制：maintainer 可加人/移除普通成员；不可动 owner；member 无管理权。"""
    # maintainer 加一个新成员
    uid = admin.post(
        "/api/v1/admin/users",
        {"username": "itest-extra", "password": "Ex@123456", "realName": "后加成员"},
    )["id"]
    life_world.mt.post(f"/api/v1/projects/{life_world.pid}/members", {"userId": uid, "role": "member"})
    # maintainer 可移除普通成员
    life_world.mt.delete(f"/api/v1/projects/{life_world.pid}/members/{uid}")
    # owner（admin=1）不可被移除
    with pytest.raises(BizError):
        life_world.mt.delete(f"/api/v1/projects/{life_world.pid}/members/1")
    # 普通成员不能移除 maintainer
    with pytest.raises(BizError):
        life_world.member.delete(f"/api/v1/projects/{life_world.pid}/members/{life_world.mt_id}")


def test_member_leave_then_lose_access(life_world):
    """自助退出：member 可退；退出后失去访问。"""
    life_world.member.post(f"/api/v1/projects/{life_world.pid}/members/leave")
    with pytest.raises(BizError):
        life_world.member.get(f"/api/v1/projects/{life_world.pid}")


def test_owner_leave_rejected(admin, life_world):
    from apiclient import expect_biz

    expect_biz(
        lambda: admin.post(f"/api/v1/projects/{life_world.pid}/members/leave"),
        contains="",
    )


def test_archive_readonly_then_restore(admin, life_world):
    """归档：读放行、写拒绝；解档恢复。"""
    admin.put(f"/api/v1/projects/{life_world.pid}", {"status": 3})
    # 成员读 OK
    rows = admin.get(f"/api/v1/projects/{life_world.pid}/members")
    assert rows["list"]
    # 写拒绝（maintainer 建任务）
    with pytest.raises(BizError):
        life_world.mt.post(f"/api/v1/projects/{life_world.pid}/tasks", {"title": "归档中建任务", "type": "feature"})
    # 解档恢复写
    admin.put(f"/api/v1/projects/{life_world.pid}", {"status": 1})
    tid = life_world.mt.post(
        f"/api/v1/projects/{life_world.pid}/tasks", {"title": "解档后建任务", "type": "feature"}
    )["id"]
    assert tid


def test_owner_transfer_invite_flow(admin, life_world, data):
    """移交邀请制：发起 → 目标收通知 → 非目标拒答 → 接受换 owner（原 owner 留任降级）。"""
    # member 退出过项目：先把目标用户加回来（接受路径会自动入项，这里用 maintainer 做目标）
    invite = admin.post(
        f"/api/v1/projects/{life_world.pid}/owner-transfer", {"userId": life_world.mt_id}
    )
    tid = invite["id"]
    # 同项目单 pending 防重
    with pytest.raises(BizError):
        admin.post(f"/api/v1/projects/{life_world.pid}/owner-transfer", {"userId": life_world.mb_id})
    # 非受邀人拒答
    with pytest.raises(BizError):
        life_world.stranger.post(f"/api/v1/owner-transfers/{tid}", {"action": "accept"})
    # 目标收到通知
    notes = life_world.mt.get("/api/v1/notifications")["list"]
    assert any(str(tid) in str(n.get("sourceId", "")) or n.get("type", "").find("transfer") >= 0 for n in notes)
    # 接受：mt 变 owner，admin 降级为成员（leave 缺省 false）
    life_world.mt.post(f"/api/v1/owner-transfers/{tid}", {"action": "accept"})
    members = admin.get(f"/api/v1/projects/{life_world.pid}/members")["list"]
    roles = {m["username"]: m["role"] for m in members}
    assert roles.get("itest-mt") == "owner"
    assert roles.get("admin") in ("member", "maintainer")
    # 已决邀请再响应被拒
    with pytest.raises(BizError):
        life_world.mt.post(f"/api/v1/owner-transfers/{tid}", {"action": "decline"})
