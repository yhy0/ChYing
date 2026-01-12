<script setup lang="ts">
import { ref, nextTick } from 'vue';
import { useI18n } from 'vue-i18n';
import { type ScopeRule, type ProxyConfig, type CompleteConfig } from '../../../types';

const props = defineProps<{
  configValue: ProxyConfig;
  completeConfig: CompleteConfig | null;
}>();

const emit = defineEmits<{
  (e: 'saveFilterSuffix'): void;
  (e: 'updateHttpConfig'): void;
  (e: 'openAddModal', type: string): void;
  (e: 'openEditModal', rule: ScopeRule, type: string): void;
  (e: 'deleteRule', rule: ScopeRule, type: string): void;
  (e: 'updateRuleToggle', rule: ScopeRule, type: string, field: string, event: Event): void;
}>();

const { t } = useI18n();

// 列宽调整状态
const isResizing = ref(false);
const startX = ref(0);
const startWidth = ref(0);
const currentResizingColumn = ref<HTMLElement | null>(null);

// 初始化可调整列宽功能
const initResizableColumns = () => {
  nextTick(() => {
    const tables = document.querySelectorAll('.resizable-table');
    tables.forEach(table => {
      const cols = table.querySelectorAll('th.resizable');
      cols.forEach(col => {
        // 添加调整手柄
        const resizer = document.createElement('div');
        resizer.classList.add('column-resizer');
        resizer.title = t('modules.proxy.filter.resize_column');
        col.appendChild(resizer);
        
        // 添加手柄拖动事件
        resizer.addEventListener('mousedown', (e) => {
          e.preventDefault();
          isResizing.value = true;
          startX.value = e.pageX;
          currentResizingColumn.value = col as HTMLElement;
          startWidth.value = col.getBoundingClientRect().width;
          
          // 添加全局鼠标移动和鼠标松开事件
          document.addEventListener('mousemove', handleMouseMove);
          document.addEventListener('mouseup', handleMouseUp);
          
          // 添加调整中样式
          document.body.classList.add('column-resizing');
        });
      });
    });
  });
};

// 鼠标移动处理
const handleMouseMove = (e: MouseEvent) => {
  if (!isResizing.value || !currentResizingColumn.value) return;
  
  const diffX = e.pageX - startX.value;
  const newWidth = Math.max(50, startWidth.value + diffX); // 最小宽度50px
  
  currentResizingColumn.value.style.width = `${newWidth}px`;
};

// 鼠标松开处理
const handleMouseUp = () => {
  isResizing.value = false;
  currentResizingColumn.value = null;
  document.removeEventListener('mousemove', handleMouseMove);
  document.removeEventListener('mouseup', handleMouseUp);
  document.body.classList.remove('column-resizing');
};

// 暴露方法给父组件
defineExpose({
  initResizableColumns
});
</script>

