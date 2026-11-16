'use strict';

/* ==========================================================================
   Data
   ========================================================================== */

const ALGORITHMS = [
  { name: 'MD5', bits: 128 },
  { name: 'SHA1', bits: 160 },
  { name: 'SHA224', bits: 224 },
  { name: 'SHA256', bits: 256, recommended: true },
  { name: 'SHA384', bits: 384 },
  { name: 'SHA512', bits: 512 },
  { name: 'SHA512_224', bits: 224 },
  { name: 'SHA512_256', bits: 256 },
  { name: 'SHA3_224', bits: 224 },
  { name: 'SHA3_256', bits: 256 },
  { name: 'SHA3_384', bits: 384 },
  { name: 'SHA3_512', bits: 512 },
  { name: 'BLAKE2b', bits: 512 },
  { name: 'BLAKE2s', bits: 256 },
  { name: 'SM3', bits: 256 },
  { name: 'CRC32', bits: 32 },
];

const DEFAULT_SELECTED = ['SHA256'];

const I18N = {
  zh: {
    appName: '文件完整性校验',
    tagline: 'PRISMATIC · 本地哈希工作台',
    filesUnit: '个文件',
    themeToLight: '切换到浅色',
    themeToDark: '切换到深色',
    langSwitch: '切换语言',
    about: '关于',
    dropTitle: '拖拽文件到此处',
    dropHint: '支持多选 · 全部计算在本地完成',
    browse: '浏览文件',
    dropRelease: '松开以添加文件',
    filesTitle: '文件',
    clearAll: '全部清空',
    filesEmpty: '尚未添加文件',
    filesEmptyHint: '拖拽或浏览文件，即刻开始校验',
    removeFile: '移除文件',
    algosTitle: '算法选择',
    selectAll: '全选',
    resetAlgos: '重置',
    recommended: '推荐',
    resultsTitle: '校验结果',
    copyAll: '复制全部',
    export: '导出',
    resultsEmpty: '等待计算',
    resultsEmptyHint: '校验结果将在此处逐条呈现',
    calculate: '开始计算',
    calculating: '计算中',
    statusReady: '就绪',
    statusCalculating: '正在计算',
    statusDone: '校验完成',
    statusError: '部分文件失败',
    ready: '就绪',
    busy: '计算中',
    done: '完成',
    error: '错误',
    copy: '复制',
    copied: '已复制',
    copyFailed: '复制失败',
    noFiles: '请先添加文件',
    noAlgos: '请至少选择一种算法',
    exportEmpty: '暂无可导出的结果',
    exportSuccess: '已导出结果',
    exportFailed: '导出失败',
    textFiles: '文本文件',
    calcDone: '全部文件校验完成',
    calcPartial: '部分文件校验失败，请查看状态',
    skipDuplicate: '已跳过重复文件',
    busyAddBlocked: '正在计算，请稍后再添加文件',
    resultMeta: '{size} · {n} 项',
    aboutTitle: '关于 文件完整性校验',
    aboutBody:
      '一个轻盈、通透的本地文件完整性校验工具。支持 16 种哈希算法，所有计算均在你的设备上完成，文件不会离开本机。',
    chipLocal: '本地计算',
    close: '关闭',
    glassFeel: '玻璃质感',
  },
  en: {
    appName: 'File Integrity Checker',
    tagline: 'PRISMATIC · LOCAL HASH WORKBENCH',
    filesUnit: 'files',
    themeToLight: 'Switch to light',
    themeToDark: 'Switch to dark',
    langSwitch: 'Switch language',
    about: 'About',
    dropTitle: 'Drop files here',
    dropHint: 'Multi-select · everything is computed locally',
    browse: 'Browse files',
    dropRelease: 'Release to add files',
    filesTitle: 'Files',
    clearAll: 'Clear all',
    filesEmpty: 'No files yet',
    filesEmptyHint: 'Drop or browse files to get started',
    removeFile: 'Remove file',
    algosTitle: 'Algorithms',
    selectAll: 'Select all',
    resetAlgos: 'Reset',
    recommended: 'Recommended',
    resultsTitle: 'Results',
    copyAll: 'Copy all',
    export: 'Export',
    resultsEmpty: 'No results yet',
    resultsEmptyHint: 'Hashes will appear here as they complete',
    calculate: 'Start hashing',
    calculating: 'Hashing',
    statusReady: 'Ready',
    statusCalculating: 'Hashing files',
    statusDone: 'Finished',
    statusError: 'Completed with errors',
    ready: 'Ready',
    busy: 'Hashing',
    done: 'Done',
    error: 'Error',
    copy: 'Copy',
    copied: 'Copied',
    copyFailed: 'Copy failed',
    noFiles: 'Add some files first',
    noAlgos: 'Select at least one algorithm',
    exportEmpty: 'Nothing to export yet',
    exportSuccess: 'Results exported',
    exportFailed: 'Export failed',
    textFiles: 'Text file',
    calcDone: 'All files hashed',
    calcPartial: 'Some files failed — check their status',
    skipDuplicate: 'Duplicate files skipped',
    busyAddBlocked: 'Hashing in progress — try again shortly',
    resultMeta: '{size} · {n} hashes',
    aboutTitle: 'About File Integrity Checker',
    aboutBody:
      'A light, translucent file integrity workbench. 16 hashing algorithms, all computed on your device — files never leave this machine.',
    chipLocal: 'Local only',
    close: 'Close',
    glassFeel: 'Glass intensity',
  },
};

