<script setup lang="ts">
import { useI18n } from 'vue-i18n';
import { ref, onMounted, onBeforeUnmount, computed, h } from 'vue';
// @ts-ignore
import { Fuzz, FuzzStop, SetProxy } from "../../../../bindings/github.com/yhy0/ChYing/app.js";
// @ts-ignore
import { Events } from "@wailsio/runtime";
import { message } from '../../../utils/message';
import { RequestResponsePanel } from '../../common';
import HttpTrafficTable from '../../common/HttpTrafficTable.vue';
import type { HttpTrafficColumn } from '../../../types';

const { t } = useI18n();
// 定义接口
interface FuzzResult {
  id: number;
  url: string;
  method: string;
  statusCode: number;
  length: number;
  request: string;
  response: string;
}

// 表单数据
const formValue = ref({
  targetUrl: "",
  path: "",
  checkboxGroupValue: [] as string[],
});

// 代理设置
const proxyInput = ref('http://127.0.0.1:8080');
const checkedRef = ref(false);
const inputRef = ref(null);

// 扫描状态
const isScanning = ref(false);
const percentage = ref(0);
const alertType = ref("info");
const alertContent = ref("");

// 扫描结果数据
const data = ref<FuzzResult[]>([]);
const selectedItem = ref<FuzzResult | null>(null);
const requestData = ref('');
const responseData = ref('');

// 面板高度调整
const controlPanelHeight = ref(200); // 控制面板高度
const tableHeight = ref(300); // 表格高度

// 调整状态变量
let isResizingControl = false;
let isResizingTable = false;
let startY = 0;
let startControlHeight = 0;
let startTableHeight = 0;

// 选项配置
const checkboxOptions = [
  { label: 'jsp', value: 'jsp' },
  { label: 'asp', value: 'asp' },
  { label: 'aspx', value: 'aspx' },
  { label: 'php', value: 'php' },
  { label: 'BBScan', value: 'bbscan' },
  { label: 'WSDL', value: 'wsdl' },
  { label: 'Bypass 403', value: 'bypass403' },
  { label: 'Swagger', value: 'swagger' }
];

// 处理选项切换
function handleCheckboxChange(option: string, checked: boolean) {
  if (option === 'swagger' && checked) {
    // 如果选中了swagger，清除其他选项
    formValue.value.checkboxGroupValue = ['swagger'];
  } else if (option !== 'swagger' && checked) {
    // 如果选中了其他选项并且swagger已选中，需要移除swagger
    const newOptions = [...formValue.value.checkboxGroupValue];
    
    // 移除swagger (如果存在)
    const swaggerIndex = newOptions.indexOf('swagger');
    if (swaggerIndex !== -1) {
      newOptions.splice(swaggerIndex, 1);
    }
    
    // 添加新选中的选项
    if (!newOptions.includes(option)) {
      newOptions.push(option);
    }
    
    formValue.value.checkboxGroupValue = newOptions;
  } else {
    // 取消选中的情况
    formValue.value.checkboxGroupValue = formValue.value.checkboxGroupValue.filter(item => item !== option);
  }
}

// 开始扫描
function fuzz() {
  const target = formValue.value.targetUrl.toString().trim();
  if (target !== "") {
    isScanning.value = true;
    data.value = [];
    selectedItem.value = null;
    requestData.value = '';
    responseData.value = '';
    alertType.value = "info";
    alertContent.value = target + " 正在扫描中...";
    percentage.value = 0;

    message.success(target + " 开始扫描");

    Fuzz(target, formValue.value.checkboxGroupValue, formValue.value.path.toString().trim()).then((result: string) => {
      isScanning.value = false;

      if (result === "") {
        alertType.value = "success";
        // alertContent.value = target + " 扫描完成";
        // percentage.value = 100;
        message.success(target + " 扫描完成");
      } else {
        alertType.value = "error";
        alertContent.value = target + " " + result;
        message.error(target + " " + result);
      }
    }).catch((error: any) => {
      isScanning.value = false;
      alertType.value = "error";
      alertContent.value = "扫描出错: " + error;
      message.error("扫描出错: " + error);
    });
  } else {
    message.warning("请输入目标URL");
  }
}

// 停止扫描
function fuzzStop() {
  FuzzStop().then(() => {
    isScanning.value = false;
    message.success(formValue.value.targetUrl.toString().trim() + " 扫描已停止");
  }).catch((error: any) => {
    message.error("停止扫描失败: " + error);
  });
}

