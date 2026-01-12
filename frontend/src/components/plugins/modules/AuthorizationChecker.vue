<script setup lang="ts">
import { ref, reactive, computed, onMounted, onUnmounted, nextTick, h } from 'vue';
import { useI18n } from 'vue-i18n';
import RequestResponsePanel from '../../common/RequestResponsePanel.vue';
import HttpTrafficTable from '../../common/HttpTrafficTable.vue';
import type { HttpTrafficItem, HttpTrafficColumn } from '../../../types';

const { t } = useI18n();

interface AuthTestResult {
  id: number;
  originalRequest: string;
  originalResponse: string;
  modifiedRequest: string;
  modifiedResponse: string;
  ruleID: number;
  ruleDescription: string;
  statusCode: number;
  originalStatus: number;
  url: string;
  method: string;
  timestamp: string;
}

// @ts-ignore - 忽略自动生成的Wails绑定文件类型错误
import {
  StartAuthorizationCheck,
  StopAuthorizationCheck,
  SaveAuthorizationRules,
  GetAuthorizationRules,
  GetAuthorizationTestResults,
  ClearAuthorizationTestResults,
  GetAuthorizationCheckStatus
// @ts-ignore
} from "../../../../bindings/github.com/yhy0/ChYing/app.js";

// 替换规则类型
interface ReplaceRule {
  id: number;
  enabled: boolean;
  type: 'replace' | 'remove';
  headerName: string;
  originalValue: string;
  newValue: string;
  description: string;
}

// 过滤条件类型 (逻辑保留，UI暂时移除)
interface FilterCondition {
  onlyInScope: boolean;
  includeUrls: string[];
  excludeUrls: string[];
  includeFileTypes: string[];
  excludeFileTypes: string[];
}

// 视图模式类型
type DetailViewMode = 'original' | 'modified' | 'response';

// 状态
const isRunning = ref(false);
const isLoading = ref(false);
const showAddRuleDialog = ref(false);
const showFilterConfigDialog = ref(false); // 新增：过滤配置弹窗
const editingRuleIndex = ref(-1);
const lastResultsCount = ref(0); // 用于跟踪新结果

// 测试结果
const testResults = ref<AuthTestResult[]>([]);
const selectedResult = ref<AuthTestResult | null>(null);
const currentDetailView = ref<DetailViewMode>('original');

// 结果筛选
const resultFilter = ref('');

// 表格重新渲染key
const tableRenderKey = ref(0);

// 移除了会话相关的状态

// 规则列表
const rules = ref<ReplaceRule[]>([]);

// 当前编辑的规则
const currentRule = reactive<ReplaceRule>({
  id: 0,
  enabled: true,
  type: 'replace',
  headerName: '',
  originalValue: '',
  newValue: '',
  description: ''
});

// 过滤条件 (逻辑保留)
const filterCondition = reactive<FilterCondition>({
  onlyInScope: true,
  includeUrls: [],
  excludeUrls: [],
  includeFileTypes: [],
  excludeFileTypes: []
});

// 过滤配置表单的文本输入框绑定
const includeUrlsText = ref('');
const excludeUrlsText = ref('');
const includeFileTypesText = ref('');
const excludeFileTypesText = ref('');

// 通知消息
const notification = ref({
  show: false,
  message: '',
  type: 'success'
});

// 拖拽调整面板大小相关变量 (下半部分左右)
const resultsListWidth = ref(50); // 左侧结果列表初始宽度百分比
let isHorizontalDragging = false;
let horizontalDragStartX = 0;
let initialResultsListPixelWidth = 0;
const resultsSectionRef = ref<HTMLElement | null>(null); // 下半部分容器引用

// 拖拽调整面板大小相关变量 (上下)
const rulesSectionHeight = ref(300); // 规则区域初始高度 (px)
const minRulesSectionHeight = 150; // 最小规则区高度
const minResultsSectionHeight = 200; // 最小结果区高度
let isVerticalDragging = false;
let verticalDragStartY = 0;
let initialRulesSectionPixelHeight = 0;
const layoutContainerRef = ref<HTMLElement | null>(null); // 整体容器引用

// 在组合式函数中增加规则状态获取
const authorizationRules = ref({
  rules: [] as ReplaceRule[],
  filterCondition: {
    onlyInScope: false,
    includeUrls: [] as string[],
    excludeUrls: [] as string[],
    includeFileTypes: [] as string[],
    excludeFileTypes: [] as string[]
  }
});

// 过滤后的测试结果 (用于结果列表)
const filteredResults = computed(() => {
  if (!resultFilter.value) return testResults.value;

  const filter = resultFilter.value.toLowerCase();
  return testResults.value.filter(result =>
    result.url.toLowerCase().includes(filter) ||
    result.ruleDescription.toLowerCase().includes(filter) ||
    String(result.statusCode).includes(filter)
  );
});