/* ==========================================================================
   State
   ========================================================================== */

const store = {
  get(key, fallback) {
    try {
      const value = localStorage.getItem(key);
      return value === null ? fallback : value;
    } catch {
      return fallback;
    }
  },
  set(key, value) {
    try {
      localStorage.setItem(key, String(value));
    } catch {
      /* preferences are best-effort */
    }
  },
};

const state = {
  files: [],
  selected: new Set(DEFAULT_SELECTED),
  lang: store.get('fc.lang', 'zh') === 'en' ? 'en' : 'zh',
  theme: store.get('fc.theme', 'light') === 'dark' ? 'dark' : 'light',
  glass: 0.25,
  phase: 'idle',
  busy: false,
  total: 0,
  processed: 0,
  errors: 0,
};

/* ==========================================================================
   DOM
   ========================================================================== */

const $ = (id) => document.getElementById(id);

const els = {
  metaCount: $('metaCount'),
  themeBtn: $('themeBtn'),
  langBtn: $('langBtn'),
  aboutBtn: $('aboutBtn'),
  dropzone: $('dropzone'),
  browseBtn: $('browseBtn'),
  fileInput: $('fileInput'),
  fileBadge: $('fileBadge'),
  clearBtn: $('clearBtn'),
  fileList: $('fileList'),
  filesEmpty: $('filesEmpty'),
  algoBadge: $('algoBadge'),
  selectAllBtn: $('selectAllBtn'),
  resetAlgosBtn: $('resetAlgosBtn'),
  algoGrid: $('algoGrid'),
  resultBadge: $('resultBadge'),
  copyAllBtn: $('copyAllBtn'),
  exportBtn: $('exportBtn'),
  results: $('results'),
  resultsEmpty: $('resultsEmpty'),
  statusDot: $('statusDot'),
  statusText: $('statusText'),
  statusCount: $('statusCount'),
  progressFill: $('progressFill'),
  glassRange: $('glassRange'),
  calcBtn: $('calcBtn'),
  calcLabel: document.querySelector('.primary-btn__label'),
  toastStack: $('toastStack'),
  aboutModal: $('aboutModal'),
  aboutClose: $('aboutClose'),
};

/* ==========================================================================
   Helpers
   ========================================================================== */

