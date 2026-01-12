<script setup lang="ts">
import { ref, onMounted, computed, onBeforeUnmount, nextTick, watch } from 'vue';
import { useI18n } from 'vue-i18n';
import { useModulesStore, type ProxyHistoryItem } from '../../store';
import type { HttpHistoryEvent, HttpMarkerEvent, HttpHistoryData } from '../../types';
import ProxyActionMenu from './ProxyActionMenu.vue';
import type { FilterOptions } from './ProxyFilter.vue';
import ProxyControlBar from './panels/ProxyControlBar.vue';
import ProxyHistoryPanel from './panels/ProxyHistoryPanel.vue';
import ProxyInterceptPanel from './panels/ProxyInterceptPanel.vue';
import ProxyMatchReplacePanel from './panels/ProxyMatchReplacePanel.vue';
import ProxyNotification from './panels/ProxyNotification.vue';
import { Events } from "@wailsio/runtime";
// @ts-ignore 
import {
  GetHistoryAll,
  GetHistoryDump,
  SetProxyInterceptMode,
  ClearAllHistoryData,
  NewClaudeAgentWindow
} from "../../../bindings/github.com/yhy0/ChYing/app.js";

// 暂时声明 CleanMemoryData，需要重新生成绑定文件
declare const CleanMemoryData: (keepCount: number) => Promise<{ Data: string; Error: string }>;

// 使用store
const store = useModulesStore();

// 使用国际化
const { t } = useI18n();

// 当前激活的选项卡
const activeTab = ref<'history' | 'intercept' | 'matchreplace'>('history');

// 通知信息
const notification = ref<{ message: string; visible: boolean; type: 'success' | 'error' } | null>(null);

// Filter options
const filterOptions = ref<FilterOptions>({
  method: '',
  host: '',
  path: '',
  status: '',
  contentType: '',
});

// 显示通知
const showNotification = (message: string, type: 'success' | 'error' = 'success') => {
  notification.value = {
    message,
    visible: true,
    type
  };
};

// 关闭通知
const closeNotification = () => {
  if (notification.value) {
    notification.value.visible = false;
  }
};

// 存储全部历史记录的引用
const originalProxyHistory = ref<ProxyHistoryItem[]>([]);
// 是否启用 DSL 过滤
const dslFilterEnabled = ref(false);

// 组件挂载
onMounted(() => {
  // 监听新的HTTP历史记录
  // Wails v3 事件回调接收 WailsEvent 对象，实际数据在 event.data 中
  Events.On("HttpHistory", (event: any) => {
    // event.data 是后端发送的 HTTPHistory 对象
    const item = event.data;
    if (!item) {
      console.warn('HttpHistory event received with no data');
      return;
    }
    
    const newItem = {
      id: item.id,
      method: item.method,
      url: item.full_url,
      host: item.host,
      path: item.path,
      status: parseInt(item.status),
      length: parseInt(item.length),
      mimeType: item.mime_type,
      extension: item.extension,
      title: item.title,
      ip: item.ip,
      note: item.note,
      timestamp: new Date().toISOString(),
      request: '',
      response: ''
    };
    store.addProxyHistoryItem(newItem);

    // 如果没有选中的项，自动选中新添加的项
    if (!selectedItem.value) {
      selectItem(newItem);
    }
  });

  // 获取所有历史记录
  GetHistoryAll().then((res: any) => {
    console.log('Received history data:', res);
    
    // 只有在历史记录为空时才加载，避免重复加载
    if (store.proxyHistory.length === 0 && res && res.length > 0) {
      res.forEach((item: HttpHistoryData) => {
        const newItem = {
          id: item.id,
          method: item.method,
          url: item.full_url,
          host: item.host,
          path: item.path,
          status: parseInt(item.status),
          length: parseInt(item.length),
          mimeType: item.mime_type,
          extension: item.extension,
          title: item.title,
          ip: item.ip,
          note: item.note,
          timestamp: new Date().toISOString(),
          request: '',
          response: ''
        };
        store.addProxyHistoryItem(newItem);
      });

      // 保存原始历史记录引用
      originalProxyHistory.value = [...store.proxyHistory];
    }
  }).catch((err: any) => {
    console.error('Failed to load history:', err);
  });

  // 监听标记更新
  // Wails v3 事件回调接收 WailsEvent 对象，实际数据在 event.data 中
  Events.On("HttpMarker", (event: any) => {
    const markerData = event.data;
    if (!markerData) {
      console.warn('HttpMarker event received with no data');
      return;
    }
    const { id, color } = markerData;
    console.log('Updating item color:', id, color);
    store.setProxyItemColor(id, color);
  });

  // 设置第一项为默认选中（如果有数据）
  if (store.proxyHistory.length > 0 && !selectedItem.value) {
    console.log('Setting initial selected item:', store.proxyHistory[0]);
    selectItem(store.proxyHistory[0]);
  }

  // 添加键盘事件监听器
  window.addEventListener('keydown', handleKeyDown);
});

