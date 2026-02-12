<script setup lang="ts">
import { ref, computed, watch } from 'vue';
import { useI18n } from 'vue-i18n';
import type { ProxyHistoryItem } from '../../store';

const { t } = useI18n();

export interface FilterOptions {
  method: string;
  host: string;
  hostMode: 'include' | 'exclude';
  path: string;
  pathMode: 'include' | 'exclude';
  status: string[];
  statusSearch: string;
  contentType: string;
  extension: string[];
  extensionMode: 'include' | 'exclude';
}

const props = defineProps<{
  initialFilter?: Partial<FilterOptions>;
  proxyHistory: ProxyHistoryItem[];
}>();

const emit = defineEmits<{
  (e: 'filter', filterOptions: FilterOptions): void;
  (e: 'reset'): void;
}>();

const filterOptions = ref<FilterOptions>({
  method: props.initialFilter?.method || '',
  host: props.initialFilter?.host || '',
  hostMode: props.initialFilter?.hostMode || 'include',
  path: props.initialFilter?.path || '',
  pathMode: props.initialFilter?.pathMode || 'include',
  status: props.initialFilter?.status || [],
  statusSearch: props.initialFilter?.statusSearch || '',
  contentType: props.initialFilter?.contentType || '',
  extension: props.initialFilter?.extension || [],
  extensionMode: props.initialFilter?.extensionMode || 'exclude',
});

const methodOptions = ['', 'GET', 'POST', 'PUT', 'DELETE', 'PATCH', 'OPTIONS', 'HEAD'];

// 预置的常用扩展名（静态资源类）
const presetExtensions = ['js', 'css', 'png', 'jpg', 'jpeg', 'gif', 'svg', 'ico', 'woff', 'woff2', 'ttf', 'eot', 'map'];

// 从当前数据中动态提取状态码及计数
const statusCodeOptions = computed(() => {
  const countMap = new Map<string, number>();
  for (const item of props.proxyHistory) {
    const code = String(item.status);
    if (code && code !== '0') {
      countMap.set(code, (countMap.get(code) || 0) + 1);
    }
  }
  const options = Array.from(countMap.entries())
    .map(([code, count]) => ({ code, count }))
    .sort((a, b) => Number(a.code) - Number(b.code));
  return options;
});

// 根据搜索前缀过滤状态码选项
const filteredStatusCodes = computed(() => {
  const search = filterOptions.value.statusSearch.trim();
  if (!search) return statusCodeOptions.value;
  return statusCodeOptions.value.filter(opt => opt.code.startsWith(search));
});

// 从当前数据中动态提取扩展名及计数
const extensionOptions = computed(() => {
  const countMap = new Map<string, number>();
  for (const item of props.proxyHistory) {
    const ext = (item.extension || '').toLowerCase().replace(/^\./, '');
    if (ext) {
      countMap.set(ext, (countMap.get(ext) || 0) + 1);
    }
  }
  // 合并预置扩展名（即使数据中没有也显示，计数为0）
  for (const ext of presetExtensions) {
    if (!countMap.has(ext)) {
      countMap.set(ext, 0);
    }
  }
  return Array.from(countMap.entries())
    .map(([ext, count]) => ({ ext, count }))
    .sort((a, b) => {
      // 预置的排在前面
      const aPreset = presetExtensions.includes(a.ext) ? 0 : 1;
      const bPreset = presetExtensions.includes(b.ext) ? 0 : 1;
      if (aPreset !== bPreset) return aPreset - bPreset;
      return a.ext.localeCompare(b.ext);
    });
});

// 切换状态码选中
const toggleStatusCode = (code: string) => {
  const idx = filterOptions.value.status.indexOf(code);
  if (idx > -1) {
    filterOptions.value.status.splice(idx, 1);
  } else {
    filterOptions.value.status.push(code);
  }
};

// 切换扩展名选中
const toggleExtension = (ext: string) => {
  const idx = filterOptions.value.extension.indexOf(ext);
  if (idx > -1) {
    filterOptions.value.extension.splice(idx, 1);
  } else {
    filterOptions.value.extension.push(ext);
  }
};