function el(tag, className) {
  const node = document.createElement(tag);
  if (className) node.className = className;
  return node;
}

function icon(id, className) {
  const svg = document.createElementNS('http://www.w3.org/2000/svg', 'svg');
  svg.setAttribute('class', `icon${className ? ` ${className}` : ''}`);
  svg.setAttribute('aria-hidden', 'true');
  const use = document.createElementNS('http://www.w3.org/2000/svg', 'use');
  use.setAttribute('href', `#${id}`);
  svg.appendChild(use);
  return svg;
}

function t(key, vars) {
  const dict = I18N[state.lang] || I18N.zh;
  let text = dict[key] ?? I18N.zh[key] ?? key;
  if (vars) {
    for (const [name, value] of Object.entries(vars)) {
      text = text.replaceAll(`{${name}}`, value);
    }
  }
  return text;
}

function formatBytes(bytes) {
  if (!Number.isFinite(bytes) || bytes < 0) return '—';
  if (bytes < 1024) return `${bytes} B`;
  const units = ['KB', 'MB', 'GB', 'TB'];
  let value = bytes / 1024;
  let index = 0;
  while (value >= 1024 && index < units.length - 1) {
    value /= 1024;
    index += 1;
  }
  return `${value >= 100 ? value.toFixed(0) : value.toFixed(1)} ${units[index]}`;
}

function extensionOf(name) {
  const dot = name.lastIndexOf('.');
  if (dot <= 0) return '';
  const ext = name.slice(dot + 1);
  return ext.length > 0 && ext.length <= 8 ? ext.toUpperCase() : '';
}

function toast(kind, message) {
  if (!message) return;
  const node = el('div', `toast${kind === 'error' ? ' toast--error' : ''}`);
  node.append(el('span', 'toast__dot'), Object.assign(el('span'), { textContent: message }));
  els.toastStack.appendChild(node);
  while (els.toastStack.children.length > 3) {
    els.toastStack.firstElementChild.remove();
  }
  window.setTimeout(() => {
    node.classList.add('is-leaving');
    window.setTimeout(() => node.remove(), 240);
  }, 3200);
}

async function copyText(text) {
  try {
    await navigator.clipboard.writeText(text);
    return true;
  } catch {
    try {
      const area = document.createElement('textarea');
      area.value = text;
      area.style.position = 'fixed';
      area.style.opacity = '0';
      document.body.appendChild(area);
      area.select();
      const ok = document.execCommand('copy');
      area.remove();
      return ok;
    } catch {
      return false;
    }
  }
}

/* ==========================================================================
   Preferences
   ========================================================================== */

function applyTheme() {
  document.documentElement.dataset.theme = state.theme;
  const use = els.themeBtn.querySelector('use');
  if (use) use.setAttribute('href', state.theme === 'light' ? '#i-moon' : '#i-sun');
  els.themeBtn.title = t(state.theme === 'light' ? 'themeToDark' : 'themeToLight');
  store.set('fc.theme', state.theme);
}

function applyGlass() {
  document.documentElement.style.setProperty('--glass-alpha', String(state.glass));
  els.glassRange.value = String(state.glass);
  store.set('fc.glass', state.glass.toFixed(2));
}

function applyLang() {
  const html = document.documentElement;
  html.lang = state.lang === 'zh' ? 'zh-CN' : 'en';
  document.title = t('appName');

  document.querySelectorAll('[data-i18n]').forEach((node) => {
    node.textContent = t(node.dataset.i18n);
  });
  document.querySelectorAll('[data-i18n-aria]').forEach((node) => {
    node.setAttribute('aria-label', t(node.dataset.i18nAria));
  });

  els.langBtn.textContent = state.lang === 'zh' ? 'EN' : '中';
  els.langBtn.title = t('langSwitch');
  els.aboutBtn.title = t('about');
  els.aboutClose.title = t('close');
  els.glassRange.setAttribute('aria-label', t('glassFeel'));
  els.calcLabel.textContent = state.busy ? t('calculating') : t('calculate');
  els.themeBtn.title = t(state.theme === 'light' ? 'themeToDark' : 'themeToLight');

  els.algoGrid.querySelectorAll('.algo-chip__rec').forEach((node) => {
    node.title = t('recommended');
  });

  store.set('fc.lang', state.lang);
  renderFiles();
  renderResults();
  refreshStatus();
}

