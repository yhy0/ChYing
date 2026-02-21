<script setup lang="ts">
import { ref, computed, onMounted, onUnmounted } from 'vue';
import { useI18n } from 'vue-i18n';
import { useOastStore, type OASTInteraction } from '../../../store/oast';
import OastProviderManager from './OastProviderManager.vue';
import HttpRequestViewer from '../../common/codemirror/HttpRequestViewer.vue';
import HttpResponseViewer from '../../common/codemirror/HttpResponseViewer.vue';

const { t } = useI18n();
const store = useOastStore();

// 初始化
onMounted(async () => {
  store.initEventListener();
  await store.loadProviders();
  await store.loadSettings();
});

// 选中的 provider ID
const selectedProviderId = ref('');

// 复制提示
const copyTooltip = ref(false);

// 获取 Payload 的加载状态
const isGettingPayload = ref(false);

// 当前 tab 的 payload URL
const currentPayloadUrl = computed(() => {
  if (store.activeTab) {
    return store.activeTab.payloadUrl;
  }
  return '';
});

// 当前 tab 的交互列表
const currentInteractions = computed(() => {
  if (store.activeTab) {
    return store.activeTab.interactions;
  }
  return [];
});

// 已启用的 providers
const enabledProviders = computed(() => {
  return store.providers.filter(p => p.enabled);
});

// 获取 Payload
const getPayload = async () => {
  if (!selectedProviderId.value || isGettingPayload.value) return;

  const provider = store.providers.find(p => p.id === selectedProviderId.value);
  if (!provider) return;

  isGettingPayload.value = true;
  try {
    // 注册
    const payloadUrl = await store.register(selectedProviderId.value);
    if (!payloadUrl) return;

    // 创建新 tab
    const tab = store.addTab(selectedProviderId.value, provider.name, payloadUrl);

    // 启动轮询
    await store.startPolling(selectedProviderId.value);
  } finally {
    isGettingPayload.value = false;
  }
};

// 手动 Poll
const manualPoll = async () => {
  if (!store.activeTab) return;
  await store.pollOnce(store.activeTab.providerId);
};

// 切换轮询
const togglePolling = async () => {
  if (!store.activeTab) return;
  const provider = store.providers.find(p => p.id === store.activeTab!.providerId);
  if (!provider) return;

  if (provider.polling) {
    await store.stopPolling(provider.id);
  } else {
    await store.startPolling(provider.id);
  }
};

// 清除交互
const clearInteractions = () => {
  store.clearInteractions(store.activeTab?.id);
};

// 复制 Payload URL
const copyPayloadUrl = async () => {
  if (!currentPayloadUrl.value) return;
  try {
    await navigator.clipboard.writeText(currentPayloadUrl.value);
    copyTooltip.value = true;
    setTimeout(() => { copyTooltip.value = false; }, 1500);
  } catch (e) {
    console.error('Copy failed:', e);
  }
};

// 关闭 tab
const closeTab = async (tabId: string) => {
  const tab = store.tabs.find(t => t.id === tabId);
  if (tab) {
    await store.stopPolling(tab.providerId);
  }
  store.removeTab(tabId);
};

// 选中交互
const selectInteraction = (interaction: OASTInteraction) => {
  store.selectedInteraction = interaction;
};

// 格式化时间
const formatTime = (ts: string) => {
  try {
    const d = new Date(ts);
    return d.toLocaleTimeString();
  } catch {
    return ts;
  }
};

// 获取当前活跃 tab 的 polling 状态
const isCurrentPolling = computed(() => {
  if (!store.activeTab) return false;
  const provider = store.providers.find(p => p.id === store.activeTab!.providerId);
  return provider?.polling || false;
});
</script>

