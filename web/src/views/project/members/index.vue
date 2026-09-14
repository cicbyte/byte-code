<template>
  <div>
    <n-card :bordered="false" title="项目成员" class="proCard">
      <template #header-extra>
        <!-- 管理级入口（owner/maintainer/超管）：member 不显示，后端同口径兜底 -->
        <n-space v-if="canManage">
          <n-button @click="handleGenAgentCode">Agent 接入码</n-button>
          <n-button type="primary" @click="showAddModal = true">添加成员</n-button>
        </n-space>
      </template>

      <n-spin :show="loading">
        <EmptyState type="member" title="暂无成员" v-if="!loading && memberList.length === 0" description="添加成员或让 Agent 凭接入码加入" />
        <n-table v-else :bordered="false" :single-line="false" size="small">
          <thead>
            <tr>
              <th class="col-idx">#</th>
            <th>用户名</th>
              <th>姓名</th>
              <th>角色</th>
              <th>加入时间</th>
              <th>操作</th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="(member, __ix) in memberList" :key="`${member.userType}-${member.id}`">
              <td class="col-idx">{{ __ix + 1 }}</td>
            <td>
                <n-space :size="4" align="center">
                  <span>{{ member.username }}</span>
                  <n-tag v-if="member.userType === 'ai'" size="tiny" :bordered="false" type="info">Agent</n-tag>
                </n-space>
              </td>
              <td>{{ member.realName }}</td>
              <td>{{ MEMBER_ROLE_LABELS[member.role] || member.role }}</td>
              <td>{{ member.joinedAt }}</td>
              <td>
                <n-space v-if="member.userType !== 'ai'" size="small">
                  <n-button
                    v-if="canTransfer && member.role !== 'owner'"
                    text
                    type="warning"
                    @click="handleTransfer(member)"
                  >转交负责人</n-button>
                  <n-button v-if="canRemoveRow(member)" text type="error" @click="handleRemove(member)">移除</n-button>
                </n-space>
                <n-space v-else-if="canManage" size="small">
                  <n-button text type="error" @click="handleRemove(member)">移除</n-button>
                  <n-button
                    text
                    type="info"
                    @click="openCaps(member)"
                  >{{ member.capabilities ? `能力·${member.capabilities.split(',').length}项` : '能力·全部' }}</n-button>
                </n-space>
              </td>
            </tr>
          </tbody>
        </n-table>
      </n-spin>
    </n-card>



    <!-- Agent 接入码（一次性展示） -->


    <n-modal v-model:show="showAgentCode" preset="dialog" title="Agent 项目接入" :show-icon="false" style="width: 560px">
      <n-space vertical :size="10" class="py-2">
        <n-alert type="info" :show-icon="false">
          把以下信息交给 Agent 侧（或其驱动者），复制即用。接入码一次性、24 小时有效，关闭后不再展示。
        </n-alert>

        <div class="onboard-sec">
          <div class="onboard-label">服务器地址 <n-text depth="3" style="font-size: 11px">（写入 ~/.bc/config.toml 的 server_url）</n-text></div>
          <div class="onboard-row">
            <n-code :code="serverUrl" language="text" class="onboard-code" />
            <n-button text size="tiny" type="primary" @click="copyText(serverUrl)">复制</n-button>
          </div>
        </div>

        <div class="onboard-sec">
          <div class="onboard-label">项目接入码</div>
          <div class="onboard-row">
            <n-code :code="agentCode" language="text" class="onboard-code" />
            <n-button text size="tiny" type="primary" @click="copyText(agentCode)">复制</n-button>
          </div>
          <n-text depth="3" style="font-size: 12px">有效期至：{{ agentCodeExpires }}</n-text>
        </div>

        <div class="onboard-sec">
          <div class="onboard-label">Agent 侧接入命令（bcode CLI）</div>
          <div class="onboard-row">
            <n-code :code="onboardCommands" language="bash" class="onboard-code block" />
            <n-button text size="tiny" type="primary" @click="copyText(onboardCommands)">复制</n-button>
          </div>
          <n-text depth="3" style="font-size: 12px">首次接入四步：配置服务器 → 注册身份（bc_ key 自动落本地）→ 凭码加入项目 → 建立会话（开工包）</n-text>
          <n-text depth="3" style="font-size: 12px">新接入 Agent 默认全部能力；接入后可在成员列表按项目收紧（能力·全部 按钮）</n-text>
        </div>
      </n-space>
    </n-modal>

    <!-- 转交负责人：选择转移方式（隔离交接可退出本项目） -->
    <n-modal v-model:show="showTransfer" preset="dialog" title="转交项目负责人" :show-icon="false" style="width: 460px">
      <n-space vertical :size="10" class="py-2">
        <span class="text-sm">确定将项目负责人转交给「{{ transferringTarget?.realName || transferringTarget?.username }}」吗？</span>
        <n-radio-group v-model:value="transferLeave">
          <n-space vertical>
            <n-radio :value="false">留在本项目 —— 转交后您成为普通成员</n-radio>
            <n-radio :value="true">退出本项目 —— 转交后移出成员列表（需要隔离时选择）</n-radio>
          </n-space>
        </n-radio-group>
        <n-text v-if="transferLeave" depth="3" style="font-size: 12px">
          退出后您将失去本项目全部访问权限，如需回来需新负责人重新添加
        </n-text>
      </n-space>
      <template #action>
        <n-space>
          <n-button size="small" @click="showTransfer = false">取消</n-button>
          <n-button size="small" type="warning" :loading="transferring" @click="confirmTransfer">确认转交</n-button>
        </n-space>
      </template>
    </n-modal>

    <!-- 添加成员弹窗 -->
    <n-modal
      v-model:show="showAddModal"
      preset="dialog"
      title="添加成员"
      positive-text="确定"
      negative-text="取消"
      @positive-click="handleAdd"
      style="width: 420px"
    >
      <n-form ref="formRef" :model="formData" :rules="formRules" label-placement="left" :label-width="80" class="py-4">
        <n-form-item label="用户ID" path="userId">
          <n-input-number v-model:value="formData.userId" placeholder="请输入用户ID" style="width: 100%" />
        </n-form-item>
        <n-form-item label="角色" path="role">
          <n-select v-model:value="formData.role" :options="roleOptions" placeholder="请选择角色" />
        </n-form-item>
      </n-form>
    </n-modal>

    <!-- Agent 能力集编辑（空=全部能力；接入后由管理侧收紧，即时生效） -->
    <n-modal
      v-model:show="showCaps"
      preset="dialog"
      title="Agent 能力集"
      positive-text="保存"
      negative-text="取消"
      :loading="capsSaving"
      @positive-click="saveCaps"
      style="width: 480px"
    >
      <n-space vertical :size="8" class="py-2">
        <n-alert type="info" :show-icon="false">
          {{ capsTarget?.username }} 在本项目的可行动作。全不勾选 = 全部能力（默认，存量 Agent 行为不变）。
        </n-alert>
        <n-checkbox-group v-model:value="capsSelected">
          <n-space :size="[24, 6]">
            <n-checkbox v-for="c in AGENT_CAPS" :key="c.key" :value="c.key" :label="c.label" />
          </n-space>
        </n-checkbox-group>
        <n-text depth="3" style="font-size: 12px">
          调整即时生效：受限 Agent 调用未授权端点会收到「未被授予「X」能力」的错误
        </n-text>
      </n-space>
    </n-modal>
  </div>
