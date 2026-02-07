<script setup lang="ts">
import { ref, onBeforeUnmount, onMounted, computed, watch, onActivated, onDeactivated, h } from 'vue';
import { useDebounceFn } from '@vueuse/core';
import { useI18n } from 'vue-i18n';
import HttpRequestViewer from '../../common/codemirror/HttpRequestViewer.vue';
import HttpTrafficTable from '../../common/HttpTrafficTable.vue';
import type { InterceptItem } from '../../../types/http';
import type { HttpTrafficColumn } from '../../../types';
import { usePanelResize } from '../../../composables/usePanelResize';

// @ts-ignore
import { Events } from "@wailsio/runtime";
// @ts-ignore
import { 
  ForwardProxyInterceptData 
  // @ts-ignore
} from "../../../../bindings/github.com/yhy0/ChYing/app.js";

const { t } = useI18n();

const props = defineProps<{
  interceptEnabled: boolean;
}>();

const emit = defineEmits<{
  (e: 'notify', message: string, type?: 'success' | 'error'): void;
  (e: 'queue-cleared'): void;
}>();

// 使用 usePanelResize 管理队列表格高度
const { panelHeight: queueTableHeight, startResize: startResizeQueueTable } = usePanelResize({
  panelId: 'intercept-queue-table-height',
  initialHeight: 200,
  minHeight: 120,
  maxHeightOffset: 300
});

// 使用 usePanelResize 管理编辑器区域高度
const { panelHeight: editorsHeight, startResize: startResizeEditors } = usePanelResize({
  panelId: 'intercept-editors-height',
  initialHeight: 400,
  minHeight: 200,
  maxHeightOffset: 150
});

// 左右分割管理 - 请求和响应面板的宽度控制
const requestWidth = ref(50); // 初始宽度50%
let isDragging = false;
let initialX = 0;
let initialLeftWidth = 0;
const containerRef = ref<HTMLElement | null>(null);

// 计算请求部分的宽度
const computedRequestWidth = computed(() => {
  return isResponsePhase.value ? requestWidth.value : 100;
});

// 计算响应部分的宽度  
const computedResponseWidth = computed(() => {
  return 100 - computedRequestWidth.value;
});

// 全局状态管理 - 使用持久化存储
const INTERCEPT_STATE_KEY = 'proxy_intercept_state';

// 状态恢复函数 - 增强版本
const restoreState = () => {
  try {
    const savedState = sessionStorage.getItem(INTERCEPT_STATE_KEY);
    if (savedState) {
      const state = JSON.parse(savedState);

      // 验证状态数据的有效性
      if (state && typeof state === 'object') {
        // 检查时间戳，超过1小时的状态不恢复
        const stateAge = Date.now() - (state.timestamp || 0);
        const maxAge = 60 * 60 * 1000; // 1小时

        if (stateAge > maxAge) {
          console.log('拦截状态过期，使用默认状态');
          sessionStorage.removeItem(INTERCEPT_STATE_KEY);
          return;
        }

        // 验证并恢复状态，过滤掉无效的拦截项
        const validQueue = (state.interceptQueue || []).filter((item: InterceptItem) => {
          return item.id && item.method && item.method !== 'UNKNOWN' && item.url && item.url !== '/';
        });

        interceptQueue.value = validQueue;
        selectedInterceptId.value = state.selectedInterceptId || null;
        interceptSequence.value = state.interceptSequence || 0;
        editedRequestData.value = state.editedRequestData || '';
        editedResponseData.value = state.editedResponseData || '';

        console.log('拦截状态已恢复:', {
          totalItems: state.interceptQueue?.length || 0,
          validItems: validQueue.length,
          selectedId: state.selectedInterceptId
        });
      }
    }
  } catch (error) {
    console.error('恢复拦截状态失败:', error);
    // 清理无效状态
    sessionStorage.removeItem(INTERCEPT_STATE_KEY);
  }
};

// 状态保存函数
const saveState = () => {
  try {
    const state = {
      interceptQueue: interceptQueue.value,
      selectedInterceptId: selectedInterceptId.value,
      interceptSequence: interceptSequence.value,
      editedRequestData: editedRequestData.value,
      editedResponseData: editedResponseData.value,
      timestamp: Date.now()
    };
    sessionStorage.setItem(INTERCEPT_STATE_KEY, JSON.stringify(state));
  } catch (error) {
    console.error('保存拦截状态失败:', error);
  }
};

// 拦截队列 - 存储所有被拦截的项目
const interceptQueue = ref<InterceptItem[]>([]);
// 当前选中的拦截项ID
const selectedInterceptId = ref<string | null>(null);
// 拦截序号计数器
const interceptSequence = ref(0);
// 当前编辑的请求/响应数据
const editedRequestData = ref<string>('');
const editedResponseData = ref<string>('');

// 监听拦截事件
let cleanupInterceptRequestListener: (() => void) | null = null;
let cleanupInterceptResponseListener: (() => void) | null = null;

// 计算属性：当前选中的拦截项
const selectedIntercept = computed(() => {
  return interceptQueue.value.find(item => item.id === selectedInterceptId.value) || null;
});

// 计算属性：是否为响应阶段
const isResponsePhase = computed(() => {
  return selectedIntercept.value?.type === 'response';
});

// 计算属性：请求是否只读（已发送或响应阶段）
const isRequestReadOnly = computed(() => {
  const item = selectedIntercept.value;
  return item?.status === 'sent' || item?.type === 'response';
});

// 计算属性：队列统计信息
const queueStats = computed(() => {
  const total = interceptQueue.value.length;
  const requests = interceptQueue.value.filter(item => item.type === 'request').length;
  const responses = interceptQueue.value.filter(item => item.type === 'response').length;
  return { total, requests, responses };
});

