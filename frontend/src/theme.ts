// 获取当前主题设置
export function getCurrentTheme(): 'light' | 'dark' | 'system' {
  return localStorage.getItem('app-theme') as 'light' | 'dark' | 'system' || 'system';
}

// 设置主题
export function setTheme(themeMode: 'light' | 'dark' | 'system'): void {
  // 保存当前主题选择
  localStorage.setItem('app-theme', themeMode);
  
  // 应用主题
  const isDarkMode = themeMode === 'dark' || (themeMode === 'system' && window.matchMedia('(prefers-color-scheme: dark)').matches);
  
  if (isDarkMode) {
    document.documentElement.classList.add('dark');
  } else {
    document.documentElement.classList.remove('dark');
  }
  
  // 触发跨窗口同步事件（包括当前窗口）
  window.dispatchEvent(new StorageEvent('storage', {
    key: 'app-theme',
    newValue: themeMode,
    oldValue: null,
    storageArea: localStorage,
    url: window.location.href
  }));
}


// 防止重复初始化的标志
let isInitialized = false;

// 初始化主题
export function initTheme(): void {
  if (isInitialized) {
    return;
  }
  
  const savedTheme = getCurrentTheme();
  // 初始化时直接应用主题，不触发事件
  const isDarkMode = savedTheme === 'dark' || (savedTheme === 'system' && window.matchMedia('(prefers-color-scheme: dark)').matches);
  if (isDarkMode) {
    document.documentElement.classList.add('dark');
  } else {
    document.documentElement.classList.remove('dark');
  }
  
  // 监听系统主题变化
  const mediaQuery = window.matchMedia('(prefers-color-scheme: dark)');
  const handleSystemThemeChange = () => {
    const currentTheme = getCurrentTheme();
    if (currentTheme === 'system') {
      setTheme('system'); // 重新应用系统主题
    }
  };
  
  // 添加监听器
  if (mediaQuery.addEventListener) {
    mediaQuery.addEventListener('change', handleSystemThemeChange);
  } else {
    // 兼容旧浏览器
    mediaQuery.addListener(handleSystemThemeChange);
  }
  
  isInitialized = true;
} 