/* ==========================================================================
   Algorithms
   ========================================================================== */

function buildAlgorithms() {
  const fragment = document.createDocumentFragment();
  for (const algo of ALGORITHMS) {
    const chip = el('button', 'algo-chip');
    chip.type = 'button';
    chip.dataset.name = algo.name;
    chip.setAttribute('role', 'checkbox');
    chip.setAttribute('aria-checked', String(state.selected.has(algo.name)));
    chip.classList.toggle('is-on', state.selected.has(algo.name));

    const check = el('span', 'algo-chip__check');
    check.appendChild(icon('i-check'));
    const name = el('span', 'algo-chip__name');
    name.textContent = algo.name;
    const bits = el('span', 'algo-chip__bits');
    bits.textContent = `${algo.bits} bit`;
    chip.append(check, name, bits);

    if (algo.recommended) {
      const rec = el('span', 'algo-chip__rec');
      rec.title = t('recommended');
      chip.appendChild(rec);
    }

    chip.addEventListener('click', () => toggleAlgorithm(algo.name));
    fragment.appendChild(chip);
  }
  els.algoGrid.replaceChildren(fragment);
  updateAlgoBadge();
}

function toggleAlgorithm(name) {
  if (state.busy) return;
  if (state.selected.has(name)) {
    state.selected.delete(name);
  } else {
    state.selected.add(name);
  }
  const chip = els.algoGrid.querySelector(`[data-name="${name}"]`);
  if (chip) {
    chip.classList.toggle('is-on', state.selected.has(name));
    chip.setAttribute('aria-checked', String(state.selected.has(name)));
  }
  updateAlgoBadge();
}

function selectAllAlgorithms() {
  if (state.busy) return;
  state.selected = new Set(ALGORITHMS.map((algo) => algo.name));
  syncAlgorithmChips();
}

function resetAlgorithms() {
  if (state.busy) return;
  state.selected = new Set(DEFAULT_SELECTED);
  syncAlgorithmChips();
}

function syncAlgorithmChips() {
  els.algoGrid.querySelectorAll('.algo-chip').forEach((chip) => {
    const on = state.selected.has(chip.dataset.name);
    chip.classList.toggle('is-on', on);
    chip.setAttribute('aria-checked', String(on));
  });
  updateAlgoBadge();
}

function updateAlgoBadge() {
  els.algoBadge.textContent = `${state.selected.size}/${ALGORITHMS.length}`;
}

/* ==========================================================================
   File list
   ========================================================================== */

function buildFileRow(item) {
  const row = el('div', 'file-row');
  row.dataset.id = item.id;
  row.setAttribute('role', 'listitem');
  row.classList.toggle('is-busy', item.status === 'busy');

  const glyph = el('span', 'file-row__glyph');
  glyph.appendChild(icon('i-file'));

  const main = el('div', 'file-row__main');
  const name = el('span', 'file-row__name');
  name.textContent = item.file.name;
  name.title = item.file.name;
  const ext = extensionOf(item.file.name);
  const meta = el('span', 'file-row__meta');
  meta.textContent = ext ? `${formatBytes(item.file.size)} · ${ext}` : formatBytes(item.file.size);
  main.append(name, meta);

  const chip = buildStatusChip(item.status);
  const remove = el('button', 'row-btn');
  remove.type = 'button';
  remove.title = t('removeFile');
  remove.setAttribute('aria-label', `${t('removeFile')}: ${item.file.name}`);
  remove.appendChild(icon('i-x'));
  remove.disabled = state.busy;
  remove.addEventListener('click', () => removeFile(item.id));

  row.append(glyph, main, chip, remove);
  return row;
}

