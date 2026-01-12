<script setup lang="ts">
import { ref, watch, h } from 'vue';
import RequestResponsePanel from '../common/RequestResponsePanel.vue';
import HttpTrafficTable from '../common/HttpTrafficTable.vue';
import type { ScanLogItem, RequestScanDetail } from '../../types/scanLog';
import { usePanelResize } from '../../composables/usePanelResize';
import { getModuleColor } from '../../utils/colors';
// @ts-ignore
import { GetRequestScanPBody } from "../../../bindings/github.com/yhy0/ChYing/app.js";

const props = defineProps<{
  items: ScanLogItem[];
  selectedItem: ScanLogItem | null;
  loading?: boolean;
}>();

const emit = defineEmits<{
  (e: 'select-item', item: ScanLogItem): void;
  (e: 'context-menu', event: MouseEvent, item: ScanLogItem): void;
  (e: 'set-color', item: ScanLogItem, color: string): void;
}>();

// 使用面板调整大小
const { panelHeight: tableHeight, startResize: startResizeTable } = usePanelResize({
  panelId: 'scan-log-table-height',
  initialHeight: 350,
  minHeight: 200,
  maxHeightOffset: 300 
});

// 请求和响应数据
const requestData = ref<string>('');
const responseData = ref<string>('');

// 获取请求响应详情
const fetchRequestDetail = async (id: number): Promise<RequestScanDetail | null> => {
  try {
    const result = await GetRequestScanPBody(id);
    console.log('ScanLogPanel - fetchRequestDetail result:', result);
    
    // 检查返回结果的结构
    if (result && result.data) {
      return result.data;
    } else if (result && (result as any).request_raw !== undefined) {
      // 如果直接返回数据结构
      return result as any;
    }
    
    console.warn('获取请求详情返回了意外的数据结构:', result);
    return null;
  } catch (error) {
    console.error('获取请求详情失败:', error);
    return null;
  }
};

// 监听选中项变化
watch(() => props.selectedItem, async (newItem) => {
  console.log('ScanLogPanel - selectedItem 变化:', newItem);
  
  if (newItem) {
    // 如果已有数据，直接使用
    if (newItem.request && newItem.response) {
      console.log('ScanLogPanel - 使用已有的请求响应数据');
      requestData.value = newItem.request;
      responseData.value = newItem.response;
    } else {
      // 按需加载请求响应详情
      console.log('ScanLogPanel - 开始获取请求响应详情，ID:', newItem.id);
      const detail = await fetchRequestDetail(newItem.id);
      
      if (detail) {
        console.log('ScanLogPanel - 成功获取请求响应详情:', detail);
        requestData.value = (detail as any).request_raw || '';
        responseData.value = (detail as any).response_raw || '';
        
        // 同时更新原始item的数据，避免重复请求
        newItem.request = (detail as any).request_raw || '';
        newItem.response = (detail as any).response_raw || '';
      } else {
        console.log('ScanLogPanel - 未能获取到请求响应详情');
        requestData.value = '';
        responseData.value = '';
      }
    }
  } else {
    console.log('ScanLogPanel - 清空请求响应数据');
    requestData.value = '';
    responseData.value = '';
  }
}, { immediate: true });

// 处理行选择
const handleSelectItem = (item: ScanLogItem) => {
  emit('select-item', item);
};

// 处理右键菜单
const handleContextMenu = (event: MouseEvent, item: ScanLogItem) => {
  emit('context-menu', event, item);
};

// 处理行颜色设置
const handleSetColor = (item: ScanLogItem, color: string) => {
  emit('set-color', item, color);
};

// 更新请求数据
const updateRequestData = (data: string) => {
  console.log('ScanLogPanel - 更新请求数据:', data.length, '字符');
  requestData.value = data;
};

// 更新响应数据
const updateResponseData = (data: string) => {
  console.log('ScanLogPanel - 更新响应数据:', data.length, '字符');
  responseData.value = data;
};

