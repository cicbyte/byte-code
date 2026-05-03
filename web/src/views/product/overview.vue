<template>
  <div>
    <n-card :bordered="false" class="proCard">
      <n-spin :show="loading">
        <template v-if="product">
          <n-descriptions label-placement="left" :column="2" bordered size="small" class="mb-4">
            <n-descriptions-item label="产品名称" :span="2">
              <span class="text-lg font-medium">{{ product.name }}</span>
            </n-descriptions-item>
            <n-descriptions-item label="状态">
              <n-tag :type="product.status === 'active' ? 'success' : 'default'" size="small">
                {{ product.status === 'active' ? '启用' : '禁用' }}
              </n-tag>
            </n-descriptions-item>
            <n-descriptions-item label="负责人">{{ product.ownerName || '-' }}</n-descriptions-item>
            <n-descriptions-item label="创建时间">{{ product.createdAt }}</n-descriptions-item>
            <n-descriptions-item label="更新时间">{{ product.updatedAt }}</n-descriptions-item>
            <n-descriptions-item label="描述" :span="2">
              {{ product.description || '暂无描述' }}
            </n-descriptions-item>
          </n-descriptions>
        </template>
        <n-empty v-else description="未找到产品信息" />
      </n-spin>
    </n-card>
  </div>
</template>

<script lang="ts" setup>
  import { ref, computed, onMounted } from 'vue';
  import { useRoute } from 'vue-router';
  import { getProduct } from '@/api/product/index';
  import type { ProductItem } from '@/api/product/index';

  const route = useRoute();
  const productId = computed(() => Number(route.params.productId));
  const loading = ref(false);
  const product = ref<ProductItem | null>(null);

  async function loadProduct() {
    loading.value = true;
    try {
      const res = await getProduct(productId.value);
      product.value = res || null;
    } catch {
      // ignore
    } finally {
      loading.value = false;
    }
  }

  onMounted(loadProduct);
</script>
