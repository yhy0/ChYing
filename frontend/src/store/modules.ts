import { defineStore } from 'pinia';
import { ref, computed, type Ref } from 'vue';
import { generateUUID } from '../utils';
import {
  prepareIntruderSourceTarget,
} from '../utils';
import eventBus, {
  SEND_TO_REPEATER_REQUESTED,
  SEND_TO_INTRUDER_FROM_PROXY_REQUESTED,
  SEND_TO_INTRUDER_FROM_REPEATER_REQUESTED
} from '../utils/eventBus';
import type {
  ProxyHistoryItem,
  RepeaterTab,
  RepeaterGroup,
  IntruderTab,
  DecoderTab,
  NotificationState,
  NotificationItem,
  NotificationSeverity,
  IntruderGroup,
} from '../types';

// 创建store
export const useModulesStore = defineStore('modules', () => {
  // 注意：activeModule 状态已移除，现在由路由管理模块切换

  // Proxy模块状态
  const proxyHistory = ref<ProxyHistoryItem[]>([]);
  
  // Repeater模块状态
  const repeaterTabs = ref<RepeaterTab[]>([]);
  const repeaterTabCounter = ref(1);
  
  // Repeater分组状态（初始为空，用户可自行创建分组）
  const repeaterGroups = ref<RepeaterGroup[]>([]);
  
  // Intruder模块状态
  const intruderTabs = ref<IntruderTab[]>([]);
  const intruderTabCounter = ref(1);
  
  // Intruder分组状态
  const intruderGroups = ref<IntruderGroup[]>([]);
  
  // Decoder模块状态
  const decoderTabs = ref<DecoderTab[]>([]);
  const decoderTabCounter = ref(1);
  
  // 通知相关状态
  const notifications = ref<NotificationState>({
    showNotifications: false,
    unreadCount: 0,
    items: []
  });
  
  interface BaseTabForCreation {
    id: string;
    name: string;
    isActive: boolean;
    color?: string;
    groupId?: string | null;
  }
  
  // 内部通用辅助函数，处理标签页创建的共同逻辑
    // 内部通用辅助函数，处理标签页创建的共同逻辑
  function _createGenericTab<T extends BaseTabForCreation>(
    tabsArrayRef: Ref<T[]>,
    tabCounterRef: Ref<number>,
    defaultBaseName: string, 
    specificPropsFactory: (id: string, generatedName: string, isActive: boolean) => T 
  ): string {
    tabsArrayRef.value.forEach(tab => { 
      tab.isActive = false;
    });
    const newTabId = generateUUID();
    const currentCounter = tabCounterRef.value;
    tabCounterRef.value++;
    const generatedName = defaultBaseName ? `${defaultBaseName} ${currentCounter}` : `Tab ${currentCounter}`;
    const isActive = true;
  
    const newTab = specificPropsFactory(newTabId, generatedName, isActive);
    
    tabsArrayRef.value.push(newTab);
    return newTabId;
  }

  // 创建新的Decoder标签页 (重构后)
  function createDecoderTab(name?: string) {
    return _createGenericTab<DecoderTab>(
      decoderTabs,
      decoderTabCounter,
      'Decoder', // Default base name
      (id, generatedName, isActive) => ({
        id: id,
        name: name || generatedName, // Use provided name or the generated one
        initialInput: '',
        steps: [],
        isActive: isActive
      } as DecoderTab) // Cast for safety, ensuring all DecoderTab props are met
    );
  }
  
  // 更新Decoder标签页状态
  function updateDecoderTab(tabId: string, updates: Partial<DecoderTab>) {
    const index = decoderTabs.value.findIndex(tab => tab.id === tabId);
    if (index !== -1) {
      // 使用 Object.assign 或展开运算符确保响应式更新
      decoderTabs.value[index] = Object.assign({}, decoderTabs.value[index], updates);
    }
  }
  
  // 设置活动的Decoder标签页
  function setActiveDecoderTab(tabId: string) {
    decoderTabs.value.forEach(tab => {
      tab.isActive = tab.id === tabId;
    });
  }
  
  // 关闭Decoder标签页
  function closeDecoderTab(tabId: string) {
    const index = decoderTabs.value.findIndex(tab => tab.id === tabId);
    if (index !== -1) {
      const wasActive = decoderTabs.value[index].isActive;
      
      // 移除标签页
      decoderTabs.value.splice(index, 1);
      
      // 如果关闭的是活动标签页，激活另一个标签页
      if (wasActive && decoderTabs.value.length > 0) {
        const newActiveIndex = Math.min(index, decoderTabs.value.length - 1);
        decoderTabs.value[newActiveIndex].isActive = true;
      }
    }
  }
  
  // 获取当前活动的Decoder标签页
  const activeDecoderTab = computed(() => {
    return decoderTabs.value.find(tab => tab.isActive) || null;
  });
  
  // 添加代理历史记录
  function addProxyHistoryItem(item: ProxyHistoryItem) {
    // 检查是否已存在相同 ID 的项目
    if (item.id && proxyHistory.value.some(existingItem => existingItem.id === item.id)) {
      // 找到已存在的项目并返回
      return proxyHistory.value.find(existingItem => existingItem.id === item.id) as ProxyHistoryItem;
    }
    
    // 确保项目有唯一ID
    const newId = proxyHistory.value.length > 0 
      ? Math.max(...proxyHistory.value.map(h => h.id)) + 1 
      : 1;
    
    const newItem: ProxyHistoryItem = {
      ...item,
      id: item.id || newId
    };
    
    proxyHistory.value.unshift(newItem);
    return newItem;
  }
  
  // 设置Proxy历史项目的颜色
  function setProxyItemColor(itemId: number, color: string) {
    const index = proxyHistory.value.findIndex(item => item.id === itemId);
    if (index !== -1) {
      proxyHistory.value[index].color = color;
    }
  }
  
  // 清除所有代理历史记录
  function clearProxyHistory() {
    // 通过赋值一个新的空数组来确保响应式触发
    proxyHistory.value = [];
    
    // 触发响应式更新的另一种方法
    // 虽然这里的操作没有实际效果，但会强制 Vue 认为数组发生了变化
    setTimeout(() => {
      const temp = proxyHistory.value;
      proxyHistory.value = [...temp];
    }, 0);
  }
  
  // 设置代理历史记录（用于筛选等）
  function setProxyHistory(items: ProxyHistoryItem[]) {
    proxyHistory.value = items;
  }
  
  // REVISED Actions (to be called by event listeners)
  function addRepeaterTabFromEventPayload(
    payload: { sourceItem: ProxyHistoryItem } 
  ): string { 
    const { sourceItem } = payload;
    const requestContent = sourceItem.request;
    const baseName = `# ${sourceItem.id})`; 
    const initialColor = '#4f46e5';

    // 从 ProxyHistoryItem 中提取 method 和 url
    const method = sourceItem.method || 'GET';
    
    // 从完整 URL 中提取域名部分（去除路径）
    let baseUrl = '';
    if (sourceItem.url) {
      try {
        const urlObj = new URL(sourceItem.url);
        baseUrl = `${urlObj.protocol}//${urlObj.host}`;
      } catch (error) {
        // 如果 URL 解析失败，使用原始 URL 或默认值
        baseUrl = sourceItem.url;
      }
    }

    return _createGenericTab<RepeaterTab>(
      repeaterTabs,
      repeaterTabCounter,
      baseName, 
      (id, generatedName, isActive) => ({
        id: id,
        name: generatedName, 
        color: initialColor,
        groupId: null,
        request: requestContent, 
        response: null,
        isActive: isActive,
        modified: false,
        serverDurationMs: 0,
        method: method,
        url: baseUrl,
      } as RepeaterTab)
    );
  }

  function addIntruderTabFromEventPayload(
    payload: { sourceItem: ProxyHistoryItem | RepeaterTab } 
  ): string { 
    const { sourceItem } = payload;
    
    const targetInfo = prepareIntruderSourceTarget(sourceItem);
    
    const baseName = `# ${sourceItem.id})`;
    const initialColor = '#e11d48';

    return _createGenericTab<IntruderTab>(
      intruderTabs,
      intruderTabCounter,
      baseName, 
      (id, generatedName, isActive) => ({
        id: id,
        name: generatedName,
        color: initialColor,
        groupId: null,
        target: targetInfo, 
        attackType: 'sniper', 
        payloadPositions: [],
        payloadSets: [ 
          { 
            id: 1, // Reverted to number 1, as PayloadSet expects a number ID
            type: 'simple-list', 
            items: [], 
            processing: { rules: [], encoding: { enabled: false, urlEncode: false, characterSet: 'UTF-8' } } 
          }
        ],
        results: [],
        isActive: isActive,
        isRunning: false,
        progress: { total: 0, current: 0, startTime: null, endTime: null },
      } as IntruderTab)
    );
  }

  // Modified "sendTo..." 函数
  function sendToRepeater(proxyItem: ProxyHistoryItem) {
    eventBus.emit(SEND_TO_REPEATER_REQUESTED, { sourceItem: proxyItem });
  }
  
  function sendToIntruder(proxyItem: ProxyHistoryItem) {
    // The targetDetails extraction will be done by the event handler.
    eventBus.emit(SEND_TO_INTRUDER_FROM_PROXY_REQUESTED, { 
      sourceItem: proxyItem 
      // targetDetails: targetInfo // Removed from here
    });
  }
  
  function sendRepeaterToIntruder(repeaterTab: RepeaterTab) {
    // The targetDetails extraction will be done by the event handler.
    eventBus.emit(SEND_TO_INTRUDER_FROM_REPEATER_REQUESTED, { 
      sourceItem: repeaterTab
      // targetDetails: targetInfo // Removed from here
    });
  }
  
  // Define BaseGroup if it's not already available from types
  interface BaseGroup { id: string; name: string; color: string; }

  function _createGroup<G extends BaseGroup>(
    groupsArrayRef: Ref<G[]>,
    name: string,
    color: string = '#4f46e5' 
  ): string {
    const newId = generateUUID();
    const newGroup = { id: newId, name, color } as G;
    groupsArrayRef.value.push(newGroup);
    return newId;
  }

  function _changeItemGroupGroup<T extends { id: string, groupId: string | null, color?: string }, 
                                 G extends BaseGroup>(
    itemsArrayRef: Ref<T[]>,
    itemId: string,
    newGroupId: string | null,
    groupsArrayRef?: Ref<G[]> 
  ): void {
    const item = itemsArrayRef.value.find(t => t.id === itemId);
    if (item) {
      item.groupId = newGroupId;
      if (newGroupId && groupsArrayRef) {
        const group = groupsArrayRef.value.find(g => g.id === newGroupId);
        // Ensure item.color is assignable and group exists
        if (group && item.hasOwnProperty('color')) { 
          item.color = group.color;
        }
      }
    }
  }

  // 创建Repeater分组 (重构后)
  function createRepeaterGroup(name: string, color: string = '#4f46e5') {
    return _createGroup<RepeaterGroup>(repeaterGroups, name, color);
  }
  
  // 创建Intruder分组 (重构后)
  function createIntruderGroup(name: string, color: string = '#4f46e5') {
    return _createGroup<IntruderGroup>(intruderGroups, name, color);
  }
  
  // 更改标签页分组 (Repeater - 重构后)
  function changeTabGroup(tabId: string, groupId: string | null) {
    _changeItemGroupGroup<RepeaterTab, RepeaterGroup>(repeaterTabs, tabId, groupId, repeaterGroups);
  }
  
  // 更改Intruder标签页分组 (重构后)
  function changeIntruderTabGroup(tabId: string, groupId: string | null) {
    _changeItemGroupGroup<IntruderTab, IntruderGroup>(intruderTabs, tabId, groupId, intruderGroups);
  }
  
  // 通知相关方法
  function toggleNotifications() {
    notifications.value.showNotifications = !notifications.value.showNotifications;
    
    // 当打开通知面板时，将未读数量重置为0
    if (notifications.value.showNotifications) {
      notifications.value.unreadCount = 0;
    }
  }
  
  // 关闭通知面板
  function closeNotifications() {
    notifications.value.showNotifications = false;
    notifications.value.unreadCount = 0;
  }
  
  // 设置未读通知数量
  function setUnreadCount(count: number) {
    notifications.value.unreadCount = count;
  }
  
  // 添加新通知
  function addNotification(notification: { type: string; title?: string; message: string; }) {
    const newItem: NotificationItem = {
      id: generateUUID(),
      title: notification.title || '',
      message: notification.message,
      timestamp: new Date().toISOString(),
      read: false,
      severity: notification.type as NotificationSeverity
    };

    // 添加到通知列表开头
    if (!notifications.value.items) {
      notifications.value.items = [];
    }
    notifications.value.items.unshift(newItem);

    // 增加未读数量
    notifications.value.unreadCount++;
  }

  // 清除所有通知
  function clearNotifications() {
    notifications.value.items = [];
    notifications.value.unreadCount = 0;
  }

  // 标记通知为已读
  function markNotificationAsRead(notificationId: string) {
    const item = notifications.value.items?.find(n => n.id === notificationId);
    if (item && !item.read) {
      item.read = true;
      if (notifications.value.unreadCount > 0) {
        notifications.value.unreadCount--;
      }
    }
  }

  // 标记所有通知为已读
  function markAllNotificationsAsRead() {
    notifications.value.items?.forEach(item => {
      item.read = true;
    });
    notifications.value.unreadCount = 0;
  }
  
  // 返回所有公开的状态和方法
  return {
    // 状态
    proxyHistory,
    repeaterTabs,
    repeaterTabCounter,
    repeaterGroups,
    intruderTabs,
    intruderTabCounter,
    intruderGroups,
    decoderTabs,
    decoderTabCounter,
    notifications,

    // 计算属性
    activeDecoderTab,

    // 方法
    createDecoderTab,
    updateDecoderTab,
    setActiveDecoderTab,
    closeDecoderTab,
    addProxyHistoryItem,
    setProxyItemColor,
    clearProxyHistory,
    setProxyHistory,
    sendToRepeater,
    sendToIntruder,
    sendRepeaterToIntruder,
    createRepeaterGroup,
    createIntruderGroup,
    changeTabGroup,
    changeIntruderTabGroup,
    toggleNotifications,
    closeNotifications,
    setUnreadCount,
    addNotification,
    clearNotifications,
    markNotificationAsRead,
    markAllNotificationsAsRead,
    // 新增 actions
    addRepeaterTabFromEventPayload,
    addIntruderTabFromEventPayload,
  };
});