<template>
  <div class="h-full flex flex-col oast-module">
    <!-- Tab 栏 -->
    <div class="tab-bar">
      <div class="tab-list">
        <button
          v-for="tab in store.tabs"
          :key="tab.id"
          :class="['tab-item', { active: tab.id === store.activeTabId }]"
          @click="store.setActiveTab(tab.id)"
        >
          <span class="tab-name">{{ tab.name }}</span>
          <span class="tab-count">({{ tab.interactions.length }})</span>
          <i class="bx bx-x tab-close" @click.stop="closeTab(tab.id)"></i>
        </button>
      </div>
    </div>

    <!-- 操作栏 -->
    <div class="action-bar">
      <div class="action-left">
        <select v-model="selectedProviderId" class="provider-select">
          <option value="" disabled>{{ t('modules.plugins.oast.select_provider', 'Select Provider') }}</option>
          <option v-for="p in enabledProviders" :key="p.id" :value="p.id">
            {{ p.name }} ({{ p.type }})
          </option>
        </select>
        <button class="btn btn-primary btn-sm" @click="getPayload" :disabled="!selectedProviderId || isGettingPayload">
          <i :class="isGettingPayload ? 'bx bx-loader-alt bx-spin mr-1' : 'bx bx-link mr-1'"></i>
          {{ isGettingPayload ? t('modules.plugins.oast.getting_payload', 'Connecting...') : t('modules.plugins.oast.get_payload', 'Get Payload') }}
        </button>
      </div>

      <div class="action-center" v-if="currentPayloadUrl">
        <div class="payload-display">
          <code class="payload-url">{{ currentPayloadUrl }}</code>
          <button class="btn-icon" @click="copyPayloadUrl" :title="t('modules.plugins.oast.copy', 'Copy')">
            <i :class="copyTooltip ? 'bx bx-check text-green-500' : 'bx bx-copy'"></i>
          </button>
        </div>
      </div>

      <div class="action-right">
        <button class="btn btn-secondary btn-sm" @click="manualPoll" :disabled="!store.activeTab" :title="t('modules.plugins.oast.poll_once', 'Poll Once')">
          <i class="bx bx-refresh"></i>
        </button>
        <button
          :class="['btn', 'btn-sm', isCurrentPolling ? 'btn-warning' : 'btn-secondary']"
          @click="togglePolling"
          :disabled="!store.activeTab"
          :title="isCurrentPolling ? t('modules.plugins.oast.stop_polling', 'Stop Polling') : t('modules.plugins.oast.start_polling', 'Start Polling')"
        >
          <i :class="isCurrentPolling ? 'bx bx-pause' : 'bx bx-play'"></i>
        </button>
        <button class="btn btn-secondary btn-sm" @click="clearInteractions" :disabled="!store.activeTab">
          <i class="bx bx-trash"></i>
        </button>
        <button class="btn btn-secondary btn-sm" @click="store.showProviderManager = true">
          <i class="bx bx-cog"></i>
        </button>
      </div>
    </div>

    <!-- 主内容区域 -->
    <div class="flex-1 flex flex-col overflow-hidden">
      <!-- 空状态 -->
      <div v-if="store.tabs.length === 0" class="empty-state">
        <i class="bx bx-broadcast text-5xl text-gray-400 mb-3"></i>
        <p class="text-lg text-gray-500">{{ t('modules.plugins.oast.empty_title', 'No Active OAST Sessions') }}</p>
        <p class="text-sm text-gray-400 mt-1">{{ t('modules.plugins.oast.empty_desc', 'Select a provider and click "Get Payload" to start') }}</p>
      </div>

      <!-- 交互列表 -->
      <template v-else>
        <div class="interaction-table-container">
          <table class="interaction-table">
            <thead>
              <tr>
                <th class="w-12">#</th>
                <th class="w-20">{{ t('modules.plugins.oast.protocol', 'Protocol') }}</th>
                <th class="w-20">{{ t('modules.plugins.oast.method', 'Method') }}</th>
                <th>{{ t('modules.plugins.oast.source', 'Source') }}</th>
                <th>{{ t('modules.plugins.oast.destination', 'Destination') }}</th>
                <th class="w-24">{{ t('modules.plugins.oast.provider', 'Provider') }}</th>
                <th class="w-24">{{ t('modules.plugins.oast.time', 'Time') }}</th>
              </tr>
            </thead>
            <tbody>
              <tr
                v-for="(interaction, idx) in currentInteractions"
                :key="interaction.id"
                :class="{ selected: store.selectedInteraction?.id === interaction.id }"
                @click="selectInteraction(interaction)"
              >
                <td>{{ idx + 1 }}</td>
                <td>
                  <span :class="['protocol-badge', `protocol-${interaction.protocol?.toLowerCase()}`]">
                    {{ interaction.protocol || 'HTTP' }}
                  </span>
                </td>
                <td>{{ interaction.method }}</td>
                <td class="truncate max-w-48">{{ interaction.source }}</td>
                <td class="truncate max-w-60">{{ interaction.destination }}</td>
                <td>
                  <span class="provider-badge">{{ interaction.type }}</span>
                </td>
                <td>{{ formatTime(interaction.timestamp) }}</td>
              </tr>
              <tr v-if="currentInteractions.length === 0">
                <td colspan="7" class="text-center text-gray-400 py-8">
                  {{ t('modules.plugins.oast.no_interactions', 'Waiting for interactions...') }}
                </td>
              </tr>
            </tbody>
          </table>
        </div>

        <!-- 详情面板 -->
        <div v-if="store.selectedInteraction" class="detail-panel">
          <div class="detail-header">
            <span class="detail-title">
              {{ store.selectedInteraction.protocol }} {{ store.selectedInteraction.method }}
              <span class="text-gray-400 ml-2">{{ store.selectedInteraction.source }}</span>
            </span>
          </div>
          <div class="detail-content">
            <div class="detail-section" v-if="store.selectedInteraction.rawRequest">
              <h4>{{ t('modules.plugins.oast.raw_request', 'Raw Request') }}</h4>
              <HttpRequestViewer :data="store.selectedInteraction.rawRequest" :read-only="true" />
            </div>
            <div class="detail-section" v-if="store.selectedInteraction.rawResponse">
              <h4>{{ t('modules.plugins.oast.raw_response', 'Raw Response') }}</h4>
              <HttpResponseViewer :data="store.selectedInteraction.rawResponse" :read-only="true" />
            </div>
          </div>
        </div>
      </template>
    </div>

    <!-- Provider 管理弹窗 -->
    <OastProviderManager
      v-if="store.showProviderManager"
      @close="store.showProviderManager = false"
    />
  </div>
