/**
 * Claude AI Agent 类型定义
 * @author yhy
 * @since 2026/01/09
 */

// 对话消息
export interface ChatMessage {
  id: string;
  role: 'user' | 'assistant' | 'system';
  content: string;
  timestamp: string;
  toolUses?: ToolUse[];
}

// 工具调用
export interface ToolUse {
  id: string;
  name: string;
  input: Record<string, any>;
  status: 'pending' | 'confirmed' | 'rejected' | 'completed' | 'error';
  result?: string;
  error?: string;
}

// 会话
export interface Session {
  id: string;
  projectId: string;
  context?: AgentContext;
  createdAt: string;
  updatedAt: string;
  history: ChatMessage[];
}

// 代理上下文
export interface AgentContext {
  projectId: string;
  projectName?: string;
  selectedTrafficIds?: string[];
  selectedVulnIds?: string[];
  selectedFingerprints?: string[];
  customData?: string;
  autoCollect: boolean;
}

// 流式事件
export interface StreamEvent {
  type: 'text' | 'tool_use' | 'tool_result' | 'error' | 'done' | 'cost';
  content?: string;
  toolUse?: ToolUse;
  error?: string;
  sessionId?: string;
  // 费用相关字段 (type === 'cost' 时)
  costUSD?: number;
  inputTokens?: number;
  outputTokens?: number;
}

// 外部 MCP 服务器配置
export interface ExternalMCPServer {
  id: string;                 // 唯一标识
  name: string;               // 服务器名称
  type: 'sse' | 'stdio';      // 服务器类型
  enabled: boolean;           // 是否启用
  description?: string;       // 描述
  url?: string;               // SSE 模式的 URL
  headers?: Record<string, string>; // SSE 模式的请求头
  command?: string;           // STDIO 模式的命令
  args?: string[];            // STDIO 模式的参数
  env?: string[];             // STDIO 模式的环境变量
}

// MCP 配置
export interface MCPConfig {
  enabled: boolean;           // 是否启用内置 MCP 服务器
  mode: 'sse' | 'stdio';      // 传输模式
  port: number;               // 端口
  enabled_tools?: string[];   // 启用的工具
  disabled_tools?: string[];  // 禁用的工具
  external_servers?: ExternalMCPServer[]; // 外部 MCP 服务器
}

// Claude Code CLI 配置
export interface ClaudeConfig {
  cliPath?: string;           // Claude Code CLI 路径
  workDir?: string;           // 工作目录
  model: string;              // 模型名称
  maxTurns?: number;          // 最大对话轮数
  systemPrompt?: string;      // 系统提示词
  allowedTools?: string[];    // 允许的工具列表
  disallowedTools?: string[]; // 禁止的工具列表
  permissionMode?: string;    // 权限模式
  requireToolConfirm: boolean; // 是否需要工具确认
  apiKey?: string;            // API Key
  baseURL?: string;           // Base URL
  temperature?: number;       // Temperature
  mcp?: MCPConfig;            // MCP 配置
}

// 发送消息请求
export interface SendMessageRequest {
  sessionId: string;
  message: string;
  context?: AgentContext;
}

// 发送消息响应
export interface SendMessageResponse {
  sessionId: string;
  message?: ChatMessage;
  toolUses?: ToolUse[];
  error?: string;
}

// 工具确认请求
export interface ToolConfirmRequest {
  sessionId: string;
  toolUseId: string;
  confirmed: boolean;
}

// 会话列表项
export interface SessionListItem {
  sessionId: string;
  projectId: string;
  createdAt: string;
  updatedAt: string;
}

// 费用信息
export interface CostInfo {
  costUSD: number;
  inputTokens: number;
  outputTokens: number;
}

// Claude Store 状态
export interface ClaudeState {
  initialized: boolean;
  loading: boolean;
  streaming: boolean;
  currentSessionId: string | null;
  sessions: Map<string, Session>;
  messages: ChatMessage[];
  pendingToolUses: ToolUse[];
  config: ClaudeConfig | null;
  error: string | null;
  streamingContent: string;
  mcpServerUrl: string;
  costInfo: CostInfo | null;
}

// 工具信息（用于显示）
export interface ToolInfo {
  name: string;
  displayName: string;
  description: string;
  dangerous: boolean;
  icon: string;
}

// 预定义的工具信息
export const TOOL_INFO: Record<string, ToolInfo> = {
  get_http_history: {
    name: 'get_http_history',
    displayName: 'Get HTTP History',
    description: 'Retrieve HTTP traffic history from the proxy',
    dangerous: false,
    icon: 'bx-history'
  },
  get_http_detail: {
    name: 'get_http_detail',
    displayName: 'Get HTTP Detail',
    description: 'Get detailed request/response for a specific traffic record',
    dangerous: false,
    icon: 'bx-detail'
  },
  get_vulnerabilities: {
    name: 'get_vulnerabilities',
    displayName: 'Get Vulnerabilities',
    description: 'List discovered vulnerabilities',
    dangerous: false,
    icon: 'bx-bug'
  },
  get_fingerprints: {
    name: 'get_fingerprints',
    displayName: 'Get Fingerprints',
    description: 'Get technology fingerprints',
    dangerous: false,
    icon: 'bx-fingerprint'
  },
  get_site_map: {
    name: 'get_site_map',
    displayName: 'Get Site Map',
    description: 'Get the site map of discovered URLs',
    dangerous: false,
    icon: 'bx-sitemap'
  },
  get_collection_info: {
    name: 'get_collection_info',
    displayName: 'Get Collection Info',
    description: 'Get collected information (domains, IPs, emails, etc.)',
    dangerous: false,
    icon: 'bx-collection'
  },
  send_http_request: {
    name: 'send_http_request',
    displayName: 'Send HTTP Request',
    description: 'Send a custom HTTP request to a target',
    dangerous: true,
    icon: 'bx-send'
  },
  run_scan: {
    name: 'run_scan',
    displayName: 'Run Scan',
    description: 'Execute a security scan on a target',
    dangerous: true,
    icon: 'bx-radar'
  },
  analyze_target: {
    name: 'analyze_target',
    displayName: 'Analyze Target',
    description: 'Get a comprehensive analysis of a target host',
    dangerous: false,
    icon: 'bx-analyse'
  }
};

// 获取工具信息
export function getToolInfo(toolName: string): ToolInfo {
  return TOOL_INFO[toolName] || {
    name: toolName,
    displayName: toolName,
    description: 'Unknown tool',
    dangerous: false,
    icon: 'bx-cog'
  };
}
