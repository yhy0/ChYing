<script setup lang="ts">
import { ref, onMounted, onUnmounted, computed } from 'vue';
import { useI18n } from 'vue-i18n';
import { Events } from "@wailsio/runtime";
import { useVulnerabilityStore } from '../../store';
import { useScanLogStore } from '../../store/scanLog';
import { message } from '../../utils/message';
import type { MemoryInfo } from '../../types/memory';
import type { VulnerabilityMessage, RequestScanMsg } from '../../types';

// 数据库错误事件类型
interface DBError {
  operation: string;
  error: string;
  time: string;
}
// @ts-ignore
import { GetMemoryUsage, NewScanLogWindow, NewVulnerabilityWindow } from "../../../bindings/github.com/yhy0/ChYing/app.js";

const { t } = useI18n();
const vulnerabilityStore = useVulnerabilityStore();
const scanLogStore = useScanLogStore();

// 内存使用状态
const memoryUsage = ref('获取中...');
const goroutineCount = ref(0);
const gcCount = ref(0);

// 漏洞总量（从 store 获取）
const vulnerabilityCount = computed(() => vulnerabilityStore.totalCount);

let memoryInterval: number | null = null;
let unlistenVulMessage: (() => void) | undefined;
let unlistenRequestScanMsg: (() => void) | undefined;
let unlistenDBError: (() => void) | undefined;

// 获取内存使用情况
const fetchMemoryUsage = async () => {
  try {
    // 直接调用 Wails 运行时，而不是通过绑定文件
    // 这种方式在绑定文件重新生成之前是一个临时解决方案
    const memInfo: MemoryInfo = await GetMemoryUsage();

    if (memInfo) {
      memoryUsage.value = memInfo.allocFormatted;
      goroutineCount.value = memInfo.numGoroutine;
      gcCount.value = memInfo.numGC;
    } else {
      memoryUsage.value = '获取失败';
    }
  } catch (error) {
    console.error('获取内存使用情况失败:', error);
    memoryUsage.value = '获取失败';
  }
};

// 处理漏洞消息，更新 store
const handleVulnerabilityMessage = (payload: { data: VulnerabilityMessage[] }) => {
  // 直接调用 store 的方法来处理漏洞消息
  vulnerabilityStore.handleVulnerabilityMessages(payload);
};

// 处理请求扫描消息，更新扫描日志 store 并同步到其他窗口
const handleRequestScanMessage = (payload: { data: RequestScanMsg | RequestScanMsg[] }) => {
  // 更新当前窗口的 store
  scanLogStore.handleRequestScanMessages(payload);

  // 触发跨窗口同步事件
  const syncData = {
    type: 'add',
    data: payload.data,
    timestamp: Date.now()
  };

  // 通过localStorage触发跨窗口事件
  localStorage.setItem('scan-log-data', JSON.stringify(syncData));

  // 立即触发storage事件（对于当前窗口）
  window.dispatchEvent(new StorageEvent('storage', {
    key: 'scan-log-data',
    newValue: JSON.stringify(syncData),
    oldValue: null,
    storageArea: localStorage,
    url: window.location.href
  }));
};

// 处理数据库错误消息
const handleDBError = (payload: { data: DBError }) => {
  const err = payload.data;
  console.warn('Footer - 数据库错误:', err);
  // 显示错误通知，持续 5 秒
  message.error(`数据库操作失败 [${err.operation}]: ${err.error}`, 5);
};


// 组件挂载时开始获取内存信息和监听漏洞事件
onMounted(async () => {
  fetchMemoryUsage(); // 立即获取一次
  // 每1秒获取一次内存使用情况
  memoryInterval = setInterval(fetchMemoryUsage, 1000);
  
  // 监听漏洞消息事件
  unlistenVulMessage = Events.On('VulMessage', handleVulnerabilityMessage as (event: any) => void);
  
  // 监听请求扫描消息事件
  unlistenRequestScanMsg = Events.On('RequestScanMsg', handleRequestScanMessage as (event: any) => void);

  // 监听数据库错误事件
  unlistenDBError = Events.On('db:error', handleDBError as (event: any) => void);
  
  // 清理旧的 localStorage 漏洞数据（已迁移到数据库）
  try {
    localStorage.removeItem('vulnerabilities');
  } catch (e) {
    console.warn('清理 localStorage 失败:', e);
  }
  
  // 从数据库加载漏洞数据
  await vulnerabilityStore.loadFromDatabase();
});

// 组件卸载时清除定时器和事件监听
onUnmounted(() => {
  if (memoryInterval) {
    clearInterval(memoryInterval);
    memoryInterval = null;
  }
  if (unlistenVulMessage) {
    unlistenVulMessage();
  }
  if (unlistenRequestScanMsg) {
    unlistenRequestScanMsg();
  }
  if (unlistenDBError) {
    unlistenDBError();
  }
});

// 日志显示状态
const showScanLog = ref(false);

// 模拟日志数据
const eventLogs = ref([
  { id: 1, message: t('layout.footer.logs.system_started'), timestamp: new Date().toLocaleTimeString() },
  { id: 2, message: t('layout.footer.logs.module_loaded', { module: 'Intruder' }), timestamp: new Date().toLocaleTimeString() },
  { id: 3, message: t('layout.footer.logs.request_sent'), timestamp: new Date().toLocaleTimeString() },
]);

// 打开扫描日志窗口
const toggleScanLog = async () => {
  try {
    NewScanLogWindow();
  } catch (error) {
    console.error('创建扫描日志窗口失败:', error);
    // 出错时降级处理：显示原来的弹窗
    showScanLog.value = !showScanLog.value;
  }
};

