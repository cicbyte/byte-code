<template>
  <div>
    <n-card :bordered="false" class="proCard">
      <n-spin :show="loading">
        <template v-if="project">
          <n-descriptions label-placement="left" :column="2" bordered size="small" class="mb-4">
            <n-descriptions-item label="项目名称" :span="2">
              <span class="text-lg font-medium">{{ project.name }}</span>
            </n-descriptions-item>
            <n-descriptions-item label="状态">
              <n-tag :type="project.status === 1 ? 'success' : 'default'" size="small">
                {{ project.status === 1 ? '进行中' : '已结束' }}
              </n-tag>
            </n-descriptions-item>
            <n-descriptions-item label="创建人">{{ project.creatorName || '-' }}</n-descriptions-item>
            <n-descriptions-item label="产品">{{ project.productName || '-' }}</n-descriptions-item>
            <n-descriptions-item label="创建时间">{{ project.createdAt }}</n-descriptions-item>
            <n-descriptions-item label="描述" :span="2">
              {{ project.description || '暂无描述' }}
            </n-descriptions-item>
          </n-descriptions>
        </template>
        <n-empty v-else description="未找到项目信息" />
      </n-spin>
    </n-card>
  </div>
</template>

<script lang="ts" setup>
  import { ref, computed, onMounted } from 'vue';
  import { useRoute } from 'vue-router';
  import { getProject } from '@/api/project/index';
  import type { ProjectItem } from '@/api/project/index';

  const route = useRoute();
  const projectId = computed(() => Number(route.params.projectId));
  const loading = ref(false);
  const project = ref<ProjectItem | null>(null);

  async function loadProject() {
    loading.value = true;
    try {
      const res = await getProject(projectId.value);
      project.value = res || null;
    } catch {
      // ignore
    } finally {
      loading.value = false;
    }
  }

  onMounted(loadProject);
</script>
