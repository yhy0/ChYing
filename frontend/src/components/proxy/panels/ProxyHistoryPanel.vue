<script setup lang="ts">
import { ref, watch, computed } from 'vue';
import { RequestResponsePanel, HttpTrafficTable, type HttpTrafficItem } from '../../../components/common';
import type { ProxyHistoryItem } from '../../../types';
import { usePanelResize } from '../../../composables/usePanelResize';

const props = defineProps<{
  items: ProxyHistoryItem[];
  selectedItem: ProxyHistoryItem | null;
  enableMultiSelect?: boolean;
  checkedItems?: ProxyHistoryItem[];
}>();

const emit = defineEmits<{
  (e: 'select-item', item: ProxyHistoryItem): void;
  (e: 'context-menu', event: MouseEvent, item: ProxyHistoryItem): void;
  (e: 'update:requestData', data: string): void;
  (e: 'update:responseData', data: string): void;
  (e: 'set-color', item: ProxyHistoryItem, color: string): void;
  (e: 'update:checkedItems', items: ProxyHistoryItem[]): void;
}>();

// Use usePanelResize composable
const { panelHeight: tableHeight, startResize: startResizeTable } = usePanelResize({
  panelId: 'proxy-history-table-height',
  initialHeight: 200,
  minHeight: 100,
  maxHeightOffset: 200 
});

// Computed properties for request and response data
const requestData = ref<string>('');
const responseData = ref<string>('');

// 流量统计
const trafficStats = computed(() => {
  return { total: props.items.length };
});

// Watch for selected item changes
watch(() => props.selectedItem, (newItem) => {
  if (newItem) {
    requestData.value = newItem.request;
    responseData.value = newItem.response;
  } else {
    requestData.value = "";
    responseData.value = "";
  }
}, { immediate: true });

// Update request/response data
const updateRequestData = (data: string) => {
  requestData.value = data;
  emit('update:requestData', data);
};

const updateResponseData = (data: string) => {
  responseData.value = data;
  emit('update:responseData', data);
};

// Selection and analysis menu
const handleSelectItem = (item: HttpTrafficItem) => {
  emit('select-item', item as ProxyHistoryItem);
};

const handleContextMenu = (event: MouseEvent, item: HttpTrafficItem) => {
  emit('context-menu', event, item as ProxyHistoryItem);
};

// 处理行颜色设置
const handleSetColor = (item: HttpTrafficItem, color: string) => {
  emit('set-color', item as ProxyHistoryItem, color);
};

// 处理多选更新
const handleCheckedItemsUpdate = (items: HttpTrafficItem[]) => {
  emit('update:checkedItems', items as ProxyHistoryItem[]);
};
</script>

<template>
  <div class="proxy-history-panel flex flex-col h-full">
    <!-- 数据源控制栏 -->
    <div class="data-source-controls bg-white dark:bg-gray-800 border-b border-gray-200 dark:border-gray-700 p-3">
      <div class="flex items-center justify-between">
        <!-- 左侧：流量统计 -->
        <div class="flex items-center space-x-4">
          <div class="text-sm text-gray-600 dark:text-gray-300">
            <span class="font-medium">流量统计:</span>
          </div>
          <div class="flex items-center space-x-3">
            <div class="flex items-center">
              <div class="w-2 h-2 bg-blue-500 rounded-full mr-1"></div>
              <span class="text-xs text-gray-600 dark:text-gray-300">
                总计: {{ trafficStats.total }}
              </span>
            </div>
          </div>
        </div>
      </div>
    </div>

    <!-- 可选插槽 -->
    <slot></slot>
    
    <!-- 历史表格 - 修改为固定高度容器 -->
    <div class="history-table-container" :style="{ height: tableHeight + 'px' }">
      <div class="h-full overflow-auto pl-2">
        <HttpTrafficTable
          :items="items"
          :selectedItem="selectedItem"
          :key="items.length + '-proxy-history-table'"
          tableId="proxy-history"
          tableClass="compact-table"
          :enableMultiSelect="enableMultiSelect"
          :checkedItems="checkedItems"
          @select-item="handleSelectItem"
          @context-menu="handleContextMenu"
          @set-color="handleSetColor"
          @update:checkedItems="handleCheckedItemsUpdate"
        />
      </div>
    </div>

    <!-- 移到了这里 - 表格与详情之间的分隔线 -->
    <div class="panel-divider-horizontal cursor-ns-resize" @mousedown="startResizeTable"></div>
    
    <!-- 详情区域 - 改为flex-1让它填充剩余空间 -->
    <div v-if="selectedItem" class="flex-1 flex flex-col overflow-hidden">
      <!-- 加载指示器 -->
      <div v-if="selectedItem.isLoading" class="flex items-center justify-center p-4 text-gray-500 dark:text-gray-400">
        <i class="bx bx-loader-alt bx-spin mr-2"></i>
        <span>正在加载请求和响应数据...</span>
      </div>
      
      <RequestResponsePanel
        :requestData="requestData"
        :responseData="responseData"
        @update:requestData="updateRequestData"
        @update:responseData="updateResponseData"
        :requestReadOnly="true"
        :responseReadOnly="true"
        :serverDurationMs="selectedItem?.serverDurationMs || 0"
      />
    </div>
  </div>
</template>

<style scoped>
.data-source-controls {
  flex-shrink: 0;
}

.history-table-container {
  flex-shrink: 0;
  overflow: hidden;
}
</style>
