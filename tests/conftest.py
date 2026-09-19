"""根 conftest：后端/前端编排（session）+ YAML 数据装载。

依赖顺序三层纪律（详见 tests/README）：
1. 跨模块状态链 = fixture 图（backend → project_env → task_chain → review_env）
2. 模块内步骤序 = @pytest.mark.order(n)
3. 上游失败联动跳过 = @pytest.mark.dependency(depends=[...])
"""

from pathlib import Path

import pytest
import yaml

TESTS_DIR = Path(__file__).resolve().parent
HELPERS = TESTS_DIR / "helpers"
import sys

sys.path.insert(0, str(HELPERS))

from booted import boot as _boot_backend  # noqa: E402
from apiclient import Client  # noqa: E402
import json  # noqa: E402

DATA_DIR = TESTS_DIR / "data"


def pytest_addoption(parser):
    parser.addoption(
        "--ui-headed",
        action="store_true",
        default=False,
        help="UI 用例以有头浏览器运行（缺省 headless）",
    )


def load_yaml(name: str) -> dict:
    with open(DATA_DIR / f"{name}.yaml", encoding="utf-8") as f:
        return yaml.safe_load(f)


@pytest.fixture(scope="session")
def data():
    """全部 YAML 测试数据的只读聚合（users/projects/tasks/testruns/ui）"""
    names = [p.stem for p in DATA_DIR.glob("*.yaml")]
    return {n: load_yaml(n) for n in names}


@pytest.fixture(scope="session")
def backend(data):
    """干净后端：随机端口 + 独立 SQLite + 全量迁移 + admin 已过首登改密。"""
    users = data["users"]
    b = _boot_backend(
        admin_username=users["admin"]["username"],
        admin_initial=users["admin"]["initial_password"],
        admin_new=users["admin"]["password"],
    )
    yield b
    b.stop()


@pytest.fixture(scope="session")
def admin(backend, data):
    """管理员 API 客户端（人类 token 头认证）"""
    users = data["users"]
    return Client.login(backend.base_url, users["admin"]["username"], users["admin"]["password"])


# ---- 世界状态链（api/ui 共享；session 级只建一次）----
sys.path.insert(0, str(Path(__file__).resolve().parents[1] / "helpers"))

from apiclient import Client  # noqa: E402


@pytest.fixture(scope="session")
def project_env(admin, backend, data):
    """主项目 + 人类成员 + 全能 agent（会话已建）+ 只读 agent + 关联项目。

    命名空间字段：pid / pid2 / agent / agent_key / agent_id / ro_agent / member
    """
    proj = data["projects"]["projects"][0]
    pid = admin.post("/api/v1/projects", proj)["id"]

    users = data["users"]
    member = users["member"]
    member_uid = admin.post(
        "/api/v1/admin/users",
        {"username": member["username"], "password": member["password"], "realName": member["real_name"]},
    )["id"]
    admin.post(f"/api/v1/projects/{pid}/members", {"userId": member_uid, "role": "member"})

    # 全能 agent：注册 → owner 发接入码 → 凭码加入 → 建会话
    agent_id, key = Client.register_agent(backend.base_url, users["agent"]["username"])
    join_code = admin.post(f"/api/v1/projects/{pid}/agent-codes")["code"]
    Client.agent_join(backend.base_url, key, join_code)
    agent = Client.agent_session(backend.base_url, key, pid)

    # 只读 agent（能力集负路径用）：加入后由 owner 收窄能力
    ro_id, ro_key = Client.register_agent(backend.base_url, f'{users["agent"]["username"]}-ro')
    ro_join = admin.post(f"/api/v1/projects/{pid}/agent-codes")["code"]
    Client.agent_join(backend.base_url, ro_key, ro_join)
    admin.put(
        f"/api/v1/projects/{pid}/agents/{ro_id}/capabilities",
        {"capabilities": [c for c in users["agent"]["readonly_capabilities"].split(",") if c]},
    )
    ro_agent = Client.agent_session(backend.base_url, ro_key, pid)

    # 关联项目（反馈路由/分组场景）
    proj2 = data["projects"]["projects"][1]
    pid2 = admin.post("/api/v1/projects", proj2)["id"]

    class NS:
        pass

    ns = NS()
    ns.pid = pid
    ns.pid2 = pid2
    ns.agent = agent
    ns.agent_id = agent_id
    ns.agent_key = key
    ns.ro_agent = ro_agent
    ns.member = Client.login(backend.base_url, member["username"], member["password"])
    return ns


@pytest.fixture(scope="session")
def task_chain(admin, project_env, data):
    """任务状态链：feature（带 checklist）→ agent 认领 → 完成 → 进审核队列。"""
    t = data["tasks"]["tasks"][0]
    body = {
        "title": t["title"],
        "description": t["description"],
        "type": t["type"],
        "priority": t["priority"],
    }
    if t.get("checklist"):
        body["checklist"] = json.dumps(t["checklist"], ensure_ascii=False)
    tid = admin.post(f"/api/v1/projects/{project_env.pid}/tasks", body)["id"]

    project_env.agent.post(f"/api/v1/tasks/{tid}/claim")
    project_env.agent.post(f"/api/v1/tasks/{tid}/complete", {"note": "集成测试完成"})

    class NS:
        pass

    ns = NS()
    ns.task_id = tid
    return ns


@pytest.fixture(scope="session")
def review_env(admin, project_env, task_chain, data):
    """审核链：task_chain 的待审任务被通过 → done。"""
    admin.post(
        "/api/v1/reviews/batch",
        {"ids": [task_chain.task_id], "status": "approved", "comment": data["tasks"]["review"]["approve_comment"]},
    )
    return task_chain


# ---- UI 层（ui/ 目录才拉起 vite dev；API 层不碰前端） ----
@pytest.fixture(scope="session")
def frontend(backend):
    """vite dev server（data-test-id 天然保留），/api 代理指向临时后端。"""
    import os
    import shutil
    import socket
    import subprocess
    import time

    web_dir = TESTS_DIR.parent / "web"
    if not shutil.which("pnpm") and not (web_dir / "node_modules").exists():
        pytest.skip("前端依赖未安装（web/ 下先 pnpm install）")

    with socket.socket() as s:
        s.bind(("127.0.0.1", 0))
        port = s.getsockname()[1]

    env = os.environ.copy()
    env["VITE_PROXY"] = f'[["/api","{backend.base_url}"]]'
    env["VITE_PORT"] = str(port)
    log = open(TESTS_DIR / ".vite-itest.log", "w", encoding="utf-8")
    proc = subprocess.Popen(
        ["npx", "vite", "--port", str(port), "--strictPort"],
        cwd=web_dir,
        env=env,
        stdout=log,
        stderr=subprocess.STDOUT,
        shell=(os.name == "nt"),  # Windows 下 npx 需要 shell 解析
    )

    import httpx

    base = f"http://127.0.0.1:{port}"
    deadline = time.time() + 90
    ready = False
    while time.time() < deadline:
        try:
            if httpx.get(base, timeout=2.0).status_code == 200:
                ready = True
                break
        except httpx.HTTPError:
            time.sleep(0.5)
    if not ready:
        proc.kill()
        pytest.fail("vite dev 未就绪（见 tests/.vite-itest.log）")

    yield base
    proc.terminate()
    try:
        proc.wait(timeout=10)
    except subprocess.TimeoutExpired:
        proc.kill()
