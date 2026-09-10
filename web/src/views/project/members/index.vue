<template>
  <div>
    <n-card :bordered="false" title="项目成员" class="proCard">
      <template #header-extra>
        <n-space>
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
                <n-button text type="error" @click="handleRemove(member)">移除</n-button>
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
        </div>
      </n-space>
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
          <n-input v-model:value="formData.role" placeholder="请输入角色" />
        </n-form-item>
      </n-form>
    </n-modal>
  </div>
</template>

<script lang="ts" setup>
  import EmptyState from '@/components/EmptyState/EmptyState.vue';
  import { ref, reactive, computed, onMounted } from 'vue';
  import { useRoute } from 'vue-router';
  import { useMessage, useDialog } from 'naive-ui';
  import { getMembers, addMember, removeMember, removeAgentProject } from '@/api/project/index';
  import { createAgentJoinCode } from '@/api/agent/index';
  import type { MemberItem } from '@/api/project/index';

  const route = useRoute();
  const message = useMessage();
  const dialog = useDialog();
  const projectId = computed(() => Number(route.params.projectId));

  const loading = ref(false);
  const memberList = ref<MemberItem[]>([]);

  const showAddModal = ref(false);
  const formRef = ref<any>(null);
  // 角色显示映射：表单历史数据可能是自由文本，回落原值
  const MEMBER_ROLE_LABELS: Record<string, string> = { owner: '负责人', member: '成员' };

  const formData = reactive({ userId: null as number | null, role: '' });
  const formRules = {
    userId: { required: true, type: 'number', message: '请输入用户ID', trigger: 'blur' },
    role: { required: true, message: '请输入角色', trigger: 'blur' },
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
      formData.userId = null; formData.role = '';
      loadMembers();
    } catch { message.error('添加失败'); return false; }
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
        } catch { message.error('移除失败'); }
      },
    });
  }

  // Agent 接入码：owner 生成，一次性展示
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
