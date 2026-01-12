<script setup lang="ts">
import { useI18n } from 'vue-i18n';

const { t } = useI18n();

defineProps<{
  moduleId: string;
}>();
</script>

<template>
  <div class="flex items-center justify-center h-full">
    <div class="text-center">
      <i class="bx bx-wrench text-5xl text-gray-400 mb-2"></i>
      <p class="text-lg">{{ t('modules.plugins.in_development') }}</p>
      <p class="text-sm text-gray-500 mt-2">{{ t('modules.plugins.coming_soon') }}</p>
    </div>
  </div>
</template> 