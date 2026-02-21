<script setup lang="ts">
import { ref, computed, onMounted, onBeforeUnmount, watch } from 'vue';
import { useI18n } from 'vue-i18n';
import IntruderTabs from './IntruderTabs.vue';
import IntruderRequestEditor from './IntruderRequestEditor.vue';
import IntruderPayloadConfig from './IntruderPayloadConfig.vue';
import IntruderHistoryPanel, { type IntruderResult } from './IntruderHistoryPanel.vue';
import IntruderGroupModal from './IntruderGroupModal.vue';
import type { RequestEditorInstance, PayloadPosition, IntruderTab, IntruderGroup, PayloadSet } from '../../types/intruder';
import { useEventListener } from '../../composables/useEventListener';
// @ts-ignore导入 Wails 后端函数
import { Intruder, GetAttackDump } from "../../../bindings/github.com/yhy0/ChYing/app.js";
import { Events } from "@wailsio/runtime";

// 导入拆分出的功能模块
import { useIntruderStore } from './components/IntruderStore';
import { useIntruderPanelResizer } from './components/IntruderPanelResizer';
import { useIntruderAttackController } from './components/IntruderAttackController';
import { useIntruderResultsManager } from './components/IntruderResultsManager';
import { useIntruderUtils } from './components/IntruderUtils';

// Initialize i18n
const { t } = useI18n();

// 使用各个功能模块
const store = useIntruderStore();
const resizer = useIntruderPanelResizer();
const attackController = useIntruderAttackController();
const resultsManager = useIntruderResultsManager();
const utils = useIntruderUtils();

// 添加请求和响应数据的响应式变量
const request = ref('');
const response = ref('');

// 将常用的错误处理逻辑包装为高阶函数
const handleSelectResult = utils.withErrorHandling((result: IntruderResult) => {
  if (!result) return;
  
  const adaptedResult = utils.intruderResultToAttackResult(result);
  resultsManager.selectResult(adaptedResult);

  // 获取详细的请求和响应数据
  if (store.activeTab?.value) {
    // 从Wails后端获取详细数据
    GetAttackDump(store.activeTab.value.id, result.id).then(HTTPBody => {
      request.value = HTTPBody["request_raw"] || '';
      response.value = HTTPBody["response_raw"] || '';
      
      // 如果返回的数据格式不同于预期，进行额外处理
      if (typeof HTTPBody["request_raw"] === 'object') {
        request.value = JSON.stringify(HTTPBody["request_raw"], null, 2);
      }
      if (typeof HTTPBody["response_raw"] === 'object') {
        response.value = JSON.stringify(HTTPBody["response_raw"], null, 2);
      }
    }).catch(err => {
      console.error('获取攻击详情失败:', err);
      request.value = '获取数据失败';
      response.value = '获取数据失败';
    });
  }
}, '处理选中结果');

const startAttack = utils.withErrorHandling(async (tabId: string) => {
  if (!tabId || !store.tabs.value) return;

  // 查找目标标签
  const tab = store.tabs.value.find(t => t.id === tabId);
  if (!tab) return;
  
  // 检查是否有 payload 位置
  if (!tab.payloadPositions || tab.payloadPositions.length === 0) {
    console.warn('开始攻击: 没有设置有效载荷位置，可能会影响结果');
    // 根据实际需求，可能需要提示用户或阻止攻击
  }
  
  // 检查是否有足够的 payload 项
  const hasPayloadItems = tab.payloadSets && 
                         tab.payloadSets.length > 0 && 
                         tab.payloadSets.some(set => set.items.length > 0);
  if (!hasPayloadItems) {
    console.warn('开始攻击: 没有设置有效载荷项');
    // 根据实际需求，可能需要提示用户或阻止攻击
  }
  
  // 启动攻击 - 调用 Wails 函数
  try {
    // 清空之前的结果，重置标签页状态
    tab.results = [];
    
    // 确保结果面板不显示旧数据
    resultsManager.clearSelectedResult();
    if (request) request.value = '';
    if (response) response.value = '';
    
    tab.isRunning = true;
    tab.progress = { total: 0, current: 0, startTime: Date.now(), endTime: null };
    
    const fullRequest = tab.target.fullRequest;

    // 提取和处理payload数据，确保id是连续的数字
    const payloads = tab.payloadSets
      .map((set, index) => ({
        ...set,
        id: index + 1 // 确保id从1开始连续递增
      }));

    const attackType = tab.attackType;
    const requestCount = attackController.calculateTotalRequests(tab);
    tab.progress.total = requestCount;
    
    // 序列化payloads对象为JSON字符串
    const payloadsJson = JSON.stringify(payloads);

    // 修改: 调用后端函数时，添加 tab.id 参数
    Intruder(tab.target.url, fullRequest, payloadsJson, attackType, requestCount, tab.id);
  } catch (error) {
    console.error('启动攻击失败:', error);
    tab.isRunning = false;
    tab.progress.endTime = Date.now();
  }
}, '启动攻击');

