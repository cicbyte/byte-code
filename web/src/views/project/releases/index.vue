<template>
  <div>
    <n-card :bordered="false" class="proCard">
      <n-space size="small" align="center" justify="space-between" class="mb-3">
        <n-space size="small" align="center">
          <!-- 渠道筛选：分段 pill（当前态可见，优于下拉框） -->
          <div class="rel-filter" data-test-id="releases.channel-filter">
            <button
              v-for="opt in filterTabs"
              :key="opt.value"
              class="rel-filter-chip"
              :class="{ active: channelFilter === opt.value }"
              :data-test-id="`releases.channel-${opt.value || 'all'}`"
              @click="setChannel(opt.value)"
            >
              {{ opt.label }}
            </button>
          </div>
          <span class="text-xs text-gray-400">本地打包上传，团队内直接下载，免IM传文件</span>
        </n-space>
        <n-button type="primary" size="small" @click="openCreate" data-test-id="releases.create-btn">
          <template #icon>
            <n-icon><PlusOutlined /></n-icon>
          </template>
          新建发布
        </n-button>
      </n-space>

      <n-spin :show="loading">
        <EmptyState
          type="data"
          title="暂无发布"
          description="本地构建后新建发布并上传安装包，团队成员即可在此下载"
          v-if="!loading && list.length === 0"
        />
        <!-- GitHub Releases 风格：左侧版本时间轴导航 + 右侧一版一卡 -->
        <div v-else class="rel-body">
          <aside class="rel-timeline" data-test-id="releases.timeline">
            <div class="rel-tl-head">版本 · {{ list.length }}</div>
            <div class="rel-tl-list">
              <div
                v-for="r in list"
                :key="r.id"
                class="rel-tl-item"
                :class="{ active: r.id === activeId }"
                :data-test-id="`releases.timeline-item-${r.id}`"
                @click="gotoRelease(r)"
              >
                <span class="rel-tl-dot" :class="`ch-${r.channel}`"></span>
                <div class="rel-tl-main">
                  <div class="rel-tl-ver-row">
                    <span class="rel-tl-version">{{ r.version }}</span>
                    <span v-if="isLatest(r)" class="rel-tl-latest">Latest</span>
                  </div>
                  <span class="rel-tl-time">{{ relTime(r.createdAt) }}</span>
                </div>
              </div>
            </div>
          </aside>
          <div class="rel-cards">
            <div
              v-for="r in list"
              :key="r.id"
              class="rel-card"
              :class="{ flash: r.id === flashId }"
              :ref="(el) => setCardRef(r.id, el)"
              :data-test-id="`releases.row-${r.id}`"
            >
            <div class="rel-head">
              <div class="rel-head-main">
                <n-space size="small" align="center" :wrap="false">
                  <span class="rel-version">{{ r.version }}</span>
                  <n-tag size="small" :bordered="false" :type="channelTagType(r.channel)">
                    {{ channelLabel(r.channel) }}
                  </n-tag>
                  <n-tag v-if="isLatest(r)" size="small" type="success" data-test-id="releases.latest-badge">
                    Latest
                  </n-tag>
                </n-space>
                <div class="rel-title-row">
                  <span class="rel-title">{{ r.title }}</span>
                  <span class="rel-meta">
                    {{ r.createdByName || '-' }} · {{ relTime(r.createdAt) }}（{{ r.createdAt }}）
                  </span>
                </div>
              </div>
              <n-space size="small" :wrap="false">
                <n-button text type="primary" size="small" @click="openFiles(r)" :data-test-id="`releases.files-btn-${r.id}`">
                  文件
                </n-button>
                <n-button text type="info" size="small" @click="openEdit(r)" :data-test-id="`releases.edit-btn-${r.id}`">
                  编辑
                </n-button>
                <n-button text type="error" size="small" @click="handleDelete(r)" :data-test-id="`releases.del-btn-${r.id}`">
                  删除
                </n-button>
              </n-space>
            </div>

            <!-- 发布说明：markdown 渲染，超长折叠 + 展开/收起（GitHub 同款交互） -->
            <div v-if="r.notes" class="rel-notes-wrap">
              <div
                class="rel-notes"
                :class="{ clamp: !expandedNotes.has(r.id) }"
                :ref="(el) => setNotesRef(r.id, el)"
              >
                <MdPreview :id="'md-rel-' + r.id" :model-value="r.notes" :sanitize="safeHtml" />
              </div>
              <div v-if="overflowNotes.has(r.id)" class="rel-notes-toggle" @click="toggleNotes(r.id)">
                {{ expandedNotes.has(r.id) ? '收起' : '展开完整说明' }}
              </div>
            </div>

            <!-- 资产区块：Assets N · 总大小 + 文件行 -->
            <div class="rel-assets">
              <div class="rel-assets-head">
                <span class="text-xs font-medium">Assets {{ r.fileCount }}</span>
                <span class="text-xs text-gray-400" v-if="r.fileCount">{{ fmtSize(r.totalSizeBytes) }}</span>
                <n-upload
                  :show-file-list="false"
                  :custom-request="(o: any) => uploadFile(r.id, o)"
                  multiple
                >
                  <n-button text type="info" size="tiny" :data-test-id="`releases.upload-${r.id}`">
                    + 上传
                  </n-button>
                </n-upload>
              </div>
              <div v-if="filesOf(r.id).length" class="rel-asset-rows">
                <div
                  v-for="f in filesOf(r.id)"
                  :key="f.id"
                  class="rel-asset-row"
                  :data-test-id="`releases.file-${f.id}`"
                  :title="`${f.fileName}（${fmtSize(f.fileSize)}）点击下载`"
                  @click="downloadFile(f)"
                >
                  <n-icon size="15" class="rel-asset-icon">
                    <FileZipOutlined v-if="isArchive(f.fileName)" />
                    <FileOutlined v-else />
                  </n-icon>
                  <span class="rel-asset-name">{{ f.fileName }}</span>
                  <n-tag v-if="f.shared" size="tiny" :bordered="false" type="success">已分享</n-tag>
                  <span class="rel-asset-count text-xs text-gray-400">{{ f.downloadCount }} 次下载</span>
                  <span class="rel-asset-size">{{ fmtSize(f.fileSize) }}</span>
                  <!-- 行内快捷操作（hover 显现，GitHub 同款）：免进「文件」弹窗即可直链/删除 -->
                  <span class="rel-asset-ops" @click.stop>
                    <n-button
                      text
                      type="info"
                      size="tiny"
                      @click="doShare(f, 0)"
                      :data-test-id="`releases.row-share-${f.id}`"
                    >
                      直链
                    </n-button>
                    <n-button
                      text
                      type="error"
                      size="tiny"
                      @click="doDeleteFile(f)"
                      :data-test-id="`releases.row-del-${f.id}`"
                    >
                      删除
                    </n-button>
                  </span>
                </div>
              </div>
              <div v-else class="text-xs text-gray-400" style="padding: 4px 0 2px">未上传文件</div>
            </div>
          </div>
          </div>
        </div>
      </n-spin>

      <div class="mt-4 flex justify-end" v-if="total > pageSize">
        <n-pagination v-model:page="page" :page-size="pageSize" :item-count="total" @update:page="loadData" />
      </div>
    </n-card>

    <!-- 新建/编辑发布 -->
    <n-modal
      v-model:show="showModal"
      preset="dialog"
      :title="isEdit ? '编辑发布' : '新建发布'"
      positive-text="确定"
      negative-text="取消"
      @positive-click="submit"
      style="width: 620px"
    >
      <n-form :model="form" label-placement="left" :label-width="72" class="py-4" :rules="formRules" ref="formRef">
        <n-form-item label="版本号" path="version" v-if="!isEdit">
          <n-input v-model:value="form.version" placeholder="如 v1.2.0（项目内唯一）" data-test-id="releases.version-input" />
        </n-form-item>
        <n-form-item label="标题">
          <n-input v-model:value="form.title" placeholder="缺省同版本号" />
        </n-form-item>
        <n-form-item label="渠道">
          <n-select v-model:value="form.channel" :options="channelOptions" />
        </n-form-item>
        <n-form-item label="发布说明">
          <n-input v-model:value="form.notes" type="textarea" placeholder="changelog / 安装说明（支持 Markdown，列表页渲染展示）" :rows="8" />
        </n-form-item>
      </n-form>
    </n-modal>

    <!-- 文件管理/分享弹窗 -->
    <n-modal
      v-model:show="showShare"
      preset="card"
      :title="`文件 · ${shareRelease?.version ?? ''}`"
      style="width: 680px"
      data-test-id="releases.files-modal"
    >
      <n-table :bordered="false" :single-line="false" size="small">
        <thead>
          <tr>
            <th>文件</th>
            <th>大小</th>
            <th>下载次数</th>
            <th>分享</th>
            <th>操作</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="f in shareFiles" :key="f.id" :data-test-id="`releases.fm-row-${f.id}`">
            <td style="word-break: break-all">{{ f.fileName }}</td>
            <td>{{ fmtSize(f.fileSize) }}</td>
            <td>{{ f.downloadCount }}</td>
            <td>
              <n-tag v-if="f.shared" size="small" type="success" :bordered="false">
                已分享{{ f.shareExpiresAt ? `（至 ${f.shareExpiresAt}）` : '（永久）' }}
              </n-tag>
              <span v-else class="text-gray-400 text-xs">未分享</span>
            </td>
            <td>
              <n-space size="small">
                <n-button text type="info" size="small" @click="downloadFile(f)">下载</n-button>
                <n-button text type="primary" size="small" @click="doShare(f, 0)" :data-test-id="`releases.share-btn-${f.id}`">
                  {{ f.shared ? '直链' : '生成直链' }}
                </n-button>
                <n-button text type="warning" size="small" @click="doShare(f, 7)" :title="f.shared ? '重新生成（7 天有效）' : '生成 7 天有效直链'">
                  7天链
                </n-button>
                <n-button v-if="f.shared" text type="error" size="small" @click="doRevoke(f)">吊销</n-button>
                <n-button text type="error" size="small" @click="doDeleteFile(f)">删除</n-button>
              </n-space>
            </td>
          </tr>
          <tr v-if="shareFiles.length === 0">
            <td colspan="5" class="text-gray-400" style="text-align: center; padding: 16px">该发布还没有文件</td>
          </tr>
        </tbody>
      </n-table>
      <div class="mt-3">
        <n-upload :show-file-list="false" :custom-request="(o: any) => uploadFile(shareRelease?.id ?? 0, o)" multiple>
          <n-button size="small" dashed block>上传文件（单文件上限 2GB）</n-button>
        </n-upload>
      </div>
    </n-modal>

    <!-- 直链展示弹窗 -->
    <n-modal
      v-model:show="showLink"
      preset="dialog"
      title="分享直链"
      positive-text="复制链接"
      negative-text="关闭"
      @positive-click="copyLink"
      style="width: 640px"
      data-test-id="releases.link-modal"
    >
      <div class="py-2">
        <n-input :value="shareUrl" readonly data-test-id="releases.link-input" />
        <div class="text-xs text-gray-400 mt-2">
          {{ linkExpires ? `有效期至 ${linkExpires}` : '永久有效（在文件管理里可吊销）' }}；拿到链接的人无需登录平台即可下载。
        </div>
      </div>
    </n-modal>
  </div>
