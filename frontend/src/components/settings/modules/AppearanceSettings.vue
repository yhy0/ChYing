<script setup lang="ts">
import { ref, onMounted, onUnmounted } from 'vue';
import { useI18n } from 'vue-i18n';
import { setLanguage } from '../../../i18n';
import { setTheme, getCurrentTheme } from '../../../theme';

const { t, locale } = useI18n();

// 主题设置
const selectedTheme = ref(getCurrentTheme());

// 应用主题变更
const updateTheme = (themeId: 'light' | 'dark' | 'system') => {
  selectedTheme.value = themeId;
  setTheme(themeId);
};

// 应用语言变更
const updateLanguage = (lang: 'en' | 'zh') => {
  setLanguage(lang);
};

// 主题选项
const themeOptions = [
  { id: 'light', label: t('common.theme.light'), icon: 'bx-sun' },
  { id: 'dark', label: t('common.theme.dark'), icon: 'bx-moon' },
  { id: 'system', label: t('common.theme.system'), icon: 'bx-desktop' }
];

// 语言选项
const languageOptions = [
  { id: 'en', label: t('common.language.en'), flag: '🇺🇸' },
  { id: 'zh', label: t('common.language.zh'), flag: '🇨🇳' }
];

// 界面设置
const showWelcomeScreen = ref(true);

// 字体大小设置 - 只保留自定义
const customFontSize = ref(16); // 默认值

// 设置自定义字体大小 (直接操作DOM并存入localStorage)
const setAndStoreCustomFontSize = (value: number) => {
  const root = document.documentElement;
  const remSize = value / 16; // 转换为rem单位
  
  root.style.setProperty('--font-size-base', `${remSize}rem`);
  root.style.setProperty('--font-size-small', `${remSize * 0.875}rem`);
  root.style.setProperty('--font-size-large', `${remSize * 1.125}rem`);
  
  // 强制重新计算样式
  document.body.style.display = 'none';
  document.body.offsetHeight; // 触发重排
  document.body.style.display = '';
  
  localStorage.setItem('app-custom-font-size', value.toString());
  console.log(`设置自定义字体大小为: ${value}px (${remSize}rem) 并已保存`);
};

// 应用自定义字体大小
const applyCustomFontSize = (value: number) => {
  customFontSize.value = value;
  setAndStoreCustomFontSize(value);
};




// localStorage变化处理函数（统一的主题和语言同步）
const handleStorageChange = (e: StorageEvent) => {
  if (e.key === 'app-theme' && e.newValue) {
    selectedTheme.value = getCurrentTheme();
  }
  
  if (e.key === 'language' && e.newValue) {
    const newLang = e.newValue as 'en' | 'zh';
    // 同步语言设置（AppearanceSettings使用的是composition API的locale）
    locale.value = newLang;
    document.querySelector('html')?.setAttribute('lang', newLang);
  }
};

// 组件初始化时获取最新设置
onMounted(() => {
  selectedTheme.value = getCurrentTheme();
  
  // 监听storage事件（包括所有窗口的主题变化）
  window.addEventListener('storage', handleStorageChange);

  // 初始化自定义字体大小
  const storedFontSize = localStorage.getItem('app-custom-font-size');
  if (storedFontSize) {
    customFontSize.value = parseInt(storedFontSize, 10);
  }
  
  // 首次加载时应用已保存或默认的设置
  setTimeout(() => {
    setAndStoreCustomFontSize(customFontSize.value);
  }, 100); // 短暂延迟确保DOM准备好
});

// 组件卸载时清理监听器
onUnmounted(() => {
  window.removeEventListener('storage', handleStorageChange);
});
</script>

