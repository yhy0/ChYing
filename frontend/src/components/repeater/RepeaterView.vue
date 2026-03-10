<script setup lang="ts">
import { ref, computed, onMounted, onBeforeUnmount, nextTick, watch } from 'vue';
import { useI18n } from 'vue-i18n';
import RepeaterTabs from './RepeaterTabs.vue';
import RepeaterTabPanel from './RepeaterTabPanel.vue';
import RepeaterModal from './RepeaterModal.vue';
import { useModulesStore } from '../../store';
import type { RepeaterTab, RepeaterGroup } from '../../types';
import { generateUUID } from '../../utils';
import { SaveRepeaterState, LoadRepeaterState, DeleteRepeaterTabData } from "../../../bindings/github.com/yhy0/ChYing/app.js";

const { t } = useI18n();

// 添加历史记录接口
interface RequestHistory {
  id: number;
  sequenceId: number; // 自增序列ID
  timestamp: number;
  method: string; 
  url: string;
  request: string;
  response: string | null;
  statusCode?: number;
  statusText?: string;
}

// 扩展RepeaterTab接口以包含history字段
interface RepeaterTabWithHistory extends RepeaterTab {
  history?: RequestHistory[];
  method?: string;
  url?: string;
}

// 使用store
const store = useModulesStore();

// Default tab colors
const tabColors = [
  { id: 'default', value: '#4f46e5', label: 'Default (Purple)' },
  { id: 'red', value: '#ef4444', label: 'Red' },
  { id: 'green', value: '#10b981', label: 'Green' },
  { id: 'blue', value: '#3b82f6', label: 'Blue' },
  { id: 'yellow', value: '#f59e0b', label: 'Yellow' },
  { id: 'orange', value: '#f97316', label: 'Orange' },
  { id: 'teal', value: '#14b8a6', label: 'Teal' },
];

// 从store获取分组数据，不再使用本地变量
const groups = computed(() => store.repeaterGroups);

// 分组模态框状态
const showGroupModal = ref(false);

// Generate a unique ID for a new tab
const generateTabId = () => generateUUID();

// 获取tabs的引用
const tabs = computed(() => store.repeaterTabs);

// Current active tab
const activeTab = computed(() => {
  const active = tabs.value.find(tab => tab.isActive);
  return active || null;
});

// 默认的 HTTP 请求模板
const defaultRequest = `GET / HTTP/1.1
Host: 127.0.0.1
User-Agent: Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36
Accept: text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,*/*;q=0.8
Accept-Language: en-US,en;q=0.5
Accept-Encoding: gzip, deflate
Connection: close

`;

// Create a new empty tab
const createTab = () => {
  // Deactivate all existing tabs
  tabs.value.forEach(tab => {
    tab.isActive = false;
  });

  // Create and add new tab
  const newTabId = generateTabId();
  
  // 获取当前计数器值并增加
  const currentCounter = store.repeaterTabCounter;
  store.repeaterTabCounter++;
  
  // 创建新标签，使用默认请求模板
  const newTab: RepeaterTabWithHistory = {
    id: newTabId,
    name: `# ${currentCounter}`,
    color: tabColors[0].value,
    groupId: null,
    request: defaultRequest,
    response: null,
    isActive: true,
    isRunning: false,
    modified: false,
    history: [], // 初始化空的历史记录数组
    serverDurationMs: 0,
    method: 'GET',
    url: 'http://127.0.0.1/',
  };
  
  // 添加新标签并确保在下一个周期更新视图
  store.repeaterTabs.push(newTab);
  
  // 确保新标签被正确激活
  nextTick(() => {
    setActiveTab(newTabId);
  });
};

// Set the active tab
const setActiveTab = (tabId: string) => {
  tabs.value.forEach(tab => {
    tab.isActive = tab.id === tabId;
  });
};

// Update tab details
const updateTab = (tabId: string, data: Partial<RepeaterTab>) => {
  const index = tabs.value.findIndex(tab => tab.id === tabId);
  if (index !== -1) {
    tabs.value[index] = { ...tabs.value[index], ...data };
  }
};

// Rename a tab
const renameTab = (tabId: string, newName: string) => {
  updateTab(tabId, { name: newName });
};

