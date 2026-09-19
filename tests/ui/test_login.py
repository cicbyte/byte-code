"""UI 登录链路：正确登录跳转控制台；错误凭据被拒留在登录页。

模块内步骤序（pytest-order）+ 上游失败联动（pytest-dependency）。
"""

import pytest

from ui import LoginPage

pytestmark = [pytest.mark.ui, pytest.mark.order(1)]


def test_login_ok(browser, frontend, backend, data):
    lp = LoginPage(browser, frontend)
    lp.goto(data["ui"]["routes"]["login"])
    users = data["users"]
    lp.login(users["admin"]["username"], users["admin"]["password"])
    lp.expect_login_ok()
    assert "/login" not in browser.url


@pytest.mark.dependency()
def test_login_rejected(browser, frontend, data):
    """登出后错误密码登录：停留登录页。"""
    lp = LoginPage(browser, frontend)
    lp.goto(data["ui"]["routes"]["login"])
    lp.login(data["users"]["admin"]["username"], "wrong-password")
    lp.expect_login_rejected()