// 状态码样式类函数
const getStatusClass = (status: number) => {
  if (status >= 200 && status < 300) return 'status-2xx';
  if (status >= 300 && status < 400) return 'status-3xx';
  if (status >= 400 && status < 500) return 'status-4xx';
  if (status >= 500) return 'status-5xx';
  return '';
};

// 自定义表格列定义 - 调整列顺序：id 在第一列，时间在最后一列
const customColumns = computed(() => [
  { id: 'id', name: '#', width: 80 },
  { id: 'url', name: 'URL', width: 300 },
  { id: 'ruleDescription', name: t('modules.plugins.authorization_checker.description'), width: 200 },
  { 
    id: 'statusCode', 
    name: t('modules.plugins.authorization_checker.status_codes_display'), 
    width: 120,
    cellRenderer: ({ item }: { item: any }) => {
      // 创建一个 VNode 显示 "原始状态/修改后状态"
      return h('div', { class: 'status-codes-display' }, [
        h('span', { class: ['status-code', getStatusClass(item.originalStatus)] }, item.originalStatus),
        h('span', { class: 'status-separator' }, '/'),
        h('span', { class: ['status-code', getStatusClass(item.status)] }, item.status)
      ]);
    }
  },
  { id: 'timestamp', name: t('common.time'), width: 140 }
]);

// 适配 AuthTestResult 到 HttpTrafficItem 接口
const adaptTestResult = (result: AuthTestResult): HttpTrafficItem => {
  const url = new URL(result.url);
  return {
    id: result.id,
    method: result.method || 'GET', // 使用后端提供的 method 字段
    url: result.url,
    status: result.statusCode,
    host: url.host,
    path: url.pathname,
    length: result.modifiedResponse.length,
    mimeType: '', // 根据响应头推断，目前默认为空
    extension: '', // 授权测试结果通常不关心文件扩展名
    title: result.ruleDescription,
    ip: '', // 授权测试结果通常不需要IP
    note: '', // 授权测试结果通常不需要备注
    timestamp: result.timestamp,
    // 保留原始数据用于扩展
    originalStatus: result.originalStatus,
    originalRequest: result.originalRequest,
    originalResponse: result.originalResponse,
    modifiedRequest: result.modifiedRequest,
    modifiedResponse: result.modifiedResponse,
    ruleID: result.ruleID,
    ruleDescription: result.ruleDescription
  };
};

// 将每个测试结果适配为 HttpTrafficItem
const adaptedResults = computed(() => 
  filteredResults.value.map(adaptTestResult)
);

// HttpTrafficTable 选中项适配
const adaptedResult = computed(() => 
  selectedResult.value ? adaptTestResult(selectedResult.value) : null
);

// 处理 HttpTrafficTable 的项目选择
const handleItemSelect = (item: HttpTrafficItem) => {
  const found = testResults.value.find(result => result.id === item.id);
  if (found) {
    selectedResult.value = found;
    currentDetailView.value = 'original'; // 默认显示原始请求
  }
};

// 定时获取结果
const startPollingResults = () => {
  const pollInterval = setInterval(async () => {
    if (isRunning.value) {
      await fetchTestResults();
    } else {
      clearInterval(pollInterval);
    }
  }, 3000);

  onUnmounted(() => {
    clearInterval(pollInterval);
  });
};

// 开始/停止检测
const toggleCheck = async () => {
  isLoading.value = true;
  try {
    if (isRunning.value) {
      await StopAuthorizationCheck();
      isRunning.value = false;
      showNotification(t('modules.plugins.authorization_checker.notifications.detection_stopped'), 'success');
    } else {
      await saveRules(); // 确保最新的规则、会话、过滤条件已保存
      await StartAuthorizationCheck();
      isRunning.value = true;
      showNotification(t('modules.plugins.authorization_checker.notifications.detection_started'), 'success');
      startPollingResults();
    }
  } catch (error) {
    showNotification(t('modules.plugins.authorization_checker.notifications.operation_failed', { error }), 'error');
  } finally {
    isLoading.value = false;
  }
};

