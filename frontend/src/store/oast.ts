import { defineStore } from 'pinia';
import { ref, computed } from 'vue';
// @ts-ignore
import {
  OASTCreateProvider,
  OASTUpdateProvider,
  OASTDeleteProvider,
  OASTListProviders,
  OASTToggleProvider,
  OASTRegister,
  OASTStartPolling,
  OASTStopPolling,
  OASTPollOnce,
  OASTDeregister,
  OASTGetSettings,
  OASTUpdateSettings,
} from '../../bindings/github.com/yhy0/ChYing/app.js';
// @ts-ignore
import { Events } from '@wailsio/runtime';

export interface OASTProvider {
  id: string;
  name: string;
  type: string;
  url: string;
  token: string;
  enabled: boolean;
  builtin?: boolean;
  payloadUrl?: string;
  registered?: boolean;
  polling?: boolean;
  createdAt?: string;
}

export interface OASTInteraction {
  id: string;
  providerId: string;
  type: string;
  protocol: string;
  method: string;
  source: string;
  destination: string;
  timestamp: string;
  correlationId: string;
  rawRequest: string;
  rawResponse: string;
  data?: any;
}

export interface OASTTab {
  id: string;
  name: string;
  providerId: string;
  payloadUrl: string;
  interactions: OASTInteraction[];
}

export interface OASTSettings {
  pollingInterval: number;
  payloadPrefix: string;
}