// 设置事件监听器
const setupEventListeners = () => {
  // 清理已有的监听器
  if (cleanupInterceptRequestListener) {
    cleanupInterceptRequestListener();
  }
  if (cleanupInterceptResponseListener) {
    cleanupInterceptResponseListener();
  }

  // 监听请求拦截
  // Wails v3: result 是 WailsEvent 对象，result.data 是后端发送的数据
  cleanupInterceptRequestListener = Events.On("InterceptRequest", result => {
    if (!props.interceptEnabled) return;
    
    // 后端发送的是 map[string]interface{}{"id": flowID, "data": reqDump, "type": "request"}
    const eventData = result.data;
    if (!eventData) {
      console.warn('InterceptRequest: No event data received');
      return;
    }
    const interceptId = eventData.id;
    const requestData = eventData.data;
    
    console.log('Received InterceptRequest:', { interceptId, requestDataLength: requestData?.length });

    // 数据验证
    if (!interceptId || !requestData || requestData.trim() === '') {
      console.warn('Invalid InterceptRequest data:', { interceptId, hasRequestData: !!requestData });
      return;
    }

    // 检查是否已存在相同ID的项目，避免重复
    const existingItem = interceptQueue.value.find(item => item.id === interceptId);
    if (existingItem) {
      console.warn('InterceptRequest: Item already exists with same ID:', interceptId);
      return;
    }

    // 解析请求信息 - 改进URL解析逻辑
    console.log('Parsing request data, first 500 chars:', requestData.substring(0, 500));

    const lines = requestData.split('\n');
    const firstLine = lines[0] || '';
    console.log('First line of request:', firstLine);

    // 更强健的HTTP请求行解析
    const requestLineParts = firstLine.trim().split(' ');
    const method = requestLineParts[0] || 'UNKNOWN';
    let fullPath = requestLineParts[1] || '/';
    const httpVersion = requestLineParts[2] || 'HTTP/1.1';

    console.log('Parsed request line:', { method, fullPath, httpVersion });

    // 初始化变量
    let host = '';
    let port = '';
    let path = '/';
    let protocol = 'http';

    // 检查 fullPath 是否是完整的 URL (如 https://example.com/path)
    if (fullPath.startsWith('http://') || fullPath.startsWith('https://')) {
      try {
        const urlObj = new URL(fullPath);
        host = urlObj.hostname;
        port = urlObj.port || (urlObj.protocol === 'https:' ? '443' : '80');
        path = urlObj.pathname + urlObj.search;
        protocol = urlObj.protocol.replace(':', '');
        console.log('Parsed full URL from request line:', { host, port, path, protocol });
      } catch (e) {
        console.warn('Failed to parse URL from request line:', e);
      }
    }

    // 如果没有从请求行解析出 host，尝试从 Host 头获取
    if (!host) {
      // 查找 Host 头 - 不过滤空行，因为 Host 头可能在任何位置
      const hostLine = lines.find((line: string) => line.toLowerCase().startsWith('host:'));
      if (hostLine) {
        const hostValue = hostLine.substring(5).trim(); // 移除 "Host:" 前缀
        console.log('Found Host header:', hostValue);
        if (hostValue.includes(':')) {
          const parts = hostValue.split(':');
          host = parts[0];
          port = parts[1];
        } else {
          host = hostValue;
          // 如果没有端口，根据请求行中的路径判断
          if (!port) {
            port = '80'; // 默认HTTP端口
          }
        }
      } else {
        console.warn('No Host header found in request data');
        console.log('All lines:', lines.slice(0, 10)); // 打印前10行用于调试
      }
      
      // 使用请求行中的路径
      path = fullPath || '/';
    }

    // 根据端口判断协议
    if (port === '443' || port === '8443') {
      protocol = 'https';
    }

    const url = host ? `${protocol}://${host}${port !== '80' && port !== '443' ? ':' + port : ''}${path}` : path;

    console.log('Final parsed request info:', { method, host, port, path, url });

    console.log('Parsed request info:', { method, host, port, path, url });
    
    // 增加序号
    interceptSequence.value++;
    
    // 创建拦截项
    const interceptItem: InterceptItem = {
      id: interceptId,
      sequence: interceptSequence.value,
      type: 'request',
      method: method || 'UNKNOWN',
      url: url,
      host: host,
      path: path,
      request: requestData,
      timestamp: new Date().toISOString()
    };
    
    // 添加到队列
    interceptQueue.value.push(interceptItem);

    // 只有在当前没有选中项时才自动选中新项，避免干扰用户正在编辑的内容
    if (!selectedInterceptId.value) {
      selectedInterceptId.value = interceptId;
      editedRequestData.value = requestData;
      editedResponseData.value = '';
    }
    
    // 保存状态
    saveState();
    
    console.log('Request intercepted and added to queue:', interceptId, 'Sequence:', interceptSequence.value);
  });
  
  // 监听响应拦截
  // Wails v3: result 是 WailsEvent 对象，result.data 是后端发送的数据
  cleanupInterceptResponseListener = Events.On("InterceptResponse", result => {
    if (!props.interceptEnabled) return;
    
    // 后端发送的是 map[string]interface{}{"id": flowID, "request": reqRaw, "response": respRaw, "type": "response"}
    const eventData = result.data;
    if (!eventData) {
      console.warn('InterceptResponse: No event data received');
      return;
    }
    const interceptId = eventData.id;
    const requestData = eventData.request;
    const responseData = eventData.response;

    console.log('Received InterceptResponse:', {
      interceptId,
      requestLength: requestData?.length,
      responseLength: responseData?.length,
      currentQueueLength: interceptQueue.value.length,
      currentQueueIds: interceptQueue.value.map(item => ({ id: item.id, type: item.type }))
    });

    // 数据验证
    if (!interceptId || !requestData || !responseData) {
      console.warn('Invalid InterceptResponse data:', {
        interceptId,
        hasRequestData: !!requestData,
        hasResponseData: !!responseData
      });
      return;
    }
    
    // 解析响应状态码
    const responseFirstLine = responseData.split('\n')[0] || '';
    const statusMatch = responseFirstLine.match(/HTTP\/[\d.]+\s+(\d+)/);
    const status = statusMatch ? parseInt(statusMatch[1]) : 0;
    
        // 查找对应的已发送请求项目
    console.log('Response intercepted, looking for corresponding sent request item');
    
    let existingIndex = interceptQueue.value.findIndex(item => 
      item.id === interceptId && item.status === 'sent'
    );
    
    if (existingIndex !== -1) {
      // 情况1：找到对应的已发送请求项目，更新为响应阶段
      const existingItem = interceptQueue.value[existingIndex];
      const updatedItem: InterceptItem = {
        ...existingItem, // 保持原有的请求信息
        type: 'response', // 更新为响应阶段
        response: responseData, // 添加响应数据
        status: status, // 添加状态码
        timestamp: new Date().toISOString() // 更新时间戳
      };
      
      // 使用Vue的响应式更新
      interceptQueue.value.splice(existingIndex, 1, updatedItem);
      
      console.log('Updated sent request item to response phase:', interceptId, 'Status:', status);
      
      // 如果这个项目当前被选中，更新编辑器中的数据
      if (selectedInterceptId.value === interceptId) {
        editedRequestData.value = existingItem.request; // 使用已发送的请求数据
        editedResponseData.value = responseData;
      }
    } else {
      // 情况2：没找到对应的已发送请求，可能是：
      // - 只开启了响应拦截
      // - 请求没有被拦截
      // - 请求已经被自动放行
      console.log('No corresponding sent request found, creating standalone response intercept for:', interceptId);
      
      // 解析请求信息创建独立的响应拦截项目
      const requestLines = requestData.split('\n').filter((line: string) => line.trim() !== '');
      const requestFirstLine = requestLines[0] || '';
      const requestLineParts = requestFirstLine.trim().split(' ');
      const method = requestLineParts[0] || 'UNKNOWN';
      const fullPath = requestLineParts[1] || '/';
      
      const hostLine = requestLines.find((line: string) => line.toLowerCase().startsWith('host:'));
      let host = '';
      let port = '';
      if (hostLine) {
        const hostValue = hostLine.substring(5).trim();
        if (hostValue.includes(':')) {
          [host, port] = hostValue.split(':');
        } else {
          host = hostValue;
          port = '80';
        }
      }
      
      let protocol = 'http';
      if (port === '443' || port === '8443') {
        protocol = 'https';
      }
      
      const path = fullPath || '/';
      const url = host ? `${protocol}://${host}${port !== '80' && port !== '443' ? ':' + port : ''}${path}` : path;
    
      // 创建新的独立响应拦截项目
      const interceptItem: InterceptItem = {
        id: interceptId,
        sequence: ++interceptSequence.value,
        type: 'response',
        method: method || 'UNKNOWN',
        url: url,
        host: host,
        path: path,
        request: requestData,
        response: responseData,
        status: status,
        timestamp: new Date().toISOString()
      };
    
      interceptQueue.value.push(interceptItem);
      console.log('Created standalone response intercept item:', interceptId, 'Status:', status);
      
      // 只有在当前没有选中项时才自动选中新创建的独立响应项
      if (!selectedInterceptId.value) {
        selectedInterceptId.value = interceptId;
        editedRequestData.value = requestData;
        editedResponseData.value = responseData;
      }
    }
    
    // 保存状态
    saveState();
  });
};

