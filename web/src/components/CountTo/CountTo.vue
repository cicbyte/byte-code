<template>
  <span :style="{ color }">
    {{ value }}
  </span>
</template>
<script lang="ts" setup>
  import { ref, computed, watchEffect, unref, onMounted, watch } from 'vue';
  import { useTransition, TransitionPresets } from '@vueuse/core';
  import { isNumber } from '@/utils/is';

  defineOptions({ name: 'CountTo' });

  const props = defineProps({
    startVal: { type: Number, default: 0 },
    endVal: { type: Number, default: 2021 },
    duration: { type: Number, default: 1500 },
    autoplay: { type: Boolean, default: true },
    decimals: {
      type: Number,
      default: 0,
      validator(value: number) {
        return value >= 0;
      },
    },
    prefix: { type: String, default: '' },
    suffix: { type: String, default: '' },
    separator: { type: String, default: ',' },
    decimal: { type: String, default: '.' },
    /**
     * font color
     */
    color: { type: String },
    /**
     * Turn on digital animation
     */
    useEasing: { type: Boolean, default: true },
    /**
     * Digital animation
     */
    transition: { type: String, default: 'linear' },
  });

  const emit = defineEmits(['onStarted', 'onFinished']);

  const source = ref(props.startVal);
  // 单一 transition 实例：此前 run() 里重绑 let outputValue 会让模板 computed
  // 的依赖与旧实例脱钩，动画期间 computed 不重算——仪表盘统计卡长期显示 0 的根因。
  // 正确姿势：实例只在 setup 建一次，start() 仅改 source 让其动画到新值。
  const outputValue = useTransition(source, {
    duration: props.duration,
    onFinished: () => emit('onFinished'),
    onStarted: () => emit('onStarted'),
    ...(props.useEasing ? { transition: TransitionPresets[props.transition] } : {}),
  });

  const value = computed(() => formatNumber(unref(outputValue)));

  watchEffect(() => {
    source.value = props.startVal;
  });

  watch([() => props.startVal, () => props.endVal], () => {
    if (props.autoplay) {
      start();
    }
  });

  onMounted(() => {
    props.autoplay && start();
  });

  function start() {
    source.value = props.endVal;
  }

  function reset() {
    source.value = props.startVal;
  }

  function formatNumber(num: number | string) {
    if (!num && num !== 0) {
      return '';
    }
    const { decimals, decimal, separator, suffix, prefix } = props;
    num = Number(num).toFixed(decimals);
    num += '';

    const x = num.split('.');
    let x1 = x[0];
    const x2 = x.length > 1 ? decimal + x[1] : '';

    const rgx = /(\d+)(\d{3})/;
    if (separator && !isNumber(separator)) {
      while (rgx.test(x1)) {
        x1 = x1.replace(rgx, '$1' + separator + '$2');
      }
    }
    return prefix + x1 + x2 + suffix;
  }

  defineExpose({
    reset,
  });
</script>