const stopAttack = utils.withErrorHandling((tabId: string) => {
  if (!tabId || !store.tabs.value) return;
  
  const tab = store.tabs.value.find(t => t.id === tabId);
  if (!tab) return;
  
  attackController.stopAttack(tab);
}, '停止攻击');

// 分组模态框状态
const showGroupModal = ref(false);

// 分组操作
const openCreateGroupModal = () => {
  showGroupModal.value = true;
};

const handleCreateGroup = (name: string, color: string) => {
  store.createGroup(name, color);
  showGroupModal.value = false;
};

const closeGroupModal = () => {
  showGroupModal.value = false;
};

// Add new ref for the editor component
const requestEditor = ref<RequestEditorInstance | null>(null);

// 有效载荷位置相关操作
const wrapSelectionWithPayloadMarker = () => {
  if (!store.activeTab || !store.activeTab.value) return;
  
  // 使用工具函数获取完整请求
  const fullRequest = store.activeTab.value.target.fullRequest;

  // 获取用户文本选择范围
  const selection = window.getSelection();
  if (!selection || selection.rangeCount === 0 || selection.toString().trim() === '') {
    return;
  }
  
  const selectedText = selection.toString().trim();
  // 使用工具函数处理选中内容
  const updatedRequest = utils.wrapSelectionInRequestWithPayloadMarker(fullRequest, selectedText);
  // 更新请求
  store.activeTab.value.target.fullRequest = updatedRequest;
};

// 清除所有有效载荷标记
const clearAllPayloadMarkers = () => {
  if (!store.activeTab || !store.activeTab.value) return;
  
  // 使用工具函数获取完整请求
  const fullRequest = store.activeTab.value.target.fullRequest;
  
  // 使用工具函数处理
  const updatedRequest = utils.clearAllPayloadMarkersInRequest(fullRequest);
  
  store.activeTab.value.target.fullRequest = updatedRequest;
  
  // 清空payloadPositions
  if (store.activeTab.value) {
    store.updatePayloadPositions(String(store.activeTabId?.value || ''), []);
  }
};

// 当有效载荷位置发生变化
const handlePayloadPositionsChanged = (positions: PayloadPosition[]) => {
  if (!store.activeTab || !store.activeTab.value) return;
  
  store.updatePayloadPositions(String(store.activeTabId?.value || ''), positions);
};

// 处理攻击类型更改
const handleAttackTypeChange = (event: Event) => {
  if (!store.activeTab || !store.activeTab.value) return;
  
  const value = (event.target as HTMLSelectElement).value;
  store.updateAttackType(String(store.activeTabId?.value || ''), value as any);
};

// 监听键盘快捷键
const handleKeyDown = (e: KeyboardEvent) => {
  // 确保没有在输入框中
  if (e.target instanceof HTMLInputElement || e.target instanceof HTMLTextAreaElement) {
    return;
  }
  
  // 可以添加全局快捷键
};

// 使用通用事件监听组合式函数
useEventListener(window, 'keydown', handleKeyDown);

