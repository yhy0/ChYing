import { createI18n } from 'vue-i18n';
import en from './locales/en.json';
import zh from './locales/zh.json';

// Get saved language preference from localStorage or use browser language
const getBrowserLanguage = (): string => {
  const savedLang = localStorage.getItem('language');
  if (savedLang && ['en', 'zh'].includes(savedLang)) {
    return savedLang;
  }
  
  const browserLang = navigator.language.toLowerCase();
  return browserLang.startsWith('zh') ? 'zh' : 'en';
};

// Create i18n instance
export const i18n = createI18n({
  legacy: false, // Use composition API
  locale: getBrowserLanguage(), // Set the locale
  fallbackLocale: 'en', // Set fallback locale
  messages: {
    en,
    zh
  }
});

// Function to change language dynamically
export const setLanguage = (lang: 'en' | 'zh'): void => {
  i18n.global.locale.value = lang;
  localStorage.setItem('language', lang);
  document.querySelector('html')?.setAttribute('lang', lang);
  
  // 触发跨窗口同步事件（包括当前窗口）
  window.dispatchEvent(new StorageEvent('storage', {
    key: 'language',
    newValue: lang,
    oldValue: null,
    storageArea: localStorage,
    url: window.location.href
  }));
}; 