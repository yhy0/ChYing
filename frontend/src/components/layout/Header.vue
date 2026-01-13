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
  
  console.log('启动Chrome浏览器，代理地址:', proxyUrl.value);
  
  // 调用后端函数启动Chrome浏览器
  OpenChromeBrowser(proxyUrl.value)
    .then(() => {
      console.log('Chrome浏览器启动成功');
      showChromeModal.value = false;
    })
    .catch((error: Error) => {
      console.error('启动Chrome浏览器失败:', error);
      alert('启动Chrome浏览器失败: ' + error.message);
    })
    .finally(() => {
      isLaunchingBrowser.value = false;
    });
};

// 打开Chrome配置模态框
const openChromeModal = () => {
  showChromeModal.value = true;
};

// 打开Claude Agent窗口
const openClaudeAgent = () => {
  NewClaudeAgentWindow([])  // 传递空数组表示不带流量 ID
    .then(() => {
      console.log('Claude Agent窗口已打开');
    })
    .catch((error: Error) => {
      console.error('打开Claude Agent窗口失败:', error);
    });
};

// 打开配置目录
const openConfigDirectory = () => {
  OpenConfigDir()
    .then(() => {
      console.log('配置目录已打开');
    })
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

// 通知计数（初始为0）
const unreadCount = ref(0);

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
  <header class="px-4 py-2 border-b border-gray-200 dark:border-gray-700 flex items-center justify-between bg-white dark:bg-gray-900 shadow-sm">
    <div class="flex items-center">
      <h1 class="text-base font-medium text-gray-800 dark:text-gray-100">{{ moduleInfo.title }}</h1>
      <div class="h-3.5 mx-2 border-r border-gray-200 dark:border-gray-700"></div>
      <p class="text-xs text-gray-500 dark:text-gray-400">{{ moduleInfo.description }}</p>
    </div>
    <div class="flex items-center space-x-2">

      <!-- Claude AI Agent 按钮 -->
      <div class="tooltip-container">
        <button
          class="btn p-1 rounded-md text-gray-600 dark:text-gray-300 hover:text-gray-900 dark:hover:text-white hover:bg-gray-100 dark:hover:bg-gray-800 transition-colors duration-200"
          @click="openClaudeAgent"
          :aria-label="t('common.ui.claudeAgent')"
        >
          <i class="bx bx-bot text-base flex items-center justify-center w-5 h-5"></i>
        </button>
        <span class="tooltip-text tooltip-bottom">{{ t('common.ui.claudeAgent') }}</span>
      </div>

      <!-- 新增启动Chrome浏览器按钮 -->
      <div class="tooltip-container">
        <button
          class="btn p-1 rounded-md text-gray-600 dark:text-gray-300 hover:text-gray-900 dark:hover:text-white hover:bg-gray-100 dark:hover:bg-gray-800 transition-colors duration-200"
          @click="openChromeModal"
          :aria-label="t('common.ui.launchChrome')"
        >
          <i class="bx bx-globe text-base flex items-center justify-center w-5 h-5"></i>
        </button>
        <span class="tooltip-text tooltip-bottom">{{ t('common.ui.launchChrome') }}</span>
      </div>

      <!-- 打开配置目录按钮 -->
      <div class="tooltip-container">
        <button
          class="btn p-1 rounded-md text-gray-600 dark:text-gray-300 hover:text-gray-900 dark:hover:text-white hover:bg-gray-100 dark:hover:bg-gray-800 transition-colors duration-200"
          @click="openConfigDirectory"
          :aria-label="t('common.ui.openConfigDir')"
        >
          <i class="bx bx-folder-open text-base flex items-center justify-center w-5 h-5"></i>
        </button>
        <span class="tooltip-text tooltip-bottom">{{ t('common.ui.openConfigDir') }}</span>
      </div>

      <!-- 通知按钮 -->
      <div class="tooltip-container">
        <button
          class="btn p-1 rounded-md text-gray-600 dark:text-gray-300 hover:text-gray-900 dark:hover:text-white hover:bg-gray-100 dark:hover:bg-gray-800 transition-colors duration-200 relative"
          @click="toggleNotifications"
          :aria-label="t('common.ui.notifications')"
        >
          <i class="bx bx-bell text-base flex items-center justify-center w-5 h-5"></i>
          <!-- 未读消息数量 -->
          <div
            v-if="unreadCount > 0"
            class="absolute -top--0.5 -right--0.5 w-4 h-4 rounded-full bg-red-500 text-white text-xs flex items-center justify-center"
          >
            {{ unreadCount > 9 ? '9+' : unreadCount }}
          </div>
        </button>
        <span class="tooltip-text tooltip-bottom">{{ t('common.ui.notifications') }}</span>
      </div>

      <!-- Language Toggle -->
      <div class="tooltip-container">
        <button
          class="btn p-1 rounded-md text-gray-600 dark:text-gray-300 hover:text-gray-900 dark:hover:text-white hover:bg-gray-100 dark:hover:bg-gray-800 transition-colors duration-200"
          @click="toggleLanguage"
          :aria-label="language === 'en' ? t('common.language.switch_to_zh') : t('common.language.switch_to_en')"
        >
          <span class="bx text-base flex items-center justify-center w-5 h-5">{{ language === 'en' ? '🇺🇸' : '🇨🇳' }}</span>
        </button>
        <span class="tooltip-text tooltip-bottom">{{ language === 'en' ? t('common.language.switch_to_zh') : t('common.language.switch_to_en') }}</span>
      </div>

      <!-- Theme Toggle -->
      <div class="tooltip-container">
        <button
          class="btn p-1 rounded-md text-gray-600 dark:text-gray-300 hover:text-gray-900 dark:hover:text-white hover:bg-gray-100 dark:hover:bg-gray-800 transition-colors duration-200"
          @click="toggleTheme"
          :aria-label="getThemeButtonTitle()"
        >
          <i class="bx text-base flex items-center justify-center w-5 h-5" :class="getThemeIcon()"></i>
        </button>
        <span class="tooltip-text tooltip-bottom">{{ getThemeButtonTitle() }}</span>
      </div>
    </div>
  </header>

      <!-- Chrome浏览器配置弹窗 -->
      <div v-if="showChromeModal" class="fixed inset-0 z-50 flex items-center justify-center bg-black/50">
        <div class="bg-white dark:bg-gray-900 rounded-lg shadow-xl w-full max-w-md overflow-hidden">
          <div class="flex justify-between items-center px-4 py-3 border-b border-gray-200 dark:border-gray-700 bg-gray-50 dark:bg-gray-800">
            <h3 class="text-sm font-medium text-gray-700 dark:text-gray-300">启动Chrome浏览器</h3>
            <button 
              @click="showChromeModal = false" 
              class="btn btn-icon text-gray-500 hover:text-gray-700 dark:text-gray-400 dark:hover:text-gray-300 w-6 h-6"
              :disabled="isLaunchingBrowser"
            >
              <i class="bx bx-x"></i>
            </button>
          </div>
          <div class="p-5">
            <div class="mb-5">
              <label class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">
                代理地址
              </label>
              <input 
                v-model="proxyUrl" 
                type="text" 
                class="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-md shadow-sm focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-blue-500 dark:bg-gray-800 dark:text-gray-100"
                placeholder="http://127.0.0.1:9080"
                spellcheck="false"
              />
              <p class="mt-1 text-xs text-gray-500 dark:text-gray-400">
                Chrome将使用此代理地址并添加 --ignore-certificate-errors 参数启动
              </p>
            </div>
            
            <div class="flex justify-end space-x-3">
              <button 
                @click="showChromeModal = false" 
                class="btn btn-secondary px-4 py-2 text-sm"
                :disabled="isLaunchingBrowser"
              >
                取消
              </button>
              <button 
                @click="launchChrome" 
                class="btn btn-primary bg-blue-600 hover:bg-blue-700 text-white px-4 py-2 text-sm"
                :disabled="isLaunchingBrowser"
              >
                <i class="bx bx-loader bx-spin mr-1" v-if="isLaunchingBrowser"></i>
                {{ isLaunchingBrowser ? '启动中...' : '启动Chrome' }}
              </button>
            </div>
          </div>
        </div>
      </div>
</template>

<style scoped>
header {
  backdrop-filter: blur(5px);
  -webkit-backdrop-filter: blur(5px);
}
</style> 