function buildStatusChip(status) {
  const chip = el('span', 'status-chip');
  chip.dataset.status = status;
  chip.append(el('span', 'dot'), Object.assign(el('span'), { textContent: t(status) }));
  return chip;
}

function renderFiles() {
  const count = state.files.length;
  els.fileBadge.textContent = String(count);
  els.metaCount.textContent = String(count);
  els.filesEmpty.hidden = count > 0;
  els.fileList.hidden = count === 0;

  const fragment = document.createDocumentFragment();
  for (const item of state.files) fragment.appendChild(buildFileRow(item));
  els.fileList.replaceChildren(fragment);
  refreshActionStates();
}

function updateFileRow(item) {
  const row = els.fileList.querySelector(`[data-id="${item.id}"]`);
  if (!row) return;
  row.classList.toggle('is-busy', item.status === 'busy');
  const chip = row.querySelector('.status-chip');
  if (chip) chip.replaceWith(buildStatusChip(item.status));
  const remove = row.querySelector('.row-btn');
  if (remove) remove.disabled = state.busy;
}

function addFiles(fileList) {
  if (state.busy) {
    toast('error', t('busyAddBlocked'));
    return;
  }
  const incoming = Array.from(fileList || []);
  if (incoming.length === 0) return;

  let added = 0;
  let duplicates = 0;
  for (const file of incoming) {
    const key = `${file.name}|${file.size}|${file.lastModified}`;
    if (state.files.some((item) => item.key === key)) {
      duplicates += 1;
      continue;
    }
    state.files.push({
      id: `f${Date.now().toString(36)}${Math.random().toString(36).slice(2, 8)}`,
      key,
      file,
      status: 'ready',
      hashes: null,
    });
    added += 1;
  }

  if (added > 0) {
    state.phase = 'idle';
    state.processed = 0;
    state.total = state.files.length;
    state.errors = 0;
    renderFiles();
    renderResults();
    refreshStatus();
  }
  if (duplicates > 0) toast('success', t('skipDuplicate'));
}

function removeFile(id) {
  if (state.busy) return;
  state.files = state.files.filter((item) => item.id !== id);
  state.phase = 'idle';
  state.processed = 0;
  state.errors = 0;
  state.total = state.files.length;
  renderFiles();
  renderResults();
  refreshStatus();
}

function clearAll() {
  if (state.busy) return;
  state.files = [];
  state.phase = 'idle';
  state.processed = 0;
  state.errors = 0;
  state.total = 0;
  renderFiles();
  renderResults();
  refreshStatus();
}

/* ==========================================================================
   Results
   ========================================================================== */

function orderedHashes(item) {
  return ALGORITHMS.map((algo) => algo.name).filter((name) => name in item.hashes);
}

function buildResultBlock(item) {
  const block = el('article', 'result-block');

  const head = el('div', 'result-block__head');
  const name = el('div', 'result-block__name');
  name.textContent = item.file.name;
  name.title = item.file.name;
  const algorithms = orderedHashes(item);
  const meta = el('div', 'result-block__meta');
  meta.textContent = t('resultMeta', {
    size: formatBytes(item.file.size),
    n: String(algorithms.length),
  });
  head.append(name, meta);
  block.appendChild(head);

  for (const algo of algorithms) {
    const value = item.hashes[algo];
    const row = el('div', 'hash-row');

    const label = el('span', 'hash-row__algo');
    label.textContent = algo;
    label.title = algo;

    const code = el('code', 'hash-row__value');
    code.textContent = value;
    code.title = value;

    const copy = el('button', 'copy-btn');
    copy.type = 'button';
    copy.title = t('copy');
    copy.setAttribute('aria-label', `${t('copy')} ${algo}`);
    copy.appendChild(icon('i-copy'));
    copy.addEventListener('click', async () => {
      const ok = await copyText(value);
      if (!ok) {
        toast('error', t('copyFailed'));
        return;
      }
      copy.classList.add('is-copied');
      const use = copy.querySelector('use');
      if (use) use.setAttribute('href', '#i-check');
      window.setTimeout(() => {
        copy.classList.remove('is-copied');
        if (use) use.setAttribute('href', '#i-copy');
      }, 1200);
      toast('success', t('copied'));
    });

    row.append(label, code, copy);
    block.appendChild(row);
  }
  return block;
}

