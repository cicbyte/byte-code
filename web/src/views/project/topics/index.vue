<template>
  <div>
    <n-card :bordered="false" class="proCard">
      <template #header>专题 <n-text depth="3" style="font-size: 12px; font-weight: 400">长时间自动执行的工程，与日常任务分池</n-text></template>
      <template #header-extra>
        <n-space :size="8">
          <n-radio-group v-model:value="statusFilter" size="small" @update:value="onFilterChange">
            <n-radio-button value="active">进行中</n-radio-button>
            <n-radio-button value="completed">已完成</n-radio-button>
            <n-radio-button value="abandoned">已放弃</n-radio-button>
            <n-radio-button value="all">全部</n-radio-button>
          </n-radio-group>
          <n-button size="small" type="primary" ghost @click="openCreate">新建专题</n-button>
        </n-space>
      </template>

      <n-spin :show="loading">
        <EmptyState v-if="!loading && list.length === 0" type="task" title="没有专题"
          description="先在知识库写 PRD，再创建专题关联它，拆解阶段后交给 Agent 持续推进" />
        <div v-for="t in list" :key="t.id" class="topic-card">
          <div class="tp-head" @click="toggle(t.id)">
            <span class="tp-caret">{{ expanded === t.id ? '▾' : '▸' }}</span>
            <n-tag size="small" :type="t.status === 'active' ? 'info' : t.status === 'completed' ? 'success' : 'default'">
              {{ statusLabel(t.status) }}
            </n-tag>
            <span class="tp-title">{{ t.title }}</span>
            <span class="tp-meta">
              <n-tag v-if="t.assigneeName" size="tiny" type="info" :bordered="false">{{ t.assigneeName }}</n-tag>
              <span v-if="t.phaseTotal">{{ t.phaseDone }}/{{ t.phaseTotal }} 阶段</span>
            </span>
          </div>
          <n-progress v-if="t.phaseTotal" :percentage="Math.round((t.phaseDone / t.phaseTotal) * 100)" :height="5" :show-indicator="false" class="tp-bar" />
          <div v-if="t.lastHandoff" class="tp-handoff">上次交接：{{ t.lastHandoff.slice(0, 80) }}{{ t.lastHandoff.length > 80 ? '…' : '' }}</div>

          <div v-if="expanded === t.id" class="tp-detail">
            <div v-if="t.goal" class="tp-sec"><div class="tp-label">目标</div><div class="tp-md">{{ t.goal }}</div></div>
            <div v-if="t.acceptance" class="tp-sec"><div class="tp-label">验收标准</div><div class="tp-md">{{ t.acceptance }}</div></div>
            <div v-if="t.docPath" class="tp-sec"><div class="tp-label">PRD 文档</div>
              <n-button text type="info" size="small" @click="openDoc(t)">{{ t.docPath }}</n-button>
            </div>
            <div class="tp-sec">
              <div class="tp-label">阶段清单</div>
              <div v-if="t.phases.length === 0" class="tp-empty">未拆解（Agent 读 PRD 后批量导入，或在下方手动添加）</div>

              <!-- 表头 -->
              <div v-if="t.phases.length" class="ph-table ph-head-row">
                <span class="ph-c-idx">#</span>
                <span class="ph-c-status">状态</span>
                <span class="ph-c-title">标题</span>
                <span class="ph-c-assignee">负责人</span>
                <span class="ph-c-ops"></span>
              </div>

              <template v-for="(p, pi) in t.phases" :key="p.id">
                <div class="ph-table ph-row" :class="{ done: p.status === 'done' }">
                  <span class="ph-c-idx">{{ pi + 1 }}</span>
                  <span class="ph-c-status">
                    <n-tag size="tiny" :bordered="false" :type="phaseTag(p.status)" class="ph-status-tag" @click="t.status === 'active' && cyclePhase(t, p)">
                      {{ phaseStatusText(p.status) }}
                    </n-tag>
                  </span>
                  <span class="ph-c-title" @click="openPhaseDrawer(t, p)">
                    <span class="ph-caret">›</span>
                    <span class="ph-title">{{ p.title }}</span>
                    <n-tag v-if="p.taskId" size="tiny" :bordered="false" type="info" class="ph-task-tag" @click.stop="openTask(p.taskId)">
                      历史任务 #{{ p.taskId }}
                    </n-tag>
                  </span>
                  <span class="ph-c-assignee">{{ p.assigneeName || '—' }}</span>
                  <span class="ph-c-ops">
                    <n-button v-if="t.status === 'active' && !p.taskId" text size="tiny" type="error" @click="removePhase(t, p)">删</n-button>
                  </span>
                </div>
              </template>

              <div v-if="t.status === 'active'" class="phase-add">
                <n-input
                  v-model:value="phaseInput[t.id]"
                  size="tiny"
                  placeholder="添加阶段，回车确认"
                  @keydown.enter.exact.prevent="addPhase(t)"
                >
                  <template #prefix><n-icon size="12"><svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M12 5v14M5 12h14"/></svg></n-icon></template>
                </n-input>
              </div>
            </div>
            <div class="tp-acts">
              <n-button v-if="t.status === 'active'" size="tiny" type="warning" ghost @click="finish(t, 'abandoned')">放弃</n-button>
              <n-button v-if="t.status === 'active'" size="tiny" type="success" @click="finish(t, 'completed')">验收通过</n-button>
            </div>
          </div>
        </div>
        <div v-if="total > pagination.size" class="mt-4 flex justify-end">
          <n-pagination
            v-model:page="pagination.page"
            :page-size="pagination.size"
            :item-count="total"
            @update:page="load"
          />
        </div>
      </n-spin>
    </n-card>

    <TaskDetailModal ref="taskDetailRef" @updated="load" />

    <!-- 阶段编辑抽屉：表格行点击进入，右侧滑出 -->
    <n-drawer v-model:show="showPhaseDrawer" :width="560" :mask="false" placement="right">
      <n-drawer-content :title="`阶段 · ${editForm.title || ''}`" closable>
        <n-space vertical :size="14">
          <div>
            <div class="pd-label">状态</div>
            <n-radio-group v-model:value="editForm.status" size="small" :disabled="currentTopic?.status !== 'active'">
              <n-radio-button value="pending">待处理</n-radio-button>
              <n-radio-button value="in_progress">进行中</n-radio-button>
              <n-radio-button value="done">已完成</n-radio-button>
          </n-radio-group>
          </div>
          <div>
            <div class="pd-label">标题</div>
            <n-input v-model:value="editForm.title" placeholder="阶段标题" :disabled="currentTopic?.status !== 'active'" />
          </div>
          <div>
            <div class="pd-label">负责人</div>
            <n-select v-model:value="editForm.assigneeId" :options="memberOptions" placeholder="选择负责人（可空）" clearable :disabled="currentTopic?.status !== 'active'" />
          </div>
          <div>
            <div class="pd-label">描述</div>
            <MdEditor
              :model-value="editForm.detail"
              :editor-id="'ph-editor-detail'"
              :sanitize="safeHtml"
              :toolbars="mdToolbars"
              :footers="[]"
              placeholder="这个阶段要做什么（markdown，右侧实时预览）"
              style="height: 260px"
              :read-only="currentTopic?.status !== 'active'"
              @update:model-value="(v: string) => (editForm.detail = v)"
            />
          </div>
          <div>
            <div class="pd-label">产出</div>
            <MdEditor
              :model-value="editForm.artifacts"
              :editor-id="'ph-editor-artifacts'"
              :sanitize="safeHtml"
              :toolbars="mdToolbars"
              :footers="[]"
              placeholder="报告 / 提交 / 验证说明（markdown）"
              style="height: 260px"
              :read-only="currentTopic?.status !== 'active'"
              @update:model-value="(v: string) => (editForm.artifacts = v)"
            />
          </div>
        </n-space>
        <template #footer>
          <n-space :size="10">
            <n-button size="small" :disabled="currentTopic?.status !== 'active'" @click="showPhaseDrawer = false">取消</n-button>
            <n-button size="small" type="primary" :disabled="!editForm.title.trim()" :loading="savingPhase" @click="savePhaseDrawer">保存</n-button>
          </n-space>
        </template>
      </n-drawer-content>
    </n-drawer>

    <!-- 新建专题：goal/验收/PRD 路径/执行 agent -->
    <n-modal v-model:show="showCreate" preset="dialog" title="新建专题" :show-icon="false" style="width: 520px">
      <n-space vertical :size="8" class="py-2">
        <n-input v-model:value="form.title" placeholder="专题标题（如：补齐全项目端到端测试）" size="small" />
        <n-input v-model:value="form.goal" type="textarea" placeholder="目标与范围（markdown）" :rows="3" size="small" />
        <n-input v-model:value="form.acceptance" type="textarea" placeholder="验收标准（完成判据，人验收依据）" :rows="2" size="small" />
        <n-input v-model:value="form.docPath" placeholder="PRD 文档路径（知识库内，如 PRD/xxx.md，可空）" size="small" />
        <n-select v-model:value="form.assigneeId" :options="agentOptions" placeholder="执行 Agent（持续推进会分配给它）" size="small" clearable />
      </n-space>
      <template #action>
        <n-space>
          <n-button size="small" @click="showCreate = false">取消</n-button>
          <n-button size="small" type="primary" :disabled="!form.title.trim()" :loading="creating" @click="handleCreate">创建</n-button>
        </n-space>
      </template>
    </n-modal>
  </div>