// 重新排序标签
const reorderTabs = (newTabs: RepeaterTab[]) => {
  // 直接使用新的排序替换现有标签
  store.repeaterTabs.splice(0, store.repeaterTabs.length, ...newTabs);
};

// Create a new group
const createGroup = (name: string, color: string = '#4f46e5') => {
  store.createRepeaterGroup(name, color);
};

// Close a tab
const closeTab = (tabId: string) => {
  const index = tabs.value.findIndex(tab => tab.id === tabId);
  if (index !== -1) {
    tabs.value.splice(index, 1);

    // If we closed the active tab, activate another one if available
    if (tabs.value.length > 0 && activeTab.value === null) {
      const newActiveIndex = Math.min(index, tabs.value.length - 1);
      tabs.value[newActiveIndex].isActive = true;
    }

    // 从数据库删除该 tab 及其 history
    DeleteRepeaterTabData(tabId).catch(() => {});
  }
};

// Update request data in a tab
const updateRequestData = (tabId: string, data: string) => {
  const tab = tabs.value.find(tab => tab.id === tabId);
  if (tab) {
    // 将 RequestData 对象转换为字符串
    tab.request = data;
  }
};

// Update response data in a tab
const updateResponseData = (tabId: string, data: string | null) => {
  const tab = tabs.value.find(tab => tab.id === tabId);
  if (tab) {
    if (data) {
      // 将 ResponseData 对象转换为字符串
      tab.response = data;
    } else {
      tab.response = null;
    }
  }
};

// Send request to Intruder
const sendToIntruder = () => {
  if (!activeTab.value) return;
  
  store.sendRepeaterToIntruder(activeTab.value);
};

// Clone current tab
const cloneTab = () => {
  if (!activeTab.value) return;

  // 先保存原始 tab 数据，再修改 isActive（避免 computed 失效导致 null）
  const originalTab = activeTab.value;
  const originalColor = originalTab.color;
  const originalGroupId = originalTab.groupId;
  const originalRequest = originalTab.request;
  const originalMethod = originalTab.method || 'GET';
  const originalUrl = originalTab.url;

  // Deactivate all existing tabs
  tabs.value.forEach(tab => {
    tab.isActive = false;
  });

  const newTabId = generateTabId();

  // 获取当前计数器值并增加
  const currentCounter = store.repeaterTabCounter;
  store.repeaterTabCounter++;

  const newTab: RepeaterTabWithHistory = {
    id: newTabId,
    name: `# ${currentCounter}`,
    color: originalColor,
    groupId: originalGroupId,
    request: originalRequest,
    response: null, // Reset response
    isActive: true,
    isRunning: false,
    modified: false,
    history: [], // 新的空历史记录，不从原标签复制
    serverDurationMs: 0,
    method: originalMethod,
    url: originalUrl,
  };

  tabs.value.push(newTab);
};

// 更新历史记录
const updateHistory = (tabId: string, history: RequestHistory[]) => {
  const tab = tabs.value.find(tab => tab.id === tabId);
  if (tab) {
    // 把历史记录添加到标签中
    (tab as RepeaterTabWithHistory).history = history;
  }
};

// 持久化相关：防抖保存 tabs 和 groups 到数据库
let saveTimer: ReturnType<typeof setTimeout> | null = null;
const debouncedSaveState = () => {
  if (saveTimer) clearTimeout(saveTimer);
  saveTimer = setTimeout(() => {
    const tabsData = tabs.value.map((tab, index) => ({
      id: tab.id,
      name: tab.name,
      color: tab.color,
      group_id: tab.groupId || '',
      request: tab.request || '',
      response: tab.response || '',
      method: tab.method || 'GET',
      url: tab.url || '',
      sort_order: index,
      is_active: tab.isActive,
      server_duration_ms: tab.serverDurationMs || 0,
    }));
    const groupsData = groups.value.map((group, index) => ({
      id: group.id,
      name: group.name,
      color: group.color,
      sort_order: index,
    }));
    SaveRepeaterState(JSON.stringify(tabsData), JSON.stringify(groupsData)).catch(() => {});
  }, 2000);
};

// 标记是否已从数据库加载完成，避免加载过程中触发 watch 保存
const loaded = ref(false);