const selectedItem = ref<ProxyHistoryItem | null>(null);
const proxyInterceptEnabled = ref(false);
const showContextMenu = ref(false);
const contextMenuPos = ref({ x: 0, y: 0 });
const contextMenuRequest = ref<ProxyHistoryItem | null>(null);

// 多选相关状态
const enableMultiSelect = ref(true); // 默认启用多选
const checkedItems = ref<ProxyHistoryItem[]>([]);

// 计算是否显示浮动操作栏
const showFloatingBar = computed(() => checkedItems.value.length > 0);

// Filter the proxy history items based on the current filter options
const filteredProxyHistory = computed(() => {
  return store.proxyHistory.filter(item => {
    // If all filter options are empty, return all items
    if (!filterOptions.value.method &&
        !filterOptions.value.host &&
        !filterOptions.value.path &&
        !filterOptions.value.status &&
        !filterOptions.value.contentType) {
      return true;
    }

    // Apply method filter
    if (filterOptions.value.method &&
        item.method.toLowerCase() !== filterOptions.value.method.toLowerCase()) {
      return false;
    }

    // Apply host filter
    if (filterOptions.value.host &&
        !item.host.toLowerCase().includes(filterOptions.value.host.toLowerCase())) {
      return false;
    }

    // Apply path filter
    if (filterOptions.value.path &&
        !item.path.toLowerCase().includes(filterOptions.value.path.toLowerCase())) {
      return false;
    }

    // Apply status filter
    if (filterOptions.value.status) {
      // Allow filtering by status code range (e.g. 2xx, 4xx)
      if (filterOptions.value.status.endsWith('xx')) {
        const statusPrefix = filterOptions.value.status.charAt(0);
        const itemStatusPrefix = String(item.status).charAt(0);
        if (statusPrefix !== itemStatusPrefix) {
          return false;
        }
      } else if (String(item.status) !== filterOptions.value.status) {
        return false;
      }
    }

    // Apply content type filter
    if (filterOptions.value.contentType &&
        !item.mimeType.toLowerCase().includes(filterOptions.value.contentType.toLowerCase())) {
      return false;
    }

    return true;
  });
});

// 选择项目
const selectItem = (item: ProxyHistoryItem) => {
  store.proxyHistory.forEach(historyItem => {
    historyItem.selected = historyItem.id === item.id;
  });
  
  // 先清空之前的请求和响应数据，并设置加载状态
  if (selectedItem.value !== item) {
    item.isLoading = true;
    item.request = '';
    item.response = '';
  }
  
  selectedItem.value = item;

  // 获取请求和响应数据
  GetHistoryDump(Number(item.id)).then((HTTPBody: any) => {
    console.log('HTTPBody:', HTTPBody);
    if (selectedItem.value && selectedItem.value.id === item.id) {
      selectedItem.value.request = HTTPBody.request_raw;
      selectedItem.value.response = HTTPBody.response_raw;
      selectedItem.value.isLoading = false;
      
      // 强制更新视图
      nextTick(() => {
        const tempItem = selectedItem.value;
        selectedItem.value = null;
        setTimeout(() => {
          selectedItem.value = tempItem;
        }, 0);
      });
    }
  }).catch((err: any) => {
    console.log('Error getting history dump:', err);
    if (selectedItem.value && selectedItem.value.id === item.id) {
      selectedItem.value.isLoading = false;
    }
  });
};