</template>

<script lang="ts" setup>
  import { ref, computed, reactive, onMounted } from 'vue';
  import { useRoute, useRouter } from 'vue-router';
  import { useMessage, useDialog } from 'naive-ui';
  import EmptyState from '@/components/EmptyState/EmptyState.vue';
  import TaskDetailModal from '@/views/project/components/TaskDetailModal.vue';
  import { MdEditor } from 'md-editor-v3';
  import 'md-editor-v3/lib/style.css';
  import DOMPurify from 'dompurify';
  import { statusLabel, statusTagType } from '@/enums/task';
  import { mdToolbars } from '@/utils/mdEditor';
  import {
    getTopics,
    appendTopicPhase,
    createTopic,
    toggleTopicPhase,
    updateTopicPhase,
    deleteTopicPhase,
    finishTopic,
    getMembers,
  } from '@/api/project/index';
  import type { TopicItem } from '@/api/project/index';

  const route = useRoute();
  const router = useRouter();
  const message = useMessage();
  const dialog = useDialog();
  const projectId = computed(() => Number(route.params.projectId));

  const loading = ref(false);
  const list = ref<TopicItem[]>([]);
  const statusFilter = ref('active');
  const pagination = reactive({ page: 1, size: 20 });
  const total = ref(0);

  function onFilterChange() {
    pagination.page = 1;
    load();
  }
  const expanded = ref<number | null>(null);

  function statusLabel(s: string): string {
    return { active: '进行中', completed: '已完成', abandoned: '已放弃' }[s] || s;
  }
  function toggle(id: number) {
    expanded.value = expanded.value === id ? null : id;
  }
  const taskDetailRef = ref();
  function openTask(taskId: number) {
    taskDetailRef.value?.openModal(taskId);
  }
  function statusText(s?: string): string {
    return s ? statusLabel(s) : '';
  }
  function statusTag(s?: string): 'default' | 'info' | 'warning' | 'success' | 'error' {
    return (s ? statusTagType(s) : 'default') as 'default' | 'info' | 'warning' | 'success' | 'error';
  }

  function openDoc(t: TopicItem) {
    router.push(`/project/${projectId.value}/knowledge?path=${encodeURIComponent(t.docPath)}`);
  }

  async function load() {
    loading.value = true;
    try {
      const res = await getTopics(projectId.value, { status: statusFilter.value, page: pagination.page, size: pagination.size });
      list.value = res?.list || [];
      total.value = res?.total || 0;
    } catch {
      message.error('加载专题失败');
    } finally {
      loading.value = false;
    }
  }

  // ---- 新建 ----
  const showCreate = ref(false);
  const creating = ref(false);
  const form = reactive({ title: '', goal: '', acceptance: '', docPath: '', assigneeId: null as number | null });
  const agentOptions = ref<Array<{ label: string; value: number }>>([]);

  function openCreate() {
    Object.assign(form, { title: '', goal: '', acceptance: '', docPath: '', assigneeId: null });
    loadAgentOptions();
    showCreate.value = true;
  }
  async function loadAgentOptions() {
    try {
      const res = await getMembers(projectId.value);
      agentOptions.value = (res?.list || [])
        .filter((m: any) => m.userType === 'ai')
        .map((m: any) => ({ label: m.realName || m.username, value: m.userId }));
    } catch { /* ignore */ }
  }
  async function handleCreate() {
    if (!form.title.trim()) return;
    creating.value = true;
    try {
      await createTopic(projectId.value, { ...form, assigneeId: form.assigneeId || 0 });
      message.success('专题已创建（阶段可在详情中拆解，或由 Agent 从 PRD 导入）');
      showCreate.value = false;
      load();
    } catch (e: any) {
      message.error(e?.message || '创建失败');
    } finally {
      creating.value = false;
    }
  }

  // ---- 阶段操作 ----
  const phaseInput = reactive<Record<number, string>>({});
  async function addPhase(t: TopicItem) {
    const v = (phaseInput[t.id] || '').trim();
    if (!v) return;
    try {
      await appendTopicPhase(projectId.value, t.id, v);
      phaseInput[t.id] = '';
      load();
    } catch (e: any) {
      message.error(e?.message || '添加失败');
    }
  }

  // 阶段交互：行点击开抽屉编辑，状态在抽屉内切
  const showPhaseDrawer = ref(false);
  const savingPhase = ref(false);
  const currentTopic = ref<TopicItem | null>(null);
  const currentPhase = ref<{ id: number } | null>(null);
  const editForm = reactive({
    title: '', detail: '', status: 'pending',
    assigneeId: null as number | null, artifacts: '',
  });
  const memberOptions = ref<Array<{ label: string; value: number }>>([]);
  async function loadMemberOptions() {
    if (memberOptions.value.length) return;
    try {
      const res = await getMembers(projectId.value);
      memberOptions.value = (res?.list || []).map((m: any) => ({ label: m.realName || m.username, value: m.userId }));
    } catch { /* ignore */ }
  }
  function phaseStatusText(st: string): string {
    return { pending: '待处理', in_progress: '进行中', done: '已完成' }[st] || st;
  }
  function phaseTag(st: string): 'default' | 'info' | 'success' {
    return ({ pending: 'default', in_progress: 'info', done: 'success' }[st] || 'default') as 'default' | 'info' | 'success';
  }
  // 描述/产出渲染前消毒（与任务详情同口径）
  function safeHtml(md: string): string {
    return DOMPurify.sanitize(md);
  }

  async function openPhaseDrawer(t: TopicItem, p: any) {
    await loadMemberOptions();
    currentTopic.value = t;
    currentPhase.value = { id: p.id };
    Object.assign(editForm, {
      title: p.title, detail: p.detail || '', status: p.status,
      assigneeId: p.assigneeId || null, artifacts: p.artifacts || '',
    });
    showPhaseDrawer.value = true;
  }
  async function savePhaseDrawer() {
    if (!currentTopic.value || !currentPhase.value || !editForm.title.trim()) return;
    savingPhase.value = true;
    try {
      // 状态与字段一起保存（状态走 toggle 端点，字段走 update）
      const t = currentTopic.value, pid = currentPhase.value.id;
      const d = await getTopics(projectId.value, { status: 'all', page: 1, size: 100 });
      const fresh = (d?.list || []).find((x) => x.id === t.id)?.phases.find((x) => x.id === pid);
      if (fresh && fresh.status !== editForm.status) {
        await toggleTopicPhase(projectId.value, t.id, pid, editForm.status);
      }
      await updateTopicPhase(projectId.value, t.id, pid, {
        title: editForm.title.trim(), detail: editForm.detail,
        assigneeId: editForm.assigneeId ?? 0, artifacts: editForm.artifacts,
      });
      message.success('阶段已保存');
      showPhaseDrawer.value = false;
      load();
    } catch (e: any) {
      message.error(e?.message || '保存失败');
    } finally {
      savingPhase.value = false;
    }
  }
  async function cyclePhase(t: TopicItem, p: { id: number; status: string }) {
    const next = p.status === 'pending' ? 'in_progress' : p.status === 'in_progress' ? 'done' : 'pending';
    try {
      await toggleTopicPhase(projectId.value, t.id, p.id, next);
      load();
    } catch (e: any) { message.error(e?.message || '操作失败'); }
  }
  function removePhase(t: TopicItem, p: { id: number; title: string }) {
    dialog.warning({
      title: '删除阶段',
      content: `删除「${p.title}」？`,
      positiveText: '删除',
      negativeText: '取消',
      onPositiveClick: async () => {
        try { await deleteTopicPhase(projectId.value, t.id, p.id); message.success('已删除'); load(); }
        catch (e: any) { message.error(e?.message || '删除失败'); }
      },
    });
  }

  function finish(t: TopicItem, result: 'completed' | 'abandoned') {
    dialog.warning({
      title: result === 'completed' ? '验收通过' : '放弃专题',
      content: result === 'completed'
        ? `确认「${t.title}」已满足验收标准？`
        : `放弃后「${t.title}」不再推进，确定？`,
      positiveText: '确定',
      negativeText: '取消',
      onPositiveClick: async () => {
        try {
          await finishTopic(projectId.value, t.id, result);
          message.success('已终态');
          load();
        } catch (e: any) { message.error(e?.message || '操作失败'); }
      },
    });
  }

  onMounted(load);
