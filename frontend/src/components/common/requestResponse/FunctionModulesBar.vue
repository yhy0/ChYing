<script setup lang="ts">
// 定义props
const props = defineProps<{
  activeModule: string | null;
}>();

// 定义事件
const emit = defineEmits<{
  (e: 'toggleModule', moduleName: string): void;
}>();

// 处理模块切换
const toggleModule = (moduleName: string) => {
  emit('toggleModule', moduleName);
};
</script>

<template>
  <div class="function-modules flex flex-col border-l border-gray-200 dark:border-gray-700 bg-white dark:bg-gray-800 shadow-sm">
    <!-- 功能按钮 -->
    <div class="function-buttons flex flex-col items-center py-4 space-y-5">
      <!-- Inspector按钮 -->
      <button 
        @click="toggleModule('inspector')"
        class="module-btn w-9 h-9 flex items-center justify-center rounded-md text-gray-600 dark:text-gray-400 hover:bg-gray-50 dark:hover:bg-gray-750 transition-all duration-300 ease-in-out"
        :class="{
          'active-module-btn': props.activeModule === 'inspector',
          'hover:scale-105': props.activeModule !== 'inspector',
          'shadow-md': props.activeModule === 'inspector'
        }"
        title="Inspector"
      >
        <i class="bx bx-analyse text-xl"></i>
      </button>
      
      <!-- Notes按钮 -->
      <button 
        @click="toggleModule('notes')"
        class="module-btn w-9 h-9 flex items-center justify-center rounded-md text-gray-600 dark:text-gray-400 hover:bg-gray-50 dark:hover:bg-gray-750 transition-all duration-300 ease-in-out"
        :class="{
          'active-module-btn': props.activeModule === 'notes',
          'hover:scale-105': props.activeModule !== 'notes',
          'shadow-md': props.activeModule === 'notes'
        }"
        title="Notes"
      >
        <i class="bx bx-note text-xl"></i>
      </button>

      <!-- Extractor按钮 -->
      <button 
        @click="toggleModule('extractor')"
        class="module-btn w-9 h-9 flex items-center justify-center rounded-md text-gray-600 dark:text-gray-400 hover:bg-gray-50 dark:hover:bg-gray-750 transition-all duration-300 ease-in-out"
        :class="{
          'active-module-btn': props.activeModule === 'extractor',
          'hover:scale-105': props.activeModule !== 'extractor',
          'shadow-md': props.activeModule === 'extractor'
        }"
        title="Data Extractor"
      >
        <i class="bx bx-file-find text-xl"></i>
      </button>
    </div>
  </div>
</template> 