<script setup lang="ts">
import { ref, computed, watch, onMounted, onUnmounted } from 'vue';
import { useI18n } from 'vue-i18n';
import { getCurrentTheme } from '../theme';
import { i18n } from '../i18n';
import { useScanLogStore } from '../store/scanLog';
import { useModulesStore } from '../store';
import ScanLogPanel from '../components/scanLog/ScanLogPanel.vue';
import ScanLogFilter from '../components/scanLog/ScanLogFilter.vue';
import ScanLogStatistics from '../components/scanLog/ScanLogStatistics.vue';
import EmptyState from '../components/common/EmptyState.vue';
import type {
  ScanLogItem,
  ScanLogFilter as ScanLogFilterType,
  ScanLogWindowConfig
} from '../types/scanLog';
import type { ProxyHistoryItem } from '../types';

const { t } = useI18n();
const scanLogStore = useScanLogStore();
const modulesStore = useModulesStore();

// 主题响应状态
const currentTheme = ref(getCurrentTheme());
const themeChangeKey = ref(0); // 用于强制重新渲染

// 组件状态
const loading = ref(false);
const showStatistics = ref(true);
const showFilter = ref(true);

// 窗口配置
const windowConfig = ref<ScanLogWindowConfig>({
  autoRefresh: true,
  refreshInterval: 5000,
  maxLogCount: 10000,
  showStatistics: true
});

// 从 store 获取数据
const filteredLogs = computed(() => scanLogStore.filteredLogs);
const selectedLogItem = computed(() => scanLogStore.selectedLogItem);
const statistics = computed(() => scanLogStore.statistics);
const currentFilter = computed(() => scanLogStore.filter);

// 自动刷新定时器
let refreshTimer: number | null = null;

// 通知状态
const notification = ref<{ message: string; visible: boolean; type: 'success' | 'error' } | null>(null);

// 显示通知
const showNotification = (message: string, type: 'success' | 'error' = 'success') => {
  notification.value = {
    message,
    visible: true,
    type
  };
  // 3秒后自动关闭
  setTimeout(() => {
    if (notification.value) {
      notification.value.visible = false;
    }
  }, 3000);
};

// 将 ScanLogItem 转换为 ProxyHistoryItem
const convertToProxyHistoryItem = (item: ScanLogItem): ProxyHistoryItem => {
  return {
    id: item.id,
    method: item.method,
    url: item.target || item.url,
    host: item.target ? new URL(item.target).host : '',
    path: item.path,
    status: item.status || 0,
    length: item.length || 0,
    mimeType: item.contentType || '',
    extension: '',
    title: item.title || '',
    ip: item.ip || '',
    note: '',
    timestamp: item.timestamp,
    request: item.request || '',
    response: item.response || ''
  };
};

// 发送到 Repeater
const sendToRepeater = () => {
  const item = selectedLogItem.value;
  if (!item) return;

  // 确保有请求数据
  if (!item.request) {
    showNotification(t('scanLog.notifications.noRequestData', '请先选择一条有请求数据的记录'), 'error');
    return;
  }

  const proxyItem = convertToProxyHistoryItem(item);
  modulesStore.sendToRepeater(proxyItem);
  showNotification(t('modules.proxy.notifications.sentToRepeater', '已发送到中继器'));
};

// 发送到 Intruder
const sendToIntruder = () => {
  const item = selectedLogItem.value;
  if (!item) return;

  // 确保有请求数据
  if (!item.request) {
    showNotification(t('scanLog.notifications.noRequestData', '请先选择一条有请求数据的记录'), 'error');
    return;
  }

  const proxyItem = convertToProxyHistoryItem(item);
  modulesStore.sendToIntruder(proxyItem);
  showNotification(t('modules.proxy.notifications.sentToIntruder', '已发送到入侵者'));
};

