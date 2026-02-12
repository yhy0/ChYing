<script setup lang="ts">
import { ref, onMounted } from 'vue';
import { useI18n } from 'vue-i18n';
import { CheckForUpdates, GetCurrentVersion, OpenURL } from '../../../../bindings/github.com/yhy0/ChYing/app.js';

const { t } = useI18n();

const currentVersion = ref('');
const checking = ref(false);
const updateResult = ref<{
  hasUpdate: boolean;
  currentVersion: string;
  latestVersion: string;
  releaseUrl: string;
  releaseNotes: string;
  publishedAt: string;
} | null>(null);
const checkError = ref('');

onMounted(async () => {
  try {
    currentVersion.value = await GetCurrentVersion();
  } catch {
    currentVersion.value = '';
  }
});

const handleCheckUpdates = async () => {
  checking.value = true;
  updateResult.value = null;
  checkError.value = '';
  try {
    const result = await CheckForUpdates();
    if (result.error) {
      checkError.value = result.error;
    } else {
      updateResult.value = result.data as typeof updateResult.value;
    }
  } catch (e: any) {
    checkError.value = e?.message || String(e);
  } finally {
    checking.value = false;
  }
};

const handleOpenRelease = () => {
  if (updateResult.value?.releaseUrl) {
    OpenURL(updateResult.value.releaseUrl);
  }
};

const handleReportBug = () => {
  OpenURL('https://github.com/yhy0/CHYing/issues/new');
};
</script>

<template>
  <div class="space-y-6 max-w-3xl">
    <h2 class="text-lg font-medium mb-5 flex items-center">
      <i class="bx bx-info-circle text-indigo-500 mr-2"></i>
      {{ t('settings.about') }}
    </h2>
    
    <div class="bg-white dark:bg-[#282838] rounded-lg p-4 shadow-sm border border-gray-100 dark:border-gray-700">
      <h3 class="font-medium mb-1">{{ t('layout.app.title') }}</h3>
      <p class="text-sm text-gray-500 dark:text-gray-400">Version {{ currentVersion || '...' }}</p>
      <p class="text-sm text-gray-500 dark:text-gray-400 mt-2">
        A security testing application built with Vue 3, TypeScript, and TailwindCSS.
      </p>
      <div class="mt-4 flex space-x-2">
        <button 
          class="btn btn-secondary btn-sm"
          :disabled="checking"
          @click="handleCheckUpdates"
        >
          <i class="bx mr-1" :class="checking ? 'bx-loader-alt bx-spin' : 'bx-refresh'"></i>
          {{ checking ? t('settings.checking_updates') : t('settings.check_updates') }}
        </button>
        <button class="btn btn-secondary btn-sm" @click="handleReportBug">
          <i class="bx bx-bug mr-1"></i>
          {{ t('settings.report_bug') }}
        </button>
      </div>
    </div>

    <!-- 检查结果 -->
    <div v-if="updateResult" class="bg-white dark:bg-[#282838] rounded-lg p-4 shadow-sm border border-gray-100 dark:border-gray-700">
      <!-- 有新版本 -->
      <template v-if="updateResult.hasUpdate">
        <div class="flex items-center mb-3">
          <i class="bx bx-up-arrow-circle text-green-500 text-xl mr-2"></i>
          <span class="font-medium text-green-600 dark:text-green-400">{{ t('settings.update_available') }}</span>
        </div>
        <div class="text-sm space-y-1 text-gray-600 dark:text-gray-300">
          <p>{{ t('settings.current_version') }}: <span class="font-mono">{{ updateResult.currentVersion }}</span></p>
          <p>{{ t('settings.latest_version') }}: <span class="font-mono font-semibold text-green-600 dark:text-green-400">{{ updateResult.latestVersion }}</span></p>
          <p v-if="updateResult.publishedAt">{{ t('settings.published_at') }}: {{ new Date(updateResult.publishedAt).toLocaleDateString() }}</p>
        </div>
        <div v-if="updateResult.releaseNotes" class="mt-3">
          <p class="text-xs font-medium text-gray-500 dark:text-gray-400 mb-1">{{ t('settings.release_notes') }}:</p>
          <div class="text-xs text-gray-500 dark:text-gray-400 bg-gray-50 dark:bg-gray-800 rounded p-2 max-h-32 overflow-y-auto whitespace-pre-wrap">{{ updateResult.releaseNotes }}</div>
        </div>
        <button 
          class="mt-3 btn btn-sm bg-green-500 hover:bg-green-600 text-white border-0"
          @click="handleOpenRelease"
        >
          <i class="bx bx-download mr-1"></i>
          {{ t('settings.download_update') }}
        </button>
      </template>

      <!-- 已是最新 -->
      <template v-else>
        <div class="flex items-center">
          <i class="bx bx-check-circle text-blue-500 text-xl mr-2"></i>
          <span class="font-medium text-blue-600 dark:text-blue-400">{{ t('settings.already_latest') }}</span>
        </div>
        <p class="text-sm text-gray-500 dark:text-gray-400 mt-1">
          {{ t('settings.already_latest_msg', { version: updateResult.currentVersion }) }}
        </p>
      </template>
    </div>

    <!-- 错误信息 -->
    <div v-if="checkError" class="bg-red-50 dark:bg-red-900/20 rounded-lg p-4 border border-red-200 dark:border-red-800">
      <div class="flex items-center">
        <i class="bx bx-error-circle text-red-500 text-xl mr-2"></i>
        <span class="font-medium text-red-600 dark:text-red-400">{{ t('settings.update_check_failed') }}</span>
      </div>
      <p class="text-sm text-red-500 dark:text-red-400 mt-1">{{ checkError }}</p>
    </div>
  </div>
</template>