function completedItems() {
  return state.files.filter((item) => item.hashes);
}

function renderResults() {
  const scrollTop = els.results.scrollTop;
  const children = document.createDocumentFragment();
  const completed = completedItems();
  for (const item of completed) children.appendChild(buildResultBlock(item));

  els.results.replaceChildren(children);
  els.resultBadge.textContent = String(completed.length);
  els.resultsEmpty.hidden = completed.length > 0;
  els.results.hidden = completed.length === 0;
  els.results.scrollTop = scrollTop;
  refreshActionStates();
}

function appendResult(item) {
  els.results.appendChild(buildResultBlock(item));
  els.resultsEmpty.hidden = true;
  els.results.hidden = false;
  els.resultBadge.textContent = String(completedItems().length);
  refreshActionStates();
}

function buildExportText() {
  const lines = [];
  for (const item of state.files) {
    if (!item.hashes) continue;
    let line = `${item.file.name}--`;
    for (const algo of orderedHashes(item)) {
      line += ` ${algo}:${item.hashes[algo]}`;
    }
    lines.push(line);
  }
  return lines.length > 0 ? `${lines.join('\n')}\n` : '';
}

/* ==========================================================================
   Actions
   ========================================================================== */

function refreshActionStates() {
  const hasResults = state.files.some((item) => item.hashes);
  els.copyAllBtn.disabled = !hasResults;
  els.exportBtn.disabled = !hasResults;
  els.clearBtn.disabled = state.busy || state.files.length === 0;
  els.browseBtn.disabled = state.busy;
  els.calcBtn.disabled = state.busy;
}

function setBusy(busy) {
  state.busy = busy;
  document.body.classList.toggle('is-calculating', busy);
  els.calcBtn.classList.toggle('is-busy', busy);
  els.calcLabel.textContent = busy ? t('calculating') : t('calculate');
  els.algoGrid.querySelectorAll('.algo-chip').forEach((chip) => {
    chip.disabled = busy;
  });
  els.selectAllBtn.disabled = busy;
  els.resetAlgosBtn.disabled = busy;
  refreshActionStates();
  els.fileList.querySelectorAll('.row-btn').forEach((button) => {
    button.disabled = busy;
  });
}

function refreshStatus() {
  const total = state.total;
  const done = state.processed;
  const percent = total > 0 ? Math.min(100, (done / total) * 100) : 0;
  els.progressFill.style.width = `${percent}%`;
  els.statusCount.textContent = `${done} / ${total}`;
  els.statusDot.className = 'status__dot';

  if (state.busy) {
    els.statusText.textContent = t('statusCalculating');
    els.statusDot.classList.add('is-busy');
    return;
  }
  if (state.phase === 'done') {
    els.statusText.textContent = state.errors > 0 ? t('statusError') : t('statusDone');
    els.statusDot.classList.add(state.errors > 0 ? 'is-error' : 'is-done');
    return;
  }
  els.statusText.textContent = t('statusReady');
}

