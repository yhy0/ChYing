<script setup lang="ts">
import { ref, inject, onMounted, onBeforeUnmount, watch } from 'vue';
import type { Ref } from 'vue';
import type { SiteMapNode, SecurityIssue } from '../../../types/project';
import SiteMapTreePanel from './SiteMapTreePanel.vue';
import ContentsPanel from './ContentsPanel.vue';
import SecurityIssuesPanel from './SecurityIssuesPanel.vue';
import { useProjectStore } from '../../../store/project';

// 注入共享的数据
const securityIssues = inject('securityIssues') as Ref<SecurityIssue[]>;
const getSeverityColorClass = inject('getSeverityColorClass') as (severity: string) => string;
const getSeverityText = inject('getSeverityText') as (severity: string) => string;
const getMethodColor = inject('getMethodColor') as (method: string) => string;
const getStatusColor = inject('getStatusColor') as (status: number) => string;
const formatBytes = inject('formatBytes') as (bytes: number, decimals?: number) => string;

// 注入更新主机数据的函数
const updateSiteMapHosts = inject('updateSiteMapHosts', (hosts: string[]) => {}) as (hosts: string[]) => void;

// 使用 ProjectStore
const projectStore = useProjectStore();

// 面板宽度管理
const getPanelWidths = () => {
  try {
    const saved = localStorage.getItem('siteMapPanelWidths');
    return saved ? JSON.parse(saved) : {
      leftPanelWidth: 20,
      middlePanelWidth: 50,
      rightPanelWidth: 30
    };
  } catch {
    return { leftPanelWidth: 20, middlePanelWidth: 50, rightPanelWidth: 30 };
  }
};

const { leftPanelWidth, middlePanelWidth, rightPanelWidth } = getPanelWidths();
const leftWidth = ref(leftPanelWidth);
const middleWidth = ref(middlePanelWidth);
const rightWidth = ref(rightPanelWidth);

// 当前选中的节点数据
const selectedNode = ref<SiteMapNode | null>(null);
const requestData = ref('');
const responseData = ref('');

// 处理节点选择
const handleNodeSelected = (node: SiteMapNode, request: string = '', response: string = '') => {
  selectedNode.value = node;
  requestData.value = request;
  responseData.value = response;
};

// 容器引用
const containerRef = ref<HTMLElement | null>(null);

// 左侧面板调整
let isDraggingLeft = false;
let initialXLeft = 0;
let initialLeftWidth = 0;

// 右侧面板调整
let isDraggingRight = false;
let initialXRight = 0;
let initialMiddleWidth = 0;

// 开始调整左侧面板
const startResizeLeft = (e: MouseEvent) => {
  e.preventDefault();
  isDraggingLeft = true;
  initialXLeft = e.clientX;
  
  if (containerRef.value) {
    const containerRect = containerRef.value.getBoundingClientRect();
    initialLeftWidth = (leftWidth.value / 100) * containerRect.width;
  }
  
  document.addEventListener('mousemove', handleMouseMoveLeft);
  document.addEventListener('mouseup', stopResizeLeft);
  document.body.classList.add('dragging');
};

// 左侧拖拽过程
const handleMouseMoveLeft = (e: MouseEvent) => {
  if (!isDraggingLeft || !containerRef.value) return;
  
  const containerRect = containerRef.value.getBoundingClientRect();
  const deltaX = e.clientX - initialXLeft;
  const newLeftPanelWidth = initialLeftWidth + deltaX;
  
  // 计算百分比宽度
  const percentWidth = (newLeftPanelWidth / containerRect.width) * 100;
  
  // 限制宽度范围（10% - 40%）
  const newLeftWidth = Math.max(10, Math.min(40, percentWidth));
  
  // 更新左侧宽度和中间宽度
  leftWidth.value = newLeftWidth;
  middleWidth.value = 100 - newLeftWidth - rightWidth.value;
};

// 停止左侧拖拽
const stopResizeLeft = () => {
  isDraggingLeft = false;
  document.removeEventListener('mousemove', handleMouseMoveLeft);
  document.removeEventListener('mouseup', stopResizeLeft);
  document.body.classList.remove('dragging');
  
  // 保存宽度到本地存储
  localStorage.setItem('siteMapPanelWidths', JSON.stringify({
    leftPanelWidth: leftWidth.value,
    middlePanelWidth: middleWidth.value,
    rightPanelWidth: rightWidth.value
  }));
};

// 开始调整右侧面板
const startResizeRight = (e: MouseEvent) => {
  e.preventDefault();
  isDraggingRight = true;
  initialXRight = e.clientX;
  
  if (containerRef.value) {
    const containerRect = containerRef.value.getBoundingClientRect();
    initialMiddleWidth = (middleWidth.value / 100) * containerRect.width;
  }
  
  document.addEventListener('mousemove', handleMouseMoveRight);
  document.addEventListener('mouseup', stopResizeRight);
  document.body.classList.add('dragging');
};

