"""任务列表时间筛选：按 completed_at 复盘口径过滤 + completedAt 字段暴露。"""

import datetime

import pytest

from apiclient import expect_biz

pytestmark = pytest.mark.order(4)


@pytest.fixture(scope="module")
def tf_env(project_env, admin):
    pid = project_env.pid
    tid = project_env.member.post(
        f"/api/v1/projects/{pid}/tasks", {"title": "时间筛选-今日完成", "description": "", "type": "chore", "priority": 3}
    )["id"]
    project_env.agent.post(f"/api/v1/tasks/{tid}/claim")
    project_env.agent.post(f"/api/v1/tasks/{tid}/complete", {"artifacts": ""})
    admin.post("/api/v1/reviews/batch", {"ids": [tid], "status": "approved"})
    open_id = project_env.member.post(
        f"/api/v1/projects/{pid}/tasks", {"title": "时间筛选-未完成", "description": "", "type": "chore", "priority": 3}
    )["id"]

    class NS:
        pass

    ns = NS()
    ns.pid = pid
    ns.done_id = tid
    ns.open_id = open_id
    return ns


def test_filter_today_returns_completed_only(project_env, tf_env):
    today = datetime.date.today().isoformat()
    res = project_env.member.get(
        f"/api/v1/projects/{tf_env.pid}/tasks", params={"from": today, "to": today, "size": 200}
    )
    ids = {t["id"] for t in res["list"]}
    assert tf_env.done_id in ids
    assert tf_env.open_id not in ids, "未完成任务 completed_at 为空串，不应命中时间筛选"
    row = next(t for t in res["list"] if t["id"] == tf_env.done_id)
    assert row["completedAt"].startswith(today), "列表应暴露完成时刻"


def test_filter_empty_range(project_env, tf_env):
    """昨天区间无完成记录 → 空结果（其它用例的任务都完成于今天；
    ListTasks 空结果序列化为 null，与 [] 同义）。"""
    yest = (datetime.date.today() - datetime.timedelta(days=1)).isoformat()
    res = project_env.member.get(
        f"/api/v1/projects/{tf_env.pid}/tasks", params={"from": yest, "to": yest}
    )
    assert (res["list"] or []) == [] and res["total"] == 0


def test_filter_bad_format_rejected(project_env, tf_env):
    expect_biz(
        lambda: project_env.member.get(
            f"/api/v1/projects/{tf_env.pid}/tasks", params={"from": "2026/09/01", "to": "2026-09-21"}
        ),
        contains="格式",
    )


def test_filter_one_sided(project_env, tf_env):
    """只给 from：当天 00:00 起的全部已完成任务。"""
    today = datetime.date.today().isoformat()
    res = project_env.member.get(
        f"/api/v1/projects/{tf_env.pid}/tasks", params={"from": today, "size": 200}
    )
    assert tf_env.done_id in {t["id"] for t in res["list"]}