// 代理切换
function handleCheckedChange(checked: boolean) {
  checkedRef.value = checked;
  if (checked) {
    const proxyValue = proxyInput.value.trim();
    if (proxyValue) {
      SetProxy(proxyValue).then(() => {
        message.success("代理设置成功: " + proxyValue);
      }).catch((error: any) => {
        message.error("代理设置失败: " + error);
      });
    } else {
      message.warning("请输入有效的代理地址");
      checkedRef.value = false;
    }
  } else {
    SetProxy("").then(() => {
      message.warning("代理已关闭");
    }).catch((error: any) => {
      message.error("关闭代理失败: " + error);
    });
  }
}

// 表格行点击
function handleRowClick(row: FuzzResult) {
  selectedItem.value = row;
  requestData.value = row.request || '';
  responseData.value = row.response || '';
}

// 数据转换函数：将 FuzzResult 转换为 HttpTrafficItem 格式
const transformedResults = computed(() => {
  return data.value.map(item => ({
    id: item.id,
    url: item.url,
    method: item.method,
    status: item.statusCode,
    size: item.length,
    timestamp: Date.now(),
    // 扩展字段
    request: item.request,
    response: item.response,
    // 为了兼容 HttpTrafficItem 接口
    host: (() => {
      try {
        return new URL(item.url).hostname;
      } catch {
        return 'unknown';
      }
    })(),
    path: (() => {
      try {
        return new URL(item.url).pathname;
      } catch {
        return item.url;
      }
    })()
  }));
});

// 列定义
const fuzzScannerColumns = computed<HttpTrafficColumn<any>[]>(() => [
  {
    id: 'id',
    name: '#',
    width: 60,
    cellRenderer: ({ item }) => h('span', {
      class: 'font-mono text-xs text-gray-700 dark:text-gray-300'
    }, (item.id + 1).toString())
  },
  {
    id: 'url',
    name: 'URL',
    width: 300,
    cellRenderer: ({ item }) => h('span', {
      class: 'text-sm text-gray-700 dark:text-gray-300 truncate'
    }, item.url)
  },
  {
    id: 'method',
    name: '方法',
    width: 80,
    cellRenderer: ({ item }) => h('span', {
      class: [
        'px-2 py-0.5 text-xs rounded-full font-medium',
        {
          'bg-green-100 text-green-800 dark:bg-green-900/50 dark:text-green-300': item.method === 'GET',
          'bg-blue-100 text-blue-800 dark:bg-blue-900/50 dark:text-blue-300': item.method === 'POST',
          'bg-purple-100 text-purple-800 dark:bg-purple-900/50 dark:text-purple-300': item.method === 'PUT',
          'bg-red-100 text-red-800 dark:bg-red-900/50 dark:text-red-300': item.method === 'DELETE',
          'bg-gray-100 text-gray-800 dark:bg-gray-900/50 dark:text-gray-300': !['GET', 'POST', 'PUT', 'DELETE'].includes(item.method)
        }
      ]
    }, item.method)
  },
  {
    id: 'status',
    name: '状态码',
    width: 100,
    cellRenderer: ({ item }) => h('span', {
      class: [
        'px-2 py-0.5 text-xs rounded-full font-medium',
        {
          'bg-green-100 text-green-800 dark:bg-green-900/50 dark:text-green-300': item.status >= 200 && item.status < 300,
          'bg-yellow-100 text-yellow-800 dark:bg-yellow-900/50 dark:text-yellow-300': item.status >= 300 && item.status < 400,
          'bg-red-100 text-red-800 dark:bg-red-900/50 dark:text-red-300': item.status >= 400 && item.status < 500,
          'bg-purple-100 text-purple-800 dark:bg-purple-900/50 dark:text-purple-300': item.status >= 500,
          'bg-gray-100 text-gray-800 dark:bg-gray-900/50 dark:text-gray-300': !item.status
        }
      ]
    }, item.status?.toString() || '-')
  },
  {
    id: 'size',
    name: '长度',
    width: 100,
    cellRenderer: ({ item }) => h('span', {
      class: 'text-sm font-mono text-gray-700 dark:text-gray-300'
    }, item.size?.toString() || '-')
  }
]);

// 处理结果选择
const handleResultSelect = (result: any) => {
  handleRowClick(result);
};

// 控制面板高度调整
const startControlResize = (e: MouseEvent) => {
  isResizingControl = true;
  startY = e.clientY;
  startControlHeight = controlPanelHeight.value;
  document.addEventListener('mousemove', handleControlResize);
  document.addEventListener('mouseup', stopControlResize);
  document.body.classList.add('cursor-ns-resize');
};

const handleControlResize = (e: MouseEvent) => {
  if (!isResizingControl) return;
  const delta = e.clientY - startY;
  controlPanelHeight.value = Math.max(150, Math.min(window.innerHeight - 400, startControlHeight + delta));
};

const stopControlResize = () => {
  isResizingControl = false;
  document.removeEventListener('mousemove', handleControlResize);
  document.removeEventListener('mouseup', stopControlResize);
  document.body.classList.remove('cursor-ns-resize');
};