export const useOastStore = defineStore('oast', () => {
  // ==================== 状态定义 ====================
  const providers = ref<OASTProvider[]>([]);
  const tabs = ref<OASTTab[]>([]);
  const activeTabId = ref<string>('');
  const settings = ref<OASTSettings>({
    pollingInterval: 5000,
    payloadPrefix: '',
  });
  const isLoading = ref(false);
  const selectedInteraction = ref<OASTInteraction | null>(null);
  const showProviderManager = ref(false);

  // ==================== 计算属性 ====================
  const activeTab = computed(() => tabs.value.find(t => t.id === activeTabId.value));
  const totalInteractions = computed(() => tabs.value.reduce((sum, t) => sum + t.interactions.length, 0));

  // ==================== Provider 管理 ====================
  const loadProviders = async () => {
    isLoading.value = true;
    try {
      const result = await OASTListProviders();
      if (!result.error && result.data) {
        providers.value = result.data;
      }
    } catch (e) {
      console.error('Load OAST providers failed:', e);
    } finally {
      isLoading.value = false;
    }
  };

  const createProvider = async (provider: Partial<OASTProvider>) => {
    try {
      const result = await OASTCreateProvider(provider as any);
      if (result.error) {
        console.error('Create provider failed:', result.error);
        return null;
      }
      await loadProviders();
      return result.data;
    } catch (e) {
      console.error('Create provider failed:', e);
      return null;
    }
  };

  const updateProvider = async (id: string, updates: Partial<OASTProvider>) => {
    try {
      const result = await OASTUpdateProvider(id, updates as any);
      if (result.error) {
        console.error('Update provider failed:', result.error);
        return false;
      }
      await loadProviders();
      return true;
    } catch (e) {
      console.error('Update provider failed:', e);
      return false;
    }
  };

  const deleteProvider = async (id: string) => {
    try {
      const result = await OASTDeleteProvider(id);
      if (result.error) {
        console.error('Delete provider failed:', result.error);
        return false;
      }
      await loadProviders();
      // 清理相关 tab
      tabs.value = tabs.value.filter(t => t.providerId !== id);
      return true;
    } catch (e) {
      console.error('Delete provider failed:', e);
      return false;
    }
  };

  const toggleProvider = async (id: string, enabled: boolean) => {
    try {
      const result = await OASTToggleProvider(id, enabled);
      if (result.error) {
        console.error('Toggle provider failed:', result.error);
        return false;
      }
      const idx = providers.value.findIndex(p => p.id === id);
      if (idx !== -1) {
        providers.value[idx].enabled = enabled;
      }
      return true;
    } catch (e) {
      console.error('Toggle provider failed:', e);
      return false;
    }
  };

  // ==================== OAST 操作 ====================
  const register = async (providerId: string) => {
    try {
      const result = await OASTRegister(providerId);
      if (result.error) {
        console.error('Register failed:', result.error);
        return null;
      }
      const payloadUrl = result.data as string;

      // 更新 provider 状态
      const idx = providers.value.findIndex(p => p.id === providerId);
      if (idx !== -1) {
        providers.value[idx].payloadUrl = payloadUrl;
        providers.value[idx].registered = true;
      }

      return payloadUrl;
    } catch (e) {
      console.error('Register failed:', e);
      return null;
    }
  };

  const startPolling = async (providerId: string) => {
    try {
      const result = await OASTStartPolling(providerId, settings.value.pollingInterval);
      if (result.error) {
        console.error('Start polling failed:', result.error);
        return false;
      }
      const idx = providers.value.findIndex(p => p.id === providerId);
      if (idx !== -1) {
        providers.value[idx].polling = true;
      }
      return true;
    } catch (e) {
      console.error('Start polling failed:', e);
      return false;
    }
  };

  const stopPolling = async (providerId: string) => {
    try {
      const result = await OASTStopPolling(providerId);
      if (result.error) {
        console.error('Stop polling failed:', result.error);
        return false;
      }
      const idx = providers.value.findIndex(p => p.id === providerId);
      if (idx !== -1) {
        providers.value[idx].polling = false;
      }
      return true;
    } catch (e) {
      console.error('Stop polling failed:', e);
      return false;
    }
  };

  const pollOnce = async (providerId: string) => {
    try {
      const result = await OASTPollOnce(providerId);
      if (result.error) {
        console.error('Poll once failed:', result.error);
        return [];
      }
      return (result.data || []) as OASTInteraction[];
    } catch (e) {
      console.error('Poll once failed:', e);
      return [];
    }
  };

  const deregister = async (providerId: string) => {
    try {
      const result = await OASTDeregister(providerId);
      if (result.error) {
        console.error('Deregister failed:', result.error);
        return false;
      }
      const idx = providers.value.findIndex(p => p.id === providerId);
      if (idx !== -1) {
        providers.value[idx].payloadUrl = '';
        providers.value[idx].registered = false;
        providers.value[idx].polling = false;
      }
      return true;
    } catch (e) {
      console.error('Deregister failed:', e);
      return false;
    }
  };

  // ==================== Tab 管理 ====================
  const addTab = (providerId: string, providerName: string, payloadUrl: string) => {
    const id = `tab-${Date.now()}`;
    const tab: OASTTab = {
      id,
      name: providerName,
      providerId,
      payloadUrl,
      interactions: [],
    };
    tabs.value.push(tab);
    activeTabId.value = id;
    return tab;
  };

  const removeTab = (tabId: string) => {
    const idx = tabs.value.findIndex(t => t.id === tabId);
    if (idx !== -1) {
      tabs.value.splice(idx, 1);
      if (activeTabId.value === tabId) {
        activeTabId.value = tabs.value.length > 0 ? tabs.value[tabs.value.length - 1].id : '';
      }
    }
  };

  const setActiveTab = (tabId: string) => {
    activeTabId.value = tabId;
  };

  // ==================== 事件处理 ====================
  const addInteraction = (interaction: OASTInteraction) => {
    // 添加到匹配的 tab
    for (const tab of tabs.value) {
      if (tab.providerId === interaction.providerId) {
        tab.interactions.unshift(interaction);
        return;
      }
    }
    // 如果没有匹配的 tab，添加到第一个
    if (tabs.value.length > 0) {
      tabs.value[0].interactions.unshift(interaction);
    }
  };

  const clearInteractions = (tabId?: string) => {
    if (tabId) {
      const tab = tabs.value.find(t => t.id === tabId);
      if (tab) {
        tab.interactions = [];
      }
    } else {
      tabs.value.forEach(t => { t.interactions = []; });
    }
    selectedInteraction.value = null;
  };

  // ==================== 设置管理 ====================
  const loadSettings = async () => {
    try {
      const result = await OASTGetSettings();
      if (!result.error && result.data) {
        settings.value = result.data;
      }
    } catch (e) {
      console.error('Load OAST settings failed:', e);
    }
  };

  const saveSettings = async (newSettings: Partial<OASTSettings>) => {
    try {
      const merged = { ...settings.value, ...newSettings };
      const result = await OASTUpdateSettings(merged);
      if (!result.error) {
        settings.value = merged;
      }
    } catch (e) {
      console.error('Save OAST settings failed:', e);
    }
  };

  // ==================== 事件监听初始化 ====================
  let eventListenerRegistered = false;

  const initEventListener = () => {
    if (eventListenerRegistered) return;
    eventListenerRegistered = true;

    Events.On('oast:interaction', (event: any) => {
      const data = event.data?.[0] || event.data;
      if (data) {
        addInteraction(data);
      }
    });
  };

  return {
    // 状态
    providers,
    tabs,
    activeTabId,
    settings,
    isLoading,
    selectedInteraction,
    showProviderManager,

    // 计算属性
    activeTab,
    totalInteractions,

    // Provider 管理
    loadProviders,
    createProvider,
    updateProvider,
    deleteProvider,
    toggleProvider,

    // OAST 操作
    register,
    startPolling,
    stopPolling,
    pollOnce,
    deregister,

    // Tab 管理
    addTab,
    removeTab,
    setActiveTab,

    // 事件处理
    addInteraction,
    clearInteractions,

    // 设置
    loadSettings,
    saveSettings,

    // 初始化
    initEventListener,
  };
});