// 获取测试结果
const fetchTestResults = async () => {
  try {
    const response = await GetAuthorizationTestResults();
    console.log('AuthorizationChecker - 获取测试结果:', response);
    
    if (response && Array.isArray(response)) { // 假设后端直接返回 TestResult[]
      // 确保响应式更新 - 创建新数组而不是直接赋值
      const newResults = [...response];
      console.log('AuthorizationChecker - 更新测试结果 (数组格式):', newResults.length, '条记录');
      testResults.value = newResults;

      if (testResults.value.length > lastResultsCount.value) {
        const newCount = testResults.value.length - lastResultsCount.value;
        showNotification(t('modules.plugins.authorization_checker.notifications.new_results', { count: newCount }), 'info');
        lastResultsCount.value = testResults.value.length;
      }
    } else if (response && response.data && Array.isArray(response.data)) { // 兼容旧的包裹格式
      // 确保响应式更新 - 创建新数组而不是直接赋值
      const newResults = [...response.data];
      console.log('AuthorizationChecker - 更新测试结果 (包裹格式):', newResults.length, '条记录');
      testResults.value = newResults;
      
      if (testResults.value.length > lastResultsCount.value) {
        const newCount = testResults.value.length - lastResultsCount.value;
        showNotification(t('modules.plugins.authorization_checker.notifications.new_results', { count: newCount }), 'info');
        lastResultsCount.value = testResults.value.length;
      }
    }
    
    console.log('AuthorizationChecker - 当前测试结果:', testResults.value.length, '条');
    console.log('AuthorizationChecker - 过滤后结果:', filteredResults.value.length, '条');
    console.log('AuthorizationChecker - 适配后结果:', adaptedResults.value.length, '条');
    
    // 强制触发重新渲染
    tableRenderKey.value++;
    console.log('AuthorizationChecker - 表格渲染key更新为:', tableRenderKey.value);
  } catch (error) {
    console.error("获取测试结果失败:", error);
    showNotification(t('modules.plugins.authorization_checker.notifications.operation_failed', { error }), 'error'); // 避免过于频繁的错误提示
  }
};

// 清除测试结果
const clearResults = async () => {
  try {
    await ClearAuthorizationTestResults();
    testResults.value = [];
    selectedResult.value = null; // 清除选中的结果
    lastResultsCount.value = 0;
    // 强制触发重新渲染
    tableRenderKey.value++;
    showNotification(t('modules.plugins.authorization_checker.notifications.results_cleared'), 'success');
  } catch (error) {
    showNotification(t('modules.plugins.authorization_checker.notifications.clear_failed', { error }), 'error');
  }
};

// 保存规则和过滤条件
const saveRules = async (showSavedNotification = true) => {
  isLoading.value = true;
  try {
    const dataToSave = {
      rules: rules.value,
      filterCondition: filterCondition
    };
    console.log(dataToSave);
    await SaveAuthorizationRules(dataToSave);
    if (showSavedNotification) {
      showNotification(t('modules.plugins.authorization_checker.notifications.config_saved'), 'success');
    }
  } catch (error) {
    showNotification(t('modules.plugins.authorization_checker.notifications.save_failed', { error }), 'error');
  } finally {
    isLoading.value = false;
  }
};

// 加载规则和过滤条件
const loadRules = async () => {
  isLoading.value = true;
  try {
    const response = await GetAuthorizationRules();
    if (response.error) {
      throw new Error(response.error);
    }

    const data = response.data || { rules: [], filterCondition: {} };
    rules.value = data.rules || [];
    
    // 更新规则状态
    authorizationRules.value = data;

    if (data.filterCondition) {
      Object.assign(filterCondition, data.filterCondition);
    }

    // 获取当前检测状态
    try {
      isRunning.value = await GetAuthorizationCheckStatus();
      if (isRunning.value) {
        startPollingResults(); // 如果正在运行，开始轮询结果
      }
    } catch (error) {
      console.warn('获取检测状态失败:', error);
    }

    showNotification(t('modules.plugins.authorization_checker.notifications.rules_loaded'), 'success');
    await fetchTestResults(); // 同时加载已有结果
  } catch (error) {
    showNotification(t('modules.plugins.authorization_checker.notifications.load_failed', { error }), 'error');
  } finally {
    isLoading.value = false;
  }
};

// 添加/编辑规则
const openRuleDialog = (rule?: ReplaceRule) => {
  if (rule) {
    Object.assign(currentRule, rule);
    editingRuleIndex.value = rules.value.findIndex(r => r.id === rule.id);
  } else {
    Object.assign(currentRule, {
      id: Date.now(),
      enabled: true,
      type: 'replace',
      headerName: '',
      originalValue: '',
      newValue: '',
      description: ''
    });
    editingRuleIndex.value = -1;
  }
  showAddRuleDialog.value = true;
};

// 保存当前编辑的规则
const saveRule = () => {
  if (!currentRule.headerName) {
    showNotification(t('modules.plugins.authorization_checker.header_name') + ' ' + t('common.status.required'), 'error');
    return;
  }

  if (editingRuleIndex.value >= 0) {
    rules.value[editingRuleIndex.value] = { ...currentRule };
  } else {
    rules.value.push({ ...currentRule });
  }

  showAddRuleDialog.value = false;
  saveRules();
};

// 删除规则
const deleteRule = (id: number) => {
  const index = rules.value.findIndex(r => r.id === id);
  if (index !== -1) {
    rules.value.splice(index, 1);
    saveRules();
  }
};

