<script setup lang="ts">
import { ref, watch, computed } from 'vue';
import { useI18n } from 'vue-i18n';
import { useModulesStore } from '../../store/modules';
import { useVulnerabilityStore } from '../../store/vulnerability';

// 定义后端 VulMessage 结构 (根据 type.go)
interface BackendVulnData {
  CreateTime: string;
  VulnType: string;
  Target: string;
  Ip?: string;
  Method?: string;
  Param?: string;
  Payload?: string;
  CURLCommand?: string;
  Description?: string;
  Request?: string;
  Header?: string;
  Response?: string;
}

interface BackendVulMessage {
  data_type: string;
  vul_data: BackendVulnData;
  plugin: string;
  level: string;
}

// Type for the payload expected from Wails event, based on user's snippet
interface VulEventPayload {
  data: BackendVulMessage[];
}

const { t } = useI18n();
const store = useModulesStore();
const vulnerabilityStore = useVulnerabilityStore();

// Props
const props = defineProps<{
  show: boolean;
}>();

// Emits
const emit = defineEmits<{
  close: [];
}>();

// 前端通知数据接口
interface Notification {
  id: number;
  type: 'info' | 'warning' | 'error' | 'success';
  title: string;
  message: string;
  time: string;
  read: boolean;
  vulnerabilityId?: number; // 关联的漏洞ID
}

// 从漏洞 store 生成通知
const notifications = computed(() => {
  return vulnerabilityStore.vulnerabilities.map(vuln => ({
    id: vuln.id,
    type: mapLevelToType(vuln.level),
    title: `${vuln.plugin}: ${vuln.vulnType}`,
    message: vuln.description || vuln.target || '无具体详情',
    time: formatVulnTimestamp(vuln.createTime),
    read: false,
    vulnerabilityId: vuln.id
  }));
});

// 将后端 VulMessage.Level 映射到前端 Notification.type
const mapLevelToType = (level?: string): Notification['type'] => {
  switch (level?.toLowerCase()) {
    case 'critical':
    case 'high':
      return 'error';
    case 'medium':
      return 'warning';
    case 'low':
      return 'info';
    default:
      console.warn(`Unknown notification level: ${level}`);
      return 'info';
  }
};

// 格式化时间戳 (如果后端返回的是标准时间字符串)
const formatVulnTimestamp = (timestamp?: string): string => {
  if (!timestamp) {
    const now = new Date();
    return `${String(now.getHours()).padStart(2, '0')}:${String(now.getMinutes()).padStart(2, '0')}`;
  }
  try {
    const date = new Date(timestamp);
    // 可以根据需要调整日期时间格式
    return date.toLocaleString(); 
  } catch (e) {
    console.warn('Error formatting timestamp:', timestamp, e);
    // 如果格式化失败，返回原始时间或一个通用占位符
    const now = new Date();
    return `${String(now.getHours()).padStart(2, '0')}:${String(now.getMinutes()).padStart(2, '0')}`;
  }
};

// Escape key handler
const handleKeyDown = (e: KeyboardEvent) => {
  if (e.key === 'Escape' && drawerVisible.value) {
    closeDrawer();
  }
};

// 添加键盘事件监听
document.addEventListener('keydown', handleKeyDown);

// 组件卸载时移除键盘事件监听
const cleanup = () => {
  document.removeEventListener('keydown', handleKeyDown);
};

// 在组件卸载时清理
if (typeof window !== 'undefined') {
  window.addEventListener('beforeunload', cleanup);
}

// 抽屉显示状态
const drawerVisible = computed(() => props.show);

// 关闭抽屉
const closeDrawer = () => {
  emit('close');
};

// 过滤未读通知
const unreadNotifications = computed(() => {
  return notifications.value.filter(notification => !notification.read);
});

// 点击了遮罩层
const handleOverlayClick = (e: MouseEvent) => {
  if ((e.target as HTMLElement).classList.contains('notification-overlay')) {
    closeDrawer();
  }
};

// 获取通知图标
const getNotificationIcon = (type: string) => {
  switch (type) {
    case 'info': return 'bx-info-circle';
    case 'warning': return 'bx-error';
    case 'error': return 'bx-error-circle';
    case 'success': return 'bx-check-circle';
    default: return 'bx-bell';
  }
};

// 获取通知颜色类
const getNotificationColorClass = (type: string) => {
  switch (type) {
    case 'info': return 'bg-blue-500';
    case 'warning': return 'bg-yellow-500';
    case 'error': return 'bg-red-500';
    case 'success': return 'bg-green-500';
    default: return 'bg-gray-500';
  }
};

// 监听 props.show 的变化
watch(() => props.show, (isVisible) => {
  if (isVisible) {
    setTimeout(() => {
      // 这里可以标记漏洞为已读，或者其他逻辑
      store.setUnreadCount(0);
    }, 1500);
  }
});

// 清除所有通知（清空漏洞数据）
const clearAllNotifications = () => {
  vulnerabilityStore.clearAllVulnerabilities();
  store.setUnreadCount(0);
};
</script>