// 选择拦截项
const selectIntercept = (item: InterceptItem) => {
  selectedInterceptId.value = item.id;
  editedRequestData.value = item.request || '';
  editedResponseData.value = item.response || '';
  saveState();
};

// 数据转换函数：将拦截队列转换为 HttpTrafficItem 格式
const transformedInterceptQueue = computed(() => {
  return interceptQueue.value.map(item => {
    // 确保 host 和 path 有值
    const host = item.host || '';
    const path = item.path || '/';
    const url = item.url || (host ? `${host}${path}` : path);
    
    return {
      id: item.id,
      sequence: item.sequence,
      method: item.method,
      host: host,
      path: path,
      url: url,
      status: typeof item.status === 'number' ? item.status : 0,
      type: item.type,
      timestamp: item.timestamp,
      size: 0,
      // 扩展字段
      request: item.request,
      response: item.response
    };
  });
});

// 当前选中的拦截项
const selectedInterceptItem = computed(() => {
  return transformedInterceptQueue.value.find(item => item.id === selectedInterceptId.value) || null;
});

// 列定义
const interceptColumns = computed<HttpTrafficColumn<any>[]>(() => [
  {
    id: 'sequence',
    name: '#',
    width: 60,
    cellRenderer: ({ item }) => h('span', {
      class: 'font-mono text-xs text-gray-700 dark:text-gray-300'
    }, item.sequence?.toString())
  },
  {
    id: 'method',
    name: t('common.method') || '方法',
    width: 80,
    cellRenderer: ({ item }) => h('span', {
      class: [
        'inline-flex items-center px-2 py-1 rounded text-xs font-medium',
        {
          'bg-blue-100 text-blue-800 dark:bg-blue-900/50 dark:text-blue-200': item.method === 'GET',
          'bg-green-100 text-green-800 dark:bg-green-900/50 dark:text-green-200': item.method === 'POST',
          'bg-yellow-100 text-yellow-800 dark:bg-yellow-900/50 dark:text-yellow-200': item.method === 'PUT',
          'bg-red-100 text-red-800 dark:bg-red-900/50 dark:text-red-200': item.method === 'DELETE',
          'bg-gray-100 text-gray-800 dark:bg-gray-900/50 dark:text-gray-200': !['GET', 'POST', 'PUT', 'DELETE'].includes(item.method)
        }
      ]
    }, item.method)
  },
  {
    id: 'url',
    name: t('common.url') || 'URL',
    width: 400,
    cellRenderer: ({ item }) => {
      const host = item.host || '';
      const path = item.path || '/';
      const displayUrl = item.url || `${host}${path}`;
      
      // 如果有 host，显示 host + path 格式
      if (host) {
        return h('div', { class: 'flex truncate min-w-0 w-full' }, [
          h('span', { class: 'font-medium text-gray-800 dark:text-gray-200 truncate' }, host),
          h('span', { class: 'text-gray-500 dark:text-gray-400 ml-1 truncate' }, path)
        ]);
      }
      // 如果没有 host，显示完整 URL 或 path
      return h('div', { class: 'flex truncate min-w-0 w-full' }, [
        h('span', { class: 'text-gray-600 dark:text-gray-300' }, displayUrl || '(未知)')
      ]);
    }
  },
  {
    id: 'status',
    name: t('common.status') || '状态',
    width: 100,
    cellRenderer: ({ item }) => {
      if (item.status && typeof item.status === 'number') {
        return h('span', {
          class: [
            'text-sm font-medium',
            {
              'text-green-600 dark:text-green-400': item.status >= 200 && item.status < 300,
              'text-yellow-600 dark:text-yellow-400': item.status >= 300 && item.status < 400,
              'text-red-600 dark:text-red-400': item.status >= 400
            }
          ]
        }, item.status.toString());
      } else if (item.status === 'sent') {
        return h('span', { class: 'text-sm font-medium text-blue-600 dark:text-blue-400' }, '已发送');
      } else {
        return h('span', { class: 'text-gray-400 dark:text-gray-500 text-sm' }, '-');
      }
    }
  },
  {
    id: 'type',
    name: t('common.type') || '类型',
    width: 100,
    cellRenderer: ({ item }) => h('div', { class: 'flex flex-col gap-1' }, [
      h('span', {
        class: [
          'inline-flex items-center px-2 py-1 rounded text-xs font-medium',
          {
            'bg-orange-100 text-orange-800 dark:bg-orange-900/50 dark:text-orange-200': item.type === 'request' && item.status !== 'sent',
            'bg-blue-100 text-blue-800 dark:bg-blue-900/50 dark:text-blue-200': item.status === 'sent',
            'bg-purple-100 text-purple-800 dark:bg-purple-900/50 dark:text-purple-200': item.type === 'response'
          }
        ]
      }, item.type === 'request' ? '请求' : '响应'),
      item.status === 'sent' ? h('span', { class: 'text-xs text-gray-500 dark:text-gray-400' }, '请求已发送') : null
    ])
  },
  {
    id: 'timestamp',
    name: t('common.time') || '时间',
    width: 120,
    cellRenderer: ({ item }) => h('span', {
      class: 'text-xs font-mono text-gray-500 dark:text-gray-400'
    }, new Date(item.timestamp).toLocaleTimeString())
  }
]);

