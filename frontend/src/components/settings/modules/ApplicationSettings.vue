<script setup lang="ts">
import { ref } from 'vue';
import { useI18n } from 'vue-i18n';

const { t } = useI18n();

// 应用设置
const autoLoadLastProject = ref(true);
const autoSaveInterval = ref(5);
const defaultProjectPath = ref('/Users/Documents/Projects');

// 处理文件路径浏览
const handleBrowse = () => {
  // TODO: 实际应用中应该调用系统文件选择器
};
</script>

<template>
  <div class="space-y-6 max-w-3xl">
    <h2 class="text-lg font-medium mb-5 flex items-center">
      <i class="bx bx-cog text-blue-500 mr-2"></i>
      {{ t('settings.application_behavior') }}
    </h2>
    
    <div class="space-y-5">
      <!-- 启动时自动加载上次项目 -->
      <div class="bg-white dark:bg-[#282838] rounded-lg p-4 shadow-sm border border-gray-100 dark:border-gray-700">
        <div class="flex items-center">
          <input 
            id="autoLoadProject" 
            v-model="autoLoadLastProject" 
            type="checkbox" 
            class="w-4 h-4 text-blue-600 border-gray-300 rounded focus:ring-blue-500 dark:focus:ring-blue-600 dark:ring-offset-gray-800 focus:ring-2 dark:bg-gray-700 dark:border-gray-600"
          >
          <label for="autoLoadProject" class="ml-2 text-sm font-medium text-gray-900 dark:text-gray-300">
            {{ t('settings.auto_load_project') }}
          </label>
        </div>
        
        <p class="text-xs text-gray-500 dark:text-gray-400 mt-2 ml-6">
          {{ t('settings.auto_load_description') }}
        </p>
      </div>
      
      <!-- 自动保存间隔 -->
      <div class="bg-white dark:bg-[#282838] rounded-lg p-4 shadow-sm border border-gray-100 dark:border-gray-700">
        <label class="block mb-2 text-sm font-medium text-gray-900 dark:text-gray-300">
          {{ t('settings.auto_save_interval') }}
        </label>
        <div class="flex items-center">
          <input 
            v-model="autoSaveInterval" 
            type="number" 
            min="1" 
            max="60"
            class="w-20 bg-gray-50 border border-gray-300 text-gray-900 text-sm rounded-lg focus:ring-blue-500 focus:border-blue-500 p-2 dark:bg-gray-700 dark:border-gray-600 dark:placeholder-gray-400 dark:text-white dark:focus:ring-blue-500 dark:focus:border-blue-500"
          >
          <span class="ml-2 text-sm text-gray-700 dark:text-gray-300">{{ t('common.ui.minutes') }}</span>
        </div>
        <p class="text-xs text-gray-500 dark:text-gray-400 mt-2">
          {{ t('settings.auto_save_description') }}
        </p>
      </div>
      
      <!-- 默认项目位置 -->
      <div class="bg-white dark:bg-[#282838] rounded-lg p-4 shadow-sm border border-gray-100 dark:border-gray-700">
        <label class="block mb-2 text-sm font-medium text-gray-900 dark:text-gray-300">
          {{ t('settings.default_project_location') }}
        </label>
        <div class="flex">
          <input 
            v-model="defaultProjectPath" 
            type="text" 
            class="flex-1 bg-gray-50 border border-gray-300 text-gray-900 text-sm rounded-lg rounded-r-none focus:ring-blue-500 focus:border-blue-500 p-2 dark:bg-gray-700 dark:border-gray-600 dark:placeholder-gray-400 dark:text-white dark:focus:ring-blue-500 dark:focus:border-blue-500"
            spellcheck="false"
          >
          <button 
            @click="handleBrowse" 
            class="btn btn-secondary rounded-l-none"
          >
            {{ t('common.actions.browse') }}
          </button>
        </div>
        <p class="text-xs text-gray-500 dark:text-gray-400 mt-2">
          {{ t('settings.default_location_description') }}
        </p>
      </div>
    </div>
  </div>
</template> 