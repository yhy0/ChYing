<script setup lang="ts">

// 定义props
const props = defineProps<{
  panelWidth: number;
  notes: string;
}>();

// 定义事件
const emit = defineEmits<{
  (e: 'close'): void;
  (e: 'startResizePanel', event: MouseEvent, panel: string): void;
  (e: 'update:notes', value: string): void;
}>();

// 关闭面板
const closePanel = () => {
  emit('close');
};

// 开始调整面板大小
const startResizePanel = (e: MouseEvent) => {
  emit('startResizePanel', e, 'notes');
};

// 更新笔记内容
const updateNotes = (e: Event) => {
  const value = (e.target as HTMLTextAreaElement).value;
  emit('update:notes', value);
};
</script>

<template>
  <div 
    class="notes-panel flex flex-col border-l border-gray-200 dark:border-gray-700 bg-white dark:bg-gray-800 shadow-sm"
    :style="{ width: props.panelWidth + 'px' }"
  >
    <!-- 拖动调整大小的手柄 -->
    <div 
      class="resize-handle absolute top-0 left-0 bottom-0 w-3 bg-transparent cursor-ew-resize z-10 flex items-center justify-center hover:bg-gray-200/50 dark:hover:bg-gray-700/50 opacity-0 hover:opacity-100 transition-opacity duration-200"
      @mousedown="startResizePanel"
    >
      <div class="h-12 w-0.5 bg-gray-300/80 dark:bg-gray-600/80 rounded-full"></div>
    </div>

    <!-- Notes Header -->
    <div class="flex items-center justify-between p-3 border-b border-gray-200 dark:border-gray-700 bg-gray-50 dark:bg-gray-800">
      <h3 class="text-sm font-medium">Notes</h3>
      <button 
        @click="closePanel"
        class="p-1 rounded-full hover:bg-gray-200 dark:hover:bg-gray-700 transition-colors"
        title="Close"
      >
        <i class="bx bx-x text-lg text-gray-600 dark:text-gray-400"></i>
      </button>
    </div>
    
    <!-- Notes Content -->
    <div class="flex-1 overflow-auto p-3">
      <textarea 
        spellcheck="false"
        :value="props.notes"
        @input="updateNotes"
        class="w-full p-3 h-full text-sm border border-gray-200 dark:border-gray-700 rounded-md bg-white dark:bg-gray-800 text-gray-800 dark:text-gray-200 resize-none"
        placeholder="Add your notes about this request/response..."
      ></textarea>
    </div>
  </div>
</template> 


<style scoped>
.notes-panel {
  position: relative; /* For resize handle positioning */
}
</style> 
