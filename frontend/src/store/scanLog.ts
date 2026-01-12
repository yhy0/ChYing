import { defineStore } from 'pinia';
import { ref, computed } from 'vue';
import type { 
  RequestScanMsg, 
  ScanLogItem, 
  ScanStatistics, 
  ScanLogFilter 
} from '../types';

// localStorage 键名
const SCAN_LOG_STORAGE_KEY = 'chying-scan-logs';
const SCAN_LOG_FILTER_STORAGE_KEY = 'chying-scan-log-filter';

/**
 * 从 localStorage 加载扫描日志数据
 */
const loadScanLogsFromStorage = (): ScanLogItem[] => {
  try {
    const stored = localStorage.getItem(SCAN_LOG_STORAGE_KEY);
    if (stored) {
      const parsed = JSON.parse(stored);
      if (Array.isArray(parsed)) {
        console.log('从 localStorage 恢复扫描日志数据:', parsed.length, '条');
        return parsed;
      }
    }
  } catch (error) {
    console.error('加载扫描日志数据失败:', error);
  }
  return [];
};

/**
 * 从 localStorage 加载过滤器设置
 */
const loadFilterFromStorage = (): ScanLogFilter => {
  try {
    const stored = localStorage.getItem(SCAN_LOG_FILTER_STORAGE_KEY);
    if (stored) {
      const parsed = JSON.parse(stored);
      console.log('从 localStorage 恢复过滤器设置:', parsed);
      return {
        moduleTypes: parsed.moduleTypes || [],
        searchKeyword: parsed.searchKeyword || '',
        targetHost: parsed.targetHost || '',
        statusCodes: parsed.statusCodes || []
      };
    }
  } catch (error) {
    console.error('加载过滤器设置失败:', error);
  }
  return {
    moduleTypes: [],
    searchKeyword: '',
    targetHost: '',
    statusCodes: []
  };
};

/**
 * 保存扫描日志数据到 localStorage
 */
const saveScanLogsToStorage = (logs: ScanLogItem[]) => {
  try {
    // 限制存储数量，避免 localStorage 过大（保留最近 1000 条）
    const maxStorageCount = 1000;
    const logsToSave = logs.length > maxStorageCount ? logs.slice(0, maxStorageCount) : logs;
    localStorage.setItem(SCAN_LOG_STORAGE_KEY, JSON.stringify(logsToSave));
    console.log('扫描日志数据已保存到 localStorage:', logsToSave.length, '条');
  } catch (error) {
    console.error('保存扫描日志数据失败:', error);
  }
};

/**
 * 保存过滤器设置到 localStorage
 */
const saveFilterToStorage = (filterData: ScanLogFilter) => {
  try {
    localStorage.setItem(SCAN_LOG_FILTER_STORAGE_KEY, JSON.stringify(filterData));
    console.log('过滤器设置已保存到 localStorage');
  } catch (error) {
    console.error('保存过滤器设置失败:', error);
  }
};

/**
 * 扫描日志状态管理
 */
