<script setup lang="ts">
import { computed, ref, onMounted, watch, onBeforeUnmount as vueOnBeforeUnmount } from 'vue';
import { HttpRequestViewer } from '../codemirror';
import type { RequestViewType } from '../../../utils/viewerUtils';
import {
  convertToHex,
  copyToClipboard
} from '../../../utils';

// 定义props接收请求数据和配置
const props = defineProps<{
  normalizedRequestData: string;
  requestWidth: number;
  requestViewType: RequestViewType;
  readOnly?: boolean;
  requestTitle?: string;
  wordWrap?: boolean;
}>();

// 定义事件
const emit = defineEmits<{
  (e: 'set-request-view-type', type: string): void;
  (e: 'update:request-data', data: string): void;
  (e: 'toggle-word-wrap'): void;
}>();

// 请求标题
const computedRequestTitle = computed(() => props.requestTitle || 'Request');

// 处理请求视图类型切换
const setRequestViewType = (type: RequestViewType) => {
  emit('set-request-view-type', type);
};

// 处理请求更新
const handleRequestUpdate = (data: string) => {
  emit('update:request-data', data);
};

// 复制请求
const copyRequest = () => {
  copyToClipboard(props.normalizedRequestData);
};

// 控制标签滑块位置
const tabsContainerRef = ref<HTMLElement | null>(null);
const tabSliderRef = ref<HTMLElement | null>(null);
const activeTabRef = ref<HTMLElement | null>(null);

// 设置标签引用的函数
const setActiveTabRef = (el: any) => {
  if (el instanceof HTMLElement) {
    activeTabRef.value = el;
  }
};

// 更新滑块位置
const updateSliderPosition = () => {
  if (!tabSliderRef.value || !activeTabRef.value || !tabsContainerRef.value) return;
  
  const tabRect = activeTabRef.value.getBoundingClientRect();
  const containerRect = tabsContainerRef.value.getBoundingClientRect();
  
  // 计算相对位置
  const leftPosition = tabRect.left - containerRect.left;
  
  tabSliderRef.value.style.width = `${tabRect.width}px`;
  tabSliderRef.value.style.transform = `translateX(${leftPosition}px)`;
};

// 监听视图类型变化更新滑块
watch(() => props.requestViewType, () => {
  setTimeout(updateSliderPosition, 50);
}, { immediate: true });

// 组件挂载后初始化滑块位置
onMounted(() => {
  updateSliderPosition();
  // 监听窗口大小变化
  window.addEventListener('resize', updateSliderPosition);
});

// 在组件卸载前移除事件监听器
vueOnBeforeUnmount(() => {
  window.removeEventListener('resize', updateSliderPosition);
});
</script>

