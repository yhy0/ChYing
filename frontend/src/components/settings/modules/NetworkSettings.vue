<script setup lang="ts">
import { useI18n } from 'vue-i18n';

const { t } = useI18n();
</script>

<template>
  <div class="space-y-6 max-w-3xl">
    <h2 class="text-lg font-medium mb-5 flex items-center">
      <i class="bx bx-globe text-cyan-500 mr-2"></i>
      {{ t('settings.network') }}
    </h2>
    
    <div class="bg-white dark:bg-[#282838] rounded-lg p-4 shadow-sm border border-gray-100 dark:border-gray-700">
      <div class="bg-blue-50 dark:bg-blue-900/20 border border-blue-200 dark:border-blue-800 rounded-lg p-4 text-sm text-blue-700 dark:text-blue-300">
        <div class="flex items-start">
          <i class="bx bx-info-circle mt-0.5 mr-2 text-lg"></i>
          <p>{{ t('settings.network_coming_soon') }}</p>
        </div>
      </div>
    </div>
  </div>
</template> 