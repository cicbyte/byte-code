"""附件域：多实体类型上传/列表/下载回环 + 归属门禁（外人拒、非法类型拒）。"""

import pytest

from apiclient import BizError, Client, expect_biz

pytestmark = pytest.mark.order(2)


def upload(client, entity_type, entity_id=0, name="attach.txt", content=b"attachment-bytes"):
    r = client.http.post(
        "/api/v1/attachments/upload",
        data={"entityType": entity_type, "entityId": str(entity_id)},
        files={"file": (name, content, "text/plain")},
    )
    return client._unwrap(r)


def test_invalid_entity_type_rejected(admin, task_chain):
    with pytest.raises(BizError):
        upload(admin, "bogus-type", task_chain.task_id)


def test_task_attachment_roundtrip(admin, task_chain):
    aid = upload(admin, "task", task_chain.task_id, name="roundtrip.txt")["id"]
    rows = admin.get(
        "/api/v1/attachments", params={"entityType": "task", "entityId": task_chain.task_id}
    )["list"]
    assert any(r["id"] == aid and r["originalName"] == "roundtrip.txt" for r in rows)
    # 下载走 URL 接口再取流（本地存储直读），字节一致
    dl = admin.get(f"/api/v1/attachments/{aid}/download")
    assert dl.get("url")
    content = admin.http.get(dl["url"]).content
    assert content == b"attachment-bytes"


def test_outsider_cannot_read_project_attachment(admin, backend, task_chain):
    """非成员对项目内附件：归属门禁拒绝（附件是项目内资产）。"""
    aid = upload(admin, "task", task_chain.task_id, name="secret.txt")["id"]
    admin.post(
        "/api/v1/admin/users",
        {"username": "itest-outsider", "password": "Outsider@1", "realName": "外部用户"},
    )
    outsider = Client.login(backend.base_url, "itest-outsider", "Outsider@1")
    with pytest.raises(BizError):
        outsider.get(f"/api/v1/attachments/{aid}/download")


def test_run_case_attachment_two_hop(admin, project_env):
    """test_run_case 附件：归属两跳解析（用例行→run→项目）。"""
    rid = admin.post(
        f"/api/v1/projects/{project_env.pid}/test-runs",
        {"source": "manual", "cases": [{"externalKey": "t::attach", "status": "fail", "message": "boom"}]},
    )["id"]
    det = admin.get(f"/api/v1/test-runs/{rid}")
    row_id = det["cases"][0]["id"]
    aid = upload(admin, "test_run_case", row_id, name="fail.png", content=b"png-bytes")["id"]
    rows = admin.get(
        "/api/v1/attachments", params={"entityType": "test_run_case", "entityId": row_id}
    )["list"]
    assert any(r["id"] == aid for r in rows)


def test_owner_delete_attachment(admin, task_chain):
    aid = upload(admin, "task", task_chain.task_id, name="todel.txt")["id"]
    admin.delete(f"/api/v1/attachments/{aid}")
    rows = admin.get(
        "/api/v1/attachments", params={"entityType": "task", "entityId": task_chain.task_id}
    )["list"]
    assert all(r["id"] != aid for r in rows)