// 键盘快捷键处理
const handleKeyDown = (event: KeyboardEvent) => {
  // 确保没有在输入框中
  if (event.target instanceof HTMLInputElement ||
      event.target instanceof HTMLTextAreaElement) {
    return;
  }

  if (event.ctrlKey && event.key.toLowerCase() === 'i' && selectedLogItem.value) {
    event.preventDefault();
    sendToIntruder();
  } else if (event.ctrlKey && event.key.toLowerCase() === 'r' && selectedLogItem.value) {
    event.preventDefault();
    sendToRepeater();
  }
};

// 处理日志项选择（只设置选中状态，数据获取由子组件处理）
const handleSelectLogItem = (item: ScanLogItem) => {
  scanLogStore.setSelectedLogItem(item);
};

// 处理过滤器变化
const handleFilterChange = (newFilter: ScanLogFilterType) => {
  scanLogStore.updateFilter(newFilter);
};

// 切换统计面板显示
const toggleStatistics = () => {
  showStatistics.value = !showStatistics.value;
};

// 切换过滤器显示
const toggleFilter = () => {
  showFilter.value = !showFilter.value;
};

// 刷新数据（实际上数据是实时从后端推送的）
const refreshData = () => {
  console.log('扫描日志数据来自实时推送，无需手动刷新');
};

// 清空日志（带跨窗口同步）
const clearLogs = () => {
  scanLogStore.clearLogsWithSync();
};

// 导出日志
const exportLogs = () => {
  try {
    const logsToExport = filteredLogs.value;
    const exportData = {
      timestamp: new Date().toISOString(),
      total: logsToExport.length,
      logs: logsToExport
    };
    
    const dataStr = JSON.stringify(exportData, null, 2);
    const blob = new Blob([dataStr], { type: 'application/json' });
    const url = URL.createObjectURL(blob);
    
    const a = document.createElement('a');
    a.href = url;
    a.download = `scan-logs-${new Date().toISOString().split('T')[0]}.json`;
    document.body.appendChild(a);
    a.click();
    document.body.removeChild(a);
    URL.revokeObjectURL(url);
    
    console.log('已导出', logsToExport.length, '条扫描日志');
  } catch (error) {
    console.error('导出日志失败:', error);
  }
};

// 设置自动刷新（扫描日志是实时推送的，这里保留接口但不实际使用）
const setupAutoRefresh = () => {
  if (refreshTimer) {
    clearInterval(refreshTimer);
  }
  // 由于数据是实时推送的，这里不需要定时刷新
};

// 主题变化处理函数
const handleThemeChange = () => {
  currentTheme.value = getCurrentTheme();
  themeChangeKey.value++; // 强制重新渲染子组件
};

// 监听配置变化
watch(() => windowConfig.value.autoRefresh, setupAutoRefresh);
watch(() => windowConfig.value.refreshInterval, setupAutoRefresh);

// localStorage变化处理函数（接收来自主界面的主题、语言和扫描日志变化）
const handleStorageChange = (e: StorageEvent) => {
  if (e.key === 'app-theme' && e.newValue) {
    const newTheme = e.newValue as 'light' | 'dark' | 'system';
    const isDark = newTheme === 'dark' || (newTheme === 'system' && window.matchMedia('(prefers-color-scheme: dark)').matches);
    
    // 每个窗口需要自己更新DOM
    if (isDark) {
      document.documentElement.classList.add('dark');
    } else {
      document.documentElement.classList.remove('dark');
    }
    
    handleThemeChange();
  }
  
  if (e.key === 'language' && e.newValue) {
    const newLang = e.newValue as 'en' | 'zh';
    // 同步语言设置
    i18n.global.locale.value = newLang;
    document.querySelector('html')?.setAttribute('lang', newLang);
      }
    
    console.log("ScanLogView - handleStorageChange", e.key, e.newValue);
    // 处理扫描日志数据同步
  if (e.key === 'scan-log-data' && e.newValue) {
    try {
      const scanLogData = JSON.parse(e.newValue);
      console.log('ScanLogView - 接收到跨窗口扫描日志数据:', scanLogData);
      
      // 直接调用 store 的方法来处理扫描消息
      if (scanLogData.type === 'add') {
        scanLogStore.addScanLog(scanLogData.data);
      } else if (scanLogData.type === 'clear') {
        scanLogStore.clearLogs();
      }
    } catch (error) {
      console.error('解析跨窗口扫描日志数据失败:', error);
    }
  }
};

