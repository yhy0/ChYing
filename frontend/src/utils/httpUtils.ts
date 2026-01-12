/**
 * HTTP请求响应处理工具函数
 */

import { toRaw } from 'vue';
import type { ProxyHistoryItem, RepeaterTab, IntruderSourceTarget, HttpTrafficItem } from '../types'; // 确保导入 IntruderSourceTarget 和 HttpTrafficItem

/**
 * 检查对象是否为Proxy对象
 */
export function isProxy(obj: any): boolean {
  return obj && typeof obj === 'object' && 
    (obj.__v_isProxy || Object.hasOwnProperty.call(obj, '__v_isProxy'));
}

/**
 * 转换数据为普通对象（如果是Proxy）
 */
export function normalizeData<T extends string>(data: T): T {
  // 如果是Proxy对象，转换为普通对象；否则保持原样
  const rawData = isProxy(data) ? toRaw(data) : data;
  return rawData as T;
}

// 从完整HTTP文本中分离头部和主体
export const extractHeadersAndBody = (content: string): { headers: string, body: string } => {
  const parts = content.split('\n\n');
  return {
    headers: parts[0] || '',
    body: parts.slice(1).join('\n\n') || ''
  };
};

/**
 * 转换为十六进制显示
 */
export function convertToHex(text: string): string {
  let result = '';
  let asciiLine = '';
  let hexLine = '';
  let lineCount = 0;
  
  for (let i = 0; i < text.length; i++) {
    const charCode = text.charCodeAt(i);
    // 转换为2位十六进制并补零
    const hex = charCode.toString(16).padStart(2, '0');
    
    hexLine += hex + ' ';
    // 可显示ASCII则显示字符，否则显示点
    asciiLine += (charCode >= 32 && charCode <= 126) ? text[i] : '.';
    
    lineCount++;
    if (lineCount === 16 || i === text.length - 1) {
      // 填充空白使对齐
      while (lineCount < 16) {
        hexLine += '   ';
        lineCount++;
      }
      
      result += hexLine + '  ' + asciiLine + '\n';
      hexLine = '';
      asciiLine = '';
      lineCount = 0;
    }
  }
  
  return result;
}

/**
 * 获取响应的Content-Type
 */
export function getResponseContentType(headers: string): string {
  const match = headers.match(/Content-Type:\s*([^;\r\n]+)/i);
  return match ? match[1].trim().toLowerCase() : '';
}

/**
 * 检查是否为HTML内容
 */
export function isHtmlContent(contentType: string, body: string): boolean {
  if (contentType.includes('html')) return true;
  // 简单检查是否包含HTML标签
  return /<html|<!doctype html|<head|<body/i.test(body);
}

/**
 * 处理HTML内容的安全显示（防止XSS）
 */
export function sanitizeHtml(html: string): string {
  // 基本的安全处理
  return html
    .replace(/<script\b[^<]*(?:(?!<\/script>)<[^<]*)*<\/script>/gi, '<!-- script removed -->')
    .replace(/javascript:/gi, 'removed:');
}

// 新增函数定义
export interface RequestLineDetails {
  method: string;
  path: string; // 对于代理请求，这通常是包含协议和主机的完整URL；对于原始服务器请求，这通常是路径。
  httpVersion: string;
}

export function extractRequestLineDetails(requestLine: string): RequestLineDetails {
  // 正则表达式尝试匹配常见的请求行格式, e.g., GET /path HTTP/1.1 or CONNECT example.com:443 HTTP/1.1
  const match = requestLine.match(/^(\S+)\s+(\S+)\s+(HTTP\/\d\.\d)$/i);
  if (match) {
    return {
      method: match[1].toUpperCase(),
      path: match[2], // path 可能是一个完整的URL，也可能只是路径
      httpVersion: match[3].toUpperCase(),
    };
  }
  // 如果正则不匹配 (例如格式不标准或缺少HTTP版本)，则进行更简单的分割
  const parts = requestLine.split(' ');
  return {
    method: (parts[0] || 'GET').toUpperCase(),
    path: parts[1] || '/',
    httpVersion: (parts[2] || 'HTTP/1.1').toUpperCase(),
  };
}

// 新增函数: 从 ProxyHistoryItem 或 RepeaterTab 准备 IntruderSourceTarget
export function prepareIntruderSourceTarget(
  sourceItem: ProxyHistoryItem | RepeaterTab
): IntruderSourceTarget {
  let requestContent: string;
  let sourceUrl: string;
  let methodFromSource: string | undefined;

  // Type guard to differentiate ProxyHistoryItem and RepeaterTab
  // ProxyHistoryItem will have a `response` property (string | null) and extends HttpTrafficItem which has `url` and `method`
  // RepeaterTab also has `response` (string | null) but its primary identifier might be its specific properties like `modified` or lack of direct `url` that ProxyHistoryItem has.
  // A more robust way: ProxyHistoryItem directly has `url` and `method` from HttpTrafficItem.
  // RepeaterTab has `request` string but not necessarily direct `url` or `method` properties for the overall item.

  if ('url' in sourceItem && 'method' in sourceItem && typeof (sourceItem as HttpTrafficItem).url === 'string' && typeof (sourceItem as HttpTrafficItem).method === 'string' && 'response' in sourceItem && typeof (sourceItem as ProxyHistoryItem).response === 'string') {
    // Likely ProxyHistoryItem: it has url, method (from HttpTrafficItem) and a string response.
    const proxyItem = sourceItem as ProxyHistoryItem;
    requestContent = proxyItem.request;
    sourceUrl = proxyItem.url; // Use pre-parsed URL
    methodFromSource = proxyItem.method; // Use pre-parsed method
  } else {
    // Assume RepeaterTab or a similar structure that doesn't fit the strict ProxyHistoryItem check above
    const repeaterItem = sourceItem as RepeaterTab;
    requestContent = repeaterItem.request; // RepeaterTab does not have currentRequest
    // For RepeaterTab, URL and method must be parsed from its request string
    const firstLine = requestContent ? requestContent.split('\n')[0] || '' : '';
    const details = extractRequestLineDetails(firstLine);
    sourceUrl = details.path; // path from request line for Repeater acts as its URL analysis
    methodFromSource = details.method;
  }

  // If methodFromSource wasn't determined (e.g. malformed first line for repeater), parse it again or default.
  // However, extractRequestLineDetails called above for Repeater already sets it.
  // For Proxy, method is already available. If not, then parse from requestContent.
  const finalMethod = methodFromSource || (extractRequestLineDetails(requestContent ? requestContent.split('\n')[0] || '' : '')).method;
  const { headers, body } = extractHeadersAndBody(requestContent || '');

  return {
    url: sourceUrl,       
    method: finalMethod,  
    headers: headers,     
    body: body,           
    fullRequest: requestContent || '',
  };
} 