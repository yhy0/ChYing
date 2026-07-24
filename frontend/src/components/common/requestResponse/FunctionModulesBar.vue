<script setup lang="ts">
// 定义props
const props = defineProps<{
  activeModule: string | null;
  hasNotes?: boolean;
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

      <!-- Notes按钮：有内容且未激活时图标变主题色 + 右上角圆点，提示该 tab 有 note -->
      <button
        @click="toggleModule('notes')"
        class="module-btn relative w-9 h-9 flex items-center justify-center rounded-md transition-all duration-300 ease-in-out"
        :class="{
          'active-module-btn': props.activeModule === 'notes',
          'text-gray-600 dark:text-gray-400 hover:bg-gray-50 dark:hover:bg-gray-750': !props.hasNotes || props.activeModule === 'notes',
          'text-indigo-500 dark:text-indigo-400 hover:bg-indigo-50 dark:hover:bg-indigo-900/20': props.hasNotes && props.activeModule !== 'notes',
          'hover:scale-105': props.activeModule !== 'notes',
          'shadow-md': props.activeModule === 'notes'
        }"
        title="Notes"
      >
        <i class="bx bx-note text-xl"></i>
        <span
          v-if="props.hasNotes && props.activeModule !== 'notes'"
          class="absolute top-1 right-1 w-2 h-2 rounded-full bg-indigo-500 ring-2 ring-white dark:ring-gray-800"
        ></span>
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