/**
 * 扫描目标管理相关类型定义
 */

// 扫描目标状态
export type ScanTargetStatus = 'pending' | 'running' | 'completed' | 'failed' | 'cancelled' | 'paused';

// 扫描目标类型
export type ScanTargetType = 'single' | 'domain' | 'subnet' | 'cidr' | 'batch';

// 扫描配置
export interface ScanConfig {
  enable_port_scan: boolean;      // 端口扫描
  enable_dir_scan: boolean;       // 目录扫描
  enable_vuln_scan: boolean;      // 漏洞扫描
  enable_fingerprint: boolean;    // 指纹识别
  enable_crawler: boolean;        // 爬虫
  enable_xss: boolean;            // XSS扫描
  enable_sql: boolean;            // SQL注入扫描
  enable_bypass_403: boolean;     // 403绕过
  threads: number;                // 扫描线程数
  timeout: number;                // 超时时间(秒)
  custom_headers: string[];       // 自定义请求头
  exclude: string[];              // 排除规则
  port_range: string;             // 端口范围
  max_depth: number;              // 爬虫最大深度
}

// 扫描目标
export interface ScanTarget {
  id: number;
  name: string;                   // 目标名称
  type: ScanTargetType;           // 目标类型
  target: string;                 // 目标地址
  description: string;            // 描述
  status: ScanTargetStatus;       // 状态
  priority: number;               // 优先级 (1-10, 10最高)
  config: string;                 // 扫描配置JSON
  
  // 调度信息
  schedule_type: string;          // 调度类型
  schedule_time: string;          // 调度时间
  next_run_time?: string;         // 下次运行时间
  
  // 分配信息
  assigned_node: string;          // 分配的节点ID
  node_name: string;              // 节点名称
  
  // 执行信息
  started_at?: string;            // 开始时间
  completed_at?: string;          // 完成时间
  duration: number;               // 持续时间(秒)
  progress: number;               // 进度百分比
  
  // 结果统计
  found_hosts: number;            // 发现的主机数
  found_ports: number;            // 发现的端口数
  found_vulns: number;            // 发现的漏洞数
  found_dirs: number;             // 发现的目录数
  total_requests: number;         // 总请求数
  failed_requests: number;        // 失败请求数
  
  // 错误信息
  error_message: string;          // 错误信息
  last_error: string;             // 最后错误
  retry_count: number;            // 重试次数
  
  // 创建者信息
  created_by: string;             // 创建者
  created_from: string;           // 创建来源
  
  created_at: string;
  updated_at: string;
}

// 扫描目标查询参数
export interface ScanTargetQuery {
  status?: ScanTargetStatus;
  limit?: number;
  offset?: number;
  search?: string;
  type?: ScanTargetType;
  node_id?: string;
}

// 扫描目标列表响应
export interface ScanTargetListResponse {
  targets: ScanTarget[];
  total: number;
}

// 扫描统计信息
export interface ScanStatistics {
  by_status: Array<{
    status: ScanTargetStatus;
    count: number;
  }>;
  by_type: Array<{
    type: ScanTargetType;
    count: number;
  }>;
  total: number;
  today: number;
}

// 批量添加扫描目标参数
export interface BatchAddScanTargetsParams {
  targets: string[];
  target_type: ScanTargetType;
  created_by: string;
  config?: ScanConfig;
}

// 添加扫描目标参数
export interface AddScanTargetParams {
  name: string;
  type: ScanTargetType;
  target: string;
  description?: string;
  priority?: number;
  config?: ScanConfig;
  schedule_type?: string;
  schedule_time?: string;
  created_by?: string;
}

// 更新扫描目标参数
export interface UpdateScanTargetParams {
  id: number;
  name?: string;
  description?: string;
  priority?: number;
  config?: ScanConfig;
  schedule_type?: string;
  schedule_time?: string;
  status?: ScanTargetStatus;
}

// 扫描目标状态更新参数
export interface UpdateScanTargetStatusParams {
  id: number;
  status: ScanTargetStatus;
  message?: string;
}

// 正在运行的任务
export interface RunningTask {
  id: number;
  target: ScanTarget;
  start_time: string;
  duration: string;
}

// 调度器状态
export interface SchedulerStatus {
  running: boolean;
  node_id: string;
  node_name: string;
  max_concurrent: number;
  poll_interval: string;
  running_tasks: number;
  task_list: Array<{
    id: number;
    name: string;
    target: string;
    start_time: string;
    duration: string;
  }>;
}

// 状态选项
export const STATUS_OPTIONS: Array<{label: string; value: ScanTargetStatus}> = [
  { label: '待扫描', value: 'pending' },
  { label: '扫描中', value: 'running' },
  { label: '已完成', value: 'completed' },
  { label: '失败', value: 'failed' },
  { label: '已取消', value: 'cancelled' },
  { label: '暂停', value: 'paused' }
];

// 类型选项
export const TYPE_OPTIONS: Array<{label: string; value: ScanTargetType}> = [
  { label: '单个URL', value: 'single' },
  { label: '域名', value: 'domain' },
  { label: '子网', value: 'subnet' },
  { label: 'CIDR', value: 'cidr' },
  { label: '批量目标', value: 'batch' }
];

// 优先级选项
export const PRIORITY_OPTIONS = [
  { label: '最高', value: 10 },
  { label: '高', value: 8 },
  { label: '中', value: 5 },
  { label: '低', value: 3 },
  { label: '最低', value: 1 }
];

// 调度类型选项
export const SCHEDULE_TYPE_OPTIONS = [
  { label: '立即执行', value: 'once' },
  { label: '每日', value: 'daily' },
  { label: '每周', value: 'weekly' },
  { label: '每月', value: 'monthly' }
];

// 状态颜色映射
export const STATUS_COLOR_MAP: Record<ScanTargetStatus, string> = {
  pending: 'orange',
  running: 'blue',
  completed: 'green',
  failed: 'red',
  cancelled: 'gray',
  paused: 'purple'
};

// 状态文本映射
export const STATUS_TEXT_MAP: Record<ScanTargetStatus, string> = {
  pending: '待扫描',
  running: '扫描中',
  completed: '已完成',
  failed: '失败',
  cancelled: '已取消',
  paused: '暂停'
};

// 类型文本映射
export const TYPE_TEXT_MAP: Record<ScanTargetType, string> = {
  single: '单个URL',
  domain: '域名',
  subnet: '子网',
  cidr: 'CIDR',
  batch: '批量目标'
}; 
