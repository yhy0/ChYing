import { ref, computed, watch } from 'vue';
import { useModulesStore } from '../../../store/modules';
import type { AttackType, PayloadPosition, AttackResult, IntruderTab, IntruderGroup, IntruderResult } from '../../../types/intruder.ts';
import { useI18n } from 'vue-i18n';

// 导出组合式API hook
export function useIntruderStore() {
  const { t } = useI18n();
  
  // 使用store
  const store = useModulesStore();

  // Define state
  const tabs = computed(() => store.intruderTabs);
  const groups = computed(() => store.intruderGroups);
  const activeTabId = ref<string | null>(null);
  
  // Computed active tab
  const activeTab = computed(() => {
    // 首先检查store中是否有被标记为active的标签
    const storeActiveTab = tabs.value.find(tab => tab.isActive);
    
    // 如果store中有active的标签但与当前activeTabId不一致，则更新activeTabId
    if (storeActiveTab && storeActiveTab.id !== activeTabId.value) {
      activeTabId.value = storeActiveTab.id;
    }
    
    // 然后根据activeTabId查找标签
    const foundTab = tabs.value.find(tab => tab.id === activeTabId.value);
    if (foundTab) {
      // Ensure attackType is correctly typed
      return {
        ...foundTab,
        attackType: foundTab.attackType as AttackType
      };
    }
    return null;
  });

  // Activate a tab
  const activateTab = (tabId: string) => {
    // First deactivate all tabs
    tabs.value.forEach(tab => {
      tab.isActive = false;
    });
    
    // Then activate only the selected tab
    const tab = tabs.value.find(tab => tab.id === tabId);
    if (tab) {
      tab.isActive = true;
      activeTabId.value = tabId;
    }
  };

  // 默认的 HTTP 请求模板
  const defaultRequest = `GET / HTTP/1.1
Host: 127.0.0.1
User-Agent: Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36
Accept: text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,*/*;q=0.8
Accept-Language: en-US,en;q=0.5
Accept-Encoding: gzip, deflate
Connection: close

`;

  // Add a new tab
  const addTab = () => {
    const currentCounter = store.intruderTabCounter;
    store.intruderTabCounter++;
    const newId = `tab-${Date.now()}`;
    const newTab: IntruderTab = {
      id: newId,
      name: `${t('modules.intruder.attack')} ${currentCounter}`,
      color: '#4f46e5', // 使用固定的颜色
      groupId: null,
      target: {
        url: 'http://127.0.0.1/',
        method: 'GET',
        headers: '',
        body: '',
        fullRequest: defaultRequest
      },
      attackType: 'sniper', // 默认使用狙击手模式
      payloadPositions: [],
      payloadSets: [
        {
          id: 1,
          type: 'simple-list',
          items: [],
          processing: {
            rules: [],
            encoding: {
              enabled: false,
              urlEncode: false,
              characterSet: 'UTF-8'
            }
          }
        },
      ],
      results: [],
      isActive: true,
      isRunning: false,
      progress: { total: 0, current: 0, startTime: null, endTime: null }
    };

    // 设置所有其他标签为非活动状态
    tabs.value.forEach(tab => tab.isActive = false);

    // 添加新标签
    tabs.value.push(newTab);
    activeTabId.value = newId;
  };

  // Close a tab
  const closeTab = (tabId: string) => {
    const index = tabs.value.findIndex(tab => tab.id === tabId);
    if (index === -1) return;

    tabs.value.splice(index, 1);

    // If we closed the active tab, activate another one
    if (activeTabId.value === tabId) {
      if (tabs.value.length > 0) {
        activeTabId.value = tabs.value[Math.min(index, tabs.value.length - 1)].id;
        tabs.value[Math.min(index, tabs.value.length - 1)].isActive = true;
      } else {
        activeTabId.value = null;
      }
    }
  };

  // Rename a tab
  const renameTab = (tabId: string, newName: string) => {
    const tab = tabs.value.find(tab => tab.id === tabId);
    if (tab) {
      tab.name = newName;
    }
  };

  // 修改标签颜色
  const changeTabColor = (tabId: string, color: string) => {
    const tab = tabs.value.find(tab => tab.id === tabId);
    if (tab) {
      tab.color = color;
    }
  };

  // 处理标签重新排序
  const handleReorderTabs = (newTabs: IntruderTab[]) => {
    // 更新store中的标签顺序
    store.intruderTabs.splice(0, store.intruderTabs.length, ...newTabs);
  };

  // 更新请求目标
  const updateTarget = (tabId: string, field: keyof IntruderTab['target'], value: string) => {
    const tab = tabs.value.find(tab => tab.id === tabId);
    if (tab) {
      tab.target[field] = value;
    }
  };

  // 更新整个请求 (新增方法)
  const updateRequest = (tabId: string, request: string) => {
    const tab = tabs.value.find(tab => tab.id === tabId);
    if (!tab) return;
    
    // 直接存储整个请求，不再解析
    // 为了保持数据结构一致性，我们先添加一个新字段，不影响现有字段
    tab.target.fullRequest = request;
  };

  // 更改标签所属分组
  const changeTabGroup = (tabId: string, groupId: string | null) => {
    store.changeIntruderTabGroup(tabId, groupId);
  };

  // 创建新分组
  const createGroup = (name: string, color: string = '#4f46e5') => {
    store.createIntruderGroup(name, color);
  };

  // 重新排序分组
  const reorderGroups = (newGroups: IntruderGroup[]) => {
    store.intruderGroups = newGroups;
  };

  // 设置结果颜色
  const setResultColor = (tabId: string, resultId: string, color: string) => {
    const tab = tabs.value.find(t => t.id === tabId);
    if (!tab) return;
    
    const resultIndex = tab.results.findIndex(r => String(r.id) === String(resultId));
    if (resultIndex !== -1) {
      tab.results[resultIndex].color = color;
    }
  };

  // 当有效载荷位置发生变化
  const updatePayloadPositions = (tabId: string, positions: PayloadPosition[]) => {
    const tab = tabs.value.find(tab => tab.id === tabId);
    if (tab) {
      tab.payloadPositions = positions;
    }
  };

  // 处理攻击类型更改
  const updateAttackType = (tabId: string, attackType: AttackType) => {
    const tab = tabs.value.find(tab => tab.id === tabId);
    if (tab) {
      tab.attackType = attackType;
    }
  };

  // 添加攻击结果
  const addAttackResult = (tabId: string, result: AttackResult) => {
    const tab = tabs.value.find(tab => tab.id === tabId);
    if (tab) {
      // 转换 AttackResult 为 IntruderResult
      const intruderResult: IntruderResult = {
        id: Number(result.id || 0),
        payload: result.payload,
        status: result.status,
        length: result.length,
        timeMs: result.timeMs,
        timestamp: String(result.timestamp || ''),
        request: result.request,
        response: result.response
      };
      
      // 使用数组方法触发响应式更新
      tab.results = [...tab.results, intruderResult];
    }
  };

  // 更新所有 payload sets
  const updatePayloadSets = (tabId: string, payloadSets: any[]) => {
    const tab = tabs.value.find(tab => tab.id === tabId);
    if (tab) {
      tab.payloadSets = [...payloadSets];
    }
  };

  // 更新指定位置的 payload set
  const updatePayloadSet = (tabId: string, index: number, payloadSet: any) => {
    const tab = tabs.value.find(tab => tab.id === tabId);
    if (tab) {
      // 确保 payloadSets 数组存在并且长度足够
      while (tab.payloadSets.length <= index) {
        tab.payloadSets.push({
          id: tab.payloadSets.length + 1,
          type: 'simple-list',
          items: [],
          processing: {
            rules: [],
            encoding: {
              enabled: false,
              urlEncode: false,
              characterSet: 'UTF-8'
            }
          }
        });
      }
      // 更新指定位置的 payload set
      tab.payloadSets[index] = { ...payloadSet };
      // 触发响应式更新
      tab.payloadSets = [...tab.payloadSets];
    }
  };

  // 添加 payload 项到指定 payload set
  const addPayloadItem = (tabId: string, index: number, item: string) => {
    const tab = tabs.value.find(tab => tab.id === tabId);
    if (tab) {
      // 确保 payloadSets 数组存在并且长度足够
      while (tab.payloadSets.length <= index) {
        tab.payloadSets.push({
          id: Date.now() + tab.payloadSets.length,
          type: 'simple-list',
          items: [],
          processing: {
            rules: [],
            encoding: {
              enabled: false,
              urlEncode: false,
              characterSet: 'UTF-8'
            }
          }
        });
      }
      // 添加项目
      tab.payloadSets[index].items.push(item);
      // 触发响应式更新
      tab.payloadSets = [...tab.payloadSets];
    }
  };

  // 清空指定 payload set 的所有项目
  const clearPayloadItems = (tabId: string, index: number) => {
    const tab = tabs.value.find(tab => tab.id === tabId);
    if (tab && tab.payloadSets[index]) {
      tab.payloadSets[index].items = [];
      // 触发响应式更新
      tab.payloadSets = [...tab.payloadSets];
    }
  };

  // 移除指定 payload set 中的特定项目
  const removePayloadItem = (tabId: string, payloadSetIndex: number, itemIndex: number) => {
    const tab = tabs.value.find(tab => tab.id === tabId);
    if (tab && tab.payloadSets[payloadSetIndex]) {
      tab.payloadSets[payloadSetIndex].items.splice(itemIndex, 1);
      // 触发响应式更新
      tab.payloadSets = [...tab.payloadSets];
    }
  };

  // 更新处理规则和编码设置 - 直接在PayloadSet中更新，不使用单独的payloadSetProcessings
  const updatePayloadSetProcessings = (tabId: string, processings: any[]) => {
    const tab = tabs.value.find(tab => tab.id === tabId);
    if (tab) {
      // 更新每个PayloadSet的processing字段
      for (let i = 0; i < processings.length; i++) {
        if (i < tab.payloadSets.length) {
          tab.payloadSets[i].processing = {
            rules: processings[i].processingRules || [],
            encoding: processings[i].encoding || {
              enabled: false,
              urlEncode: false,
              characterSet: 'UTF-8'
            }
          };
        }
      }
      // 触发响应式更新
      tab.payloadSets = [...tab.payloadSets];
    }
  };

  // 更新单个位置的处理规则和编码设置
  const updatePayloadSetProcessing = (tabId: string, index: number, processing: any) => {
    const tab = tabs.value.find(tab => tab.id === tabId);
    if (tab) {
      // 确保payloadSets数组长度足够
      while (tab.payloadSets.length <= index) {
        tab.payloadSets.push({
          id: tab.payloadSets.length,
          type: 'simple-list',
          items: [],
          processing: {
            rules: [],
            encoding: {
              enabled: false,
              urlEncode: false,
              characterSet: 'UTF-8'
            }
          }
        });
      }
      
      // 更新指定位置的PayloadSet处理规则
      tab.payloadSets[index].processing = {
        rules: processing.processingRules || [],
        encoding: processing.encoding || {
          enabled: false,
          urlEncode: false,
          characterSet: 'UTF-8'
        }
      };
      
      // 触发响应式更新
      tab.payloadSets = [...tab.payloadSets];
    }
  };

  // 监视activeTabId的变化，确保UI状态同步
  watch(activeTabId, (newTabId, oldTabId) => {
    if (newTabId) {
      // 确保所有标签的isActive属性与activeTabId保持一致
      tabs.value.forEach(tab => {
        tab.isActive = tab.id === newTabId;
      });
    }
  }, { immediate: true });

  // 监视tabs数组变化，确保新添加的激活标签被正确识别
  watch(tabs, (newTabs) => {
    const activeTabInStore = newTabs.find(tab => tab.isActive);
    if (activeTabInStore && activeTabInStore.id !== activeTabId.value) {
      // 如果store中有新的激活标签，更新当前选中的标签ID
      activeTabId.value = activeTabInStore.id;
    }
  }, { deep: true });

  // 开始攻击
  function startAttack(tabId: string) {
    const tab = tabs.value.find(t => t.id === tabId);
    if (!tab) return;
    
    // 因为这个方法不完整，我们在这里不执行任何操作
  }

  // 类型转换辅助函数
  function adaptResultType(result: IntruderResult): AttackResult {
    return {
      id: String(result.id),
      payload: result.payload,
      status: result.status,
      length: result.length,
      timeMs: result.timeMs,
      request: result.request,
      response: result.response,
      timestamp: Date.now() // 使用当前时间戳替代结果中的字符串时间戳
    };
  }

  // 返回所有需要在组件中使用的状态和方法
  return {
    tabs,
    groups,
    activeTabId,
    activeTab,
    activateTab,
    addTab,
    closeTab,
    renameTab,
    changeTabColor,
    handleReorderTabs,
    updateTarget,
    updateRequest,
    changeTabGroup,
    createGroup,
    reorderGroups,
    setResultColor,
    updatePayloadPositions,
    updateAttackType,
    addAttackResult,
    startAttack,
    adaptResultType,
    updatePayloadSets,
    updatePayloadSet,
    addPayloadItem,
    clearPayloadItems,
    removePayloadItem,
    updatePayloadSetProcessings,
    updatePayloadSetProcessing
  };
}