async function startCalculation() {
  if (state.busy) return;
  if (state.files.length === 0) {
    toast('error', t('noFiles'));
    return;
  }

  const algorithms = ALGORITHMS.filter((algo) => state.selected.has(algo.name)).map(
    (algo) => algo.name,
  );
  if (algorithms.length === 0) {
    toast('error', t('noAlgos'));
    return;
  }

  const snapshot = state.files.slice();
  state.busy = true;
  state.phase = 'busy';
  state.total = snapshot.length;
  state.processed = 0;
  state.errors = 0;

  for (const item of snapshot) {
    item.status = 'ready';
    item.hashes = null;
  }
  setBusy(true);
  renderFiles();
  renderResults();
  refreshStatus();

  try {
    for (const item of snapshot) {
      item.status = 'busy';
      updateFileRow(item);
      refreshStatus();

      try {
        const response = await fetch('/api/hash', {
          method: 'POST',
          headers: { 'X-Algorithms': algorithms.join(',') },
          body: item.file,
        });
        if (!response.ok) {
          let message = `${response.status} ${response.statusText}`.trim();
          try {
            const payload = await response.json();
            if (payload && payload.error) message = payload.error;
          } catch {
            /* keep the status text */
          }
          throw new Error(message);
        }
        const payload = await response.json();
        item.hashes = payload.hashes;
        item.status = 'done';
        appendResult(item);
      } catch (error) {
        item.status = 'error';
        state.errors += 1;
        toast('error', `${item.file.name}: ${error.message || error}`);
      }

      state.processed += 1;
      updateFileRow(item);
      refreshStatus();
    }
  } finally {
    state.busy = false;
    state.phase = 'done';
    setBusy(false);
    renderFiles();
    refreshStatus();
  }

  if (state.errors > 0) {
    toast('error', t('calcPartial'));
  } else {
    toast('success', t('calcDone'));
  }
}

async function copyAll() {
  const text = buildExportText();
  if (!text) {
    toast('error', t('exportEmpty'));
    return;
  }
  const ok = await copyText(text);
  toast(ok ? 'success' : 'error', ok ? t('copied') : t('copyFailed'));
}

async function exportResults() {
  const text = buildExportText();
  if (!text) {
    toast('error', t('exportEmpty'));
    return;
  }

  if (window.showSaveFilePicker) {
    try {
      const handle = await window.showSaveFilePicker({
        suggestedName: 'hash-results.txt',
        types: [{ description: t('textFiles'), accept: { 'text/plain': ['.txt'] } }],
      });
      const writable = await handle.createWritable();
      await writable.write(text);
      await writable.close();
      toast('success', t('exportSuccess'));
      return;
    } catch (error) {
      if (error && error.name === 'AbortError') return;
      /* fall back to a regular download */
    }
  }

  try {
    const blob = new Blob([text], { type: 'text/plain;charset=utf-8' });
    const url = URL.createObjectURL(blob);
    const link = document.createElement('a');
    link.href = url;
    link.download = 'hash-results.txt';
    document.body.appendChild(link);
    link.click();
    link.remove();
    window.setTimeout(() => URL.revokeObjectURL(url), 1000);
    toast('success', t('exportSuccess'));
  } catch {
    toast('error', t('exportFailed'));
  }
}

/* ==========================================================================
   About modal
   ========================================================================== */

let lastFocused = null;

function openAbout() {
  lastFocused = document.activeElement;
  els.aboutModal.hidden = false;
  els.aboutClose.focus();
}

function closeAbout() {
  if (els.aboutModal.hidden) return;
  els.aboutModal.hidden = true;
  if (lastFocused && typeof lastFocused.focus === 'function') lastFocused.focus();
}

/* ==========================================================================
   Events
   ========================================================================== */