// 自定义列定义，与后端数据结构一致
const customColumns = [
  { id: 'id', name: 'ID', width: 60 },
  { 
    id: 'moduleName', 
    name: 'Module', 
    width: 140,
    cellRenderer: ({ item }: { item: ScanLogItem }) => {
      const moduleColor = getModuleColor(item.moduleName);
      return h('div', {
        class: 'module-badge',
        style: {
          backgroundColor: moduleColor.color,
          display: 'inline-flex',
          alignItems: 'center',
          padding: '2px 8px',
          borderRadius: '12px',
          fontSize: '11px',
          fontWeight: '500',
          border: `1px solid ${moduleColor.color}`,
          color: 'var(--color-text-primary)'
        }
      }, item.moduleName);
    }
  },
  { id: 'target', name: 'Target', width: 300 },
  { id: 'path', name: 'Path', width: 200 },
  { id: 'method', name: 'Method', width: 80 },
  { id: 'status', name: 'Status', width: 80 },
  { id: 'length', name: 'Length', width: 100 },
  { id: 'contentType', name: 'Content-Type', width: 150 },
  { id: 'title', name: 'Title', width: 200 },
  { id: 'ip', name: 'IP', width: 120 },
  { id: 'timestamp', name: 'Timestamp', width: 160 }
];
</script>

<template>
  <div class="scan-log-panel">
    <!-- 加载状态 -->
    <div v-if="loading" class="empty-state">
      <div class="empty-state-icon">
        <i class="bx bx-loader-alt bx-spin"></i>
      </div>
      <div class="empty-state-text">加载扫描日志中...</div>
    </div>
    
    <!-- 扫描日志表格 -->
    <div v-else class="scan-log-table-container" :style="{ height: tableHeight + 'px' }">
      <HttpTrafficTable 
        :items="items" 
        :selectedItem="selectedItem"
        :customColumns="customColumns"
        :key="items.length + '-scan-log-table'"
        tableId="scan-log"
        @select-item="handleSelectItem"
        @context-menu="handleContextMenu"
        @set-color="handleSetColor"
      />
    </div>

    <!-- 表格与详情之间的分隔线 -->
    <div 
      v-if="selectedItem"
      class="panel-divider-horizontal" 
      @mousedown="startResizeTable"
    >
      <div class="resize-line"></div>
    </div>
    
    <!-- 详情区域 -->
    <div v-if="selectedItem" class="detail-panel">
      <!-- 扫描信息摘要 -->
      <div class="detail-summary">
        <div class="summary-grid">
          <div class="summary-section">
            <div class="summary-item">
              <span class="summary-label">日志ID:</span>
              <code class="code-snippet">{{ selectedItem.id }}</code>
            </div>
            <div class="summary-item">
              <span class="summary-label">模块:</span>
              <span class="text-info">{{ selectedItem.moduleName }}</span>
            </div>
            <div class="summary-item">
              <span class="summary-label">目标URL:</span>
              <span class="text-truncate">{{ selectedItem.target }}</span>
            </div>
            <div v-if="selectedItem.vulnerability" class="summary-item">
              <span class="summary-label">漏洞类型:</span>
              <span class="text-danger">{{ selectedItem.vulnerability }}</span>
            </div>
          </div>
          <div class="summary-section">
            <div class="summary-item">
              <span class="summary-label">HTTP状态:</span>
              <code class="code-snippet">{{ selectedItem.status }}</code>
            </div>
            <div class="summary-item">
              <span class="summary-label">响应长度:</span>
              <span>{{ selectedItem.length }} bytes</span>
            </div>
            <div v-if="selectedItem.payload" class="summary-item">
              <span class="summary-label">载荷:</span>
              <code class="code-snippet">{{ selectedItem.payload }}</code>
            </div>
            <div v-if="selectedItem.description" class="summary-item">
              <span class="summary-label">描述:</span>
              <span class="text-secondary">{{ selectedItem.description }}</span>
            </div>
            <div v-if="selectedItem.evidence" class="summary-item">
              <span class="summary-label">证据:</span>
              <span class="text-secondary">{{ selectedItem.evidence }}</span>
            </div>
          </div>
        </div>
      </div>
      
      <!-- 请求响应面板 -->
      <RequestResponsePanel
        :request-data="requestData"
        :response-data="responseData"
        @update:request-data="updateRequestData"
        @update:response-data="updateResponseData"
        :request-read-only="true"
        :response-read-only="true"
        :server-duration-ms="selectedItem?.serverDurationMs || 0"
        :uuid="String(selectedItem?.id || '')"
      />
    </div>
  </div>
