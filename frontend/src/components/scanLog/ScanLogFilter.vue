<script setup lang="ts">
import { ref, computed, watch } from 'vue';
import { useI18n } from 'vue-i18n';
import type { ScanLogFilter, ScanModuleType } from '../../types/scanLog';
import { useScanLogStore } from '../../store/scanLog';
import { getModuleColor } from '../../utils/colors';

const { t } = useI18n();

const props = defineProps<{
  filter: ScanLogFilter;
}>();

const emit = defineEmits<{
  (e: 'update:filter', filter: ScanLogFilter): void;
}>();

// 本地过滤器状态
const localFilter = ref<ScanLogFilter>({ 
  moduleTypes: props.filter.moduleTypes || [],
  searchKeyword: props.filter.searchKeyword || '',
  targetHost: props.filter.targetHost || ''
});

const scanLogStore = useScanLogStore();

// 从store中动态获取模块类型选项
const moduleTypeOptions = computed(() => {
  const modules = scanLogStore.statistics.byModule;
  return Object.keys(modules).map(moduleName => {
    const moduleColor = getModuleColor(moduleName);
    return {
      value: moduleName,
      label: moduleName,
      icon: getModuleIcon(moduleName),
      color: moduleColor.color,
      textClass: moduleColor.textClass,
      count: modules[moduleName]
    };
  });
});

// 根据模块名称获取合适的图标
const getModuleIcon = (moduleName: string): string => {
  const name = moduleName.toLowerCase();
  if (name.includes('xss')) return 'bx-code-alt';
  if (name.includes('sql')) return 'bx-data';
  if (name.includes('auth')) return 'bx-shield-quarter';
  if (name.includes('upload')) return 'bx-upload';
  if (name.includes('xxe')) return 'bx-file-blank';
  if (name.includes('ssrf')) return 'bx-link-external';
  if (name.includes('cmd') || name.includes('inject')) return 'bx-terminal';
  if (name.includes('sensitive')) return 'bx-shield-alt-2';
  if (name.includes('waf')) return 'bx-shield-x';
  if (name.includes('brute')) return 'bx-key';
  if (name.includes('finger')) return 'bx-fingerprint';
  if (name.includes('port') || name.includes('scan')) return 'bx-search';
  if (name.includes('subdomain') || name.includes('domain')) return 'bx-network-chart';
  if (name.includes('directory') || name.includes('bbscan')) return 'bx-folder';
  if (name.includes('json')) return 'bx-bug';
  if (name.includes('swagger')) return 'bx-file-doc';
  if (name.includes('raw')) return 'bx-code';
  if (name.includes('http')) return 'bx-transfer-alt';
  return 'bx-cog';
};



// 更新过滤器
const updateFilter = () => {
  emit('update:filter', { ...localFilter.value });
};

// 切换模块类型
const toggleModuleType = (moduleType: ScanModuleType) => {
  const index = localFilter.value.moduleTypes.indexOf(moduleType);
  if (index > -1) {
    localFilter.value.moduleTypes.splice(index, 1);
  } else {
    localFilter.value.moduleTypes.push(moduleType);
  }
  updateFilter();
};



// 清空所有过滤器
const clearAllFilters = () => {
  localFilter.value = {
    moduleTypes: [],
    searchKeyword: '',
    targetHost: ''
  };
  updateFilter();
};

// 搜索关键词更新
const updateSearchKeyword = () => {
  updateFilter();
};

// 目标主机更新
const updateTargetHost = () => {
  updateFilter();
};

// 监听 props 变化，同步本地状态
watch(() => props.filter, (newFilter) => {
  localFilter.value = {
    moduleTypes: newFilter.moduleTypes || [],
    searchKeyword: newFilter.searchKeyword || '',
    targetHost: newFilter.targetHost || ''
  };
}, { deep: true });

// 计算激活的过滤器数量
const activeFiltersCount = computed(() => {
  return localFilter.value.moduleTypes.length + 
         (localFilter.value.searchKeyword && localFilter.value.searchKeyword.trim() ? 1 : 0) +
         (localFilter.value.targetHost && localFilter.value.targetHost.trim() ? 1 : 0);
});
</script>

<template>
  <div class="filter-container scrollbar-thin">
    <!-- 过滤器标题 -->
    <div class="filter-header">
      <h3>{{ t('scanLog.filter.title', '过滤器') }}</h3>
      <div class="filter-actions">
        <span v-if="activeFiltersCount > 0" class="active-count">
          {{ activeFiltersCount }}
        </span>
        <button
          @click="clearAllFilters"
          class="btn-icon"
          :title="t('common.clear', '清空')"
        >
          <i class="bx bx-trash"></i>
        </button>
      </div>
    </div>

    <!-- 搜索框 -->
    <div class="form-group spacing-sm">
      <label>{{ t('scanLog.filter.search', '搜索关键词') }}</label>
      <input
        v-model="localFilter.searchKeyword"
        @input="updateSearchKeyword"
        type="text"
        placeholder="搜索URL、模块、漏洞类型等..."
        class="filter-input"
        spellcheck="false"
      />
    </div>

    <!-- 目标主机过滤 -->
    <div class="form-group spacing-sm">
      <label>{{ t('scanLog.filter.targetHost', '目标主机') }}</label>
      <input
        v-model="localFilter.targetHost"
        @input="updateTargetHost"
        type="text"
        placeholder="example.com"
        class="filter-input"
        spellcheck="false"
      />
    </div>

    <!-- 模块类型过滤 -->
    <div class="form-group spacing-sm">
      <label>{{ t('scanLog.filter.moduleTypes', '扫描模块') }}</label>
      <div class="checkbox-list">
        <div
          v-for="option in moduleTypeOptions"
          :key="option.value"
          @click="toggleModuleType(option.value as ScanModuleType)"
          class="checkbox-item module-item"
          :data-module="option.value"
          :style="{ borderLeftColor: option.color }"
        >
          <div class="checkbox-content">
            <i :class="option.icon" class="option-icon"></i>
            <div 
              class="module-color-badge"
              :style="{ backgroundColor: option.color }"
            >
              <span class="option-label">{{ option.label }}</span>
            </div>
            <span class="option-count">({{ option.count }})</span>
          </div>
          <div v-if="localFilter.moduleTypes.includes(option.value as ScanModuleType)" class="check-indicator">
            <i class="bx bx-check"></i>
          </div>
        </div>
      </div>
    </div>


  </div>
