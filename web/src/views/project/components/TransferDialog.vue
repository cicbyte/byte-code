<template>
  <n-modal
    :show="show"
    preset="dialog"
    title="转交项目负责人"
    :show-icon="false"
    style="width: 480px"
    @update:show="$emit('update:show', $event)"
  >
    <n-space vertical :size="10" class="py-2">
      <span class="text-sm">搜索选择接手人，对方将在通知中心收到邀请，确认接受后完成移交。</span>
      <n-select
        v-model:value="targetId"
        :options="userOptions"
        filterable
        remote
        clearable
        placeholder="输入用户名 / 姓名搜索（仅人类账号）"
        :loading="searching"
        @search="onSearch"
      />
      <n-radio-group v-model:value="leave">
        <n-space vertical>
          <n-radio :value="false">留在本项目 —— 移交完成后您成为普通成员</n-radio>
          <n-radio :value="true">退出本项目 —— 移交完成后移出成员列表（需要隔离时选择）</n-radio>
        </n-space>
      </n-radio-group>
      <n-text v-if="leave" depth="3" style="font-size: 12px">
        退出后您将失去本项目全部访问权限，如需回来需新负责人重新添加
      </n-text>
    </n-space>
    <template #action>
      <n-space>
        <n-button size="small" @click="$emit('update:show', false)">取消</n-button>
        <n-button size="small" type="warning" :loading="submitting" @click="submit">发送邀请</n-button>
      </n-space>
    </template>
  </n-modal>
</template>

<script lang="ts" setup>
  import { ref, watch } from 'vue';
  import { useMessage } from 'naive-ui';
  import { searchUsers } from '@/api/system/user';
  import { inviteOwnerTransfer } from '@/api/project/index';

  const props = defineProps<{ show: boolean; projectId: number }>();
  const emit = defineEmits<{ (e: 'update:show', v: boolean): void; (e: 'invited'): void }>();

  const message = useMessage();
  const targetId = ref<number | null>(null);
  const userOptions = ref<Array<{ label: string; value: number }>>([]);
  const searching = ref(false);
  const leave = ref(false);
  const submitting = ref(false);
  let searchTimer: ReturnType<typeof setTimeout> | null = null;

  watch(
    () => props.show,
    (v) => {
      if (v) {
        targetId.value = null;
        userOptions.value = [];
        leave.value = false;
      }
    }
  );

  // 用户远程搜索（复用 /v1/users/search，300ms 防抖；与添加成员选择器同源）
  function onSearch(q: string) {
    if (searchTimer) clearTimeout(searchTimer);
    if (!q.trim()) {
      userOptions.value = [];
      return;
    }
    searchTimer = setTimeout(async () => {
      searching.value = true;
      try {
        const res = await searchUsers(q.trim());
        userOptions.value = (res?.list || []).map((u) => ({
          label: `${u.realName || u.username}（${u.username} #${u.id}）`,
          value: u.id,
        }));
      } catch {
        // http 层统一提示
      } finally {
        searching.value = false;
      }
    }, 300);
  }

  async function submit() {
    if (!targetId.value) {
      message.warning('请先搜索选择接手人');
      return;
    }
    submitting.value = true;
    try {
      await inviteOwnerTransfer(props.projectId, { userId: targetId.value, leave: leave.value });
      message.success('邀请已发送，对方在通知中心接受后完成移交');
      emit('update:show', false);
      emit('invited');
    } catch {
      // http 层统一提示（如同项目已有待响应邀请）
    } finally {
      submitting.value = false;
    }
  }
</script>
