<script setup lang="ts">
import { computed, ref, onMounted, onUnmounted } from 'vue';
import { useRoute } from 'vue-router';
import { useI18n } from 'vue-i18n';
import { setLanguage } from '../../i18n';
import { setTheme, getCurrentTheme } from '../../theme';
import { useModulesStore } from '../../store/modules';
// @ts-ignore - 忽略模块导入错误
import { OpenChromeBrowser, NewClaudeAgentWindow, GetConfigStatus, OpenConfigDir } from "../../../bindings/github.com/yhy0/ChYing/app.js";

const route = useRoute();
const store = useModulesStore();
const { t } = useI18n();

// Get module title and description from route meta
const moduleInfo = computed(() => {
  const meta = route.meta;
  if (meta && meta.title) {
    return {
      title: t(meta.title as string),
      description: meta.description ? t(meta.description as string) : '',
    };
  }
  // 默认值
  return {
    title: t('modules.project.title'),
    description: t('modules.project.description'),
  };
});

// Chrome浏览器配置
const showChromeModal = ref(false);
const proxyUrl = ref('http://127.0.0.1:9080'); // 默认值，会在 onMounted 中更新
const isLaunchingBrowser = ref(false);

// 从后端获取代理端口配置
const loadProxyConfig = async () => {
  try {
    const status = await GetConfigStatus();
    if (status && status.proxy_address) {
      proxyUrl.value = `http://${status.proxy_address}`;
    }
  } catch (error) {
    console.error('获取代理配置失败:', error);
  }
};

// 启动Chrome浏览器
const launchChrome = () => {
  isLaunchingBrowser.value = true;
  
  // 调用后端函数启动Chrome浏览器
  OpenChromeBrowser(proxyUrl.value)
    .then(() => {
      showChromeModal.value = false;
    })
    .catch((error: Error) => {
      console.error('启动Chrome浏览器失败:', error);
      alert(t('layout.header.chrome_launch_failed') + error.message);
    })
    .finally(() => {
      isLaunchingBrowser.value = false;
    });
};

// 打开Chrome配置模态框
const openChromeModal = async () => {
  // 每次打开模态框时重新获取当前代理配置
  await loadProxyConfig();
  showChromeModal.value = true;
};

// 打开Claude Agent窗口
const openClaudeAgent = () => {
  NewClaudeAgentWindow([])  // 传递空数组表示不带流量 ID
    .catch((error: Error) => {
      console.error('打开Claude Agent窗口失败:', error);
    });
};

// 打开配置目录
const openConfigDirectory = () => {
  OpenConfigDir()
    .catch((error: Error) => {
      console.error('打开配置目录失败:', error);
    });
};


// Current language - keep as reactive reference
const language = useI18n().locale;

// Toggle language
const toggleLanguage = () => {
  const newLang = language.value === 'en' ? 'zh' : 'en';
  setLanguage(newLang);
};

// 主题管理
const currentTheme = ref(getCurrentTheme());

// 计算是否为暗色模式
const isDarkMode = computed(() => {
  const theme = currentTheme.value;
  if (theme === 'system') {
    return window.matchMedia('(prefers-color-scheme: dark)').matches;
  }
  return theme === 'dark';
});

// 切换主题（循环切换：light -> dark -> system -> light）
const toggleTheme = () => {
  const current = currentTheme.value;
  let nextTheme: 'light' | 'dark' | 'system';
  
  switch (current) {
    case 'light':
      nextTheme = 'dark';
      break;
    case 'dark':
      nextTheme = 'system';
      break;
    case 'system':
    default:
      nextTheme = 'light';
      break;
  }
  
  currentTheme.value = nextTheme;
  setTheme(nextTheme);
};

// 获取主题按钮的标题
const getThemeButtonTitle = () => {
  const theme = currentTheme.value;
  switch (theme) {
    case 'light':
      return t('common.theme.light') + ' → ' + t('common.theme.dark');
    case 'dark':
      return t('common.theme.dark') + ' → ' + t('common.theme.system');
    case 'system':
      return t('common.theme.system') + ' → ' + t('common.theme.light');
    default:
      return '';
  }
};

