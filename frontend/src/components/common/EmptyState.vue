<script setup lang="ts">
/**
 * EmptyState - 通用空状态组件
 * 用于显示列表为空、无数据等状态
 */

import { computed } from 'vue';

interface Props {
  icon?: string;           // BoxIcons 图标类名，如 'bx-folder-open'
  title?: string;          // 标题文本
  description?: string;    // 描述文本
  size?: 'small' | 'medium' | 'large';  // 尺寸
  showAction?: boolean;    // 是否显示操作按钮
  actionText?: string;     // 操作按钮文本
  actionIcon?: string;     // 操作按钮图标
}

const props = withDefaults(defineProps<Props>(), {
  icon: 'bx-inbox',
  title: '暂无数据',
  description: '',
  size: 'medium',
  showAction: false,
  actionText: '刷新',
  actionIcon: 'bx-refresh'
});

const emit = defineEmits<{
  (e: 'action'): void;
}>();

// 根据尺寸计算样式类
const sizeClasses = computed(() => {
  switch (props.size) {
    case 'small':
      return {
        container: 'py-6 px-4',
        icon: 'text-3xl mb-2',
        title: 'text-sm',
        description: 'text-xs',
        button: 'text-xs px-3 py-1.5'
      };
    case 'large':
      return {
        container: 'py-16 px-8',
        icon: 'text-6xl mb-6',
        title: 'text-xl',
        description: 'text-base',
        button: 'text-base px-6 py-3'
      };
    default: // medium
      return {
        container: 'py-10 px-6',
        icon: 'text-5xl mb-4',
        title: 'text-base',
        description: 'text-sm',
        button: 'text-sm px-4 py-2'
      };
  }
});

const handleAction = () => {
  emit('action');
};
</script>

<template>
  <div 
    class="empty-state" 
    :class="sizeClasses.container"
    role="status"
    aria-live="polite"
  >
    <!-- 图标 -->
    <div class="empty-state-icon" :class="sizeClasses.icon">
      <i class="bx" :class="icon"></i>
    </div>
    
    <!-- 标题 -->
    <h4 class="empty-state-title" :class="sizeClasses.title">
      {{ title }}
    </h4>
    
    <!-- 描述 -->
    <p v-if="description" class="empty-state-description" :class="sizeClasses.description">
      {{ description }}
    </p>
    
    <!-- 操作按钮 -->
    <button 
      v-if="showAction" 
      class="empty-state-action" 
      :class="sizeClasses.button"
      @click="handleAction"
    >
      <i v-if="actionIcon" class="bx" :class="actionIcon"></i>
      {{ actionText }}
    </button>
    
    <!-- 自定义插槽 -->
    <slot></slot>
  </div>
</template>

<style scoped>
.empty-state {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  text-align: center;
  color: var(--color-text-secondary);
  background: var(--glass-bg-tertiary);
  border-radius: var(--radius-lg);
  border: 1px dashed var(--glass-border-light);
  transition: var(--glass-transition-normal);
}

.empty-state-icon {
  color: var(--color-text-tertiary);
  opacity: 0.6;
  transition: var(--glass-transition-normal);
}

.empty-state:hover .empty-state-icon {
  opacity: 0.8;
  color: var(--color-primary);
}

.empty-state-title {
  font-weight: 600;
  color: var(--color-text-primary);
  margin: 0;
  line-height: 1.4;
}

.empty-state-description {
  color: var(--color-text-secondary);
  margin: 0.5rem 0 0 0;
  line-height: 1.5;
  max-width: 300px;
}

.empty-state-action {
  display: inline-flex;
  align-items: center;
  gap: 0.5rem;
  margin-top: 1rem;
  background: var(--color-primary);
  color: white;
  border: none;
  border-radius: var(--radius-md);
  font-weight: 500;
  cursor: pointer;
  transition: var(--glass-transition-fast);
}

.empty-state-action:hover {
  background: var(--color-primary-hover);
  box-shadow: var(--glass-shadow-light);
}

.empty-state-action:active {
  transform: scale(0.98);
}

/* 深色模式优化 */
.dark .empty-state {
  background: var(--glass-bg-secondary);
  border-color: var(--glass-border);
}

.dark .empty-state-icon {
  opacity: 0.5;
}

.dark .empty-state:hover .empty-state-icon {
  opacity: 0.7;
}
</style>