// 切换规则启用状态
const toggleRule = (id: number) => {
  const rule = rules.value.find(r => r.id === id);
  if (rule) {
    rule.enabled = !rule.enabled;
    // 强制触发响应式更新
    rules.value = [...rules.value];
    saveRules(false); // 保存但不频繁提示 "已保存"
  }
};

// 数据转换函数：将规则转换为 HttpTrafficItem 格式
const transformedRules = computed(() => {
  return rules.value.map(rule => ({
    id: rule.id,
    enabled: rule.enabled,
    type: rule.type,
    headerName: rule.headerName,
    originalValue: rule.originalValue,
    newValue: rule.newValue,
    description: rule.description,
    // HttpTrafficItem 必需字段的默认值
    method: 'RULE' as const,
    url: rule.headerName,
    status: rule.enabled ? 200 : 0,
    timestamp: Date.now(),
    size: 0,
    host: 'rule',
    path: rule.headerName
  }));
});

// 当前选中的规则
const selectedRule = ref<any>(null);

// 列定义
const authorizationColumns = computed<HttpTrafficColumn<any>[]>(() => [
  {
    id: 'enabled',
    name: t('modules.plugins.authorization_checker.rule_enabled'),
    width: 80,
    cellRenderer: ({ item }) => {
      // 找到原始规则对象以确保响应式
      const originalRule = rules.value.find(r => r.id === item.id);
      return h('label', {
        class: 'toggle-switch',
        onClick: (e: Event) => {
          e.stopPropagation();
          toggleRule(item.id);
        }
      }, [
        h('input', {
          type: 'checkbox',
          checked: originalRule?.enabled || false,
          onChange: (e: Event) => {
            e.stopPropagation();
            toggleRule(item.id);
          },
          onClick: (e: Event) => {
            e.stopPropagation();
          }
        }),
        h('span', { class: 'toggle-slider' })
      ]);
    }
  },
  {
    id: 'type',
    name: t('modules.plugins.authorization_checker.rule_type'),
    width: 100,
    cellRenderer: ({ item }) => h('span', {},
      item.type === 'replace'
        ? t('modules.plugins.authorization_checker.replace_operation')
        : t('modules.plugins.authorization_checker.remove_operation')
    )
  },
  {
    id: 'headerName',
    name: t('modules.plugins.authorization_checker.header_name'),
    width: 150
  },
  {
    id: 'originalValue',
    name: t('modules.plugins.authorization_checker.original_value'),
    width: 180,
    cellRenderer: ({ item }) => h('code', { class: 'code-snippet' },
      item.originalValue || t('modules.plugins.authorization_checker.any_value')
    )
  },
  {
    id: 'newValue',
    name: t('modules.plugins.authorization_checker.new_value'),
    width: 180,
    cellRenderer: ({ item }) => item.type === 'replace'
      ? h('code', { class: 'code-snippet' }, item.newValue)
      : h('span', { class: 'removed-value' }, t('modules.plugins.authorization_checker.empty_value'))
  },
  {
    id: 'description',
    name: t('modules.plugins.authorization_checker.description'),
    width: 200
  },
  {
    id: 'actions',
    name: t('common.actions.action'),
    width: 120,
    cellRenderer: ({ item }) => h('div', { class: 'actions-column' }, [
      h('button', {
        class: 'btn-icon edit-button',
        title: t('common.actions.edit'),
        onClick: (e: Event) => {
          e.stopPropagation();
          openRuleDialog(item);
        }
      }, [h('i', { class: 'bx bx-edit' })]),
      h('button', {
        class: 'btn-icon delete-button',
        title: t('common.actions.delete'),
        onClick: (e: Event) => {
          e.stopPropagation();
          deleteRule(item.id);
        }
      }, [h('i', { class: 'bx bx-trash' })])
    ])
  }
]);

// 处理规则选择
const handleRuleSelect = (rule: any) => {
  selectedRule.value = rule;
};

// 显示通知
const showNotification = (message: string, type: 'success' | 'error' | 'info' = 'success') => {
  notification.value = { show: true, message, type };
  setTimeout(() => { notification.value.show = false; }, 3000);
};

// 打开过滤配置弹窗
const openFilterConfigDialog = () => {
  // 将数组转换为文本格式用于编辑
  includeUrlsText.value = filterCondition.includeUrls.join('\n');
  excludeUrlsText.value = filterCondition.excludeUrls.join('\n');
  includeFileTypesText.value = filterCondition.includeFileTypes.join(', ');
  excludeFileTypesText.value = filterCondition.excludeFileTypes.join(', ');
  showFilterConfigDialog.value = true;
};

