"""个人设置：修改密码弹窗与全链路（弹窗失活缺陷回归）。"""
import time

import pytest
from ui import UiBase

pytestmark = [pytest.mark.ui, pytest.mark.order(5)]


def test_pwd_full_flow(logged_in, frontend, data):
    """改密全链路：开弹窗 → 填表 → 提交 → 旧密错误提示（不真改，避免破坏 fixture 会话）。"""
    import time
    page = UiBase(logged_in, frontend)
    page.goto("/setting/account")
    assert page.wait_ele("setting-account.item-2", timeout=30)
    page.click("setting-account.item-2")
    time.sleep(1)
    page.click("setting-account.pwd-edit-btn")
    assert page.wait_ele("setting-account.pwd-old-input", timeout=8)
    # 填一个错误旧密码 + 合法新密码 → 提交 → 后端应报「旧密码不正确」
    page.input("setting-account.pwd-old-input", "wrong-old-1")
    page.input("setting-account.pwd-new-input", "NewPass9x")
    # 确认密码在第三个输入框（组件内下钻）
    modal = page.ele("setting-account.pwd-modal")
    inputs = modal.eles("tag:input")
    inputs[2].input("NewPass9x")
    logged_in.ele("text:确认修改").click()
    assert logged_in.wait.ele_displayed("text:旧密码不正确", timeout=10)
    # 关弹窗收尾
    logged_in.ele("text:取 消") or logged_in.ele("text:取消")
    cancel = [b for b in logged_in.eles("css:.n-dialog button") if "取消" in (b.text or "")]
    if cancel:
        cancel[0].click()


def test_pwd_modal_opens(logged_in, frontend, data):
    page = UiBase(logged_in, frontend)
    page.goto("/setting/account")
    assert page.wait_ele("setting-account.item-2", timeout=30)
    page.click("setting-account.item-2")
    time.sleep(1.2)
    page.click("setting-account.pwd-edit-btn")
    ok = page.wait_ele("setting-account.pwd-old-input", timeout=8)
    assert ok, "改密弹窗未打开"
    print("MODAL>>>OPENED")