export const useScanLogStore = defineStore('scanLog', () => {
  // 状态 - 初始化时从 localStorage 恢复数据
  const scanLogs = ref<ScanLogItem[]>(loadScanLogsFromStorage());
  const selectedLogItem = ref<ScanLogItem | null>(null);
  const filter = ref<ScanLogFilter>(loadFilterFromStorage());

  // 获取器
  const totalCount = computed(() => scanLogs.value.length);
  
  const statistics = computed<ScanStatistics>(() => {
    const logs = scanLogs.value;
    
    // 计算模块统计
    const byModule: Record<string, number> = {};
    logs.forEach(log => {
      byModule[log.moduleName] = (byModule[log.moduleName] || 0) + 1;
    });

    return {
      total: logs.length,
      byModule
    };
  });

  // 过滤后的日志
  const filteredLogs = computed(() => {
    let filtered = scanLogs.value;
    console.log('filteredLogs计算 - 原始数据:', filtered.length, filtered);
    console.log('当前过滤器:', filter.value);

    if (filter.value.moduleTypes.length > 0) {
      filtered = filtered.filter(log => 
        filter.value.moduleTypes.includes(log.moduleName)
      );
      console.log('模块类型过滤后:', filtered.length);
    }

    if (filter.value.statusCodes && filter.value.statusCodes.length > 0) {
      filtered = filtered.filter(log => 
        filter.value.statusCodes!.includes(log.status)
      );
      console.log('状态码过滤后:', filtered.length);
    }

    if (filter.value.searchKeyword && filter.value.searchKeyword.trim()) {
      const keyword = filter.value.searchKeyword.toLowerCase();
      filtered = filtered.filter(log => 
        log.url.toLowerCase().includes(keyword) ||
        log.moduleName.toLowerCase().includes(keyword) ||
        log.title.toLowerCase().includes(keyword) ||
        log.host.toLowerCase().includes(keyword) ||
        log.target.toLowerCase().includes(keyword) ||
        (log.vulnerability && log.vulnerability.toLowerCase().includes(keyword)) ||
        (log.description && log.description.toLowerCase().includes(keyword))
      );
      console.log('关键词过滤后:', filtered.length);
    }

    if (filter.value.targetHost && filter.value.targetHost.trim()) {
      filtered = filtered.filter(log => 
        log.host.includes(filter.value.targetHost!)
      );
      console.log('目标主机过滤后:', filtered.length);
    }

    console.log('最终过滤结果:', filtered.length, filtered);
    return filtered;
  });

  // 操作
  
  /**
   * 添加扫描日志（从 RequestScanMsg 转换）
   */
  const addScanLog = (requestMsg: RequestScanMsg | RequestScanMsg[]) => {
    // 处理数组格式的数据
    let actualMsg: RequestScanMsg;
    if (Array.isArray(requestMsg)) {
      if (requestMsg.length === 0) {
        console.warn('Empty RequestScanMsg array received');
        return;
      }
      actualMsg = requestMsg[0];
    } else {
      actualMsg = requestMsg;
    }

    const scanLogItem: ScanLogItem = convertRequestMsgToScanLogItem(actualMsg);
    console.log("------", scanLogItem)
    
    // 检查是否已存在，避免重复添加
    const existingIndex = scanLogs.value.findIndex(log => log.id === scanLogItem.id);
    if (existingIndex !== -1) {
      // 更新现有项 - 确保响应式更新
      scanLogs.value.splice(existingIndex, 1, scanLogItem);
    } else {
      // 添加新项 - 使用splice确保响应式
      scanLogs.value.unshift(scanLogItem);
    }
    console.log("---scanLogs---", scanLogs.value)
    
    // 强制触发响应式更新
    scanLogs.value = [...scanLogs.value];

    // 保存到 localStorage
    saveScanLogsToStorage(scanLogs.value);
  };

  /**
   * 批量处理扫描消息
   */
  const handleRequestScanMessages = (payload: { data: RequestScanMsg | RequestScanMsg[] }) => {
    addScanLog(payload.data);
  };

  /**
   * 设置选中的日志项
   */
  const setSelectedLogItem = (item: ScanLogItem | null) => {
    selectedLogItem.value = item;
  };

  /**
   * 更新过滤器
   */
  const updateFilter = (newFilter: Partial<ScanLogFilter>) => {
    filter.value = { ...filter.value, ...newFilter };
    
    // 保存过滤器设置到 localStorage
    saveFilterToStorage(filter.value);
  };

  /**
   * 清空日志
   */
  const clearLogs = () => {
    scanLogs.value = [];
    selectedLogItem.value = null;
    
    // 同时清除 localStorage 中的数据
    try {
      localStorage.removeItem(SCAN_LOG_STORAGE_KEY);
      console.log('已清除 localStorage 中的扫描日志数据');
    } catch (error) {
      console.error('清除 localStorage 扫描日志数据失败:', error);
    }
  };

  /**
   * 清空日志并同步到其他窗口
   */
  const clearLogsWithSync = () => {
    // 清空当前窗口的日志
    clearLogs();
    
    // 触发跨窗口同步事件
    const syncData = {
      type: 'clear',
      data: null,
      timestamp: Date.now()
    };
    
    // 通过localStorage触发跨窗口事件
    localStorage.setItem('scan-log-data', JSON.stringify(syncData));
    
    // 立即触发storage事件（对于当前窗口）
    window.dispatchEvent(new StorageEvent('storage', {
      key: 'scan-log-data',
      newValue: JSON.stringify(syncData),
      oldValue: null,
      storageArea: localStorage,
      url: window.location.href
    }));
  };

  /**
   * 设置行颜色
   */
  const setItemColor = (item: ScanLogItem, color: string) => {
    const index = scanLogs.value.findIndex(log => log.id === item.id);
    if (index !== -1) {
      scanLogs.value[index].color = color;
      // 保存到 localStorage
      saveScanLogsToStorage(scanLogs.value);
    }
  };

  /**
   * 设置行备注
   */
  const setItemNote = (item: ScanLogItem, note: string) => {
    const index = scanLogs.value.findIndex(log => log.id === item.id);
    if (index !== -1) {
      scanLogs.value[index].note = note;
      // 保存到 localStorage
      saveScanLogsToStorage(scanLogs.value);
    }
  };

  /**
   * 清理过期数据（保留最近7天的数据）
   */
  const cleanupOldLogs = () => {
    const sevenDaysAgo = new Date(Date.now() - 7 * 24 * 60 * 60 * 1000);
    const initialCount = scanLogs.value.length;
    
    scanLogs.value = scanLogs.value.filter(log => {
      const logDate = new Date(log.timestamp);
      return logDate >= sevenDaysAgo;
    });
    
    const removedCount = initialCount - scanLogs.value.length;
    if (removedCount > 0) {
      console.log(`清理了 ${removedCount} 条过期扫描日志（7天前）`);
      saveScanLogsToStorage(scanLogs.value);
    }
  };

  /**
   * 获取存储空间使用情况
   */
  const getStorageInfo = () => {
    try {
      const data = localStorage.getItem(SCAN_LOG_STORAGE_KEY);
      const sizeInBytes = data ? new Blob([data]).size : 0;
      const sizeInMB = (sizeInBytes / (1024 * 1024)).toFixed(2);
      return {
        itemCount: scanLogs.value.length,
        sizeInBytes,
        sizeInMB: parseFloat(sizeInMB)
      };
    } catch (error) {
      console.error('获取存储信息失败:', error);
      return { itemCount: 0, sizeInBytes: 0, sizeInMB: 0 };
    }
  };

  return {
    // 状态
    scanLogs,
    selectedLogItem,
    filter,
    
    // 获取器
    totalCount,
    statistics,
    filteredLogs,
    
    // 操作
    addScanLog,
    handleRequestScanMessages,
    setSelectedLogItem,
    updateFilter,
    clearLogs,
    clearLogsWithSync,
    setItemColor,
    setItemNote,
    cleanupOldLogs,
    getStorageInfo
  };
});