// 处理拦截项选择
const handleInterceptSelect = (item: any) => {
  selectIntercept(item);
};

// Forward - 放行当前拦截
const forwardIntercept = () => {
  if (!selectedIntercept.value) {
    console.warn('No intercept selected for forwarding');
    return;
  }

  const interceptId = selectedIntercept.value.id;
  const isRequestPhase = selectedIntercept.value.type === 'request';

  console.log('Forward intercept:', {
    interceptId,
    type: selectedIntercept.value.type,
    requestDataLength: editedRequestData.value?.length,
    responseDataLength: editedResponseData.value?.length
  });

  if (isRequestPhase) {
    // 请求阶段：使用编辑后的请求数据
    console.log('Forwarding request with data:', editedRequestData.value.substring(0, 300) + '...');

    // 如果请求数据被修改，更新队列中的显示信息
    if (editedRequestData.value !== selectedIntercept.value.request) {
      updateInterceptItemFromRequestData(interceptId, editedRequestData.value);
    }

    ForwardProxyInterceptData(interceptId, editedRequestData.value, "forward").then(() => {
      console.log('Request forwarded successfully, waiting for response...');
      
      // 更新项目状态为"已发送，等待响应"，而不是移除
      const itemIndex = interceptQueue.value.findIndex(item => item.id === interceptId);
      if (itemIndex !== -1) {
        const item = interceptQueue.value[itemIndex];
        interceptQueue.value[itemIndex] = {
          ...item,
          status: 'sent', // 标记为已发送
          request: editedRequestData.value, // 保存最终发送的请求数据
          timestamp: new Date().toISOString() // 更新时间戳
        };

        console.log('Updated request item status to sent:', interceptId);
      }
      
      // 保存状态
      saveState();
      
      emit('notify', t('modules.proxy.request_forwarded_waiting') || '请求已放行，等待响应...', 'success');
    }).catch((err: any) => {
      console.error('Error forwarding request:', err);
      emit('notify', t('modules.proxy.forward_error') || '放行失败', 'error');
    });
  } else {
    // 响应阶段：使用编辑后的响应数据
    console.log('Forwarding response with data:', editedResponseData.value.substring(0, 300) + '...');

    ForwardProxyInterceptData(interceptId, editedResponseData.value, "forward").then(() => {
      console.log('Response forwarded successfully, transaction complete');
      
      // 从队列中移除已完成的项目
      interceptQueue.value = interceptQueue.value.filter(item => item.id !== interceptId);

      // 如果移除的是当前选中项，选择下一个项目或清空
      if (selectedInterceptId.value === interceptId) {
        if (interceptQueue.value.length > 0) {
          // 选择队列中的第一个项目
          const nextItem = interceptQueue.value[0];
          selectedInterceptId.value = nextItem.id;
          editedRequestData.value = nextItem.request || '';
          editedResponseData.value = nextItem.response || '';
        } else {
          // 队列为空，清空选中状态
          selectedInterceptId.value = null;
          editedRequestData.value = '';
          editedResponseData.value = '';
        }
      }
      
      // 保存状态
      saveState();
      
      emit('notify', t('modules.proxy.transaction_complete') || '事务已完成', 'success');
    }).catch((err: any) => {
      console.error('Error forwarding response:', err);
      emit('notify', t('modules.proxy.forward_error') || '放行失败', 'error');
    });
  }
};

