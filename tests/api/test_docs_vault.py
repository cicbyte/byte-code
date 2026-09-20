"""文档 vault（人类路径 + agent 免参别名 + 能力门禁）与项目记忆。"""

import pytest

from apiclient import BizError

pytestmark = pytest.mark.order(2)

PATH_API = "itest/api-written.md"
PATH_AGENT = "itest/agent-written.md"
NEEDLE = "itest-搜索锚点"


def _flatten(nodes):
    out = []
    for n in nodes or []:
        out.append(n)
        out.extend(_flatten(n.get("children")))
    return out


def test_human_write_read_roundtrip(admin, project_env):
    admin.put(
        f"/api/v1/projects/{project_env.pid}/docs/file",
        {"path": PATH_API, "content": f"# API 写入\n{NEEDLE}\n"},
    )
    got = admin.get(f"/api/v1/projects/{project_env.pid}/docs/file", params={"path": PATH_API})
    assert NEEDLE in got["content"] and got["path"] == PATH_API
    tree = _flatten(admin.get(f"/api/v1/projects/{project_env.pid}/docs/tree")["tree"])
    assert any(n["path"] == PATH_API for n in tree)
    hits = admin.get(
        f"/api/v1/projects/{project_env.pid}/docs/search", params={"q": NEEDLE}
    )["items"]
    assert any(h["path"] == PATH_API for h in hits)


def test_agent_alias_write_and_read(project_env):
    """agent 免参别名：X-Session 推导项目；写人类可读，树互通。"""
    project_env.agent.put("/api/v1/agent/docs/file", {"path": PATH_AGENT, "content": "agent 写入"})
    got = project_env.agent.get("/api/v1/agent/docs/file", params={"path": PATH_AGENT})
    assert "agent 写入" in got["content"]
    human_view = project_env.member.get(
        f"/api/v1/projects/{project_env.pid}/docs/file", params={"path": PATH_AGENT}
    )
    assert "agent 写入" in human_view["content"]
    # 别名端点树响应用 list 键（CLI 同口径）
    tree = _flatten(project_env.agent.get("/api/v1/agent/docs/tree")["list"])
    assert any(n["path"] == PATH_API for n in tree)


def test_readonly_agent_docs_write_rejected(project_env):
    """只读 agent（docs_read）：读通、写拒。"""
    got = project_env.ro_agent.get("/api/v1/agent/docs/file", params={"path": PATH_API})
    assert NEEDLE in got["content"]
    with pytest.raises(BizError):
        project_env.ro_agent.put(
            "/api/v1/agent/docs/file", {"path": "itest/ro.md", "content": "不该成功"}
        )


def test_project_memory_roundtrip(admin, project_env):
    """项目记忆 upsert/读取/列表 + 只读 agent 读通写拒。"""
    admin.put(
        f"/api/v1/projects/{project_env.pid}/memories/itest-key",
        {"value": "记忆值v1", "status": "active"},
    )
    got = admin.get(f"/api/v1/projects/{project_env.pid}/memories/itest-key")
    assert "记忆值v1" in str(got)
    rows = admin.get(f"/api/v1/projects/{project_env.pid}/memories")["list"]
    assert any(r["key"] == "itest-key" for r in rows)
    # 只读 agent 能力集是 tasks_read,docs_read——不含 memory_read：读写皆拒
    with pytest.raises(BizError):
        project_env.ro_agent.get(f"/api/v1/projects/{project_env.pid}/memories/itest-key")
    with pytest.raises(BizError):
        project_env.ro_agent.put(
            f"/api/v1/projects/{project_env.pid}/memories/itest-key-ro",
            {"value": "不该成功"},
        )
