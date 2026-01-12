<script setup lang="ts">
import { useI18n } from 'vue-i18n';

const { t } = useI18n();
</script>

<template>
  <div class="space-y-6 max-w-3xl">
    <h2 class="text-lg font-medium mb-5 flex items-center">
      <i class="bx bx-info-circle text-indigo-500 mr-2"></i>
      {{ t('settings.about') }}
    </h2>
    
    <div class="bg-white dark:bg-[#282838] rounded-lg p-4 shadow-sm border border-gray-100 dark:border-gray-700">
      <h3 class="font-medium mb-1">{{ t('layout.app.title') }}</h3>
      <p class="text-sm text-gray-500 dark:text-gray-400">Version 2.0</p>
      <div class="flex items-center mt-3 text-sm text-gray-500 dark:text-gray-400">
        <i class="bx bx-calendar mr-2"></i>
        <span>Build Date: 2026-01-12</span>
      </div>
      <p class="text-sm text-gray-500 dark:text-gray-400 mt-2">
        A security testing application built with Vue 3, TypeScript, and TailwindCSS.
      </p>
      <div class="mt-4 flex space-x-2">
        <button class="btn btn-secondary btn-sm">
          <i class="bx bx-file mr-1"></i>
          {{ t('settings.check_updates') }}
        </button>
        <button class="btn btn-secondary btn-sm">
          <i class="bx bx-bug mr-1"></i>
          {{ t('settings.report_bug') }}
        </button>
      </div>
    </div>
  </div>
</template> 