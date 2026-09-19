<template>
  <div>
    <n-space vertical :size="16">
      <n-alert type="info" :bordered="false" v-if="isAdmin">
        管理员账号：平台全部功能开放（字典授权不适用于超管）
      </n-alert>
      <n-alert type="warning" :bordered="false" v-else-if="!loading && perms.length === 0">
        当前账号未授予任何权限项——请联系管理员在「系统设置 → 角色权限」中为你的角色勾选菜单
      </n-alert>

      <div>
        <n-space :size="8" align="center" class="mb-3">
          <span class="text-sm">我的权限（{{ perms.length }} 项）</span>
          <n-button size="tiny" :loading="loading" @click="load">刷新</n-button>
        </n-space>
        <n-space :size="8">
          <n-tooltip v-for="p in perms" :key="p.value" trigger="hover" :data-test-id="`setting-account-components.item-${p.value}`">
            <template #trigger>
              <n-tag size="small" :bordered="false" round>{{ p.label }}</n-tag>
            </template>
            {{ p.value }}
          </n-tooltip>
        </n-space>
        <div class="text-xs text-gray-400 mt-3">
          悬停查看权限标识；权限由角色授予（角色权限页勾选即生效），项目内操作另受项目角色（负责人/维护者/成员）约束
        </div>
      </div>
    </n-space>
  </div>
</template>

<script lang="ts" setup>
  import { computed, onMounted, ref } from 'vue';
  import { useUserStore } from '@/store/modules/user';
  import { getUserInfo } from '@/api/system/user';
  import { usePerm } from '@/composables/usePerm';

  interface PermissionItem {
    label: string;
    value: string;
  }

  const userStore = useUserStore();
  const { isAdmin } = usePerm();
  const loading = ref(false);
  // 登录时 admin_info 已带一份；刷新保角色调整后的新鲜度
  const perms = computed<PermissionItem[]>(
    () => (userStore.permissions as PermissionItem[]) || [],
  );

  async function load() {
    loading.value = true;
    try {
      const data = await getUserInfo();
      const list = (data?.result?.permissions as PermissionItem[]) || [];
      userStore.setPermissions(list);
    } catch {
      // 保留 store 现值（admin_info 失败不阻断展示）
    } finally {
      loading.value = false;
    }
  }

  onMounted(load);
</script>
