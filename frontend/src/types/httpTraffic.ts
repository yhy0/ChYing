// HTTP流量项目基础接口
export interface HttpTrafficItem {
  id: number;
  method: string;
  url: string;
  status?: number;
  timestamp: string;
  color?: string;
}

// HTTP流量表格列定义
export interface HttpTrafficColumn<T> {
  id: string;
  header: string;
  accessorKey?: keyof T;
  size?: number;
  minSize?: number;
  maxSize?: number;
  enableSorting?: boolean;
  enableResizing?: boolean;
}

// 授权测试结果接口
export interface AuthTestResult extends HttpTrafficItem {
  originalRequest: string;
  originalResponse: string;
  modifiedRequest: string;
  modifiedResponse: string;
  ruleID: number;
  ruleDescription: string;
  statusCode: number;
  originalStatus: number;
  sessionName: string;
}

// HTTP流量详细信息（包含请求响应内容）
export interface HttpTrafficDetail extends HttpTrafficItem {
  request?: string;        // 请求内容
  response?: string;       // 响应内容
  isLoading?: boolean;     // 是否正在加载
  serverDurationMs?: number; // 服务器响应时间
}

// 流量统计信息
export interface TrafficStats {
  total: number;   // 总流量数量
}
