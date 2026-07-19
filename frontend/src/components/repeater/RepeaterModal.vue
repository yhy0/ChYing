<script setup lang="ts">
import { ref, nextTick, watch, computed } from 'vue';
import { useI18n } from 'vue-i18n';

const { t } = useI18n();

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
const colorOptions = computed(() => [
  { id: 'default', value: '#4f46e5', label: t('modules.repeater.modal.default_purple') },
  { id: 'red', value: '#ef4444', label: t('modules.repeater.modal.red') },
  { id: 'green', value: '#10b981', label: t('modules.repeater.modal.green') },
  { id: 'blue', value: '#3b82f6', label: t('modules.repeater.modal.blue') },
  { id: 'yellow', value: '#f59e0b', label: t('modules.repeater.modal.yellow') },
  { id: 'orange', value: '#f97316', label: t('modules.repeater.modal.orange') },
  { id: 'teal', value: '#14b8a6', label: t('modules.repeater.modal.cyan') },
]);

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
  <div v-if="show" class="dialog-overlay" @click="close">
    <div class="dialog-container dialog-sm" role="dialog" aria-modal="true" @click.stop>
      <div class="dialog-header">
        <h3 class="dialog-title">{{ title }}</h3>
        <button class="btn btn-icon btn-icon-xs" @click="close" :aria-label="t('common.actions.close')">
          <i class="bx bx-x"></i>
        </button>
      </div>
      <div class="dialog-content">
        <div class="form-field">
          <label for="group-name-input" class="form-label">{{ t('modules.repeater.modal.name_label') }}</label>
          <input
            id="group-name-input"
            type="text"
            v-model="inputName"
            class="form-input"
            :placeholder="t('modules.repeater.modal.enter_name')"
            @keydown="handleKeyDown"
            spellcheck="false"
          />
        </div>
        
        <div class="form-field">
          <label class="form-label">{{ t('modules.repeater.modal.color_label') }}</label>
          <div class="repeater-color-options">
            <button
              v-for="color in colorOptions" 
              :key="color.id"
              type="button"
              class="repeater-color-option"
              :class="{ 'repeater-color-selected': selectedColor === color.value }"
              :aria-pressed="selectedColor === color.value"
              @click="selectedColor = color.value"
            >
              <div class="repeater-color-swatch" :style="{ backgroundColor: color.value }"></div>
              <span class="repeater-color-label">{{ color.label }}</span>
            </button>
          </div>
        </div>
      </div>
      <div class="dialog-footer">
        <button class="btn btn-secondary btn-sm" @click="close">{{ t('common.actions.cancel') }}</button>
        <button 
          class="btn btn-primary btn-sm" 
          @click="submit"
          :disabled="!inputName.trim()"
        >
          {{ submitText || t('common.confirm.default_confirm') }}
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
  border: 1px solid transparent;
  background: transparent;
  color: var(--color-text-primary);
  text-align: left;
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