</template>

<script lang="ts" setup>
  import EmptyState from '@/components/EmptyState/EmptyState.vue';
  import { ref, reactive, computed, onMounted } from 'vue';
  import { useRoute, useRouter } from 'vue-router';
  import { useMessage, useDialog } from 'naive-ui';
  import { useUserStore } from '@/store/modules/user';
  import { usePerm } from '@/composables/usePerm';
  import { getMembers, addMember, removeMember, removeAgentProject, transferOwner } from '@/api/project/index';
  import { createAgentJoinCode, updateAgentCapabilities, AGENT_CAPS } from '@/api/agent/index';
  import type { MemberItem } from '@/api/project/index';

  const route = useRoute();
  const router = useRouter();
  const message = useMessage();
  const dialog = useDialog();
  const userStore = useUserStore();

  // 转交权限：本人在本项目是 owner，或平台管理员（后端同口径兜底）
  const { isAdmin } = usePerm();
  const canTransfer = computed(() => {
    const myId = Number((userStore?.info as any)?.userId || 0);
    const meRow = memberList.value.find((m) => m.userType !== 'ai' && m.userId === myId);
    if (meRow?.role === 'owner') return true;
    return isAdmin.value;
  });

  const showTransfer = ref(false);
  const transferring = ref(false);
  const transferLeave = ref(false);
  const transferringTarget = ref<MemberItem | null>(null);

  function handleTransfer(member: MemberItem) {
    transferringTarget.value = member;
    transferLeave.value = false;
    showTransfer.value = true;
  }

  async function confirmTransfer() {
    const m = transferringTarget.value;
    if (!m) return;
    transferring.value = true;
    try {
      await transferOwner(projectId.value, m.userId, transferLeave.value);
      message.success(transferLeave.value
        ? `负责人已转交给 ${m.realName || m.username}，您已退出本项目`
        : `负责人已转交给 ${m.realName || m.username}`);
      showTransfer.value = false;
      if (transferLeave.value) {
        // 退出后无权停留在成员页，回到项目列表
        router.push('/project/list');
      } else {
        loadMembers();
      }
    } catch (e: any) {
      message.error(e?.message || '转交失败');
    } finally {
      transferring.value = false;
    }
  }
  const projectId = computed(() => Number(route.params.projectId));

  const loading = ref(false);
  const memberList = ref<MemberItem[]>([]);

  const showAddModal = ref(false);
  const formRef = ref<any>(null);
  // 角色显示映射：表单历史数据可能是自由文本，回落原值
  const MEMBER_ROLE_LABELS: Record<string, string> = { owner: '负责人', maintainer: '维护者', member: '成员' };

  // 本人在本项目的角色（未入本项目时为空）
  const myRole = computed(() => {
    const myId = Number((userStore?.info as any)?.userId || 0);
    const meRow = memberList.value.find((m) => m.userType !== 'ai' && m.userId === myId);
    return (meRow?.role as string) || '';
  });
  // 管理级（owner/maintainer/超管）：成员管理、接入码入口（后端 IsProjectMaintainer 同口径）
  const canManage = computed(() => ['owner', 'maintainer'].includes(myRole.value) || isAdmin.value);
  // 移除按钮按档位：owner 行仅超管可动；maintainer 行仅 owner/超管；member 行管理级即可
  function canRemoveRow(member: MemberItem) {
    if (member.userType === 'ai') return canManage.value;
    if (member.role === 'owner') return isAdmin.value;
    if (member.role === 'maintainer') return myRole.value === 'owner' || isAdmin.value;
    return canManage.value;
  }
  // 添加成员角色选项：maintainer 档仅 owner/超管可授予（后端对称校验）
  const roleOptions = computed(() => {
    const opts = [{ label: '成员', value: 'member' }];
    if (myRole.value === 'owner' || isAdmin.value) {
      opts.push({ label: '维护者（可管任务/成员/接入码）', value: 'maintainer' });
    }
    return opts;
  });

  const formData = reactive({ userId: null as number | null, role: 'member' });
  const formRules = {
    userId: { required: true, type: 'number', message: '请输入用户ID', trigger: 'blur' },
    role: { required: true, message: '请选择角色', trigger: ['blur', 'change'] },
  };

  async function loadMembers() {
    loading.value = true;
    try {
      const res = await getMembers(projectId.value);
      memberList.value = res?.list || [];
    } catch { /* ignore */ } finally { loading.value = false; }
  }

  async function handleAdd() {
    try { await formRef.value?.validate(); } catch { return false; }
    try {
      await addMember(projectId.value, { userId: formData.userId!, role: formData.role });
      message.success('成员添加成功');
      showAddModal.value = false;
      formData.userId = null; formData.role = 'member';
      loadMembers();
    } catch (e: any) { message.error(e?.message || '添加失败'); return false; }
  }

  function handleRemove(member: MemberItem) {
    dialog.warning({
      title: '确认移除',
      // 协议接入的 agent（viaBinding=1）走移除准入（清 binding+会话）；
      // members 表里的行（含早期手动加的 agent）走移除成员
      content: member.viaBinding === 1
        ? `确定要移除 Agent 准入「${member.username}」吗？其已签发会话将一并失效`
        : `确定要移除成员「${member.username}」吗？`,
      positiveText: '确定',
      negativeText: '取消',
      onPositiveClick: async () => {
        try {
          if (member.viaBinding === 1) {
            await removeAgentProject(projectId.value, member.userId);
          } else {
            await removeMember(projectId.value, member.userId);
          }
          message.success('移除成功'); loadMembers();
        } catch (e: any) { message.error(e?.message || '移除失败'); }
      },
    });
  }

  // ==================== Agent 能力集（PRD §5） ====================
  const showCaps = ref(false);
  const capsTarget = ref<MemberItem | null>(null);
  const capsSelected = ref<string[]>([]);
  const capsSaving = ref(false);

  function openCaps(member: MemberItem) {
    capsTarget.value = member;
    capsSelected.value = (member.capabilities || '').split(',').filter(Boolean);
    showCaps.value = true;
  }

  async function saveCaps() {
    const m = capsTarget.value;
    if (!m) return;
    capsSaving.value = true;
    try {
      await updateAgentCapabilities(projectId.value, m.userId, capsSelected.value);
      message.success(capsSelected.value.length ? `已收紧为 ${capsSelected.value.length} 项能力` : '已恢复全部能力');
      showCaps.value = false;
      loadMembers();
    } catch (e: any) {
      message.error(e?.message || '保存失败');
      return false;
    } finally {
      capsSaving.value = false;
    }
  }

  // Agent 接入码：owner/maintainer 生成，一次性展示
  // （关联项目/分组治理已拆至项目设置页；本段在 f6830e7 拆分时被误删，自 0ca5705 恢复）
  const showAgentCode = ref(false);
  const agentCode = ref('');
  const agentCodeExpires = ref('');
  // 连接信息随码展示：agent 侧无需手动问服务器地址（取当前访问地址 + /api 前缀）
  const serverUrl = computed(() => window.location.origin + '/api');
  const onboardCommands = computed(() =>
    [
      '# 1. 配置服务器（一次性，写入 ~/.bc/config.toml）',
      `bcode config set server ${serverUrl.value}`,
      '',
      '# 2. 注册身份（bc_ key 自动落本地 ~/.bc/agents/<profile>/）',
      'bcode register <agent-name>',
      '',
      '# 3. 凭码加入本项目',
      `bcode join ${agentCode.value || '<接入码>'}`,
      '',
      '# 4. 在项目目录建立会话（展示开工包：项目记忆 + 我的任务 + 待分析反馈）',
      'bcode start',
    ].join('\n'));
  function copyText(text: string) {
    navigator.clipboard?.writeText(text).then(
      () => message.success('已复制'),
      () => message.error('复制失败，请手动选择'),
    );
  }
  async function handleGenAgentCode() {
    try {
      const res = await createAgentJoinCode(projectId.value);
      agentCode.value = res.code;
      agentCodeExpires.value = res.expiresAt;
      showAgentCode.value = true;
      loadMembers();
    } catch {
      // http 层统一提示（无权限等）
    }
  }

  onMounted(() => { loadMembers(); });
</script>

<style lang="less" scoped>
  .onboard-sec {
    .onboard-label {
      font-size: 12px;
      font-weight: 600;
      color: var(--text-color-2, #5c6470);
      margin-bottom: 4px;
    }
    .onboard-row {
      display: flex;
      align-items: flex-start;
      gap: 8px;

      .onboard-code {
        flex: 1;
        min-width: 0;
        background: #f7f8fa;
        border-radius: 6px;
        padding: 8px 10px;
        font-size: 12px;
      }
      .onboard-code.block {
        white-space: pre;
        overflow-x: auto;
      }
    }
  }
</style>