<template>
  <!-- 过滤后缀配置 -->
  <div class="bg-white dark:bg-[#282838] rounded-lg p-5 shadow-sm border border-gray-100 dark:border-gray-700">
    <div class="flex items-center justify-between mb-4">
      <h3 class="text-lg font-medium">{{ t('modules.proxy.filter.suffix') }}</h3>
      <button 
        @click="emit('saveFilterSuffix')"
        class="px-3 py-1.5 bg-amber-500 hover:bg-amber-600 text-white rounded-md text-sm font-medium transition-colors"
      >
        {{ t('common.actions.save') }}
      </button>
    </div>
    
    <div class="space-y-2">
      <p class="text-sm text-gray-600 dark:text-gray-400">
        {{ t('modules.proxy.filter.suffix_description') }}
      </p>
      <div class="max-w-2xl">
        <textarea
          spellcheck="false"
          v-model="configValue.filterSuffix"
          class="w-full h-20 rounded-md border-gray-300 dark:border-gray-600 bg-white dark:bg-[#32324c] shadow-sm resize-none"
          :placeholder="t('modules.proxy.filter.suffix_placeholder')"
        ></textarea>
      </div>
      <p class="text-xs text-gray-500 dark:text-gray-400 mt-1">
        {{ t('modules.proxy.filter.suffix_hint') }}
      </p>
    </div>
  </div>
  
  <!-- 排除目标配置 -->
  <div class="bg-white dark:bg-[#282838] rounded-lg p-5 shadow-sm border border-gray-100 dark:border-gray-700 mt-6">
    <div class="flex items-center justify-between mb-4">
      <h3 class="text-lg font-medium">{{ t('modules.proxy.filter.exclude_targets') }}</h3>
      <button 
        @click="emit('openAddModal', 'exclude')"
        class="flex items-center px-3 py-1.5 bg-green-500 hover:bg-green-600 text-white rounded-md text-sm font-medium transition-colors"
      >
        <i class="bx bx-plus mr-1"></i>
        {{ t('common.actions.add') }}
      </button>
    </div>
    
    <div class="overflow-auto max-h-64 rounded-lg border border-gray-200 dark:border-gray-700">
      <table class="min-w-full divide-y divide-gray-200 dark:divide-gray-700 resizable-table">
        <thead class="bg-gray-50 dark:bg-gray-800 sticky top-0 z-10">
          <tr>
            <th class="px-3 py-2 text-left text-xs font-medium text-gray-500 dark:text-gray-400 uppercase tracking-wider w-20 resizable">
              {{ t('common.ui.regex') }}
            </th>
            <th class="px-3 py-2 text-left text-xs font-medium text-gray-500 dark:text-gray-400 uppercase tracking-wider w-20 resizable">
              {{ t('common.ui.enabled') }}
            </th>
            <th class="px-3 py-2 text-left text-xs font-medium text-gray-500 dark:text-gray-400 uppercase tracking-wider w-40 resizable">
              {{ t('modules.proxy.filter.prefix') }}
            </th>
            <th class="px-3 py-2 text-xs font-medium text-gray-500 dark:text-gray-400 uppercase tracking-wider w-24">
              {{ t('modules.proxy.filter.action') }}
            </th>
          </tr>
        </thead>
        <tbody class="bg-white dark:bg-[#32324a] divide-y divide-gray-200 dark:divide-gray-700">
          <tr 
            v-for="rule in configValue.exclude" 
            :key="rule.id"
            class="hover:bg-gray-50 dark:hover:bg-gray-700 transition-colors"
          >
            <td class="px-3 py-2 whitespace-nowrap">
              <div class="flex justify-center">
                <label class="inline-flex items-center">
                  <input 
                    spellcheck="false"
                    type="checkbox" 
                    :checked="rule.regexp" 
                    @change="(e) => emit('updateRuleToggle', rule, 'exclude', 'regexp', e)"
                    class="rounded border-gray-300 text-blue-500 focus:ring-blue-500"
                  />
                </label>
              </div>
            </td>
            <td class="px-3 py-2 whitespace-nowrap">
              <div class="flex justify-center">
                <label class="inline-flex items-center">
                  <input 
                    spellcheck="false"
                    type="checkbox" 
                    :checked="rule.enabled" 
                    @change="(e) => emit('updateRuleToggle', rule, 'exclude', 'enabled', e)"
                    class="rounded border-gray-300 text-green-500 focus:ring-green-500"
                  />
                </label>
              </div>
            </td>
            <td class="px-3 py-2 truncate max-w-xs">
              <div class="text-sm text-gray-900 dark:text-gray-200 truncate">{{ rule.prefix }}</div>
            </td>
            <td class="px-3 py-2 whitespace-nowrap text-center space-x-2">
              <button 
                @click="emit('openEditModal', rule, 'exclude')"
                class="text-blue-500 hover:text-blue-700 focus:outline-none"
              >
                <i class="bx bx-edit text-lg"></i>
              </button>
              <button 
                @click="emit('deleteRule', rule, 'exclude')"
                class="text-red-500 hover:text-red-700 focus:outline-none"
              >
                <i class="bx bx-trash text-lg"></i>
              </button>
            </td>
          </tr>
          <tr v-if="configValue.exclude.length === 0">
            <td colspan="4" class="px-6 py-4 text-center text-sm text-gray-500 dark:text-gray-400">
              {{ t('common.status.no_data') }}
            </td>
          </tr>
        </tbody>
      </table>
    </div>
  </div>
  
  <!-- 允许目标配置 -->
  <div class="bg-white dark:bg-[#282838] rounded-lg p-5 shadow-sm border border-gray-100 dark:border-gray-700 mt-6">
    <div class="flex items-center justify-between mb-4">
      <h3 class="text-lg font-medium">{{ t('modules.proxy.filter.include_targets') }}</h3>
      <button 
        @click="emit('openAddModal', 'include')"
        class="flex items-center px-3 py-1.5 bg-green-500 hover:bg-green-600 text-white rounded-md text-sm font-medium transition-colors"
      >
        <i class="bx bx-plus mr-1"></i>
        {{ t('common.actions.add') }}
      </button>
    </div>
    
    <div class="overflow-auto max-h-64 rounded-lg border border-gray-200 dark:border-gray-700">
      <table class="min-w-full divide-y divide-gray-200 dark:divide-gray-700 resizable-table">
        <thead class="bg-gray-50 dark:bg-gray-800 sticky top-0 z-10">
          <tr>
            <th class="px-3 py-2 text-left text-xs font-medium text-gray-500 dark:text-gray-400 uppercase tracking-wider w-20 resizable">
              {{ t('common.ui.regex') }}
            </th>
            <th class="px-3 py-2 text-left text-xs font-medium text-gray-500 dark:text-gray-400 uppercase tracking-wider w-20 resizable">
              {{ t('common.ui.enabled') }}
            </th>
            <th class="px-3 py-2 text-left text-xs font-medium text-gray-500 dark:text-gray-400 uppercase tracking-wider w-40 resizable">
              {{ t('modules.proxy.filter.prefix') }}
            </th>
            <th class="px-3 py-2 text-xs font-medium text-gray-500 dark:text-gray-400 uppercase tracking-wider w-24">
              {{ t('modules.proxy.filter.action') }}
            </th>
          </tr>
        </thead>
        <tbody class="bg-white dark:bg-[#32324a] divide-y divide-gray-200 dark:divide-gray-700">
          <tr 
            v-for="rule in configValue.include" 
            :key="rule.id"
            class="hover:bg-gray-50 dark:hover:bg-gray-700 transition-colors"
          >
            <td class="px-3 py-2 whitespace-nowrap">
              <div class="flex justify-center">
                <label class="inline-flex items-center">
                  <input 
                    spellcheck="false"
                    type="checkbox" 
                    :checked="rule.regexp" 
                    @change="(e) => emit('updateRuleToggle', rule, 'include', 'regexp', e)"
                    class="rounded border-gray-300 text-blue-500 focus:ring-blue-500"
                  />
                </label>
              </div>
            </td>
            <td class="px-3 py-2 whitespace-nowrap">
              <div class="flex justify-center">
                <label class="inline-flex items-center">
                  <input 
                    spellcheck="false"
                    type="checkbox" 
                    :checked="rule.enabled" 
                    @change="(e) => emit('updateRuleToggle', rule, 'include', 'enabled', e)"
                    class="rounded border-gray-300 text-green-500 focus:ring-green-500"
                  />
                </label>
              </div>
            </td>
            <td class="px-3 py-2 truncate max-w-xs">
              <div class="text-sm text-gray-900 dark:text-gray-200 truncate">{{ rule.prefix }}</div>
            </td>
            <td class="px-3 py-2 whitespace-nowrap text-center space-x-2">
              <button 
                @click="emit('openEditModal', rule, 'include')"
                class="text-blue-500 hover:text-blue-700 focus:outline-none"
              >
                <i class="bx bx-edit text-lg"></i>
              </button>
              <button 
                @click="emit('deleteRule', rule, 'include')"
                class="text-red-500 hover:text-red-700 focus:outline-none"
              >
                <i class="bx bx-trash text-lg"></i>
              </button>
            </td>
          </tr>
          <tr v-if="configValue.include.length === 0">
            <td colspan="4" class="px-6 py-4 text-center text-sm text-gray-500 dark:text-gray-400">
              {{ t('common.status.no_data') }}
            </td>
          </tr>
        </tbody>
      </table>
    </div>
  </div>
  
  <!-- HTTP 配置部分 -->
  <div v-if="completeConfig" class="bg-white dark:bg-[#282838] rounded-lg p-5 shadow-sm border border-gray-100 dark:border-gray-700 mt-6">
    <div class="flex items-center justify-between mb-4">
      <h3 class="text-lg font-medium">{{ t('settings.http_config') }}</h3>
      <button 
        @click="emit('updateHttpConfig')"
        class="px-3 py-1.5 bg-amber-500 hover:bg-amber-600 text-white rounded-md text-sm font-medium transition-colors"
      >
        {{ t('common.actions.save') }}
      </button>
    </div>
    
    <div class="grid grid-cols-1 md:grid-cols-2 gap-4">
      <div>
        <label class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">{{ t('settings.http_proxy') }}</label>
        <input 
          v-model="completeConfig.http.proxy" 
          type="text" 
          class="w-full rounded-md border-gray-300 dark:border-gray-600 bg-white dark:bg-[#32324c] shadow-sm"
          :placeholder="t('settings.http_proxy_placeholder')"
          spellcheck="false"
        />
        <p class="text-xs text-gray-500 dark:text-gray-400 mt-1">{{ t('settings.http_proxy_hint') }}</p>
      </div>
      
      <div>
        <label class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">{{ t('settings.http_timeout') }}</label>
        <input 
          spellcheck="false"
          v-model.number="completeConfig.http.timeout" 
          type="number" 
          class="w-full rounded-md border-gray-300 dark:border-gray-600 bg-white dark:bg-[#32324c] shadow-sm"
          min="1"
        />
        <p class="text-xs text-gray-500 dark:text-gray-400 mt-1">{{ t('settings.http_timeout_hint') }}</p>
      </div>
      
      <div>
        <label class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">{{ t('settings.max_connections_per_host') }}</label>
        <input 
          v-model.number="completeConfig.http.maxConnsPerHost" 
          type="number" 
          class="w-full rounded-md border-gray-300 dark:border-gray-600 bg-white dark:bg-[#32324c] shadow-sm"
          min="1"
        />
      </div>
      
      <div>
        <label class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">{{ t('settings.retry_times') }}</label>
        <input 
          spellcheck="false"
          v-model.number="completeConfig.http.retryTimes" 
          type="number" 
          class="w-full rounded-md border-gray-300 dark:border-gray-600 bg-white dark:bg-[#32324c] shadow-sm"
          min="0"
        />
        <p class="text-xs text-gray-500 dark:text-gray-400 mt-1">{{ t('settings.retry_times_hint') }}</p>
      </div>
      
      <div>
        <label class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">{{ t('settings.allow_redirect') }}</label>
        <input 
          spellcheck="false"
          v-model.number="completeConfig.http.allowRedirect" 
          type="number" 
          class="w-full rounded-md border-gray-300 dark:border-gray-600 bg-white dark:bg-[#32324c] shadow-sm"
          min="0"
        />
        <p class="text-xs text-gray-500 dark:text-gray-400 mt-1">{{ t('settings.allow_redirect_hint') }}</p>
      </div>
      
      <div>
        <label class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">{{ t('settings.max_qps') }}</label>
        <input 
          spellcheck="false"
          v-model.number="completeConfig.http.maxQps" 
          type="number" 
          class="w-full rounded-md border-gray-300 dark:border-gray-600 bg-white dark:bg-[#32324c] shadow-sm"
          min="1"
        />
      </div>
      
      <div class="flex items-center space-x-2">
        <input
          id="verify-ssl"
          v-model="completeConfig.http.verifySSL"
          type="checkbox"
          class="rounded border-gray-300 text-blue-500 focus:ring-blue-500"
        />
        <label for="verify-ssl" class="text-sm font-medium text-gray-700 dark:text-gray-300">
          {{ t('settings.verify_ssl') }}
        </label>
      </div>
      
      <div class="flex items-center space-x-2">
        <input
          id="force-http1"
          v-model="completeConfig.http.forceHTTP1"
          type="checkbox"
          class="rounded border-gray-300 text-blue-500 focus:ring-blue-500"
        />
        <label for="force-http1" class="text-sm font-medium text-gray-700 dark:text-gray-300">
          {{ t('settings.force_http1') }}
        </label>
      </div>
    </div>
  </div>
