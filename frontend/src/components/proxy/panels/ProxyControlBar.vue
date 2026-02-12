<script setup lang="ts">
import { ref, onMounted} from 'vue';
import { useI18n } from 'vue-i18n';
import ProxyFilter from '../ProxyFilter.vue';
import ProxyDSLSearch from './ProxyDSLSearch.vue';
import type { FilterOptions } from '../ProxyFilter.vue';
import type { ProxyHistoryItem } from '../../../store';
// @ts-ignore
import {
  GetConfig
} from "../../../../bindings/github.com/yhy0/ChYing/app.js";

const { t } = useI18n();

const proxyListener = ref('');

onMounted(() => {
  // 获取默认的爆破字典文件路径
  GetConfig().then((result: any) => {
    console.log('获取配置:', result);
    try {
      // 解析JSON字符串
      const config = JSON.parse(result);
      // 提取字段
      proxyListener.value = config.proxy;
    } catch (error) {
      console.error('解析配置失败:', error);
    }
  }).catch((err: any) => {
    console.error('获取proxyListener失败:', err);
  });
});

const props = defineProps<{
  interceptEnabled: boolean;
  proxyHistory: ProxyHistoryItem[];
}>();

const emit = defineEmits<{
  (e: 'toggle-intercept'): void;
  (e: 'filter', options: FilterOptions): void;
  (e: 'reset-filter'): void;
  (e: 'clear'): void;
  (e: 'send-to-repeater'): void;
  (e: 'send-to-intruder'): void;
  // 新增DSL搜索相关事件
  (e: 'search-results', results: any[]): void;
  (e: 'clear-search'): void;
  (e: 'notify', notification: { message: string; type: 'success' | 'error' }): void;
}>();

// Filter options and modal
const showFilterModal = ref(false);

// 过滤器变更
const handleFilterChange = (options: FilterOptions) => {
  emit('filter', options);
};

// 重置过滤器
const handleResetFilter = () => {
  emit('reset-filter');
};

// 拦截控制
const toggleInterception = () => {
  emit('toggle-intercept');
};

// 发送到Repeater
const sendToRepeater = () => {
  emit('send-to-repeater');
};

// 发送到Intruder
const sendToIntruder = () => {
  emit('send-to-intruder');
};

// 清空历史记录
const clearHistory = () => {
  emit('clear');
};

// DSL搜索结果处理
const handleDSLSearchResults = (results: any[]) => {
  emit('search-results', results);
};

// 清除DSL搜索
const clearDSLSearch = () => {
  emit('clear-search');
};

// 处理通知
const handleNotify = (notification: { message: string; type: 'success' | 'error' }) => {
  emit('notify', notification);
};
</script>

<template>
  <div class="p-1 border-b border-gray-200 dark:border-gray-700 flex items-center bg-white dark:bg-[#1e1e36] shadow-sm min-h-[2.5rem]">
    <div class="flex items-center space-x-1.5">
      <!-- 拦截按钮 -->
      <button
        :class="[
          'btn btn-xs px-1.5 py-0.5 rounded-md flex items-center gap-1',
          props.interceptEnabled 
            ? 'btn-danger bg-red-600 text-white hover:bg-red-700'
            : 'btn-secondary'
        ]"
        @click="toggleInterception"
      >
        <i :class="['bx text-xs', props.interceptEnabled ? 'bx-pause' : 'bx-play']"></i>
        <span class="text-xs">{{ props.interceptEnabled ? t('modules.proxy.controls.intercept_on') : t('modules.proxy.controls.intercept_off') }}</span>
      </button>
      
      <!-- 过滤按钮 -->
      <button 
        class="btn btn-secondary btn-xs px-1.5 py-0.5 rounded-md flex items-center gap-1"
        @click="showFilterModal = !showFilterModal"
      >
        <i class="bx bx-filter text-xs"></i> 
        <span class="text-xs">Filter</span>
      </button>
      
      <!-- 清除按钮 -->
      <button 
        class="btn btn-secondary btn-xs px-1.5 py-0.5 rounded-md flex items-center gap-1"
        @click="clearHistory"
      >
        <i class="bx bx-trash text-xs"></i>
        <span class="text-xs">Clear</span>
      </button>
      
      <!-- 发送到... 按钮组 -->
      <div class="flex items-center px-1.5 rounded-md bg-gradient-to-r from-gray-50 to-gray-100 dark:from-[#292945] dark:to-[#2d2d4d] border border-gray-200 dark:border-gray-700">
        <i class="bx bx-send text-indigo-500 dark:text-indigo-400 mr-0.5 text-xs"></i>
        <span class="text-xs font-medium text-gray-700 dark:text-gray-300 mr-0.5">Send to:</span>
        <button 
          class="btn btn-xs px-1 py-0.5 bg-transparent hover:bg-indigo-50 dark:hover:bg-indigo-900/30 text-indigo-600 dark:text-indigo-400 text-xs font-medium rounded"
          @click="sendToRepeater"
          title="Send to Repeater (Ctrl+R)"
        >
          Repeater
        </button>
        <span class="text-gray-300 dark:text-gray-600 mx-0.5">|</span>
        <button 
          class="btn btn-xs px-1 py-0.5 bg-transparent hover:bg-orange-50 dark:hover:bg-orange-900/30 text-orange-600 dark:text-orange-400 text-xs font-medium rounded"
          @click="sendToIntruder"
          title="Send to Intruder (Ctrl+I)"
        >
          Intruder
        </button>
      </div>
    </div>
    
    <!-- DSL 搜索组件 -->
    <div class="flex-1 mr-2 ml-1.5 w-full max-w-[90%]">
      <ProxyDSLSearch 
        @search-results="handleDSLSearchResults"
        @clear-search="clearDSLSearch"
        @notify="handleNotify"
      />
    </div>
    
    <!-- 代理状态指示 -->
    <div class="ml-auto flex items-center">
      <span class="text-xs text-gray-500 dark:text-gray-400 mr-1">{{ t('modules.proxy.proxy_listener') }} {{ proxyListener }}</span>
      <div class="h-1.5 w-1.5 rounded-full bg-green-500 shadow-sm shadow-green-500/50 animate-pulse"></div>
    </div>
    
    <!-- Filter Modal -->
    <div v-show="showFilterModal" class="absolute inset-x-0 top-[5rem] z-20 bg-white dark:bg-[#1e1e36] shadow-lg border border-gray-200 dark:border-gray-700 rounded-md overflow-hidden">
      <div class="flex justify-end items-center px-3 py-0.5 border-b border-gray-200 dark:border-gray-700 bg-gray-50 dark:bg-[#292945]">
        <button 
          @click="showFilterModal = false" 
          class="btn btn-icon text-gray-500 hover:text-gray-700 dark:text-gray-400 dark:hover:text-gray-300 w-4 h-4"
        >
          <i class="bx bx-x"></i>
        </button>
      </div>
      <ProxyFilter 
        :proxyHistory="props.proxyHistory"
        @filter="handleFilterChange" 
        @reset="handleResetFilter" 
      />
    </div>
  </div>
</template> 