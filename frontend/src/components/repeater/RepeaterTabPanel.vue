<script setup lang="ts">
import { ref, watch, onMounted } from 'vue';
import { RequestResponsePanel } from '../common';
import { useI18n } from 'vue-i18n';
import { RawRequest } from "../../../bindings/github.com/yhy0/ChYing/app.js";

// 定义历史记录接口
interface RequestHistory {
  id: number; // 时间戳ID
  sequenceId: number; // 自增序列ID
  timestamp: number;
  method: string; 
  url: string;
  request: string;
  response: string | null;
  statusCode?: number;
  server_duration_ms: number;
  statusText?: string;
}

interface RepeaterTab {
  id: string;
  name: string;
  color: string;
  groupId: string | null;
  request: string;
  response: string | null;
  isActive: boolean;
  modified: boolean;
  serverDurationMs: number;
  history?: RequestHistory[]; // 添加历史记录数组
  method?: string; // 新增字段
  url?: string; // 新增字段
}

const props = defineProps<{
  tab: RepeaterTab;
}>();

const emit = defineEmits<{
  (e: 'update-request', data: string): void;
  (e: 'update-response', data: string | null): void;
  (e: 'update-history', history: RequestHistory[]): void; // 添加更新历史的事件
  (e: 'update-server-duration', duration: number): void; // 新增事件
  (e: 'update-url', url: string): void; // 新增更新 URL 的事件
}>();

const { t } = useI18n();
const isLoading = ref(false);
const requestMethod = ref('GET');
const requestUrl = ref('');
const history = ref<RequestHistory[]>(props.tab.history || []);
const selectedHistoryId = ref<number | null>(null);
const historyCounter = ref(0); // 添加历史记录计数器

// 提取状态码和状态文本的辅助函数
const extractStatusInfo = (responseStr: string | null): { statusCode: number, statusText: string } => {
  if (!responseStr) {
    return { statusCode: 0, statusText: 'No Response' };
  }
  
  const firstLine = responseStr.split('\n')[0] || '';
  const match = firstLine.match(/HTTP\/\d\.\d\s+(\d+)\s+(.+)/);
  
  if (match) {
    return {
      statusCode: parseInt(match[1], 10),
      statusText: match[2].trim()
    };
  }
  
  return { statusCode: 0, statusText: 'Unknown' };
};

// 更强健的解析请求方法和URL的函数
const parseRequestHeadersForMethodAndUrl = (requestStr: string) => {
  // 先分离 headers 和 body
  const parts = requestStr.split('\n\n');
  const headers = parts[0] || '';
  
  // 解析第一行来获取方法和URL
  const lines = headers.split('\n');
  const firstLine = lines[0] || '';
  const match = firstLine.match(/^(\w+)\s+(.+)\s+HTTP\/\d\.\d$/);
  
  if (match) {
    const method = match[1];
    let url = match[2];
    
    // 处理相对URL
    if (!url.startsWith('http')) {
      // 从headers提取host
      const hostLine = lines.find(line => line.toLowerCase().startsWith('host:'));
      
      if (hostLine) {
        const host = hostLine.split(':')[1]?.trim() || '';
        url = `https://${host}${url.startsWith('/') ? '' : '/'}${url}`;
      }
    }
    return { method, url };
  }
  return { method: 'GET', url: '' };
};

// 当tab变化时更新请求方法和URL和历史记录
watch(() => props.tab, (newTab, oldTab) => {
  // 确保新旧标签不同时才更新（避免重复渲染）
  if (!oldTab || newTab.id !== oldTab.id) {
    // 优先使用传入的 method 和 url，如果没有则从请求头解析
    if (newTab.method && newTab.url) {
      requestMethod.value = newTab.method;
      requestUrl.value = newTab.url;
    } else {
      const { method, url } = parseRequestHeadersForMethodAndUrl(newTab.request);
      requestMethod.value = method;
      // 只有在 tab 没有保存 URL 时才使用解析出的 URL
      if (!newTab.url) {
        requestUrl.value = url;
      } else {
        requestUrl.value = newTab.url;
      }
    }
    history.value = newTab.history || [];
    selectedHistoryId.value = null;
  }
}, { immediate: true, deep: true });

// 处理请求更新
const updateRequest = (newData: string) => {
  emit('update-request', newData);
  
  // 当请求数据更新时，只更新方法，不更新URL（保持用户手动设置的URL）
  const { method } = parseRequestHeadersForMethodAndUrl(newData);
  requestMethod.value = method;
};

