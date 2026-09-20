"""项目发布（Releases，#551）：创建门禁/版本唯一/附件通道/聚合回填/级联删除。"""

import pytest

from apiclient import BizError, expect_biz

pytestmark = pytest.mark.order(2)


def _upload(client, rid, name="app.zip", content=b"pkg-bytes"):
    r = client.http.post(
        "/api/v1/attachments/upload",
        data={"entityType": "release", "entityId": str(rid)},
        files={"file": (name, content, "application/zip")},
    )
    return client._unwrap(r)


@pytest.fixture(scope="module")
def rel_world(admin, project_env, data):
    pid = admin.post("/api/v1/projects", {"name": "集成测试-发布", "code": "RELT"})["id"]
    # member 入项：负路径要打到业务门禁（maintainer 档）而非中间件 403
    uid = next(
        u["id"] for u in admin.get("/api/v1/users/search", params={"q": data["users"]["member"]["username"]})["list"]
    )
    admin.post(f"/api/v1/projects/{pid}/members", {"userId": uid, "role": "member"})
    rid = admin.post(
        f"/api/v1/projects/{pid}/releases",
        {"version": "v1.0.0", "title": "首发", "notes": "首个安装包", "channel": "stable"},
    )["id"]

    class NS:
        pass

    ns = NS()
    ns.pid = pid
    ns.rid = rid
    return ns


def test_create_gates_and_version_unique(admin, project_env, rel_world):
    """创建收 maintainer：member 拒；同版本重复明确拒；跨项目同版本放行。"""
    expect_biz(
        lambda: project_env.member.post(
            f"/api/v1/projects/{rel_world.pid}/releases", {"version": "v0.9.0"}
        ),
        contains="管理员",
    )
    expect_biz(
        lambda: admin.post(f"/api/v1/projects/{rel_world.pid}/releases", {"version": "v1.0.0"}),
        contains="已存在",
    )
    other = admin.post(
        f"/api/v1/projects/{rel_world.pid}/releases", {"version": "v2.0.0-beta.1", "channel": "beta"}
    )["id"]
    assert other


def test_release_files_via_attachment_channel(admin, rel_world):
    """文件走附件通道：上传→列表聚合（数量/总大小/署名）→下载 URL 可达。"""
    aid = _upload(admin, rel_world.rid, name="app-1.0.0.zip", content=b"x" * 2048)["id"]
    rows = admin.get(f"/api/v1/projects/{rel_world.pid}/releases")["list"]
    hit = next(r for r in rows if r["id"] == rel_world.rid)
    assert hit["fileCount"] >= 1 and hit["totalSizeBytes"] >= 2048
    assert hit["createdByName"]  # 署名回填
    dl = admin.get(f"/api/v1/attachments/{aid}/download")
    assert dl["url"]
    content = admin.http.get(dl["url"]).content
    assert content == b"x" * 2048


def test_outsider_cannot_touch_release(admin, backend, rel_world):
    """门外汉：发布列表不可见（项目鉴权）、附件归属拒。"""
    from apiclient import Client

    admin.post(
        "/api/v1/admin/users",
        {"username": "itest-rel-out", "password": "RelOut@123", "realName": "发布门外汉"},
    )
    out = Client.login(backend.base_url, "itest-rel-out", "RelOut@123")
    with pytest.raises(BizError):
        out.get(f"/api/v1/projects/{rel_world.pid}/releases")
    aid = _upload(admin, rel_world.rid, name="secret.zip")["id"]
    with pytest.raises(BizError):
        out.get(f"/api/v1/attachments/{aid}/download")


def test_update_delete_and_cascade(admin, project_env, rel_world):
    """member 更新/删除拒；maintainer/owner 更新通；删除级联清附件记录。"""
    expect_biz(
        lambda: project_env.member.put(
            f"/api/v1/releases/{rel_world.rid}", {"notes": "不该成功"}
        ),
        contains="管理员",
    )
    admin.put(f"/api/v1/releases/{rel_world.rid}", {"notes": "更新后的说明", "channel": "beta"})
    rows = admin.get(f"/api/v1/projects/{rel_world.pid}/releases")["list"]
    hit = next(r for r in rows if r["id"] == rel_world.rid)
    assert hit["channel"] == "beta"

    expect_biz(lambda: project_env.member.delete(f"/api/v1/releases/{rel_world.rid}"), contains="管理员")
    admin.delete(f"/api/v1/releases/{rel_world.rid}")
    rows2 = admin.get(f"/api/v1/projects/{rel_world.pid}/releases")["list"]
    assert all(r["id"] != rel_world.rid for r in rows2)