function wireEvents() {
  els.themeBtn.addEventListener('click', () => {
    state.theme = state.theme === 'light' ? 'dark' : 'light';
    applyTheme();
    applyLang();
  });

  els.langBtn.addEventListener('click', () => {
    state.lang = state.lang === 'zh' ? 'en' : 'zh';
    applyLang();
  });

  els.aboutBtn.addEventListener('click', openAbout);
  els.aboutClose.addEventListener('click', closeAbout);
  els.aboutModal.querySelector('[data-close-modal]').addEventListener('click', closeAbout);

  els.glassRange.addEventListener('input', () => {
    state.glass = Number.parseFloat(els.glassRange.value);
    applyGlass();
  });

  els.selectAllBtn.addEventListener('click', selectAllAlgorithms);
  els.resetAlgosBtn.addEventListener('click', resetAlgorithms);

  els.browseBtn.addEventListener('click', (event) => {
    event.stopPropagation();
    if (!state.busy) els.fileInput.click();
  });

  els.dropzone.addEventListener('click', (event) => {
    if (event.target.closest('button')) return;
    if (!state.busy) els.fileInput.click();
  });

  els.dropzone.addEventListener('keydown', (event) => {
    if (event.key !== 'Enter' && event.key !== ' ') return;
    event.preventDefault();
    if (!state.busy) els.fileInput.click();
  });

  els.fileInput.addEventListener('change', () => {
    if (els.fileInput.files && els.fileInput.files.length > 0) {
      addFiles(els.fileInput.files);
    }
    els.fileInput.value = '';
  });

  let dragDepth = 0;
  els.dropzone.addEventListener('dragenter', (event) => {
    event.preventDefault();
    dragDepth += 1;
    els.dropzone.classList.add('is-dragging');
  });
  els.dropzone.addEventListener('dragover', (event) => {
    event.preventDefault();
    if (event.dataTransfer) event.dataTransfer.dropEffect = 'copy';
  });
  els.dropzone.addEventListener('dragleave', () => {
    dragDepth = Math.max(0, dragDepth - 1);
    if (dragDepth === 0) els.dropzone.classList.remove('is-dragging');
  });
  els.dropzone.addEventListener('drop', (event) => {
    event.preventDefault();
    event.stopPropagation();
    dragDepth = 0;
    els.dropzone.classList.remove('is-dragging');
    if (event.dataTransfer && event.dataTransfer.files.length > 0) {
      addFiles(event.dataTransfer.files);
    }
  });

  window.addEventListener('dragover', (event) => {
    if (event.dataTransfer && Array.from(event.dataTransfer.types).includes('Files')) {
      event.preventDefault();
    }
  });
  window.addEventListener('drop', (event) => {
    if (event.dataTransfer && Array.from(event.dataTransfer.types).includes('Files')) {
      event.preventDefault();
    }
  });

  els.clearBtn.addEventListener('click', clearAll);
  els.calcBtn.addEventListener('click', startCalculation);
  els.copyAllBtn.addEventListener('click', copyAll);
  els.exportBtn.addEventListener('click', exportResults);

  document.addEventListener('keydown', (event) => {
    if (event.key === 'Escape') {
      closeAbout();
      return;
    }
    const modifier = event.ctrlKey || event.metaKey;
    if (!modifier) return;
    if (event.key.toLowerCase() === 'o') {
      event.preventDefault();
      if (!state.busy) els.fileInput.click();
    } else if (event.key === 'Enter') {
      event.preventDefault();
      startCalculation();
    }
  });
}

function connectLive() {
  if (!('EventSource' in window)) return;
  try {
    const source = new EventSource('/api/live');
    window.addEventListener('pagehide', () => source.close());
  } catch {
    /* liveness is optional */
  }
}

/* ==========================================================================
   Init
   ========================================================================== */

function init() {
  const savedGlass = Number.parseFloat(store.get('fc.glass', '0.25'));
  state.glass = Number.isFinite(savedGlass) ? Math.min(0.5, Math.max(0.12, savedGlass)) : 0.25;

  buildAlgorithms();
  wireEvents();
  applyTheme();
  applyGlass();
  applyLang();
  refreshStatus();
  connectLive();
}

init();
