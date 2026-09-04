import { ref, computed } from 'vue';
import type { ComputedRef, Ref } from 'vue';
import DOMPurify from 'dompurify';
import { docsRawUrl } from '@/api/docs/index';
import type { DocsFile } from '@/api/docs/index';

// 二进制预览：docx（mammoth 转 HTML + 消毒）与图片直出。
// raw 直链接口流式直出且鉴权走 token 头——img/fetch/a 标签不会带自定义头，
// 统一改为带 token fetch → blob URL

const IMAGE_EXTS = ['.png', '.jpg', '.jpeg', '.gif', '.webp', '.svg', '.bmp'];

export interface BinaryPreviewDeps {
  projectId: ComputedRef<number>;
  currentFile: Ref<DocsFile | null>;
}

export function useBinaryPreview(deps: BinaryPreviewDeps) {
  const docxHtml = ref('');
  const docxLoading = ref(false);
  const binObjectUrl = ref('');

  async function loadBinObjectUrl() {
    if (binObjectUrl.value) URL.revokeObjectURL(binObjectUrl.value);
    binObjectUrl.value = '';
    try {
      const token = JSON.parse(localStorage.getItem('ACCESS-TOKEN') || '{"value":""}').value || '';
      const resp = await fetch(docsRawUrl(deps.projectId.value, deps.currentFile.value!.path), {
        headers: { token },
      });
      if (!resp.ok) throw new Error(String(resp.status));
      const blob = await resp.blob();
      binObjectUrl.value = URL.createObjectURL(blob);
    } catch {
      binObjectUrl.value = '';
    }
  }

  function downloadBin() {
    if (!binObjectUrl.value) return;
    const a = document.createElement('a');
    a.href = binObjectUrl.value;
    a.download = (deps.currentFile.value?.path || 'file').split('/').pop() || 'file';
    a.click();
  }

  const fileExt = computed(() => {
    const p = deps.currentFile.value?.path || '';
    const i = p.lastIndexOf('.');
    return i >= 0 ? p.slice(i).toLowerCase() : '';
  });
  const isDocx = computed(() => fileExt.value === '.docx');
  const isImage = computed(() => IMAGE_EXTS.includes(fileExt.value));

  async function loadDocxPreview() {
    docxHtml.value = '';
    docxLoading.value = true;
    try {
      const buf = await (await fetch(binObjectUrl.value)).arrayBuffer();
      // mammoth 较大（~200KB），动态导入按需加载
      const mammoth = await import('mammoth/mammoth.browser');
      const result = await (mammoth as any).convertToHtml({ arrayBuffer: buf });
      // mammoth 官方要求自行消毒：docx 超链接可携带 javascript: scheme 等注入面
      docxHtml.value = DOMPurify.sanitize(result.value || '<p>（空文档）</p>', {
        ALLOWED_TAGS: DOMPurify.allowedTags, // 默认白名单（禁 script/iframe 等）
        ALLOWED_ATTR: ['href', 'src', 'alt', 'colspan', 'rowspan'], // 最小属性集
        ALLOW_DATA_ATTR: false,
      });
    } catch {
      docxHtml.value = '';
    } finally {
      docxLoading.value = false;
    }
  }

  function onUnmountCleanup() {
    if (binObjectUrl.value) URL.revokeObjectURL(binObjectUrl.value);
  }

  return {
    docxHtml, docxLoading, binObjectUrl,
    isDocx, isImage,
    loadBinObjectUrl, loadDocxPreview, downloadBin, onUnmountCleanup,
  };
}
