"""项目发布（Releases，#551/#552）：独立文件体系/门禁/大文件/分享直链/级联。"""

import pytest

from apiclient import BizError, expect_biz

pytestmark = pytest.mark.order(2)


def _upload(client, rid, name="app.zip", content=b"pkg-bytes"):
    r = client.http.post(
        f"/api/v1/releases/{rid}/files",
        files={"file": (name, content, "application/zip")},
    )
    return client._unwrap(r)


@pytest.fixture(scope="module")
def rel_world(admin, project_env, data):
    pid = admin.post("/api/v1/projects", {"name": "集成测试-发布", "code": "RELT"})["id"]
    # member 入项：负路径打到业务门禁（maintainer 档）而非中间件 403
    uid = next(
        u["id"]
        for u in admin.get(
            "/api/v1/users/search", params={"q": data["users"]["member"]["username"]}
        )["list"]
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
    """创建收 maintainer：member 拒；同版本重复明确拒。"""
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


def test_upload_gates_and_same_name_rejected(admin, project_env, rel_world):
    """上传收 maintainer；版本内同名拒传；>20MB 文件放行（#552 上限 2GB）。"""
    fid = _upload(admin, rel_world.rid, name="app-1.0.0.zip", content=b"A" * 1024)["id"]
    assert fid
    expect_biz(
        lambda: _upload(project_env.member, rel_world.rid, name="m.zip"),
        contains="管理员",
    )
    expect_biz(
        lambda: _upload(admin, rel_world.rid, name="app-1.0.0.zip", content=b"B"),
        contains="已存在",
    )
    # 21MB：证明发布通道不受附件 20MB 上限约束
    big = _upload(admin, rel_world.rid, name="big-installer.bin", content=b"\0" * (21 * 1024 * 1024))
    assert big["id"]


def test_list_files_and_aggregates(admin, rel_world):
    rows = admin.get(f"/api/v1/projects/{rel_world.pid}/releases")["list"]
    hit = next(r for r in rows if r["id"] == rel_world.rid)
    assert hit["fileCount"] >= 2 and hit["totalSizeBytes"] >= 21 * 1024 * 1024
    assert hit["createdByName"]  # 署名回填
    files = admin.get(f"/api/v1/releases/{rel_world.rid}/files")["list"]
    names = {f["fileName"] for f in files}
    assert {"app-1.0.0.zip", "big-installer.bin"} <= names


def test_share_public_link_noauth_and_revoke(admin, backend, rel_world):
    """直链：免鉴权可下、字节一致、幂等、下载计数、吊销即失效。"""
    import httpx

    # 自建文件（不依赖其它测试的执行顺序）
    fid = _upload(admin, rel_world.rid, name="share-demo.zip", content=b"S" * 512)["id"]
    sh = admin.post(f"/api/v1/release-files/{fid}/share", {"days": 0})
    assert sh["url"].startswith("/api/v1/release-files/public/")
    sh2 = admin.post(f"/api/v1/release-files/{fid}/share", {"days": 0})
    assert sh2["url"] == sh["url"]  # 幂等复用
    # 免鉴权下载（裸 httpx，无任何头）
    r = httpx.get(f"{backend.base_url}{sh['url']}", timeout=30)
    assert r.status_code == 200 and r.content == b"S" * 512
    assert "share-demo.zip" in (r.headers.get("content-disposition") or "")
    files = admin.get(f"/api/v1/releases/{rel_world.rid}/files")["list"]
    assert next(f for f in files if f["id"] == fid)["downloadCount"] == 1
    admin.delete(f"/api/v1/release-files/{fid}/share")
    r2 = httpx.get(f"{backend.base_url}{sh['url']}", timeout=10)
    assert r2.status_code == 404


def test_update_delete_and_cascade(admin, project_env, rel_world):
    """member 改/删拒；更新生效；删文件；删发布级联清文件行。"""
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

    files = admin.get(f"/api/v1/releases/{rel_world.rid}/files")["list"]
    victim = next(f for f in files if f["fileName"] == "big-installer.bin")
    expect_biz(
        lambda: project_env.member.delete(f"/api/v1/release-files/{victim['id']}"), contains="管理员"
    )
    admin.delete(f"/api/v1/release-files/{victim['id']}")
    files2 = admin.get(f"/api/v1/releases/{rel_world.rid}/files")["list"]
    assert all(f["id"] != victim["id"] for f in files2)

    expect_biz(lambda: project_env.member.delete(f"/api/v1/releases/{rel_world.rid}"), contains="管理员")
    admin.delete(f"/api/v1/releases/{rel_world.rid}")
    rows2 = admin.get(f"/api/v1/projects/{rel_world.pid}/releases")["list"]
    assert all(r["id"] != rel_world.rid for r in rows2)
