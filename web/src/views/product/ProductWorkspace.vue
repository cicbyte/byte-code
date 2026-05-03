<template>
  <router-view />
</template>

<script lang="ts" setup>
  import { onMounted, onUnmounted } from 'vue';
  import { useRoute } from 'vue-router';
  import { useEntityContext } from '@/store/modules/entityContext';
  import { getProduct } from '@/api/product/index';

  const route = useRoute();
  const entityContext = useEntityContext();

  onMounted(async () => {
    const productId = Number(route.params.productId);
    if (!productId) return;
    try {
      const res = await getProduct(productId);
      if (res) {
        entityContext.setProduct({ id: res.id, name: res.name });
      }
    } catch {
      // ignore
    }
  });

  onUnmounted(() => {
    entityContext.clearEntity();
  });
</script>
