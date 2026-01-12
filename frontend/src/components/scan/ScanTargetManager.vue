<script setup lang="ts">
import { ref, reactive, computed, onMounted, nextTick } from 'vue';
import type { 
  ScanTarget, 
  ScanTargetStatus, 
  ScanTargetType, 
  ScanConfig,
  ScanStatistics,
  AddScanTargetParams,
  BatchAddScanTargetsParams,
  SchedulerStatus
} from '../../types/scanTarget';
import {
  STATUS_OPTIONS,
  TYPE_OPTIONS,
  PRIORITY_OPTIONS,
  SCHEDULE_TYPE_OPTIONS,
  STATUS_COLOR_MAP,
  STATUS_TEXT_MAP,
  TYPE_TEXT_MAP
} from '../../types/scanTarget';

// 响应式数据
const targets = ref<ScanTarget[]>([]);
const total = ref(0);
const loading = ref(false);
const statistics = ref<ScanStatistics | null>(null);
const schedulerStatus = ref<SchedulerStatus | null>(null);

// 查询参数
const queryParams = reactive({
  status: '' as ScanTargetStatus | '',
  limit: 20,
  offset: 0,
  search: '',
  type: '' as ScanTargetType | ''
});

// 添加目标表单
const addForm = reactive<AddScanTargetParams>({
  name: '',
  type: 'single',
  target: '',
  description: '',
  priority: 5,
  schedule_type: 'once',
  schedule_time: '',
  created_by: '本地用户'
});

// 批量添加表单
const batchForm = reactive<BatchAddScanTargetsParams>({
  targets: [],
  target_type: 'single',
  created_by: '本地用户'
});

// 弹框状态
const showAddDialog = ref(false);
const showBatchDialog = ref(false);
const showConfigDialog = ref(false);
const showDetailDialog = ref(false);

// 当前选中的目标
const selectedTarget = ref<ScanTarget | null>(null);

// 扫描配置
const scanConfig = ref<ScanConfig | null>(null);

// 批量目标文本
const batchTargetsText = ref('');

// API调用函数（需要根据实际API接口调整）
declare global {
  interface Window {
    go: {
      main: {
        App: {
          GetScanTargets: (status: string, limit: number, offset: number) => Promise<{data?: {targets: ScanTarget[], total: number}, error?: string}>;
          AddScanTarget: (target: AddScanTargetParams) => Promise<{data?: string, error?: string}>;
          BatchAddScanTargets: (targets: string[], targetType: string, createdBy: string) => Promise<{data?: string, error?: string}>;
          UpdateScanTarget: (target: ScanTarget) => Promise<{data?: string, error?: string}>;
          DeleteScanTarget: (id: number) => Promise<{data?: string, error?: string}>;
          UpdateScanTargetStatus: (id: number, status: string, message: string) => Promise<{data?: string, error?: string}>;
          GetScanStatistics: () => Promise<{data?: ScanStatistics, error?: string}>;
          GetDefaultScanConfig: () => Promise<{data?: ScanConfig, error?: string}>;
          GetSchedulerStatus: () => Promise<{data?: SchedulerStatus, error?: string}>;
          StartScheduler: () => Promise<{data?: string, error?: string}>;
          StopScheduler: () => Promise<{data?: string, error?: string}>;
        };
      };
    };
  }
}

// 计算属性
const filteredTargets = computed(() => {
  return targets.value.filter(target => {
    if (queryParams.search && !target.name.toLowerCase().includes(queryParams.search.toLowerCase()) && 
        !target.target.toLowerCase().includes(queryParams.search.toLowerCase())) {
      return false;
    }
    if (queryParams.status && target.status !== queryParams.status) {
      return false;
    }
    if (queryParams.type && target.type !== queryParams.type) {
      return false;
    }
    return true;
  });
});

const paginatedTargets = computed(() => {
  const start = queryParams.offset;
  const end = start + queryParams.limit;
  return filteredTargets.value.slice(start, end);
});

