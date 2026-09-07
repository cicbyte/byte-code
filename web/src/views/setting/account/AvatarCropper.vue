<template>
  <n-modal
    :show="show"
    preset="card"
    title="裁剪头像"
    style="width: 400px"
    :mask-closable="false"
    @update:show="(v: boolean) => emit('update:show', v)"
  >
    <div class="crop-body">
      <div
        class="crop-viewport"
        @pointerdown="onPointerDown"
        @pointermove="onPointerMove"
        @pointerup="onPointerUp"
        @pointercancel="onPointerUp"
        @wheel.prevent="onWheel"
      >
        <img
          v-if="imgUrl"
          :src="imgUrl"
          :style="imgStyle"
          :width="baseW"
          :height="baseH"
          draggable="false"
          alt="待裁剪头像"
        />
        <!-- 三分线参考 -->
        <span class="grid-v" style="left: 33.33%"></span>
        <span class="grid-v" style="left: 66.66%"></span>
        <span class="grid-h" style="top: 33.33%"></span>
        <span class="grid-h" style="top: 66.66%"></span>
      </div>
      <div class="crop-controls">
        <n-button text size="tiny" @click="resetView">重置</n-button>
        <n-slider :value="zoom" :min="1" :max="4" :step="0.02" :tooltip="false" class="crop-slider" :on-update:value="setZoom" />
        <span class="crop-zoom-val">{{ Math.round(zoom * 100) }}%</span>
      </div>
      <div class="crop-hint">拖动调整位置，滑杆或滚轮缩放；输出 256×256 方形头像</div>
    </div>
    <template #action>
      <n-space>
        <n-button @click="emit('update:show', false)">取消</n-button>
        <n-button type="primary" :loading="uploading" :disabled="!ready" @click="confirmCrop">
          确认并上传
        </n-button>
      </n-space>
    </template>
  </n-modal>
</template>

