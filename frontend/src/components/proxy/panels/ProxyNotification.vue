<script setup lang="ts">
import { ref, watch, onBeforeUnmount } from 'vue';

const props = defineProps<{
  message: string;
  type: 'success' | 'error';
  visible: boolean;
}>();

const emit = defineEmits<{
  (e: 'close'): void;
}>();

const isClosing = ref(false);
let autoCloseTimer: number | null = null;

// 设置自动关闭定时器
const setupAutoClose = () => {
  // 清除可能存在的旧定时器
  clearAutoCloseTimer();
  
  // 设置新的定时器，2.7秒后开始关闭动画
  autoCloseTimer = window.setTimeout(() => {
    startCloseAnimation();
  }, 1700);
};

// 清除自动关闭定时器
const clearAutoCloseTimer = () => {
  if (autoCloseTimer) {
    window.clearTimeout(autoCloseTimer);
    autoCloseTimer = null;
  }
};

// 监视visible属性变化
watch(() => props.visible, (isVisible) => {
  if (isVisible) {
    // 通知显示时，重置关闭状态并设置自动关闭
    isClosing.value = false;
    setupAutoClose();
  } else {
    // 通知隐藏时，清除定时器
    clearAutoCloseTimer();
  }
}, { immediate: true });

// 开始关闭动画
const startCloseAnimation = () => {
  // 确保通知仍然显示时才执行关闭动画
  if (props.visible && !isClosing.value) {
    isClosing.value = true;
    
    // 等待动画完成后发出关闭事件
    window.setTimeout(() => {
      emit('close');
    }, 300);
  }
};

// 手动关闭通知
const closeNotification = () => {
  clearAutoCloseTimer();
  startCloseAnimation();
};

// 组件卸载前清除定时器
onBeforeUnmount(() => {
  clearAutoCloseTimer();
});
</script>

<template>
  <div 
    v-if="visible" 
    :class="[
      'proxy-notification',
      isClosing ? 'closing' : '',
      type === 'success' ? 'proxy-notification-success' : 'proxy-notification-error'
    ]"
  >
    <div class="proxy-notification-content">
      <i :class="[
        'proxy-notification-icon',
        type === 'success' ? 'bx bx-check-circle' : 'bx bx-error-circle'
      ]"></i>
      <span class="proxy-notification-message">{{ message }}</span>
    </div>
    
    <button 
      class="proxy-notification-close" 
      @click="closeNotification"
      title="关闭"
    >
      <i class="bx bx-x"></i>
    </button>
    
    <!-- 倒计时进度条 -->
    <div class="proxy-notification-progress"></div>
  </div>
</template> 