// 方法
const loadTargets = async () => {
  loading.value = true;
  try {
    const result = await window.go.main.App.GetScanTargets(
      queryParams.status, 
      queryParams.limit, 
      queryParams.offset
    );
    
    if (result.error) {
      console.error('加载扫描目标失败:', result.error);
      return;
    }
    
    if (result.data) {
      targets.value = result.data.targets || [];
      total.value = result.data.total || 0;
    }
  } catch (error) {
    console.error('加载扫描目标失败:', error);
  } finally {
    loading.value = false;
  }
};

const loadStatistics = async () => {
  try {
    const result = await window.go.main.App.GetScanStatistics();
    if (result.data) {
      statistics.value = result.data;
    }
  } catch (error) {
    console.error('加载统计信息失败:', error);
  }
};

const loadSchedulerStatus = async () => {
  try {
    const result = await window.go.main.App.GetSchedulerStatus();
    if (result.data) {
      schedulerStatus.value = result.data;
    }
  } catch (error) {
    console.error('加载调度器状态失败:', error);
  }
};

const loadDefaultConfig = async () => {
  try {
    const result = await window.go.main.App.GetDefaultScanConfig();
    if (result.data) {
      scanConfig.value = result.data;
    }
  } catch (error) {
    console.error('加载默认配置失败:', error);
  }
};

const addTarget = async () => {
  try {
    const result = await window.go.main.App.AddScanTarget(addForm);
    if (result.error) {
      alert('添加失败: ' + result.error);
      return;
    }
    
    showAddDialog.value = false;
    resetAddForm();
    await loadTargets();
    await loadStatistics();
  } catch (error) {
    console.error('添加目标失败:', error);
    alert('添加失败: ' + error);
  }
};

const batchAddTargets = async () => {
  const targetList = batchTargetsText.value.split('\n')
    .map(line => line.trim())
    .filter(line => line.length > 0);
  
  if (targetList.length === 0) {
    alert('请输入目标列表');
    return;
  }
  
  try {
    const result = await window.go.main.App.BatchAddScanTargets(
      targetList,
      batchForm.target_type,
      batchForm.created_by
    );
    
    if (result.error) {
      alert('批量添加失败: ' + result.error);
      return;
    }
    
    showBatchDialog.value = false;
    resetBatchForm();
    await loadTargets();
    await loadStatistics();
  } catch (error) {
    console.error('批量添加失败:', error);
    alert('批量添加失败: ' + error);
  }
};

const deleteTarget = async (target: ScanTarget) => {
  if (!confirm(`确定要删除扫描目标 "${target.name}" 吗？`)) {
    return;
  }
  
  try {
    const result = await window.go.main.App.DeleteScanTarget(target.id);
    if (result.error) {
      alert('删除失败: ' + result.error);
      return;
    }
    
    await loadTargets();
    await loadStatistics();
  } catch (error) {
    console.error('删除目标失败:', error);
    alert('删除失败: ' + error);
  }
};

const updateTargetStatus = async (target: ScanTarget, status: ScanTargetStatus) => {
  try {
    const result = await window.go.main.App.UpdateScanTargetStatus(target.id, status, '');
    if (result.error) {
      alert('更新状态失败: ' + result.error);
      return;
    }
    
    await loadTargets();
  } catch (error) {
    console.error('更新状态失败:', error);
    alert('更新状态失败: ' + error);
  }
};

const startScheduler = async () => {
  try {
    const result = await window.go.main.App.StartScheduler();
    if (result.error) {
      alert('启动调度器失败: ' + result.error);
      return;
    }
    
    await loadSchedulerStatus();
  } catch (error) {
    console.error('启动调度器失败:', error);
    alert('启动调度器失败: ' + error);
  }
};

const stopScheduler = async () => {
  try {
    const result = await window.go.main.App.StopScheduler();
    if (result.error) {
      alert('停止调度器失败: ' + result.error);
      return;
    }
    
    await loadSchedulerStatus();
  } catch (error) {
    console.error('停止调度器失败:', error);
    alert('停止调度器失败: ' + error);
  }
};

const showTargetDetail = (target: ScanTarget) => {
  selectedTarget.value = target;
  showDetailDialog.value = true;
};

const resetAddForm = () => {
  Object.assign(addForm, {
    name: '',
    type: 'single',
    target: '',
    description: '',
    priority: 5,
    schedule_type: 'once',
    schedule_time: '',
    created_by: '本地用户'
  });
};