</script>

<style lang="less" scoped>
  .topic-card {
    border: 1px solid var(--border-color, #eef0f3);
    border-radius: 14px;
    padding: 13px 15px;
    margin-bottom: 12px;
  }
  .tp-head { display: flex; align-items: center; gap: 8px; cursor: pointer; }
  .tp-caret { color: var(--text-color-3, #999); width: 14px; }
  .tp-title { font-weight: 700; color: var(--text-color-1, #1f2329); flex: 1; }
  .tp-meta { display: flex; align-items: center; gap: 8px; font-size: 12px; color: var(--text-color-3, #9aa0a6); }
  .tp-bar { margin-top: 10px; }
  .tp-handoff { font-size: 12px; color: var(--text-color-2, #5c6470); background: #f7f8fa; border-radius: 8px; padding: 7px 10px; margin-top: 8px; }
  .tp-detail { border-top: 1px dashed var(--border-color, #eef0f3); margin-top: 12px; padding-top: 12px; }
  .tp-sec { margin-bottom: 12px; }
  .tp-label { font-size: 12px; font-weight: 700; color: var(--text-color-2, #5c6470); margin-bottom: 6px; }
  .tp-md { font-size: 13px; color: var(--text-color-2, #5c6470); white-space: pre-wrap; line-height: 1.7; }
  .tp-empty { font-size: 12px; color: var(--text-color-3, #9aa0a6); }
  .ph-table {
    display: grid;
    grid-template-columns: 34px 72px 1fr 110px 44px;
    gap: 8px;
    align-items: center;
    padding: 7px 8px;
  }
  .ph-head-row {
    font-size: 11px;
    font-weight: 600;
    color: var(--text-color-3, #9aa0a6);
    border-bottom: 1px solid var(--border-color, #eef0f3);
  }
  .ph-row {
    border-bottom: 1px dashed var(--border-color, #eef0f3);
    &.done .ph-title { color: var(--text-color-3, #999); text-decoration: line-through; }
  }
  .ph-status-tag { cursor: pointer; }
  .ph-c-idx { font-size: 12px; color: var(--text-color-3, #9aa0a6); font-variant-numeric: tabular-nums; }
  .ph-c-title { display: flex; align-items: center; gap: 5px; min-width: 0; cursor: pointer; }
  .ph-caret { color: var(--text-color-3, #999); width: 12px; flex: none; }
  .ph-title { overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
  .ph-task-tag { flex: none; cursor: pointer; }
  .ph-c-assignee { font-size: 12px; color: var(--text-color-2, #5c6470); overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
  .ph-c-ops { text-align: right; }
  .pd-label {
    font-size: 12px;
    font-weight: 700;
    color: var(--text-color-2, #5c6470);
    margin-bottom: 6px;
  }

  .phase-add { margin-top: 8px; }
  .tp-acts { display: flex; justify-content: flex-end; gap: 8px; margin-top: 4px; }
</style>
