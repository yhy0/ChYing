<script setup lang="ts">

// 定义props
const props = defineProps<{
  panelWidth: number; 
  expandedSections: Record<string, boolean>;
}>();

// 定义事件
const emit = defineEmits<{
  (e: 'close'): void;
  (e: 'toggleSection', section: string): void;
  (e: 'startResizePanel', event: MouseEvent, panel: string): void;
}>();

// 关闭面板
const closePanel = () => {
  emit('close');
};

// 切换部分展开/折叠
const toggleSection = (section: string) => {
  emit('toggleSection', section);
};

// 开始调整面板大小
const startResizePanel = (e: MouseEvent) => {
  emit('startResizePanel', e, 'inspector');
};
</script>

<template>
  <div 
    class="inspector-panel flex flex-col border-l border-gray-200 dark:border-gray-700 bg-white dark:bg-gray-800 shadow-sm"
    :style="{ width: props.panelWidth + 'px' }"
  >
    <!-- 拖动调整大小的手柄 -->
    <div 
      class="resize-handle absolute top-0 left-0 bottom-0 w-3 bg-transparent cursor-ew-resize z-10 flex items-center justify-center hover:bg-gray-200/50 dark:hover:bg-gray-700/50 opacity-0 hover:opacity-100 transition-opacity duration-200"
      @mousedown="startResizePanel"
    >
      <div class="h-12 w-0.5 bg-gray-300/80 dark:bg-gray-600/80 rounded-full"></div>
    </div>

    <!-- Inspector Header -->
    <div class="flex items-center justify-between p-3 border-b border-gray-200 dark:border-gray-700 bg-gray-50 dark:bg-gray-800">
      <h3 class="text-sm font-medium">Inspector</h3>
      <button 
        @click="closePanel"
        class="p-1 rounded-full hover:bg-gray-200 dark:hover:bg-gray-700 transition-colors"
        title="Close"
      >
        <i class="bx bx-x text-lg text-gray-600 dark:text-gray-400"></i>
      </button>
    </div>
    
    <!-- Inspector Content -->
    <div class="flex-1 overflow-auto">
      <!-- Request Attributes Section -->
      <div class="border-b border-gray-200 dark:border-gray-700">
        <button 
          class="flex items-center justify-between w-full p-3 text-left hover:bg-gray-100 dark:hover:bg-gray-750 transition-colors"
          @click="toggleSection('requestAttributes')"
        >
          <div class="flex items-center">
            <i class="bx bx-cube-alt text-lg text-gray-500 dark:text-gray-400 mr-2"></i>
            <span class="text-sm">Request Attributes</span>
          </div>
          <span class="text-xs bg-gray-100 dark:bg-gray-700 text-gray-600 dark:text-gray-400 rounded-full px-2 py-0.5">2</span>
        </button>
        <div v-if="props.expandedSections.requestAttributes" class="p-3 border-t border-gray-100 dark:border-gray-700 bg-gray-50/50 dark:bg-gray-800/50">
          <div class="space-y-2">
            <div class="grid grid-cols-3 text-sm">
              <span class="text-gray-500 dark:text-gray-400">Method</span>
              <span class="col-span-2 font-mono">POST</span>
            </div>
            <div class="grid grid-cols-3 text-sm">
              <span class="text-gray-500 dark:text-gray-400">URL</span>
              <span class="col-span-2 font-mono break-all">/gen_204?s=web&t=cap&typ=csi&ei=rV_qZ_LXMNXg1e8P2trq8AA&rt=</span>
            </div>
          </div>
        </div>
      </div>
      
      <!-- Request Query Parameters Section -->
      <div class="border-b border-gray-200 dark:border-gray-700">
        <button 
          class="flex items-center justify-between w-full p-3 text-left hover:bg-gray-100 dark:hover:bg-gray-750 transition-colors"
          @click="toggleSection('queryParameters')"
        >
          <div class="flex items-center">
            <i class="bx bx-question-mark text-lg text-gray-500 dark:text-gray-400 mr-2"></i>
            <span class="text-sm">Query Parameters</span>
          </div>
          <span class="text-xs bg-gray-100 dark:bg-gray-700 text-gray-600 dark:text-gray-400 rounded-full px-2 py-0.5">10</span>
        </button>
        <div v-if="props.expandedSections.queryParameters" class="p-3 border-t border-gray-100 dark:border-gray-700 bg-gray-50/50 dark:bg-gray-800/50">
          <div class="space-y-2">
            <div class="grid grid-cols-3 text-sm">
              <span class="text-gray-500 dark:text-gray-400">s</span>
              <span class="col-span-2 font-mono">web</span>
            </div>
            <div class="grid grid-cols-3 text-sm">
              <span class="text-gray-500 dark:text-gray-400">t</span>
              <span class="col-span-2 font-mono">cap</span>
            </div>
            <div class="grid grid-cols-3 text-sm">
              <span class="text-gray-500 dark:text-gray-400">typ</span>
              <span class="col-span-2 font-mono">csi</span>
            </div>
            <div class="grid grid-cols-3 text-sm">
              <span class="text-gray-500 dark:text-gray-400">ei</span>
              <span class="col-span-2 font-mono">rV_qZ_LXMNXg1e8P2trq8AA</span>
            </div>
            <div class="grid grid-cols-3 text-sm">
              <span class="text-gray-500 dark:text-gray-400">rt</span>
              <span class="col-span-2 font-mono">''</span>
            </div>
          </div>
        </div>
      </div>
      
      <!-- Request Cookies Section -->
      <div class="border-b border-gray-200 dark:border-gray-700">
        <button 
          class="flex items-center justify-between w-full p-3 text-left hover:bg-gray-100 dark:hover:bg-gray-750 transition-colors"
          @click="toggleSection('cookies')"
        >
          <div class="flex items-center">
            <i class="bx bx-cookie text-lg text-gray-500 dark:text-gray-400 mr-2"></i>
            <span class="text-sm">Cookies</span>
          </div>
          <span class="text-xs bg-gray-100 dark:bg-gray-700 text-gray-600 dark:text-gray-400 rounded-full px-2 py-0.5">2</span>
        </button>
        <div v-if="props.expandedSections.cookies" class="p-3 border-t border-gray-100 dark:border-gray-700 bg-gray-50/50 dark:bg-gray-800/50">
          <div class="space-y-2">
            <div class="grid grid-cols-3 text-sm">
              <span class="text-gray-500 dark:text-gray-400">AEC</span>
              <span class="col-span-2 font-mono break-all">AVcja2dhAHwfXx6pzCLxN3HE3UQUi7ASOUCWekTZeRoh7rxI9oTC8b7dpAM</span>
            </div>
            <div class="grid grid-cols-3 text-sm">
              <span class="text-gray-500 dark:text-gray-400">NID</span>
              <span class="col-span-2 font-mono break-all">511=Dumv-TaAMW515jGP_nJkTlfLJpw0axio4UwI3Hc-qyz1_3m6hZUeMFmZ69KEW</span>
            </div>
          </div>
        </div>
      </div>
      
      <!-- Request Headers Section -->
      <div class="border-b border-gray-200 dark:border-gray-700">
        <button 
          class="flex items-center justify-between w-full p-3 text-left hover:bg-gray-100 dark:hover:bg-gray-750 transition-colors"
          @click="toggleSection('requestHeaders')"
        >
          <div class="flex items-center">
            <i class="bx bx-send text-lg text-gray-500 dark:text-gray-400 mr-2"></i>
            <span class="text-sm">Request Headers</span>
          </div>
          <span class="text-xs bg-gray-100 dark:bg-gray-700 text-gray-600 dark:text-gray-400 rounded-full px-2 py-0.5">22</span>
        </button>
        <div v-if="props.expandedSections.requestHeaders" class="p-3 border-t border-gray-100 dark:border-gray-700 bg-gray-50/50 dark:bg-gray-800/50">
          <div class="space-y-2">
            <div class="grid grid-cols-3 text-sm">
              <span class="text-gray-500 dark:text-gray-400">Host</span>
              <span class="col-span-2 font-mono">www.google.com</span>
            </div>
            <div class="grid grid-cols-3 text-sm">
              <span class="text-gray-500 dark:text-gray-400">User-Agent</span>
              <span class="col-span-2 font-mono break-all">Mozilla/5.0 (Windows NT 10.0)</span>
            </div>
            <div class="grid grid-cols-3 text-sm">
              <span class="text-gray-500 dark:text-gray-400">Accept</span>
              <span class="col-span-2 font-mono">text/html,application/xhtml+xml,application/xml</span>
            </div>
          </div>
        </div>
      </div>
      
      <!-- Response Headers Section -->
      <div class="border-b border-gray-200 dark:border-gray-700">
        <button 
          class="flex items-center justify-between w-full p-3 text-left hover:bg-gray-100 dark:hover:bg-gray-750 transition-colors"
          @click="toggleSection('responseHeaders')"
        >
          <div class="flex items-center">
            <i class="bx bx-download text-lg text-gray-500 dark:text-gray-400 mr-2"></i>
            <span class="text-sm">Response Headers</span>
          </div>
          <span class="text-xs bg-gray-100 dark:bg-gray-700 text-gray-600 dark:text-gray-400 rounded-full px-2 py-0.5">11</span>
        </button>
        <div v-if="props.expandedSections.responseHeaders" class="p-3 border-t border-gray-100 dark:border-gray-700 bg-gray-50/50 dark:bg-gray-800/50">
          <div class="space-y-2">
            <div class="grid grid-cols-3 text-sm">
              <span class="text-gray-500 dark:text-gray-400">Content-Type</span>
              <span class="col-span-2 font-mono">text/html; charset=UTF-8</span>
            </div>
            <div class="grid grid-cols-3 text-sm">
              <span class="text-gray-500 dark:text-gray-400">Server</span>
              <span class="col-span-2 font-mono">gws</span>
            </div>
            <div class="grid grid-cols-3 text-sm">
              <span class="text-gray-500 dark:text-gray-400">Cache-Control</span>
              <span class="col-span-2 font-mono">no-cache, no-store, max-age=0, must-revalidate</span>
            </div>
          </div>
        </div>
      </div>

      <!-- Notes Section -->
      <div class="border-b border-gray-200 dark:border-gray-700">
        <button 
          class="flex items-center justify-between w-full p-3 text-left hover:bg-gray-100 dark:hover:bg-gray-750 transition-colors"
        >
          <div class="flex items-center">
            <i class="bx bx-note text-lg text-gray-500 dark:text-gray-400 mr-2"></i>
            <span class="text-sm">Notes</span>
          </div>
        </button>
        <div class="p-3 border-t border-gray-100 dark:border-gray-700 bg-gray-50/50 dark:bg-gray-800/50 hidden">
          <textarea 
            spellcheck="false"
            class="w-full p-2 text-sm border border-gray-200 dark:border-gray-700 rounded-md bg-white dark:bg-gray-800 text-gray-800 dark:text-gray-200"
            placeholder="Add your notes about this request/response..."
            rows="3"
          ></textarea>
        </div>
      </div>
    </div>
  </div>
</template> 

<style scoped>
.inspector-panel {
  position: relative; /* For resize handle positioning */
}
</style> 