<script lang="ts" setup>
  import { computed, ref, watch, onBeforeUnmount } from 'vue';

  // 视口与输出尺寸（正方形）
  const VIEW = 260;
  const OUT = 256;

  const props = defineProps<{
    show: boolean;
    file: File | null;
    uploading?: boolean;
  }>();
  const emit = defineEmits<{
    (e: 'update:show', v: boolean): void;
    (e: 'confirm', file: File): void;
  }>();

  const imgUrl = ref('');
  const ready = ref(false);
  // zoom=1 时图像铺满视口的基准显示尺寸（cover），随图片载入一次性确定
  const baseW = ref(0);
  const baseH = ref(0);
  const zoom = ref(1);
  const pos = ref({ x: 0, y: 0 });

  let imgEl: HTMLImageElement | null = null;
  let urlRevoke = '';
  const dragState = { active: false, px: 0, py: 0, ox: 0, oy: 0 };

  const dispW = computed(() => baseW.value * zoom.value);
  const dispH = computed(() => baseH.value * zoom.value);

  // 平移用 translate、缩放用统一 scale（transformOrigin 0 0）：
  // 均匀缩放物理上不可能产生斜切，几何与导出采样共用同一套数值
  const imgStyle = computed(() => ({
    transform: `translate(${pos.value.x}px, ${pos.value.y}px) scale(${zoom.value})`,
    transformOrigin: '0 0',
    cursor: dragState.active ? 'grabbing' : 'grab',
    userSelect: 'none',
    willChange: 'transform',
  }));

  function clampPos() {
    pos.value.x = Math.min(0, Math.max(VIEW - dispW.value, pos.value.x));
    pos.value.y = Math.min(0, Math.max(VIEW - dispH.value, pos.value.y));
  }

  // 缩放围绕视口中心：换算中心采样点后放回中心（单一入口，无 watch 时序问题）
  function setZoom(z: number) {
    const nz = Math.min(4, Math.max(1, z));
    if (nz === zoom.value) return;
    const cx = (VIEW / 2 - pos.value.x) / zoom.value;
    const cy = (VIEW / 2 - pos.value.y) / zoom.value;
    zoom.value = nz;
    pos.value.x = VIEW / 2 - cx * nz;
    pos.value.y = VIEW / 2 - cy * nz;
    clampPos();
  }

  function resetView() {
    zoom.value = 1;
    pos.value = { x: (VIEW - baseW.value) / 2, y: (VIEW - baseH.value) / 2 };
    clampPos();
  }

  function loadImage(file: File) {
    if (urlRevoke) URL.revokeObjectURL(urlRevoke);
    ready.value = false;
    imgUrl.value = '';
    const url = URL.createObjectURL(file);
    urlRevoke = url;
    const im = new Image();
    im.onload = () => {
      imgEl = im;
      // cover 基准：短边贴满视口，保证视口内永远是图像而非空白
      const base = VIEW / Math.min(im.naturalWidth, im.naturalHeight);
      baseW.value = Math.round(im.naturalWidth * base);
      baseH.value = Math.round(im.naturalHeight * base);
      imgUrl.value = url;
      ready.value = true;
      resetView();
    };
    im.onerror = () => {
      imgUrl.value = '';
    };
    im.src = url;
  }

  watch(
    () => props.file,
    (f) => {
      if (f) loadImage(f);
    }
  );

  function onPointerDown(e: PointerEvent) {
    if (!ready.value) return;
    (e.currentTarget as HTMLElement).setPointerCapture(e.pointerId);
    dragState.active = true;
    dragState.px = e.clientX;
    dragState.py = e.clientY;
    dragState.ox = pos.value.x;
    dragState.oy = pos.value.y;
  }

  function onPointerMove(e: PointerEvent) {
    if (!dragState.active) return;
    pos.value.x = dragState.ox + (e.clientX - dragState.px);
    pos.value.y = dragState.oy + (e.clientY - dragState.py);
    clampPos();
  }

  function onPointerUp() {
    dragState.active = false;
  }

  function onWheel(e: WheelEvent) {
    if (!ready.value) return;
    setZoom(zoom.value - e.deltaY * 0.0015);
  }

  function confirmCrop() {
    if (!ready.value || !imgEl) return;
    // 图像以 baseW/baseH（zoom=1 基准）渲染，scale 后有效显示尺寸 = base*zoom；
    // 视口左上角对应原图采样矩形（natural 坐标）
    const sx = (-pos.value.x / zoom.value) * (imgEl.naturalWidth / baseW.value);
    const sy = (-pos.value.y / zoom.value) * (imgEl.naturalHeight / baseH.value);
    const sw = (VIEW / zoom.value) * (imgEl.naturalWidth / baseW.value);
    const canvas = document.createElement('canvas');
    canvas.width = OUT;
    canvas.height = OUT;
    const ctx = canvas.getContext('2d');
    if (!ctx) return;
    ctx.imageSmoothingQuality = 'high';
    ctx.drawImage(imgEl, sx, sy, sw, sw, 0, 0, OUT, OUT);
    canvas.toBlob(
      (blob) => {
        if (!blob) return;
        // 裁剪后统一输出静态 PNG（动图取首帧）
        emit('confirm', new File([blob], 'avatar.png', { type: 'image/png' }));
      },
      'image/png'
    );
  }

  onBeforeUnmount(() => {
    if (urlRevoke) URL.revokeObjectURL(urlRevoke);
  });
</script>

<style lang="less" scoped>
  .crop-body {
    display: flex;
    flex-direction: column;
    gap: 12px;
    align-items: center;
  }

  .crop-viewport {
    position: relative;
    width: 260px;
    height: 260px;
    overflow: hidden;
    background: #000;
    border-radius: 8px;
    touch-action: none;

    img {
      position: absolute;
      left: 0;
      top: 0;
      display: block;
    }

    .grid-v,
    .grid-h {
      position: absolute;
      pointer-events: none;
      background: rgb(255 255 255 / 35%);
    }

    .grid-v {
      top: 0;
      bottom: 0;
      width: 1px;
    }

    .grid-h {
      left: 0;
      right: 0;
      height: 1px;
    }
  }

  .crop-controls {
    display: flex;
    gap: 10px;
    align-items: center;
    width: 100%;
    padding: 0 8px;

    .crop-slider {
      flex: 1;
    }

    .crop-zoom-val {
      width: 42px;
      font-size: 12px;
      color: var(--text-3, #8b949e);
      text-align: right;
    }
  }

  .crop-hint {
    font-size: 12px;
    color: var(--text-3, #8b949e);
  }
</style>