</template>

<script lang="ts" setup>
  import EmptyState from '@/components/EmptyState/EmptyState.vue';
  import { ref, reactive, computed, onMounted, onBeforeUnmount, nextTick } from 'vue';
  import { useRoute, useRouter } from 'vue-router';
  import { useMessage, useDialog } from 'naive-ui';
  import { PlusOutlined, FileOutlined, FileZipOutlined } from '@vicons/antd';
  import { MdPreview } from 'md-editor-v3';
  import 'md-editor-v3/lib/style.css';
  import DOMPurify from 'dompurify';
  import {
    getReleases,
    createRelease,
    updateRelease,
    deleteRelease,
    getReleaseFiles,
    uploadReleaseFile,
    deleteReleaseFile,
    shareReleaseFile,
    revokeReleaseFileShare,
  } from '@/api/project/release';
  import type { ReleaseItem, ReleaseFileItem } from '@/api/project/release';
  import { copyToClipboard } from '@/utils/clipboard';

  const message = useMessage();
  const dialog = useDialog();
  const route = useRoute();
  const router = useRouter();
  const projectId = computed(() => Number(route.params.projectId));

  const channelOptions = [
    { label: '正式（stable）', value: 'stable' },
    { label: '内测（beta）', value: 'beta' },
    { label: '每日（nightly）', value: 'nightly' },
  ];
  // 渠道筛选分段 pill：'' 即全部（不筛），去掉「全部渠道」选项的语义重复
  const filterTabs = [
    { label: '全部', value: '' },
    ...channelOptions.map((c) => ({ label: c.label.split('（')[0], value: c.value })),
  ];
  function setChannel(v: string) {
    if (channelFilter.value === v) return;
    channelFilter.value = v;
    reload();
  }
  function channelLabel(c: string) {
    return channelOptions.find((o) => o.value === c)?.label || c;
  }
  function channelTagType(c: string): 'success' | 'warning' | 'info' {
    if (c === 'stable') return 'success';
    if (c === 'beta') return 'warning';
    return 'info';
  }
  function fmtSize(bytes: number): string {
    if (!bytes || bytes <= 0) return '-';
    if (bytes < 1024) return `${bytes}B`;
    if (bytes < 1024 * 1024) return `${(bytes / 1024).toFixed(1)}KB`;
    if (bytes < 1024 * 1024 * 1024) return `${(bytes / 1024 / 1024).toFixed(1)}MB`;
    return `${(bytes / 1024 / 1024 / 1024).toFixed(2)}GB`;
  }
  function isArchive(name: string): boolean {
    return /\.(zip|gz|tar|rar|7z|bz2|xz|tgz)$/i.test(name);
  }
  // 相对时间（GitHub 同款语感）：3 分钟前 / 5 小时前 / 2 天前，超 30 天回绝对时间
  function relTime(dt: string): string {
    if (!dt) return '';
    const t = new Date(dt.replace(' ', 'T')).getTime();
    if (Number.isNaN(t)) return dt;
    const diff = Date.now() - t;
    if (diff < 60_000) return '刚刚';
    if (diff < 3_600_000) return `${Math.floor(diff / 60_000)} 分钟前`;
    if (diff < 86_400_000) return `${Math.floor(diff / 3_600_000)} 小时前`;
    if (diff < 30 * 86_400_000) return `${Math.floor(diff / 86_400_000)} 天前`;
    return dt;
  }
  // Latest 徽标：列表（id 倒序）中最新的 stable 版
  function isLatest(r: ReleaseItem): boolean {
    if (r.channel !== 'stable') return false;
    return list.value.find((x) => x.channel === 'stable')?.id === r.id;
  }
  // 说明是否溢出需要折叠开关：按真实渲染高度判定（字符数猜不准——
  // 标题/代码块等 markdown 密集内容短文也会超 180px），渲染后量 scrollHeight
  const notesEls = new Map<number, HTMLElement>();
  function setNotesRef(id: number, el: unknown) {
    if (el) notesEls.set(id, el as HTMLElement);
    else notesEls.delete(id);
  }
  const overflowNotes = ref(new Set<number>());
  async function measureNotes() {
    await nextTick();
    const s = new Set<number>();
    for (const [id, el] of notesEls) {
      if (el.scrollHeight - el.clientHeight > 2) s.add(id);
    }
    overflowNotes.value = s;
  }
  // markdown 消毒：与 QA 页/任务评论同口径
  function safeHtml(md: string): string {
    return DOMPurify.sanitize(md);
  }
  function absUrl(path: string): string {
    if (/^https?:\/\//.test(path)) return path;
    return window.location.origin + path;
  }

  // ---------- 列表 ----------
  const loading = ref(false);
  const list = ref<ReleaseItem[]>([]);
  const total = ref(0);
  const page = ref(1);
  const pageSize = 20;
  const channelFilter = ref('');
  const expandedNotes = ref(new Set<number>());

  async function loadData() {
    loading.value = true;
    try {
      const res = await getReleases(projectId.value, {
        pageNum: page.value,
        pageSize,
        channel: channelFilter.value || undefined,
      });
      list.value = res?.list || [];
      total.value = res?.total || 0;
      expandedNotes.value = new Set();
      loadFiles();
      measureNotes();
      applyQueryVersion();
    } catch (e) {
      // ignore
    } finally {
      loading.value = false;
    }
  }
  function reload() {
    page.value = 1;
    loadData();
  }
  function toggleNotes(id: number) {
    if (expandedNotes.value.has(id)) expandedNotes.value.delete(id);
    else expandedNotes.value.add(id);
    expandedNotes.value = new Set(expandedNotes.value);
  }

  // ---------- 版本时间轴（GitHub 式导航）----------
  const activeId = ref<number | null>(null);
  const flashId = ref<number | null>(null);
  let flashTimer: ReturnType<typeof setTimeout> | null = null;
  const cardEls = new Map<number, HTMLElement>();
  function setCardRef(id: number, el: unknown) {
    if (el) cardEls.set(id, el as HTMLElement);
    else cardEls.delete(id);
  }
  // 点击时间轴：滚动定位到卡片 + 短暂高亮 + URL 记版本（可分享/刷新直达）
  function gotoRelease(r: ReleaseItem) {
    activeId.value = r.id;
    flashId.value = r.id;
    if (flashTimer) clearTimeout(flashTimer);
    flashTimer = setTimeout(() => (flashId.value = null), 1600);
    scrollCardTo(r.id);
    // markdown/文件异步回填会持续撑高布局，刷新直达时首次滚动会偏：多轮校正
    for (const delay of [400, 1000, 1800]) {
      setTimeout(() => scrollCardTo(r.id), delay);
    }
    router.replace({ query: { ...route.query, tag: r.version } });
  }
  function scrollCardTo(id: number) {
    cardEls.get(id)?.scrollIntoView({ behavior: 'smooth', block: 'start' });
  }
  // 进页/刷新按 ?tag= 直达对应版本
  async function applyQueryVersion() {
    const v = route.query.tag;
    if (typeof v === 'string' && v) {
      const r = list.value.find((x) => x.version === v);
      if (r) {
        await nextTick();
        gotoRelease(r);
        return;
      }
    }
    activeId.value = list.value[0]?.id ?? null;
  }
  // scrollspy：capture 接住内层滚动容器，取最贴近视口顶部的卡片联动高亮
  let spyTick = false;
  function onScrollSpy() {
    if (spyTick) return;
    spyTick = true;
    requestAnimationFrame(() => {
      spyTick = false;
      let cur: number | null = null;
      for (const r of list.value) {
        const el = cardEls.get(r.id);
        if (el && el.getBoundingClientRect().top <= 140) cur = r.id;
        else break;
      }
      if (cur !== null) activeId.value = cur;
    });
  }

  // ---------- 文件 ----------
  const filesMap = ref(new Map<number, ReleaseFileItem[]>());
  function filesOf(rid: number): ReleaseFileItem[] {
    return filesMap.value.get(rid) || [];
  }
  async function loadFiles() {
    if (list.value.length === 0) return;
    const results = await Promise.all(
      list.value.map((r) =>
        getReleaseFiles(r.id)
          .then((res) => ({ rid: r.id, files: res?.list || [] }))
          .catch(() => ({ rid: r.id, files: [] }))
      )
    );
    const m = new Map<number, ReleaseFileItem[]>();
    for (const it of results) {
      m.set(it.rid, it.files);
    }
    filesMap.value = m;
  }
  async function uploadFile(rid: number, o: { file: { file: File } }) {
    if (!rid) return;
    try {
      await uploadReleaseFile(rid, o.file.file);
      message.success('上传成功');
      loadData();
      if (showShare.value && shareRelease.value) openFiles(shareRelease.value);
    } catch (e: any) {
      message.error(e?.message || '上传失败');
    }
  }
  // 下载走公开直链（免登录头问题）：无令牌则先生成
  async function downloadFile(f: ReleaseFileItem) {
    try {
      const res = await shareReleaseFile(f.id, 0);
      if (res?.url) window.open(absUrl(res.url), '_blank');
    } catch (e: any) {
      message.error(e?.message || '获取下载链接失败');
    }
  }

  // ---------- 文件管理/分享弹窗 ----------
  const showShare = ref(false);
  const shareRelease = ref<ReleaseItem | null>(null);
  const shareFiles = ref<ReleaseFileItem[]>([]);
  function openFiles(r: ReleaseItem) {
    shareRelease.value = r;
    getReleaseFiles(r.id)
      .then((res) => (shareFiles.value = res?.list || []))
      .catch(() => (shareFiles.value = []));
    showShare.value = true;
  }
  async function doShare(f: ReleaseFileItem, days: number) {
    try {
      const res = await shareReleaseFile(f.id, days);
      shareUrl.value = absUrl(res.url);
      linkExpires.value = res.expiresAt || '';
      showLink.value = true;
      const nl = await getReleaseFiles(f.releaseId).catch(() => null);
      if (nl) shareFiles.value = nl.list || [];
      loadFiles();
    } catch (e: any) {
      message.error(e?.message || '生成直链失败');
    }
  }
  async function doRevoke(f: ReleaseFileItem) {
    try {
      await revokeReleaseFileShare(f.id);
      message.success('已吊销，旧链接即刻失效');
      const nl = await getReleaseFiles(f.releaseId).catch(() => null);
      if (nl) shareFiles.value = nl.list || [];
      loadFiles();
    } catch (e: any) {
      message.error(e?.message || '吊销失败');
    }
  }
  function doDeleteFile(f: ReleaseFileItem) {
    dialog.warning({
      title: '删除文件',
      content: `确定删除 ${f.fileName}？已分享的直链将同时失效，不可恢复。`,
      positiveText: '删除',
      negativeText: '取消',
      onPositiveClick: async () => {
        try {
          await deleteReleaseFile(f.id);
          message.success('已删除');
          const nl = await getReleaseFiles(f.releaseId).catch(() => null);
          if (nl) shareFiles.value = nl.list || [];
          loadData();
        } catch (e: any) {
          message.error(e?.message || '删除失败');
        }
      },
    });
  }

  // ---------- 直链弹窗 ----------
  const showLink = ref(false);
  const shareUrl = ref('');
  const linkExpires = ref('');
  async function copyLink() {
    const ok = await copyToClipboard(shareUrl.value);
    ok ? message.success('已复制') : message.error('复制失败，请手动选择');
  }

  // ---------- 新建/编辑 ----------
  const showModal = ref(false);
  const isEdit = ref(false);
  const editId = ref<number | null>(null);
  const formRef = ref<any>(null);
  const form = reactive({ version: '', title: '', notes: '', channel: 'stable' });
  const formRules = {
    version: { required: true, message: '请输入版本号', trigger: 'blur' },
  };

  function openCreate() {
    isEdit.value = false;
    editId.value = null;
    form.version = '';
    form.title = '';
    form.notes = '';
    form.channel = 'stable';
    showModal.value = true;
  }
  function openEdit(r: ReleaseItem) {
    isEdit.value = true;
    editId.value = r.id;
    form.title = r.title;
    form.notes = r.notes;
    form.channel = r.channel;
    showModal.value = true;
  }
  async function submit() {
    if (!isEdit.value) {
      try {
        await formRef.value?.validate();
      } catch {
        return false;
      }
    }
    try {
      if (isEdit.value && editId.value) {
        await updateRelease(editId.value, {
          title: form.title,
          notes: form.notes,
          channel: form.channel,
        });
        message.success('更新成功');
      } else {
        await createRelease(projectId.value, { ...form });
        message.success('发布已创建，可上传安装包');
      }
      showModal.value = false;
      loadData();
    } catch (e: any) {
      message.error(e?.message || '操作失败');
      return false;
    }
  }

  function handleDelete(r: ReleaseItem) {
    dialog.warning({
      title: '删除发布',
      content: `确定删除发布 ${r.version}（含 ${r.fileCount} 个文件及其直链）？该操作不可恢复。`,
      positiveText: '删除',
      negativeText: '取消',
      onPositiveClick: async () => {
        try {
          await deleteRelease(r.id);
          message.success('已删除');
          loadData();
        } catch (e: any) {
          message.error(e?.message || '删除失败');
        }
      },
    });
  }

  onMounted(() => {
    loadData();
    window.addEventListener('scroll', onScrollSpy, true);
  });
  onBeforeUnmount(() => {
    window.removeEventListener('scroll', onScrollSpy, true);
    if (flashTimer) clearTimeout(flashTimer);
  });
</script>

<style scoped>
  /* 渠道筛选分段 pill：容器胶囊描边，active 实心主题色 */
  .rel-filter {
    display: inline-flex;
    align-items: center;
    gap: 2px;
    padding: 2px;
    border: 1px solid var(--line, #e9e9e7);
    border-radius: 999px;
    background: var(--panel-bg, #fff);
  }
  .rel-filter-chip {
    border: 0;
    background: transparent;
    padding: 3px 12px;
    border-radius: 999px;
    font-size: 12px;
    line-height: 18px;
    color: var(--text-2, #57606a);
    cursor: pointer;
    transition:
      background 0.15s,
      color 0.15s;
  }
  .rel-filter-chip:hover {
    background: var(--hover-bg, rgba(0, 0, 0, 0.035));
  }
  .rel-filter-chip.active {
    background: var(--primary-color, #16a34a);
    color: #fff;
  }
  .rel-body {
    display: flex;
    align-items: flex-start;
    gap: 18px;
  }
  .rel-cards {
    flex: 1;
    min-width: 0;
    display: flex;
    flex-direction: column;
    gap: 14px;
  }
  /* 左侧版本时间轴（GitHub tags 导航同款视觉：竖线 + 渠道色节点） */
  .rel-timeline {
    position: sticky;
    top: 12px;
    width: 176px;
    flex-shrink: 0;
    max-height: calc(100vh - 120px);
    overflow-y: auto;
    padding: 2px 0;
  }
  .rel-tl-head {
    font-size: 12px;
    font-weight: 600;
    color: var(--text-2, #57606a);
    padding: 2px 0 8px 20px;
  }
  .rel-tl-list {
    position: relative;
  }
  .rel-tl-list::before {
    content: '';
    position: absolute;
    left: 5px;
    top: 8px;
    bottom: 8px;
    width: 2px;
    border-radius: 1px;
    background: var(--line, #e9e9e7);
  }
  .rel-tl-item {
    position: relative;
    display: flex;
    align-items: flex-start;
    padding: 5px 8px 5px 20px;
    cursor: pointer;
    border-radius: 6px;
  }
  .rel-tl-item:hover {
    background: var(--hover-bg, rgba(0, 0, 0, 0.035));
  }
  .rel-tl-item.active {
    background: var(--hover-bg, rgba(0, 0, 0, 0.035));
  }
  .rel-tl-dot {
    position: absolute;
    left: 0;
    top: 9px;
    width: 12px;
    height: 12px;
    border-radius: 50%;
    border: 2px solid var(--panel-bg, #fff);
    background: var(--text-3, #8b949e);
    box-sizing: border-box;
  }
  .rel-tl-dot.ch-stable {
    background: #18a058;
  }
  .rel-tl-dot.ch-beta {
    background: #f0a020;
  }
  .rel-tl-dot.ch-nightly {
    background: #2080f0;
  }
  .rel-tl-item.active .rel-tl-dot {
    transform: scale(1.2);
  }
  .rel-tl-main {
    min-width: 0;
  }
  .rel-tl-ver-row {
    display: flex;
    align-items: center;
    gap: 6px;
  }
  .rel-tl-version {
    font-size: 12px;
    color: var(--text-2, #57606a);
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }
  .rel-tl-item.active .rel-tl-version {
    color: var(--primary-color, #16a34a);
    font-weight: 600;
  }
  .rel-tl-latest {
    flex-shrink: 0;
    font-size: 10px;
    line-height: 16px;
    color: #18a058;
    border: 1px solid currentColor;
    border-radius: 10px;
    padding: 0 5px;
  }
  .rel-tl-time {
    font-size: 11px;
    color: var(--text-3, #8b949e);
    white-space: nowrap;
  }
  .rel-card {
    border: 1px solid var(--line, #e9e9e7);
    border-radius: 8px;
    padding: 14px 18px;
    background: var(--panel-bg, #fff);
    scroll-margin-top: 16px;
    transition:
      border-color 0.3s,
      box-shadow 0.3s;
  }
  /* 时间轴点击直达后的短暂高亮 */
  .rel-card.flash {
    border-color: var(--primary-color, #16a34a);
    box-shadow: 0 0 0 3px rgba(24, 160, 88, 0.12);
  }
  @media (max-width: 1080px) {
    .rel-timeline {
      display: none;
    }
  }
  .rel-head {
    display: flex;
    align-items: flex-start;
    justify-content: space-between;
    gap: 12px;
  }
  .rel-head-main {
    min-width: 0;
    flex: 1;
  }
  .rel-version {
    font-size: 16px;
    font-weight: 600;
    color: var(--text-1, #24292f);
  }
  .rel-title-row {
    margin-top: 2px;
    display: flex;
    align-items: baseline;
    gap: 10px;
    flex-wrap: wrap;
  }
  .rel-title {
    font-size: 13px;
    color: var(--text-2, #57606a);
  }
  .rel-meta {
    font-size: 12px;
    color: var(--text-3, #8b949e);
  }
  .rel-notes-wrap {
    margin-top: 8px;
  }
  .rel-notes {
    position: relative;
  }
  .rel-notes.clamp {
    max-height: 180px;
    overflow: hidden;
    /* 底部渐隐提示可展开（GitHub 同款视觉） */
    mask-image: linear-gradient(to bottom, #000 70%, transparent 100%);
    -webkit-mask-image: linear-gradient(to bottom, #000 70%, transparent 100%);
  }
  .rel-notes-toggle {
    margin-top: 4px;
    font-size: 12px;
    color: var(--primary-color, #16a34a);
    cursor: pointer;
    user-select: none;
  }
  .rel-notes-toggle:hover {
    text-decoration: underline;
  }
  .rel-assets {
    margin-top: 10px;
    border-top: 1px dashed var(--line, #e9e9e7);
    padding-top: 8px;
  }
  .rel-assets-head {
    display: flex;
    align-items: center;
    gap: 10px;
  }
  /* n-upload 根节点参与 flex 收缩会把手头的文字 span 压到 min-content
     （Assets N 竖排、3.0KB 断词）：全部禁止收缩，上传按钮推到行尾 */
  .rel-assets-head > span {
    white-space: nowrap;
    flex-shrink: 0;
  }
  .rel-assets-head > :deep(.n-upload) {
    flex-shrink: 0;
    margin-left: auto;
    width: fit-content; /* naive 内部 100% 宽残留会反向溢出容器，收敛到按钮实际宽度 */
  }
  .rel-asset-rows {
    margin-top: 4px;
    display: flex;
    flex-direction: column;
  }
  .rel-asset-row {
    display: flex;
    align-items: center;
    gap: 8px;
    padding: 5px 8px;
    border-radius: 6px;
    cursor: pointer;
  }
  .rel-asset-row:hover {
    background: var(--hover-bg, rgba(0, 0, 0, 0.035));
  }
  .rel-asset-icon {
    color: var(--text-3, #8b949e);
    flex-shrink: 0;
  }
  .rel-asset-name {
    font-size: 13px;
    color: var(--primary-color, #16a34a);
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }
  .rel-asset-row:hover .rel-asset-name {
    text-decoration: underline;
  }
  .rel-asset-size {
    margin-left: auto;
    flex-shrink: 0;
    font-size: 12px;
    color: var(--text-3, #8b949e);
  }
  .rel-asset-count {
    flex-shrink: 0;
    min-width: 56px;
    text-align: right;
  }
  /* 行内快捷操作：hover 才显现，免进文件管理弹窗 */
  .rel-asset-ops {
    display: none;
    flex-shrink: 0;
    align-items: center;
    gap: 2px;
    margin-left: 6px;
  }
  .rel-asset-row:hover .rel-asset-ops {
    display: flex;
  }
</style>