// Drop - 丢弃当前拦截
const dropIntercept = () => {
  if (!selectedIntercept.value) {
    console.warn('No intercept selected for dropping');
    return;
  }
  
  const interceptId = selectedIntercept.value.id;
  const interceptType = selectedIntercept.value.type;

  console.log('Drop intercept:', { interceptId, type: interceptType });
  
  ForwardProxyInterceptData(interceptId, "", "drop").then(() => {
    console.log('Intercept dropped successfully:', interceptId);
    
    // 从队列中移除已丢弃的项目
    interceptQueue.value = interceptQueue.value.filter(item => item.id !== interceptId);

    // 如果移除的是当前选中项，选择下一个项目或清空
    if (selectedInterceptId.value === interceptId) {
      if (interceptQueue.value.length > 0) {
        // 选择队列中的第一个项目
        const nextItem = interceptQueue.value[0];
        selectedInterceptId.value = nextItem.id;
        editedRequestData.value = nextItem.request || '';
        editedResponseData.value = nextItem.response || '';
      } else {
        // 队列为空，清空选中状态
        selectedInterceptId.value = null;
        editedRequestData.value = '';
        editedResponseData.value = '';
      }
    }
    
    // 保存状态
    saveState();
    
    const message = interceptType === 'request' ?
      (t('modules.proxy.request_dropped') || '请求已丢弃') :
      (t('modules.proxy.response_dropped') || '响应已丢弃');

    emit('notify', message, 'success');
  }).catch((err: any) => {
    console.error('Error dropping intercept:', err);
    emit('notify', t('modules.proxy.drop_error') || '丢弃失败', 'error');
  });
};

// 清空队列
const clearQueue = () => {
  interceptQueue.value = [];
  selectedInterceptId.value = null;
  editedRequestData.value = '';
  editedResponseData.value = '';
  saveState();
  emit('queue-cleared');
};

// 放行所有队列中的请求并清空队列（拦截关闭时调用）
const forwardAllAndClear = () => {
  console.log('Forwarding all intercepted items and clearing queue...');
  
  // 记录要放行的项目数量
  const totalItems = interceptQueue.value.length;
  if (totalItems === 0) {
    console.log('No items in queue to forward');
    return;
  }
  
  console.log(`Forwarding ${totalItems} items from intercept queue`);
  
  // 为每个拦截项目发送放行指令
  const forwardPromises = interceptQueue.value.map(item => {
    console.log(`Forwarding item: ${item.id} (${item.type})`);
    
    // 根据类型使用对应的数据
    const dataToSend = item.type === 'request' ? item.request : 
                      item.response || item.request;
    
    return ForwardProxyInterceptData(item.id, dataToSend, "forward").catch((err: any) => {
      console.error(`Error forwarding item ${item.id}:`, err);
      // 即使单个项目失败，也继续处理其他项目
    });
  });
  
  // 等待所有放行操作完成
  Promise.allSettled(forwardPromises).then((results) => {
    const successful = results.filter(r => r.status === 'fulfilled').length;
    const failed = results.filter(r => r.status === 'rejected').length;
    
    console.log(`Forwarded ${successful} items successfully, ${failed} failed`);
    
    // 清空队列
    clearQueue();
    
    // 显示通知
    emit('notify', 
      failed === 0 
        ? `已放行所有 ${totalItems} 个拦截项目` 
        : `已放行 ${successful} 个项目，${failed} 个失败`, 
      failed === 0 ? 'success' : 'error'
    );
  });
};

// 暴露方法给父组件
defineExpose({
  forwardAllAndClear,
  clearQueue
});

// 更新请求数据
const updateRequestData = (data: string) => {
  editedRequestData.value = data;
  saveState(); // 保存状态变更
};

// 更新响应数据
const updateResponseData = (data: string) => {
  editedResponseData.value = data;
  saveState(); // 保存状态变更
};

// 从修改后的请求数据更新拦截项的显示信息
const updateInterceptItemFromRequestData = (interceptId: string, requestData: string) => {
  const itemIndex = interceptQueue.value.findIndex(item => item.id === interceptId);
  if (itemIndex === -1) {
    console.warn('Could not find intercept item to update:', interceptId);
    return;
  }

  try {
    // 解析修改后的请求数据
    const lines = requestData.split('\n');
    const firstLine = lines[0] || '';
    const [method, fullPath] = firstLine.split(' ');

    // 正确解析Host头
    const hostLine = lines.find((line: string) => line.toLowerCase().startsWith('host:'));
    let host = '';
    let port = '';
    if (hostLine) {
      const hostValue = hostLine.substring(5).trim();
      if (hostValue.includes(':')) {
        [host, port] = hostValue.split(':');
      } else {
        host = hostValue;
        port = '80';
      }
    }

    // 构建完整URL
    let protocol = 'http';
    if (port === '443' || port === '8443') {
      protocol = 'https';
    }

    const path = fullPath || '/';
    const url = host ? `${protocol}://${host}${port !== '80' && port !== '443' ? ':' + port : ''}${path}` : path;

    // 更新拦截项
    const item = interceptQueue.value[itemIndex];
    interceptQueue.value[itemIndex] = {
      ...item,
      method: method || item.method,
      url: url,
      host: host || item.host,
      path: path,
      request: requestData, // 更新请求数据
      timestamp: new Date().toISOString() // 更新时间戳
    };

    console.log('Updated intercept item with modified request data:', {
      id: interceptId,
      method: method,
      url: url,
      host: host,
      path: path
    });

    // 保存状态
    saveState();
  } catch (error) {
    console.error('Error parsing modified request data:', error);
  }
};

// 左右分割拖拽处理方法
const startResizeRequest = (e: MouseEvent) => {
  e.preventDefault();
  isDragging = true;
  initialX = e.clientX;

  // 获取初始请求面板宽度
  if (containerRef.value) {
    const containerRect = containerRef.value.getBoundingClientRect();
    initialLeftWidth = (requestWidth.value / 100) * containerRect.width;
  }

  // 添加全局事件监听器
  document.addEventListener('mousemove', handleMouseMove);
  document.addEventListener('mouseup', stopResize);

  // 添加拖拽中的样式
  document.body.classList.add('dragging');
};

// 拖拽过程
const handleMouseMove = (e: MouseEvent) => {
  if (!isDragging || !containerRef.value) return;

  const containerRect = containerRef.value.getBoundingClientRect();
  const deltaX = e.clientX - initialX;
  const newLeftPanelWidth = initialLeftWidth + deltaX;

  // 计算百分比宽度
  const percentWidth = (newLeftPanelWidth / containerRect.width) * 100;

  // 限制宽度范围（20% - 80%）
  requestWidth.value = Math.max(20, Math.min(80, percentWidth));
};

