<script setup lang="ts">
import { ref } from 'vue';
import { useI18n } from 'vue-i18n';
import JwtAnalyzer from './modules/JwtAnalyzer.vue';
import FuzzScanner from './modules/FuzzScanner.vue';
import APIGenerate from './modules/APIGenerate.vue';
import DevModule from './modules/DevModule.vue';
import AuthorizationChecker from './modules/AuthorizationChecker.vue';

const { t } = useI18n();

// 当前选中的插件
const activePlugin = ref('jwt');

// 插件列表
const plugins = [
  { 
    id: 'jwt', 
    name: t('modules.plugins.jwt_analyzer.title'), 
    icon: 'bx-key',
    color: 'text-blue-500'
  },
  { 
    id: 'fuzz', 
    name: t('modules.plugins.fuzz.title'), 
    icon: 'bx-search-alt',
    color: 'text-green-500'
  },
  { 
    id: 'auth', 
    name: t('modules.plugins.auth_checker.title', '越权检测'), 
    icon: 'bx-shield-quarter',
    color: 'text-red-500'
  },
  { 
    id: 'apigen', 
    name: t('modules.plugins.api_generate.title', 'API 生成器'), 
    icon: 'bx-code-alt',
    color: 'text-purple-500'
  },
  { 
    id: 'sqlinjection', 
    name: t('modules.plugins.sqlinjection.title'), 
    icon: 'bx-data',
    color: 'text-amber-500'
  },
  { 
    id: 'xss', 
    name: t('modules.plugins.xss.title'), 
    icon: 'bx-code-block',
    color: 'text-purple-500'
  }
];
</script>

<template>
  <div class="h-full flex flex-col">
    <!-- 标签页导航 -->
    <div class="bg-white dark:bg-[#1e1e2e] border-b border-gray-200 dark:border-gray-700 py-2 px-2">
      <div class="flex overflow-x-auto scrollbar-thin">
        <button 
          v-for="plugin in plugins" 
          :key="plugin.id"
          @click="activePlugin = plugin.id"
          :class="[
            'btn',
            activePlugin === plugin.id 
              ? 'btn-primary' 
              : 'btn-secondary',
            'mx-1'
          ]"
        >
          <i :class="`bx ${plugin.icon} mr-1.5 ${plugin.color}`"></i>
          <span>{{ plugin.name }}</span>
        </button>
      </div>
    </div>
    
    <!-- 插件内容区域 -->
    <div class="flex-1 overflow-auto p-1 bg-white dark:bg-[#1e1e2e]">
      <keep-alive>
        <JwtAnalyzer v-if="activePlugin === 'jwt'" />
        <FuzzScanner v-else-if="activePlugin === 'fuzz'" />
        <AuthorizationChecker v-else-if="activePlugin === 'auth'" />
        <APIGenerate v-else-if="activePlugin === 'apigen'" />
        <DevModule v-else :moduleId="activePlugin" />
      </keep-alive>
    </div>
  </div>
</template>

<style scoped>
pre {
  white-space: pre-wrap;
  word-break: break-word;
}

.scrollbar-thin::-webkit-scrollbar {
  height: 6px;
}

.scrollbar-thin::-webkit-scrollbar-track {
  background: transparent;
}

.scrollbar-thin::-webkit-scrollbar-thumb {
  background: #d1d5db;
  border-radius: 3px;
}

.dark .scrollbar-thin::-webkit-scrollbar-thumb {
  background: #4b5563;
}

.scrollbar-thin::-webkit-scrollbar-thumb:hover {
  background: #9ca3af;
}

.dark .scrollbar-thin::-webkit-scrollbar-thumb:hover {
  background: #6b7280;
}

@keyframes pulse {
  0%, 100% {
    opacity: 1;
  }
  50% {
    opacity: 0.7;
  }
}

.animate-pulse {
  animation: pulse 2s cubic-bezier(0.4, 0, 0.6, 1) infinite;
}
</style> 