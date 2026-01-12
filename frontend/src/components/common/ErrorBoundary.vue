<script setup lang="ts">
import { ref, onErrorCaptured } from 'vue';
import { useI18n } from 'vue-i18n';

// 定义错误类型
interface ErrorInfo {
  error: Error;
  info: string;
  timestamp: number;
  componentStack?: string;
}

// 定义Props，允许自定义错误信息和样式
const props = defineProps<{
  // 自定义显示的组件名称或标识
  componentName?: string;
  
  // 自定义错误回调，可用于错误上报
  onError?: (errorInfo: ErrorInfo) => void;
  
  // 是否显示重试按钮
  showRetry?: boolean;
  
  // 是否总是显示后备UI，用于调试
  fallback?: boolean;
  
  // 是否显示详细的错误信息（在生产环境中应设为false）
  showDetails?: boolean;
  
  // 最大重试次数
  maxRetries?: number;
}>();

// 使用i18n
const { t } = useI18n();

// 当前错误状态
const error = ref<ErrorInfo | null>(null);
const retryCount = ref(0);
const componentName = ref(props.componentName || '组件');

// 重置错误状态
const resetError = () => {
  if (props.maxRetries && retryCount.value >= props.maxRetries) {
    console.warn('已达到最大重试次数');
    return;
  }
  
  error.value = null;
  retryCount.value++;
};

// 捕获渲染错误
onErrorCaptured((err: Error, instance, info) => {
  console.error(`Error in ${componentName.value}:`, err);
  
  // 构建错误信息对象
  const errorInfo: ErrorInfo = {
    error: err,
    info,
    timestamp: Date.now(),
    componentStack: instance?.$options?.__file
  };
  
  // 设置错误状态
  error.value = errorInfo;
  
  // 调用自定义错误处理函数
  if (props.onError) {
    props.onError(errorInfo);
  }
  
  // 阻止错误继续传播
  return false;
});

// 提供给父组件的方法
defineExpose({
  resetError,
  getError: () => error.value,
  getRetryCount: () => retryCount.value
});
</script>

<template>
  <!-- 显示错误状态或正常内容 -->
  <div class="error-boundary">
    <div v-if="error || fallback" class="error-display">
      <div class="error-content">
        <div class="error-icon">
          <i class="bx bx-error-circle text-red-500 dark:text-red-400 text-3xl"></i>
        </div>
        <h3 class="error-title">{{ t('common.error.error_boundary_title', { component: componentName }) }}</h3>
        <p class="error-message">{{ t('common.error.error_boundary_message') }}</p>
        
        <!-- 错误详情（仅在开发环境或显示详情时显示） -->
        <div v-if="showDetails && error" class="error-details">
          <p class="font-mono text-xs text-red-600 dark:text-red-400 bg-red-50 dark:bg-red-900/20 p-2 rounded overflow-auto max-h-32">
            {{ error.error.message }}
          </p>
          <details class="mt-2">
            <summary class="text-xs cursor-pointer text-gray-600 dark:text-gray-400">{{ t('common.error.stack_trace') }}</summary>
            <pre class="font-mono text-2xs text-gray-600 dark:text-gray-400 bg-gray-50 dark:bg-gray-800/50 p-2 mt-1 rounded overflow-auto max-h-40">{{ error.error.stack }}</pre>
          </details>
        </div>
        
        <!-- 显示重试按钮 -->
        <button 
          v-if="showRetry" 
          @click="resetError" 
          class="retry-button"
        >
          <i class="bx bx-refresh mr-1"></i>
          {{ t('common.actions.retry') }}
        </button>
        
        <!-- 后备UI插槽 -->
        <slot name="fallback" :error="error"></slot>
      </div>
    </div>
    
    <!-- 正常内容 -->
    <template v-else>
      <slot></slot>
    </template>
  </div>
</template>

<style scoped>
.error-boundary {
  width: 100%;
  height: 100%;
  position: relative;
}

.error-display {
  width: 100%;
  height: 100%;
  display: flex;
  align-items: center;
  justify-content: center;
  text-align: center;
  padding: 1rem;
  background-color: var(--color-bg-primary);
}

.error-content {
  max-width: 30rem;
  padding: 1.5rem;
  border-radius: 0.5rem;
  background-color: var(--color-bg-primary);
  box-shadow: var(--glass-shadow-strong);
  transition: all 0.3s ease;
}

.error-icon {
  margin-bottom: 1rem;
  animation: pulse-error 2s infinite;
}

.error-title {
  font-size: 1.125rem;
  font-weight: 600;
  color: var(--color-text-primary);
  margin-bottom: 0.5rem;
}

.error-message {
  color: var(--color-text-secondary);
  margin-bottom: 1rem;
}

.error-details {
  margin-top: 1rem;
  transition: all 0.3s ease;
}

.retry-button {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  padding: 0.5rem 1rem;
  font-size: 0.875rem;
  font-weight: 500;
  color: white;
  background-color: var(--color-primary);
  border-radius: 0.375rem;
  border: none;
  cursor: pointer;
  transition: all 0.2s ease;
  margin-top: 1rem;
}

.retry-button:hover {
  background-color: var(--color-primary-hover);
  transform: translateY(-1px);
}

.retry-button:active {
  transform: translateY(0);
}

.text-2xs {
  font-size: 0.625rem;
  line-height: 0.875rem;
}

/* 动画已移至 animations.css 统一管理 */
</style> 