const resetBatchForm = () => {
  Object.assign(batchForm, {
    targets: [],
    target_type: 'single',
    created_by: '本地用户'
  });
  batchTargetsText.value = '';
};

const formatDuration = (seconds: number): string => {
  if (seconds < 60) return `${seconds}秒`;
  if (seconds < 3600) return `${Math.floor(seconds / 60)}分${seconds % 60}秒`;
  return `${Math.floor(seconds / 3600)}时${Math.floor((seconds % 3600) / 60)}分`;
};

const formatDateTime = (dateStr: string): string => {
  if (!dateStr) return '-';
  return new Date(dateStr).toLocaleString('zh-CN');
};

const getStatusColor = (status: ScanTargetStatus): string => {
  return STATUS_COLOR_MAP[status] || 'gray';
};

const getStatusText = (status: ScanTargetStatus): string => {
  return STATUS_TEXT_MAP[status] || status;
};

const getTypeText = (type: ScanTargetType): string => {
  return TYPE_TEXT_MAP[type] || type;
};

// 生命周期
onMounted(async () => {
  await Promise.all([
    loadTargets(),
    loadStatistics(),
    loadSchedulerStatus(),
    loadDefaultConfig()
  ]);
  
  // 定期刷新调度器状态
  setInterval(loadSchedulerStatus, 5000);
});
</script>