// 停止拖拽
const stopResize = () => {
  isDragging = false;

  // 移除全局事件监听器
  document.removeEventListener('mousemove', handleMouseMove);
  document.removeEventListener('mouseup', stopResize);

  // 移除拖拽中的样式
  document.body.classList.remove('dragging');
};

// 键盘快捷键
const handleKeyDown = (event: KeyboardEvent) => {
  if (!selectedIntercept.value) return;
  
  // 确保没有在输入框中
  if (event.target instanceof HTMLInputElement ||
      event.target instanceof HTMLTextAreaElement) {
    return;
  }
  
  if (event.key.toLowerCase() === 'f') {
    event.preventDefault();
    forwardIntercept();
  } else if (event.key.toLowerCase() === 'd') {
    event.preventDefault();
    dropIntercept();
  }
};

// 使用防抖保存状态，避免每次按键都触发 sessionStorage 写入
const debouncedSaveState = useDebounceFn(() => {
  saveState();
}, 500);

// 监听拦截状态变化，保存状态（使用防抖）
watch([interceptQueue, selectedInterceptId, editedRequestData, editedResponseData], () => {
  debouncedSaveState();
}, { deep: true });

// 组件激活时（标签页切换到拦截面板）
onActivated(() => {
  console.log('拦截面板激活，恢复状态...');
  restoreState();
  setupEventListeners();
  window.addEventListener('keydown', handleKeyDown);
});

// 组件失活时（切换到其他标签页）
onDeactivated(() => {
  console.log('拦截面板失活，保存状态...');
  saveState();
  window.removeEventListener('keydown', handleKeyDown);
});

onMounted(() => {
  console.log('拦截面板挂载，初始化...');
  restoreState();
  setupEventListeners();
  window.addEventListener('keydown', handleKeyDown);

  // 清理可能存在的拖拽事件监听器
  document.removeEventListener('mousemove', handleMouseMove);
  document.removeEventListener('mouseup', stopResize);
  document.body.classList.remove('dragging');
});

onBeforeUnmount(() => {
  console.log('拦截面板卸载，清理资源...');
  saveState();
  window.removeEventListener('keydown', handleKeyDown);

  // 清理拖拽事件监听器
  document.removeEventListener('mousemove', handleMouseMove);
  document.removeEventListener('mouseup', stopResize);
  document.body.classList.remove('dragging');
  
  // 移除事件监听器
  if (cleanupInterceptRequestListener) {
    cleanupInterceptRequestListener();
    cleanupInterceptRequestListener = null;
  }
  if (cleanupInterceptResponseListener) {
    cleanupInterceptResponseListener();
    cleanupInterceptResponseListener = null;
  }
});
</script>