<template>
  <div class="space-y-6 max-w-3xl">
    <h2 class="text-lg font-medium mb-5 flex items-center">
      <i class="bx bx-paint text-green-500 mr-2"></i>
      {{ t('settings.appearance') }}
    </h2>
    
    <div class="space-y-5">
      <!-- Theme Selection -->
      <div class="bg-white dark:bg-[#282838] rounded-lg p-4 shadow-sm border border-gray-100 dark:border-gray-700">
        <h3 class="text-sm font-medium mb-3 text-gray-700 dark:text-gray-300">{{ t('common.theme.light') }} / {{ t('common.theme.dark') }}</h3>
        
        <div class="grid grid-cols-3 gap-4">
          <div 
            v-for="option in themeOptions" 
            :key="option.id"
            class="p-4 rounded-lg border cursor-pointer transition-colors duration-200 text-center relative"
            :class="[
              selectedTheme === option.id 
                ? 'border-[#4f46e5] bg-[#4f46e5]/5 text-[#4f46e5] dark:text-[#818cf8]' 
                : 'border-gray-200 dark:border-gray-700 hover:bg-gray-50 dark:hover:bg-gray-800/50'
            ]"
            @click="updateTheme(option.id as 'light' | 'dark' | 'system')"
          >
            <i :class="`bx ${option.icon} text-2xl mb-2`"></i>
            <div class="text-sm">{{ option.label }}</div>
            <div v-if="selectedTheme === option.id" class="absolute top-2 right-2 text-[#4f46e5]">
              <i class="bx bx-check-circle"></i>
            </div>
          </div>
        </div>
      </div>
      
      <!-- Language Selection -->
      <div class="bg-white dark:bg-[#282838] rounded-lg p-4 shadow-sm border border-gray-100 dark:border-gray-700">
        <h3 class="text-sm font-medium mb-3 text-gray-700 dark:text-gray-300">{{ t('common.language.title') }}</h3>
        
        <div class="grid grid-cols-2 gap-4">
          <div 
            v-for="option in languageOptions" 
            :key="option.id"
            class="p-4 rounded-lg border cursor-pointer transition-colors duration-200 flex items-center"
            :class="[
              locale === option.id 
                ? 'border-[#4f46e5] bg-[#4f46e5]/5 text-[#4f46e5] dark:text-[#818cf8]' 
                : 'border-gray-200 dark:border-gray-700 hover:bg-gray-50 dark:hover:bg-gray-800/50'
            ]"
            @click="updateLanguage(option.id as 'en' | 'zh')"
          >
            <span class="text-xl mr-2">{{ option.flag }}</span>
            <span class="text-sm">{{ option.label }}</span>
          </div>
        </div>
      </div>
      
      <!-- 字体大小设置 - 只保留自定义输入 -->
      <div class="bg-white dark:bg-[#282838] rounded-lg p-4 shadow-sm border border-gray-100 dark:border-gray-700">
        <h3 class="text-sm font-medium mb-3 text-gray-700 dark:text-gray-300">{{ t('settings.font_size.title') }}</h3>
        
        <div class="space-y-4">
          <!-- 自定义字体大小数字输入 -->
          <div>
            <label class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">
              {{ t('settings.font_size.custom') }}
            </label>
            <div class="flex items-center space-x-2">
              <input 
                type="number" 
                v-model="customFontSize"
                min="10"
                max="24"
                step="1"
                class="w-24 px-3 py-2 bg-white dark:bg-[#323248] border border-gray-300 dark:border-gray-600 rounded-md shadow-sm text-sm focus:outline-none focus:ring-2 focus:ring-[#4f46e5]"
              />
              <span class="text-sm text-gray-600 dark:text-gray-400">px</span>
              <button 
                @click="applyCustomFontSize(customFontSize)"
                class="ml-2 px-3 py-2 bg-[#4f46e5] hover:bg-[#4338ca] text-white rounded-md text-sm transition-colors"
              >
                {{ t('common.actions.apply') }}
              </button>
            </div>
            <div class="text-xs text-gray-500 dark:text-gray-400 mt-1">
              {{ t('settings.font_size.custom_hint') }}
            </div>
          </div>
        </div>
        
        <!-- 字体预览 -->
        <div class="font-sample mt-6 p-4 border border-gray-200 dark:border-gray-700 rounded-lg bg-gray-50 dark:bg-[#1e1e36] text-gray-800 dark:text-gray-200" :style="{ fontSize: 'var(--font-size-base)' }">
          <div class="font-sample-heading font-medium mb-2">{{ t('settings.font_sample_heading') }}</div>
          <div class="font-sample-text mb-2">{{ t('settings.font_sample_text') }}</div>
          <div class="font-sample-code font-mono text-green-600 dark:text-green-400">console.log("{{ t('settings.font_sample_code') }}");</div>
        </div>
      </div>
      
      <!-- 界面设置分割线 -->
      <div class="relative my-6">
        <div class="absolute inset-0 flex items-center">
          <div class="w-full border-t border-gray-200 dark:border-gray-700"></div>
        </div>
        <div class="relative flex justify-center">
          <span class="bg-white dark:bg-[#1e1e2e] px-3 text-sm text-gray-500 dark:text-gray-400">
            {{ t('settings.interface_settings') }}
          </span>
        </div>
      </div>
      
      <!-- 显示启动屏幕 -->
      <div class="bg-white dark:bg-[#282838] rounded-lg p-4 shadow-sm border border-gray-100 dark:border-gray-700">
        <div class="flex items-center">
          <input 
            id="showWelcomeScreen" 
            v-model="showWelcomeScreen" 
            type="checkbox" 
            class="w-4 h-4 text-blue-600 border-gray-300 rounded focus:ring-blue-500 dark:focus:ring-blue-600 dark:ring-offset-gray-800 focus:ring-2 dark:bg-gray-700 dark:border-gray-600"
          >
          <label for="showWelcomeScreen" class="ml-2 text-sm font-medium text-gray-900 dark:text-gray-300">
            {{ t('settings.show_welcome_screen') }}
          </label>
        </div>
        
        <p class="text-xs text-gray-500 dark:text-gray-400 mt-2 ml-6">
          {{ t('settings.welcome_screen_description') }}
        </p>
      </div>
      
    </div>
  </div>
</template>

<style scoped>
/* 字体设置卡片样式 */
.font-preview {
  font-size: 1.5rem;
  font-weight: 500;
  margin-top: 0.5rem;
}

/* 字体样本样式 */
.font-sample-heading {
  font-size: 1.1em;
  font-weight: 500;
  margin-bottom: 0.5rem;
}

.font-sample-text {
  margin-bottom: 0.5rem;
}

.font-sample-code {
  font-family: var(--font-family-mono);
  padding: 0.5rem;
  background-color: rgba(0, 0, 0, 0.05);
  border-radius: 0.25rem;
}

.dark .font-sample-code {
  background-color: rgba(255, 255, 255, 0.05);
}
</style> 