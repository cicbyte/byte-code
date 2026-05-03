<template>
  <router-view />
</template>

<script lang="ts" setup>
  import { onMounted, onUnmounted } from 'vue';
  import { useRoute } from 'vue-router';
  import { useEntityContext } from '@/store/modules/entityContext';
  import { getProject } from '@/api/project/index';

  const route = useRoute();
  const entityContext = useEntityContext();

  onMounted(async () => {
    const projectId = Number(route.params.projectId);
    if (!projectId) return;
    try {
      const res = await getProject(projectId);
      if (res) {
        entityContext.setProject({ id: res.id, name: res.name });
      }
    } catch {
      // ignore
    }
  });

  onUnmounted(() => {
    entityContext.clearEntity();
  });
</script>