// 监听proxyHistory变化
watch(
  () => store.proxyHistory,
  async (newHistory) => {
    // 如果不是 DSL 过滤模式，则更新原始历史记录引用
    if (!dslFilterEnabled.value) {
      originalProxyHistory.value = [...newHistory];
    }

    // 如果有新的历史记录
    if (newHistory.length > 0) {
      // 如果当前没有选中项，自动选中最新的一项
      if (!selectedItem.value) {
        const latestItem = newHistory[newHistory.length - 1];
        await selectItem(latestItem);
      } else {
        // 如果当前选中项在历史记录中，更新它的数据
        const currentItem = newHistory.find(item => item.id === selectedItem.value?.id);
        if (currentItem) {
          selectedItem.value = currentItem;
          // 重新获取详细数据
          GetHistoryDump(Number(currentItem.id)).then((HTTPBody: any) => {
            if (selectedItem.value) {
              selectedItem.value.request = HTTPBody.request_raw;
              selectedItem.value.response = HTTPBody.response_raw;
            }
          }).catch((err: any) => {
            console.log('Error getting history dump:', err);
          });
        }
      }
    }
  },
  { deep: true }
);

// 拦截面板引用
const interceptPanelRef = ref<InstanceType<typeof ProxyInterceptPanel>>();

// 拦截开关
const toggleInterception = () => {
  const newStatus = !proxyInterceptEnabled.value;
  
  console.log('Toggling intercept mode:', { 
    current: proxyInterceptEnabled.value, 
    new: newStatus 
  });

  // 如果是关闭拦截，先放行所有队列中的请求
  if (!newStatus && interceptPanelRef.value) {
    console.log('Intercept is being disabled, forwarding all queued items...');
    interceptPanelRef.value.forwardAllAndClear();
  }

  // 调用后端拦截API
  SetProxyInterceptMode(newStatus).then(() => {
    proxyInterceptEnabled.value = newStatus;
    console.log('Intercept status changed successfully:', newStatus);
    
    const message = newStatus
      ? (t('modules.proxy.controls.intercept_on') || '拦截已开启')
      : (t('modules.proxy.controls.intercept_off') || '拦截已关闭，所有队列项目已放行');
    
    showNotification(message, 'success');
  }).catch((err: any) => {
    console.error('Error setting intercept status:', err);
  showNotification(
        t('modules.proxy.intercept_toggle_failed') || '拦截状态切换失败', 
        'error'
  );
  });
};

// 更新请求数据
const updateRequestData = (data: string) => {
  if (selectedItem.value) {
    selectedItem.value.request = data;
  }
};

// 更新响应数据
const updateResponseData = (data: string) => {
  if (selectedItem.value) {
    selectedItem.value.response = data;
  }
};

// 旧的拦截逻辑已移除，现在由拦截面板内部管理

// 处理右键菜单
const handleContextMenu = (event: MouseEvent, item: ProxyHistoryItem) => {
  event.preventDefault();
  contextMenuPos.value = { x: event.clientX, y: event.clientY };
  contextMenuRequest.value = item;
  showContextMenu.value = true;
};

// 关闭右键菜单
const closeContextMenu = () => {
  showContextMenu.value = false;
};

// 发送到Repeater
const sendToRepeater = () => {
  if (!selectedItem.value) return;

  // 使用store发送到Repeater
  store.sendToRepeater(selectedItem.value);

  // 显示成功通知
  showNotification(t('modules.proxy.notifications.sentToRepeater'));

  closeContextMenu();
};

// 发送到Intruder
const sendToIntruder = () => {
  if (!selectedItem.value) return;

  // 使用store发送到Intruder
  store.sendToIntruder(selectedItem.value);

  // 显示成功通知
  showNotification(t('modules.proxy.notifications.sentToIntruder'));

  closeContextMenu();
};

// 设置行颜色
const setRowColor = (item: ProxyHistoryItem, color: string) => {
  // 使用store中的方法设置颜色
  store.setProxyItemColor(item.id, color);
};

// 旧的拦截项删除逻辑已移除

// 接收子组件的通知
const handleNotification = (message: string | { message: string, type: string }) => {
  // 处理不同格式的消息参数
  if (typeof message === 'object' && message !== null) {
    // 如果是对象格式，解构出message和type
    const { message: msg, type = 'success' } = message as { message: string, type: string };
    showNotification(msg, type as 'success' | 'error');
  } else {
    // 如果是字符串格式，使用默认type
    showNotification(message as string, 'success');
  }
};

// 处理过滤器变化
const handleFilterChange = (options: FilterOptions) => {
  filterOptions.value = options;
};

