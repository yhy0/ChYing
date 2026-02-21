<script setup lang="ts">
import { ref, watch, computed } from 'vue';
import { RequestResponsePanel, HttpTrafficTable, type HttpTrafficItem } from '../../components/common';
import type {ProxyHistoryItem} from "../../types";
import { usePanelResize } from '../../composables/usePanelResize';

// 初始化面板调整组合式函数
const { panelHeight, startResize } = usePanelResize({ panelId: 'intruder-history', initialHeight: 300 });

// 定义Intruder结果项的接口
export interface IntruderResult {
  id: number;
  payload: string[];
  status: number;
  length: number;
  timeMs: number;
  timestamp: string;
  request: string;
  response: string;
  selected?: boolean;
  color?: string;
  [key: string]: any;
}

const props = defineProps<{
  results: IntruderResult[];
  selectedResult: IntruderResult | null;
  request?: string;
  response?: string;
}>();

const emit = defineEmits<{
  (e: 'select-result', item: IntruderResult): void;
  (e: 'context-menu', event: MouseEvent, item: IntruderResult): void;
  (e: 'set-color', item: ProxyHistoryItem, color: string): void;
}>();

// 从结果数据中提取HTTP方法
const extractMethodFromRequest = (request: string): string => {
  if (!request) return 'GET';
  
  try {
    const lines = request.split('\\n');
    const firstLine = lines[0] || '';
    const parts = firstLine.split(' ');
    if (parts.length > 0) {
      return parts[0];
    }
  } catch (e) {
    // 保持默认值
  }
  
  return 'GET';
};

// 计算结果中最大的payload数量
const maxPayloads = computed(() => {
  return props.results.reduce((max, result) => {
    const payloadLength = Array.isArray(result.payload) ? result.payload.length : 0;
    return Math.max(max, payloadLength);
  }, 0);
});

// 动态生成列配置
const dynamicColumns = computed(() => {
  // 分离 ID 列和其他基础列
  const idColumn = { id: 'id', name: '#', width: 60 };
  const otherBaseColumns = [
    { id: 'status', name: 'Status', width: 80 },
    { id: 'length', name: 'Length', width: 100 },
    { id: 'timeMs', name: 'Time (ms)', width: 100 },
    { id: 'timestamp', name: 'Timestamp', width: 150 } // 增加时间戳宽度
  ];

  const payloadColumns: Array<{ id: string; name: string; width: number }> = [];
  for (let i = 1; i <= maxPayloads.value; i++) {
    payloadColumns.push({
      id: `payload#${i}`,
      name: `Payload #${i}`,
      width: 150 // 可以根据需要调整默认宽度
    });
  }

  // 按 #, Payload #N..., 其他基础列 的顺序组合
  return [idColumn, ...payloadColumns, ...otherBaseColumns];
});

// 获取原始IntruderResult
const getOriginalResult = (item: HttpTrafficItem): IntruderResult | null => {
  return (item as any)._original as IntruderResult || null;
};

// 将Intruder结果转换为HttpTrafficItem格式
const trafficItems = computed<HttpTrafficItem[]>(() => {
  // 使用解构赋值确保数据变化触发重新计算
  return [...props.results].map(result => {
    // 使用提取方法获取HTTP方法
    const method = extractMethodFromRequest(result.request);
    
    // 安全的类型转换
    const id = typeof result.id === 'number' ? result.id : Number(result.id || 0);
    const timestamp = typeof result.timestamp === 'string' 
      ? result.timestamp 
      : (typeof result.timestamp === 'number' 
        ? new Date(result.timestamp).toISOString() 
        : new Date().toISOString());
    
    // 创建基础HttpTrafficItem
    const trafficItem: HttpTrafficItem = {
      id: id,
      method: method,
      url: '',
      host: '',
      path: '',
      status: Number(result.status || 0),
      length: Number(result.length || 0),
      mimeType: '',
      extension: '',
      title: '',
      ip: '',
      note: '',
      timestamp: timestamp,
      selected: result.selected,
      color: result.color,
      timeMs: Number(result.timeMs || 0),
      // 添加原始数据的引用，以便在选择时能够获取
      _original: result
    };
    
    // 动态添加 payload 列
    if (Array.isArray(result.payload)) {
        for (let i = 0; i < result.payload.length; i++) {
            const columnName = `payload#${i + 1}`;
            trafficItem[columnName] = result.payload[i] || '';
        }
    }

    return trafficItem;
  });
});

