<script setup lang="ts">
import { ref, computed, onMounted, onUnmounted, watch } from 'vue';
import { useI18n } from 'vue-i18n';
import { useRoute, useRouter } from 'vue-router';
import { getCurrentTheme } from '../theme';
import { i18n } from '../i18n';
import { useClaudeStore } from '../store/claude';
import ClaudeChat from '../components/claude/ClaudeChat.vue';
import ClaudeContextSelector from '../components/claude/ClaudeContextSelector.vue';
import ClaudeSettings from '../components/claude/ClaudeSettings.vue';
import ClaudeToolConfirm from '../components/claude/ClaudeToolConfirm.vue';
import ClaudeSessionList from '../components/claude/ClaudeSessionList.vue';
import type { AgentContext } from '../types/claude';
import { Events } from "@wailsio/runtime";

const { t } = useI18n();
const claudeStore = useClaudeStore();
const route = useRoute();
const router = useRouter();

// 主题响应状态
const currentTheme = ref(getCurrentTheme());
const themeChangeKey = ref(0);

// 组件状态
const showContextSelector = ref(false);
const showSettings = ref(false);
const showSessionList = ref(true); // 默认显示会话列表
const projectId = ref('default');
const sessionListRef = ref<InstanceType<typeof ClaudeSessionList> | null>(null);
const claudeChatRef = ref<InstanceType<typeof ClaudeChat> | null>(null);

// 预填充消息（从路由参数获取）
const prefillMessage = ref('');

// 选中的流量 ID 列表
const selectedTrafficIds = ref<number[]>([]);

// 上下文
const agentContext = ref<AgentContext>({
  projectId: 'default',
  autoCollect: true
});

// 计算属性
const isInitialized = computed(() => claudeStore.initialized);
const isLoading = computed(() => claudeStore.loading);
const hasError = computed(() => !!claudeStore.error);
const errorMessage = computed(() => claudeStore.error);
const hasPendingTools = computed(() => claudeStore.hasPendingTools);
const pendingToolUses = computed(() => claudeStore.pendingToolUses);

// 初始化
const initializeClaude = async (): Promise<boolean> => {
  const success = await claudeStore.initializeService();
  if (success) {
    // 创建新会话
    await claudeStore.createSession(projectId.value, agentContext.value);
  }
  return success;
};

// 创建新会话
const createNewSession = async () => {
  await claudeStore.createSession(projectId.value, agentContext.value);
  // 刷新会话列表
  sessionListRef.value?.refresh();
};

// 切换会话列表显示
const toggleSessionList = () => {
  showSessionList.value = !showSessionList.value;
};

// 清除会话
const clearSession = async () => {
  await claudeStore.clearCurrentSession();
};

// 切换上下文选择器
const toggleContextSelector = () => {
  showContextSelector.value = !showContextSelector.value;
};

// 切换设置面板
const toggleSettings = () => {
  showSettings.value = !showSettings.value;
};

// 更新上下文
const handleContextUpdate = (context: AgentContext) => {
  agentContext.value = context;
  claudeStore.updateContext(context);
};

// 确认工具执行
const handleToolConfirm = async (toolUseId: string, confirmed: boolean) => {
  await claudeStore.confirmToolExecution(toolUseId, confirmed);
};

// 主题变化处理
const handleThemeChange = () => {
  currentTheme.value = getCurrentTheme();
  themeChangeKey.value++;
};

// localStorage 变化处理
const handleStorageChange = (e: StorageEvent) => {
  if (e.key === 'app-theme' && e.newValue) {
    const newTheme = e.newValue as 'light' | 'dark' | 'system';
    const isDark = newTheme === 'dark' || (newTheme === 'system' && window.matchMedia('(prefers-color-scheme: dark)').matches);

    if (isDark) {
      document.documentElement.classList.add('dark');
    } else {
      document.documentElement.classList.remove('dark');
    }

    handleThemeChange();
  }

  if (e.key === 'language' && e.newValue) {
    const newLang = e.newValue as 'en' | 'zh';
    i18n.global.locale.value = newLang;
    document.querySelector('html')?.setAttribute('lang', newLang);
  }
};