// 当标签ID变化或者组件挂载时，重置结果显示
watch(store.activeTabId, () => {
  resultsManager.clearSelectedResult();
});

// 添加用于存储 Wails 事件数据的响应式变量
const uuid = ref('');
const data = ref<any[]>([]);

// 跟踪所有注册的动态事件监听器，用于清理
const registeredEventKeys = ref<Set<string>>(new Set());

// 组件挂载
onMounted(() => {
  // 初始化面板引用元素，确保resizer能正确获取到面板元素
  setTimeout(() => {
    resizer.initializePanels();
  }, 50);

  // 查找store中已标记为活动的标签
  const activeTabInStore = store.tabs.value.find(tab => tab.isActive);
  
  if (activeTabInStore) {
    // 如果store中有活动标签，使用它的ID
    store.activeTabId.value = activeTabInStore.id;
  } else if (store.tabs.value.length > 0) {
    // 如果没有活动标签但有标签，激活第一个
    store.activeTabId.value = store.tabs.value[0].id;
    store.tabs.value[0].isActive = true;
  } else {
    // 如果没有标签，创建一个新标签
    store.addTab();
  }

  // 修改: 监听 Attack-Data 事件，并根据 tabId 匹配对应的标签页
  // Wails v3: result 是 WailsEvent 对象，result.data 是后端发送的 AttackData 对象
  Events.On("Attack-Data", result => {
    // 后端发送的是 AttackData{Name, UUID, Len} 单个对象
    const attackData = result.data;
    if (!attackData || !attackData.uuid) {
      console.warn("Attack-Data: 无效的事件数据");
      return;
    }
    
    // 获取后端返回的标签页 ID
    const tabId = attackData.uuid || '';
    
    // 查找匹配的标签页
    const matchedTab = store.tabs.value.find(tab => tab.id === tabId);
    if (!matchedTab) {
      console.warn("未找到匹配的标签页:", tabId);
      return;
    }

    // 清空该标签页的结果数据，避免重复
    matchedTab.results = [];
    
    // 清空数据数组，准备接收新结果
    data.value = [];
    
    // 保存 UUID 和数据长度
    const uuidKey = attackData.uuid;
    const dataLength = attackData.len;

    // 先尝试移除之前可能存在的同名事件监听器，避免重复
    try {
      Events.Off(uuidKey);
      registeredEventKeys.value.delete(uuidKey);
    } catch (error) {
      // Listener may not exist, which is fine
    }
    
    // 跟踪新注册的事件监听器
    registeredEventKeys.value.add(uuidKey);
    
    // 监听特定 UUID 的结果数据
    // Wails v3: resultData 是 WailsEvent 对象，resultData.data 是后端发送的 IntruderRes 对象
    Events.On(uuidKey, resultData => {
      // 后端发送的是 *IntruderRes 单个对象
      const resultItem = resultData.data;
      if (!resultItem) {
        console.warn(`UUID ${uuidKey}: 无效的结果数据`);
        return;
      }
      
      // 提取 payload 数组
      const payloadArray: string[] = [];
      if (resultItem.payload && Array.isArray(resultItem.payload)) {
        payloadArray.push(...resultItem.payload);
      }
      
      // 创建 IntruderResult 对象
      const intruderResult = {
        id: String(resultItem.id || 0),
        payload: payloadArray,
        status: resultItem.status || 0,
        length: resultItem.length || 0,
        timeMs: resultItem.timeMs || 0,
        timestamp: Date.now(),
        request: resultItem.request || '',
        response: resultItem.response || ''
      };
      
      // 添加到对应标签页的结果中
      store.addAttackResult(tabId, intruderResult);
      
      // 更新进度信息
      if (matchedTab) {
        matchedTab.progress.current += 1;
        // 如果完成了所有请求，标记为完成
        if (matchedTab.progress.current >= matchedTab.progress.total) {
          matchedTab.isRunning = false;
          matchedTab.progress.endTime = Date.now();
        }
      }
    });
  });
});