<template>
  <!-- 通知抽屉 - 使用Vue的Transition组件实现滑动效果 -->
  <Transition name="drawer">
    <div v-if="drawerVisible" class="fixed inset-0 z-50 overflow-hidden">
      <!-- 背景遮罩 -->
      <div 
        class="absolute inset-0 bg-black bg-opacity-25 notification-overlay"
        @click="handleOverlayClick"
      ></div>
      
      <!-- 抽屉面板 -->
      <div 
        class="absolute inset-y-0 right-0 bg-white dark:bg-[#1e1e2e] shadow-xl flex flex-col notification-drawer"
        :class="{ 'sm:w-96': drawerVisible }"
      >
        <!-- 抽屉头部 -->
        <div class="px-4 py-3 border-b border-gray-200 dark:border-gray-700 flex items-center justify-between">
          <h2 class="text-lg font-medium text-gray-900 dark:text-white flex items-center">
            <i class="bx bx-bell mr-2"></i>
            {{ t('common.ui.notifications') }}
            <span v-if="unreadNotifications.length > 0" 
              class="ml-2 text-xs bg-red-500 text-white rounded-full px-1.5 py-0.5">
              {{ unreadNotifications.length }}
            </span>
          </h2>
          <button 
            class="p-1 rounded-md text-gray-500 hover:text-gray-700 dark:text-gray-400 dark:hover:text-gray-200 transition-colors"
            @click="closeDrawer"
            :aria-label="t('common.ui.close_notification')"
          >
            <i class="bx bx-x text-xl"></i>
          </button>
        </div>
        
        <!-- 通知内容 -->
        <div class="flex-1 overflow-y-auto">
          <div v-if="notifications.length === 0" class="flex flex-col items-center justify-center h-full p-4">
            <i class="bx bx-bell-off text-5xl text-gray-300 dark:text-gray-600"></i>
            <p class="mt-2 text-sm text-gray-500 dark:text-gray-400">{{ t('common.ui.no_notifications') }}</p>
          </div>
          
          <div v-else>
            <!-- 通知列表 -->
            <div class="divide-y divide-gray-200 dark:divide-gray-700">
              <div 
                v-for="notification in notifications" 
                :key="notification.id"
                class="px-4 py-3 hover:bg-gray-50 dark:hover:bg-gray-800/50 transition-colors"
                :class="{ 'bg-blue-50 dark:bg-blue-900/10': !notification.read }"
              >
                <div class="flex items-start">
                  <!-- 图标 -->
                  <div 
                    class="flex-shrink-0 mr-3 w-8 h-8 rounded-full flex items-center justify-center text-white"
                    :class="getNotificationColorClass(notification.type)"
                  >
                    <i class="bx text-lg" :class="getNotificationIcon(notification.type)"></i>
                  </div>
                  
                  <!-- 内容 -->
                  <div class="flex-1 min-w-0">
                    <h3 class="text-sm font-medium text-gray-900 dark:text-white">
                      {{ notification.title }}
                    </h3>
                    <p class="text-xs text-gray-500 dark:text-gray-400 mt-1 line-clamp-3" :title="notification.message">
                      {{ notification.message }}
                    </p>
                    <span class="text-xs text-gray-400 dark:text-gray-500 mt-1 block">
                      {{ notification.time }}
                    </span>
                  </div>
                </div>
              </div>
            </div>
          </div>
        </div>
        
        <!-- 底部操作栏 -->
        <div class="p-3 border-t border-gray-200 dark:border-gray-700 flex justify-between">
          <button 
            class="text-xs text-gray-600 dark:text-gray-400 hover:text-gray-900 dark:hover:text-white transition-colors"
            @click="clearAllNotifications"
            :disabled="notifications.length === 0"
          >
            {{ t('common.actions.clear_all') }}
          </button>
          <button 
            class="text-xs text-blue-600 dark:text-blue-400 hover:text-blue-800 dark:hover:text-blue-300 transition-colors"
            @click="closeDrawer"
          >
            {{ t('common.actions.close') }}
          </button>
        </div>
      </div>
    </div>
  </Transition>
</template>

<style scoped>
/* 抽屉动画 */
.drawer-enter-active,
.drawer-leave-active {
  transition: transform 0.3s ease;
}

.drawer-enter-from,
.drawer-leave-to {
  transform: translateX(100%);
}

/* 确保抽屉在最高层级 */
.notification-overlay {
  z-index: 50;
}

/* 抽屉面板样式 */
.notification-drawer {
  z-index: 60;
  box-shadow: -4px 0 15px rgba(0, 0, 0, 0.1);
  width: 320px;
}

@media (min-width: 640px) {
  .notification-drawer {
    width: 384px;
  }
}

/* 确保在暗黑模式下有足够的对比度 */
.dark .notification-drawer {
  box-shadow: -4px 0 15px rgba(0, 0, 0, 0.3);
}

/* 滚动条美化 */
.overflow-y-auto::-webkit-scrollbar {
  width: 4px;
}

.overflow-y-auto::-webkit-scrollbar-track {
  background: transparent;
}

.overflow-y-auto::-webkit-scrollbar-thumb {
  background-color: rgba(156, 163, 175, 0.5);
  border-radius: 2px;
}

.overflow-y-auto::-webkit-scrollbar-thumb:hover {
  background-color: rgba(156, 163, 175, 0.7);
}

/* 消息内容截断 */
.line-clamp-3 {
  display: -webkit-box;
  -webkit-line-clamp: 3;
  -webkit-box-orient: vertical;  
  overflow: hidden;
  text-overflow: ellipsis;
}
</style> 