// 表格高度调整
const startTableResize = (e: MouseEvent) => {
  isResizingTable = true;
  startY = e.clientY;
  startTableHeight = tableHeight.value;
  document.addEventListener('mousemove', handleTableResize);
  document.addEventListener('mouseup', stopTableResize);
  document.body.classList.add('cursor-ns-resize');
};

const handleTableResize = (e: MouseEvent) => {
  if (!isResizingTable) return;
  const delta = e.clientY - startY;
  tableHeight.value = Math.max(100, Math.min(window.innerHeight - 300, startTableHeight + delta));
};

const stopTableResize = () => {
  isResizingTable = false;
  document.removeEventListener('mousemove', handleTableResize);
  document.removeEventListener('mouseup', stopTableResize);
  document.body.classList.remove('cursor-ns-resize');
};

// 监听扫描进度和结果
onMounted(() => {
  // 调整初始高度，分配合理的空间
  const windowHeight = window.innerHeight;
  const totalHeight = windowHeight - 100; // 留出一些边距
  controlPanelHeight.value = Math.min(200, totalHeight * 0.3);
  tableHeight.value = Math.min(300, totalHeight * 0.35);

  // 监听扫描结果
  // Wails v3: e 是 WailsEvent 对象，e.data 是后端发送的 ui.Result 对象
  Events.On("Fuzz", (e: any) => {
    const fuzzResult = e.data;
    if (!fuzzResult) {
      console.warn('Fuzz: 无效的事件数据');
      return;
    }
    let cnt = data.value.length;
    data.value.push({
      id: cnt,
      url: fuzzResult.url,
      method: fuzzResult.method,
      statusCode: fuzzResult.status,
      length: fuzzResult.length,
      request: fuzzResult.request,
      response: fuzzResult.response,
    });
  });

  // 监听扫描进度
  // Wails v3: Percentage 是 WailsEvent 对象，Percentage.data 是后端发送的 float64 值
  Events.On("Percentage", (Percentage: any) => {
    percentage.value = Percentage.data;
  });
});

onBeforeUnmount(() => {
  Events.Off("Fuzz");
  Events.Off("Percentage");
});
</script>