// 组件卸载
onBeforeUnmount(() => {
  // 移除拖拽相关事件监听器
  resizer.cleanupResizer();
  
  // 移除所有 Wails 事件监听器
  try {
    Events.Off("Attack-Data");

    // 清理所有动态注册的 UUID 事件监听器
    registeredEventKeys.value.forEach(key => {
      try {
        Events.Off(key);
      } catch (e) {
        console.warn(`移除 ${key} 事件监听器失败:`, e);
      }
    });
    registeredEventKeys.value.clear();
    
  } catch (error) {
    console.error("移除事件监听器失败:", error);
  }
});

// 追踪结果数量变化，用于在结果更新时触发视图刷新
const resultsCount = computed(() => store.activeTab.value?.results.length || 0);

// 设置结果颜色
function setResultColor(resultId: string, color: string) {
  if (!store.activeTab?.value) return;
  store.setResultColor(String(store.activeTabId?.value || ''), resultId, color);
}

// 处理结果右键菜单
function handleResultContextMenu(result: IntruderResult, event: MouseEvent) {
  // 使用store来处理右键菜单，而不是resultsManager
  const adaptedResult = utils.intruderResultToAttackResult(result);
  // 先选中结果
  resultsManager.selectResult(adaptedResult);
  // 然后由外部处理上下文菜单
}

// 为IntruderHistoryPanel组件准备格式化的结果数据
const formattedResults = computed(() => {
  if (!store.activeTab?.value?.results) return [];
  // 使用工具函数格式化结果
  return store.activeTab.value.results.map(utils.formatIntruderResult);
});

// Make sure tabs and groups are properly reactive with correct types
const tabsForComponent = computed<IntruderTab[]>(() => {
  return Array.isArray(store.tabs.value) ? store.tabs.value : [];
});

const groupsForComponent = computed<IntruderGroup[]>(() => {
  return Array.isArray(store.groups.value) ? store.groups.value : [];
});

// Make sure resizer.leftPanelWidth is a number
const leftPanelWidthNumber = computed(() => {
  const width = resizer.leftPanelWidth;
  return typeof width === 'number' ? width : Number(width || 50);
});

// 为 IntruderHistoryPanel 组件创建一个新的计算属性 selectedResultForPanel，确保类型兼容性
const selectedResultForPanel = computed<IntruderResult | null>(() => {
  const rawResult = resultsManager.selectedResultWithType.value;
  if (!rawResult) return null;
  return rawResult as IntruderResult;
});

// 确保载荷集类型兼容性
function ensurePayloadSetTypeCompatibility(payloadSets: any[]): any[] {
  if (!payloadSets || !Array.isArray(payloadSets)) return [];
  
  return payloadSets.map(set => {
    const newSet = { ...set };
    
    // 将 'brute-forcer' 转换为 'brute-force'，确保类型兼容
    if (newSet.type === 'brute-forcer') {
      newSet.type = 'brute-force';
    }
    
    // 如果存在旧的 options 字段，但没有 processing 字段，创建 processing 字段
    if (newSet.options && !newSet.processing) {
      newSet.processing = {
        rules: [],
        encoding: {
          enabled: false,
          urlEncode: false,
          characterSet: 'UTF-8'
        }
      };
      // 可以考虑从 options 中迁移数据到 processing
    }
    
    // 确保 processing 字段存在
    if (!newSet.processing) {
      newSet.processing = {
        rules: [],
        encoding: {
          enabled: false,
          urlEncode: false,
          characterSet: 'UTF-8'
        }
      };
    }
    
    return newSet;
  });
}

// 处理请求更新
function handleRequestUpdate(request: string) {
  if (!store.activeTab || !store.activeTab.value) return;

  // 直接更新整个请求字符串，不再解析各部分
  store.updateRequest(String(store.activeTabId?.value || ''), request);
  
  // 提取并更新payload位置
  const positions = extractPayloadPositions(request);
  store.updatePayloadPositions(String(store.activeTabId?.value || ''), positions);
}

// 提取payload位置的函数(从IntruderRequestEditor复制过来)
const extractPayloadPositions = (text: string): PayloadPosition[] => {
  // 使用工具函数的实现
  return utils.extractPayloadPositions(text);
};
</script>

