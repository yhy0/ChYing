/**
 * Project 相关类型定义
 */

// 数据库模式
export type DatabaseMode = 'local';

// 项目信息
export interface ProjectInfo {
  id: string;
  name: string;
  created_at: string;
  last_opened_at?: string;
  description?: string;
  database_mode: DatabaseMode;
  is_current?: boolean;
  // 兼容旧接口的字段
  requests?: number;
  source?: 'local';
  // 数据库信息
  database_file?: string;  // 本地SQLite数据库文件名
  database?: string;       // 数据库名称
  database_path?: string;  // 数据库文件完整路径
  // 数据库大小信息
  size_bytes?: number;     // 数据库文件大小（字节）
  size_formatted?: string; // 格式化的文件大小（如 "1.2 MB"）
  modified_time?: string;  // 文件修改时间
  file_exists?: boolean;   // 文件是否存在
  // 统计信息
  total_requests?: number; // 总请求数
  total_hosts?: number;    // 总主机数
  first_request?: string;  // 第一个请求时间
  last_request?: string;   // 最后一个请求时间
}

// 向后兼容的别名
export type Project = ProjectInfo;

// 项目创建请求
export interface CreateProjectRequest {
  project_id: string;
  project_name: string;
  database_mode: DatabaseMode;
}

// 项目详情
export interface ProjectDetails {
  name: string;
  createdDate: string;
  totalRequests: number;
  issuesFound: number;
  hosts: number;
  scanProgress: number;
}

// 安全问题
export interface SecurityIssue {
  id: number;
  name: string;
  severity: 'high' | 'medium' | 'low' | 'info';
  host: string;
  path: string;
  description: string;
  timestamp: string;
}

// HTTP历史记录项
export interface HttpHistoryItem {
  id: number;
  method: string;
  url: string;
  host: string;
  path: string;
  status: number;
  length: number;
  mimeType: string;
  timestamp: string;
}

// 站点地图节点
export interface SiteMapNode {
  id: number;
  name: string;
  path: string;
  fullUrl: string;
  nodeType: 'host' | 'directory' | 'file';
  icon: string;
  children: SiteMapNode[];
  isExpanded?: boolean;
  lastAccessed?: string;
  rawData?: any; // 存储节点的原始数据
}

// 项目设置
export interface ProjectSettings {
  name: string;
  scope: string[];
  scanOptions: {
    active: boolean;
    passive: boolean;
    spider: boolean;
    depth: number;
  };
  notifications: {
    email: boolean;
    slack: boolean;
    emailAddress: string;
    slackWebhook: string;
  };
} 
