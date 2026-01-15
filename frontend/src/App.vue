<script setup lang="ts">
import { onMounted, onUnmounted, ref } from 'vue';
import { useRouter } from 'vue-router';
import { useI18n } from 'vue-i18n';
import 'boxicons/css/boxicons.min.css';
import { useModulesStore } from './store/modules';
import { Events } from '@wailsio/runtime';
import { GetCertificateInfo } from '../bindings/github.com/yhy0/ChYing/app';
import eventBus, {
  SEND_TO_REPEATER_REQUESTED,
  SEND_TO_INTRUDER_FROM_PROXY_REQUESTED,
  SEND_TO_INTRUDER_FROM_REPEATER_REQUESTED
} from './utils/eventBus';

const router = useRouter();
const store = useModulesStore();
const { t } = useI18n();

// 全局通知弹窗状态
const globalNotification = ref({
  visible: false,
  message: '',
  type: 'info' as 'info' | 'success' | 'warning' | 'error'
});

// 显示全局通知弹窗
const showGlobalNotification = (message: string, type: 'info' | 'success' | 'warning' | 'error' = 'info') => {
  globalNotification.value = { visible: true, message, type };
  // 5秒后自动关闭
  setTimeout(() => {
    globalNotification.value.visible = false;
  }, 5000);
};

// 处理代理启动事件
const handleProxyStarted = (event: any) => {
  const data = event.data;
  if (data) {
    if (data.success) {
      // 显示全局通知弹窗
      showGlobalNotification(data.message, 'success');
      // 添加到全局通知中心
      store.addNotification({
        type: 'success',
        title: t('modules.proxy.title'),
        message: data.message
      });
    } else {
      showGlobalNotification(data.message || t('modules.proxy.startFailed'), 'error');
      store.addNotification({
        type: 'error',
        title: t('modules.proxy.title'),
        message: data.message || t('modules.proxy.startFailed')
      });
    }
  }
};

// Event Bus Handler Functions
const handleSendToRepeater = (payload: { sourceItem: import('./types').ProxyHistoryItem }) => {
  console.log('App.vue: Event SEND_TO_REPEATER_REQUESTED received', payload);
  const newTabId = store.addRepeaterTabFromEventPayload(payload); 
  if (newTabId) {
    router.push('/app/repeater');
  }
};

const handleSendToIntruder = (payload: { sourceItem: import('./types').ProxyHistoryItem | import('./types').RepeaterTab }) => {
  console.log('App.vue: Event SEND_TO_INTRUDER_REQUESTED received', payload);
  const newTabId = store.addIntruderTabFromEventPayload(payload); 
  if (newTabId) {
    router.push('/app/intruder');
  }
};

// Event Bus Listeners Setup
onMounted(async () => {
  // 初始化 store 中的通知状态
  store.setUnreadCount?.(0);

  console.log('App.vue: Mounting and setting up event listeners.');

  // 注册 Wails 后端事件监听器
  Events.On("ProxyStarted", handleProxyStarted);

  // 注册 eventBus 事件监听器
  eventBus.on(SEND_TO_REPEATER_REQUESTED, handleSendToRepeater);
  eventBus.on(SEND_TO_INTRUDER_FROM_PROXY_REQUESTED, handleSendToIntruder);
  eventBus.on(SEND_TO_INTRUDER_FROM_REPEATER_REQUESTED, handleSendToIntruder);

  // 添加证书安装提示通知
  try {
    const result = await GetCertificateInfo();
    if (result.data) {
      const certInfo = result.data as {
        certDir: string;
        certPem: string;
        certCrt: string;
        platform: string;
      };
      const platform = certInfo.platform;

      let hint = '';
      if (platform === 'windows') {
        hint = t('notifications.certificate.windowsHint');
      } else if (platform === 'darwin') {
        hint = t('notifications.certificate.macHint');
      } else {
        hint = t('notifications.certificate.linuxHint');
      }

      store.addNotification({
        type: 'info',
        title: t('notifications.certificate.title'),
        message: `${t('notifications.certificate.message')}\n${t('notifications.certificate.path', { path: certInfo.certDir })}\n\n${hint}`
      });
    }
  } catch (e) {
    console.error('Failed to get certificate info:', e);
  }
});

onUnmounted(() => {
  console.log('App.vue: Unmounting and tearing down event listeners.');

  // 取消注册 Wails 后端事件监听器
  Events.Off("ProxyStarted");

  // 取消注册 eventBus 事件监听器
  eventBus.off(SEND_TO_REPEATER_REQUESTED, handleSendToRepeater);
  eventBus.off(SEND_TO_INTRUDER_FROM_PROXY_REQUESTED, handleSendToIntruder);
  eventBus.off(SEND_TO_INTRUDER_FROM_REPEATER_REQUESTED, handleSendToIntruder);
});
</script>

<template>
  <div class="app-container">
    <RouterView />

    <!-- 全局通知弹窗 -->
    <Transition name="notification-slide">
      <div
        v-if="globalNotification.visible"
        class="global-notification"
        :class="globalNotification.type"
      >
        <div class="notification-content">
          <i
            class="bx notification-icon"
            :class="{
              'bx-check-circle': globalNotification.type === 'success',
              'bx-error-circle': globalNotification.type === 'error',
              'bx-info-circle': globalNotification.type === 'info',
              'bx-error': globalNotification.type === 'warning'
            }"
          ></i>
          <span class="notification-message">{{ globalNotification.message }}</span>
        </div>
        <button class="notification-close" @click="globalNotification.visible = false">
          <i class="bx bx-x"></i>
        </button>
      </div>
    </Transition>
  </div>
</template>

<style scoped>
.app-container {
  height: 100vh;
  width: 100vw;
  overflow: auto; /* 允许滚动 */
}

/* 全局通知弹窗样式 */
.global-notification {
  position: fixed;
  top: 20px;
  right: 20px;
  z-index: 9999;
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 12px 16px;
  border-radius: 8px;
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.15);
  min-width: 300px;
  max-width: 450px;
  backdrop-filter: blur(10px);
}

.global-notification.success {
  background: linear-gradient(135deg, rgba(34, 197, 94, 0.95), rgba(22, 163, 74, 0.95));
  color: white;
}

.global-notification.error {
  background: linear-gradient(135deg, rgba(239, 68, 68, 0.95), rgba(220, 38, 38, 0.95));
  color: white;
}

.global-notification.info {
  background: linear-gradient(135deg, rgba(59, 130, 246, 0.95), rgba(37, 99, 235, 0.95));
  color: white;
}

.global-notification.warning {
  background: linear-gradient(135deg, rgba(245, 158, 11, 0.95), rgba(217, 119, 6, 0.95));
  color: white;
}

.notification-content {
  display: flex;
  align-items: center;
  gap: 10px;
  flex: 1;
}

.notification-icon {
  font-size: 20px;
}

.notification-message {
  font-size: 14px;
  font-weight: 500;
}

.notification-close {
  background: transparent;
  border: none;
  color: white;
  cursor: pointer;
  padding: 4px;
  display: flex;
  align-items: center;
  justify-content: center;
  opacity: 0.8;
  transition: opacity 0.2s;
}

.notification-close:hover {
  opacity: 1;
}

.notification-close i {
  font-size: 18px;
}

/* 通知弹窗动画 */
.notification-slide-enter-active,
.notification-slide-leave-active {
  transition: all 0.3s ease;
}

.notification-slide-enter-from {
  opacity: 0;
  transform: translateX(100%);
}

.notification-slide-leave-to {
  opacity: 0;
  transform: translateX(100%);
}
</style>