</template>

<style scoped>
.scan-log-panel {
  display: flex;
  flex-direction: column;
  height: 100%;
  background: var(--glass-bg-primary);
}

.scan-log-table-container {
  background: var(--glass-bg-card);
  border-bottom: 1px solid var(--glass-border-light);
  flex-shrink: 0;
  position: relative;
}

.detail-panel {
  flex: 1;
  display: flex;
  flex-direction: column;
  overflow: hidden;
  min-height: 200px;
  background: var(--glass-bg-secondary);
}

.detail-summary {
  padding: var(--spacing-md);
  background: var(--glass-bg-card);
  backdrop-filter: var(--glass-blur-subtle);
  -webkit-backdrop-filter: var(--glass-blur-subtle);
  border-bottom: 1px solid var(--glass-border-light);
}

/* 详情摘要网格样式 */
.summary-grid {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: var(--spacing-lg);
  font-size: var(--text-sm);
}

.summary-section {
  display: flex;
  flex-direction: column;
  gap: var(--spacing-sm);
  padding: var(--spacing-sm);
  background: var(--glass-bg-tertiary);
  border: 1px solid var(--glass-border-light);
  border-radius: var(--radius-md);
}

.summary-item {
  display: flex;
  align-items: flex-start;
  gap: var(--spacing-sm);
  padding: var(--spacing-xs) 0;
}

.summary-item:not(:last-child) {
  border-bottom: 1px solid var(--glass-border-light);
  padding-bottom: var(--spacing-sm);
}

.summary-label {
  font-weight: var(--font-weight-medium);
  color: var(--color-text-secondary);
  white-space: nowrap;
  flex-shrink: 0;
  min-width: 80px;
  font-size: var(--text-xs);
  text-transform: uppercase;
  letter-spacing: 0.03em;
}

.code-snippet {
  background: var(--glass-bg-hover);
  padding: 2px 6px;
  border-radius: var(--radius-xs);
  font-family: var(--font-mono);
  font-size: var(--text-xs);
  color: var(--color-primary);
  border: 1px solid var(--glass-border-light);
}

.text-danger {
  color: var(--color-danger);
  font-weight: var(--font-weight-semibold);
  background: rgba(239, 68, 68, 0.1);
  padding: 2px 8px;
  border-radius: var(--radius-sm);
}

.text-secondary {
  color: var(--color-text-secondary);
  font-size: var(--text-sm);
}

.text-truncate {
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  max-width: 300px;
}

.text-info {
  color: var(--color-info);
  font-weight: var(--font-weight-medium);
  background: rgba(59, 130, 246, 0.1);
  padding: 2px 8px;
  border-radius: var(--radius-sm);
}

/* 分隔线样式 */
.panel-divider-horizontal {
  user-select: none;
  height: 6px;
  background: var(--glass-bg-tertiary);
  border-top: 1px solid var(--glass-border-light);
  border-bottom: 1px solid var(--glass-border-light);
  cursor: row-resize;
  display: flex;
  align-items: center;
  justify-content: center;
  transition: background var(--glass-transition-fast);
}

.panel-divider-horizontal:hover {
  background: var(--glass-bg-hover);
}

.resize-line {
  width: 40px;
  height: 3px;
  background: var(--color-text-tertiary);
  border-radius: var(--radius-full);
  opacity: 0.5;
  transition: opacity var(--glass-transition-fast);
}

.panel-divider-horizontal:hover .resize-line {
  opacity: 1;
  background: var(--color-primary);
}

/* 空状态样式 */
.empty-state {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  height: 200px;
  gap: var(--spacing-md);
  color: var(--color-text-tertiary);
}

.empty-state-icon {
  font-size: 48px;
  opacity: 0.5;
}

.empty-state-icon i {
  color: var(--color-primary);
}

.empty-state-text {
  font-size: var(--text-sm);
  color: var(--color-text-secondary);
}

/* 响应式设计 */
@media (max-width: 768px) {
  .summary-grid {
    grid-template-columns: 1fr;
    gap: var(--spacing-md);
  }

  .summary-label {
    min-width: 60px;
  }

  .text-truncate {
    max-width: 200px;
  }
}
</style> 