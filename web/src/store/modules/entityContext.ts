import { defineStore } from 'pinia';
import { store } from '@/store';
import { ref } from 'vue';

export interface EntityInfo {
  id: number;
  name: string;
}

// 项目上下文：项目名由 Header 面包屑承载（原 EntityNavBar tabs 导航已删除）
export const useEntityContextStore = defineStore('entity-context', () => {
  const currentProject = ref<EntityInfo | null>(null);

  function setProject(info: EntityInfo | null) {
    currentProject.value = info;
  }

  function clearEntity() {
    currentProject.value = null;
  }

  return {
    currentProject,
    setProject,
    clearEntity,
  };
});

export function useEntityContext() {
  return useEntityContextStore(store);
}