</template>

<style scoped>
/* 美化滚动条 */
::-webkit-scrollbar {
  width: 6px;
  height: 6px;
}

::-webkit-scrollbar-track {
  background: transparent;
}

::-webkit-scrollbar-thumb {
  background-color: rgba(156, 163, 175, 0.5);
  border-radius: 3px;
}

::-webkit-scrollbar-thumb:hover {
  background-color: rgba(156, 163, 175, 0.7);
}

/* 确保表格不会因为内容过多而拉伸 */
table {
  table-layout: fixed;
  width: 100%;
}

/* 自定义单选框和复选框样式 */
input[type="checkbox"] {
  color-scheme: light dark;
  cursor: pointer;
}

/* 确保暗色模式下勾选框可见 */
.dark input[type="checkbox"] {
  background-color: #32324c;
  border-color: #4b5563;
}

/* 设置选中状态样式 */
input[type="checkbox"]:checked {
  background-image: url("data:image/svg+xml,%3csvg viewBox='0 0 16 16' fill='white' xmlns='http://www.w3.org/2000/svg'%3e%3cpath d='M5.707 7.293a1 1 0 0 0-1.414 1.414l2 2a1 1 0 0 0 1.414 0l4-4a1 1 0 0 0-1.414-1.414L7 8.586 5.707 7.293z'/%3e%3c/svg%3e");
  background-size: 100% 100%;
  background-position: center;
  background-repeat: no-repeat;
}

input.text-green-500:checked {
  background-color: #10b981;
  border-color: #10b981;
}

input.text-blue-500:checked {
  background-color: #3b82f6;
  border-color: #3b82f6;
}

input:focus-visible {
  outline: none;
}

/* 可调整列宽样式 */
.resizable {
  position: relative;
  overflow: visible !important;
}

.column-resizer {
  position: absolute;
  top: 0;
  right: 0;
  width: 5px;
  height: 100%;
  cursor: col-resize;
  user-select: none;
  background-color: transparent;
  z-index: 1;
}

.column-resizer:hover,
.column-resizer:active {
  background-color: rgba(59, 130, 246, 0.5);
}

/* 拖动调整列宽时的鼠标样式 */
body.column-resizing {
  cursor: col-resize !important;
  user-select: none;
}

/* 截断过长文本 */
.truncate {
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}
</style> 