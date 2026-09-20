"""认证扩展：忘记密码（防枚举恒定成功）+ 凭令牌重置（白盒种令牌走真实校验链）。

SMTP 未配置时邮件链路拿不到令牌——直接往 password_resets 落 sha256 摘要
（与生产同构），重置端点仍完整执行令牌校验/单次/过期/改密/踢会话逻辑。
"""

import hashlib
import sqlite3
import time

import pytest

from apiclient import BizError, Client, expect_biz

pytestmark = pytest.mark.order(2)


def _seed_token(backend, user_id, token, minutes=30, used=0):
    conn = sqlite3.connect(backend.db_path, timeout=10)
    conn.execute("PRAGMA busy_timeout=10000")
    digest = hashlib.sha256(token.encode()).hexdigest()
    expires = time.strftime("%Y-%m-%d %H:%M:%S", time.localtime(time.time() + minutes * 60))
    with conn:
        conn.execute(
            "INSERT INTO password_resets (user_id, token_hash, expires_at, used) VALUES (?,?,?,?)",
            (user_id, digest, expires, used),
        )
    conn.close()


def test_forgot_password_silent_for_unknown_account(backend):
    """防枚举：不存在账号与存在账号返回完全一致的成功形态。"""
    import httpx

    with httpx.Client(base_url=backend.base_url, timeout=10) as c:
        r1 = c.post("/api/auth/forgot-password", json={"account": "no-such-user-xyz"})
        r2 = c.post("/api/auth/forgot-password", json={"account": "admin"})
        assert r1.status_code == r2.status_code
        assert r1.json().get("code") == r2.json().get("code") == 0


def test_reset_rejects_bad_token_and_short_password(admin):
    expect_biz(
        lambda: admin.post(
            "/api/auth/reset-password", {"token": "not-a-real-token", "newPassword": "NewPass@123"}
        ),
        contains="无效",
    )
    expect_biz(
        lambda: admin.post("/api/auth/reset-password", {"token": "whatever", "newPassword": "short"}),
        contains="长度",
    )


def test_reset_happy_path_single_use_and_expiry(admin, backend, data):
    """白盒种令牌：重置成功可登录、旧密码失效、令牌单次、过期令牌拒绝。"""
    uid = admin.post(
        "/api/v1/admin/users",
        {"username": "itest-reset", "password": "Reset@Old1", "realName": "重置用例"},
    )["id"]

    _seed_token(backend, uid, "tok-happy-1")
    admin.post("/api/auth/reset-password", {"token": "tok-happy-1", "newPassword": "Reset@New1"})

    newc = Client.login(backend.base_url, "itest-reset", "Reset@New1")
    # 新密码会话有效（profile 可达即认证通过）
    assert newc.get("/api/account/profile") is not None
    with pytest.raises(BizError):
        Client.login(backend.base_url, "itest-reset", "Reset@Old1")

    # 令牌单次：同令牌再用被拒；过期令牌拒绝
    with pytest.raises(BizError):
        admin.post("/api/auth/reset-password", {"token": "tok-happy-1", "newPassword": "Reset@New2x"})
    _seed_token(backend, uid, "tok-expired", minutes=-1)
    with pytest.raises(BizError):
        admin.post("/api/auth/reset-password", {"token": "tok-expired", "newPassword": "Reset@New3x"})
