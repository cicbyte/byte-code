<template>
  <div class="console">
    <n-grid cols="1 s:2 m:2 l:4 xl:4 2xl:4" responsive="screen" :x-gap="12" :y-gap="8">
      <n-grid-item>
        <n-card title="用户总数" size="small" :bordered="false">
          <template #header-extra>
            <n-icon size="24" color="#69c0ff"><UsergroupAddOutlined /></n-icon>
          </template>
          <n-skeleton v-if="loading" :width="60" size="medium" />
          <CountTo v-else :startVal="0" :endVal="data.userCount" class="text-3xl" />
          <template #footer>
            <span class="text-gray-400">系统注册用户数</span>
          </template>
        </n-card>
      </n-grid-item>

      <n-grid-item>
        <n-card title="角色总数" size="small" :bordered="false">
          <template #header-extra>
            <n-icon size="24" color="#b37feb"><TeamOutlined /></n-icon>
          </template>
          <n-skeleton v-if="loading" :width="60" size="medium" />
          <CountTo v-else :startVal="0" :endVal="data.roleCount" class="text-3xl" />
          <template #footer>
            <span class="text-gray-400">系统角色数量</span>
          </template>
        </n-card>
      </n-grid-item>

      <n-grid-item>
        <n-card title="菜单总数" size="small" :bordered="false">
          <template #header-extra>
            <n-icon size="24" color="#5cdbd3"><MenuOutlined /></n-icon>
          </template>
          <n-skeleton v-if="loading" :width="60" size="medium" />
          <CountTo v-else :startVal="0" :endVal="data.menuCount" class="text-3xl" />
          <template #footer>
            <span class="text-gray-400">系统菜单数量</span>
          </template>
        </n-card>
      </n-grid-item>

      <n-grid-item>
        <n-card title="在线用户" size="small" :bordered="false">
          <template #header-extra>
            <n-icon size="24" color="#95de64"><CheckCircleOutlined /></n-icon>
          </template>
          <n-skeleton v-if="loading" :width="60" size="medium" />
          <CountTo v-else :startVal="0" :endVal="data.onlineUser" class="text-3xl" />
          <template #footer>
            <span class="text-gray-400">当前在线 Token 数</span>
          </template>
        </n-card>
      </n-grid-item>
    </n-grid>

    <n-card title="系统信息" class="mt-4" :bordered="false">
      <n-descriptions label-placement="left" :column="2" bordered>
        <n-descriptions-item label="系统名称">ByteAdmin</n-descriptions-item>
        <n-descriptions-item label="前端框架">Vue 3 + Naive UI</n-descriptions-item>
        <n-descriptions-item label="后端框架">Go (GoFrame v2)</n-descriptions-item>
        <n-descriptions-item label="数据库">SQLite</n-descriptions-item>
      </n-descriptions>
    </n-card>
  </div>
</template>

<script lang="ts" setup>
  import { ref, reactive, onMounted } from 'vue';
  import { getConsoleInfo } from '@/api/dashboard/console';
  import { CountTo } from '@/components/CountTo/index';
  import {
    UsergroupAddOutlined,
    TeamOutlined,
    MenuOutlined,
    CheckCircleOutlined,
  } from '@vicons/antd';

  const loading = ref(true);
  const data = reactive({
    userCount: 0,
    roleCount: 0,
    menuCount: 0,
    onlineUser: 0,
  });

  onMounted(async () => {
    try {
      const res = await getConsoleInfo();
      if (res) {
        data.userCount = res.userCount || 0;
        data.roleCount = res.roleCount || 0;
        data.menuCount = res.menuCount || 0;
        data.onlineUser = res.onlineUser || 0;
      }
    } catch (e) {
      // ignore
    } finally {
      loading.value = false;
    }
  });
</script>