// 组件挂载
onMounted(async () => {
  // 设置事件监听
  claudeStore.setupEventListeners();

  // 监听后端发送的流量 ID 事件
  Events.On("claude:traffic-ids", handleTrafficIdsEvent);

  // 初始化 Claude 服务
  const success = await initializeClaude();

  // 只有初始化成功后才处理路由参数
  if (success) {
    // 检查路由参数，处理预填充消息
    handleRouteQuery();
  }

  // 监听 localStorage 变化
  window.addEventListener('storage', handleStorageChange);
});

// 处理路由查询参数
const handleRouteQuery = () => {
  const { prefill, trafficIds } = route.query;
  
  // 处理预填充消息
  if (prefill && typeof prefill === 'string') {
    prefillMessage.value = prefill;
  }
  
  // 处理流量 ID 列表
  if (trafficIds && typeof trafficIds === 'string') {
    const ids = trafficIds.split(',').map(id => parseInt(id.trim(), 10)).filter(id => !isNaN(id));
    if (ids.length > 0) {
      handleTrafficIds(ids);
    }
  }
  
  // 清除 URL 中的查询参数，避免刷新时重复填充
  if (prefill || trafficIds) {
    router.replace({ path: route.path, query: {} });
  }
};

// 处理流量 ID 列表
const handleTrafficIds = (ids: number[]) => {
  selectedTrafficIds.value = ids;
  // 构建预填充消息
  const message = t('claude.prefill.analyzeTraffic', { 
    count: ids.length, 
    ids: ids.join(', ') 
  }, `Please analyze the following ${ids.length} HTTP traffic records (IDs: ${ids.join(', ')}). Look for potential security vulnerabilities, sensitive information leaks, and suspicious patterns.`);
  prefillMessage.value = message;
};

// 监听路由变化
watch(() => route.query, (newQuery) => {
  if (newQuery.prefill || newQuery.trafficIds) {
    handleRouteQuery();
  }
}, { immediate: false });

// 监听后端发送的流量 ID 事件（当窗口已存在时）
// Wails v3 事件回调接收 WailsEvent 对象，实际数据在 event.data 中
const handleTrafficIdsEvent = (event: any) => {
  const ids = event?.data;
  if (ids && Array.isArray(ids) && ids.length > 0) {
    handleTrafficIds(ids);
  }
};

// 组件卸载
onUnmounted(() => {
  claudeStore.cleanupEventListeners();
  Events.Off("claude:traffic-ids");
  window.removeEventListener('storage', handleStorageChange);
});
</script>