// 处理响应更新
const updateResponse = (newData: string) => {
  emit('update-response', newData);
};

// 添加到历史记录
const addToHistory = (request: string, response: string | null, server_duration_ms: number) => {
  const { method, url } = parseRequestHeadersForMethodAndUrl(request);
  const { statusCode, statusText } = extractStatusInfo(response);
  
  // 更新服务器响应时间并添加日志
  emit('update-server-duration', server_duration_ms);
  
  // 自增历史记录序列ID
  historyCounter.value++;
  
  const historyItem: RequestHistory = {
    id: Date.now(),
    sequenceId: historyCounter.value,
    timestamp: Date.now(),
    method,
    url,
    request,
    response,
    statusCode,
    statusText,
    server_duration_ms
  };

  history.value = [historyItem, ...history.value];
  emit('update-history', history.value);
  
  // 选中新添加的历史记录
  selectedHistoryId.value = historyItem.id;
};

// 发送请求
const sendRequest = async () => {
  isLoading.value = true;
  selectedHistoryId.value = null; // 重置选择的历史记录
  console.log('Sending request:', props.tab.request);
  console.log('URL:', requestUrl.value);
  console.log('h:', history.value);
  console.log('s:', selectedHistoryId.value);

  RawRequest(props.tab.request, requestUrl.value, props.tab.id).then(result => {
    console.log(result);
    if (result && result.error !== "") {
      console.log(result.error)
      updateResponse("error: " + result.error);
      // 添加到历史记录
      addToHistory(props.tab.request, "error: " + result.error, result.data?.server_duration_ms || 0);
    } else if (result && result.data) {
      updateResponse(result.data.response_raw || "");
      // 添加到历史记录
      addToHistory(props.tab.request, result.data.response_raw || "", result.data.server_duration_ms || 0);
    } else {
      updateResponse("error: No response received");
      addToHistory(props.tab.request, "error: No response received", 0);
    }
    isLoading.value = false;
  })
};

// 清除响应
const clearResponse = () => {
  emit('update-response', null);
};

// 更新请求头中的方法和URL
const updateMethodAndUrl = () => {
  // 更新请求
  updateRequest(props.tab.request);
};

// 更新URL
const handleUrlChange = (event: Event) => {
  const target = event.target as HTMLInputElement;
  requestUrl.value = target.value;
  emit('update-url', requestUrl.value);
  updateMethodAndUrl();
};

// 选择历史记录项
const selectHistoryItem = (historyId: number) => {
  const item = history.value.find(h => h.id === historyId);
  if (!item) return;
  
  // 更新请求和响应
  emit('update-request', item.request);
  emit('update-response', item.response);
  
  // 更新服务器响应时间
  emit('update-server-duration', item.server_duration_ms);
  
  // 更新方法，但不更新URL（保持用户手动设置的URL）
  const { method } = parseRequestHeadersForMethodAndUrl(item.request);
  requestMethod.value = method;
  
  // 更新选择的历史记录ID
  selectedHistoryId.value = historyId;
};

// 格式化时间
const formatTime = (timestamp: number): string => {
  const date = new Date(timestamp);
  return date.toLocaleTimeString([], { hour: '2-digit', minute: '2-digit', second: '2-digit' });
};

// 格式化日期
const formatDate = (timestamp: number): string => {
  const date = new Date(timestamp);
  return date.toLocaleDateString();
};

// 获取历史记录的简短描述
const getStatusClass = (statusCode: number = 0): string => {
  if (statusCode >= 200 && statusCode < 300) {
    return 'text-green-500 dark:text-green-400';
  } else if (statusCode >= 400 && statusCode < 500) {
    return 'text-orange-500 dark:text-orange-400';
  } else if (statusCode >= 500) {
    return 'text-red-500 dark:text-red-400';
  }
  return 'text-gray-500 dark:text-gray-400';
};

// 手动实现按日期分组历史记录
const groupHistoryByDate = () => {
  const groups: Record<string, RequestHistory[]> = {};
  
  history.value.forEach(item => {
    const dateKey = formatDate(item.timestamp);
    if (!groups[dateKey]) {
      groups[dateKey] = [];
    }
    groups[dateKey].push(item);
  });
  
  return groups;
};

// 模拟初始化tab的history
onMounted(() => {
  if (!props.tab.history) {
    history.value = [];
  }
  
  // 初始化历史记录计数器，设置为当前历史记录的最大序列ID
  if (history.value.length > 0) {
    const maxSequenceId = Math.max(...history.value.map(item => item.sequenceId || 0));
    historyCounter.value = maxSequenceId;
  }
});
</script>

