"""干净后端实例编排：临时目录 + 全量迁移 + 随机端口 + 绝对路径 SQLite。

session 级唯一实例（tests/conftest.py 的 backend fixture 调用）：
- 二进制来源：环境变量 BYTECODE_TEST_BIN 指定的现成构建；否则 go build 产临时件
- 迁移链来自仓库 resource/sql/sqlite（启动自动应用，兼容到最新迁移号）
- admin 首登强制改密由本模块完成（新密码取 data/users.yaml）
"""

import os
import shutil
import socket
import subprocess
import tempfile
import time
from pathlib import Path

import httpx

REPO_ROOT = Path(__file__).resolve().parents[2]
FIXTURE_SECRET = "itest-jwt-secret-do-not-use-in-prod"


def free_port() -> int:
    with socket.socket() as s:
        s.bind(("127.0.0.1", 0))
        return s.getsockname()[1]


class BootedBackend:
    def __init__(self, proc, base_url, workdir, db_path, admin_password):
        self.proc = proc
        self.base_url = base_url
        self.workdir = workdir
        self.db_path = db_path
        self.admin_password = admin_password

    def stop(self):
        if self.proc.poll() is None:
            self.proc.terminate()
            try:
                self.proc.wait(timeout=10)
            except subprocess.TimeoutExpired:
                self.proc.kill()
        shutil.rmtree(self.workdir, ignore_errors=True)


def build_binary(target: Path) -> Path:
    env = os.environ.copy()
    env["CGO_ENABLED"] = "0"
    subprocess.run(
        ["go", "build", "-o", str(target), "."],
        cwd=REPO_ROOT,
        check=True,
        env=env,
        timeout=300,
    )
    return target


def boot(admin_username: str, admin_initial: str, admin_new: str) -> BootedBackend:
    workdir = Path(os.environ.get("BYTECODE_ITEST_KEEP") or "") if os.environ.get("BYTECODE_ITEST_KEEP") else None
    keep = workdir is not None
    if workdir is None:
        workdir = Path(tempfile.mkdtemp(prefix="bytecode-itest-"))

    # 迁移与静态资源随行（服务按 CWD 相对路径读 resource/）
    res = workdir / "resource"
    (res / "data").mkdir(parents=True, exist_ok=True)
    shutil.copytree(REPO_ROOT / "resource" / "sql", res / "sql", dirs_exist_ok=True)
    public = REPO_ROOT / "resource" / "public"
    if public.exists():
        shutil.copytree(public, res / "public", dirs_exist_ok=True)

    port = free_port()
    db_path = (res / "data" / "itest.db").resolve()
    (workdir / "config.yaml").write_text(
        f"""server:
  address: ":{port}"
  openapiPath: ""
  swaggerPath: ""
logger:
  level: "error"
  stdout: true
database:
  default:
    link: "sqlite::@file({db_path.as_posix()})"
    extra: "busy_timeout=10000&journal_mode=WAL"
token:
  secret: "{FIXTURE_SECRET}"
""",
        encoding="utf-8",
    )

    bin_env = os.environ.get("BYTECODE_TEST_BIN")
    binary = Path(bin_env) if bin_env else build_binary(workdir / "bytecode-itest.exe")
    env = os.environ.copy()
    env["JWT_SECRET"] = FIXTURE_SECRET
    log = open(workdir / "server.log", "w", encoding="utf-8")
    proc = subprocess.Popen(
        [str(binary)], cwd=workdir, env=env, stdout=log, stderr=subprocess.STDOUT
    )

    base = f"http://127.0.0.1:{port}"
    _wait_health(base)

    booted = BootedBackend(proc, base, workdir, db_path, admin_new)
    _finish_admin_onboarding(booted, admin_username, admin_initial, admin_new)
    booted.keep = keep  # BYTECODE_ITEST_KEEP 指定目录时保留现场便于排查
    return booted


def _wait_health(base: str, timeout_s: float = 60.0):
    deadline = time.time() + timeout_s
    last = None
    with httpx.Client(base_url=base, timeout=2.0) as c:
        while time.time() < deadline:
            try:
                # /api/v1/health 在鉴权链后：401 也证明路由/中间件已活着
                r = c.get("/api/v1/health")
                if r.status_code < 500:
                    return
                last = r.status_code
            except httpx.HTTPError as e:
                last = e
            time.sleep(0.4)
    raise RuntimeError(f"后端未在 {timeout_s}s 内就绪：{last}")


def _finish_admin_onboarding(b: BootedBackend, username: str, initial: str, new: str):
    """种子 admin 带 must_change_password=1：登录 → 首改（免旧密）→ 换长效 token。"""
    with httpx.Client(base_url=b.base_url, timeout=10.0) as c:
        r = c.post("/api/login", json={"username": username, "password": initial})
        data = _unwrap(r)
        if not data.get("mustChangePassword"):
            return  # 现场复用（KEEP 模式）已改过
        r = c.put(
            "/api/account/password",
            json={"newPassword": new},
            headers={"token": data["token"]},
        )
        _unwrap(r)


def _unwrap(r: httpx.Response):
    r.raise_for_status()
    body = r.json()
    code = body.get("code")
    if code not in (0, 200, None):
        raise AssertionError(f"业务拒绝 [{r.request.method} {r.request.url.path}]: {body.get('message')}")
    return body.get("data", body)
