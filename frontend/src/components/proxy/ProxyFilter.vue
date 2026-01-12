<script setup lang="ts">
import { ref, watch } from 'vue';
import { useI18n } from 'vue-i18n';

const { t } = useI18n();

export interface FilterOptions {
  method: string;
  host: string;
  path: string;
  status: string;
  contentType: string;
}

const props = defineProps<{
  initialFilter?: Partial<FilterOptions>;
}>();

const emit = defineEmits<{
  (e: 'filter', filterOptions: FilterOptions): void;
  (e: 'reset'): void;
}>();

const filterOptions = ref<FilterOptions>({
  method: props.initialFilter?.method || '',
  host: props.initialFilter?.host || '',
  path: props.initialFilter?.path || '',
  status: props.initialFilter?.status || '',
  contentType: props.initialFilter?.contentType || '',
});

const methodOptions = ['', 'GET', 'POST', 'PUT', 'DELETE', 'PATCH', 'OPTIONS', 'HEAD'];

const applyFilter = () => {
  emit('filter', filterOptions.value);
};

const resetFilter = () => {
  filterOptions.value = {
    method: '',
    host: '',
    path: '',
    status: '',
    contentType: '',
  };
  emit('reset');
};

// Watch for changes to immediately apply filter if any field is changed
watch(filterOptions, () => {
  applyFilter();
}, { deep: true });
</script>

<template>
  <div class="proxy-filter">
    <div class="proxy-filter-title">
      {{ t('modules.proxy.filter.title') }}
    </div>
    <div class="grid grid-cols-1 md:grid-cols-5 gap-3">
      <!-- Method filter -->
      <div>
        <label class="block text-xs font-medium text-gray-700 dark:text-gray-300 mb-1">
          {{ t('modules.proxy.filter.method') }}
        </label>
        <select
          v-model="filterOptions.method"
          class="proxy-filter-input"
        >
          <option value="">{{ t('common.ui.none') }}</option>
          <option v-for="method in methodOptions.slice(1)" :key="method" :value="method">
            {{ method }}
          </option>
        </select>
      </div>

      <!-- Host filter -->
      <div>
        <label class="block text-xs font-medium text-gray-700 dark:text-gray-300 mb-1">
          {{ t('modules.proxy.filter.host') }}
        </label>
        <input
          v-model="filterOptions.host"
          type="text"
          class="proxy-filter-input"
          :placeholder="t('modules.proxy.filter.placeholder')"
        />
      </div>

      <!-- Path filter -->
      <div>
        <label class="block text-xs font-medium text-gray-700 dark:text-gray-300 mb-1">
          {{ t('modules.proxy.filter.path') }}
        </label>
        <input
          v-model="filterOptions.path"
          type="text"
          class="proxy-filter-input"
          :placeholder="t('modules.proxy.filter.placeholder')"
        />
      </div>

      <!-- Status filter -->
      <div>
        <label class="block text-xs font-medium text-gray-700 dark:text-gray-300 mb-1">
          {{ t('modules.proxy.filter.status') }}
        </label>
        <input
          v-model="filterOptions.status"
          type="text"
          class="proxy-filter-input"
          :placeholder="t('modules.proxy.filter.placeholder')"
        />
      </div>

      <!-- Content Type filter -->
      <div>
        <label class="block text-xs font-medium text-gray-700 dark:text-gray-300 mb-1">
          {{ t('modules.proxy.filter.contentType') }}
        </label>
        <input
          v-model="filterOptions.contentType"
          type="text"
          class="proxy-filter-input"
          :placeholder="t('modules.proxy.filter.placeholder')"
        />
      </div>
    </div>

    <div class="mt-3 flex items-center justify-end">
      <button
        @click="resetFilter"
        class="btn btn-secondary btn-sm"
      >
        <i class="bx bx-reset"></i>
        {{ t('modules.proxy.filter.reset') }}
      </button>
    </div>
  </div>
</template> 