// 保存过滤配置
const saveFilterConfig = async () => {
  try {
    // 将文本转换回数组格式
    filterCondition.includeUrls = includeUrlsText.value
      .split('\n')
      .map(url => url.trim())
      .filter(url => url.length > 0);
    
    filterCondition.excludeUrls = excludeUrlsText.value
      .split('\n')
      .map(url => url.trim())
      .filter(url => url.length > 0);
    
    filterCondition.includeFileTypes = includeFileTypesText.value
      .split(',')
      .map(type => type.trim())
      .filter(type => type.length > 0);
    
    filterCondition.excludeFileTypes = excludeFileTypesText.value
      .split(',')
      .map(type => type.trim())
      .filter(type => type.length > 0);

    const dataToSave = {
      rules: rules.value,
      filterCondition: filterCondition
    };
    await SaveAuthorizationRules(dataToSave);
    authorizationRules.value.filterCondition = { ...filterCondition };
    showFilterConfigDialog.value = false;
    showNotification(t('modules.plugins.authorization_checker.notifications.filter_config_saved'), 'success');
  } catch (error) {
    showNotification(t('modules.plugins.authorization_checker.notifications.save_failed', { error }), 'error');
  }
};

// RequestResponsePanel的动态props
const requestResponsePanelProps = computed(() => {
  if (!selectedResult.value) return null;
  
  // 根据视图类型决定显示原始请求/响应还是修改后的请求/响应
  const isOriginalView = currentDetailView.value === 'original';
  
  return {
    requestData: isOriginalView ? selectedResult.value.originalRequest : selectedResult.value.modifiedRequest,
    responseData: isOriginalView ? selectedResult.value.originalResponse : selectedResult.value.modifiedResponse,
    requestReadOnly: true,
    responseReadOnly: true,
    serverDurationMs: 0,
    title: {
      request: isOriginalView ? t('modules.plugins.authorization_checker.original_request') : t('modules.plugins.authorization_checker.modified_request'),
      response: isOriginalView ? t('modules.plugins.authorization_checker.original_response') : t('modules.plugins.authorization_checker.modified_response')
    }
  };
});


// 水平拖拽逻辑 (下半部分左右)
const startHorizontalResize = (e: MouseEvent) => {
  e.preventDefault();
  isHorizontalDragging = true;
  horizontalDragStartX = e.clientX;
  if (resultsSectionRef.value) {
    initialResultsListPixelWidth = (resultsListWidth.value / 100) * resultsSectionRef.value.clientWidth;
  }
  document.addEventListener('mousemove', handleHorizontalMouseMove);
  document.addEventListener('mouseup', stopHorizontalResize);
  document.body.style.cursor = 'col-resize';
};

const handleHorizontalMouseMove = (e: MouseEvent) => {
  if (!isHorizontalDragging || !resultsSectionRef.value) return;
  const deltaX = e.clientX - horizontalDragStartX;
  const containerWidth = resultsSectionRef.value.clientWidth;
  if (containerWidth === 0) return;

  let newPixelWidth = initialResultsListPixelWidth + deltaX;
  // 限制最小20% 最大80%
  newPixelWidth = Math.max(containerWidth * 0.2, Math.min(containerWidth * 0.8, newPixelWidth));
  resultsListWidth.value = (newPixelWidth / containerWidth) * 100;
};

const stopHorizontalResize = () => {
  isHorizontalDragging = false;
  document.removeEventListener('mousemove', handleHorizontalMouseMove);
  document.removeEventListener('mouseup', stopHorizontalResize);
  document.body.style.cursor = 'default';
};

// 垂直拖拽逻辑 (上下)
const startVerticalResize = (e: MouseEvent) => {
  e.preventDefault();
  isVerticalDragging = true;
  verticalDragStartY = e.clientY;
  initialRulesSectionPixelHeight = rulesSectionHeight.value;

  document.addEventListener('mousemove', handleVerticalMouseMove);
  document.addEventListener('mouseup', stopVerticalResize);
  document.body.style.cursor = 'row-resize';
};

const handleVerticalMouseMove = (e: MouseEvent) => {
  if (!isVerticalDragging || !layoutContainerRef.value) return;
  const deltaY = e.clientY - verticalDragStartY;
  const totalHeight = layoutContainerRef.value.clientHeight;

  // --panel-divider-width (8px) is the height of the horizontal divider's draggable area
  const dividerHeight = 8; // Or parse from CSS var if needed, but 8px is default from panel-divider.css
  let newHeight = initialRulesSectionPixelHeight + deltaY;
  newHeight = Math.max(minRulesSectionHeight, Math.min(totalHeight - minResultsSectionHeight - dividerHeight, newHeight));
  rulesSectionHeight.value = newHeight;
};

const stopVerticalResize = () => {
  isVerticalDragging = false;
  document.removeEventListener('mousemove', handleVerticalMouseMove);
  document.removeEventListener('mouseup', stopVerticalResize);
  document.body.style.cursor = 'default';
};


