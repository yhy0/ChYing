<script setup lang="ts">
import { ref } from 'vue';

defineProps<{
  show: boolean;
}>();

const emit = defineEmits<{
  (e: 'close'): void;
  (e: 'create', name: string, color: string): void;
}>();

// 分组相关状态
const newGroupName = ref('');
const selectedColor = ref('#4f46e5'); // 添加颜色选择状态

// 预定义的颜色选项
const colorOptions = [
  { id: 'default', value: '#4f46e5', label: '默认 (紫色)' },
  { id: 'red', value: '#ef4444', label: '红色' },
  { id: 'green', value: '#10b981', label: '绿色' },
  { id: 'blue', value: '#3b82f6', label: '蓝色' },
  { id: 'yellow', value: '#f59e0b', label: '黄色' },
  { id: 'orange', value: '#f97316', label: '橙色' },
  { id: 'teal', value: '#14b8a6', label: '青色' },
];

// 提交新分组
const submitNewGroup = () => {
  if (newGroupName.value.trim()) {
    emit('create', newGroupName.value.trim(), selectedColor.value);
    resetForm();
  }
};

// 关闭模态框
const closeGroupModal = () => {
  resetForm();
  emit('close');
};

// 重置表单
const resetForm = () => {
  newGroupName.value = '';
  selectedColor.value = '#4f46e5'; // 重置为默认颜色
};

// 处理模态框内键盘事件
const handleModalKeyDown = (event: KeyboardEvent) => {
  if (event.key === 'Enter') {
    submitNewGroup();
  } else if (event.key === 'Escape') {
    closeGroupModal();
  }
};
</script>

<template>
  <div v-if="show" class="dialog-overlay" @click="closeGroupModal">
    <div class="dialog-container dialog-sm" @click.stop>
      <div class="dialog-header">
        <div class="flex justify-between items-center">
          <h3 class="text-lg font-medium">创建新分组</h3>
          <button class="text-gray-500 hover:text-gray-700 dark:text-gray-400 dark:hover:text-gray-200"
            @click="closeGroupModal">
            <i class="bx bx-x text-xl"></i>
          </button>
        </div>
      </div>

      <div class="dialog-content">
        <div class="mb-4">
          <label class="settings-label">分组名称</label>
          <input id="new-group-input-intruder" v-model="newGroupName" type="text" class="w-full px-3 py-2 border border-gray-300 dark:border-gray-700
                 rounded-md text-gray-800 dark:text-gray-200
                 bg-white dark:bg-[#282838]
                 focus:outline-none focus:ring-2 focus:ring-indigo-500" placeholder="输入分组名称" spellcheck="false" 
            @keydown="handleModalKeyDown" />
        </div>

        <div class="mb-4">
          <label class="block text-sm text-gray-700 dark:text-gray-300 mb-1">颜色</label>
          <div class="grid grid-cols-2 gap-2">
            <div v-for="color in colorOptions" :key="color.id"
              class="flex items-center p-2 rounded cursor-pointer transition-colors" :class="{
                'bg-gray-100 dark:bg-[#282838] border border-gray-300 dark:border-gray-700': selectedColor === color.value
              }" @click="selectedColor = color.value">
              <div class="w-4 h-4 rounded-full mr-2 border border-gray-300 dark:border-gray-700"
                :style="{ backgroundColor: color.value }"></div>
              <span class="text-xs text-gray-700 dark:text-gray-300">{{ color.label }}</span>
            </div>
          </div>
        </div>
      </div>

      <div class="dialog-footer">
        <button class="px-3 py-1.5 text-sm text-gray-700 dark:text-gray-300
            border border-gray-300 dark:border-gray-700
            hover:bg-gray-100 dark:hover:bg-[#282838] rounded-md" @click="closeGroupModal">
          取消
        </button>
        <button class="px-3 py-1.5 text-sm text-white bg-indigo-600 hover:bg-indigo-700
            rounded-md" @click="submitNewGroup" :disabled="!newGroupName.trim()"
          :class="{ 'opacity-50 cursor-not-allowed': !newGroupName.trim() }">
          创建
        </button>
      </div>
    </div>
    </div>

</template>