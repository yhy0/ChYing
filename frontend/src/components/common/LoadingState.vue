<script setup lang="ts">
import { useI18n } from 'vue-i18n';
import { ref, watch, computed, onUnmounted } from 'vue';

// 定义Props
const props = defineProps<{
  // 是否正在加载
  loading: boolean;
  
  // 显示的加载消息
  message?: string;
  
  // 加载类型：spinner, dots, pulse
  type?: 'spinner' | 'dots' | 'pulse';
  
  // 加载指示器尺寸
  size?: 'sm' | 'md' | 'lg';
  
  // 是否显示在整屏
  fullscreen?: boolean;
  
  // 自定义加载遮罩的z-index
  zIndex?: number;
  
  // 是否使用透明背景
  transparent?: boolean;
  
  // 延迟显示(毫秒)，避免闪烁
  delay?: number;
}>();

// 解构Props，设置默认值
const {
  loading = false,
  message = '',
  type = 'spinner',
  size = 'md',
  fullscreen = false,
  zIndex = 50,
  transparent = false,
  delay = 200
} = props;

// 使用i18n
const { t } = useI18n();

// 计算默认消息
const defaultMessage = t('common.loading');

// 根据size生成样式类
const sizeClass = {
  sm: 'w-4 h-4',
  md: 'w-8 h-8',
  lg: 'w-12 h-12'
}[size];

// 延迟显示控制
const showDelayed = ref(false);
let delayTimer: ReturnType<typeof setTimeout> | null = null;

// 监控loading属性变化
watch(() => props.loading, (newValue) => {
  if (newValue) {
    // 设置延迟显示定时器
    if (delayTimer) clearTimeout(delayTimer);
    delayTimer = setTimeout(() => {
      showDelayed.value = true;
    }, delay);
  } else {
    // 重置状态
    if (delayTimer) clearTimeout(delayTimer);
    showDelayed.value = false;
  }
}, { immediate: true });

// 组件卸载时清理定时器
onUnmounted(() => {
  if (delayTimer) clearTimeout(delayTimer);
});

// 计算显示加载器的条件
const shouldShow = computed(() => loading && (delay === 0 || showDelayed.value));
</script>

<template>
  <div v-if="shouldShow" class="loading-container" :class="{ 'fullscreen': fullscreen, 'transparent-bg': transparent }" :style="{ zIndex }">
    <div class="loading-content">
      <!-- 加载动画 -->
      <div class="loading-indicator" :class="[sizeClass]">
        <!-- 旋转加载 -->
        <div v-if="type === 'spinner'" class="spinner" :class="[sizeClass]"></div>
        
        <!-- 点状加载 -->
        <div v-else-if="type === 'dots'" class="dots" :class="[sizeClass]">
          <div class="dot"></div>
          <div class="dot"></div>
          <div class="dot"></div>
        </div>
        
        <!-- 脉冲加载 -->
        <div v-else-if="type === 'pulse'" class="pulse-loader" :class="[sizeClass]"></div>
      </div>
      
      <!-- 加载消息 -->
      <div v-if="message || defaultMessage" class="loading-message">
        {{ message || defaultMessage }}
      </div>
    </div>
    
    <!-- 默认插槽作为背景内容 -->
    <div class="loading-background" v-if="!fullscreen">
      <slot></slot>
    </div>
  </div>
  
  <!-- 未加载时显示默认内容 -->
  <template v-else>
    <slot></slot>
  </template>
</template>

<style scoped>
/* 加载容器 */
.loading-container {
  position: relative;
  width: 100%;
  height: 100%;
  min-height: 4rem;
  display: flex;
  align-items: center;
  justify-content: center;
}

/* 全屏模式 */
.loading-container.fullscreen {
  position: fixed;
  top: 0;
  left: 0;
  width: 100vw;
  height: 100vh;
  background-color: rgba(255, 255, 255, 0.8);
}

/* 暗色主题全屏 */
.dark .loading-container.fullscreen {
  background-color: rgba(0, 0, 0, 0.6);
}

/* 透明背景 */
.loading-container.transparent-bg {
  background-color: transparent;
}

/* 加载内容 */
.loading-content {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  z-index: 1;
}

/* 加载指示器 */
.loading-indicator {
  margin-bottom: 0.75rem;
}

/* 加载消息 */
.loading-message {
  font-size: 0.875rem;
  color: var(--color-text-secondary, #4b5563);
  text-align: center;
}

.dark .loading-message {
  color: var(--color-text-secondary, #d1d5db);
}

/* 背景内容 */
.loading-background {
  position: absolute;
  top: 0;
  left: 0;
  width: 100%;
  height: 100%;
  filter: blur(1px) opacity(0.6);
  z-index: 0;
  pointer-events: none;
}

/* 加载动画样式已移至 animations.css 统一管理 */




</style> 