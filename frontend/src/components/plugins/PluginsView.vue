<script setup lang="ts">
import { ref, reactive, onMounted, onUnmounted } from 'vue';
import { useI18n } from 'vue-i18n';
import JwtAnalyzer from './modules/JwtAnalyzer.vue';
import FuzzScanner from './modules/FuzzScanner.vue';
import APIGenerate from './modules/APIGenerate.vue';
import DevModule from './modules/DevModule.vue';
import AuthorizationChecker from './modules/AuthorizationChecker.vue';
import OastModule from './modules/OastModule.vue';
import { NewPluginWindow, IsPluginWindowOpen, NewPluginsWindow, IsPluginsWindowOpen } from '../../../bindings/github.com/yhy0/ChYing/app.js';
import { Events } from '@wailsio/runtime';

const { t } = useI18n();

// 当前选中的插件
const activePlugin = ref('jwt');

// 已弹出到独立窗口的插件集合
const poppedOutPlugins = reactive(new Set<string>());

// 整个插件系统是否已弹出
const allPluginsPopped = ref(false);

// 弹出整个插件系统到独立窗口
const popOutAllPlugins = async () => {
  try {
    await NewPluginsWindow();
    allPluginsPopped.value = true;
  } catch (err) {
    console.error('Failed to open plugins window:', err);
  }
};

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
  },
  {
    id: 'oast',
    name: t('modules.plugins.oast.title', 'OAST'),
    icon: 'bx-broadcast',
    color: 'text-cyan-500'
  }
];

// 弹出插件到独立窗口
const popOutPlugin = async (pluginId: string) => {
  try {
    await NewPluginWindow(pluginId);
    poppedOutPlugins.add(pluginId);
  } catch (err) {
    console.error('Failed to open plugin window:', err);
  }
};

// 拖拽相关
const handleDragStart = (e: DragEvent, pluginId: string) => {
  e.dataTransfer?.setData('text/plain', pluginId);
};

const handleDragEnd = (e: DragEvent, pluginId: string) => {
  // 检测是否拖到了窗口外
  const isOutsideWindow =
    e.clientX < 0 || e.clientY < 0 ||
    e.clientX > window.innerWidth || e.clientY > window.innerHeight;

  if (isOutsideWindow) {
    popOutPlugin(pluginId);
  }
};

// 监听窗口关闭事件
const handleWindowClosed = (event: any) => {
  const pluginId = event.data?.[0] || event.data;
  if (typeof pluginId === 'string') {
    poppedOutPlugins.delete(pluginId);
  }
};

const handleAllPluginsWindowClosed = () => {
  allPluginsPopped.value = false;
};

onMounted(async () => {
  // 监听 plugin:window-closed 事件
  Events.On('plugin:window-closed', handleWindowClosed);
  Events.On('plugins:window-closed', handleAllPluginsWindowClosed);

  // 初始化时检查哪些插件窗口已打开
  for (const plugin of plugins) {
    try {
      const isOpen = await IsPluginWindowOpen(plugin.id);
      if (isOpen) {
        poppedOutPlugins.add(plugin.id);
      }
    } catch {
      // 忽略错误
    }
  }

  // 检查整个插件系统窗口是否已打开
  try {
    allPluginsPopped.value = await IsPluginsWindowOpen();
  } catch {
    // 忽略错误
  }
});

onUnmounted(() => {
  Events.Off('plugin:window-closed');
  Events.Off('plugins:window-closed');
});
</script>

<template>
  <div class="h-full flex flex-col">
    <!-- 标签页导航 -->
    <div class="bg-white dark:bg-[#1e1e2e] border-b border-gray-200 dark:border-gray-700 py-2 px-2">
      <div class="flex items-center overflow-x-auto scrollbar-thin">
        <div
          v-for="plugin in plugins"
          :key="plugin.id"
          class="plugin-tab-wrapper mx-1"
        >
          <button
            @click="activePlugin = plugin.id"
            draggable="true"
            @dragstart="handleDragStart($event, plugin.id)"
            @dragend="handleDragEnd($event, plugin.id)"
            :class="[
              'btn',
              activePlugin === plugin.id
                ? 'btn-primary'
                : 'btn-secondary',
            ]"
          >
            <i :class="`bx ${plugin.icon} mr-1.5 ${plugin.color}`"></i>
            <span>{{ plugin.name }}</span>
            <i
              v-if="poppedOutPlugins.has(plugin.id)"
              class="bx bx-link-external ml-1 text-xs opacity-60"
              :title="t('modules.plugins.popped_out', '已弹出到独立窗口')"
            ></i>
          </button>
          <button
            class="pop-out-btn"
            :title="t('modules.plugins.pop_out', '弹出到独立窗口')"
            @click.stop="popOutPlugin(plugin.id)"
          >
            <i class="bx bx-window-open"></i>
          </button>
        </div>

        <!-- 弹出全部插件按钮 -->
        <div class="ml-auto pl-2">
          <button
            class="pop-out-all-btn"
            :title="t('modules.plugins.pop_out_all', '弹出全部插件到独立窗口')"
            @click="popOutAllPlugins"
          >
            <i class="bx bx-windows"></i>
            <i
              v-if="allPluginsPopped"
              class="bx bx-link-external ml-0.5 text-xs opacity-60"
            ></i>
          </button>
        </div>
      </div>
    </div>

    <!-- 插件内容区域 -->
    <div class="flex-1 overflow-auto p-1 bg-white dark:bg-[#1e1e2e]">
      <keep-alive>
        <JwtAnalyzer v-if="activePlugin === 'jwt'" />
        <FuzzScanner v-else-if="activePlugin === 'fuzz'" />
        <AuthorizationChecker v-else-if="activePlugin === 'auth'" />
        <APIGenerate v-else-if="activePlugin === 'apigen'" />
        <OastModule v-else-if="activePlugin === 'oast'" />
        <DevModule v-else :moduleId="activePlugin" />
      </keep-alive>
    </div>
  </div>
</template>

<style scoped>
.plugin-tab-wrapper {
  display: inline-flex;
  align-items: center;
  gap: 2px;
}

.pop-out-btn {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 28px;
  height: 28px;
  border-radius: 6px;
  border: none;
  background: transparent;
  color: var(--color-text-secondary, #6b7280);
  cursor: pointer;
  transition: all 0.15s ease;
  font-size: 14px;
}

.pop-out-btn:hover {
  background: var(--glass-bg-secondary, rgba(0, 0, 0, 0.05));
  color: var(--color-text-primary, #374151);
}

.dark .pop-out-btn:hover {
  background: rgba(255, 255, 255, 0.1);
  color: var(--color-text-primary, #e5e7eb);
}

.pop-out-all-btn {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  height: 32px;
  padding: 0 10px;
  border-radius: 6px;
  border: 1px dashed var(--color-text-secondary, #9ca3af);
  background: transparent;
  color: var(--color-text-secondary, #6b7280);
  cursor: pointer;
  transition: all 0.15s ease;
  font-size: 16px;
  white-space: nowrap;
}

.pop-out-all-btn:hover {
  background: var(--glass-bg-secondary, rgba(0, 0, 0, 0.05));
  color: var(--color-text-primary, #374151);
  border-color: var(--color-text-primary, #374151);
}

.dark .pop-out-all-btn:hover {
  background: rgba(255, 255, 255, 0.1);
  color: var(--color-text-primary, #e5e7eb);
  border-color: var(--color-text-primary, #e5e7eb);
}

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