// 更新服务器响应时间
const handleServerDurationUpdate = (duration: number, tabId: string) => {
  const tab = store.repeaterTabs.find(t => t.id === tabId);
  if (tab) {
    tab.serverDurationMs = duration;
  }
};

// 更新 tab 的 URL
const updateTabUrl = (tabId: string, url: string) => {
  const tab = tabs.value.find(tab => tab.id === tabId);
  if (tab) {
    (tab as RepeaterTabWithHistory).url = url;
  }
};

// Handle keyboard shortcuts
const handleKeyDown = (event: KeyboardEvent) => {
  // Ctrl+T: New Tab
  if (event.ctrlKey && event.key === 't') {
    event.preventDefault();
    createTab();
  }
  
  // Ctrl+I: Send to Intruder
  else if (event.ctrlKey && event.key === 'i' && activeTab.value) {
    event.preventDefault();
    sendToIntruder();
  }
  
  // Ctrl+D: Clone Tab
  else if (event.ctrlKey && event.key === 'd' && activeTab.value) {
    event.preventDefault();
    cloneTab();
  }
  
  // Ctrl+W: Close Tab
  else if (event.ctrlKey && event.key === 'w' && activeTab.value) {
    event.preventDefault();
    closeTab(activeTab.value.id);
  }
};