onMounted(() => {
  loadRules();
  // 初始化时设置一个合理的下半区高度
  nextTick(() => {
    if (layoutContainerRef.value) {
      // --panel-divider-width (8px) is the height of the horizontal divider's draggable area
      const dividerHeight = 8;
      const calculatedResultsHeight = layoutContainerRef.value.clientHeight - rulesSectionHeight.value - dividerHeight;
      if (calculatedResultsHeight < minResultsSectionHeight) {
         rulesSectionHeight.value = layoutContainerRef.value.clientHeight - minResultsSectionHeight - dividerHeight;
       }
    }
  });
});
</script>

<template>
  <div class="flex flex-col h-full bg-white dark:bg-[#1e1e2e] border border-gray-200 dark:border-gray-700 rounded-lg shadow-sm overflow-hidden" ref="layoutContainerRef">
    <!-- 通知提示 - 使用 Teleport 确保正确的层级和位置 -->
    <Teleport to="body">
      <div v-if="notification.show" class="notification" :class="notification.type">
        {{ notification.message }}
      </div>
    </Teleport>

    <div class="flex flex-col bg-white dark:bg-[#1e1e2e] border-b border-gray-200 dark:border-gray-700 p-4" :style="{ height: rulesSectionHeight + 'px' }">
      <div class="flex items-center justify-between mb-3 pb-3 border-b border-gray-200 dark:border-gray-700">
        <h3 class="text-base font-medium text-gray-800 dark:text-gray-200 flex items-center gap-2">{{ t('modules.plugins.authorization_checker.rules_section') }}</h3>
        <div class="flex items-center gap-2">
          <button class="btn btn-sm btn-secondary" @click="openFilterConfigDialog()" :title="t('modules.plugins.authorization_checker.filter_config')">
            <i class="bx bx-filter"></i> {{ t('modules.plugins.authorization_checker.filter_config') }}
          </button>
          <button class="btn btn-sm btn-warning" @click="openRuleDialog()">
            <i class="bx bx-plus"></i> {{ t('modules.plugins.authorization_checker.add_rule') }}
          </button>
          <button class="btn btn-sm" :class="{ 'btn-danger': isRunning, 'btn-success': !isRunning }"
            @click="toggleCheck" :disabled="isLoading">
            <i class="bx" :class="isRunning ? 'bx-stop' : 'bx-play'"></i>
            {{ isRunning ? t('modules.plugins.authorization_checker.stop_detection') : t('modules.plugins.authorization_checker.start_detection') }}
          </button>
        </div>
      </div>

      <div v-if="isLoading" class="flex flex-col items-center justify-center flex-1 gap-3 text-gray-500 dark:text-gray-400">
        <div class="w-8 h-8 border-2 border-gray-300 dark:border-gray-600 border-t-blue-500 rounded-full animate-spin"></div>
        <span class="text-sm">{{ rules.length === 0 ? t('modules.plugins.authorization_checker.loading_rules') : t('modules.plugins.authorization_checker.refreshing_rules') }}</span>
      </div>
      <div v-else-if="rules.length === 0" class="flex flex-col items-center justify-center flex-1 gap-3 text-gray-400 dark:text-gray-500">
        <div class="text-4xl opacity-50"><i class="bx bx-file-blank"></i></div>
        <div class="text-sm">{{ t('modules.plugins.authorization_checker.no_rules') }}</div>
      </div>
      <div v-else class="flex-1 overflow-hidden">
        <HttpTrafficTable
          :items="transformedRules"
          :selectedItem="selectedRule"
          :customColumns="authorizationColumns"
          :tableClass="'compact-table'"
          :containerHeight="'400px'"
          tableId="authorization-rules-table"
          @select-item="handleRuleSelect"
        />
      </div>
    </div>

    <div class="panel-divider-horizontal" @mousedown="startVerticalResize"></div>

    <div class="flex flex-row bg-white dark:bg-[#1e1e2e]" ref="resultsSectionRef"
      :style="{ height: `calc(100% - ${rulesSectionHeight}px - var(--panel-divider-width, 8px))` }">
      <div class="flex flex-col border-r border-gray-200 dark:border-gray-700 overflow-hidden" :style="{ width: resultsListWidth + '%' }">
        <div class="flex items-center justify-between p-3 border-b border-gray-200 dark:border-gray-700 gap-2 flex-wrap">
          <h3 class="text-base font-medium text-gray-800 dark:text-gray-200 whitespace-nowrap">{{ t('modules.plugins.authorization_checker.test_results') }} ({{ filteredResults.length }})</h3>
          <!-- 增加过滤配置提示 -->
          <div class="flex items-center gap-1 text-xs text-blue-600 dark:text-blue-400" v-if="authorizationRules.filterCondition.onlyInScope && authorizationRules.filterCondition.includeUrls.length > 0">
            <i class="bx bx-info-circle"></i>
            <span class="truncate max-w-[200px]">{{ t('modules.plugins.authorization_checker.detection_scope_active', { urls: authorizationRules.filterCondition.includeUrls.join(', ') }) }}</span>
          </div>
          <div class="flex items-center gap-1 text-xs text-green-600 dark:text-green-400" v-else-if="!authorizationRules.filterCondition.onlyInScope">
            <i class="bx bx-check-circle"></i>
            <span>{{ t('modules.plugins.authorization_checker.detection_scope_all') }}</span>
          </div>
          <div class="flex items-center gap-2 ml-auto">
            <div class="search-box">
              <input type="text" v-model="resultFilter" :placeholder="t('modules.plugins.authorization_checker.search_placeholder')" class="search-input" />
              <i class="bx bx-search search-icon"></i>
            </div>
            <button class="btn btn-secondary btn-sm" @click="clearResults" :disabled="!testResults.length || isLoading"
              :title="t('modules.plugins.authorization_checker.clear_results')">
              <i class="bx bx-trash"></i>
            </button>
            <button class="btn btn-secondary btn-sm refresh" @click="fetchTestResults" :disabled="isLoading"
              :title="t('modules.plugins.authorization_checker.refresh_results')">
              <i class="bx bx-refresh"></i>
            </button>
          </div>
        </div>

        <div class="flex-1 overflow-hidden">
          <div v-if="isLoading" class="flex flex-col items-center justify-center h-full gap-3 text-gray-500 dark:text-gray-400">
            <div class="w-8 h-8 border-2 border-gray-300 dark:border-gray-600 border-t-blue-500 rounded-full animate-spin"></div>
            <span class="text-sm">{{ t('modules.plugins.authorization_checker.loading_results') }}</span>
          </div>
          <div v-else-if="!filteredResults.length" class="flex flex-col items-center justify-center h-full gap-3 text-gray-400 dark:text-gray-500">
            <div class="text-4xl opacity-50"><i class="bx bx-data"></i></div>
            <div class="text-sm">{{ t('modules.plugins.authorization_checker.no_results') }}</div>
          </div>
          <div v-else class="http-traffic-table-container h-full">
            <HttpTrafficTable
              container-height="100%"
              :items="adaptedResults"
              :selected-item="adaptedResult"
              :custom-columns="customColumns"
              tableId="auth-test-results"
              :tableClass="'compact-table'"
              :key="'auth-table-' + tableRenderKey"
              @select-item="handleItemSelect"
            />
          </div>
        </div>
      </div>

      <div class="panel-divider" @mousedown="startHorizontalResize"></div>

      <div class="flex flex-col overflow-hidden"
        :style="{ width: `calc(100% - ${resultsListWidth}% - var(--panel-divider-width, 8px))` }">
        <div v-if="!selectedResult" class="flex flex-col items-center justify-center h-full gap-3 text-gray-400 dark:text-gray-500">
          <div class="text-4xl opacity-50"><i class="bx bx-info-circle"></i></div>
          <div class="text-sm">{{ t('modules.plugins.authorization_checker.select_result') }}</div>
        </div>
        <div v-else class="flex flex-col h-full">
          <div class="flex items-center gap-2 p-3 border-b border-gray-200 dark:border-gray-700">
            <button :class="['btn', 'btn-sm', currentDetailView === 'original' ? 'btn-primary' : 'btn-secondary']"
              @click="currentDetailView = 'original'">
              {{ t('modules.plugins.authorization_checker.original_request') }}
            </button>
            <button :class="['btn', 'btn-sm', currentDetailView === 'modified' ? 'btn-primary' : 'btn-secondary']"
              @click="currentDetailView = 'modified'">
              {{ t('modules.plugins.authorization_checker.modified_request') }}
            </button>
          </div>
          <div class="flex-1 overflow-hidden">
            <RequestResponsePanel v-if="requestResponsePanelProps" v-bind="requestResponsePanelProps" />
          </div>
        </div>
      </div>
    </div>

    <div v-if="showAddRuleDialog" class="rule-dialog-overlay">
      <div class="rule-dialog">
        <div class="dialog-header">
          <h3>{{ editingRuleIndex >= 0 ? t('modules.plugins.authorization_checker.edit_rule') : t('modules.plugins.authorization_checker.add_rule') }}</h3>
          <button class="dialog-close" @click="showAddRuleDialog = false">
            <i class="bx bx-x"></i>
          </button>
        </div>
        <div class="dialog-body">
          <div class="form-group">
            <label for="rule-type">{{ t('modules.plugins.authorization_checker.operation_type') }}</label>
            <select id="rule-type" v-model="currentRule.type">
              <option value="replace">{{ t('modules.plugins.authorization_checker.replace_operation') }}</option>
              <option value="remove">{{ t('modules.plugins.authorization_checker.remove_operation') }}</option>
            </select>
          </div>
          <div class="form-group">
            <label for="header-name">{{ t('modules.plugins.authorization_checker.header_name') }}</label>
            <input id="header-name" type="text" v-model="currentRule.headerName"
              :placeholder="t('modules.plugins.authorization_checker.header_name_placeholder')">
          </div>
          <div class="form-group">
            <label for="original-value">{{ t('modules.plugins.authorization_checker.original_value') }}</label>
            <input id="original-value" type="text" v-model="currentRule.originalValue" :placeholder="t('modules.plugins.authorization_checker.original_value_placeholder')">
          </div>
          <div v-if="currentRule.type === 'replace'" class="form-group">
            <label for="new-value">{{ t('modules.plugins.authorization_checker.new_value') }}</label>
            <input id="new-value" type="text" v-model="currentRule.newValue" :placeholder="t('modules.plugins.authorization_checker.new_value_placeholder')">
          </div>
          <div class="form-group">
            <label for="description">{{ t('modules.plugins.authorization_checker.description') }}</label>
            <input id="description" type="text" v-model="currentRule.description" :placeholder="t('modules.plugins.authorization_checker.description_placeholder')">
          </div>
          <div class="form-group">
            <label class="checkbox-label">
              <input type="checkbox" v-model="currentRule.enabled">
              <span>{{ t('modules.plugins.authorization_checker.enable_rule') }}</span>
            </label>
          </div>
        </div>
        <div class="dialog-footer">
          <button class="btn btn-secondary" @click="showAddRuleDialog = false">{{ t('common.actions.cancel') }}</button>
          <button class="btn btn-primary" @click="saveRule">{{ t('common.actions.save') }}</button>
        </div>
      </div>
    </div>

    <!-- 过滤配置弹窗 -->
    <div v-if="showFilterConfigDialog" class="rule-dialog-overlay">
      <div class="rule-dialog filter-config-dialog">
        <div class="dialog-header">
          <h3>{{ t('modules.plugins.authorization_checker.filter_config') }}</h3>
          <button class="dialog-close" @click="showFilterConfigDialog = false">
            <i class="bx bx-x"></i>
          </button>
        </div>
        <div class="dialog-body">
          <div class="form-group">
            <label class="checkbox-label">
              <input type="checkbox" v-model="filterCondition.onlyInScope">
              <span>{{ t('modules.plugins.authorization_checker.filter_scope_only') }}</span>
            </label>
            <small class="form-help">{{ t('modules.plugins.authorization_checker.filter_scope_help') }}</small>
          </div>

          <div class="form-group">
            <label for="include-urls">{{ t('modules.plugins.authorization_checker.include_urls') }}</label>
            <textarea 
              spellcheck="false"
              id="include-urls" 
              v-model="includeUrlsText" 
              :placeholder="t('modules.plugins.authorization_checker.include_urls_placeholder')"
              rows="4"
            ></textarea>
            <small class="form-help">{{ t('modules.plugins.authorization_checker.include_urls_help') }}</small>
          </div>

          <div class="form-group">
            <label for="exclude-urls">{{ t('modules.plugins.authorization_checker.exclude_urls') }}</label>
            <textarea 
              spellcheck="false"
              id="exclude-urls" 
              v-model="excludeUrlsText" 
              :placeholder="t('modules.plugins.authorization_checker.exclude_urls_placeholder')"
              rows="4"
            ></textarea>
            <small class="form-help">{{ t('modules.plugins.authorization_checker.exclude_urls_help') }}</small>
          </div>

          <div class="form-group">
            <label for="include-file-types">{{ t('modules.plugins.authorization_checker.include_file_types') }}</label>
            <input 
              spellcheck="false"
              id="include-file-types" 
              type="text" 
              v-model="includeFileTypesText" 
              :placeholder="t('modules.plugins.authorization_checker.include_file_types_placeholder')"
            >
            <small class="form-help">{{ t('modules.plugins.authorization_checker.include_file_types_help') }}</small>
          </div>

          <div class="form-group">
            <label for="exclude-file-types">{{ t('modules.plugins.authorization_checker.exclude_file_types') }}</label>
            <input 
              spellcheck="false"
              id="exclude-file-types" 
              type="text" 
              v-model="excludeFileTypesText" 
              :placeholder="t('modules.plugins.authorization_checker.exclude_file_types_placeholder')"
            >
            <small class="form-help">{{ t('modules.plugins.authorization_checker.exclude_file_types_help') }}</small>
          </div>
        </div>
        <div class="dialog-footer">
          <button class="btn btn-secondary" @click="showFilterConfigDialog = false">{{ t('common.actions.cancel') }}</button>
          <button class="btn btn-primary" @click="saveFilterConfig">{{ t('modules.plugins.authorization_checker.save_config') }}</button>
        </div>
      </div>
    </div>
  </div>
</template>

<!-- 样式已迁移到 styles/modules/plugins.css 统一管理 -->
