"""平台 API 客户端：人类 token 头与 agent bc_/X-Session 双认证封装。

统一处理平台响应壳（成功 code ∈ {0, 200}；0 为 GoFrame 原生、200 为
中间件归一）；业务拒绝抛 BizError 携带端点与消息，便于断言负路径。
"""

from __future__ import annotations

import httpx

OK_CODES = (0, 200, None)


class BizError(AssertionError):
    """平台业务拒绝（code 非 0/200）——负路径断言用 expect_biz 捕获。"""

    def __init__(self, method: str, path: str, code, message: str):
        self.method, self.path, self.code, self.message = method, path, code, message
        super().__init__(f"[{method} {path}] code={code}: {message}")


class Client:
    def __init__(self, base_url: str, token: str | None = None,
                 bearer: str | None = None, session: str | None = None, timeout: float = 15.0):
        headers = {}
        if token:
            headers["token"] = token
        if bearer:
            headers["Authorization"] = f"Bearer {bearer}"
        if session:
            headers["X-Session"] = session
        self.http = httpx.Client(base_url=base_url, headers=headers, timeout=timeout)

    # ---- 基础方法：返回解壳后的 data ----
    def get(self, path, params=None):
        return self._unwrap(self.http.get(path, params=params))

    def post(self, path, json=None):
        return self._unwrap(self.http.post(path, json=json))

    def put(self, path, json=None):
        return self._unwrap(self.http.put(path, json=json))

    def delete(self, path):
        return self._unwrap(self.http.delete(path))

    def raw_get(self, path, params=None) -> httpx.Response:
        return self.http.get(path, params=params)

    def _unwrap(self, r: httpx.Response):
        r.raise_for_status()
        body = r.json()
        if body.get("code") not in OK_CODES:
            raise BizError(r.request.method, r.request.url.path, body.get("code"), body.get("message", ""))
        return body.get("data", body)

    # ---- 人类登录 ----
    @classmethod
    def login(cls, base_url, username, password) -> "Client":
        c = cls(base_url)
        data = c.post("/api/login", {"username": username, "password": password})
        return cls(base_url, token=data["token"])

    # ---- agent 身份：注册（bc_ key 只此一次返回）→ 接入码加入 → 会话 ----
    @classmethod
    def register_agent(cls, base_url, name) -> tuple[int, str]:
        """返回 (agentId, apiKey)——key 只此一次，能力调整等后续操作用 id。"""
        c = cls(base_url)
        data = c.post("/api/v1/agent/register", {"name": name})
        return data["agentId"], data["apiKey"]

    @classmethod
    def agent_join(cls, base_url, api_key, join_code) -> dict:
        c = cls(base_url, bearer=api_key)
        return c.post("/api/v1/agent/projects/join", {"code": join_code})

    @classmethod
    def agent_session(cls, base_url, api_key, project_id) -> "Client":
        c = cls(base_url, bearer=api_key)
        data = c.post("/api/v1/agent/sessions", {"projectId": project_id})
        return cls(base_url, bearer=api_key, session=data["sessionId"])


def expect_biz(fn, *args, contains: str = "", **kwargs):
    """负路径断言辅助：调用应抛 BizError 且消息含 contains。"""
    try:
        fn(*args, **kwargs)
    except BizError as e:
        if contains and contains not in e.message:
            raise AssertionError(f"业务拒绝消息不符：期望含「{contains}」实际「{e.message}」") from e
        return e
    raise AssertionError("预期业务拒绝，实际成功")