<template>
  <div class="flex-1 flex flex-col overflow-hidden">
    <!-- 顶部状态栏 -->
    <div class="px-4 py-2 border-b border-gray-200 dark:border-gray-700 bg-gray-50 dark:bg-[#292945] flex-shrink-0">
      <div class="flex items-center justify-between">
        <!-- 左侧：拦截状态 -->
        <div class="flex items-center">
          <h3 class="text-sm font-medium text-gray-700 dark:text-gray-300 mr-3">
            {{ t('modules.proxy.intercept') }}
          </h3>
          <span class="text-xs py-1 px-2 rounded"
            :class="interceptEnabled ? 'bg-red-100 dark:bg-red-900/30 text-red-800 dark:text-red-300' : 'bg-gray-100 dark:bg-gray-800 text-gray-600 dark:text-gray-400'">
            {{ interceptEnabled ? t('modules.proxy.controls.intercept_on') : t('modules.proxy.controls.intercept_off')
            }}
          </span>
        </div>
        
        <!-- 右侧：队列统计和操作 -->
        <div class="flex items-center space-x-4">
          <div class="text-xs text-gray-600 dark:text-gray-400">
            {{ t('modules.proxy.queue_stats', {
              total: queueStats.total, requests: queueStats.requests, responses:
                queueStats.responses
            }) || `队列: ${queueStats.total} 项 (${queueStats.requests} 请求, ${queueStats.responses}
            响应)` }}
          </div>
          <!-- 调试信息 -->
          <div v-if="selectedIntercept" class="text-xs text-blue-600 dark:text-blue-400">
            {{ t('modules.proxy.selected') || '选中' }}: #{{ selectedIntercept.sequence }}
            {{ selectedIntercept.type === 'request' ? '请求' : '响应' }}
            <span v-if="selectedIntercept.type === 'response'" class="ml-1 text-gray-500">(含请求)</span>
          </div>
          <button v-if="interceptQueue.length > 0" @click="clearQueue"
            class="text-xs px-2 py-1 text-gray-600 dark:text-gray-400 hover:text-red-600 dark:hover:text-red-400 transition-colors">
            {{ t('modules.proxy.clear_queue') || '清空队列' }}
          </button>
        </div>
      </div>

    </div>

    <!-- 主要内容区域 -->
    <div class="flex-1 flex flex-col overflow-hidden">
      <!-- 拦截队列表格 -->
      <div class="border-b border-gray-200 dark:border-gray-700 overflow-hidden"
        :style="{ height: queueTableHeight + 'px' }">
        <HttpTrafficTable
          :items="transformedInterceptQueue"
          :selectedItem="selectedInterceptItem"
          :customColumns="interceptColumns"
          :tableClass="'compact-table'"
          :containerHeight="queueTableHeight + 'px'"
          tableId="proxy-intercept-queue-table"
          @select-item="handleInterceptSelect"
        />

      </div>

      <!-- 队列表格与编辑器区域之间的分隔线 -->
      <div class="panel-divider-horizontal cursor-ns-resize border-b border-gray-200 dark:border-gray-700"
        @mousedown="startResizeQueueTable"></div>

      <!-- 下半部分：选中项的请求/响应编辑器 -->
      <div v-if="selectedIntercept" class="flex-1 flex flex-col overflow-hidden">
        <!-- 选中项信息头 -->
                <div
          class="px-4 py-2 border-b border-gray-200 dark:border-gray-700 bg-gray-50/50 dark:bg-gray-800/20 flex-shrink-0">
        <div class="flex items-center justify-between">
            <div class="flex items-center min-w-0 flex-1 mr-4">
              <span class="text-xs font-medium text-gray-600 dark:text-gray-400 mr-2 flex-shrink-0">#{{ selectedIntercept.sequence
                }}</span>
              <span class="intercept-method flex-shrink-0">{{ selectedIntercept.method }}</span>
              <span class="font-mono text-xs truncate min-w-0 flex-1" :title="selectedIntercept.url">
                <span class="font-medium text-gray-800 dark:text-gray-300">{{ selectedIntercept.host }}</span>
                <span class="text-gray-500 dark:text-gray-400">{{ selectedIntercept.path }}</span>
            </span>
              <span v-if="isResponsePhase && selectedIntercept.status && typeof selectedIntercept.status === 'number'"
                class="ml-3 text-sm font-medium flex-shrink-0" :class="{
                  'text-green-600 dark:text-green-400': selectedIntercept.status >= 200 && selectedIntercept.status < 300,
                  'text-yellow-600 dark:text-yellow-400': selectedIntercept.status >= 300 && selectedIntercept.status < 400,
                  'text-red-600 dark:text-red-400': selectedIntercept.status >= 400
            }">
                {{ selectedIntercept.status }}
            </span>
          </div>
          <!-- 操作按钮 -->
             <div class="flex items-center space-x-2 flex-shrink-0">
               <button @click="forwardIntercept"
                 class="px-3 py-1.5 bg-green-500 hover:bg-green-600 text-white text-sm font-medium rounded transition-colors duration-150 flex items-center">
              <i class="bx bx-play mr-1.5"></i>
              {{ t('modules.proxy.forward') || 'Forward' }}
              <span class="ml-2 text-xs opacity-75">(F)</span>
            </button>
               <button @click="dropIntercept"
                 class="px-3 py-1.5 bg-red-500 hover:bg-red-600 text-white text-sm font-medium rounded transition-colors duration-150 flex items-center">
              <i class="bx bx-trash mr-1.5"></i>
              {{ t('modules.proxy.drop') || 'Drop' }}
              <span class="ml-2 text-xs opacity-75">(D)</span>
            </button>
          </div>
        </div>
      </div>

      <!-- 编辑器区域 -->
        <div class="flex overflow-hidden" :style="{ height: editorsHeight + 'px' }" ref="containerRef">
        <!-- 请求编辑器 -->
          <div class="flex flex-col overflow-hidden" :style="{ width: computedRequestWidth + '%' }">
            <div
              class="px-3 py-2 border-b border-gray-200 dark:border-gray-700 bg-gray-50 dark:bg-[#292945] flex-shrink-0">
            <span class="text-xs font-medium text-gray-600 dark:text-gray-400">
                {{ t('common.request') }} - {{ isRequestReadOnly ? (t('modules.proxy.final_request') || '最终请求') :
                  (t('modules.proxy.intercepted_request') || '拦截请求') }}
                <span v-if="!isRequestReadOnly" class="ml-2 text-green-600 dark:text-green-400">({{
                  t('modules.proxy.editable') || '可编辑' }})</span>
                <span v-else class="ml-2 text-gray-500 dark:text-gray-400">({{ t('modules.proxy.readonly') || '只读'
                }})</span>
            </span>
          </div>
            <HttpRequestViewer class="flex-1" :data="editedRequestData" @update:data="updateRequestData"
              :readOnly="isRequestReadOnly" />
        </div>

          <!-- 左右分隔线（仅在响应阶段显示） -->
          <div v-if="isResponsePhase" class="panel-divider" @mousedown="startResizeRequest"></div>
        
        <!-- 响应编辑器（仅在响应阶段显示） -->
          <div v-if="isResponsePhase" class="flex flex-col overflow-hidden"
            :style="{ width: computedResponseWidth + '%' }">
            <div
              class="px-3 py-2 border-b border-gray-200 dark:border-gray-700 bg-gray-50 dark:bg-[#292945] flex-shrink-0">
            <span class="text-xs font-medium text-gray-600 dark:text-gray-400">
              {{ t('common.response') }} - {{ t('modules.proxy.intercepted_response') || '拦截响应' }}
                <span class="ml-2 text-green-600 dark:text-green-400">({{ t('modules.proxy.editable') || '可编辑'
                }})</span>
            </span>
          </div>
            <HttpRequestViewer class="flex-1" :data="editedResponseData" @update:data="updateResponseData"
              :readOnly="false" />
        </div>
        
        <!-- 等待响应提示（请求阶段） -->
          <div v-else class="flex flex-col overflow-hidden" :style="{ width: '50%' }">
            <div
              class="px-3 py-2 border-b border-gray-200 dark:border-gray-700 bg-gray-50 dark:bg-[#292945] flex-shrink-0">
            <span class="text-xs font-medium text-gray-600 dark:text-gray-400">
              {{ t('common.response') }}
            </span>
          </div>
          <div class="flex-1 flex items-center justify-center text-gray-500 dark:text-gray-400">
            <div class="text-center">
              <i class="bx bx-time-five text-3xl mb-3"></i>
              <p>{{ t('modules.proxy.no_response_yet') || '暂无响应' }}</p>
              <p class="text-xs mt-2">{{ t('modules.proxy.forward_request_first') || '请先放行请求以获取响应' }}</p>
            </div>
          </div>
        </div>
      </div>
    </div>

      <!-- 空状态（没有选中项时） -->
    <div v-else class="flex-1 flex items-center justify-center text-gray-500 dark:text-gray-400">
        <div class="text-center max-w-md">
        <i class="bx bx-intersect text-4xl mb-4"></i>
          <h3 class="text-lg font-medium mb-2">{{ t('modules.proxy.no_item_selected') || '未选择拦截项' }}</h3>
          <p class="text-sm mb-4">
            {{ interceptQueue.length > 0 ? (t('modules.proxy.select_item_hint') || '请从上方队列中选择一个拦截项') : (interceptEnabled
              ? (t('modules.proxy.intercept_waiting_hint') || '当有请求被拦截时，将在队列中显示') :
              (t('modules.proxy.enable_intercept_hint') || '请启用拦截功能')) }}
        </p>

          <!-- 测试提示 -->
          <div v-if="interceptEnabled && interceptQueue.length === 0"
            class="mt-6 p-4 bg-blue-50 dark:bg-blue-900/20 rounded-lg border border-blue-200 dark:border-blue-800">
            <h4 class="text-sm font-medium text-blue-800 dark:text-blue-200 mb-2">
              <i class="bx bx-info-circle mr-1"></i>
              {{ t('modules.proxy.test_intercept') || '测试拦截功能' }}
            </h4>
            <div class="text-xs text-blue-700 dark:text-blue-300 space-y-1">
              <p>1. 确保浏览器代理设置为 ChYing 代理地址（点击顶部浏览器图标可自动配置）</p>
              <p>2. 访问任意网站（如 httpbin.org/get）</p>
              <p>3. 请求将出现在上方队列中等待处理</p>
              <p>4. 可以修改请求内容，然后点击 Forward 或 Drop</p>
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
/* 使用 CSS 变量和统一的设计系统 */
/* 所有样式已在 proxy.css 中定义，这里只做必要的补充 */