// 组件挂载
onMounted(() => {
  setupAutoRefresh();

  // 清理过期数据（保留最近7天）
  scanLogStore.cleanupOldLogs();

  // 输出存储使用情况
  const storageInfo = scanLogStore.getStorageInfo();
  console.log('扫描日志存储信息:', storageInfo);

  // 监听localStorage变化（跨窗口主题同步）
  window.addEventListener('storage', handleStorageChange);

  // 添加键盘事件监听器
  window.addEventListener('keydown', handleKeyDown);
});

// 组件卸载
onUnmounted(() => {
  if (refreshTimer) {
    clearInterval(refreshTimer);
  }

  // 移除localStorage监听器
  window.removeEventListener('storage', handleStorageChange);

  // 移除键盘事件监听器
  window.removeEventListener('keydown', handleKeyDown);
});
</script>

<template>
  <div class="scan-log-view" :key="`scan-log-${themeChangeKey}`">
    <!-- 通知组件 -->
    <Transition name="notification">
      <div
        v-if="notification?.visible"
        class="notification-toast"
        :class="notification.type"
      >
        <i :class="notification.type === 'success' ? 'bx bx-check-circle' : 'bx bx-error-circle'"></i>
        <span>{{ notification.message }}</span>
      </div>
    </Transition>

    <!-- 头部工具栏 -->
    <div class="section-header">
      <div class="header-left">
        <h3>{{ t('scanLog.title', '扫描日志') }}</h3>
        <div class="header-actions">
          <button
            @click="toggleFilter"
            class="btn btn-sm"
            :class="showFilter ? 'btn-primary' : 'btn-secondary'"
          >
            <i class="bx bx-filter"></i>
            {{ t('scanLog.filter', '过滤器') }}
          </button>
          <button
            @click="toggleStatistics"
            class="btn btn-sm"
            :class="showStatistics ? 'btn-primary' : 'btn-secondary'"
          >
            <i class="bx bx-bar-chart"></i>
            {{ t('scanLog.statistics', '统计') }}
          </button>
        </div>
      </div>
      
      <div class="header-actions">
        <button
          @click="refreshData"
          :disabled="loading"
          class="btn btn-sm btn-primary"
          :class="{ 'loading': loading }"
        >
          <i class="bx bx-refresh" :class="{ 'bx-spin': loading }"></i>
          {{ t('common.ui.refresh') }}
        </button>
        <button
          @click="exportLogs"
          class="btn btn-sm btn-success"
        >
          <i class="bx bx-download"></i>
          {{ t('common.export') }}
        </button>
        <button
          @click="clearLogs"
          class="btn btn-sm btn-danger"
        >
          <i class="bx bx-trash"></i>
          {{ t('common.actions.clear') }}
        </button>
      </div>
    </div>

    <!-- 内容区域 -->
    <div class="main-content">
      <!-- 侧边栏区域 -->
      <div 
        class="sidebar scrollbar-thin" 
        v-if="showStatistics || showFilter"
        :class="{
          'single-panel': (showStatistics && !showFilter) || (!showStatistics && showFilter),
          'dual-panel': showStatistics && showFilter
        }"
      >
        <!-- 统计面板 -->
        <ScanLogStatistics
          v-if="showStatistics"
          :statistics="statistics"
          class="sidebar-panel statistics-panel"
          :key="`scan-log-stats-${themeChangeKey}`"
        />
        
        <!-- 过滤器面板 -->
        <ScanLogFilter
          v-if="showFilter"
          :filter="currentFilter"
          @update:filter="handleFilterChange"
          class="sidebar-panel filter-panel"
          :key="`scan-log-filter-${themeChangeKey}`"
        />
      </div>

      <!-- 主内容区域 -->
      <div class="scan-log-content">
        <!-- 空状态 -->
        <EmptyState
          v-if="!loading && filteredLogs.length === 0"
          icon="bx-search-alt"
          :title="t('scanLog.empty.title', '暂无扫描日志')"
          :description="t('scanLog.empty.description', '扫描日志将在执行安全扫描时自动记录')"
          size="large"
        />
        
        <!-- 扫描日志面板 -->
        <ScanLogPanel
          v-else
          :items="filteredLogs"
          :selectedItem="selectedLogItem"
          :loading="loading"
          @select-item="handleSelectLogItem"
          :key="`scan-log-panel-${themeChangeKey}`"
        />
      </div>
    </div>
  </div>
