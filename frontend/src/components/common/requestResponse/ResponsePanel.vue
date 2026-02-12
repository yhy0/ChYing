<script setup lang="ts">
import { computed, ref, onMounted, watch, onBeforeUnmount } from 'vue';
import { HttpResponseViewer } from '../codemirror';
import type { ResponseViewType } from '../../../utils';
import {
  convertToHex, 
  isHtmlContent, 
  sanitizeHtml, 
  getResponseContentType,
  copyToClipboard,
  extractHeadersAndBody
} from '../../../utils';

// 定义props接收响应数据和配置
const props = defineProps<{
  normalizedResponseData: string;
  responseWidth: number;
  responseViewType: ResponseViewType;
  readOnly?: boolean;
  responseTitle?: string;
  serverDurationMs: number; // 新增时间参数
  wordWrap?: boolean;
}>();

// 定义事件
const emit = defineEmits<{
  (e: 'set-response-view-type', type: string): void;
  (e: 'update:response-data', data: string): void;
  (e: 'toggle-word-wrap'): void;
}>();

// 响应标题和时间
const computedResponseTitle = computed(() => props.responseTitle || 'Response');

// 处理响应视图类型切换
const setResponseViewType = (type: ResponseViewType) => {
  emit('set-response-view-type', type);
};

// 处理响应更新
const handleResponseUpdate = (data: string) => {
  emit('update:response-data', data);
};

// 复制响应
const copyResponse = () => {
  copyToClipboard(props.normalizedResponseData);
};

// Body 格式化状态
const bodyFormatted = ref(true);

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
watch(() => props.responseViewType, () => {
  setTimeout(updateSliderPosition, 50);
}, { immediate: true });

// 组件挂载后初始化滑块位置
onMounted(() => {
  updateSliderPosition();
  // 监听窗口大小变化
  window.addEventListener('resize', updateSliderPosition);
});

// 在组件卸载前移除事件监听器
onBeforeUnmount(() => {
  window.removeEventListener('resize', updateSliderPosition);
});
</script>

<template>
  <div class="relative flex flex-col" :style="{ width: responseWidth + '%' }">
    <!-- 标签栏 -->
    <div class="flex items-center border-b border-gray-200 dark:border-gray-700 bg-gray-50 dark:bg-gray-800 py-2 px-3">
      <h3 class="panel-title mr-3">
        <i class="bx bx-arrow-to-right panel-title-icon"></i>
        {{ computedResponseTitle }}
        <span v-if="serverDurationMs > 0" class="response-time ml-2">
          ({{ serverDurationMs }}ms)
        </span>
      </h3>
      
      <!-- 标签容器 -->
      <div class="view-tabs-container" ref="tabsContainerRef">
        <!-- 滑块背景 -->
        <div class="view-tab-slider" ref="tabSliderRef"></div>
        
        <!-- 标签按钮 -->
        <button 
          class="view-tab-button"
          :class="{ 'active': responseViewType === 'pretty' }"
          @click="setResponseViewType('pretty')"
          :ref="responseViewType === 'pretty' ? setActiveTabRef : undefined"
        >
          <i class="bx bx-code-alt view-tab-icon"></i>Pretty
        </button>
      
        <button 
          class="view-tab-button"
          :class="{ 'active': responseViewType === 'hex' }"
          @click="setResponseViewType('hex')"
          :ref="responseViewType === 'hex' ? setActiveTabRef : undefined"
        >
          <i class="bx bx-data view-tab-icon"></i>Hex
        </button>
        
        <button 
          class="view-tab-button"
          :class="{ 'active': responseViewType === 'render' }"
          @click="setResponseViewType('render')"
          :ref="responseViewType === 'render' ? setActiveTabRef : undefined"
        >
          <i class="bx bx-window view-tab-icon"></i>Render
        </button>
      </div>
      
      <div class="ml-auto flex items-center">
        <button 
          class="action-button mr-2 tooltip-container"
          @click="bodyFormatted = !bodyFormatted"
          :class="{ 'active': bodyFormatted }"
        >
          <i class="bx bx-code-curly text-base"></i>
          <span class="tooltip-text">{{ bodyFormatted ? '显示原始格式' : '格式化 Body' }}</span>
        </button>
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
          @click="copyResponse"
        >
          <i class="bx bx-clipboard text-base"></i>
          <span class="tooltip-text">复制响应</span>
        </button>
      </div>
    </div>
    
    <!-- 编辑器容器 -->
    <div class="flex-1 overflow-hidden relative editor-container">
      <!-- 使用Transition组件添加切换动画 -->
      <transition name="view-transition" mode="out-in">
        <!-- Pretty 视图 - 使用 HttpResponseViewer -->
        <HttpResponseViewer 
          v-if="responseViewType === 'pretty'"
          :data="normalizedResponseData"
          :read-only="readOnly"
          :body-formatted="bodyFormatted"
          @update:data="handleResponseUpdate"
          class="editor-view"
          :key="'pretty'"
          :class="{ 'word-wrap': wordWrap }"
        />
        
        <!-- Hex 视图 - 十六进制显示 -->
        <div v-else-if="responseViewType === 'hex'" class="h-full w-full overflow-auto p-4 pb-14" :key="'hex'">
          <pre class="font-mono text-xs text-gray-800 dark:text-gray-200 overflow-x-auto" :class="{ 'whitespace-pre-wrap': wordWrap, 'whitespace-pre': !wordWrap }">{{ convertToHex(normalizedResponseData) }}</pre>
        </div>
        
        <!-- Render 视图 - 渲染HTML内容 -->
        <div v-else-if="responseViewType === 'render'" class="h-full w-full overflow-auto bg-white dark:bg-gray-900 p-2 pb-12" :key="'render'">
          <div v-if="isHtmlContent(getResponseContentType(extractHeadersAndBody(normalizedResponseData).headers), extractHeadersAndBody(normalizedResponseData).body)" class="render-container border border-gray-200 dark:border-gray-700 h-full w-full overflow-auto">
            <iframe 
              sandbox="allow-same-origin" 
              class="w-full h-full"
              :srcdoc="sanitizeHtml(extractHeadersAndBody(normalizedResponseData).body)" 
            ></iframe>
          </div>
          <div v-else class="flex items-center justify-center h-full">
            <p class="text-gray-500 dark:text-gray-400">该内容无法渲染，请尝试其他视图模式</p>
          </div>
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

.response-time {
  font-size: 0.7rem;
  color: var(--text-color-light, #6b7280);
  font-weight: normal;
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

.tooltip-container {
  position: relative;
}

.tooltip-text {
  visibility: hidden;
  position: absolute;
  bottom: -30px;
  left: 50%;
  transform: translateX(-50%);
  background-color: rgba(0, 0, 0, 0.75);
  color: #fff;
  text-align: center;
  border-radius: 4px;
  padding: 4px 8px;
  font-size: 0.7rem;
  white-space: nowrap;
  z-index: 100;
  pointer-events: none;
  opacity: 0;
  transition: opacity 0.2s;
}

.tooltip-container:hover .tooltip-text {
  visibility: visible;
  opacity: 1;
}
</style>