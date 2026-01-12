<script setup lang="ts">
import { ref, nextTick, watch } from 'vue';

const props = defineProps<{
  show: boolean;
  title: string;
  submitText?: string;
}>();

const emit = defineEmits<{
  (e: 'close'): void;
  (e: 'submit', name: string, color: string): void;
}>();

const inputName = ref('');
const selectedColor = ref('#4f46e5');

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

const submit = () => {
  if (inputName.value.trim()) {
    emit('submit', inputName.value.trim(), selectedColor.value);
    inputName.value = '';
    selectedColor.value = '#4f46e5';
  }
};

const close = () => {
  inputName.value = '';
  selectedColor.value = '#4f46e5';
  emit('close');
};

const handleKeyDown = (event: KeyboardEvent) => {
  if (event.key === 'Enter') {
    submit();
  } else if (event.key === 'Escape') {
    close();
  }
};

const focusInput = () => {
  nextTick(() => {
    const input = document.getElementById('group-name-input');
    if (input) {
      input.focus();
    }
  });
};

// 监听模态框显示状态，管理 body 类和焦点
watch(() => props.show, (newShow) => {
  if (newShow) {
    // 显示模态框时
    document.body.classList.add('overflow-hidden');
    focusInput();
  } else {
    // 隐藏模态框时
    document.body.classList.remove('overflow-hidden');
  }
}, { immediate: true });
</script>

<template>
  <div v-if="show" class="repeater-modal-backdrop" @click="close">
    <div class="repeater-modal" @click.stop>
      <div class="repeater-modal-header">
        <h3>{{ title }}</h3>
        <button class="repeater-modal-close" @click="close">
          <i class="bx bx-x"></i>
        </button>
      </div>
      <div class="repeater-modal-body">
        <div class="repeater-form-group">
          <label for="group-name-input" class="repeater-form-label">名称</label>
          <input 
            id="group-name-input"
            type="text" 
            v-model="inputName" 
            class="repeater-input w-full" 
            placeholder="请输入名称"
            @keydown="handleKeyDown"
            spellcheck="false"
          />
        </div>
        
        <div class="repeater-form-group">
          <label class="repeater-form-label">颜色</label>
          <div class="repeater-color-options">
            <div 
              v-for="color in colorOptions" 
              :key="color.id"
              class="repeater-color-option"
              :class="{ 'repeater-color-selected': selectedColor === color.value }"
              @click="selectedColor = color.value"
            >
              <div class="repeater-color-swatch" :style="{ backgroundColor: color.value }"></div>
              <span class="repeater-color-label">{{ color.label }}</span>
            </div>
          </div>
        </div>
      </div>
      <div class="repeater-modal-footer">
        <button class="btn btn-secondary mr-2" @click="close">取消</button>
        <button 
          class="btn btn-primary" 
          @click="submit"
          :disabled="!inputName.trim()"
        >
          {{ submitText || '确认' }}
        </button>
      </div>
    </div>
  </div>
</template>

<style scoped>
.repeater-color-options {
  display: grid;
  grid-template-columns: repeat(2, 1fr);
  gap: 8px;
}

.repeater-color-option {
  display: flex;
  align-items: center;
  padding: 4px 8px;
  border-radius: 4px;
  cursor: pointer;
  transition: all 0.2s ease;
}

.repeater-color-option:hover {
  background-color: var(--repeater-bg-secondary);
}

.repeater-color-selected {
  background-color: var(--repeater-bg-secondary);
  border: 1px solid var(--repeater-border);
}

.repeater-color-swatch {
  width: 16px;
  height: 16px;
  border-radius: 50%;
  margin-right: 8px;
  border: 1px solid var(--repeater-border);
}

.repeater-color-label {
  font-size: 0.75rem;
  color: var(--repeater-text-primary);
}
</style> 