// 获取主题按钮的图标
const getThemeIcon = () => {
  const theme = currentTheme.value;
  switch (theme) {
    case 'light':
      return 'bx-sun';
    case 'dark':
      return 'bx-moon';
    case 'system':
      return 'bx-desktop';
    default:
      return isDarkMode.value ? 'bx-moon' : 'bx-sun';
  }
};

// 通知计数（从 store 获取）
const unreadCount = computed(() => store.notifications.unreadCount);

// 定义向上发送事件的emit
const emit = defineEmits(['toggleNotifications']);

// 打开/关闭通知抽屉
const toggleNotifications = () => {
  // 向上触发事件，由App.vue处理
  emit('toggleNotifications');
};

// localStorage变化处理函数（统一的主题和语言同步）
const handleStorageChange = (e: StorageEvent) => {
  if (e.key === 'app-theme' && e.newValue) {
    currentTheme.value = getCurrentTheme();
  }
  
  if (e.key === 'language' && e.newValue) {
    const newLang = e.newValue as 'en' | 'zh';
    // 同步语言设置（Header使用的是composition API的locale）
    language.value = newLang;
    document.querySelector('html')?.setAttribute('lang', newLang);
  }
};

// 组件挂载时添加监听器
onMounted(() => {
  // 监听storage事件（包括所有窗口的主题变化）
  window.addEventListener('storage', handleStorageChange);

  // 加载代理配置
  loadProxyConfig();

  // 保存清理函数
  onUnmounted(() => {
    window.removeEventListener('storage', handleStorageChange);
  });
});
</script>