<template>
  <div class="relative flex flex-col" :style="{ width: requestWidth + '%' }">
    <!-- 标签栏 -->
    <div class="flex items-center border-b border-gray-200 dark:border-gray-700 bg-gray-50 dark:bg-gray-800 py-2 px-3">
      <h3 class="panel-title mr-3">
        <i class="bx bx-arrow-from-left panel-title-icon"></i>
        {{ computedRequestTitle }}
      </h3>
      
      <!-- 标签容器 -->
      <div class="view-tabs-container" ref="tabsContainerRef">
        <!-- 滑块背景 -->
        <div class="view-tab-slider" ref="tabSliderRef"></div>
        
        <!-- 标签按钮 -->
        <button 
          class="view-tab-button"
          :class="{ 'active': requestViewType === 'pretty' }"
          @click="setRequestViewType('pretty')"
          :ref="requestViewType === 'pretty' ? setActiveTabRef : undefined"
        >
          <i class="bx bx-code-alt view-tab-icon"></i>Pretty
        </button>
        
        <button 
          class="view-tab-button"
          :class="{ 'active': requestViewType === 'hex' }"
          @click="setRequestViewType('hex')"
          :ref="requestViewType === 'hex' ? setActiveTabRef : undefined"
        >
          <i class="bx bx-data view-tab-icon"></i>Hex
        </button>
      </div>
      
      <div class="ml-auto flex items-center">
        <button 
          class="action-button mr-2 tooltip-container"
          @click="emit('toggle-word-wrap')"
          :class="{ 'active': wordWrap }"
        >
          <i class="bx bx-text text-base"></i>
          <span class="tooltip-text">切换自动换行</span>
        </button>
        <button 
          class="action-button tooltip-container"
          @click="copyRequest"
        >
          <i class="bx bx-clipboard text-base"></i>
          <span class="tooltip-text">复制请求</span>
        </button>
      </div>
    </div>
    
    <!-- 编辑器容器 -->
    <div class="flex-1 overflow-hidden relative editor-container">
      <!-- 使用Transition组件添加切换动画 -->
      <transition name="view-transition" mode="out-in">
        <!-- Pretty 视图 - 使用 HttpRequestViewer -->
        <HttpRequestViewer 
          v-if="requestViewType === 'pretty'"
          :data="normalizedRequestData"
          :read-only="readOnly"
          @update:data="handleRequestUpdate"
          class="editor-view"
          :key="'pretty'"
          :class="{ 'word-wrap': wordWrap }"
        />
        
        <!-- Hex 视图 - 十六进制显示 -->
        <div v-else-if="requestViewType === 'hex'" class="h-full w-full overflow-auto p-4 pb-14" :key="'hex'">
          <pre class="font-mono text-xs text-gray-800 dark:text-gray-200 overflow-x-auto" :class="{ 'whitespace-pre-wrap': wordWrap, 'whitespace-pre': !wordWrap }">{{ convertToHex(normalizedRequestData) }}</pre>
        </div>
      </transition>
    </div>
  </div>
</template> 

<style scoped>
.panel-title {
  font-size: 0.85rem;
  font-weight: 500;
  color: var(--text-color-medium, #4b5563);
  display: flex;
  align-items: center;
}

.panel-title-icon {
  margin-right: 6px;
  font-size: 1rem;
}

.view-tabs-container {
  position: relative;
  display: flex;
  align-items: center;
  gap: 2px;
}

.view-tab-slider {
  position: absolute;
  bottom: -2px;
  left: 0;
  height: 2px;
  background-color: var(--primary-color, #4f46e5);
  transition: transform 0.2s ease, width 0.2s ease;
  z-index: 1;
}

.view-tab-button {
  position: relative;
  padding: 0.35rem 0.65rem;
  font-size: 0.8rem;
  border-radius: 0.25rem 0.25rem 0 0;
  color: var(--text-color-medium, #4b5563);
  background: transparent;
  border: none;
  cursor: pointer;
  transition: color 0.2s ease;
  display: flex;
  align-items: center;
  gap: 4px;
}

.view-tab-button:hover {
  color: var(--primary-color, #4f46e5);
}

.view-tab-button.active {
  color: var(--primary-color, #4f46e5);
  font-weight: 500;
}

.view-tab-icon {
  font-size: 0.95rem;
}

.action-button {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 26px;
  height: 26px;
  border-radius: 4px;
  color: var(--text-color-light, #6b7280);
  background: transparent;
  border: none;
  cursor: pointer;
  transition: all 0.2s ease;
}

.action-button:hover {
  color: var(--primary-color, #4f46e5);
  background-color: var(--bg-hover, rgba(79, 70, 229, 0.1));
}

.action-button.active {
  color: var(--primary-color, #4f46e5);
  background-color: var(--bg-hover, rgba(79, 70, 229, 0.1));
}

.action-button i {
  font-size: 0.95rem;
}

.editor-container {
  position: relative;
  height: 100%;
}

.editor-view {
  height: 100%;
  width: 100%;
  overflow: auto;
}

.view-transition-enter-active,
.view-transition-leave-active {
  transition: opacity 0.15s ease;
}

.view-transition-enter-from,
.view-transition-leave-to {
  opacity: 0;
}

.word-wrap :deep(.cm-content) {
  white-space: pre-wrap;
  word-break: break-word;
}

/* Tooltip 样式已移至 tooltip.css 统一管理 */
</style> 