/* 表格行选中高亮 */
.bg-blue-50 {
  background-color: rgba(59, 130, 246, 0.05);
}

.dark .bg-blue-50 {
  background-color: rgba(59, 130, 246, 0.1);
}

.border-l-2 {
  border-left-width: 2px;
}

.border-blue-500 {
  border-left-color: rgb(59, 130, 246);
}

/* 确保表格的响应式和溢出处理 */
.h-48 {
  height: 12rem;
}

/* 方法标签颜色 */
.bg-blue-100 {
  background-color: rgba(59, 130, 246, 0.1);
}

.text-blue-800 {
  color: rgb(30, 64, 175);
}

.bg-green-100 {
  background-color: rgba(34, 197, 94, 0.1);
}

.text-green-800 {
  color: rgb(22, 101, 52);
}

.bg-yellow-100 {
  background-color: rgba(245, 158, 11, 0.1);
}

.text-yellow-800 {
  color: rgb(146, 64, 14);
}

.bg-red-100 {
  background-color: rgba(239, 68, 68, 0.1);
}

.text-red-800 {
  color: rgb(153, 27, 27);
}

.bg-orange-100 {
  background-color: rgba(249, 115, 22, 0.1);
}

.text-orange-800 {
  color: rgb(154, 52, 18);
}

.bg-purple-100 {
  background-color: rgba(168, 85, 247, 0.1);
}

.text-purple-800 {
  color: rgb(107, 33, 168);
}

/* 深色模式下的方法标签颜色 */
.dark .bg-blue-100 {
  background-color: rgba(59, 130, 246, 0.2);
}

.dark .text-blue-800 {
  color: rgb(147, 197, 253);
}

.dark .bg-green-100 {
  background-color: rgba(34, 197, 94, 0.2);
}

.dark .text-green-800 {
  color: rgb(134, 239, 172);
}

.dark .bg-yellow-100 {
  background-color: rgba(245, 158, 11, 0.2);
}

.dark .text-yellow-800 {
  color: rgb(253, 224, 71);
}

.dark .bg-red-100 {
  background-color: rgba(239, 68, 68, 0.2);
}

.dark .text-red-800 {
  color: rgb(252, 165, 165);
}

.dark .bg-orange-100 {
  background-color: rgba(249, 115, 22, 0.2);
}

.dark .text-orange-800 {
  color: rgb(251, 146, 60);
}

.dark .bg-purple-100 {
  background-color: rgba(168, 85, 247, 0.2);
}

.dark .text-purple-800 {
  color: rgb(196, 181, 253);
}

/* 方法标签样式 */
.intercept-method {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  padding: 0.125rem 0.375rem;
  font-size: 0.75rem;
  font-weight: 600;
  border-radius: 0.25rem;
  margin-right: 0.5rem;
  min-width: 3.5rem;
  text-align: center;
  background-color: rgba(59, 130, 246, 0.1);
  color: rgb(37, 99, 235);
}

.dark .intercept-method {
  background-color: rgba(59, 130, 246, 0.2);
  color: rgb(96, 165, 250);
}

/* 面板分隔线样式 */
.panel-divider-horizontal {
  height: 4px;
  background-color: transparent;
  position: relative;
  flex-shrink: 0;
}

.panel-divider-horizontal:hover {
  background-color: var(--color-brand);
}

.panel-divider-horizontal::before {
  content: '';
  position: absolute;
  top: 50%;
  left: 50%;
  transform: translate(-50%, -50%);
  width: 40px;
  height: 2px;
  background-color: var(--color-border);
  border-radius: 1px;
  transition: background-color 0.2s ease;
}

.panel-divider-horizontal:hover::before {
  background-color: white;
}

/* 拖拽时的样式 */
.dragging {
  user-select: none;
  cursor: ew-resize !important;
  }

.dragging * {
  pointer-events: none;
}

/* === 拦截队列表格特有样式 === */

/* 特别针对类型列中可能有多个标签的情况 */
.intercept-queue-table.compact-table td:nth-child(5) .inline-flex {
  margin-bottom: 1px !important;
}

.intercept-queue-table.compact-table td:nth-child(5) .flex {
  gap: 2px !important;
}

/* 拦截队列表格列宽定义 */
.intercept-queue-table {
  table-layout: fixed;
}

.intercept-queue-table th:nth-child(1) { /* # */
  width: 60px;
}

.intercept-queue-table th:nth-child(2) { /* 方法 */
  width: 80px;
}

.intercept-queue-table th:nth-child(3) { /* URL */
  width: auto; /* 自动宽度，占据剩余空间 */
}

.intercept-queue-table th:nth-child(4) { /* 状态 */
  width: 80px;
}

.intercept-queue-table th:nth-child(5) { /* 类型 */
  width: 80px;
}

.intercept-queue-table th:nth-child(6) { /* 时间 */
  width: 140px;
}

/* URL列特殊样式 - 允许适当的文本截断 */
.intercept-queue-table td:nth-child(3) {
  max-width: 0 !important; /* 强制文本截断 */
  text-overflow: ellipsis !important;
}
</style> 