// 打开漏洞窗口
const toggleVulnerabilityWindow = async () => {
  try {
    NewVulnerabilityWindow();
  } catch (error) {
    console.error('创建漏洞窗口失败:', error);
  }
};

</script>

<template>
  <div
    class="h-7 bg-gray-100 dark:bg-[#282838] border-t border-gray-200 dark:border-gray-700 flex items-center px-4 text-xs text-gray-700 dark:text-gray-300 flex-shrink-0 app-layout-footer">
    <div class="flex items-center gap-3">
      <!-- 扫描日志按钮 -->
      <button 
        class="btn btn-sm btn-secondary footer-btn"
        @click="toggleScanLog"
        :title="t('layout.footer.scan_log_title')"
      >
        <i class="bx bx-list-ul no-rotate"></i>
        <span>{{ t('layout.footer.scan_log_title') }}</span>
      </button>
      
      <!-- 漏洞列表按钮 -->
      <button 
        class="btn btn-sm btn-secondary footer-btn vulnerability-btn"
        :class="{ 'has-vulnerabilities': vulnerabilityCount > 0 }"
        @click="toggleVulnerabilityWindow"
        :title="t('layout.footer.all_issues')"
      >
        <i class="bx bx-shield-alt-2 no-rotate"></i>
        <span>{{ t('layout.footer.all_issues') }}</span>
        <!-- 漏洞数量徽章 -->
        <span 
          v-if="vulnerabilityCount > 0" 
          class="vulnerability-badge"
          :class="{ 'high-count': vulnerabilityCount >= 100 }"
        >
          {{ vulnerabilityCount > 999 ? '999+' : vulnerabilityCount }}
        </span>
      </button>
    </div>

    <div class="ml-auto flex items-center space-x-4">
      <div class="flex items-center">
        <i class="bx bx-microchip mr-1"></i>
        <span>{{ t('layout.footer.memory') }}: {{ memoryUsage }}</span>
      </div>
      <div class="flex items-center text-xs text-gray-500 dark:text-gray-400" v-if="goroutineCount > 0">
        <i class="bx bx-code-alt mr-1"></i>
        <span>协程: {{ goroutineCount }}</span>
      </div>
      <div class="flex items-center text-xs text-gray-500 dark:text-gray-400" v-if="gcCount > 0">
        <i class="bx bx-recycle mr-1"></i>
        <span>GC: {{ gcCount }}</span>
      </div>
    </div>
  </div>
</template>

<style scoped>
.app-layout-footer {
  z-index: 30;
  position: relative;
}

/* 页脚按钮特殊样式 */
.footer-btn {
  height: 22px !important;
  min-height: 22px !important;
  padding: 2px 8px !important;
  font-size: 10px !important;
  border-radius: 12px !important;
  border-width: 1px !important;
  transition: all 0.2s ease !important;
}

.footer-btn:hover {
  transform: translateY(-1px) !important;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.15) !important;
}

.footer-btn i {
  font-size: 12px !important;
  margin-right: 4px !important;
}

/* 漏洞按钮特殊样式 */
.vulnerability-btn {
  position: relative;
  overflow: visible; /* 允许徽章突出显示 */
}

/* 当有漏洞时，按钮边框变红 */
.vulnerability-btn.has-vulnerabilities {
  border-color: var(--color-error, #ef4444) !important;
  color: var(--color-error, #ef4444) !important;
}

.vulnerability-btn.has-vulnerabilities:hover {
  background: rgba(239, 68, 68, 0.1) !important;
  border-color: var(--color-error-hover, #dc2626) !important;
  color: var(--color-error-hover, #dc2626) !important;
}

/* 漏洞数量徽章样式 */
.vulnerability-badge {
  position: absolute;
  top: -6px;
  right: -6px;
  background: var(--color-error, #ef4444); /* 红色，确保在所有主题下都显示正确 */
  color: white;
  font-size: 9px;
  font-weight: 600;
  padding: 1px 5px;
  border-radius: 10px;
  min-width: 16px;
  height: 16px;
  display: flex;
  align-items: center;
  justify-content: center;
  box-shadow: 0 2px 4px rgba(0, 0, 0, 0.2);
  border: 2px solid var(--color-bg-primary);
  animation: pulse-badge 2s infinite;
  z-index: 10;
}

.vulnerability-badge.high-count {
  background: var(--color-warning, #f59e0b); /* 橙色，表示高数量 */
  animation: pulse-badge-warning 2s infinite;
}

/* 徽章脉冲动画 */
@keyframes pulse-badge {
  0%, 100% {
    transform: scale(1);
    box-shadow: 0 2px 4px rgba(239, 68, 68, 0.3);
  }
  50% {
    transform: scale(1.1);
    box-shadow: 0 4px 8px rgba(239, 68, 68, 0.5);
  }
}

@keyframes pulse-badge-warning {
  0%, 100% {
    transform: scale(1);
    box-shadow: 0 2px 4px rgba(245, 158, 11, 0.3);
  }
  50% {
    transform: scale(1.1);
    box-shadow: 0 4px 8px rgba(245, 158, 11, 0.5);
  }
}

/* 按钮悬停时徽章样式调整 */
.vulnerability-btn:hover .vulnerability-badge {
  animation: none;
  transform: scale(1.05);
}

/* 确保按钮文本和图标不会被徽章遮挡 */
.vulnerability-btn .bx,
.vulnerability-btn span:not(.vulnerability-badge) {
  z-index: 1;
  position: relative;
}
</style>