// Computed properties for request and response data
const requestData = ref<string>("");
const responseData = ref<string>("");

// 监听从外部传入的请求和响应数据
watch(() => props.request, (newRequest) => {
  if (newRequest) {
    requestData.value = newRequest;
  }
}, { immediate: true });

watch(() => props.response, (newResponse) => {
  if (newResponse) {
    responseData.value = newResponse;
  }
}, { immediate: true });

// Watch for selected result changes
watch(() => props.selectedResult, (newResult) => {
  if (newResult) {
    // 只有在没有外部数据时才使用选中结果中的数据
    if (!props.request) {
      requestData.value = newResult.request;
    }
    if (!props.response) {
      responseData.value = newResult.response;
    }
  } else {
    // 如果没有选中项且没有外部数据，则清空显示
    if (!props.request) {
      requestData.value = "";
    }
    if (!props.response) {
      responseData.value = "";
    }
  }
}, { immediate: true });

// Selection and analysis menu
const handleSelectItem = (item: HttpTrafficItem) => {
  const originalResult = getOriginalResult(item);
  if (originalResult) {
    emit('select-result', originalResult);
  } else {
    console.error('无法获取原始结果数据');
  }
};

const handleContextMenu = (event: MouseEvent, item: HttpTrafficItem) => {
  const originalResult = getOriginalResult(item);
  if (originalResult) {
    emit('context-menu', event, originalResult);
  }
};

// 处理行颜色设置
const handleSetColor = (item: HttpTrafficItem, color: string) => {
  const originalResult = getOriginalResult(item);
  if (originalResult) {
    emit('set-color', originalResult as any, color);
  }
};

// 存储当前选中项的 HttpTrafficItem 版本
const selectedTrafficItem = computed(() => {
  if (!props.selectedResult) return null;
  
  // 在 trafficItems 中查找匹配的项
  return trafficItems.value.find(item => item.id === props.selectedResult?.id) || null;
});
</script>

<template>
  <div class="flex-1 flex flex-col overflow-hidden">
    <!-- Table section with resizable height -->
    <div class="relative" :style="{ height: panelHeight + 'px' }">
      <div class="absolute inset-0 overflow-auto">
        <HttpTrafficTable
          :key="'intruder-table-dynamic-' + maxPayloads"
          :items="trafficItems"
          :selectedItem="selectedTrafficItem"
          :customColumns="dynamicColumns"
          tableId="intruder-results"
          :tableClass="'compact-table'"
          @select-item="handleSelectItem"
          @context-menu="handleContextMenu"
          @set-color="handleSetColor"
        />
      </div>
      <!-- Resizer handle -->
      <div 
        class="absolute bottom-0 left-0 right-0 h-1 bg-gray-200 dark:bg-gray-700 cursor-ns-resize hover:bg-indigo-500 dark:hover:bg-indigo-400"
        @mousedown="startResize"
      ></div>
    </div>

    <!-- Editors section -->
    <RequestResponsePanel
      v-if="selectedResult"
      :requestData="requestData || ''"
      :responseData="responseData || ''"
      :title="{
        request: 'Request',
        response: 'Response'
      }"
      :requestReadOnly="true"
      :responseReadOnly="true"
      :serverDurationMs="props.selectedResult?.timeMs || 0"
    />
  </div>
</template> 