<template>
  <div>
    <n-grid :x-gap="24">
      <n-grid-item span="6">
        <n-card :bordered="false" size="small" class="proCard">
          <n-thing
            class="thing-cell"
            v-for="item in typeTabList"
            :key="item.key"
            :class="{ 'thing-cell-on': state.type === item.key }"
            @click="switchType(item)"
          >
            <template #header>{{ item.name }}</template>
            <template #description>{{ item.desc }}</template>
          </n-thing>
        </n-card>
      </n-grid-item>
      <n-grid-item span="18">
        <n-card :bordered="false" size="small" :title="state.typeTitle" class="proCard">
          <BasicSetting v-if="state.type === 1" />
          <EmailSetting v-if="state.type === 2" />
        </n-card>
      </n-grid-item>
    </n-grid>
  </div>
</template>
<script lang="ts" setup>
  import { reactive } from 'vue';
  import BasicSetting from './components/BasicSetting.vue';
  import EmailSetting from './components/EmailSetting.vue';

  const typeTabList = [
    {
      name: '基本设置',
      desc: '系统常规设置',
      key: 1,
    },
    {
      name: '邮件设置',
      desc: '系统邮件设置',
      key: 2,
    },
  ];

  const state = reactive({
    type: 1,
    typeTitle: '基本设置',
  });

  function switchType(e) {
    state.type = e.key;
    state.typeTitle = e.name;
  }
</script>
<style lang="less" scoped>
  .thing-cell {
    margin: 0 -16px 10px;
    padding: 5px 16px;

    &:hover {
      background: var(--hover-bg);
      cursor: pointer;
    }
  }

  .thing-cell-on {
    background: var(--hover-bg);
    color: #16a34a;

    ::v-deep(.n-thing-main .n-thing-header .n-thing-header__title) {
      color: #16a34a;
    }

    &:hover {
      background: var(--hover-bg);
    }
  }
</style>
