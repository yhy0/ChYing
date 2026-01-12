/**
 * 扫描日志相关类型定义
 */

import type { HttpTrafficItem } from './http';

/**
 * 后端请求扫描消息接口（与后端 RequestScanMsg 对应）
 */
export interface RequestScanMsg {
  id: number;              // 请求ID（后端是 int64，前端接收为 number）
  module_name: string;     // 模块名称
  target: string;          // 目标URL
  path: string;            // 请求路径
  method: string;          // HTTP方法
  status: number;          // 状态码
  length: number;          // 响应长度
  title: string;           // 页面标题
  ip: string;              // IP地址
  content_type: string;    // 内容类型
  timestamp: string;       // 时间戳
}

/**
 * 前端请求详情接口（包含完整的请求响应数据）
 */
export interface RequestScanDetail {
  id: number;
  request_raw: string;     // 原始请求
  response_raw: string;    // 原始响应
}

/**
 * 扫描模块类型枚举（保持现有定义用于过滤）
 */
export enum ScanModuleType {
  XSS = 'xss',
  SQL_INJECTION = 'sql_injection', 
  DIRECTORY_TRAVERSAL = 'directory_traversal',
  FILE_UPLOAD = 'file_upload',
  XXE = 'xxe',
  SSRF = 'ssrf',
  COMMAND_INJECTION = 'command_injection',
  DESERIALIZATION = 'deserialization',
  SENSITIVE_INFO = 'sensitive_info',
  WAF_BYPASS = 'waf_bypass',
  BRUTE_FORCE = 'brute_force',
  FINGERPRINT = 'fingerprint',
  PORT_SCAN = 'port_scan',
  SUBDOMAIN = 'subdomain',
  DIRECTORY_SCAN = 'directory_scan',
  VULNERABILITY_SCAN = 'vulnerability_scan',
  RAW = 'raw',
  UPLOAD = 'upload',
  HTTP10 = 'http10',
  OTHER = 'other'
}



/**
 * 扫描日志项接口（扩展 HttpTrafficItem，适配前端显示）
 */
export interface ScanLogItem extends HttpTrafficItem {
  // 从 RequestScanMsg 映射而来的字段
  moduleName: string;          // 模块名称（来自 module_name）
  target: string;              // 目标URL（完整URL）
  contentType: string;         // 内容类型（来自 content_type）
  
  // 扩展字段（前端计算或默认值）
  moduleType?: ScanModuleType; // 模块类型（从 moduleName 推导）
  vulnerability?: string;      // 检测到的漏洞类型
  description?: string;        // 描述信息
  evidence?: string;           // 漏洞证据
  payload?: string;            // 攻击载荷
  
  // HTTP详情（按需加载）
  request?: string;            // 原始请求（通过API获取）
  response?: string;           // 原始响应（通过API获取）
  serverDurationMs?: number;   // 服务器响应时间
  
  // 其他
  tags?: string[];             // 标签
}

/**
 * 扫描统计信息
 */
export interface ScanStatistics {
  total: number;
  
  // 按模块分组统计
  byModule: Record<string, number>;
}

/**
 * 扫描过滤器配置
 */
export interface ScanLogFilter {
  moduleTypes: string[];               // 模块类型过滤（使用字符串而不是枚举）
  timeRange?: {                        // 时间范围
    start: string;
    end: string;
  };
  searchKeyword?: string;              // 搜索关键词
  targetHost?: string;                 // 目标主机过滤
  statusCodes?: number[];              // HTTP状态码过滤
}

/**
 * 扫描日志窗口配置
 */
export interface ScanLogWindowConfig {
  autoRefresh: boolean;                // 自动刷新
  refreshInterval: number;             // 刷新间隔（毫秒）
  maxLogCount: number;                 // 最大日志条数
  showStatistics: boolean;             // 显示统计信息
} 