</template>

<style scoped>
.oast-module {
  background: var(--color-bg-primary);
}

.tab-bar {
  display: flex;
  align-items: center;
  padding: 4px 8px;
  border-bottom: 1px solid var(--glass-border-light, #e5e7eb);
  background: var(--glass-bg-secondary, rgba(255,255,255,0.6));
  min-height: 36px;
}

.tab-list {
  display: flex;
  gap: 4px;
  overflow-x: auto;
}

.tab-item {
  display: inline-flex;
  align-items: center;
  gap: 4px;
  padding: 4px 10px;
  border-radius: 6px;
  border: 1px solid transparent;
  background: transparent;
  color: var(--color-text-secondary, #6b7280);
  cursor: pointer;
  font-size: 12px;
  white-space: nowrap;
  transition: all 0.15s;
}

.tab-item:hover {
  background: var(--glass-bg-secondary, rgba(0,0,0,0.05));
}

.tab-item.active {
  background: var(--color-primary, #3b82f6);
  color: white;
  border-color: var(--color-primary, #3b82f6);
}

.tab-close {
  font-size: 14px;
  opacity: 0.6;
  cursor: pointer;
}

.tab-close:hover {
  opacity: 1;
}

.tab-count {
  font-size: 11px;
  opacity: 0.7;
}

.action-bar {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 6px 8px;
  border-bottom: 1px solid var(--glass-border-light, #e5e7eb);
  background: var(--glass-bg-secondary, rgba(255,255,255,0.4));
  gap: 8px;
}

.action-left,
.action-right {
  display: flex;
  align-items: center;
  gap: 6px;
}

.action-center {
  flex: 1;
  min-width: 0;
}

.provider-select {
  padding: 4px 8px;
  border: 1px solid var(--glass-border-light, #d1d5db);
  border-radius: 6px;
  background: var(--color-bg-primary);
  color: var(--color-text-primary);
  font-size: 12px;
  min-width: 140px;
}

.payload-display {
  display: flex;
  align-items: center;
  gap: 6px;
  padding: 3px 8px;
  background: var(--glass-bg-secondary, rgba(0,0,0,0.03));
  border: 1px solid var(--glass-border-light, #d1d5db);
  border-radius: 6px;
}

.payload-url {
  font-size: 12px;
  color: var(--color-primary, #3b82f6);
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  flex: 1;
}

.btn-icon {
  background: none;
  border: none;
  cursor: pointer;
  color: var(--color-text-secondary);
  padding: 2px;
  font-size: 14px;
}

.btn-icon:hover {
  color: var(--color-primary);
}

.btn-sm {
  padding: 4px 8px;
  font-size: 12px;
}

.btn-warning {
  background: #f59e0b;
  color: white;
  border-color: #f59e0b;
}

.empty-state {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  height: 100%;
  min-height: 300px;
}

.interaction-table-container {
  flex: 1;
  overflow: auto;
  min-height: 0;
}

.interaction-table {
  width: 100%;
  border-collapse: collapse;
  font-size: 12px;
}

.interaction-table thead {
  position: sticky;
  top: 0;
  z-index: 1;
  background: var(--glass-bg-secondary, #f3f4f6);
}

.interaction-table th {
  padding: 6px 8px;
  text-align: left;
  font-weight: 600;
  color: var(--color-text-secondary);
  border-bottom: 1px solid var(--glass-border-light, #e5e7eb);
  white-space: nowrap;
}

.interaction-table td {
  padding: 6px 8px;
  border-bottom: 1px solid var(--glass-border-light, #f3f4f6);
  color: var(--color-text-primary);
}

.interaction-table tbody tr {
  cursor: pointer;
  transition: background 0.1s;
}

.interaction-table tbody tr:hover {
  background: var(--glass-bg-secondary, rgba(59, 130, 246, 0.05));
}

.interaction-table tbody tr.selected {
  background: rgba(59, 130, 246, 0.1);
}

.protocol-badge {
  display: inline-block;
  padding: 1px 6px;
  border-radius: 4px;
  font-size: 11px;
  font-weight: 600;
}

.protocol-http {
  background: #dbeafe;
  color: #1d4ed8;
}

.protocol-dns {
  background: #dcfce7;
  color: #166534;
}

.protocol-smtp {
  background: #fef3c7;
  color: #92400e;
}

.provider-badge {
  display: inline-block;
  padding: 1px 6px;
  border-radius: 4px;
  font-size: 11px;
  background: #f3f4f6;
  color: #374151;
}

.dark .protocol-http {
  background: rgba(59, 130, 246, 0.2);
  color: #93c5fd;
}

.dark .protocol-dns {
  background: rgba(34, 197, 94, 0.2);
  color: #86efac;
}

.dark .protocol-smtp {
  background: rgba(245, 158, 11, 0.2);
  color: #fcd34d;
}

.dark .provider-badge {
  background: rgba(255, 255, 255, 0.1);
  color: #d1d5db;
}

.detail-panel {
  border-top: 1px solid var(--glass-border-light, #e5e7eb);
  max-height: 40%;
  overflow: auto;
}

.detail-header {
  padding: 6px 8px;
  background: var(--glass-bg-secondary, #f9fafb);
  border-bottom: 1px solid var(--glass-border-light, #e5e7eb);
}

.detail-title {
  font-size: 12px;
  font-weight: 600;
}

.detail-content {
  display: flex;
  gap: 8px;
  padding: 8px;
}

.detail-section {
  flex: 1;
  min-width: 0;
}

.detail-section h4 {
  font-size: 11px;
  font-weight: 600;
  color: var(--color-text-secondary);
  margin-bottom: 4px;
}
</style>