// 全选/取消全选状态码
const toggleAllStatusCodes = () => {
  if (filterOptions.value.status.length === filteredStatusCodes.value.length) {
    filterOptions.value.status = [];
  } else {
    filterOptions.value.status = filteredStatusCodes.value.map(opt => opt.code);
  }
};

// 快速预置：排除静态资源
const applyPresetExcludeStatic = () => {
  filterOptions.value.extensionMode = 'exclude';
  filterOptions.value.extension = [...presetExtensions.filter(ext => 
    ['js', 'css', 'png', 'jpg', 'jpeg', 'gif', 'svg', 'ico', 'woff', 'woff2', 'ttf', 'eot', 'map'].includes(ext)
  )];
};

const applyFilter = () => {
  emit('filter', { ...filterOptions.value, status: [...filterOptions.value.status], extension: [...filterOptions.value.extension] });
};

const resetFilter = () => {
  filterOptions.value = {
    method: '',
    host: '',
    hostMode: 'include',
    path: '',
    pathMode: 'include',
    status: [],
    statusSearch: '',
    contentType: '',
    extension: [],
    extensionMode: 'exclude',
  };
  emit('reset');
};

// 计算激活的筛选条件数
const activeFiltersCount = computed(() => {
  let count = 0;
  if (filterOptions.value.method) count++;
  if (filterOptions.value.host) count++;
  if (filterOptions.value.path) count++;
  if (filterOptions.value.status.length > 0) count++;
  if (filterOptions.value.contentType) count++;
  if (filterOptions.value.extension.length > 0) count++;
  return count;
});

watch(filterOptions, () => {
  applyFilter();
}, { deep: true });
</script>

