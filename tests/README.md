# ByteCode 集成测试（tests/）

pytest 驱动的平台自身集成测试：**API 层**（HTTP 全链）与 **UI 层**（DrissionPage 真
浏览器）共享同一个临时后端世界；uv 管理 Python 环境；YAML 管理测试数据。

## 运行

```bash
cd tests
uv sync                 # CI / 常规路径（uv.lock 锁定 37 包）
uv run pytest api -q    # API 集成测试（自包含，无外部依赖）
uv run pytest ui -q     # UI E2E（需本机 Chrome；--ui-headed 可视化）
uv run pytest -q        # 全量
```

本机若遇 `uv sync` 报 `Failed to update Windows PE resources`（杀软拦截
UpdateResource 的环境特征），回退 stdlib 环境跑：

```bash
python -m venv .venv
.venv/Scripts/python -m pip install pytest pytest-dependency pytest-order httpx pyyaml DrissionPage byte-code-pytest
.venv/Scripts/python -m pytest -q
```

## 目录

```
tests/
├── conftest.py        # 后端/vite 编排（session）+ 世界状态链（api/ui 共享）
├── helpers/booted.py  # 干净实例：随机端口+绝对路径 SQLite+迁移+首登改密
├── helpers/apiclient.py # API 客户端（token 头 / bc_+X-Session 双认证）
├── helpers/ui.py      # DrissionPage 页面对象（data-test-id 定位）
├── data/*.yaml        # 测试数据（users.yaml 是账号密码唯一事实源）
├── api/               # 认证 / 项目域 / 任务全链 / 执行记录四模块
└── ui/                # 登录 / 看板+任务详情 / 执行记录页
```

环境旋钮：`BYTECODE_TEST_BIN`（复用现成后端二进制，省一次 go build）、
`BYTECODE_ITEST_KEEP`（指定目录保留现场排查）。

## 依赖顺序三层纪律

1. **跨模块状态链 = fixture 图**：`project_env`（项目+成员+agent）→
   `task_chain`（建→认领→完成）→ `review_env`（审核通过），session 级只建
   一次，API/UI 共享同一世界；
2. **模块内步骤序 = `@pytest.mark.order(n)`**：登录→导航→操作→断言；
3. **上游失败联动跳过 = `@pytest.mark.dependency(depends=[...])`**：前置
   失败时后续步骤 skip 而非假绿。

## data-test-id 铺设规范（前端）

- 命名 `data-test-id="<scope>.<element>"`，scope 为页面域：
  `login` / `project-list` / `board` / `task-detail` / `runs` / `system-user`…
- 常用后缀：`.create-btn` / `.confirm-btn` / `.search-input` / `.row-{id}` /
  `.item-{id}` / `.page`（页面根）；循环体用动态绑定
  `:data-test-id="\`scope.row-${item.id}\`"`；
- **dev/测试保留、生产构建剥离**：`web/build/vite/plugin/testId.ts` 在
  `vite build` 时整属性移除（静态与 `:动态` 两种形态都剥）；CI 有
  `Test anchors stripped from dist` 门禁，泄漏即红；
- UI 用例定位统一走 `helpers/ui.py`（输入类锚点会下钻组件根内 `<input>`）。

## 用例元数据注册表（tests/cases/）

`--bcode-sync` 自动建的用例只有 nodeid 索引；人读元数据以
`tests/cases/*.yaml` 为准（每用例一文件）：中文标题、分类
（API/UI 自动化）、业务模块、优先级、前置条件、步骤、预期，
`externalKey` 即 pytest nodeid。

```bash
# 改完注册表后上推（externalKey 幂等 upsert，原地更新平台用例）
bcode test --cases --push --dir tests/cases
# 反向拉平台用例到本地（owner 在 Web 手工改过时同步回仓库）
bcode test --cases --pull --dir tests/cases
```

**运行时必须带 `--bcode --bcode-sync`**：sync 对已存在用例是
「按 externalKey 查找复用」（不会重复建），执行记录详情才能回显
中文标题与映射；只带 `--bcode` 时按设计仅记录 nodeid 不做映射。

## dogfood：执行记录上报平台

测试即产品场景——跑完直接上报 ByteCode 平台的执行记录（`byte-code-pytest`）：

```bash
# 一次性：owner 在 Web 生成接入码，注册 CI agent 并勾「上报测试执行」能力
bcode register ci-itest && bcode join <接入码>

# 上报（BCODE_PROJECT 为平台项目 id）
uv run pytest --bcode --bcode-sync
#   BCODE_URL=https://<平台>/api  BCODE_KEY=<bc_...>  BCODE_PROJECT=<id>
```

`--bcode-sync` 按 nodeid 自动建平台用例映射；Flaky/趋势/失败转缺陷在平台
「测试 → 执行记录」可视。CI 侧接入（secrets 配 `BCODE_KEY` 后启用
`--bcode`）留作下一步。