<template>
  <header class="app-header">
    <div class="header-title-group">
      <h1 class="header-title">{{ moduleInfo.title }}</h1>
      <div class="header-divider"></div>
      <p class="header-description">{{ moduleInfo.description }}</p>
    </div>
    <div class="header-actions">

      <!-- Claude AI Agent 按钮 -->
      <div class="tooltip-container">
        <button
          class="btn btn-icon btn-icon-sm"
          @click="openClaudeAgent"
          :aria-label="t('common.ui.claudeAgent')"
        >
          <i class="bx bx-bot"></i>
        </button>
        <span class="tooltip-text tooltip-bottom">{{ t('common.ui.claudeAgent') }}</span>
      </div>

      <!-- 新增启动Chrome浏览器按钮 -->
      <div class="tooltip-container">
        <button
          class="btn btn-icon btn-icon-sm"
          @click="openChromeModal"
          :aria-label="t('common.ui.launchChrome')"
        >
          <i class="bx bx-globe"></i>
        </button>
        <span class="tooltip-text tooltip-bottom">{{ t('common.ui.launchChrome') }}</span>
      </div>

      <!-- 打开配置目录按钮 -->
      <div class="tooltip-container">
        <button
          class="btn btn-icon btn-icon-sm"
          @click="openConfigDirectory"
          :aria-label="t('common.ui.openConfigDir')"
        >
          <i class="bx bx-folder-open"></i>
        </button>
        <span class="tooltip-text tooltip-bottom">{{ t('common.ui.openConfigDir') }}</span>
      </div>

      <!-- 通知按钮 -->
      <div class="tooltip-container overflow-visible">
        <button
          class="btn btn-icon btn-icon-sm notification-btn"
          @click="toggleNotifications"
          :aria-label="t('common.ui.notifications')"
        >
          <i class="bx bx-bell"></i>
          <!-- 未读消息数量 -->
          <span
            v-if="unreadCount > 0"
            class="notification-count"
          >
            {{ unreadCount > 9 ? '9+' : unreadCount }}
          </span>
        </button>
        <span class="tooltip-text tooltip-bottom">{{ t('common.ui.notifications') }}</span>
      </div>

      <!-- Language Toggle -->
      <div class="tooltip-container">
        <button
          class="btn btn-icon btn-icon-sm"
          @click="toggleLanguage"
          :aria-label="language === 'en' ? t('common.language.switch_to_zh') : t('common.language.switch_to_en')"
        >
          <span class="language-glyph">{{ language === 'en' ? '🇺🇸' : '🇨🇳' }}</span>
        </button>
        <span class="tooltip-text tooltip-bottom">{{ language === 'en' ? t('common.language.switch_to_zh') : t('common.language.switch_to_en') }}</span>
      </div>

      <!-- Theme Toggle -->
      <div class="tooltip-container">
        <button
          class="btn btn-icon btn-icon-sm"
          @click="toggleTheme"
          :aria-label="getThemeButtonTitle()"
        >
          <i class="bx" :class="getThemeIcon()"></i>
        </button>
        <span class="tooltip-text tooltip-bottom">{{ getThemeButtonTitle() }}</span>
      </div>
    </div>
  </header>

  <!-- Chrome浏览器配置弹窗 -->
  <div
    v-if="showChromeModal"
    class="dialog-overlay"
    role="dialog"
    aria-modal="true"
    :aria-label="t('layout.header.chrome_dialog_title')"
    @keydown.esc="showChromeModal = false"
  >
    <div class="dialog-container dialog-sm" @click.stop>
      <div class="dialog-header">
        <h3 class="dialog-title">{{ t('layout.header.chrome_dialog_title') }}</h3>
        <button
          @click="showChromeModal = false"
          class="btn btn-icon btn-icon-xs"
          :disabled="isLaunchingBrowser"
          :aria-label="t('common.actions.close')"
        >
          <i class="bx bx-x"></i>
        </button>
      </div>
      <div class="dialog-content">
        <div class="form-field">
          <label class="form-label" for="chrome-proxy-url">代理地址</label>
          <input
            id="chrome-proxy-url"
            v-model="proxyUrl"
            type="text"
            class="form-input"
            placeholder="http://127.0.0.1:9080"
            spellcheck="false"
          />
          <p class="field-hint">Chrome将使用此代理地址并添加 --ignore-certificate-errors 参数启动</p>
        </div>
      </div>
      <div class="dialog-footer">
        <button
          @click="showChromeModal = false"
          class="btn btn-secondary btn-sm"
          :disabled="isLaunchingBrowser"
        >
          取消
        </button>
        <button
          @click="launchChrome"
          class="btn btn-primary btn-sm"
          :disabled="isLaunchingBrowser"
        >
          <i class="bx bx-loader bx-spin" v-if="isLaunchingBrowser"></i>
          {{ isLaunchingBrowser ? t('layout.header.launching') : t('layout.header.launch_chrome') }}
        </button>
      </div>
    </div>
  </div>
</template>

<style scoped>
.app-header {
  min-height: 2.75rem;
  padding: 0.5rem 1rem;
  border-bottom: 1px solid var(--color-border);
  background: var(--glass-bg-card);
  backdrop-filter: var(--glass-blur-light);
  -webkit-backdrop-filter: var(--glass-blur-light);
  box-shadow: var(--shadow-sm);
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 1rem;
}

.header-title-group {
  min-width: 0;
  display: flex;
  align-items: center;
  gap: 0.5rem;
}

.header-title {
  margin: 0;
  font-size: 0.9375rem;
  font-weight: 600;
  color: var(--color-text-primary);
  white-space: nowrap;
}

.header-divider {
  width: 1px;
  height: 0.875rem;
  background: var(--color-border);
}

.header-description {
  margin: 0;
  font-size: 0.75rem;
  color: var(--color-text-secondary);
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.header-actions {
  display: flex;
  align-items: center;
  gap: 0.5rem;
}

.notification-btn {
  position: relative;
  overflow: visible !important;
}

.notification-count {
  position: absolute;
  top: -0.375rem;
  right: -0.375rem;
  min-width: 1rem;
  height: 1rem;
  padding: 0 0.25rem;
  border-radius: var(--radius-full);
  background: var(--color-danger);
  color: var(--color-primary-text);
  font-size: 0.625rem;
  font-weight: 600;
  line-height: 1rem;
  pointer-events: none;
}

.language-glyph {
  font-size: 0.875rem;
  line-height: 1;
}
</style> 
