<script setup lang="ts">
import { computed } from 'vue';

interface UrlItem {
  url: string;
  type?: string;
  source?: string;
  confidence?: string;
}

interface SecretItem {
  type: string;
  value: string;
  confidence?: string;
  source?: string;
}

interface JsluiceResult {
  urls: UrlItem[];
  secrets: SecretItem[];
  loading: boolean;
  error?: string | null;
}

const props = defineProps<{
  panelWidth: number;
  results: JsluiceResult;
}>();

const emit = defineEmits<{
  (e: 'close'): void;
  (e: 'start-resize-panel', panel: string, event: MouseEvent): void;
}>();

const closePanel = () => {
  emit('close');
};

const startResize = (event: MouseEvent) => {
  emit('start-resize-panel', 'extractor', event);
};

const panelStyle = computed(() => ({
  width: `${props.panelWidth}px`,
}));

</script>

<template>
  <div 
    class="extractor-panel flex flex-col border-l border-gray-200 dark:border-gray-700 bg-white dark:bg-gray-800 shadow-sm"
    :style="panelStyle"
  >
    <div 
      class="panel-resize-handle absolute top-0 bottom-0 left-0 w-1.5 cursor-ew-resize z-10"
      @mousedown="startResize"
    ></div>
    
    <div class="flex items-center justify-between p-2 border-b border-gray-200 dark:border-gray-700">
      <h3 class="text-sm font-semibold text-gray-700 dark:text-gray-300">Data Extractor</h3>
      <button 
        @click="closePanel" 
        class="p-1 rounded-md hover:bg-gray-100 dark:hover:bg-gray-700 text-gray-500 dark:text-gray-400"
        title="Close Panel"
      >
        <i class="bx bx-x text-lg"></i>
      </button>
    </div>

    <div class="flex-1 p-3 overflow-y-auto text-xs">
      <div v-if="props.results.loading" class="text-center py-4">
        <p class="text-gray-500 dark:text-gray-400">Loading extracted data...</p>
      </div>
      <div v-else-if="props.results.error" class="text-center py-4">
        <p class="text-red-500 dark:text-red-400">Error: {{ props.results.error }}</p>
      </div>
      <div v-else>
        <div v-if="props.results.urls.length > 0" class="mb-4">
          <h4 class="font-semibold text-gray-600 dark:text-gray-300 mb-1.5">URLs ({{ props.results.urls.length }})</h4>
          <ul class="space-y-1">
            <li 
              v-for="(url, index) in props.results.urls" 
              :key="'url-' + index" 
              class="p-1.5 bg-gray-50 dark:bg-gray-700 rounded break-all hover:bg-gray-100 dark:hover:bg-gray-600"
            >
              <a :href="url.url" target="_blank" class="text-blue-600 dark:text-blue-400 hover:underline">{{ url.url }}</a>
              <div v-if="url.type" class="text-gray-500 dark:text-gray-400 text-xxs">Type: {{ url.type }}</div>
              <div v-if="url.source" class="text-gray-500 dark:text-gray-400 text-xxs">Source: {{ url.source }}</div>
            </li>
          </ul>
        </div>
        <div v-else class="mb-3 text-gray-500 dark:text-gray-400">
          No URLs found.
        </div>

        <div v-if="props.results.secrets.length > 0">
          <h4 class="font-semibold text-gray-600 dark:text-gray-300 mb-1.5">Secrets ({{ props.results.secrets.length }})</h4>
          <ul class="space-y-1">
            <li 
              v-for="(secret, index) in props.results.secrets" 
              :key="'secret-' + index" 
              class="p-1.5 bg-gray-50 dark:bg-gray-700 rounded break-all"
            >
              <strong class="text-orange-600 dark:text-orange-400">{{ secret.type }}:</strong>
              <span class="ml-1 text-gray-700 dark:text-gray-300">{{ secret.value }}</span>
              <div v-if="secret.source" class="text-gray-500 dark:text-gray-400 text-xxs">Source: {{ secret.source }}</div>
            </li>
          </ul>
        </div>
        <div v-else class="text-gray-500 dark:text-gray-400">
          No secrets found.
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
.extractor-panel {
  position: relative; /* For resize handle positioning */
}
.panel-resize-handle {
  /* Ensure it's visually distinct for debugging if needed, but typically subtle */
  /* background-color: rgba(0,0,0,0.1); */ 
}
.text-xxs {
  font-size: 0.65rem; /* Smaller text for metadata */
  line-height: 0.85rem;
}
</style> 