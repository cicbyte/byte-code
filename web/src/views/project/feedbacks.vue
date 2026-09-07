<template>
  <div>
    <n-card :bordered="false" class="proCard">
      <template #header>
        跨项目反馈
        <n-tag v-if="openCount" type="warning" size="small" class="ml-2">{{ openCount }} 条待分析</n-tag>
      </template>
      <template #header-extra>
        <n-space :size="8">
          <n-radio-group v-model:value="statusFilter" size="small" @update:value="onFilter">
            <n-radio-button value="open">待分析</n-radio-button>
            <n-radio-button value="converted">已转化</n-radio-button>
            <n-radio-button value="dismissed">已忽略</n-radio-button>
            <n-radio-button value="all">全部</n-radio-button>
          </n-radio-group>
          <n-button size="small" type="primary" ghost @click="openSend">投递反馈</n-button>
        </n-space>
      </template>

      <n-spin :show="loading">
        <EmptyState
          v-if="!loading && list.length === 0"
          type="notify"
          title="没有反馈"
          description="其他项目发现的线索会投递到这里，由你或 Agent 分析是否建立任务"
        />
        <div v-for="fb in list" :key="fb.id" class="fb-card">
          <div class="fb-head">
            <n-tag size="small" :type="fb.status === 'open' ? 'warning' : fb.status === 'converted' ? 'success' : 'default'">
              {{ statusLabel(fb.status) }}
            </n-tag>
            <span class="fb-title">{{ fb.title }}</span>
            <span class="fb-src">来自 {{ fb.sourceProjectName }}<template v-if="fb.sourceTaskId"> · 任务 #{{ fb.sourceTaskId }}</template></span>
          </div>
          <div v-if="fb.content" class="fb-content">{{ fb.content }}</div>
          <div v-if="fb.status === 'converted' && fb.convertedTaskId" class="fb-result">
            已转化为任务 #{{ fb.convertedTaskId }}
          </div>
          <div v-if="fb.status === 'dismissed' && fb.dismissReason" class="fb-result dismissed">
            未采纳：{{ fb.dismissReason }}
          </div>
          <div v-if="fb.status === 'open'" class="fb-acts">
            <n-button size="tiny" type="error" ghost @click="openDismiss(fb)">忽略</n-button>
            <n-button size="tiny" type="success" @click="handleConvert(fb)">转化为任务</n-button>
          </div>
          <div class="fb-time">{{ (fb.createdAt || '').slice(0, 16) }}</div>
        </div>
      </n-spin>
    </n-card>

    <!-- 投递反馈：目标须是已关联且可访问的项目 -->
    <n-modal v-model:show="showSend" preset="dialog" title="投递跨项目反馈" :show-icon="false">
      <n-space vertical :size="10" class="py-2">
        <n-select
          v-model:value="sendForm.projectId"
          :options="sendTargets"
          placeholder="目标项目（须已建立关联）"
          size="small"
        />
        <n-input v-model:value="sendForm.title" placeholder="反馈标题（一句话说清线索）" size="small" />
        <n-input
          v-model:value="sendForm.content"
          type="textarea"
          placeholder="现象 / 线索 / 怀疑点（markdown），对方 Agent 将阅读分析是否建任务"
          :rows="4"
        />
      </n-space>
      <template #action>
        <n-space>
          <n-button size="small" @click="showSend = false">取消</n-button>
          <n-button size="small" type="primary" :disabled="!sendForm.projectId || !sendForm.title.trim()" :loading="sending" @click="handleSend">
            投递
          </n-button>
        </n-space>
      </template>
    </n-modal>

    <!-- 忽略理由 -->
    <n-modal v-model:show="showDismiss" preset="dialog" title="忽略反馈" :show-icon="false">
      <n-input v-model:value="dismissReason" type="textarea" placeholder="忽略理由（必填，将通知发起方）" :rows="3" />
      <template #action>
        <n-space>
          <n-button size="small" @click="showDismiss = false">取消</n-button>
          <n-button size="small" type="error" :disabled="!dismissReason.trim()" @click="handleDismiss">确认忽略</n-button>
        </n-space>
      </template>
    </n-modal>
  </div>
</template>