<template>
  <div class="claude-agent-view" :key="`claude-${themeChangeKey}`">
    <!-- 头部工具栏 -->
    <div class="section-header">
      <div class="header-left">
        <h3>
          <i class="bx bx-bot"></i>
          {{ t('claude.title', 'AI Security Assistant') }}
        </h3>
        <div class="header-status">
          <span v-if="isInitialized" class="status-badge status-connected">
            <i class="bx bx-check-circle"></i>
            {{ t('claude.status.connected', 'Connected') }}
          </span>
          <span v-else class="status-badge status-disconnected">
            <i class="bx bx-x-circle"></i>
            {{ t('claude.status.disconnected', 'Disconnected') }}
          </span>
        </div>
      </div>

      <div class="header-actions">
        <button
          @click="toggleSessionList"
          class="btn btn-sm"
          :class="showSessionList ? 'btn-primary' : 'btn-secondary'"
          :title="t('claude.sessions.toggle', 'Toggle Sessions')"
        >
          <i class="bx bx-history"></i>
        </button>
        <button
          @click="toggleContextSelector"
          class="btn btn-sm"
          :class="showContextSelector ? 'btn-primary' : 'btn-secondary'"
          :title="t('claude.context.title', 'Context')"
        >
          <i class="bx bx-data"></i>
          {{ t('claude.context.title', 'Context') }}
        </button>
        <button
          @click="createNewSession"
          class="btn btn-sm btn-success"
          :disabled="!isInitialized || isLoading"
          :title="t('claude.actions.newSession', 'New Session')"
        >
          <i class="bx bx-plus"></i>
          {{ t('claude.actions.newSession', 'New') }}
        </button>
        <button
          @click="clearSession"
          class="btn btn-sm btn-warning"
          :disabled="!isInitialized || isLoading"
          :title="t('claude.actions.clearSession', 'Clear')"
        >
          <i class="bx bx-trash"></i>
          {{ t('claude.actions.clearSession', 'Clear') }}
        </button>
        <button
          @click="toggleSettings"
          class="btn btn-sm btn-secondary"
          :title="t('common.settings', 'Settings')"
        >
          <i class="bx bx-cog"></i>
        </button>
      </div>
    </div>

    <!-- 错误提示 -->
    <div 
      v-if="hasError" 
      class="error-banner"
      role="alert"
      aria-live="polite"
    >
      <i class="bx bx-error-circle" aria-hidden="true"></i>
      <span>{{ errorMessage }}</span>
      <button @click="initializeClaude" class="btn btn-sm btn-primary">
        {{ t('common.retry', 'Retry') }}
      </button>
    </div>


    <!-- 主内容区域 -->
    <div class="main-content">
      <!-- 会话列表侧边栏（左侧） -->
      <div
        class="session-sidebar"
        v-if="showSessionList"
      >
        <ClaudeSessionList
          ref="sessionListRef"
          :currentProjectId="projectId"
          @newSession="createNewSession"
          :key="`session-list-${themeChangeKey}`"
        />
      </div>

      <!-- 聊天区域 -->
      <div class="chat-content">
        <ClaudeChat
          ref="claudeChatRef"
          :initialMessage="prefillMessage"
          :selectedTrafficIds="selectedTrafficIds"
          @clearTrafficIds="selectedTrafficIds = []"
          :key="`chat-${themeChangeKey}`"
        />
      </div>

      <!-- 上下文选择器侧边栏（右侧） -->
      <div
        class="context-sidebar scrollbar-thin"
        v-if="showContextSelector"
      >
        <ClaudeContextSelector
          :context="agentContext"
          @update:context="handleContextUpdate"
          :key="`context-${themeChangeKey}`"
        />
      </div>
    </div>

    <!-- 工具确认模态框 -->
    <ClaudeToolConfirm
      v-if="hasPendingTools"
      :toolUses="pendingToolUses"
      @confirm="handleToolConfirm"
      :key="`tool-confirm-${themeChangeKey}`"
    />

    <!-- 设置面板 -->
    <ClaudeSettings
      v-if="showSettings"
      @close="showSettings = false"
      :key="`settings-${themeChangeKey}`"
    />
  </div>
</template>

<style scoped>
.claude-agent-view {
  display: flex;
  flex-direction: column;
  height: 100vh;
  background: var(--color-bg-primary);
  color: var(--color-text-primary);
}

.header-left {
  display: flex;
  align-items: center;
  gap: 16px;
}

.header-left h3 {
  display: flex;
  align-items: center;
  gap: 8px;
  margin: 0;
}

.header-left h3 i {
  font-size: 1.25rem;
  color: var(--color-primary);
}

.header-status {
  display: flex;
  align-items: center;
}

.status-badge {
  display: flex;
  align-items: center;
  gap: 4px;
  padding: 4px 8px;
  border-radius: 4px;
  font-size: 0.75rem;
  font-weight: 500;
}

.status-connected {
  background: var(--color-success-bg, rgba(34, 197, 94, 0.1));
  color: var(--color-success, #22c55e);
}

.status-disconnected {
  background: var(--color-danger-bg, rgba(239, 68, 68, 0.1));
  color: var(--color-danger, #ef4444);
}

.error-banner {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 12px 16px;
  background: var(--color-danger-bg, rgba(239, 68, 68, 0.1));
  border-bottom: 1px solid var(--color-danger, #ef4444);
  color: var(--color-danger, #ef4444);
}

.error-banner i {
  font-size: 1.25rem;
}

.error-banner span {
  flex: 1;
}

.main-content {
  display: flex;
  flex: 1;
  overflow: hidden;
}

/* 会话列表侧边栏（左侧） */
.session-sidebar {
  width: 280px;
  background: var(--color-bg-secondary);
  border-right: 1px solid var(--color-border);
  flex-shrink: 0;
  display: flex;
  flex-direction: column;
}

/* 上下文选择器侧边栏（右侧） */
.context-sidebar {
  width: 320px;
  background: var(--color-bg-secondary);
  border-left: 1px solid var(--color-border);
  overflow-y: auto;
  flex-shrink: 0;
}

.chat-content {
  flex: 1;
  overflow: hidden;
  display: flex;
  flex-direction: column;
  min-width: 0; /* 防止 flex 子元素溢出 */
}
</style>