</template>

<style scoped>
.filter-container {
  padding: var(--spacing-md);
  display: flex;
  flex-direction: column;
  gap: var(--spacing-md);
}

.filter-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding-bottom: var(--spacing-sm);
  border-bottom: 1px solid var(--glass-border-light);
}

.filter-header h3 {
  font-size: var(--text-sm);
  font-weight: var(--font-weight-semibold);
  color: var(--color-text-primary);
  margin: 0;
  display: flex;
  align-items: center;
  gap: var(--spacing-xs);
}

.filter-actions {
  display: flex;
  align-items: center;
  gap: var(--spacing-xs);
}

.active-count {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  min-width: 20px;
  height: 20px;
  padding: 0 6px;
  background: linear-gradient(135deg, var(--color-primary), var(--color-primary-dark));
  color: white;
  font-size: 11px;
  font-weight: var(--font-weight-semibold);
  border-radius: var(--radius-full);
}

.btn-icon {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 28px;
  height: 28px;
  background: var(--glass-bg-tertiary);
  border: 1px solid var(--glass-border-light);
  border-radius: var(--radius-sm);
  color: var(--color-text-secondary);
  cursor: pointer;
  transition: all var(--glass-transition-fast);
}

.btn-icon:hover {
  background: var(--glass-bg-hover);
  border-color: var(--color-danger);
  color: var(--color-danger);
}

/* 表单组 */
.form-group {
  display: flex;
  flex-direction: column;
  gap: var(--spacing-xs);
}

.form-group label {
  font-size: var(--text-xs);
  font-weight: var(--font-weight-medium);
  color: var(--color-text-secondary);
  text-transform: uppercase;
  letter-spacing: 0.05em;
}

.filter-input {
  width: 100%;
  padding: var(--spacing-sm) var(--spacing-md);
  background: var(--glass-bg-tertiary);
  border: 1px solid var(--glass-border-light);
  border-radius: var(--radius-sm);
  font-size: var(--text-sm);
  color: var(--color-text-primary);
  transition: all var(--glass-transition-fast);
}

.filter-input::placeholder {
  color: var(--color-text-tertiary);
}

.filter-input:focus {
  outline: none;
  border-color: var(--color-primary);
  background: var(--glass-bg-card);
  box-shadow: 0 0 0 3px rgba(var(--color-primary-rgb), 0.1);
}

/* 复选框列表 */
.checkbox-list {
  display: flex;
  flex-direction: column;
  gap: var(--spacing-xs);
  max-height: 300px;
  overflow-y: auto;
}

.checkbox-item {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: var(--spacing-sm);
  background: var(--glass-bg-tertiary);
  border: 1px solid var(--glass-border-light);
  border-radius: var(--radius-sm);
  cursor: pointer;
  transition: all var(--glass-transition-fast);
}

.checkbox-item:hover {
  background: var(--glass-bg-hover);
  border-color: var(--glass-border);
}

.checkbox-content {
  display: flex;
  align-items: center;
  gap: var(--spacing-xs);
  flex: 1;
  min-width: 0;
}

.option-icon {
  font-size: 14px;
  color: var(--color-text-secondary);
  flex-shrink: 0;
}

.option-count {
  font-size: var(--text-xs);
  color: var(--color-text-tertiary);
  margin-left: auto;
  flex-shrink: 0;
}

/* 模块项特殊样式 */
.module-item {
  border-left: 3px solid transparent;
  transition: all var(--glass-transition-fast);
}

.module-item:hover {
  transform: translateX(2px);
}

.module-color-badge {
  display: inline-flex;
  align-items: center;
  padding: 2px 8px;
  border-radius: var(--radius-full);
  font-size: 11px;
  font-weight: 500;
  color: white;
  text-shadow: 0 1px 2px rgba(0, 0, 0, 0.2);
  max-width: 120px;
  overflow: hidden;
}

.module-color-badge .option-label {
  color: inherit;
  font-size: inherit;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

/* 选中指示器 */
.check-indicator {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 20px;
  height: 20px;
  background: linear-gradient(135deg, var(--color-success), var(--color-success-dark, #059669));
  border-radius: var(--radius-full);
  color: white;
  font-size: 14px;
  flex-shrink: 0;
}

/* 间距工具类 */
.spacing-sm {
  margin-top: var(--spacing-sm);
}
</style> 