// Initialize a default tab if none exists
onMounted(async () => {
  // 从数据库加载持久化的 tabs 和 groups
  try {
    const result = await LoadRepeaterState();
    if (result && !result.error && result.data) {
      const data = result.data as { tabs: any[]; groups: any[] };
      if (data.groups && data.groups.length > 0) {
        store.repeaterGroups.splice(0, store.repeaterGroups.length, ...data.groups.map((g: any) => ({
          id: g.id,
          name: g.name,
          color: g.color,
        })));
      }
      if (data.tabs && data.tabs.length > 0) {
        // 找到最大的 tab 计数器（从 name 中解析 # 后面的数字）
        let maxCounter = 0;
        const restoredTabs: RepeaterTab[] = data.tabs.map((t: any) => {
          const match = t.name.match(/^#\s*(\d+)/);
          if (match) {
            const num = parseInt(match[1], 10);
            if (num > maxCounter) maxCounter = num;
          }
          return {
            id: t.id,
            name: t.name,
            color: t.color,
            groupId: t.group_id || null,
            request: t.request || defaultRequest,
            response: t.response || null,
            isActive: t.is_active || false,
            isRunning: false,
            modified: false,
            history: [],
            serverDurationMs: t.server_duration_ms || 0,
            method: t.method || 'GET',
            url: t.url || '',
          };
        });

        // 确保至少有一个 tab 是激活的
        if (!restoredTabs.some(tab => tab.isActive) && restoredTabs.length > 0) {
          restoredTabs[0].isActive = true;
        }

        store.repeaterTabs.splice(0, store.repeaterTabs.length, ...restoredTabs);
        if (maxCounter >= store.repeaterTabCounter) {
          store.repeaterTabCounter = maxCounter + 1;
        }
      }
    }
  } catch {
    // 加载失败时忽略，使用默认状态
  }

  if (tabs.value.length === 0) {
    createTab();
  }

  loaded.value = true;
  window.addEventListener('keydown', handleKeyDown);
});

// 监听 tabs 和 groups 变化，自动保存到数据库
watch(() => [...tabs.value.map(t => ({ id: t.id, name: t.name, color: t.color, groupId: t.groupId, request: t.request, response: t.response, isActive: t.isActive, method: t.method, url: t.url, serverDurationMs: t.serverDurationMs }))], () => {
  if (loaded.value) debouncedSaveState();
}, { deep: true });

watch(() => [...groups.value.map(g => ({ id: g.id, name: g.name, color: g.color }))], () => {
  if (loaded.value) debouncedSaveState();
}, { deep: true });

onBeforeUnmount(() => {
  window.removeEventListener('keydown', handleKeyDown);
});

// 处理标签颜色变更
const handleColorChange = (tabId: string, color: string) => {
  const index = tabs.value.findIndex(tab => tab.id === tabId);
  if (index !== -1) {
    tabs.value[index].color = color;
  }
};

// 处理标签分组变更
const handleGroupChange = (tabId: string, groupId: string | null) => {
  store.changeTabGroup(tabId, groupId);
};

// New function to handle tab selection
const handleTabSelect = (tabId: string) => {
  setActiveTab(tabId);
};

// 处理创建分组
const handleCreateGroup = (name: string, color: string) => {
  createGroup(name, color);
  closeCreateGroupModal();
};

// New function to handle reordering tabs
const handleReorderTabs = (newTabs: RepeaterTab[]) => {
  reorderTabs(newTabs);
};

// 处理分组重新排序
const handleReorderGroups = (newGroups: RepeaterGroup[]) => {
  store.repeaterGroups.splice(0, store.repeaterGroups.length, ...newGroups);
};

// 打开创建分组模态框
const openCreateGroupModal = () => {
  showGroupModal.value = true;
};

// 关闭创建分组模态框
const closeCreateGroupModal = () => {
  showGroupModal.value = false;
};
</script>

<template>
  <div class="h-full flex flex-col">
    <!-- Control Bar -->
    <div class="repeater-control-bar">
      <div class="flex items-center space-x-4">
        <button
          class="btn btn-primary"
          @click="createTab"
          :title="t('modules.repeater.new_tab_tooltip')"
        >
          <i class="bx bx-plus mr-1"></i> {{ t('modules.repeater.new_tab') }}
        </button>
        
        <button
          class="btn btn-secondary"
          @click="cloneTab"
          :disabled="!activeTab"
          :class="{ 'opacity-50 cursor-not-allowed': !activeTab }"
          :title="t('modules.repeater.clone_tooltip')"
        >
          <i class="bx bx-duplicate mr-1"></i> {{ t('modules.repeater.clone') }}
        </button>
        
        <button
          class="btn btn-secondary"
          @click="sendToIntruder"
          :disabled="!activeTab"
          :class="{ 'opacity-50 cursor-not-allowed': !activeTab }"
          :title="t('modules.repeater.send_to_intruder_tooltip')"
        >
          <i class="bx bx-target-lock mr-1"></i> {{ t('modules.repeater.send_to_intruder') }}
        </button>

        <button
          class="btn btn-secondary"
          @click="openCreateGroupModal"
          :title="t('modules.repeater.create_new_group')"
        >
          <i class="bx bx-folder-plus mr-1"></i> {{ t('modules.repeater.new_group') }}
        </button>
      </div>
    </div>
    
    <!-- Tabs Bar -->
    <div class="flex-none overflow-visible">
      <RepeaterTabs
        :tabs="tabs"
        :groups="groups"
        @select-tab="handleTabSelect"
        @close-tab="closeTab"
        @rename-tab="renameTab"
        @change-tab-color="handleColorChange"
        @change-tab-group="handleGroupChange"
        @create-group="openCreateGroupModal"
        @reorder-tabs="handleReorderTabs"
        @reorder-groups="handleReorderGroups"
      />
    </div>
    
    <!-- Tab Content -->
    <div v-if="activeTab" class="flex-1 overflow-hidden">
      <RepeaterTabPanel 
        :key="activeTab.id"
        :tab="activeTab"
        @update-request="updateRequestData(activeTab.id, $event)"
        @update-response="updateResponseData(activeTab.id, $event)"
        @update-history="updateHistory(activeTab.id, $event)"
        @update-server-duration="(duration) => activeTab && handleServerDurationUpdate(duration, activeTab.id)"
        @update-url="updateTabUrl(activeTab.id, $event)"
      />
    </div>
    
    <!-- Empty State -->
    <div 
      v-else 
      class="empty-state"
    >
      <i class="bx bx-repeat empty-state-icon"></i>
      <h3 class="empty-state-text">{{ t('modules.repeater.no_request_open') }}</h3>
      <p class="empty-state-text">{{ t('modules.repeater.create_tab_hint') }}</p>
      <button
        class="btn btn-primary"
        @click="createTab"
      >
        <i class="bx bx-plus mr-1"></i> {{ t('modules.repeater.new_tab') }}
      </button>
    </div>

    <!-- 分组创建模态框 -->
    <RepeaterModal
      :show="showGroupModal"
      :title="t('modules.repeater.create_new_group')"
      :submit-text="t('common.actions.create')"
      @close="closeCreateGroupModal"
      @submit="handleCreateGroup"
    />
  </div>
</template>