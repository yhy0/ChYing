/**
 * 视图管理相关工具函数
 */

import { ref, computed } from 'vue';
import { convertToHex, getResponseContentType, isHtmlContent, sanitizeHtml, extractHeadersAndBody } from './httpUtils';

// 视图类型定义
export type RequestViewType = 'pretty' | 'raw' | 'hex';
export type ResponseViewType = 'pretty' | 'raw' | 'hex' | 'render';

/**
 * 请求视图管理
 */
export function useRequestViewManager(requestData: string) {
  // 请求视图类型
  const requestViewType = ref<RequestViewType>('pretty');

  // 设置请求视图类型
  const setRequestViewType = (type: RequestViewType) => {
    requestViewType.value = type;
  };

  // 获取当前视图下的请求内容
  const requestContent = computed(() => {
    switch (requestViewType.value) {
      case 'raw':
        return requestData;
      case 'hex':
        return convertToHex(requestData);
      case 'pretty':
      default:
        return requestData;
    }
  });

  return {
    requestViewType,
    setRequestViewType,
    requestContent
  };
}

/**
 * 响应视图管理
 */
export function useResponseViewManager(responseData: string) {
  // 响应视图类型
  const responseViewType = ref<ResponseViewType>('pretty');

  // 设置响应视图类型
  const setResponseViewType = (type: ResponseViewType) => {
    responseViewType.value = type;
  };

  // 获取当前视图下的响应内容
  const responseContent = computed(() => {
    switch (responseViewType.value) {
      case 'raw':
        return responseData;
      case 'hex':
        return convertToHex(responseData);
      case 'render':
        const newData = extractHeadersAndBody(responseData);
        const contentType = getResponseContentType(newData.headers);
        if (isHtmlContent(contentType, newData.body)) {
          return sanitizeHtml(newData.body);
        }
        return newData.body || '<!-- 无法渲染此内容类型 -->';
      case 'pretty':
      default:
        return responseData;
    }
  });

  // 检查响应是否为空
  const isResponseEmpty = computed(() => {
    return !responseData;
  });

  return {
    responseViewType,
    setResponseViewType,
    responseContent,
    isResponseEmpty
  };
} 