// 重置过滤器
const handleResetFilter = () => {
  filterOptions.value = {
    method: '',
    host: '',
    path: '',
    status: '',
    contentType: '',
  };
};

// 清除历史记录
const clearHistory = () => {
  store.clearProxyHistory();
  selectedItem.value = null;
};

// 旧的添加拦截项逻辑已移除

// 键盘快捷键处理
const handleKeyDown = (event: KeyboardEvent) => {
  // 确保没有在输入框中
  if (event.target instanceof HTMLInputElement ||
      event.target instanceof HTMLTextAreaElement) {
    return;
  }

  if (event.ctrlKey && event.key.toLowerCase() === 'i' && selectedItem.value) {
    // 阻止默认行为
    event.preventDefault();
    // 发送到Intruder
    sendToIntruder();
  } else if (event.ctrlKey && event.key.toLowerCase() === 'r' && selectedItem.value) {
    // 阻止默认行为
    event.preventDefault();
    // 发送到Repeater
    sendToRepeater();
  }
};

// 组件卸载
onBeforeUnmount(() => {
  // 清理 Wails 事件监听器
  Events.Off("HttpHistory");
  Events.Off("HttpMarker");
  // 清理键盘事件监听器
  window.removeEventListener('keydown', handleKeyDown);
});

// 旧的拦截项更新逻辑已移除

// 处理DSL搜索结果
const handleDSLSearchResults = (results: any[]) => {
  if (!results || results.length === 0) {
    // 如果结果为空，显示通知并清空表格
    showNotification(t('modules.proxy.dsl.no_results'), 'error');
    // 使用空数组更新表格，显示"无结果"状态
    store.setProxyHistory([]);
  } else {
    // 格式化时间戳并转换状态码为数字
    results.forEach(item => {
      if (item.Timestamp) {
        item.timestamp = item.Timestamp;
      }
      if (item.Status) {
        item.status = Number(item.Status);
      }
    });
    // 更新表格数据
    store.setProxyHistory(results);
  }
};

// 清除 DSL 过滤，恢复显示所有历史记录
const clearDSLFilter = () => {
  if (!dslFilterEnabled.value) return;

  // 禁用 DSL 过滤标志
  dslFilterEnabled.value = false;

  // 恢复原始历史记录
  store.setProxyHistory([...originalProxyHistory.value]);

  // 选中第一项（如果有）
  if (store.proxyHistory.length > 0) {
    selectItem(store.proxyHistory[0]);
  } else {
    selectedItem.value = null;
  }

  showNotification(t('modules.proxy.dsl.filter_cleared'), 'success');
};

// 清除所有历史记录
const handleClearAllHistory = async () => {
  try {
    await ClearAllHistoryData(); // 调用后端清空数据
    store.clearProxyHistory();     // 清空前端 store
    originalProxyHistory.value = []; // 清空原始历史记录的本地引用
    selectedItem.value = null;     // 清除选中项
    showNotification('所有历史记录已清空（数据库+内存）', 'success');
  } catch (error) {
    console.error('Failed to clear all history:', error);
    showNotification('清空所有历史记录失败', 'error');
  }
  closeContextMenu(); // 关闭右键菜单
};

// 仅清理内存数据
const handleCleanMemoryOnly = async () => {
  try {
    // 调用 Wails API 清理内存数据
    const result = await CleanMemoryData(0); // 0 表示清空所有内存数据
    
    if (result.Error) {
      throw new Error(result.Error);
    }
    
    showNotification('内存数据已清理（数据库数据保留）', 'success');
  } catch (error) {
    console.error('Failed to clean memory data:', error);
    showNotification('清理内存数据失败', 'error');
  }
  closeContextMenu();
};

// 按主机筛选
const handleFilterByHost = (host: string) => {
  filterOptions.value = {
    method: '',
    host: host,
    path: '',
    status: '',
    contentType: '',
  };
  showNotification(`已筛选主机: ${host}`, 'success');
  closeContextMenu();
};

// 按方法筛选
const handleFilterByMethod = (method: string) => {
  filterOptions.value = {
    method: method,
    host: '',
    path: '',
    status: '',
    contentType: '',
  };
  showNotification(`已筛选方法: ${method}`, 'success');
  closeContextMenu();
};

