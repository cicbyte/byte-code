"""任务 watcher（订阅扇出）+ 我的效率统计。"""

import pytest

from apiclient import BizError

pytestmark = pytest.mark.order(2)


@pytest.fixture(scope="module")
def watch_task(admin, project_env):
    tid = admin.post(
        f"/api/v1/projects/{project_env.pid}/tasks", {"title": "观察者目标", "type": "feature"}
    )["id"]

    class NS:
        pass

    ns = NS()
    ns.tid = tid
    return ns


def _flatten_tree(nodes):
    out = []
    for n in nodes or []:
        out.append(n)
        out.extend(_flatten_tree(n.get("children")))
    return out


def test_watch_enriches_detail(admin, project_env, watch_task):
    """只读语义：agent 可关注；详情回填 watching/watcherCount/watchers。"""
    project_env.agent.post(f"/api/v1/tasks/{watch_task.tid}/watch")
    det = project_env.agent.get(f"/api/v1/tasks/{watch_task.tid}")
    assert det["watching"] is True
    assert det["watcherCount"] == 1
    assert det["watchers"]


def test_watch_fans_out_on_comment(admin, project_env, watch_task):
    """评论事件向关注者扇出：agent 通知中心出现该任务评论通知。"""
    before = project_env.agent.get("/api/v1/notifications/unread-count")
    admin.post(f"/api/v1/tasks/{watch_task.tid}/comments", {"content": " watchers 应收到这条"})
    notes = project_env.agent.get("/api/v1/notifications")["list"]
    assert any(n.get("sourceId") == watch_task.tid or "评论" in str(n.get("title", "")) for n in notes)
    after = project_env.agent.get("/api/v1/notifications/unread-count")
    assert after.get("count", 0) >= before.get("count", 0)


def test_unwatch_stops_subscription(admin, project_env, watch_task):
    project_env.agent.post(f"/api/v1/tasks/{watch_task.tid}/unwatch")
    det = project_env.agent.get(f"/api/v1/tasks/{watch_task.tid}")
    assert det["watching"] is False and det["watcherCount"] == 0


def test_my_task_stats_shape(admin, project_env):
    """我的效率：结构完整 + admin 自闭环（认领→完成→过审）后 completed30d 生效。

    注：agent 视角统计恒 0——作用域限定 project_members，agent 走
    agent_project_bindings 不在成员表（行为观察，已反馈平台任务）。"""
    tid = admin.post(
        f"/api/v1/projects/{project_env.pid}/tasks", {"title": "效率统计任务", "type": "chore"}
    )["id"]
    admin.post(f"/api/v1/tasks/{tid}/claim")
    admin.post(f"/api/v1/tasks/{tid}/complete", {"note": "完成"})
    admin.post("/api/v1/reviews/batch", {"ids": [tid], "status": "approved", "comment": "过审"})
    res = admin.get("/api/v1/my-tasks/stats")
    for k in ("statusCounts", "activeTotal", "overdue", "completed30d", "trend", "avgLeadHours", "byProject"):
        assert k in res, k
    assert res["completed30d"] >= 1
