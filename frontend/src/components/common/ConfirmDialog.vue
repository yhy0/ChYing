<script setup lang="ts">
/**
 * ConfirmDialog - 确认对话框组件
 * 用于替代原生 confirm()，保持设计一致性
 */

import { ref, watch, onMounted, onUnmounted } from 'vue';

interface Props {
  modelValue: boolean;      // v-model 绑定
  title?: string;           // 对话框标题
  message?: string;         // 确认消息
  confirmText?: string;     // 确认按钮文本
  cancelText?: string;      // 取消按钮文本
  type?: 'info' | 'warning' | 'danger';  // 类型，影响样式
  icon?: string;            // 自定义图标
  loading?: boolean;        // 确认按钮加载状态
  closeOnOverlay?: boolean; // 点击遮罩是否关闭
}

const props = withDefaults(defineProps<Props>(), {
  modelValue: false,
  title: '确认操作',
  message: '确定要执行此操作吗？',
  confirmText: '确认',
  cancelText: '取消',
  type: 'warning',
  icon: '',
  loading: false,
  closeOnOverlay: true
});

const emit = defineEmits<{
  (e: 'update:modelValue', value: boolean): void;
  (e: 'confirm'): void;
  (e: 'cancel'): void;
}>();

// 内部显示状态
const visible = ref(props.modelValue);

// 同步外部状态
watch(() => props.modelValue, (val) => {
  visible.value = val;
});

// 获取类型对应的图标
const getTypeIcon = () => {
  if (props.icon) return props.icon;
  switch (props.type) {
    case 'danger':
      return 'bx-error-circle';
    case 'warning':
      return 'bx-error';
    case 'info':
    default:
      return 'bx-info-circle';
  }
};

// 获取类型对应的颜色类
const getTypeClass = () => {
  switch (props.type) {
    case 'danger':
      return 'type-danger';
    case 'warning':
      return 'type-warning';
    case 'info':
    default:
      return 'type-info';
  }
};

// 关闭对话框
const close = () => {
  visible.value = false;
  emit('update:modelValue', false);
};

// 确认操作
const handleConfirm = () => {
  if (props.loading) return;
  emit('confirm');
  close();
};

// 取消操作
const handleCancel = () => {
  emit('cancel');
  close();
};

// 点击遮罩
const handleOverlayClick = () => {
  if (props.closeOnOverlay) {
    handleCancel();
  }
};

// ESC 键关闭
const handleKeydown = (e: KeyboardEvent) => {
  if (e.key === 'Escape' && visible.value) {
    handleCancel();
  }
};

onMounted(() => {
  document.addEventListener('keydown', handleKeydown);
});

onUnmounted(() => {
  document.removeEventListener('keydown', handleKeydown);
});
</script>

<template>
  <Teleport to="body">
    <Transition name="dialog-fade">
      <div 
        v-if="visible" 
        class="confirm-dialog-overlay"
        @click.self="handleOverlayClick"
        role="dialog"
        aria-modal="true"
        :aria-labelledby="title ? 'dialog-title' : undefined"
        :aria-describedby="message ? 'dialog-message' : undefined"
      >
        <div class="confirm-dialog" :class="getTypeClass()">
          <!-- 图标 -->
          <div class="dialog-icon">
            <i class="bx" :class="getTypeIcon()"></i>
          </div>
          
          <!-- 标题 -->
          <h3 v-if="title" id="dialog-title" class="dialog-title">
            {{ title }}
          </h3>
          
          <!-- 消息 -->
          <p v-if="message" id="dialog-message" class="dialog-message">
            {{ message }}
          </p>
          
          <!-- 自定义内容插槽 -->
          <div v-if="$slots.default" class="dialog-content">
            <slot></slot>
          </div>
          
          <!-- 操作按钮 -->
          <div class="dialog-actions">
            <button 
              class="btn btn-cancel"
              @click="handleCancel"
              :disabled="loading"
            >
              {{ cancelText }}
            </button>
            <button 
              class="btn btn-confirm"
              :class="[`btn-${type}`, { 'btn-loading': loading }]"
              @click="handleConfirm"
              :disabled="loading"
            >
              <i v-if="loading" class="bx bx-loader-alt bx-spin"></i>
              {{ confirmText }}
            </button>
          </div>
        </div>
      </div>
    </Transition>
  </Teleport>