// 右侧拖拽过程
const handleMouseMoveRight = (e: MouseEvent) => {
  if (!isDraggingRight || !containerRef.value) return;
  
  const containerRect = containerRef.value.getBoundingClientRect();
  const deltaX = e.clientX - initialXRight;
  const newMiddlePanelWidth = initialMiddleWidth + deltaX;
  
  // 计算百分比宽度
  const percentWidth = (newMiddlePanelWidth / containerRect.width) * 100;
  
  // 中间面板允许的最大宽度 = 100% - 左侧宽度 - 右侧最小宽度(10%)
  const maxMiddleWidth = 100 - leftWidth.value - 10;
  
  // 限制宽度范围
  const newMiddleWidth = Math.max(20, Math.min(maxMiddleWidth, percentWidth));
  
  // 更新中间宽度和右侧宽度
  middleWidth.value = newMiddleWidth;
  rightWidth.value = 100 - leftWidth.value - newMiddleWidth;
};

// 停止右侧拖拽
const stopResizeRight = () => {
  isDraggingRight = false;
  document.removeEventListener('mousemove', handleMouseMoveRight);
  document.removeEventListener('mouseup', stopResizeRight);
  document.body.classList.remove('dragging');
  
  // 保存宽度到本地存储
  localStorage.setItem('siteMapPanelWidths', JSON.stringify({
    leftPanelWidth: leftWidth.value,
    middlePanelWidth: middleWidth.value,
    rightPanelWidth: rightWidth.value
  }));
};

// 组件挂载时清理可能存在的事件监听器
onMounted(() => {
  document.removeEventListener('mousemove', handleMouseMoveLeft);
  document.removeEventListener('mouseup', stopResizeLeft);
  document.removeEventListener('mousemove', handleMouseMoveRight);
  document.removeEventListener('mouseup', stopResizeRight);
  document.body.classList.remove('dragging');

  // 检查是否已有数据，如果有，手动更新一次主机列表
  const siteMapData = projectStore.siteMapData;
  if (siteMapData && siteMapData.length > 0) {
    const hosts = siteMapData.map(node => node.name);
    updateSiteMapHosts(hosts);
  }
});

// 组件卸载时清理事件监听器
onBeforeUnmount(() => {
  document.removeEventListener('mousemove', handleMouseMoveLeft);
  document.removeEventListener('mouseup', stopResizeLeft);
  document.removeEventListener('mousemove', handleMouseMoveRight);
  document.removeEventListener('mouseup', stopResizeRight);
  document.body.classList.remove('dragging');
});

// 监听站点地图数据变化，更新主机信息
watch(() => projectStore.siteMapData, (newSiteMapData) => {
  if (newSiteMapData && newSiteMapData.length > 0) {
    // 从站点地图数据中提取主机列表
    const hosts = newSiteMapData.map(node => node.name);

    // 确保提取的是有效的主机名
    const validHosts = hosts.filter(host => host && typeof host === 'string' && host.length > 0);

    // 调用父组件注入的函数更新主机数据
    if (validHosts.length > 0) {
      updateSiteMapHosts(validHosts);
    }
  }
}, { immediate: true, deep: true });

// 导出更新主机方法
defineExpose({ updateHost: (host: string) => {} });
</script>

<template>
  <!-- 站点地图集成视图：占满剩余空间 -->
  <div class="h-full flex flex-col bg-white dark:bg-gray-900" ref="containerRef">
    <!-- 主内容区域 - 添加底部间距 -->
    <div class="flex flex-1 sitemap-container pb-20"> <!-- 添加pb-5类为页脚留出空间 -->
      <!-- 左侧站点地图面板 -->
      <div 
        class="h-full flex-shrink-0 bg-white dark:bg-gray-900 border-r border-gray-200 dark:border-gray-800"
        :style="{ width: `${leftWidth}%` }"
      >
        <SiteMapTreePanel
          class="h-full w-full"
          @node-selected="handleNodeSelected"
        />
      </div>
      
      <!-- 左侧调整器 -->
      <div 
        class="panel-divider"
        @mousedown="startResizeLeft"
      ></div>
      
      <!-- 中间 HTTP 历史面板 -->
      <div 
        class="h-full flex-shrink-0 bg-white dark:bg-gray-900 border-r border-gray-200 dark:border-gray-800"
        :style="{ width: `${middleWidth}%` }"
      >
        <ContentsPanel
          class="h-full w-full"
          :selected-node="selectedNode"
          :format-bytes="formatBytes"
          :get-method-color="getMethodColor"
          :get-status-color="getStatusColor"
          :request-data="requestData"
          :response-data="responseData"
        />
      </div>
      
      <!-- 右侧调整器 -->
      <div 
        class="panel-divider"
        @mousedown="startResizeRight"
      ></div>
      
      <!-- 右侧安全问题面板 -->
      <div 
        class="h-full flex-shrink-0 bg-white dark:bg-gray-900"
        :style="{ width: `${rightWidth}%` }"
      >
        <SecurityIssuesPanel
          class="h-full w-full"
          :security-issues="securityIssues"
          :selected-node="selectedNode"
          :get-severity-color-class="getSeverityColorClass"
          :get-severity-text="getSeverityText"
        />
      </div>
    </div>
  </div>
</template>

<style scoped>
/* 项目特定样式 */
.sitemap-container {
  position: relative;
  overflow: hidden;
}

/* 面板样式 */
.panel-wrapper {
  position: relative;
  height: 100%;
  display: flex;
  flex-direction: column;
  min-height: 0;
  flex-shrink: 0;
  background-color: var(--bg-primary);
}

.panel-content {
  flex: 1;
  min-height: 0;
  height: 100%;
  overflow: hidden;
}

/* 使用panel-divider.css中的样式 */
</style> 