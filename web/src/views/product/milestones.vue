<template>
  <div>
    <n-card :bordered="false" class="proCard">
      <template #header-extra>
        <n-button type="primary" @click="showCreate = true">新建里程碑</n-button>
      </template>

      <n-spin :show="loading">
        <n-empty v-if="!loading && milestones.length === 0" description="暂无里程碑" />
        <n-table v-else :bordered="false" :single-line="false" size="small">
          <thead>
            <tr>
              <th>名称</th>
              <th>描述</th>
              <th>目标日期</th>
              <th>状态</th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="item in milestones" :key="item.id">
              <td>{{ item.name }}</td>
              <td>{{ item.description }}</td>
              <td>{{ item.targetDate }}</td>
              <td>
                <n-tag size="small">{{ item.status }}</n-tag>
              </td>
            </tr>
          </tbody>
        </n-table>
      </n-spin>
    </n-card>

    <n-modal v-model:show="showCreate" title="新建里程碑" preset="card" style="width: 500px">
      <n-form ref="formRef" :model="form" :rules="rules" label-placement="left" label-width="80">
        <n-form-item label="名称" path="name">
          <n-input v-model:value="form.name" placeholder="里程碑名称" />
        </n-form-item>
        <n-form-item label="描述" path="description">
          <n-input v-model:value="form.description" type="textarea" placeholder="描述" />
        </n-form-item>
        <n-form-item label="目标日期" path="targetDate">
          <n-date-picker v-model:formatted-value="form.targetDate" type="date" placeholder="选择目标日期" style="width: 100%" />
        </n-form-item>
      </n-form>
      <template #action>
        <n-space>
          <n-button @click="showCreate = false">取消</n-button>
          <n-button type="primary" @click="handleCreate" :loading="submitting">创建</n-button>
        </n-space>
      </template>
    </n-modal>
  </div>
</template>

<script lang="ts" setup>
  import { ref, computed, onMounted } from 'vue';
  import { useRoute } from 'vue-router';
  import { useMessage } from 'naive-ui';
  import { getMilestones, createMilestone } from '@/api/product/index';
  import type { MilestoneItem } from '@/api/product/index';

  const route = useRoute();
  const message = useMessage();
  const productId = computed(() => Number(route.params.productId));

  const loading = ref(false);
  const submitting = ref(false);
  const milestones = ref<MilestoneItem[]>([]);
  const showCreate = ref(false);

  const form = ref({ name: '', description: '', targetDate: '' });
  const rules = { name: { required: true, message: '请输入名称' } };

  async function loadMilestones() {
    loading.value = true;
    try {
      const res = await getMilestones(productId.value);
      milestones.value = res?.list || [];
    } catch {
      // ignore
    } finally {
      loading.value = false;
    }
  }

  async function handleCreate() {
    submitting.value = true;
    try {
      await createMilestone(productId.value, { ...form.value });
      message.success('创建成功');
      showCreate.value = false;
      form.value = { name: '', description: '', targetDate: '' };
      loadMilestones();
    } catch (e: any) {
      message.error(e.message || '创建失败');
    } finally {
      submitting.value = false;
    }
  }

  onMounted(loadMilestones);
</script>
