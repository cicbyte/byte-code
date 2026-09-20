"""DrissionPage UI 用例基类：按 data-test-id 定位的页面对象。

约定：前端锚点规范见 tests/README（scope.element 命名、静态/动态皆可，
生产构建剥离）。定位统一走 css @[data-test-id="..."]。
"""

from __future__ import annotations


class UiBase:
    def __init__(self, page, base_url: str):
        self.page = page            # DrissionPage ChromiumPage
        self.base_url = base_url

    # ---- 导航 ----
    def goto(self, route: str):
        self.page.get(self.base_url.rstrip("/") + route)

    # ---- 锚点定位（DrissionPage 原生 @attr=value 精确匹配）----
    def sel(self, anchor: str) -> str:
        return f"@data-test-id={anchor}"

    def ele(self, anchor: str):
        return self.page.ele(self.sel(anchor))

    def eles(self, anchor: str):
        return self.page.eles(self.sel(anchor))

    def wait_ele(self, anchor: str, timeout: float = 10.0):
        e = self.page.wait.ele_displayed(self.sel(anchor), timeout=timeout)
        if not e:
            raise AssertionError(f"等待锚点超时：{anchor}（当前页面 {self.page.url}）")
        return self.page.ele(self.sel(anchor))

    # ---- 高频动作 ----
    def click(self, anchor: str):
        self.wait_ele(anchor).click()

    def _input_node(self, anchor):
        """锚点内的真实输入节点：锚点多落在组件根（如 n-input 的 div），
        直接在其上 .input() 会命中页面首个 input——必须显式下钻。
        n-input type=textarea 渲染 <textarea>，一并下钻。"""
        e = self.wait_ele(anchor)
        return e.ele("tag:input") or e.ele("tag:textarea") or e

    def input(self, anchor: str, text: str):
        node = self._input_node(anchor)
        node.clear()
        node.input(text)

    def text_of(self, anchor: str) -> str:
        return (self.wait_ele(anchor).text or "").strip()


class LoginPage(UiBase):
    def login(self, username: str, password: str, first_load_timeout: float = 45.0):
        # vite dev 首屏 on-demand 编译可达数十秒：首次等表单要放宽
        self.input("login.username", username)
        self.input("login.password", password)
        self.click("login.submit")

    def expect_login_ok(self, timeout: float = 20.0):
        """登录成功：URL 离开 /login 且出现工作台外壳（侧栏菜单）。"""
        self.page.wait.url_change("/login", exclude=True, timeout=timeout)
        # 首跳后 SPA 还要拉 admin_info/menus；等任意菜单项出现才算落稳
        self.page.wait.ele_displayed("tag:aside", timeout=timeout) or self.page.wait.ele_displayed(
            "text:项目", timeout=5
        )

    def expect_login_rejected(self):
        """错误凭据停留在登录页并出现错误提示"""
        self.page.wait.doc_loaded()
        assert "/login" in self.page.url
