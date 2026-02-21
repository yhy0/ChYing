<script setup lang="ts">
import { ref } from 'vue';
import { useI18n } from 'vue-i18n';
// @ts-ignore
import { PredictionApi } from "../../../../bindings/github.com/yhy0/ChYing/app.js";

const { t } = useI18n();

// API 生成状态
const api = ref('');
const apiGenerator = ref([]);
const show = ref(false);
const error = ref('');

// 生成 API
function generateApi() {
  if (!api.value.trim()) {
    error.value = t('modules.plugins.api_generate.empty_input', '请输入 API 描述');
    return;
  }
  
  error.value = '';
  show.value = true;
  
  PredictionApi(api.value).then((res: any) => {
    apiGenerator.value = res;
    show.value = false;
  }).catch((err: any) => {
    console.error('API 生成失败:', err);
    error.value = t('modules.plugins.api_generate.generate_failed', 'API 生成失败') + ': ' + err;
    show.value = false;
  });
}

// 复制结果到剪贴板
function copyToClipboard(text: string) {
  navigator.clipboard.writeText(text).catch(err => {
    console.error('复制失败:', err);
  });
}
</script>

<template>
  <div class="api-generator h-full flex flex-col">
    <!-- 标题和说明 -->
    <div class="mb-4">
      <h2 class="text-lg font-semibold">{{ t('modules.plugins.api_generate.title', 'API 生成器') }}</h2>
    </div>
    
    <!-- 错误提示 -->
    <div v-if="error" class="mb-4 p-3 bg-red-100 text-red-700 dark:bg-red-900/30 dark:text-red-400 rounded-md">
      {{ error }}
    </div>
    
    <!-- 主内容区域：左右分栏 -->
    <div class="flex-1 flex gap-4 min-h-0">
      <!-- 左侧输入面板 -->
      <div class="w-1/2 flex flex-col border border-gray-200 dark:border-gray-700 rounded-md overflow-hidden">
        <div class="p-3 bg-gray-50 dark:bg-gray-800 border-b border-gray-200 dark:border-gray-700">
          <h3 class="font-medium text-gray-700 dark:text-gray-300">
            {{ t('modules.plugins.api_generate.input', '输入 API 描述') }}
          </h3>
        </div>
        <div class="flex-1 p-3 overflow-hidden">
          <textarea
            spellcheck="false"
            v-model="api"
            class="w-full h-full p-3 border border-gray-200 dark:border-gray-700 rounded-md bg-white dark:bg-gray-900 text-gray-800 dark:text-gray-200 resize-none focus:ring-1 focus:ring-indigo-500 focus:border-indigo-500 outline-none"
            :placeholder="t('modules.plugins.api_generate.input_placeholder', '请输入您需要的 API 功能描述...')"
          ></textarea>
        </div>
        <div class="p-3 bg-gray-50 dark:bg-gray-800 border-t border-gray-200 dark:border-gray-700">
          <button
            @click="generateApi"
            class="w-full px-4 py-2 bg-indigo-600 hover:bg-indigo-700 text-white font-medium rounded-md transition-colors duration-200 flex items-center justify-center"
            :disabled="show"
          >
            <template v-if="show">
              <svg class="animate-spin -ml-1 mr-2 h-4 w-4 text-white" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24">
                <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle>
                <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
              </svg>
              {{ t('modules.plugins.api_generate.generating', '生成中...') }}
            </template>
            <template v-else>
              <i class="bx bx-code-alt mr-1.5"></i>
              {{ t('modules.plugins.api_generate.generate', '生成 API') }}
            </template>
          </button>
        </div>
      </div>
      
      <!-- 右侧输出面板 -->
      <div class="w-1/2 flex flex-col border border-gray-200 dark:border-gray-700 rounded-md overflow-hidden">
        <div class="p-3 bg-gray-50 dark:bg-gray-800 border-b border-gray-200 dark:border-gray-700 flex justify-between items-center">
          <h3 class="font-medium text-gray-700 dark:text-gray-300">
            {{ t('modules.plugins.api_generate.output', '生成结果') }}
          </h3>
          <button 
            v-if="apiGenerator.length > 0"
            @click="copyToClipboard(JSON.stringify(apiGenerator, null, 2))" 
            class="text-sm text-indigo-600 hover:text-indigo-800 dark:text-indigo-400 dark:hover:text-indigo-300 flex items-center"
          >
            <i class="bx bx-copy mr-1"></i>
            {{ t('modules.plugins.api_generate.copy', '复制') }}
          </button>
        </div>
        <div class="flex-1 p-3 overflow-auto bg-gray-50 dark:bg-gray-900">
          <div v-if="show" class="h-full flex items-center justify-center">
            <div class="text-center">
              <div class="inline-block animate-spin rounded-full h-8 w-8 border-t-2 border-b-2 border-indigo-500 mb-2"></div>
              <p class="text-gray-500 dark:text-gray-400">{{ t('modules.plugins.api_generate.loading', '正在生成 API...') }}</p>
            </div>
          </div>
          <div v-else-if="apiGenerator.length === 0" class="h-full flex items-center justify-center">
            <div class="text-center">
              <i class="bx bx-code-curly text-5xl text-gray-300 dark:text-gray-600 mb-2"></i>
              <p class="text-gray-500 dark:text-gray-400">{{ t('modules.plugins.api_generate.no_result', '暂无生成结果') }}</p>
            </div>
          </div>
          <pre v-else class="w-full h-full p-3 rounded-md bg-white dark:bg-gray-800 text-gray-800 dark:text-gray-200 overflow-auto border border-gray-200 dark:border-gray-700">{{ JSON.stringify(apiGenerator, null, 2) }}</pre>
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
/* 确保组件占据整个高度 */
.api-generator {
  min-height: 0;
}

/* 滚动条样式已移至 scrollbar.css 统一管理 */

/* 动画 */
@keyframes pulse {
  0%, 100% {
    opacity: 1;
  }
  50% {
    opacity: 0.7;
  }
}

.animate-pulse {
  animation: pulse 2s cubic-bezier(0.4, 0, 0.6, 1) infinite;
}
</style> 