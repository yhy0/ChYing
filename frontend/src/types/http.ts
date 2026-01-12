// /**
//  * HTTP 请求数据接口
//  */
// export interface RequestData {
//   headers: string;
//   body: string;
// }
//
// /**
//  * HTTP 响应数据接口
//  */
// export interface ResponseData {
//   headers: string;
//   body: string;
// }

/**
 * HTTP 流量数据基础接口
 * 用于代理历史和各模块的流量展示
 */
export interface HttpTrafficItem {
  id: number;
  method: string;
  url: string;
  host: string;
  path: string;
  status: number;
  length: number;
  mimeType: string;
  extension: string;
  title: string;
  ip: string;
  note: string;
  timestamp: string;
  selected?: boolean;
  color?: string;
  [key: string]: any; // 额外属性
}

/**
 * 代理历史项接口
 * 扩展自 HttpTrafficItem，添加请求和响应数据
 */
export interface ProxyHistoryItem extends HttpTrafficItem {
  request: string;
  response: string;
}

/**
 * 拦截项的数据结构
 */
export interface InterceptItem {
  id: string;
  sequence: number; // 拦截序号
  type: 'request' | 'response'; // 当前拦截阶段
  status?: number | 'sent'; // 响应状态码或请求状态
  method: string;
  url: string;
  host: string;
  path: string;
  request: string; // 原始请求数据
  response?: string; // 响应数据（仅在响应阶段有）
  timestamp: string;
}

export interface HttpHistoryData {
  id: number;
  method: string;
  full_url: string;
  host: string;
  path: string;
  status: string;
  length: string;
  mime_type: string;
  extension: string;
  title: string;
  ip: string;
  note: string;
}

export interface HttpMarkerData {
  id: number;
  color: string;
  note: string;
}

export interface HttpHistoryEvent {
  data: HttpHistoryData[];
}

export interface HttpMarkerEvent {
  data: HttpMarkerData[];
}

export interface IntruderSourceTarget {
  url: string;
  method: string;
  headers: string;
  body: string;
  fullRequest: string;
} 