<template>
  <div class="scan-target-manager h-full flex flex-col bg-white dark:bg-gray-900">
    <!-- 头部工具栏 -->
    <div class="header-toolbar border-b border-gray-200 dark:border-gray-700 p-4">
      <div class="flex items-center justify-between">
        <div class="flex items-center space-x-4">
          <h2 class="text-lg font-semibold text-gray-900 dark:text-white">扫描目标管理</h2>
          
          <!-- 统计信息 -->
          <div v-if="statistics" class="flex items-center space-x-4 text-sm text-gray-600 dark:text-gray-300">
            <span>总计: {{ statistics.total }}</span>
            <span>今日: {{ statistics.today }}</span>
          </div>
        </div>
        
        <div class="flex items-center space-x-3">
          <!-- 调度器状态 -->
          <div v-if="schedulerStatus" class="flex items-center space-x-2">
            <div class="flex items-center">
              <div 
                :class="[
                  'w-2 h-2 rounded-full mr-2',
                  schedulerStatus.running ? 'bg-green-500' : 'bg-red-500'
                ]"
              ></div>
              <span class="text-sm text-gray-600 dark:text-gray-300">
                调度器: {{ schedulerStatus.running ? '运行中' : '已停止' }}
              </span>
            </div>
            <button
              @click="schedulerStatus.running ? stopScheduler() : startScheduler()"
              class="px-3 py-1 text-xs rounded"
              :class="[
                schedulerStatus.running 
                  ? 'bg-red-100 text-red-800 hover:bg-red-200 dark:bg-red-900 dark:text-red-200' 
                  : 'bg-green-100 text-green-800 hover:bg-green-200 dark:bg-green-900 dark:text-green-200'
              ]"
            >
              {{ schedulerStatus.running ? '停止' : '启动' }}
            </button>
          </div>
          
          <!-- 操作按钮 -->
          <button
            @click="showAddDialog = true"
            class="btn btn-primary"
          >
            <i class="bx bx-plus mr-1"></i>
            添加目标
          </button>
          
          <button
            @click="showBatchDialog = true"
            class="btn btn-secondary"
          >
            <i class="bx bx-list-plus mr-1"></i>
            批量添加
          </button>
          
          <button
            @click="loadTargets()"
            class="btn btn-secondary"
            :disabled="loading"
          >
            <i class="bx bx-refresh" :class="{ 'bx-spin': loading }"></i>
          </button>
        </div>
      </div>
      
      <!-- 筛选条件 -->
      <div class="mt-4 flex items-center space-x-4">
        <div class="flex items-center space-x-2">
          <label class="text-sm text-gray-600 dark:text-gray-300">状态:</label>
          <select
            v-model="queryParams.status"
            @change="loadTargets()"
            class="form-select"
          >
            <option value="">全部</option>
            <option v-for="option in STATUS_OPTIONS" :key="option.value" :value="option.value">
              {{ option.label }}
            </option>
          </select>
        </div>
        
        <div class="flex items-center space-x-2">
          <label class="text-sm text-gray-600 dark:text-gray-300">类型:</label>
          <select
            v-model="queryParams.type"
            @change="loadTargets()"
            class="form-select"
          >
            <option value="">全部</option>
            <option v-for="option in TYPE_OPTIONS" :key="option.value" :value="option.value">
              {{ option.label }}
            </option>
          </select>
        </div>
        
        <div class="flex items-center space-x-2 flex-1">
          <label class="text-sm text-gray-600 dark:text-gray-300">搜索:</label>
          <input
            v-model="queryParams.search"
            @input="loadTargets()"
            type="text"
            placeholder="搜索目标名称或地址..."
            class="form-input flex-1"
          >
        </div>
      </div>
    </div>

    <!-- 目标列表 -->
    <div class="flex-1 overflow-auto">
      <div v-if="loading" class="flex items-center justify-center h-32">
        <i class="bx bx-loader-alt bx-spin text-2xl text-gray-400"></i>
      </div>
      
      <div v-else-if="paginatedTargets.length === 0" class="empty-state">
        <div class="text-center py-12">
          <i class="bx bx-target-lock text-4xl text-gray-400 mb-4"></i>
          <p class="text-gray-500 dark:text-gray-400 mb-4">暂无扫描目标</p>
          <button @click="showAddDialog = true" class="btn btn-primary">
            添加第一个目标
          </button>
        </div>
      </div>
      
      <div v-else class="divide-y divide-gray-200 dark:divide-gray-700">
        <div
          v-for="target in paginatedTargets"
          :key="target.id"
          class="p-4 hover:bg-gray-50 dark:hover:bg-gray-800 cursor-pointer"
          @click="showTargetDetail(target)"
        >
          <div class="flex items-start justify-between">
            <div class="flex-1 min-w-0">
              <!-- 第一行：名称、状态、类型 -->
              <div class="flex items-center mb-2">
                <h3 class="text-base font-medium text-gray-900 dark:text-white truncate mr-3">
                  {{ target.name }}
                </h3>
                
                <span
                  class="inline-flex items-center px-2 py-1 rounded-full text-xs font-medium mr-2"
                  :class="`bg-${getStatusColor(target.status)}-100 text-${getStatusColor(target.status)}-800 dark:bg-${getStatusColor(target.status)}-900 dark:text-${getStatusColor(target.status)}-200`"
                >
                  {{ getStatusText(target.status) }}
                </span>
                
                <span class="inline-flex items-center px-2 py-1 rounded-full text-xs bg-gray-100 text-gray-800 dark:bg-gray-700 dark:text-gray-200">
                  {{ getTypeText(target.type) }}
                </span>
                
                <span v-if="target.priority >= 8" class="inline-flex items-center px-2 py-1 rounded-full text-xs bg-red-100 text-red-800 dark:bg-red-900 dark:text-red-200 ml-2">
                  高优先级
                </span>
              </div>
              
              <!-- 第二行：目标地址 -->
              <p class="text-sm text-gray-600 dark:text-gray-300 mb-2 truncate">
                <i class="bx bx-link-external mr-1"></i>
                {{ target.target }}
              </p>
              
              <!-- 第三行：描述和节点信息 -->
              <div class="flex items-center text-xs text-gray-500 dark:text-gray-400 space-x-4">
                <span v-if="target.description">{{ target.description }}</span>
                <span v-if="target.node_name">
                  <i class="bx bx-server mr-1"></i>
                  {{ target.node_name }}
                </span>
                <span>
                  <i class="bx bx-time mr-1"></i>
                  {{ formatDateTime(target.created_at) }}
                </span>
              </div>
              
              <!-- 进度条（运行中时显示） -->
              <div v-if="target.status === 'running'" class="mt-2">
                <div class="flex items-center justify-between text-xs text-gray-600 dark:text-gray-300 mb-1">
                  <span>扫描进度</span>
                  <span>{{ target.progress }}%</span>
                </div>
                <div class="w-full bg-gray-200 rounded-full h-2 dark:bg-gray-700">
                  <div
                    class="bg-blue-600 h-2 rounded-full transition-all duration-300"
                    :style="{ width: target.progress + '%' }"
                  ></div>
                </div>
              </div>
              
              <!-- 统计信息（已完成时显示） -->
              <div v-if="target.status === 'completed'" class="mt-2 grid grid-cols-4 gap-4 text-xs">
                <div class="text-center">
                  <div class="text-gray-500 dark:text-gray-400">主机</div>
                  <div class="font-medium text-gray-900 dark:text-white">{{ target.found_hosts }}</div>
                </div>
                <div class="text-center">
                  <div class="text-gray-500 dark:text-gray-400">端口</div>
                  <div class="font-medium text-gray-900 dark:text-white">{{ target.found_ports }}</div>
                </div>
                <div class="text-center">
                  <div class="text-gray-500 dark:text-gray-400">漏洞</div>
                  <div class="font-medium text-red-600 dark:text-red-400">{{ target.found_vulns }}</div>
                </div>
                <div class="text-center">
                  <div class="text-gray-500 dark:text-gray-400">目录</div>
                  <div class="font-medium text-gray-900 dark:text-white">{{ target.found_dirs }}</div>
                </div>
              </div>
            </div>
            
            <!-- 操作按钮 -->
            <div class="flex items-center space-x-2 ml-4" @click.stop>
              <button
                v-if="target.status === 'pending'"
                @click="updateTargetStatus(target, 'running')"
                class="btn-icon btn-icon-sm text-green-600 hover:bg-green-100 dark:hover:bg-green-900"
                title="开始扫描"
              >
                <i class="bx bx-play"></i>
              </button>
              
              <button
                v-if="target.status === 'running'"
                @click="updateTargetStatus(target, 'paused')"
                class="btn-icon btn-icon-sm text-yellow-600 hover:bg-yellow-100 dark:hover:bg-yellow-900"
                title="暂停扫描"
              >
                <i class="bx bx-pause"></i>
              </button>
              
              <button
                v-if="target.status === 'failed'"
                @click="updateTargetStatus(target, 'pending')"
                class="btn-icon btn-icon-sm text-blue-600 hover:bg-blue-100 dark:hover:bg-blue-900"
                title="重新扫描"
              >
                <i class="bx bx-refresh"></i>
              </button>
              
              <button
                @click="deleteTarget(target)"
                class="btn-icon btn-icon-sm text-red-600 hover:bg-red-100 dark:hover:bg-red-900"
                title="删除"
              >
                <i class="bx bx-trash"></i>
              </button>
            </div>
          </div>
        </div>
      </div>
    </div>

    <!-- 分页 -->
    <div v-if="total > queryParams.limit" class="border-t border-gray-200 dark:border-gray-700 px-4 py-3">
      <div class="flex items-center justify-between">
        <div class="text-sm text-gray-700 dark:text-gray-300">
          显示 {{ queryParams.offset + 1 }} 到 {{ Math.min(queryParams.offset + queryParams.limit, total) }} 项，共 {{ total }} 项
        </div>
        <div class="flex items-center space-x-2">
          <button
            @click="queryParams.offset = Math.max(0, queryParams.offset - queryParams.limit); loadTargets()"
            :disabled="queryParams.offset === 0"
            class="btn btn-secondary btn-sm"
          >
            上一页
          </button>
          <button
            @click="queryParams.offset += queryParams.limit; loadTargets()"
            :disabled="queryParams.offset + queryParams.limit >= total"
            class="btn btn-secondary btn-sm"
          >
            下一页
          </button>
        </div>
      </div>
    </div>

    <!-- 添加目标对话框 -->
    <div v-if="showAddDialog" class="dialog-overlay" @click.self="showAddDialog = false">
      <div class="dialog">
        <div class="dialog-header">
          <h3>添加扫描目标</h3>
          <button @click="showAddDialog = false" class="btn-icon">
            <i class="bx bx-x"></i>
          </button>
        </div>
        
        <div class="dialog-body">
          <form @submit.prevent="addTarget" class="space-y-4">
            <div class="grid grid-cols-2 gap-4">
              <div>
                <label class="form-label">目标名称 *</label>
                <input v-model="addForm.name" type="text" required class="form-input" placeholder="输入目标名称">
              </div>
              <div>
                <label class="form-label">目标类型 *</label>
                <select v-model="addForm.type" required class="form-select">
                  <option v-for="option in TYPE_OPTIONS" :key="option.value" :value="option.value">
                    {{ option.label }}
                  </option>
                </select>
              </div>
            </div>
            
            <div>
              <label class="form-label">目标地址 *</label>
              <input v-model="addForm.target" type="text" required class="form-input" 
                     placeholder="输入URL、域名、IP地址或CIDR">
            </div>
            
            <div>
              <label class="form-label">描述</label>
              <textarea spellcheck="false" v-model="addForm.description" class="form-input" rows="3" 
                        placeholder="输入目标描述"></textarea>
            </div>
            
            <div class="grid grid-cols-2 gap-4">
              <div>
                <label class="form-label">优先级</label>
                <select v-model="addForm.priority" class="form-select">
                  <option v-for="option in PRIORITY_OPTIONS" :key="option.value" :value="option.value">
                    {{ option.label }}
                  </option>
                </select>
              </div>
              <div>
                <label class="form-label">调度类型</label>
                <select v-model="addForm.schedule_type" class="form-select">
                  <option v-for="option in SCHEDULE_TYPE_OPTIONS" :key="option.value" :value="option.value">
                    {{ option.label }}
                  </option>
                </select>
              </div>
            </div>
            
            <div class="dialog-footer">
              <button type="button" @click="showAddDialog = false" class="btn btn-secondary">
                取消
              </button>
              <button type="submit" class="btn btn-primary">
                添加
              </button>
            </div>
          </form>
        </div>
      </div>
    </div>

    <!-- 批量添加对话框 -->
    <div v-if="showBatchDialog" class="dialog-overlay" @click.self="showBatchDialog = false">
      <div class="dialog">
        <div class="dialog-header">
          <h3>批量添加扫描目标</h3>
          <button @click="showBatchDialog = false" class="btn-icon">
            <i class="bx bx-x"></i>
          </button>
        </div>
        
        <div class="dialog-body">
          <form @submit.prevent="batchAddTargets" class="space-y-4">
            <div>
              <label class="form-label">目标类型 *</label>
              <select v-model="batchForm.target_type" required class="form-select">
                <option v-for="option in TYPE_OPTIONS" :key="option.value" :value="option.value">
                  {{ option.label }}
                </option>
              </select>
            </div>
            
            <div>
              <label class="form-label">目标列表 *</label>
              <textarea spellcheck="false"
                v-model="batchTargetsText"
                required
                class="form-input"
                rows="10"
                placeholder="每行一个目标，例如：&#10;https://example.com&#10;192.168.1.1&#10;10.0.0.0/24"
              ></textarea>
              <p class="text-xs text-gray-500 mt-1">每行输入一个目标地址</p>
            </div>
            
            <div class="dialog-footer">
              <button type="button" @click="showBatchDialog = false" class="btn btn-secondary">
                取消
              </button>
              <button type="submit" class="btn btn-primary">
                批量添加
              </button>
            </div>
          </form>
        </div>
      </div>
    </div>

    <!-- 目标详情对话框 -->
    <div v-if="showDetailDialog && selectedTarget" class="dialog-overlay" @click.self="showDetailDialog = false">
      <div class="dialog dialog-lg">
        <div class="dialog-header">
          <h3>扫描目标详情</h3>
          <button @click="showDetailDialog = false" class="btn-icon">
            <i class="bx bx-x"></i>
          </button>
        </div>
        
        <div class="dialog-body">
          <div class="space-y-6">
            <!-- 基本信息 -->
            <div class="grid grid-cols-2 gap-6">
              <div>
                <h4 class="text-sm font-medium text-gray-900 dark:text-white mb-3">基本信息</h4>
                <div class="space-y-2 text-sm">
                  <div class="flex justify-between">
                    <span class="text-gray-500 dark:text-gray-400">名称:</span>
                    <span class="text-gray-900 dark:text-white">{{ selectedTarget.name }}</span>
                  </div>
                  <div class="flex justify-between">
                    <span class="text-gray-500 dark:text-gray-400">类型:</span>
                    <span class="text-gray-900 dark:text-white">{{ getTypeText(selectedTarget.type) }}</span>
                  </div>
                  <div class="flex justify-between">
                    <span class="text-gray-500 dark:text-gray-400">状态:</span>
                    <span :class="`text-${getStatusColor(selectedTarget.status)}-600`">
                      {{ getStatusText(selectedTarget.status) }}
                    </span>
                  </div>
                  <div class="flex justify-between">
                    <span class="text-gray-500 dark:text-gray-400">优先级:</span>
                    <span class="text-gray-900 dark:text-white">{{ selectedTarget.priority }}</span>
                  </div>
                </div>
              </div>
              
              <div>
                <h4 class="text-sm font-medium text-gray-900 dark:text-white mb-3">执行信息</h4>
                <div class="space-y-2 text-sm">
                  <div class="flex justify-between">
                    <span class="text-gray-500 dark:text-gray-400">开始时间:</span>
                    <span class="text-gray-900 dark:text-white">{{ formatDateTime(selectedTarget.started_at || '') }}</span>
                  </div>
                  <div class="flex justify-between">
                    <span class="text-gray-500 dark:text-gray-400">完成时间:</span>
                    <span class="text-gray-900 dark:text-white">{{ formatDateTime(selectedTarget.completed_at || '') }}</span>
                  </div>
                  <div class="flex justify-between">
                    <span class="text-gray-500 dark:text-gray-400">持续时间:</span>
                    <span class="text-gray-900 dark:text-white">{{ formatDuration(selectedTarget.duration) }}</span>
                  </div>
                  <div class="flex justify-between">
                    <span class="text-gray-500 dark:text-gray-400">进度:</span>
                    <span class="text-gray-900 dark:text-white">{{ selectedTarget.progress }}%</span>
                  </div>
                </div>
              </div>
            </div>
            
            <!-- 目标地址 -->
            <div>
              <h4 class="text-sm font-medium text-gray-900 dark:text-white mb-3">目标地址</h4>
              <p class="text-sm text-gray-600 dark:text-gray-300 bg-gray-50 dark:bg-gray-800 p-3 rounded">
                {{ selectedTarget.target }}
              </p>
            </div>
            
            <!-- 扫描结果统计 -->
            <div v-if="selectedTarget.status === 'completed'">
              <h4 class="text-sm font-medium text-gray-900 dark:text-white mb-3">扫描结果</h4>
              <div class="grid grid-cols-4 gap-4">
                <div class="text-center p-3 bg-gray-50 dark:bg-gray-800 rounded">
                  <div class="text-2xl font-bold text-gray-900 dark:text-white">{{ selectedTarget.found_hosts }}</div>
                  <div class="text-xs text-gray-500 dark:text-gray-400">发现主机</div>
                </div>
                <div class="text-center p-3 bg-gray-50 dark:bg-gray-800 rounded">
                  <div class="text-2xl font-bold text-gray-900 dark:text-white">{{ selectedTarget.found_ports }}</div>
                  <div class="text-xs text-gray-500 dark:text-gray-400">开放端口</div>
                </div>
                <div class="text-center p-3 bg-gray-50 dark:bg-gray-800 rounded">
                  <div class="text-2xl font-bold text-red-600 dark:text-red-400">{{ selectedTarget.found_vulns }}</div>
                  <div class="text-xs text-gray-500 dark:text-gray-400">发现漏洞</div>
                </div>
                <div class="text-center p-3 bg-gray-50 dark:bg-gray-800 rounded">
                  <div class="text-2xl font-bold text-gray-900 dark:text-white">{{ selectedTarget.found_dirs }}</div>
                  <div class="text-xs text-gray-500 dark:text-gray-400">发现目录</div>
                </div>
              </div>
            </div>
            
            <!-- 错误信息 -->
            <div v-if="selectedTarget.error_message">
              <h4 class="text-sm font-medium text-gray-900 dark:text-white mb-3">错误信息</h4>
              <p class="text-sm text-red-600 dark:text-red-400 bg-red-50 dark:bg-red-900/20 p-3 rounded">
                {{ selectedTarget.error_message }}
              </p>
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
/* 引入已有的样式类 */
</style> 