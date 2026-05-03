import { defineStore } from 'pinia';
import { store } from '@/store';
import { ref, computed } from 'vue';

export interface EntityInfo {
  id: number;
  name: string;
}

export const useEntityContextStore = defineStore('entity-context', () => {
  const currentProject = ref<EntityInfo | null>(null);
  const currentProduct = ref<EntityInfo | null>(null);

  const currentEntityType = computed<'project' | 'product' | null>(() => {
    if (currentProject.value) return 'project';
    if (currentProduct.value) return 'product';
    return null;
  });

  const currentEntityName = computed(() => {
    return currentProject.value?.name || currentProduct.value?.name || '';
  });

  function setProject(info: EntityInfo | null) {
    currentProject.value = info;
    if (info) currentProduct.value = null;
  }

  function setProduct(info: EntityInfo | null) {
    currentProduct.value = info;
    if (info) currentProject.value = null;
  }

  function clearEntity() {
    currentProject.value = null;
    currentProduct.value = null;
  }

  return {
    currentProject,
    currentProduct,
    currentEntityType,
    currentEntityName,
    setProject,
    setProduct,
    clearEntity,
  };
});

export function useEntityContext() {
  return useEntityContextStore(store);
}
