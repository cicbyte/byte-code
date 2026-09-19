"""测试执行记录链：上报（汇总服务端重算）→ flaky 检测 → 趋势 → 失败转缺陷。"""

import pytest

from apiclient import expect_biz

pytestmark = pytest.mark.order(4)


@pytest.fixture(scope="module")
def runs_reported(project_env, data):
    """按 testruns.yaml 依次上报三次执行（构造 flaky/恒败/恒过）。"""
    cfg = data["testruns"]
    run_ids = []
    for spec in cfg["runs"]:
        cases = [
            {
                "externalKey": c["key"],
                "title": c["key"].rsplit("::", 1)[-1],
                "status": c["status"],
                "durationMs": c["duration"],
                "message": c.get("message", ""),
            }
            for c in spec["cases"]
        ]
        rid = project_env.agent.post(
            f"/api/v1/projects/{project_env.pid}/test-runs",
            {
                "source": "pytest",
                "branch": cfg["branch"],
                "env": cfg["env"],
                "cases": cases,
            },
        )["id"]
        run_ids.append(rid)
    return run_ids


def test_report_recomputes_summary_server_side(admin, project_env, runs_reported):
    """服务端重算汇总：末次（ok 过 + flip 复过 + always 败 = 2 过 1 败）正确。"""
    last = admin.get(f"/api/v1/test-runs/{runs_reported[-1]}")
    assert (last["total"], last["passed"], last["failed"]) == (3, 2, 1)


def test_flaky_detection(admin, project_env, runs_reported, data):
    flaky = admin.get(f"/api/v1/projects/{project_env.pid}/test-runs/flaky", params={"window": 10})
    keys = {x["externalKey"] for x in flaky["list"]}
    for k in data["testruns"]["expect"]["flaky_keys"]:
        assert k in keys
    assert "tests/test_a.py::test_ok" not in keys          # 恒过不标
    assert "tests/test_bad.py::test_always" not in keys    # 恒败不标


def test_trends_and_top_failed(admin, project_env, runs_reported, data):
    tr = admin.get(f"/api/v1/projects/{project_env.pid}/test-runs/trends", params={"limit": 10})
    assert len(tr["runs"]) >= 3
    assert tr["topFailed"][0]["externalKey"] == data["testruns"]["expect"]["top_failed_first"]


def test_failure_to_bug_closes_loop(admin, project_env, runs_reported, data):
    """失败行一键转缺陷：建 bug 任务 + 挂钩 + 重复转明确拒绝。"""
    det = admin.get(f"/api/v1/test-runs/{runs_reported[-1]}")
    bad = next(c for c in det["cases"] if c["externalKey"].endswith("test_always") and c["status"] == "fail")
    res = admin.post(f"/api/v1/test-run-cases/{bad['id']}/bug", {"title": data["testruns"]["expect"]["bug_title"]})
    assert res["created"] and res["taskId"]

    task = admin.get(f"/api/v1/tasks/{res['taskId']}")
    assert task["type"] == "bug" and task["status"] == "open"

    expect_biz(
        lambda: admin.post(f"/api/v1/test-run-cases/{bad['id']}/bug", {}),
        contains="已挂接",
    )


def test_readonly_agent_cannot_report(project_env):
    """test_execute 能力门禁：只读 agent 上报被拒。"""
    expect_biz(
        lambda: project_env.ro_agent.post(
            f"/api/v1/projects/{project_env.pid}/test-runs",
            {"source": "pytest", "cases": [{"externalKey": "x::y", "status": "pass"}]},
        ),
        contains="上报测试执行",
    )