<template>
  <div class="proxy-filter-enhanced">
    <!-- 头部 -->
    <div class="proxy-filter-header">
      <div class="proxy-filter-title">
        <i class="bx bx-filter-alt"></i>
        {{ t('modules.proxy.filter.title') }}
        <span v-if="activeFiltersCount > 0" class="active-filter-badge">{{ activeFiltersCount }}</span>
      </div>
      <button @click="resetFilter" class="btn btn-secondary btn-sm">
        <i class="bx bx-reset"></i>
        {{ t('modules.proxy.filter.reset') }}
      </button>
    </div>

    <!-- 第一行：Method / Host / Path / ContentType -->
    <div class="proxy-filter-row">
      <!-- Method -->
      <div class="proxy-filter-field">
        <label class="proxy-filter-label">{{ t('modules.proxy.filter.method') }}</label>
        <select v-model="filterOptions.method" class="proxy-filter-input">
          <option value="">{{ t('common.ui.none') }}</option>
          <option v-for="method in methodOptions.slice(1)" :key="method" :value="method">{{ method }}</option>
        </select>
      </div>

      <!-- Host -->
      <div class="proxy-filter-field">
        <label class="proxy-filter-label">
          {{ t('modules.proxy.filter.host') }}
          <button
            class="mode-toggle-btn"
            :class="{ 'mode-exclude': filterOptions.hostMode === 'exclude' }"
            @click="filterOptions.hostMode = filterOptions.hostMode === 'include' ? 'exclude' : 'include'"
            :title="filterOptions.hostMode === 'include' ? t('modules.proxy.filter.include_targets') : t('modules.proxy.filter.exclude_targets')"
          >
            <i :class="filterOptions.hostMode === 'include' ? 'bx bx-plus-circle' : 'bx bx-minus-circle'"></i>
            {{ filterOptions.hostMode === 'include' ? 'Include' : 'Exclude' }}
          </button>
        </label>
        <input
          v-model="filterOptions.host"
          type="text"
          class="proxy-filter-input"
          :placeholder="t('modules.proxy.filter.placeholder')"
        />
      </div>

      <!-- Path -->
      <div class="proxy-filter-field">
        <label class="proxy-filter-label">
          {{ t('modules.proxy.filter.path') }}
          <button
            class="mode-toggle-btn"
            :class="{ 'mode-exclude': filterOptions.pathMode === 'exclude' }"
            @click="filterOptions.pathMode = filterOptions.pathMode === 'include' ? 'exclude' : 'include'"
            :title="filterOptions.pathMode === 'include' ? t('modules.proxy.filter.include_targets') : t('modules.proxy.filter.exclude_targets')"
          >
            <i :class="filterOptions.pathMode === 'include' ? 'bx bx-plus-circle' : 'bx bx-minus-circle'"></i>
            {{ filterOptions.pathMode === 'include' ? 'Include' : 'Exclude' }}
          </button>
        </label>
        <input
          v-model="filterOptions.path"
          type="text"
          class="proxy-filter-input"
          :placeholder="t('modules.proxy.filter.placeholder')"
        />
      </div>

      <!-- Content Type -->
      <div class="proxy-filter-field">
        <label class="proxy-filter-label">{{ t('modules.proxy.filter.content_type') }}</label>
        <input
          v-model="filterOptions.contentType"
          type="text"
          class="proxy-filter-input"
          :placeholder="t('modules.proxy.filter.placeholder')"
        />
      </div>
    </div>

    <!-- 第二行：Status Codes + Extension -->
    <div class="proxy-filter-panels">
      <!-- Status Code Panel -->
      <div class="proxy-filter-panel">
        <div class="panel-header">
          <span class="panel-title">
            <i class="bx bx-hash"></i>
            {{ t('modules.proxy.filter.status') }}
            <span v-if="filterOptions.status.length > 0" class="panel-count">{{ filterOptions.status.length }}</span>
          </span>
          <button class="panel-action-btn" @click="toggleAllStatusCodes" :title="filterOptions.status.length === filteredStatusCodes.length ? '取消全选' : '全选'">
            <i :class="filterOptions.status.length === filteredStatusCodes.length && filteredStatusCodes.length > 0 ? 'bx bx-checkbox-checked' : 'bx bx-checkbox'"></i>
          </button>
        </div>
        <input
          v-model="filterOptions.statusSearch"
          type="text"
          class="proxy-filter-input panel-search"
          placeholder="输入前缀匹配，如 20、30..."
        />
        <div class="checkbox-list-compact">
          <div
            v-for="opt in filteredStatusCodes"
            :key="opt.code"
            class="checkbox-item-compact"
            :class="{ 'is-checked': filterOptions.status.includes(opt.code) }"
            @click="toggleStatusCode(opt.code)"
          >
            <span class="status-code-badge" :class="'status-' + opt.code.charAt(0) + 'xx'">{{ opt.code }}</span>
            <span class="item-count">({{ opt.count }})</span>
            <i v-if="filterOptions.status.includes(opt.code)" class="bx bx-check check-icon"></i>
          </div>
          <div v-if="filteredStatusCodes.length === 0" class="empty-hint">
            {{ statusCodeOptions.length === 0 ? '暂无数据' : '无匹配的状态码' }}
          </div>
        </div>
      </div>

      <!-- Extension Panel -->
      <div class="proxy-filter-panel">
        <div class="panel-header">
          <span class="panel-title">
            <i class="bx bx-file"></i>
            Extension
            <span v-if="filterOptions.extension.length > 0" class="panel-count">{{ filterOptions.extension.length }}</span>
          </span>
          <div class="panel-header-actions">
            <button
              class="mode-toggle-btn"
              :class="{ 'mode-exclude': filterOptions.extensionMode === 'exclude' }"
              @click="filterOptions.extensionMode = filterOptions.extensionMode === 'include' ? 'exclude' : 'include'"
            >
              <i :class="filterOptions.extensionMode === 'include' ? 'bx bx-show' : 'bx bx-hide'"></i>
              {{ filterOptions.extensionMode === 'include' ? 'Include' : 'Exclude' }}
            </button>
            <button class="preset-btn" @click="applyPresetExcludeStatic" title="快速排除静态资源">
              <i class="bx bx-bolt-circle"></i>
            </button>
          </div>
        </div>
        <div class="checkbox-list-compact ext-list">
          <div
            v-for="opt in extensionOptions"
            :key="opt.ext"
            class="checkbox-item-compact"
            :class="{ 'is-checked': filterOptions.extension.includes(opt.ext) }"
            @click="toggleExtension(opt.ext)"
          >
            <span class="ext-name">.{{ opt.ext }}</span>
            <span v-if="opt.count > 0" class="item-count">({{ opt.count }})</span>
            <i v-if="filterOptions.extension.includes(opt.ext)" class="bx bx-check check-icon"></i>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