</template>

<style scoped>
.scan-log-view {
  display: flex;
  flex-direction: column;
  height: 100vh;
  background: var(--color-bg-primary);
  color: var(--color-text-primary);
}

.header-left {
  display: flex;
  align-items: center;
  gap: var(--spacing-md);
}

.header-left h3 {
  margin: 0;
  font-size: var(--text-lg);
  font-weight: var(--font-weight-semibold);
  color: var(--color-text-primary);
}

.main-content {
  display: flex;
  flex: 1;
  overflow: hidden;
  gap: 0;
}

/* 侧边栏优化 */
.sidebar {
  width: 300px;
  background: var(--glass-bg-secondary);
  backdrop-filter: var(--glass-blur-light);
  -webkit-backdrop-filter: var(--glass-blur-light);
  border-right: 1px solid var(--glass-border-light);
  overflow: hidden;
  flex-shrink: 0;
  display: flex;
  flex-direction: column;
  transition: width var(--glass-transition-normal);
}

.sidebar-panel {
  overflow-y: auto;
  border-bottom: 1px solid var(--glass-border-light);
}

/* 单个面板时占满100%高度 */
.sidebar.single-panel .sidebar-panel {
  flex: 1;
  border-bottom: none;
}

/* 双面板时各占50%高度 */
.sidebar.dual-panel .statistics-panel {
  flex: 0 0 50%;
  max-height: 50%;
}

.sidebar.dual-panel .filter-panel {
  flex: 0 0 50%;
  max-height: 50%;
  border-bottom: none;
}

.scan-log-content {
  flex: 1;
  overflow: hidden;
  background: var(--color-bg-primary);
}

/* 通知样式优化 */
.notification-toast {
  position: fixed;
  top: var(--spacing-lg);
  right: var(--spacing-lg);
  z-index: 9999;
  display: flex;
  align-items: center;
  gap: var(--spacing-sm);
  padding: var(--spacing-sm) var(--spacing-md);
  border-radius: var(--radius-md);
  font-size: var(--text-sm);
  font-weight: var(--font-weight-medium);
  box-shadow: var(--glass-shadow-medium);
  backdrop-filter: var(--glass-blur-light);
  -webkit-backdrop-filter: var(--glass-blur-light);
}

.notification-toast.success {
  background: rgba(16, 185, 129, 0.95);
  color: white;
  border: 1px solid rgba(16, 185, 129, 0.3);
}

.notification-toast.error {
  background: rgba(239, 68, 68, 0.95);
  color: white;
  border: 1px solid rgba(239, 68, 68, 0.3);
}

.notification-toast i {
  font-size: 18px;
}

/* 通知动画 */
.notification-enter-active,
.notification-leave-active {
  transition: all var(--glass-transition-normal);
}

.notification-enter-from,
.notification-leave-to {
  opacity: 0;
  transform: translateX(20px);
}

/* 响应式设计 */
@media (max-width: 1200px) {
  .sidebar {
    width: 280px;
  }
}

@media (max-width: 992px) {
  .sidebar {
    width: 260px;
  }
}
</style> 