/**
 * 将 RequestScanMsg 转换为 ScanLogItem
 */
function convertRequestMsgToScanLogItem(requestMsg: RequestScanMsg): ScanLogItem {
  console.log('Processing RequestScanMsg:', requestMsg);
  
  // 从 target 解析 host 和其他URL信息
  let host = '';
  let url = requestMsg.target || '';
  
  // 安全地解析 URL
  if (url) {
    try {
      const urlObj = new URL(url);
      host = urlObj.hostname;
    } catch {
      // 如果解析失败，尝试简单提取
      try {
        const hostMatch = url.match(/^https?:\/\/([^\/]+)/);
        if (hostMatch) {
          host = hostMatch[1];
        } else {
          host = 'unknown';
        }
      } catch (e) {
        console.warn('Failed to extract host from URL:', url, e);
        host = 'unknown';
      }
    }
  }

  // 从 content_type 解析 mimeType 和 extension
  let mimeType = requestMsg.content_type || '';
  let extension = '';
  
  // 更精确的扩展名推导
  if (mimeType.includes('html')) {
    extension = 'html';
  } else if (mimeType.includes('json')) {
    extension = 'json';
  } else if (mimeType.includes('xml')) {
    extension = 'xml';
  } else if (mimeType.includes('javascript') || mimeType.includes('js')) {
    extension = 'js';
  } else if (mimeType.includes('css')) {
    extension = 'css';
  } else if (mimeType.includes('image')) {
    if (mimeType.includes('png')) extension = 'png';
    else if (mimeType.includes('jpg') || mimeType.includes('jpeg')) extension = 'jpg';
    else if (mimeType.includes('gif')) extension = 'gif';
    else extension = 'img';
  } else {
    // 从 URL 路径尝试提取扩展名
    const pathExt = requestMsg.path.split('.').pop();
    if (pathExt && pathExt.length < 5 && !pathExt.includes('/') && !pathExt.includes('?')) {
      extension = pathExt;
    }
  }



  // 构建符合 HttpTrafficItem 和 ScanLogItem 的对象
  const scanLogItem: ScanLogItem = {
    // HttpTrafficItem 必需字段
    id: requestMsg.id,
    method: requestMsg.method,
    url: url,
    host: host,
    path: requestMsg.path,
    status: requestMsg.status,
    length: requestMsg.length,
    mimeType: mimeType,
    extension: extension,
    title: requestMsg.title || '',
    ip: requestMsg.ip || '',
    note: '',
    timestamp: requestMsg.timestamp,
    
    // HttpTrafficItem 可选字段
    selected: false,
    color: '',
    
    // ScanLogItem 扩展字段
    moduleName: requestMsg.module_name,
    target: requestMsg.target,
    contentType: requestMsg.content_type,
    serverDurationMs: 0,
    
    // HTTP详情（初始为 undefined，按需加载）
    request: undefined,
    response: undefined
  };

  console.log('Converted ScanLogItem:', scanLogItem);
  return scanLogItem;
} 