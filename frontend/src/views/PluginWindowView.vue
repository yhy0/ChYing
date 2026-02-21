<script setup lang="ts">
import { computed, onMounted, onUnmounted } from 'vue';
import { useRoute } from 'vue-router';
import { useI18n } from 'vue-i18n';
import { i18n } from '../i18n';
import JwtAnalyzer from '../components/plugins/modules/JwtAnalyzer.vue';
import FuzzScanner from '../components/plugins/modules/FuzzScanner.vue';
import APIGenerate from '../components/plugins/modules/APIGenerate.vue';
import DevModule from '../components/plugins/modules/DevModule.vue';
import AuthorizationChecker from '../components/plugins/modules/AuthorizationChecker.vue';
import PluginsView from '../components/plugins/PluginsView.vue';

const route = useRoute();
const { t } = useI18n();

const pluginId = computed(() => route.params.pluginId as string);
const isAllPlugins = computed(() => pluginId.value === 'all');

// 插件元数据
const pluginMeta: Record<string, { nameKey: string; icon: string; color: string }> = {
  jwt: { nameKey: 'modules.plugins.jwt_analyzer.title', icon: 'bx-key', color: 'text-blue-500' },
  fuzz: { nameKey: 'modules.plugins.fuzz.title', icon: 'bx-search-alt', color: 'text-green-500' },
  auth: { nameKey: 'modules.plugins.auth_checker.title', icon: 'bx-shield-quarter', color: 'text-red-500' },
  apigen: { nameKey: 'modules.plugins.api_generate.title', icon: 'bx-code-alt', color: 'text-purple-500' },
  sqlinjection: { nameKey: 'modules.plugins.sqlinjection.title', icon: 'bx-data', color: 'text-amber-500' },
  xss: { nameKey: 'modules.plugins.xss.title', icon: 'bx-code-block', color: 'text-purple-500' },
};

const currentPlugin = computed(() => pluginMeta[pluginId.value]);
const pluginName = computed(() => {
  if (isAllPlugins.value) {
    return t('modules.plugins.title', '插件');
  }
  const meta = currentPlugin.value;
  return meta ? t(meta.nameKey) : pluginId.value;
});

// localStorage 监听：跨窗口主题/语言同步
const handleStorageChange = (e: StorageEvent) => {
  if (e.key === 'app-theme' && e.newValue) {
    const newTheme = e.newValue as 'light' | 'dark' | 'system';
    const isDark = newTheme === 'dark' || (newTheme === 'system' && window.matchMedia('(prefers-color-scheme: dark)').matches);
    if (isDark) {
      document.documentElement.classList.add('dark');
    } else {
      document.documentElement.classList.remove('dark');
    }
  }
  if (e.key === 'language' && e.newValue) {
    const newLang = e.newValue as 'en' | 'zh';
    i18n.global.locale.value = newLang;
    document.querySelector('html')?.setAttribute('lang', newLang);
  }
};

onMounted(() => {
  window.addEventListener('storage', handleStorageChange);
});

onUnmounted(() => {
  window.removeEventListener('storage', handleStorageChange);
});
</script>

<template>
  <div class="plugin-window">
    <!-- 标题栏（仅单插件模式显示） -->
    <div v-if="!isAllPlugins" class="plugin-header">
      <div class="header-title">
        <i v-if="currentPlugin" :class="`bx ${currentPlugin.icon} ${currentPlugin.color}`"></i>
        <h3>{{ pluginName }} - {{ t('modules.plugins.independent_window', '独立窗口') }}</h3>
      </div>
    </div>

    <!-- 插件内容 -->
    <div :class="isAllPlugins ? 'plugin-content-full' : 'plugin-content'">
      <!-- 整个插件系统模式：复用 PluginsView -->
      <PluginsView v-if="isAllPlugins" />
      <!-- 单插件模式 -->
      <JwtAnalyzer v-else-if="pluginId === 'jwt'" />
      <FuzzScanner v-else-if="pluginId === 'fuzz'" />
      <AuthorizationChecker v-else-if="pluginId === 'auth'" />
      <APIGenerate v-else-if="pluginId === 'apigen'" />
      <DevModule v-else :moduleId="pluginId" />
    </div>
  </div>
</template>

<style scoped>
.plugin-window {
  display: flex;
  flex-direction: column;
  height: 100vh;
  background: var(--color-bg-primary);
  color: var(--color-text-primary);
}

.plugin-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 12px 16px;
  border-bottom: 1px solid var(--glass-border-light);
  background: var(--glass-bg-secondary);
  backdrop-filter: var(--glass-blur-light);
  -webkit-backdrop-filter: var(--glass-blur-light);
  -webkit-app-region: drag;
}

.header-title {
  display: flex;
  align-items: center;
  gap: 8px;
  -webkit-app-region: no-drag;
}

.header-title i {
  font-size: 20px;
}

.header-title h3 {
  margin: 0;
  font-size: var(--text-base);
  font-weight: var(--font-weight-semibold);
  color: var(--color-text-primary);
}

.plugin-content {
  flex: 1;
  overflow: auto;
  padding: 4px;
  background: var(--color-bg-primary);
}

.plugin-content-full {
  flex: 1;
  overflow: hidden;
  background: var(--color-bg-primary);
}
</style>