<template>
  <div class="h-full flex flex-col">
    <!-- Control Bar -->
    <div class="intruder-control-bar">
      <div class="flex items-center space-x-4">
        <button
          class="btn btn-primary"
          @click="store.addTab"
          :title="t('modules.intruder.new_attack')"
        >
          <i class="bx bx-plus mr-1"></i> {{ t('modules.intruder.new_attack') }}
        </button>
        
        <button
          v-if="store.activeTab && store.activeTab.value && !store.activeTab.value.isRunning"
          class="btn btn-secondary"
          @click="startAttack(String(store.activeTabId?.value || ''))"
          :disabled="!store.activeTab"
          :class="{ 'opacity-50 cursor-not-allowed': !store.activeTab }"
        >
          <i class="bx bx-play mr-1"></i> {{ t('modules.intruder.start_attack') }}
        </button>
        
        <button
          v-else-if="store.activeTab && store.activeTab.value && store.activeTab.value.isRunning"
          class="btn btn-secondary"
          @click="stopAttack(String(store.activeTabId?.value || ''))"
        >
          <i class="bx bx-stop mr-1"></i> {{ t('modules.intruder.stop_attack') }}
        </button>
        
        <button
          class="btn btn-secondary"
          @click="openCreateGroupModal"
          :title="t('modules.intruder.create_group')"
        >
          <i class="bx bx-folder-plus mr-1"></i> {{ t('modules.intruder.create_group') }}
        </button>
      </div>
      
      <div class="flex items-center">
        <div 
          v-if="store.activeTab && store.activeTab.value && store.activeTab.value.progress && store.activeTab.value.progress.total > 0" 
          class="flex items-center space-x-3"
        >
          <div class="text-xs text-gray-600 dark:text-gray-400">
            {{ t('modules.intruder.progress') }}: {{ store.activeTab.value.progress.current }} / {{ store.activeTab.value.progress.total }}
            <span v-if="store.activeTab.value.progress.startTime" class="ml-2">
              {{ store.activeTab.value.progress.endTime ? t('modules.intruder.completed') : t('modules.intruder.running') }}
            </span>
          </div>
          <div class="w-40 h-2 bg-gray-200 dark:bg-gray-700 rounded-full overflow-hidden">
            <div 
              class="h-full bg-btn-primary rounded-full transition-all" 
              :style="{
                width: `${(store.activeTab.value.progress.current / store.activeTab.value.progress.total) * 100}%`
              }"
            ></div>
          </div>
        </div>
      </div>
    </div>
    
    <!-- Tabs Bar -->
    <div class="flex-none overflow-visible">
      <IntruderTabs 
        :tabs="tabsForComponent"
        :groups="groupsForComponent"
        @select-tab="store.activateTab"
        @close-tab="store.closeTab"
        @rename-tab="store.renameTab"
        @change-tab-color="store.changeTabColor"
        @change-tab-group="store.changeTabGroup"
        @create-group="openCreateGroupModal"
        @reorder-tabs="store.handleReorderTabs"
        @reorder-groups="store.reorderGroups"
      />
    </div>

    <!-- Main content -->
    <div v-if="store.activeTab && store.activeTab.value" class="flex-1 flex flex-col overflow-hidden">
      <!-- Configuration and results panel -->
      <div class="flex-1 flex flex-col overflow-hidden">
        <!-- 攻击配置面板 -->
        <div class="p-3 border-b border-gray-200 dark:border-gray-700 bg-white dark:bg-[#1e1e2e]">
          <div class="flex flex-wrap sm:flex-nowrap items-center gap-3">
            <!-- 攻击类型选择 -->
            <div class="flex items-center gap-2 whitespace-nowrap">
              <label class="text-xs text-gray-500 dark:text-gray-400">{{ t('modules.intruder.attack_type') }}:</label>
              <select
                class="text-sm border border-gray-300 dark:border-gray-700 rounded bg-white dark:bg-[#282838] text-gray-700 dark:text-gray-300 py-1 px-2 min-w-[120px]"
                :value="store.activeTab.value.attackType" 
                @change="handleAttackTypeChange"
              >
                <option v-for="type in attackController.attackTypes" :key="type.value" :value="type.value">
                  {{ t(`modules.intruder.attack_types.${type.value}`) }}
                </option>
              </select>
            </div>

            <!-- 攻击类型描述 -->
            <div class="flex-1 text-xs text-gray-600 dark:text-gray-400 italic overflow-hidden text-ellipsis">
              {{ t(`modules.intruder.attack_descriptions.${store.activeTab.value.attackType}`) }}
            </div>
          </div>
        </div>

        <!-- 内容区域：分为请求编辑器和结果区域 -->
        <div 
          class="flex-1 flex flex-col md:flex-row overflow-hidden intruder-resizable-panels" 
          id="intruder-panels-container"
        >
          <!-- 左侧请求编辑面板 -->
          <div 
            class="w-full md:w-1/2 border-r border-gray-200 dark:border-gray-700 overflow-y-auto intruder-panel-left flex flex-col" 
            id="intruder-panel-left"
            :style="{ width: `${leftPanelWidthNumber}%` }"
          >
            <div class="p-4 flex flex-col flex-grow">
              <h3 class="text-sm font-medium text-gray-700 dark:text-gray-300 mb-3">
                {{ t('modules.intruder.request_editor') }}
              </h3>
              <div class="mb-2 flex items-start">
                <div class="flex space-x-2">
                  <button
                    class="btn btn-success"
                    @click="wrapSelectionWithPayloadMarker"
                    title="Mark selected text as payload position"
                  >
                    <i class="bx bx-plus-circle"></i> {{ t('modules.intruder.add_position') }}
                  </button>
                  <button
                    class="btn btn-danger"
                    @click="clearAllPayloadMarkers"
                    title="Clear all payload positions"
                  >
                    <i class="bx bx-trash"></i> {{ t('modules.intruder.clear_all') }}
                  </button>
                </div>
                <div class="ml-4 text-xs text-gray-600 dark:text-gray-400 flex-1">
                  <div class="flex items-center justify-between">
                    <p v-html="t('modules.intruder.add_marker_instructions')"></p>
                    <span class="text-xs bg-gray-100 dark:bg-gray-800 px-2 py-0.5 rounded-full">
                      {{ t('modules.intruder.payload_positions') }}: {{ store.activeTab.value.payloadPositions.length }}
                    </span>
                  </div>
                  <p class="mt-1">{{ t('modules.intruder.marker_example') }}: <span class="font-mono">username=<span class="text-[#4f46e5] font-bold">§</span>admin<span class="text-[#4f46e5] font-bold">§</span>&password=secret</span></p>
                </div>
              </div>
              <div class="editor-container flex-grow" style="min-height: 400px; height: calc(50vh - 100px);">
                <IntruderRequestEditor
                  ref="requestEditor"
                  :request="store.activeTab.value.target.fullRequest"
                  payloadMarker="§"
                  @update:request="handleRequestUpdate"
                  @payload-positions-changed="handlePayloadPositionsChanged"
                />
              </div>
            </div>

            <!-- Payload Positions section - show only when there are positions -->
            <div v-if="store.activeTab.value.payloadPositions.length > 0" class="p-3 border-t border-gray-200 dark:border-gray-700">
              <div v-for="(pos, index) in store.activeTab.value.payloadPositions" :key="index"
                class="mb-2 p-2 border border-gray-200 dark:border-gray-700 rounded-md bg-gray-50 dark:bg-[#282838]">
                <div class="flex items-center justify-between mb-1">
                  <span class="text-xs font-medium text-gray-700 dark:text-gray-300">
                    {{ t('modules.intruder.position_number', { number: index + 1 }) }}
                    <span v-if="pos.paramName" class="text-[#4f46e5]">({{ pos.paramName }})</span>
                  </span>
                </div>
                <div class="text-xs text-gray-600 dark:text-gray-400 font-mono bg-white dark:bg-[#1e1e2e] p-1 rounded">
                  {{ pos.value }}
                </div>
              </div>
            </div>

            <!-- 使用IntruderPayloadConfig组件 -->
            <IntruderPayloadConfig
              :payload-sets="store.activeTab.value.payloadSets ? ensurePayloadSetTypeCompatibility(store.activeTab.value.payloadSets) : []"
              :payload-positions="store.activeTab.value.payloadPositions"
              :attack-type="store.activeTab.value.attackType"
              @update:payload-sets="updatedSets => { 
                if (store.activeTab && store.activeTab.value) {
                  store.updatePayloadSets(store.activeTabId.value || '', ensurePayloadSetTypeCompatibility(updatedSets));
                }
              }"
              @add-payload-item="(value, tabIndex) => {
                if (store.activeTab && store.activeTab.value) {
                  store.addPayloadItem(store.activeTabId.value || '', tabIndex, value);
                }
              }"
              @paste-payload-items="(items, tabIndex) => {
                if (store.activeTab && store.activeTab.value) {
                  items.forEach(item => {
                    store.addPayloadItem(store.activeTabId.value || '', tabIndex, item);
                  });
                }
              }"
              @load-payload-items="(items, tabIndex) => {
                if (store.activeTab && store.activeTab.value) {
                  items.forEach(item => {
                    store.addPayloadItem(store.activeTabId.value || '', tabIndex, item);
                  });
                }
              }"
              @remove-payload-item="(itemIndex, tabIndex) => {
                if (store.activeTab && store.activeTab.value) {
                  store.removePayloadItem(store.activeTabId.value || '', tabIndex, itemIndex);
                }
              }"
              @clear-payload-items="(tabIndex) => {
                if (store.activeTab && store.activeTab.value) {
                  store.clearPayloadItems(store.activeTabId.value || '', tabIndex);
                }
              }"
            />
          </div>

          <!-- 分隔线 -->
          <div 
            class="intruder-panel-divider" 
            title="拖动调整面板大小"
            @mousedown.prevent="resizer.startResize($event)"
          ></div>

          <!-- 右侧结果面板 -->
          <div 
            class="w-full md:w-1/2 flex flex-col overflow-hidden intruder-panel-right"
            id="intruder-panel-right"
            :style="{ width: `${100 - Number(leftPanelWidthNumber)}%` }"
          >
            <!-- 使用IntruderHistoryPanel显示结果 -->
            <IntruderHistoryPanel
              :key="'history-panel-' + resultsCount"
              :results="formattedResults"
              :selectedResult="selectedResultForPanel"
              :request="request"
              :response="response"
              @select-result="handleSelectResult"
              @set-color="(item, color) => setResultColor(String(item.id), color)"
              @context-menu="(event, item) => handleResultContextMenu(item, event)"
            />
          </div>
        </div>
      </div>
    </div>

    <!-- 空状态 -->
    <div v-else class="flex-1 flex flex-col items-center justify-center text-gray-400 dark:text-gray-600">
      <i class="bx bx-target-lock text-5xl mb-4"></i>
      <h3 class="text-lg font-medium mb-2">{{ t('modules.intruder.no_tab_open') }}</h3>
      <p class="text-sm mb-4">{{ t('modules.intruder.create_new_tab') }}</p>
      <button class="px-3 py-1.5 text-xs bg-[#4f46e5] hover:bg-[#4338ca] text-white rounded-md flex items-center"
        @click="store.addTab">
        <i class="bx bx-plus mr-1"></i> {{ t('modules.intruder.new_tab') }}
      </button>
    </div>

    <!-- 分组模态框 -->
    <IntruderGroupModal
      :show="showGroupModal"
      @close="closeGroupModal"
      @create="handleCreateGroup"
    />
  </div>
</template>

<style scoped>
/* 简单的过渡效果 */
.fade-enter-active,
.fade-leave-active {
  transition: opacity 0.3s;
}

.fade-enter-from,
.fade-leave-to {
  opacity: 0;
}
</style>