</template>

<style scoped>
.confirm-dialog-overlay {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background: rgba(0, 0, 0, 0.5);
  backdrop-filter: blur(4px);
  -webkit-backdrop-filter: blur(4px);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 9999;
  padding: 1rem;
}

.confirm-dialog {
  background: var(--glass-bg-popup);
  backdrop-filter: var(--glass-blur-medium);
  -webkit-backdrop-filter: var(--glass-blur-medium);
  border-radius: var(--radius-xl);
  padding: 1.5rem;
  max-width: 400px;
  width: 100%;
  text-align: center;
  box-shadow: var(--glass-shadow-strong);
  border: 1px solid var(--glass-border-light);
  animation: dialog-enter 0.2s ease-out;
}

@keyframes dialog-enter {
  from {
    opacity: 0;
    transform: scale(0.95) translateY(-10px);
  }
  to {
    opacity: 1;
    transform: scale(1) translateY(0);
  }
}

.dialog-icon {
  width: 56px;
  height: 56px;
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  margin: 0 auto 1rem;
  font-size: 1.75rem;
}

.type-info .dialog-icon {
  background: rgba(59, 130, 246, 0.1);
  color: var(--color-info);
}

.type-warning .dialog-icon {
  background: rgba(245, 158, 11, 0.1);
  color: var(--color-warning);
}

.type-danger .dialog-icon {
  background: rgba(239, 68, 68, 0.1);
  color: var(--color-danger);
}

.dialog-title {
  font-size: 1.125rem;
  font-weight: 600;
  color: var(--color-text-primary);
  margin: 0 0 0.5rem 0;
}

.dialog-message {
  font-size: 0.9375rem;
  color: var(--color-text-secondary);
  margin: 0 0 1.5rem 0;
  line-height: 1.5;
}

.dialog-content {
  margin-bottom: 1.5rem;
}

.dialog-actions {
  display: flex;
  gap: 0.75rem;
  justify-content: center;
}

.dialog-actions .btn {
  flex: 1;
  max-width: 140px;
  padding: 0.625rem 1rem;
  border-radius: var(--radius-md);
  font-size: 0.875rem;
  font-weight: 500;
  cursor: pointer;
  transition: var(--glass-transition-fast);
  display: inline-flex;
  align-items: center;
  justify-content: center;
  gap: 0.5rem;
}

.btn-cancel {
  background: var(--glass-bg-secondary);
  color: var(--color-text-primary);
  border: 1px solid var(--glass-border);
}

.btn-cancel:hover:not(:disabled) {
  background: var(--glass-bg-hover);
  border-color: var(--glass-border-strong);
}

.btn-confirm {
  border: none;
  color: white;
}

.btn-info {
  background: var(--color-info);
}

.btn-info:hover:not(:disabled) {
  background: var(--color-info-hover);
}

.btn-warning {
  background: var(--color-warning);
  color: var(--color-text-primary);
}

.btn-warning:hover:not(:disabled) {
  background: var(--color-warning-hover);
}

.btn-danger {
  background: var(--color-danger);
}

.btn-danger:hover:not(:disabled) {
  background: var(--color-danger-hover);
}

.btn:disabled {
  opacity: 0.6;
  cursor: not-allowed;
}

.btn-loading {
  pointer-events: none;
}

/* 过渡动画 */
.dialog-fade-enter-active,
.dialog-fade-leave-active {
  transition: opacity 0.2s ease;
}

.dialog-fade-enter-active .confirm-dialog,
.dialog-fade-leave-active .confirm-dialog {
  transition: transform 0.2s ease, opacity 0.2s ease;
}

.dialog-fade-enter-from,
.dialog-fade-leave-to {
  opacity: 0;
}

.dialog-fade-enter-from .confirm-dialog,
.dialog-fade-leave-to .confirm-dialog {
  transform: scale(0.95) translateY(-10px);
  opacity: 0;
}

/* 深色模式 */
.dark .confirm-dialog {
  background: var(--glass-bg-popup);
  border-color: var(--glass-border);
}

.dark .type-info .dialog-icon {
  background: rgba(99, 102, 241, 0.15);
}

.dark .type-warning .dialog-icon {
  background: rgba(245, 158, 11, 0.15);
}

.dark .type-danger .dialog-icon {
  background: rgba(239, 68, 68, 0.15);
}
</style>
