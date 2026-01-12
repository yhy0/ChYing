<script setup lang="ts">
import { ref, computed } from 'vue';
import { useRoute } from 'vue-router';
import { useI18n } from 'vue-i18n';
import Sidebar from './Sidebar.vue';
import Header from './Header.vue';
import Footer from './Footer.vue';
import NotificationDrawer from './NotificationDrawer.vue';
import { ErrorBoundary } from '../common';
import { useModulesStore } from '../../store/modules';

const { t } = useI18n();
const route = useRoute();
const store = useModulesStore();

// 显示通知抽屉的状态
const isNotificationDrawerVisible = ref(false);

// 打开/关闭通知抽屉
const toggleNotifications = () => {
  isNotificationDrawerVisible.value = !isNotificationDrawerVisible.value;
};

// 关闭通知抽屉
const closeNotifications = () => {
  isNotificationDrawerVisible.value = false;
};

// 获取当前模块名称（用于错误边界显示）
const currentModuleName = computed(() => {
  const meta = route.meta;
  return meta && meta.title ? t(meta.title as string) : t('modules.project.title');
});

// 错误处理函数
interface ErrorInfoForHandler {
  error: Error;
  info?: string;
}

const handleModuleError = (errorInfo: ErrorInfoForHandler) => {
  console.error(
    'Module error caught by ErrorBoundary:',
    errorInfo.error, 
    errorInfo.info ? `(Info: ${errorInfo.info})` : ''
  );
  store.addNotification?.({
    type: 'error',
    message: errorInfo.error.message || t('common.error.unknown')
  });
};
</script>

<template>
  <div class="font-sans transition-colors duration-200 h-screen bg-white dark:bg-[#1e1e2e] text-gray-900 dark:text-white flex flex-col app-layout-container">
    <div class="flex flex-1 overflow-hidden">
      <!-- Sidebar Navigation -->
      <div class="app-layout-sidebar">
        <Sidebar />
      </div>

      <!-- Main Content -->
      <div class="flex-1 flex flex-col overflow-hidden app-layout-main">
        <!-- Header -->
        <div class="app-layout-header flex-shrink-0">
          <Header @toggle-notifications="toggleNotifications" />
        </div>

        <!-- Module Content with Simplified Transition -->
        <div class="flex-1 overflow-hidden app-layout-content">
          <Transition 
            name="module-fade" 
            mode="out-in"
          >
            <!-- 使用ErrorBoundary包裹组件 -->
            <ErrorBoundary 
              :component-name="currentModuleName"
              :show-retry="true" 
              :show-details="true"
              :on-error="handleModuleError"
            >
              <RouterView :key="route.fullPath" />
            </ErrorBoundary>
          </Transition>
        </div>
        
        <!-- Footer -->
        <Footer />
      </div>
    </div>
    
    <!-- Global Notification Drawer -->
    <NotificationDrawer 
      :show="isNotificationDrawerVisible" 
      @close="closeNotifications"
    />
  </div>
</template>

<style scoped>
/* 模块切换过渡动画 - 简化版 */
.module-fade-enter-active,
.module-fade-leave-active {
  transition: opacity 0.25s ease;
}

.module-fade-enter-from,
.module-fade-leave-to {
  opacity: 0;
}

/* 布局类 */
.app-layout-container {
  display: flex;
  flex-direction: column;
  height: 100vh;
  width: 100vw;
  max-width: 100vw;
  max-height: 100vh;
  overflow: hidden;
}

.app-layout-sidebar {
  flex: 0 0 56px;
  width: 56px !important;
  min-width: 56px !important;
  max-width: 56px !important;
  overflow: visible;
  z-index: 50;
  position: relative;
}

.app-layout-main {
  flex: 1;
  min-width: 0;
  display: flex;
  flex-direction: column;
  overflow: hidden;
}

.app-layout-header {
  z-index: 40;
  position: relative;
}

.app-layout-content {
  flex: 1;
  min-height: 0;
  position: relative;
  z-index: 1;
}

.app-layout-footer {
  z-index: 30;
  position: relative;
}
</style> 