<template>
  <div class="flex flex-col h-full overflow-hidden">
    <!-- 控制面板 -->
    <div class="relative" :style="{ height: controlPanelHeight + 'px' }">
      <div
        class="absolute inset-0 overflow-auto p-4 bg-white dark:bg-[#1e1e2e] rounded-lg border border-gray-200 dark:border-gray-700 shadow-sm">
        <div class="grid grid-cols-1 md:grid-cols-2 gap-4">
          <!-- 目标URL和路径 -->
          <div class="space-y-3">
            <div>
              <label class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">
                {{ t('modules.plugins.fuzz.target_url') }}
              </label>
              <div class="flex">
                <input v-model="formValue.targetUrl" type="text" placeholder="请输入目标URL，例如：https://example.com"
                  class="flex-1 border border-gray-200 dark:border-gray-700 rounded-lg bg-white dark:bg-[#282838] px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-indigo-400 transition-all duration-300 shadow-inner" spellcheck="false"/>
              </div>
            </div>

            <!-- 代理设置 -->
            <div>
              <label class="flex items-center text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">
                <input type="checkbox" v-model="checkedRef"
                  @change="(e) => handleCheckedChange(Boolean((e.target as HTMLInputElement)?.checked))"
                  class="mr-2 rounded text-indigo-600 focus:ring-indigo-500 dark:bg-gray-700" />
                  {{ t('modules.plugins.fuzz.use_proxy') }}
              </label>
              <div class="flex">
                <input ref="inputRef" v-model="proxyInput" :disabled="!checkedRef" type="text"
                  placeholder="127.0.0.1:8080"
                  class="flex-1 border border-gray-200 dark:border-gray-700 rounded-lg bg-white dark:bg-[#282838] px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-indigo-400 transition-all duration-300 shadow-inner disabled:opacity-50" />
              </div>
            </div>
          </div>

          <!-- 选项和按钮 -->
          <div class="space-y-3">
            <div>
              <label class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">
                {{ t('modules.plugins.fuzz.scan_options') }}
              </label>
              <div class="grid grid-cols-3 gap-3">
                <label v-for="option in checkboxOptions" :key="option.value"
                  class="flex items-center text-sm text-gray-700 dark:text-gray-300"
                  :class="{ 'cursor-pointer': true, 'text-indigo-600 dark:text-indigo-400 font-medium': option.value === 'swagger' }">
                  <input type="checkbox" 
                    :checked="formValue.checkboxGroupValue.includes(option.value)"
                    @change="(e) => handleCheckboxChange(option.value, (e.target as HTMLInputElement).checked)"
                    class="mr-2 rounded focus:ring-indigo-500 dark:bg-gray-700"
                    :class="{ 'text-indigo-600': option.value !== 'swagger', 'text-green-600': option.value === 'swagger' }" />
                  {{ option.label }}
                </label>
              </div>
            </div>

            <!-- 操作按钮 -->
            <div class="flex mt-auto pt-4 space-x-3">
              <button @click="fuzz" :disabled="isScanning" class="btn btn-primary flex-1"
                :class="{ 'opacity-50 cursor-not-allowed': isScanning }">
                <i class="bx bx-search-alt mr-1.5"></i> {{ t('modules.plugins.fuzz.start_scan') }}
              </button>
              <button @click="fuzzStop" :disabled="!isScanning" class="btn btn-danger flex-1"
                :class="{ 'opacity-50 cursor-not-allowed': !isScanning }">
                <i class="bx bx-stop-circle mr-1.5"></i> {{ t('modules.plugins.fuzz.stop_scan') }}
              </button>
            </div>

            <!-- 提示信息 -->
            <div v-if="alertContent" class="mt-2 p-3 rounded-lg text-sm" :class="{
              'bg-blue-100/80 dark:bg-blue-900/30 text-blue-800 dark:text-blue-200': alertType === 'info',
              'bg-green-100/80 dark:bg-green-900/30 text-green-800 dark:text-green-200': alertType === 'success',
              'bg-red-100/80 dark:bg-red-900/30 text-red-800 dark:text-red-200': alertType === 'error'
            }">
              <div class="flex items-center">
                <i :class="{
                  'bx bx-info-circle text-blue-500 dark:text-blue-400': alertType === 'info',
                  'bx bx-check-circle text-green-500 dark:text-green-400': alertType === 'success',
                  'bx bx-error text-red-500 dark:text-red-400': alertType === 'error'
                }" class="mr-2 text-lg"></i>
                {{ alertContent }}
              </div>

              <!-- 进度条 -->
              <div v-if="isScanning && alertType === 'info'" class="mt-2">
                <div class="flex items-center justify-between mb-1">
                  <span class="text-xs font-medium text-blue-700 dark:text-blue-300">{{ t('modules.plugins.fuzz.scan_progress') }}</span>
                  <span class="text-xs font-medium text-blue-700 dark:text-blue-300">{{ Math.round(percentage)
                  }}%</span>
                </div>
                <div class="w-full bg-gray-200 dark:bg-gray-700 rounded-full h-2.5">
                  <div class="bg-blue-600 dark:bg-blue-500 h-2.5 rounded-full transition-all duration-300"
                    :style="{ width: `${percentage}%` }"></div>
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>

    </div>

    <!-- 控制面板底部分隔条 -->
    <div class="panel-divider-horizontal" @mousedown="startControlResize"></div>

    <!-- 结果表格区域 -->
    <div class="relative flex-shrink-0" :style="{ height: tableHeight + 'px' }">
      <HttpTrafficTable
        :items="transformedResults"
        :selectedItem="selectedItem"
        :customColumns="fuzzScannerColumns"
        :tableClass="'compact-table'"
        :containerHeight="tableHeight + 'px'"
        tableId="fuzz-scanner-results-table"
        @select-item="handleResultSelect"
      />
    </div>

    <!-- 表格底部分隔条 -->
    <div class="panel-divider-horizontal" @mousedown="startTableResize"></div>

    <!-- 请求/响应面板 -->
    <div
      class="flex-1 flex flex-col overflow-hidden bg-white dark:bg-[#1e1e2e] border border-gray-200 dark:border-gray-700 shadow-sm">
      <div v-if="selectedItem" class="h-full overflow-hidden">
        <RequestResponsePanel :requestData="requestData" :responseData="responseData" :requestReadOnly="true"
          :responseReadOnly="true" :serverDurationMs="0" />
      </div>

      <div v-else class="h-full flex items-center justify-center text-gray-500 dark:text-gray-400">
        <div class="text-center">
          <i class="bx bx-search-alt text-4xl mb-2"></i>
          <p>{{ t('modules.plugins.fuzz.select_scan_result') }}</p>
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
/* 表格样式优化 */
table {
  border-collapse: separate;
  border-spacing: 0;
  width: 100%;
}

/* 滚动条美化 */
::-webkit-scrollbar {
  width: 6px;
  height: 6px;
}

::-webkit-scrollbar-track {
  background: transparent;
}

::-webkit-scrollbar-thumb {
  background: #d1d5db;
  border-radius: 3px;
}

.dark ::-webkit-scrollbar-thumb {
  background: #4b5563;
}

::-webkit-scrollbar-thumb:hover {
  background: #9ca3af;
}

.dark ::-webkit-scrollbar-thumb:hover {
  background: #6b7280;
}
</style>