<script lang="ts" setup>
  import { ref, computed, reactive, onMounted } from 'vue';
  import { useRoute } from 'vue-router';
  import { useMessage } from 'naive-ui';
  import EmptyState from '@/components/EmptyState/EmptyState.vue';
  import {
    getFeedbacks,
    createFeedback,
    convertFeedback,
    dismissFeedback,
    getRelations,
    getProjects,
  } from '@/api/project/index';
  import type { FeedbackItem } from '@/api/project/index';

  const route = useRoute();
  const message = useMessage();
  const projectId = computed(() => Number(route.params.projectId));

  const loading = ref(false);
  const list = ref<FeedbackItem[]>([]);
  const statusFilter = ref('open');
  const openCount = computed(() => list.value.length);

  function statusLabel(s: string): string {
    return { open: '待分析', converted: '已转化', dismissed: '已忽略' }[s] || s;
  }

  async function load() {
    loading.value = true;
    try {
      const res = await getFeedbacks(projectId.value, { status: statusFilter.value, size: 50 });
      list.value = res?.list || [];
    } catch {
      message.error('加载反馈失败');
    } finally {
      loading.value = false;
    }
  }

  function onFilter() { load(); }

  // ---- 投递 ----
  const showSend = ref(false);
  const sending = ref(false);
  const sendForm = reactive({ projectId: null as number | null, title: '', content: '' });
  const sendTargets = ref<Array<{ label: string; value: number }>>([]);

  async function openSend() {
    // 目标 = 本项目已建立关联、且当前账号可访问的项目
    try {
      const [rel, proj] = await Promise.all([getRelations(projectId.value), getProjects({ size: 100 })]);
      const relatedIds = new Set((rel?.list || []).map((r) => r.projectId));
      sendTargets.value = (proj?.list || [])
        .filter((p: any) => relatedIds.has(p.id) && p.id !== projectId.value)
        .map((p: any) => ({ label: p.name, value: p.id }));
      if (sendTargets.value.length === 0) {
        message.warning('还没有关联项目（项目管理员先在成员管理页建立关联）');
      }
    } catch {
      message.error('加载关联项目失败');
    }
    sendForm.projectId = null;
    sendForm.title = '';
    sendForm.content = '';
    showSend.value = true;
  }

  async function handleSend() {
    if (!sendForm.projectId || !sendForm.title.trim()) return;
    sending.value = true;
    try {
      await createFeedback(sendForm.projectId, { title: sendForm.title.trim(), content: sendForm.content });
      message.success('反馈已投递，对方成员与 Agent 会收到通知');
      showSend.value = false;
    } catch (e: any) {
      message.error(e?.message || '投递失败');
    } finally {
      sending.value = false;
    }
  }

  // ---- 转化 / 忽略 ----
  async function handleConvert(fb: FeedbackItem) {
    try {
      const res = await convertFeedback(projectId.value, fb.id, {});
      message.success(`已转化为任务 #${res?.taskId}（已回告发起方）`);
      load();
    } catch (e: any) {
      message.error(e?.message || '转化失败');
    }
  }

  const showDismiss = ref(false);
  const dismissReason = ref('');
  const dismissing = ref<FeedbackItem | null>(null);
  function openDismiss(fb: FeedbackItem) {
    dismissing.value = fb;
    dismissReason.value = '';
    showDismiss.value = true;
  }
  async function handleDismiss() {
    if (!dismissing.value || !dismissReason.value.trim()) return;
    try {
      await dismissFeedback(projectId.value, dismissing.value.id, dismissReason.value.trim());
      message.success('已忽略并回告发起方');
      showDismiss.value = false;
      load();
    } catch (e: any) {
      message.error(e?.message || '操作失败');
    }
  }

  onMounted(load);
</script>

<style lang="less" scoped>
  .fb-card {
    border: 1px solid var(--border-color, #eef0f3);
    border-radius: 12px;
    padding: 13px 15px;
    margin-bottom: 11px;
  }
  .fb-head {
    display: flex;
    align-items: center;
    gap: 8px;
    flex-wrap: wrap;
  }
  .fb-title { font-weight: 600; color: var(--text-color-1, #1f2329); }
  .fb-src { font-size: 12px; color: var(--text-color-3, #9aa0a6); margin-left: auto; }
  .fb-content {
    font-size: 13px;
    color: var(--text-color-2, #5c6470);
    margin-top: 8px;
    white-space: pre-wrap;
    line-height: 1.7;
  }
  .fb-result { font-size: 12px; color: #18a058; margin-top: 8px; }
  .fb-result.dismissed { color: var(--text-color-3, #9aa0a6); }
  .fb-acts { display: flex; gap: 8px; justify-content: flex-end; margin-top: 10px; }
  .fb-time { font-size: 11px; color: var(--text-color-3, #9aa0a6); margin-top: 6px; }
</style>