.proxy-filter-enhanced {
  padding: 0.5rem 0.75rem;
}

.proxy-filter-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 0.5rem;
}

.proxy-filter-title {
  display: flex;
  align-items: center;
  gap: 0.375rem;
  font-size: 0.8125rem;
  font-weight: 600;
  color: var(--color-text-primary);
}

.proxy-filter-title i {
  font-size: 1rem;
  color: var(--color-primary, #6366f1);
}

.active-filter-badge {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  min-width: 18px;
  height: 18px;
  padding: 0 5px;
  background: var(--color-primary, #6366f1);
  color: white;
  font-size: 0.6875rem;
  font-weight: 600;
  border-radius: 9px;
}

/* 第一行 */
.proxy-filter-row {
  display: grid;
  grid-template-columns: 0.8fr 1.2fr 1.2fr 1fr;
  gap: 0.5rem;
  margin-bottom: 0.5rem;
}

.proxy-filter-field {
  display: flex;
  flex-direction: column;
  gap: 0.25rem;
}

.proxy-filter-label {
  display: flex;
  align-items: center;
  justify-content: space-between;
  font-size: 0.6875rem;
  font-weight: 500;
  color: var(--color-text-secondary);
  text-transform: uppercase;
  letter-spacing: 0.03em;
}

.proxy-filter-input {
  width: 100%;
  padding: 0.3rem 0.5rem;
  font-size: 0.8125rem;
  border: 1px solid var(--color-border);
  border-radius: 0.3rem;
  background-color: var(--color-bg-primary);
  color: var(--color-text-primary);
  transition: border-color 0.15s ease;
}

.proxy-filter-input:focus {
  outline: none;
  border-color: var(--color-primary, #6366f1);
  box-shadow: 0 0 0 2px rgba(99, 102, 241, 0.15);
}

.dark .proxy-filter-input {
  background-color: rgba(49, 49, 70, 0.8);
  border-color: var(--color-border-light);
}

.dark .proxy-filter-input:focus {
  border-color: var(--color-primary, #6366f1);
  box-shadow: 0 0 0 2px rgba(99, 102, 241, 0.2);
}

/* Mode Toggle Button */
.mode-toggle-btn {
  display: inline-flex;
  align-items: center;
  gap: 0.2rem;
  padding: 0.1rem 0.35rem;
  font-size: 0.625rem;
  font-weight: 500;
  border: 1px solid var(--color-border);
  border-radius: 0.25rem;
  background: var(--color-bg-secondary);
  color: var(--color-text-secondary);
  cursor: pointer;
  transition: all 0.15s ease;
  line-height: 1;
}

.mode-toggle-btn i {
  font-size: 0.75rem;
}

.mode-toggle-btn:hover {
  border-color: var(--color-primary, #6366f1);
  color: var(--color-primary, #6366f1);
}

.mode-toggle-btn.mode-exclude {
  background: rgba(239, 68, 68, 0.08);
  border-color: rgba(239, 68, 68, 0.3);
  color: rgb(239, 68, 68);
}

.mode-toggle-btn.mode-exclude:hover {
  background: rgba(239, 68, 68, 0.15);
}

/* Panels Row */
.proxy-filter-panels {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 0.5rem;
}

.proxy-filter-panel {
  border: 1px solid var(--color-border);
  border-radius: 0.375rem;
  overflow: hidden;
  background: var(--color-bg-primary);
}

.dark .proxy-filter-panel {
  background: rgba(49, 49, 70, 0.5);
  border-color: var(--color-border-light);
}

.panel-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 0.35rem 0.5rem;
  background: var(--color-bg-secondary);
  border-bottom: 1px solid var(--color-border);
}

.dark .panel-header {
  background: rgba(42, 42, 66, 0.6);
  border-color: var(--color-border-light);
}

.panel-title {
  display: flex;
  align-items: center;
  gap: 0.3rem;
  font-size: 0.75rem;
  font-weight: 600;
  color: var(--color-text-primary);
}

.panel-title i {
  font-size: 0.875rem;
  color: var(--color-text-secondary);
}

.panel-count {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  min-width: 16px;
  height: 16px;
  padding: 0 4px;
  background: var(--color-primary, #6366f1);
  color: white;
  font-size: 0.625rem;
  font-weight: 600;
  border-radius: 8px;
}

.panel-header-actions {
  display: flex;
  align-items: center;
  gap: 0.25rem;
}

.panel-action-btn {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 22px;
  height: 22px;
  border: none;
  background: transparent;
  color: var(--color-text-secondary);
  cursor: pointer;
  border-radius: 0.2rem;
  transition: all 0.15s ease;
}

.panel-action-btn:hover {
  background: var(--color-bg-hover);
  color: var(--color-primary, #6366f1);
}

.panel-action-btn i {
  font-size: 1rem;
}

.preset-btn {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 22px;
  height: 22px;
  border: 1px solid var(--color-border);
  background: transparent;
  color: var(--color-text-secondary);
  cursor: pointer;
  border-radius: 0.2rem;
  transition: all 0.15s ease;
}

.preset-btn:hover {
  background: rgba(245, 158, 11, 0.1);
  border-color: rgb(245, 158, 11);
  color: rgb(245, 158, 11);
}

.preset-btn i {
  font-size: 0.875rem;
}

.panel-search {
  margin: 0.35rem 0.4rem;
  width: calc(100% - 0.8rem);
  padding: 0.25rem 0.4rem;
  font-size: 0.75rem;
}

/* Checkbox list compact */
.checkbox-list-compact {
  max-height: 160px;
  overflow-y: auto;
  padding: 0.25rem;
  display: flex;
  flex-wrap: wrap;
  gap: 0.25rem;
}

.ext-list {
  max-height: 160px;
}

.checkbox-item-compact {
  display: inline-flex;
  align-items: center;
  gap: 0.25rem;
  padding: 0.2rem 0.4rem;
  border: 1px solid var(--color-border);
  border-radius: 0.25rem;
  cursor: pointer;
  font-size: 0.75rem;
  color: var(--color-text-primary);
  transition: all 0.12s ease;
  user-select: none;
  background: transparent;
}

.checkbox-item-compact:hover {
  background: var(--color-bg-hover);
  border-color: var(--color-primary, #6366f1);
}

.checkbox-item-compact.is-checked {
  background: rgba(99, 102, 241, 0.1);
  border-color: rgba(99, 102, 241, 0.4);
}

.checkbox-item-compact.is-checked .check-icon {
  color: var(--color-primary, #6366f1);
  font-size: 0.8rem;
}

.item-count {
  font-size: 0.625rem;
  color: var(--color-text-tertiary);
}

.ext-name {
  font-family: ui-monospace, monospace;
  font-size: 0.75rem;
}

.check-icon {
  font-size: 0.8rem;
  color: var(--color-primary, #6366f1);
}

.empty-hint {
  width: 100%;
  text-align: center;
  padding: 0.5rem;
  font-size: 0.75rem;
  color: var(--color-text-tertiary);
}

/* Status code badge colors */
.status-code-badge {
  font-size: 0.6875rem;
  font-weight: 600;
  font-family: ui-monospace, monospace;
  padding: 0.05rem 0.3rem;
  border-radius: 0.2rem;
}

.status-1xx { color: #6b7280; background: rgba(107, 114, 128, 0.1); }
.status-2xx { color: #059669; background: rgba(5, 150, 105, 0.1); }
.status-3xx { color: #d97706; background: rgba(217, 119, 6, 0.1); }
.status-4xx { color: #dc2626; background: rgba(220, 38, 38, 0.1); }
.status-5xx { color: #7c3aed; background: rgba(124, 58, 237, 0.1); }

/* Scrollbar */
.checkbox-list-compact::-webkit-scrollbar {
  width: 4px;
}
.checkbox-list-compact::-webkit-scrollbar-track {
  background: transparent;
}
.checkbox-list-compact::-webkit-scrollbar-thumb {
  background: var(--color-border);
  border-radius: 2px;
}
</style>
