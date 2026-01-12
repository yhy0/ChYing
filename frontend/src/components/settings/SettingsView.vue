<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useI18n } from 'vue-i18n'

// 导入设置模块组件
import AppearanceSettings from './modules/AppearanceSettings.vue'
import ProxySettings from './modules/ProxySettings.vue'
import ApplicationSettings from './modules/ApplicationSettings.vue'
import NetworkSettings from './modules/NetworkSettings.vue'
import AISettings from './modules/AISettings.vue'
import AboutSettings from './modules/AboutSettings.vue'

const { t } = useI18n()

// 当前活跃的设置页
const activeTab = ref('appearance')

// 设置页面列表
const settingsTabs = [
  { 
    id: 'appearance', 
    name: t('settings.appearance'), 
    icon: 'bx-paint', 
    color: 'text-green-500' 
  },
  { 
    id: 'proxy', 
    name: t('modules.proxy.settings'), 
    icon: 'bx-network-chart', 
    color: 'text-amber-500' 
  },
  { 
    id: 'application', 
    name: t('settings.application_behavior'), 
    icon: 'bx-cog', 
    color: 'text-blue-500' 
  },
  { 
    id: 'network', 
    name: t('settings.network'), 
    icon: 'bx-wifi', 
    color: 'text-purple-500' 
  },
  { 
    id: 'ai', 
    name: t('settings.ai.title', 'AI Assistant'), 
    icon: 'bx-bot', 
    color: 'text-indigo-500' 
  },
  { 
    id: 'about', 
    name: t('settings.about'), 
    icon: 'bx-info-circle', 
    color: 'text-gray-500' 
  }
]

// 切换设置页
const switchTab = (tabId: string) => {
  activeTab.value = tabId
}

// 组件挂载
onMounted(() => {
  // 可以在这里添加初始化逻辑
})
</script>

<template>
  <div class="h-screen flex bg-gray-50 dark:bg-[#1e1e2e]">
    <!-- 侧边栏 -->
    <div class="w-72 bg-white dark:bg-[#282838] border-r border-gray-200 dark:border-gray-700 overflow-y-auto">
      <!-- 设置标题 -->
      <div class="p-6 border-b border-gray-200 dark:border-gray-700">
        <h1 class="text-xl font-semibold text-gray-900 dark:text-white flex items-center">
          <i class="bx bx-cog mr-3 text-2xl text-blue-500"></i>
          {{ t('settings.title', '设置') }}
        </h1>
      </div>
      
      <!-- 设置导航 -->
      <div class="p-4">
        <nav class="space-y-2">
          <button
            v-for="tab in settingsTabs"
            :key="tab.id"
            @click="switchTab(tab.id)"
            class="w-full flex items-center px-4 py-3 text-left rounded-lg transition-colors duration-200"
            :class="[
              activeTab === tab.id 
                ? 'bg-blue-50 dark:bg-blue-900/20 text-blue-700 dark:text-blue-400 border border-blue-200 dark:border-blue-800' 
                : 'text-gray-700 dark:text-gray-300 hover:bg-gray-50 dark:hover:bg-gray-800/50'
            ]"
          >
            <i :class="`bx ${tab.icon} ${tab.color} mr-3 text-lg`"></i>
            <span class="font-medium">{{ tab.name }}</span>
          </button>
        </nav>
      </div>
    </div>
    
    <!-- 主内容区域 -->
    <div class="flex-1 overflow-y-auto">
      <div class="p-8">
        <!-- 外观设置 -->
        <AppearanceSettings v-if="activeTab === 'appearance'" />
        
        <!-- 代理设置 -->
        <ProxySettings v-else-if="activeTab === 'proxy'" />
        
        <!-- 应用设置 -->
        <ApplicationSettings v-else-if="activeTab === 'application'" />
        
        <!-- 网络设置 -->
        <NetworkSettings v-else-if="activeTab === 'network'" />
        
        <!-- AI 设置 -->
        <AISettings v-else-if="activeTab === 'ai'" />
        
        <!-- 关于 -->
        <AboutSettings v-else-if="activeTab === 'about'" />
      </div>
    </div>
  </div>
</template>

<style scoped>
/* 自定义滚动条样式 */
::-webkit-scrollbar {
  width: 6px;
  height: 6px;
}

::-webkit-scrollbar-track {
  background: transparent;
}

::-webkit-scrollbar-thumb {
  background-color: rgba(156, 163, 175, 0.5);
  border-radius: 3px;
}

::-webkit-scrollbar-thumb:hover {
  background-color: rgba(156, 163, 175, 0.7);
}

/* 深色模式下的滚动条 */
.dark ::-webkit-scrollbar-thumb {
  background-color: rgba(75, 85, 99, 0.5);
}

.dark ::-webkit-scrollbar-thumb:hover {
  background-color: rgba(75, 85, 99, 0.7);
}
</style>
