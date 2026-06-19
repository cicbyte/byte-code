import { defineStore } from 'pinia';
import { store } from '@/store';
import { ref, computed } from 'vue';

export interface EntityInfo {
  id: number;
  name: string;
}

export const useEntityContextStore = defineStore('entity-context', () => {
  const currentProject = ref<EntityInfo | null>(null);

  const currentEntityType = computed<'project' | null>(() => {
    return currentProject.value ? 'project' : null;
  });

  const currentEntityName = computed(() => currentProject.value?.name || '');

  function setProject(info: EntityInfo | null) {
    currentProject.value = info;
  }

  function clearEntity() {
    currentProject.value = null;
  }

  return {
    currentProject,
    currentEntityType,
    currentEntityName,
    setProject,
    clearEntity,
  };
});

export function useEntityContext() {
  return useEntityContextStore(store);
}