// 清除所有筛选
const handleClearFilters = () => {
  filterOptions.value = {
    method: '',
    host: '',
    path: '',
    status: '',
    contentType: '',
  };
  showNotification('已清除所有筛选条件', 'success');
  closeContextMenu();
};

// 发送选中项到 AI 助手
const sendToClaudeAgent = () => {
  if (checkedItems.value.length === 0) return;

  // 获取选中项的 ID 列表
  const selectedIds = checkedItems.value.map(item => item.id);

  // 调用后端创建新的 Claude Agent 窗口，传递流量 ID
  NewClaudeAgentWindow(selectedIds);

  // 清空选中状态
  checkedItems.value = [];
};

// 清除选中项
const clearCheckedItems = () => {
  checkedItems.value = [];
};
</script>

<template>
  <div class="flex flex-col h-full">
    <!-- 通知组件 -->
    <ProxyNotification v-if="notification" :message="notification.message" :type="notification.type"
      :visible="notification.visible" @close="closeNotification" />

    <!-- 控制栏 -->
    <ProxyControlBar :interceptEnabled="proxyInterceptEnabled" @toggle-intercept="toggleInterception"
      @filter="handleFilterChange" @reset-filter="handleResetFilter" @clear="clearHistory"
      @send-to-repeater="sendToRepeater" @send-to-intruder="sendToIntruder" @search-results="handleDSLSearchResults"
      @clear-search="clearDSLFilter" @notify="handleNotification" />

    <!-- 内容区域 -->
    <div class="flex-1 flex flex-col overflow-hidden">
      <!-- 标签导航 -->
      <div class="flex border-b border-gray-200 dark:border-gray-700 relative z-10 bg-white dark:bg-gray-800 shadow-sm">
        <!-- (B) 固定高度，有z-index和背景 -->
        <button class="px-4 py-2 text-sm font-medium flex items-center" :class="[
            activeTab === 'history'
              ? 'border-b-2 border-indigo-500 text-indigo-600 dark:text-indigo-400'
              : 'text-gray-600 dark:text-gray-400 hover:text-gray-800 dark:hover:text-gray-200'
        ]" @click="activeTab = 'history'">
          <i class="bx bx-history mr-1.5"></i>
          {{ t('modules.proxy.tabs.history') }}
        </button>

        <button class="px-4 py-2 text-sm font-medium flex items-center" :class="[
            activeTab === 'intercept'
              ? 'border-b-2 border-indigo-500 text-indigo-600 dark:text-indigo-400'
              : 'text-gray-600 dark:text-gray-400 hover:text-gray-800 dark:hover:text-gray-200'
        ]" @click="activeTab = 'intercept'">
          <i class="bx bx-intersect mr-1.5"></i>
          {{ t('modules.proxy.tabs.intercept') }}
        </button>

        <button class="px-4 py-2 text-sm font-medium flex items-center" :class="[
          activeTab === 'matchreplace'
            ? 'border-b-2 border-indigo-500 text-indigo-600 dark:text-indigo-400'
            : 'text-gray-600 dark:text-gray-400 hover:text-gray-800 dark:hover:text-gray-200'
        ]" @click="activeTab = 'matchreplace'">
          <i class="bx bx-code-block mr-1.5"></i>
          {{ t('modules.matchReplace.title') }}
        </button>
      </div>

      <!-- 内容面板 -->
      <!-- 内容面板的容器，这个容器将负责滚动 -->
      <div class="flex-1 overflow-y-auto"> <!-- (C') 新的滚动容器 -->
      <!-- 历史记录面板 -->
        <ProxyHistoryPanel v-if="activeTab === 'history'"
          :items="filteredProxyHistory"
          :selectedItem="selectedItem"
          :enableMultiSelect="enableMultiSelect"
          :checkedItems="checkedItems"
          @select-item="selectItem"
          @context-menu="handleContextMenu"
          @update:requestData="updateRequestData"
          @update:responseData="updateResponseData"
          @set-color="setRowColor"
          @update:checkedItems="checkedItems = $event">
        </ProxyHistoryPanel>

      <!-- 拦截面板 -->
        <ProxyInterceptPanel v-else-if="activeTab === 'intercept'" 
          ref="interceptPanelRef"
          :interceptEnabled="proxyInterceptEnabled"
          @notify="handleNotification" 
          @queue-cleared="() => {}" />

        <!-- 匹配替换面板 -->
        <ProxyMatchReplacePanel v-else-if="activeTab === 'matchreplace'" @notify="handleNotification" />

      </div>
    </div>

    <!-- 右键菜单 -->
    <ProxyActionMenu v-if="showContextMenu" :request="contextMenuRequest" :x="contextMenuPos.x" :y="contextMenuPos.y"
      :visible="showContextMenu" :filterOptions="filterOptions" @close="showContextMenu = false" @send-to-repeater="sendToRepeater"
      @send-to-intruder="sendToIntruder" @set-color="setRowColor" @clear-all-history="handleClearAllHistory"
      @clean-memory-only="handleCleanMemoryOnly" @filter-by-host="handleFilterByHost" @filter-by-method="handleFilterByMethod"
      @clear-filters="handleClearFilters" />

    <!-- 浮动操作栏 - 多选时显示 -->
    <Transition name="slide-up">
      <div v-if="showFloatingBar" class="floating-action-bar">
        <div class="floating-bar-content">
          <div class="selected-count">
            <i class="bx bx-check-square"></i>
            <span>{{ t('modules.proxy.multiSelect.selectedCount', { count: checkedItems.length }) }}</span>
          </div>
          <div class="action-buttons">
            <button
              class="btn btn-primary btn-sm"
              @click="sendToClaudeAgent"
              :title="t('modules.proxy.multiSelect.sendToAI')"
            >
              <i class="bx bx-bot"></i>
              {{ t('modules.proxy.multiSelect.sendToAI') }}
            </button>
            <button
              class="btn btn-secondary btn-sm"
              @click="clearCheckedItems"
              :title="t('modules.proxy.multiSelect.clearSelection')"
            >
              <i class="bx bx-x"></i>
              {{ t('modules.proxy.multiSelect.clearSelection') }}
            </button>
          </div>
        </div>
      </div>
    </Transition>
  </div>
</template>

<style scoped>
/* 浮动操作栏样式 */
.floating-action-bar {
  position: fixed;
  bottom: 24px;
  left: 50%;
  transform: translateX(-50%);
  z-index: 1000;
  background: var(--color-bg-primary, #ffffff);
  border: 1px solid var(--color-border, #e5e7eb);
  border-radius: 12px;
  box-shadow: 0 4px 20px rgba(0, 0, 0, 0.15);
  padding: 12px 20px;
}

.dark .floating-action-bar {
  background: var(--color-bg-secondary, #1f2937);
  border-color: var(--color-border, #374151);
  box-shadow: 0 4px 20px rgba(0, 0, 0, 0.4);
}

.floating-bar-content {
  display: flex;
  align-items: center;
  gap: 20px;
}

.selected-count {
  display: flex;
  align-items: center;
  gap: 8px;
  color: var(--color-text-primary, #1f2937);
  font-weight: 500;
  font-size: 14px;
}

.selected-count i {
  font-size: 18px;
  color: var(--color-primary, #6366f1);
}

.action-buttons {
  display: flex;
  align-items: center;
  gap: 8px;
}

.action-buttons .btn {
  display: flex;
  align-items: center;
  gap: 6px;
  padding: 8px 16px;
  border-radius: 8px;
  font-size: 13px;
  font-weight: 500;
  cursor: pointer;
  transition: all 0.2s ease;
}

.action-buttons .btn i {
  font-size: 16px;
}

.action-buttons .btn-primary {
  background: var(--color-primary, #6366f1);
  color: white;
  border: none;
}

.action-buttons .btn-primary:hover {
  background: var(--color-primary-hover, #4f46e5);
}

.action-buttons .btn-secondary {
  background: transparent;
  color: var(--color-text-secondary, #6b7280);
  border: 1px solid var(--color-border, #e5e7eb);
}

.action-buttons .btn-secondary:hover {
  background: var(--color-bg-hover, #f3f4f6);
  color: var(--color-text-primary, #1f2937);
}

.dark .action-buttons .btn-secondary:hover {
  background: var(--color-bg-hover, #374151);
}

/* 过渡动画 */
.slide-up-enter-active,
.slide-up-leave-active {
  transition: all 0.3s ease;
}

.slide-up-enter-from,
.slide-up-leave-to {
  opacity: 0;
  transform: translateX(-50%) translateY(20px);
}
</style>