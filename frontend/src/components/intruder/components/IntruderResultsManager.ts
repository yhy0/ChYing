import { ref, computed } from 'vue';
import type { IntruderResult, AttackResult } from '../../../types/intruder';

// 用于请求响应数据的类型
interface RequestData {
  headers: string;
  body: string;
}

interface ResponseData {
  headers: string;
  body: string;
}

// 结果管理相关的组合式API
export function useIntruderResultsManager() {
  // 定义状态
  const selectedResult = ref<AttackResult | null>(null);
  const showRequestResponse = ref(true);
  const requestData = ref<RequestData>({ headers: '', body: '' });
  const responseData = ref<ResponseData>({ headers: '', body: '' });

  // 选择结果行
  const selectResult = (result: AttackResult) => {
    if (!result) {
      console.error('选择了空结果');
      return;
    }

    let request = result.request || '';
    let response = result.response || '';

    // 确保处理换行符
    if (typeof request === 'string' && request.includes('\\n')) {
      request = request.replace(/\\n/g, '\n');
    }

    if (typeof response === 'string' && response.includes('\\n')) {
      response = response.replace(/\\n/g, '\n');
    }

    // 先设置选定的结果
    selectedResult.value = { ...result, request, response };

    // 更新请求和响应数据
    // 解析请求
    const reqBlankLineIndex = typeof request === 'string' ? request.indexOf('\n\n') : -1;
    if (reqBlankLineIndex === -1) {
      requestData.value = {
        headers: typeof request === 'string' ? request.trim() : '',
        body: ''
      };
    } else {
      requestData.value = {
        headers: request.substring(0, reqBlankLineIndex).trim(),
        body: request.substring(reqBlankLineIndex + 2).trim()
      };
    }

    // 解析响应
    const respBlankLineIndex = typeof response === 'string' ? response.indexOf('\n\n') : -1;
    if (respBlankLineIndex === -1) {
      responseData.value = {
        headers: typeof response === 'string' ? response.trim() : '',
        body: ''
      };
    } else {
      responseData.value = {
        headers: response.substring(0, respBlankLineIndex).trim(),
        body: response.substring(respBlankLineIndex + 2).trim()
      };
    }

    // 确保显示面板
    showRequestResponse.value = true;
  };

  // 清空选中结果
  const clearSelectedResult = () => {
    selectedResult.value = null;
    requestData.value = { headers: '', body: '' };
    responseData.value = { headers: '', body: '' };
  };

  // 类型转换：确保selectedResult的类型与IntruderHistoryPanel组件兼容
  const selectedResultWithType = computed<IntruderResult | null>(() => {
    if (!selectedResult.value) return null;
    
    // 确保request和response是字符串
    let requestStr = '';
    let responseStr = '';
    
    if (typeof selectedResult.value.request === 'string') {
      requestStr = selectedResult.value.request;
    } else if (selectedResult.value.request) {
      // 如果是对象，则尝试转换为字符串
      const reqObj = selectedResult.value.request as any;
      requestStr = reqObj.headers ? reqObj.headers + (reqObj.body ? '\n\n' + reqObj.body : '') : '';
    }
    
    if (typeof selectedResult.value.response === 'string') {
      responseStr = selectedResult.value.response;
    } else if (selectedResult.value.response) {
      // 如果是对象，则尝试转换为字符串
      const respObj = selectedResult.value.response as any;
      responseStr = respObj.headers ? respObj.headers + (respObj.body ? '\n\n' + respObj.body : '') : '';
    }
    
    // 获取颜色属性，如果没有则使用undefined
    const colorValue = (selectedResult.value as any).color;
    
    return {
      id: Number(selectedResult.value.id || 0), // 确保ID是数字类型
      payload: Array.isArray(selectedResult.value.payload) 
        ? selectedResult.value.payload 
        : [String(selectedResult.value.payload || '')],
      status: selectedResult.value.status,
      length: selectedResult.value.length,
      timeMs: selectedResult.value.timeMs,
      timestamp: String(selectedResult.value.timestamp || ''),
      request: requestStr,
      response: responseStr,
      color: colorValue, // 使用提取的颜色值
      selected: true
    };
  });

  // 处理来自IntruderHistoryPanel的结果选择
  const handleSelectResult = (result: IntruderResult, allResults: AttackResult[]) => {
    // 找到原始的结果对象并选择它
    const originalResult = allResults.find(r => String(r.id) === String(result.id));
    if (originalResult) {
      selectResult(originalResult);
    }
  };

  // 处理结果右键菜单
  const handleResultContextMenu = (event: MouseEvent, result: IntruderResult, allResults: AttackResult[]) => {
    // 阻止默认的浏览器右键菜单
    event.preventDefault();
    
    // 找到原始的结果对象
    const originalResult = allResults.find(r => String(r.id) === String(result.id));
    if (originalResult) {
      selectResult(originalResult);
    }
  };

  // 为IntruderHistoryPanel组件准备格式化的结果数据
  const formatResults = (results: AttackResult[]): IntruderResult[] => {
    if (!results || results.length === 0) return [];
    
    // 强制创建新的数组引用，确保响应式系统能检测到变化
    return results.map(r => ({
      ...r,
      id: Number(r.id || 0), // 确保ID是数字类型
      timestamp: String(r.timestamp || ''),
      payload: Array.isArray(r.payload) ? r.payload : [String(r.payload || '')]
    }));
  };

  return {
    selectedResult,
    showRequestResponse,
    requestData,
    responseData,
    selectedResultWithType,
    selectResult,
    clearSelectedResult,
    handleSelectResult,
    handleResultContextMenu,
    formatResults
  };
} 