<template>
  <div class="flex flex-col h-full">
    <!-- Control Bar -->
    <div class="repeater-control-bar">
      <!-- Send button -->
      <button
        class="btn btn-primary mr-2"
        @click="sendRequest"
        :disabled="isLoading"
        :class="{ 'opacity-70 cursor-wait': isLoading }"
      >
        <i :class="['bx', isLoading ? 'bx-loader-alt animate-spin' : 'bx-send', 'mr-1']"></i>
        {{ isLoading ? (t('modules.repeater.sending') || 'Sending...') : (t('modules.repeater.send') || 'Send') }}
      </button>
      
      <!-- URL input -->
      <input 
        type="text"
        class="repeater-input flex-1"
        :value="requestUrl"
        @change="handleUrlChange"
        placeholder="https://example.com/api/endpoint"
      />
      
      <!-- 请求历史下拉框 -->
      <div class="ml-2 relative">
        <select
          class="repeater-history-select"
          :value="selectedHistoryId || ''"
          @change="(e) => selectHistoryItem(Number((e.target as HTMLSelectElement).value))"
          :class="{ 'repeater-history-select-empty': !selectedHistoryId }"
        >
          <option value="" disabled>{{ t('modules.repeater.history') || 'History' }}</option>
          <optgroup 
            v-for="(dayGroup, date) in groupHistoryByDate()"
            :key="date"
            :label="date"
          >
            <option
              v-for="item in dayGroup"
              :key="item.id"
              :value="item.id"
              :class="getStatusClass(item.statusCode)"
            >
              #{{ item.sequenceId }} - {{ formatTime(item.timestamp) }} - {{ item.method }} {{ item.statusCode ? `(${item.statusCode})` : '' }}
            </option>
          </optgroup>
        </select>
        <div class="absolute inset-y-0 right-0 flex items-center pr-2 pointer-events-none">
          <i class="bx bx-chevron-down repeater-select-icon"></i>
        </div>
      </div>
      
      <!-- 历史记录数量指示 -->
      <div v-if="history.length > 0" class="repeater-history-count">
        <i class="bx bx-history text-sm mr-1"></i>
        {{ history.length }}
      </div>
    </div>
    
    <!-- Content Area - Using RequestResponsePanel -->
    <div class="flex-1 overflow-hidden">
      <!-- 将字符串转换为 RequestData 和 ResponseData -->
      <RequestResponsePanel
        :requestData="tab.request"
        :responseData="tab.response ? tab.response : ''"
        :serverDurationMs="tab.serverDurationMs || 0"
        :hideEmptyResponse="!tab.response"
        :uuid="tab.id"
        @update:requestData="updateRequest"
        @update:responseData="updateResponse"
        :responseReadOnly="true"
      />
    </div>
  </div>
</template>

<style scoped>
.repeater-select-icon {
  color: var(--repeater-text-muted);
}

.repeater-history-select {
  min-width: 180px;
  padding: 0.5rem 1rem;
  font-size: 0.875rem;
  border: 1px solid var(--repeater-border);
  border-radius: 0.375rem;
  background-color: var(--repeater-bg-secondary);
  color: var(--repeater-text-primary);
  appearance: none;
  padding-right: 2rem;
}

.repeater-history-select-empty {
  color: var(--repeater-text-muted);
}

.repeater-history-count {
  margin-left: 0.5rem;
  font-size: 0.75rem;
  color: var(--repeater-text-secondary);
  background-color: var(--repeater-bg-secondary);
  border-radius: 9999px;
  padding: 0.25rem 0.5rem;
  display: flex;
  align-items: center;
}

.repeater-control-bar {
  display: flex;
  align-items: center;
  padding: 8px 16px;
  background-color: var(--repeater-bg-primary);
  border-bottom: 1px solid var(--repeater-border);
}

.repeater-input {
  border: 1px solid var(--repeater-border);
  border-radius: 4px;
  padding: 8px 12px;
  font-size: 14px;
  color: var(--repeater-text-primary);
  background-color: var(--repeater-bg-secondary);
  outline: none;
  transition: all 0.2s ease;
}

.repeater-input:focus {
  border-color: var(--repeater-btn-primary);
  box-shadow: 0 0 0 2px rgba(79, 70, 229, 0.1);
}

.repeater-input:hover {
  border-color: var(--repeater-btn-primary);
}
</style>