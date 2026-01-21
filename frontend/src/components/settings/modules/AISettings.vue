<script setup lang="ts">
import { ref, onMounted, computed } from 'vue';
import { useI18n } from 'vue-i18n';
// @ts-ignore
import { ClaudeUpdateConfig, ClaudeInitialize, ClaudeGetConfig, ClaudeIsInitialized, ClaudeTestConnection, ClaudeTestExternalMCPServer, A2AUpdateConfig, A2ATestConnection } from "../../../../bindings/github.com/yhy0/ChYing/app.js";

// Props
const props = defineProps<{
  compact?: boolean;  // 紧凑模式（用于模态框）
}>();

// Emits
const emit = defineEmits<{
  (e: 'saved'): void;  // 保存成功后触发
}>();

const { t } = useI18n();

// ==================== 类型定义 ====================
interface MCPTool {
  name: string;
  descKey: string;  // i18n key
  category: 'read' | 'analyze' | 'action';
}

interface ToolCategory {
  key: string;
  icon: string;
  labelKey: string;
  caution?: boolean;
}

// 外部 MCP 服务器接口（内部存储格式）
interface ExternalMCPServer {
  id: string;
  name: string;
  type: 'sse' | 'stdio';
  enabled: boolean;
  description: string;
  url: string;
  headers: Record<string, string>;
  command: string;
  args: string[];
  env: string[];
}

// Claude Desktop 风格的 MCP 服务器配置格式
interface MCPServerConfig {
  url?: string;
  headers?: Record<string, string>;
  command?: string;
  args?: string[];
  env?: Record<string, string>;
}

interface MCPServersJson {
  mcpServers: Record<string, MCPServerConfig>;
}

// ==================== Claude Code CLI 配置状态 ====================
// Agent 模式
const agentMode = ref<'claude-code' | 'a2a'>('claude-code');

const cliPath = ref('');
const workDir = ref('');
const model = ref('claude-sonnet-4');
const maxTurns = ref(0);
const systemPrompt = ref('');
const permissionMode = ref('default');
const requireToolConfirm = ref(true);

// 环境变量配置
const apiKey = ref('');
const baseURL = ref('');
const temperature = ref(0.7);

// MCP 配置状态
const mcpEnabled = ref(true);
const mcpMode = ref('sse');
const mcpPort = ref(0);
const mcpEnabledTools = ref<string[]>([]);
const mcpDisabledTools = ref<string[]>([]);
const mcpExternalServers = ref<ExternalMCPServer[]>([]);

// ==================== A2A 配置状态 ====================
const a2aEnabled = ref(false);
const a2aAgentURL = ref('');
const a2aHeaders = ref<Record<string, string>>({});
const a2aTimeout = ref(300);
const a2aEnableSSE = ref(true);
const a2aHeadersJson = ref('{}');
const a2aHeadersJsonError = ref('');
const isTestingA2A = ref(false);
const a2aAgentInfo = ref<any>(null);

// ==================== 常量定义 ====================
// 可用的 MCP 工具列表 (需要与 pkg/claude-code/mcp_server.go 保持同步)
const availableMCPTools: MCPTool[] = [
  { name: 'get_http_history', descKey: 'settings.ai.tool_get_http_history', category: 'read' },
  { name: 'get_traffic_detail', descKey: 'settings.ai.tool_get_traffic_detail', category: 'read' },
  { name: 'get_vulnerabilities', descKey: 'settings.ai.tool_get_vulnerabilities', category: 'read' },
  { name: 'search_traffic', descKey: 'settings.ai.tool_search_traffic', category: 'read' },
  { name: 'get_sitemap', descKey: 'settings.ai.tool_get_sitemap', category: 'read' },
  { name: 'get_statistics', descKey: 'settings.ai.tool_get_statistics', category: 'read' },
  { name: 'analyze_request', descKey: 'settings.ai.tool_analyze_request', category: 'analyze' },
  { name: 'send_http_request', descKey: 'settings.ai.tool_send_http_request', category: 'action' },
];

// 工具分类配置
const toolCategories: ToolCategory[] = [
  { key: 'read', icon: 'bx-book-reader', labelKey: 'settings.ai.mcp_tools_read' },
  { key: 'analyze', icon: 'bx-analyse', labelKey: 'settings.ai.mcp_tools_analyze' },
  { key: 'action', icon: 'bx-play', labelKey: 'settings.ai.mcp_tools_action', caution: true },
];

// 权限模式选项
const permissionModes = [
  { value: 'default', label: 'Default', description: 'Standard permission handling' },
  { value: 'plan', label: 'Plan Mode', description: 'Planning only, no execution' },
  { value: 'bypassPermissions', label: 'Bypass', description: 'Skip permission prompts' }
];

// ==================== UI 状态 ====================
const isLoading = ref(false);
const isSaving = ref(false);
const isTestingLLM = ref(false);
const isAutoInitializing = ref(false); // 自动初始化中
const testingMCPServerId = ref<string | null>(null); // 正在测试的外部 MCP 服务器 ID
const saveMessage = ref('');
const saveMessageType = ref<'success' | 'error'>('success');
const isInitialized = ref(false);
const initError = ref(''); // 初始化错误信息
const showMCPAdvanced = ref(false);
const showApiKey = ref(false);
const showExternalMCP = ref(false);
const mcpServersJsonText = ref(''); // JSON 文本框内容
const mcpServersJsonError = ref(''); // JSON 解析错误
let messageTimer: ReturnType<typeof setTimeout> | null = null;

// ==================== Computed 属性 ====================
// 是否已配置（基于 API Key 是否已设置）
const isConfigured = computed(() => !!apiKey.value && apiKey.value.trim().length > 0);

// 获取分类下的工具
const getToolsByCategory = (category: string) => {
  return availableMCPTools.filter(t => t.category === category);
};

// 启用的工具数量
const enabledToolCount = computed(() =>
  availableMCPTools.filter(t => isToolEnabled(t.name)).length
);

// ==================== 工具管理函数 ====================
// 检查工具是否启用
const isToolEnabled = (toolName: string): boolean => {
  // 如果在禁用列表中，则禁用
  if (mcpDisabledTools.value.includes(toolName)) {
    return false;
  }
  // 如果启用列表为空，则启用所有未禁用的工具
  if (mcpEnabledTools.value.length === 0) {
    return true;
  }
  // 如果启用列表不为空，只启用列表中的工具
  return mcpEnabledTools.value.includes(toolName);
};

// 切换工具启用状态
const toggleTool = (toolName: string) => {
  const currentlyEnabled = isToolEnabled(toolName);

  if (currentlyEnabled) {
    // 禁用工具：添加到禁用列表
    if (!mcpDisabledTools.value.includes(toolName)) {
      mcpDisabledTools.value.push(toolName);
    }
    // 从启用列表移除（如果存在）
    mcpEnabledTools.value = mcpEnabledTools.value.filter(t => t !== toolName);
  } else {
    // 启用工具：从禁用列表移除
    mcpDisabledTools.value = mcpDisabledTools.value.filter(t => t !== toolName);
    // 如果启用列表不为空，添加到启用列表
    if (mcpEnabledTools.value.length > 0 && !mcpEnabledTools.value.includes(toolName)) {
      mcpEnabledTools.value.push(toolName);
    }
  }
};

// 启用所有工具
const enableAllTools = () => {
  mcpEnabledTools.value = [];
  mcpDisabledTools.value = [];
};

// 禁用所有工具
const disableAllTools = () => {
  mcpDisabledTools.value = availableMCPTools.map(t => t.name);
  mcpEnabledTools.value = [];
};

// ==================== 外部 MCP 服务器管理 ====================
// 生成唯一 ID
const generateId = () => {
  return 'mcp-' + Date.now().toString(36) + Math.random().toString(36).substr(2, 9);
};

// 将内部格式转换为 Claude Desktop JSON 格式（用于显示）
const serversToJson = (servers: ExternalMCPServer[]): string => {
  if (servers.length === 0) {
    return JSON.stringify({ mcpServers: {} }, null, 2);
  }
  
  const mcpServers: Record<string, MCPServerConfig> = {};
  for (const server of servers) {
    const config: MCPServerConfig = {};
    
    if (server.type === 'sse') {
      config.url = server.url;
      if (Object.keys(server.headers || {}).length > 0) {
        config.headers = server.headers;
      }
    } else {
      config.command = server.command;
      if (server.args && server.args.length > 0) {
        config.args = server.args;
      }
      if (server.env && server.env.length > 0) {
        // 将 ["KEY=VALUE"] 格式转换为 {KEY: VALUE} 格式
        config.env = {};
        for (const envStr of server.env) {
          const [key, ...valueParts] = envStr.split('=');
          if (key) {
            config.env[key] = valueParts.join('=');
          }
        }
      }
    }
    
    mcpServers[server.name || server.id] = config;
  }
  
  return JSON.stringify({ mcpServers }, null, 2);
};

// 将 Claude Desktop JSON 格式转换为内部格式
const jsonToServers = (jsonText: string): ExternalMCPServer[] => {
  if (!jsonText.trim()) {
    return [];
  }
  
  const parsed = JSON.parse(jsonText);
  const mcpServers = parsed.mcpServers || parsed;
  
  const servers: ExternalMCPServer[] = [];
  
  for (const [name, config] of Object.entries(mcpServers)) {
    const serverConfig = config as MCPServerConfig;
    const server: ExternalMCPServer = {
      id: generateId(),
      name: name,
      type: serverConfig.url ? 'sse' : 'stdio',
      enabled: true,
      description: '',
      url: serverConfig.url || '',
      headers: serverConfig.headers || {},
      command: serverConfig.command || '',
      args: serverConfig.args || [],
      env: []
    };
    
    // 将 {KEY: VALUE} 格式转换为 ["KEY=VALUE"] 格式
    if (serverConfig.env && typeof serverConfig.env === 'object') {
      server.env = Object.entries(serverConfig.env).map(([k, v]) => `${k}=${v}`);
    }
    
    servers.push(server);
  }
  
  return servers;
};

// 验证并更新 JSON
const validateAndUpdateMcpJson = () => {
  mcpServersJsonError.value = '';
  
  if (!mcpServersJsonText.value.trim()) {
    mcpExternalServers.value = [];
    return;
  }
  
  try {
    const servers = jsonToServers(mcpServersJsonText.value);
    mcpExternalServers.value = servers;
  } catch (e) {
    mcpServersJsonError.value = e instanceof Error ? e.message : 'Invalid JSON format';
  }
};

// 同步 JSON 文本框（当从配置加载时）
const syncJsonFromServers = () => {
  mcpServersJsonText.value = serversToJson(mcpExternalServers.value);
};

// ==================== 消息显示函数 ====================
const showMessage = (message: string, type: 'success' | 'error', duration = 3000) => {
  // 清除之前的定时器
  if (messageTimer) {
    clearTimeout(messageTimer);
  }

  saveMessage.value = message;
  saveMessageType.value = type;

  // 错误消息显示更长时间
  const timeout = type === 'error' ? Math.max(duration, 5000) : duration;
  messageTimer = setTimeout(() => {
    saveMessage.value = '';
  }, timeout);
};

// ==================== 输入验证函数 ====================
const validateConfig = () => {
  // 验证 temperature
  if (temperature.value < 0) temperature.value = 0;
  if (temperature.value > 1) temperature.value = 1;

  // 验证 mcpPort
  if (mcpPort.value < 0) mcpPort.value = 0;
  if (mcpPort.value > 65535) mcpPort.value = 0;

  // 验证 maxTurns
  if (maxTurns.value < 0) maxTurns.value = 0;
};

// 加载配置
const loadConfig = async () => {
  isLoading.value = true;
  initError.value = '';
  try {
    // 检查初始化状态
    const initResult = await ClaudeIsInitialized();
    isInitialized.value = initResult?.data === true;

    // 获取配置
    const result = await ClaudeGetConfig();
    if (result?.data) {
      const config = typeof result.data === 'string' ? JSON.parse(result.data) : result.data;
      cliPath.value = config.cli_path || config.cliPath || '';
      workDir.value = config.work_dir || config.workDir || '';
      model.value = config.model || 'claude-sonnet-4';
      maxTurns.value = config.max_turns || config.maxTurns || 0;
      systemPrompt.value = config.system_prompt || config.systemPrompt || '';
      permissionMode.value = config.permission_mode || config.permissionMode || 'default';
      requireToolConfirm.value = config.require_tool_confirm ?? config.requireToolConfirm ?? true;

      // 加载环境变量配置
      apiKey.value = config.api_key || config.apiKey || '';
      baseURL.value = config.base_url || config.baseURL || '';
      temperature.value = config.temperature ?? 0.7;

      // 加载 MCP 配置
      if (config.mcp) {
        mcpEnabled.value = config.mcp.enabled ?? true;
        mcpMode.value = config.mcp.mode || 'sse';
        mcpPort.value = config.mcp.port || 0;
        mcpEnabledTools.value = config.mcp.enabled_tools || [];
        mcpDisabledTools.value = config.mcp.disabled_tools || [];
        mcpExternalServers.value = config.mcp.external_servers || [];
        // 同步 JSON 文本框
        syncJsonFromServers();
      }

      // 加载 A2A 配置
      agentMode.value = config.agent_mode || 'claude-code';
      if (config.a2a) {
        a2aEnabled.value = config.a2a.enabled ?? false;
        a2aAgentURL.value = config.a2a.agent_url || '';
        a2aHeaders.value = config.a2a.headers || {};
        a2aTimeout.value = config.a2a.timeout || 300;
        a2aEnableSSE.value = config.a2a.enable_sse ?? true;
        // 同步 headers JSON
        a2aHeadersJson.value = JSON.stringify(a2aHeaders.value, null, 2);
      }

      // 如果 API Key 已配置但未初始化，自动尝试初始化
      if (apiKey.value && !isInitialized.value) {
        await autoInitialize();
      }
    }
  } catch (error) {
    console.error('加载 AI 配置失败:', error);
  } finally {
    isLoading.value = false;
  }
};

// 自动初始化
const autoInitialize = async () => {
  isAutoInitializing.value = true;
  initError.value = '';
  try {
    const result = await ClaudeInitialize();
    if (result?.error) {
      initError.value = result.error;
      console.warn('自动初始化失败:', result.error);
    } else {
      isInitialized.value = true;
    }
  } catch (error: any) {
    initError.value = error.message || 'Initialization failed';
    console.warn('自动初始化失败:', error);
  } finally {
    isAutoInitializing.value = false;
  }
};

// 保存配置
const saveConfig = async () => {
  isSaving.value = true;
  saveMessage.value = '';

  try {
    // 构建 MCP 配置对象
    const mcpConfig = {
      enabled: mcpEnabled.value,
      mode: mcpMode.value,
      port: mcpPort.value,
      enabled_tools: mcpEnabledTools.value,
      disabled_tools: mcpDisabledTools.value,
      external_servers: mcpExternalServers.value.map(s => ({
        id: s.id,
        name: s.name,
        type: s.type,
        enabled: s.enabled,
        description: s.description,
        url: s.url,
        headers: s.headers,
        command: s.command,
        args: s.args,
        env: s.env
      }))
    };

    // 保存 Claude Code 配置
    const result = await ClaudeUpdateConfig(
      cliPath.value,
      workDir.value,
      model.value,
      maxTurns.value,
      systemPrompt.value,
      permissionMode.value,
      requireToolConfirm.value,
      apiKey.value,
      baseURL.value,
      temperature.value,
      mcpConfig
    );

    if (result?.error) {
      throw new Error(result.error);
    }

    // 保存 A2A 配置
    const a2aConfig = {
      enabled: a2aEnabled.value,
      agent_url: a2aAgentURL.value,
      headers: a2aHeaders.value,
      timeout: a2aTimeout.value,
      enable_sse: a2aEnableSSE.value
    };

    const a2aResult = await A2AUpdateConfig(agentMode.value, a2aConfig);
    if (a2aResult?.error) {
      throw new Error(a2aResult.error);
    }

    saveMessage.value = t('settings.ai.save_success', 'Settings saved successfully');
    saveMessageType.value = 'success';

    // 触发保存成功事件
    emit('saved');

    // 3秒后清除消息
    setTimeout(() => {
      saveMessage.value = '';
    }, 3000);
  } catch (error: any) {
    console.error('保存 AI 配置失败:', error);
    saveMessage.value = error.message || t('settings.ai.save_error', 'Failed to save settings');
    saveMessageType.value = 'error';
  } finally {
    isSaving.value = false;
  }
};

// 测试连接/初始化
const testConnection = async () => {
  isSaving.value = true;
  saveMessage.value = '';

  try {
    // 先保存配置
    await saveConfig();

    // 然后初始化
    const result = await ClaudeInitialize();

    if (result?.error) {
      throw new Error(result.error);
    }

    isInitialized.value = true;
    saveMessage.value = t('settings.ai.connection_success', 'Claude Code initialized successfully');
    saveMessageType.value = 'success';
  } catch (error: any) {
    console.error('初始化失败:', error);
    saveMessage.value = error.message || t('settings.ai.connection_error', 'Initialization failed');
    saveMessageType.value = 'error';
  } finally {
    isSaving.value = false;
  }
};

// 测试 LLM 连接
const testLLMConnection = async () => {
  isTestingLLM.value = true;
  saveMessage.value = '';

  try {
    // 先保存配置
    await saveConfig();

    // 测试 LLM 连接
    const result = await ClaudeTestConnection();

    if (result?.error) {
      throw new Error(result.error);
    }

    saveMessage.value = t('settings.ai.llm_test_success', 'LLM connection test successful');
    saveMessageType.value = 'success';
  } catch (error: any) {
    console.error('LLM 连接测试失败:', error);
    saveMessage.value = error.message || t('settings.ai.llm_test_error', 'LLM connection test failed');
    saveMessageType.value = 'error';
  } finally {
    isTestingLLM.value = false;
  }
};

// 测试外部 MCP 服务器连接
const testExternalMCPServer = async (server: ExternalMCPServer) => {
  testingMCPServerId.value = server.id;
  saveMessage.value = '';

  try {
    const result = await ClaudeTestExternalMCPServer(
      server.type,
      server.url,
      server.headers || {},
      server.command,
      server.args || []
    );

    if (result?.error) {
      throw new Error(result.error);
    }

    const data = result?.data;
    if (data?.success) {
      showMessage(data.message || t('settings.ai.mcp_test_success', 'MCP server connection successful'), 'success');
    } else {
      showMessage(data?.message || t('settings.ai.mcp_test_error', 'MCP server connection failed'), 'error');
    }
  } catch (error: any) {
    console.error('MCP 服务器测试失败:', error);
    showMessage(error.message || t('settings.ai.mcp_test_error', 'MCP server connection failed'), 'error');
  } finally {
    testingMCPServerId.value = null;
  }
};

// ==================== A2A 相关函数 ====================

// 验证并更新 A2A headers JSON
const validateAndUpdateA2AHeaders = () => {
  a2aHeadersJsonError.value = '';

  if (!a2aHeadersJson.value.trim() || a2aHeadersJson.value.trim() === '{}') {
    a2aHeaders.value = {};
    return;
  }

  try {
    const parsed = JSON.parse(a2aHeadersJson.value);
    if (typeof parsed !== 'object' || Array.isArray(parsed)) {
      throw new Error('Headers must be a JSON object');
    }
    a2aHeaders.value = parsed;
  } catch (e) {
    a2aHeadersJsonError.value = e instanceof Error ? e.message : 'Invalid JSON format';
  }
};

// 测试 A2A Agent 连接
const testA2AConnection = async () => {
  if (!a2aAgentURL.value) {
    showMessage('Agent URL is required', 'error');
    return;
  }

  isTestingA2A.value = true;
  saveMessage.value = '';

  try {
    const result = await A2ATestConnection(a2aAgentURL.value, a2aHeaders.value);

    if (result?.error) {
      throw new Error(result.error);
    }

    const data = result?.data;
    if (data?.success) {
      a2aAgentInfo.value = data.agent;
      showMessage(data.message || 'A2A Agent connection successful', 'success');
    } else {
      showMessage(data?.message || 'A2A Agent connection failed', 'error');
    }
  } catch (error: any) {
    console.error('A2A 连接测试失败:', error);
    showMessage(error.message || 'A2A Agent connection failed', 'error');
  } finally {
    isTestingA2A.value = false;
  }
};

// 组件挂载时加载配置
onMounted(() => {
  loadConfig();
});
</script>

<template>
  <div class="space-y-6 max-w-3xl pb-16">
    <!-- 标题（非紧凑模式显示） -->
    <template v-if="!props.compact">
      <h2 class="text-lg font-medium mb-5 flex items-center">
        <i class="bx bx-bot text-indigo-500 mr-2"></i>
        {{ t('settings.ai.title', 'Claude Code AI') }}
      </h2>

      <p class="text-sm text-gray-500 dark:text-gray-400 -mt-4 mb-6">
        {{ t('settings.ai.description', 'Configure Claude Code CLI for AI-powered security analysis and automation.') }}
      </p>
    </template>

    <!-- 加载状态 -->
    <div v-if="isLoading" class="flex items-center justify-center py-12">
      <i class="bx bx-loader-alt bx-spin text-3xl text-indigo-500"></i>
    </div>

    <div v-else class="space-y-5">
      <!-- 状态卡片 -->
      <div
        class="rounded-lg p-4 border flex items-center gap-4"
        :class="initError
          ? 'bg-red-50 dark:bg-red-900/20 border-red-200 dark:border-red-800'
          : (isConfigured
            ? (isInitialized ? 'bg-green-50 dark:bg-green-900/20 border-green-200 dark:border-green-800' : 'bg-blue-50 dark:bg-blue-900/20 border-blue-200 dark:border-blue-800')
            : 'bg-yellow-50 dark:bg-yellow-900/20 border-yellow-200 dark:border-yellow-800')"
      >
        <div class="text-3xl" :class="initError ? 'text-red-500' : (isConfigured ? (isInitialized ? 'text-green-500' : 'text-blue-500') : 'text-yellow-500')">
          <i :class="['bx', isAutoInitializing ? 'bx-loader-alt bx-spin' : (initError ? 'bx-error-circle' : (isConfigured ? (isInitialized ? 'bx-check-circle' : 'bx-cog') : 'bx-info-circle'))]"></i>
        </div>
        <div class="flex-1">
          <div class="font-medium" :class="initError ? 'text-red-700 dark:text-red-400' : (isConfigured ? (isInitialized ? 'text-green-700 dark:text-green-400' : 'text-blue-700 dark:text-blue-400') : 'text-yellow-700 dark:text-yellow-400')">
            {{ isAutoInitializing 
              ? t('settings.ai.initializing', 'Initializing...')
              : (initError 
                ? t('settings.ai.init_failed', 'Initialization Failed')
                : (isConfigured 
                  ? (isInitialized ? t('settings.ai.configured', 'Claude Code Ready') : t('settings.ai.configured_not_init', 'API Key Configured'))
                  : t('settings.ai.not_configured', 'API Not Configured')))
            }}
          </div>
          <div class="text-sm" :class="initError ? 'text-red-600 dark:text-red-500' : (isConfigured ? (isInitialized ? 'text-green-600 dark:text-green-500' : 'text-blue-600 dark:text-blue-500') : 'text-yellow-600 dark:text-yellow-500')">
            {{ isAutoInitializing
              ? t('settings.ai.initializing_desc', 'Please wait...')
              : (initError 
                ? initError
                : (isConfigured
                  ? (isInitialized ? t('settings.ai.configured_desc', 'Claude Code CLI is ready to use') : t('settings.ai.configured_not_init_desc', 'Click "Initialize" to start using Claude Code'))
                  : t('settings.ai.not_configured_desc', 'Please enter API Key to enable AI features')))
            }}
          </div>
        </div>
      </div>

      <!-- Agent 模式选择 -->
      <div class="bg-white dark:bg-[#282838] rounded-lg p-4 shadow-sm border border-gray-100 dark:border-gray-700">
        <h3 class="text-sm font-medium mb-3 text-gray-700 dark:text-gray-300 flex items-center">
          <i class="bx bx-bot mr-2 text-gray-400"></i>
          {{ t('settings.ai.agent_mode', 'Agent Mode') }}
        </h3>

        <div class="grid grid-cols-2 gap-3">
          <!-- Claude Code CLI 选项 -->
          <label
            class="relative flex items-start p-3 rounded-lg border-2 cursor-pointer transition-all"
            :class="agentMode === 'claude-code'
              ? 'border-indigo-500 bg-indigo-50 dark:bg-indigo-900/20'
              : 'border-gray-200 dark:border-gray-600 hover:border-gray-300 dark:hover:border-gray-500'"
          >
            <input
              type="radio"
              v-model="agentMode"
              value="claude-code"
              class="sr-only"
            />
            <div class="flex-1">
              <div class="flex items-center gap-2">
                <i class="bx bx-terminal text-lg" :class="agentMode === 'claude-code' ? 'text-indigo-500' : 'text-gray-400'"></i>
                <span class="font-medium text-sm" :class="agentMode === 'claude-code' ? 'text-indigo-700 dark:text-indigo-400' : 'text-gray-700 dark:text-gray-300'">
                  Claude Code CLI
                </span>
              </div>
              <p class="text-xs text-gray-500 dark:text-gray-400 mt-1">
                {{ t('settings.ai.claude_code_desc', 'Use local Claude Code CLI') }}
              </p>
            </div>
            <div v-if="agentMode === 'claude-code'" class="absolute top-2 right-2">
              <i class="bx bx-check-circle text-indigo-500"></i>
            </div>
          </label>

          <!-- A2A Agent 选项 -->
          <label
            class="relative flex items-start p-3 rounded-lg border-2 cursor-pointer transition-all"
            :class="agentMode === 'a2a'
              ? 'border-indigo-500 bg-indigo-50 dark:bg-indigo-900/20'
              : 'border-gray-200 dark:border-gray-600 hover:border-gray-300 dark:hover:border-gray-500'"
          >
            <input
              type="radio"
              v-model="agentMode"
              value="a2a"
              class="sr-only"
            />
            <div class="flex-1">
              <div class="flex items-center gap-2">
                <i class="bx bx-globe text-lg" :class="agentMode === 'a2a' ? 'text-indigo-500' : 'text-gray-400'"></i>
                <span class="font-medium text-sm" :class="agentMode === 'a2a' ? 'text-indigo-700 dark:text-indigo-400' : 'text-gray-700 dark:text-gray-300'">
                  A2A Agent
                </span>
              </div>
              <p class="text-xs text-gray-500 dark:text-gray-400 mt-1">
                {{ t('settings.ai.a2a_desc', 'Connect to remote A2A agent') }}
              </p>
            </div>
            <div v-if="agentMode === 'a2a'" class="absolute top-2 right-2">
              <i class="bx bx-check-circle text-indigo-500"></i>
            </div>
          </label>
        </div>
      </div>

      <!-- A2A Agent 配置 (仅在 A2A 模式下显示) -->
      <template v-if="agentMode === 'a2a'">
        <div class="bg-white dark:bg-[#282838] rounded-lg p-4 shadow-sm border border-gray-100 dark:border-gray-700">
          <h3 class="text-sm font-medium mb-3 text-gray-700 dark:text-gray-300 flex items-center">
            <i class="bx bx-globe mr-2 text-gray-400"></i>
            {{ t('settings.ai.a2a_config', 'A2A Agent Configuration') }}
          </h3>

          <div class="space-y-4">
            <!-- Agent URL -->
            <div>
              <label class="block text-xs font-medium text-gray-600 dark:text-gray-400 mb-1">
                {{ t('settings.ai.a2a_url', 'Agent URL') }}
                <span class="text-red-500">*</span>
              </label>
              <input
                type="text"
                v-model="a2aAgentURL"
                :placeholder="t('settings.ai.a2a_url_placeholder', 'https://my-agent.example.com')"
                class="w-full px-3 py-2 bg-white dark:bg-[#323248] border border-gray-300 dark:border-gray-600 rounded-md shadow-sm text-sm focus:outline-none focus:ring-2 focus:ring-[#4f46e5]"
                spellcheck="false"
              />
              <p class="text-xs text-gray-500 dark:text-gray-400 mt-1">
                {{ t('settings.ai.a2a_url_hint', 'The base URL of the A2A agent (Agent Card will be fetched from /.well-known/agent.json)') }}
              </p>
            </div>

            <!-- Custom Headers -->
            <div>
              <label class="block text-xs font-medium text-gray-600 dark:text-gray-400 mb-1">
                {{ t('settings.ai.a2a_headers', 'Custom Headers (JSON)') }}
              </label>
              <textarea
                v-model="a2aHeadersJson"
                @blur="validateAndUpdateA2AHeaders"
                :placeholder='`{\n  "Authorization": "Bearer YOUR_TOKEN"\n}`'
                rows="3"
                class="w-full px-3 py-2 bg-white dark:bg-[#323248] border rounded-md shadow-sm text-sm font-mono focus:outline-none focus:ring-2 focus:ring-[#4f46e5] resize-y"
                :class="a2aHeadersJsonError
                  ? 'border-red-300 dark:border-red-600'
                  : 'border-gray-300 dark:border-gray-600'"
                spellcheck="false"
              ></textarea>
              <div v-if="a2aHeadersJsonError" class="mt-1 text-xs text-red-600 dark:text-red-400 flex items-center gap-1">
                <i class="bx bx-error-circle"></i>
                {{ a2aHeadersJsonError }}
              </div>
            </div>

            <!-- Timeout -->
            <div>
              <label class="block text-xs font-medium text-gray-600 dark:text-gray-400 mb-1">
                {{ t('settings.ai.a2a_timeout', 'Timeout (seconds)') }}
              </label>
              <input
                type="number"
                v-model.number="a2aTimeout"
                min="30"
                max="600"
                class="w-32 px-3 py-2 bg-white dark:bg-[#323248] border border-gray-300 dark:border-gray-600 rounded-md shadow-sm text-sm focus:outline-none focus:ring-2 focus:ring-[#4f46e5]"
              />
            </div>

            <!-- Enable SSE -->
            <div class="flex items-center">
              <input
                id="a2aEnableSSE"
                v-model="a2aEnableSSE"
                type="checkbox"
                class="w-4 h-4 text-[#4f46e5] border-gray-300 rounded focus:ring-[#4f46e5] dark:focus:ring-[#4f46e5] dark:ring-offset-gray-800 focus:ring-2 dark:bg-gray-700 dark:border-gray-600"
              >
              <label for="a2aEnableSSE" class="ml-2 text-sm text-gray-700 dark:text-gray-300">
                {{ t('settings.ai.a2a_enable_sse', 'Enable streaming (SSE)') }}
              </label>
            </div>

            <!-- Test Connection Button -->
            <div class="pt-2">
              <button
                type="button"
                @click="testA2AConnection"
                :disabled="isTestingA2A || !a2aAgentURL"
                class="px-4 py-2 bg-emerald-500 hover:bg-emerald-600 text-white rounded-md text-sm transition-colors flex items-center gap-2 disabled:opacity-50 disabled:cursor-not-allowed"
              >
                <i :class="['bx', isTestingA2A ? 'bx-loader-alt bx-spin' : 'bx-plug']"></i>
                {{ isTestingA2A ? t('common.testing', 'Testing...') : t('settings.ai.a2a_test', 'Test Connection') }}
              </button>
            </div>

            <!-- Agent Info (显示连接成功后的信息) -->
            <div v-if="a2aAgentInfo" class="mt-3 p-3 bg-green-50 dark:bg-green-900/20 rounded-lg border border-green-200 dark:border-green-800">
              <div class="text-sm font-medium text-green-700 dark:text-green-400 mb-2">
                <i class="bx bx-check-circle mr-1"></i>
                {{ t('settings.ai.a2a_connected', 'Connected to Agent') }}
              </div>
              <div class="text-xs text-green-600 dark:text-green-500 space-y-1">
                <div><strong>Name:</strong> {{ a2aAgentInfo.name }}</div>
                <div v-if="a2aAgentInfo.description"><strong>Description:</strong> {{ a2aAgentInfo.description }}</div>
                <div v-if="a2aAgentInfo.version"><strong>Version:</strong> {{ a2aAgentInfo.version }}</div>
                <div v-if="a2aAgentInfo.capabilities?.length">
                  <strong>Capabilities:</strong> {{ a2aAgentInfo.capabilities.join(', ') }}
                </div>
              </div>
            </div>
          </div>
        </div>
      </template>

      <!-- Claude Code CLI 配置 (仅在 Claude Code 模式下显示) -->
      <template v-if="agentMode === 'claude-code'">

      <!-- CLI Path -->
      <div class="bg-white dark:bg-[#282838] rounded-lg p-4 shadow-sm border border-gray-100 dark:border-gray-700">
        <h3 class="text-sm font-medium mb-3 text-gray-700 dark:text-gray-300 flex items-center">
          <i class="bx bx-terminal mr-2 text-gray-400"></i>
          {{ t('settings.ai.cli_path', 'Claude Code CLI Path') }}
          <span class="text-gray-400 text-xs ml-2">({{ t('common.optional', 'Optional') }})</span>
        </h3>

        <input
          type="text"
          v-model="cliPath"
          :placeholder="t('settings.ai.cli_path_placeholder', 'Leave empty to use default (claude)')"
          class="w-full px-3 py-2 bg-white dark:bg-[#323248] border border-gray-300 dark:border-gray-600 rounded-md shadow-sm text-sm focus:outline-none focus:ring-2 focus:ring-[#4f46e5]"
          spellcheck="false"
        />

        <p class="text-xs text-gray-500 dark:text-gray-400 mt-2">
          {{ t('settings.ai.cli_path_hint', 'Path to Claude Code CLI executable. Leave empty to use system PATH.') }}
        </p>
      </div>

      <!-- Work Directory -->
      <div class="bg-white dark:bg-[#282838] rounded-lg p-4 shadow-sm border border-gray-100 dark:border-gray-700">
        <h3 class="text-sm font-medium mb-3 text-gray-700 dark:text-gray-300 flex items-center">
          <i class="bx bx-folder mr-2 text-gray-400"></i>
          {{ t('settings.ai.work_dir', 'Working Directory') }}
          <span class="text-gray-400 text-xs ml-2">({{ t('common.optional', 'Optional') }})</span>
        </h3>

        <input
          type="text"
          v-model="workDir"
          :placeholder="t('settings.ai.work_dir_placeholder', 'Leave empty to use current directory')"
          class="w-full px-3 py-2 bg-white dark:bg-[#323248] border border-gray-300 dark:border-gray-600 rounded-md shadow-sm text-sm focus:outline-none focus:ring-2 focus:ring-[#4f46e5]"
          spellcheck="false"
        />

        <p class="text-xs text-gray-500 dark:text-gray-400 mt-2">
          {{ t('settings.ai.work_dir_hint', 'Working directory for Claude Code operations.') }}
        </p>
      </div>

      <!-- Model Selection -->
      <div class="bg-white dark:bg-[#282838] rounded-lg p-4 shadow-sm border border-gray-100 dark:border-gray-700">
        <h3 class="text-sm font-medium mb-3 text-gray-700 dark:text-gray-300 flex items-center">
          <i class="bx bx-chip mr-2 text-gray-400"></i>
          {{ t('settings.ai.model', 'Model') }}
        </h3>

        <input
          type="text"
          v-model="model"
          :placeholder="t('settings.ai.model_placeholder', 'e.g., claude-sonnet-4')"
          class="w-full px-3 py-2 bg-white dark:bg-[#323248] border border-gray-300 dark:border-gray-600 rounded-md shadow-sm text-sm focus:outline-none focus:ring-2 focus:ring-[#4f46e5]"
          spellcheck="false"
        />

        <p class="text-xs text-gray-500 dark:text-gray-400 mt-2">
          {{ t('settings.ai.model_hint', 'Enter the Claude model name for AI conversations.') }}
        </p>
      </div>

      <!-- Environment Variables Section -->
      <div class="bg-white dark:bg-[#282838] rounded-lg p-4 shadow-sm border border-gray-100 dark:border-gray-700">
        <h3 class="text-sm font-medium mb-3 text-gray-700 dark:text-gray-300 flex items-center">
          <i class="bx bx-key mr-2 text-gray-400"></i>
          {{ t('settings.ai.env_config', 'API Configuration') }}
        </h3>

        <div class="space-y-4">
          <!-- API Key -->
          <div>
            <label class="block text-xs font-medium text-gray-600 dark:text-gray-400 mb-1">
              {{ t('settings.ai.api_key', 'API Key') }}
              <span class="text-gray-400 font-normal">(ANTHROPIC_API_KEY)</span>
            </label>
            <div class="relative">
              <input
                :type="showApiKey ? 'text' : 'password'"
                v-model="apiKey"
                :placeholder="t('settings.ai.api_key_placeholder', 'sk-ant-...')"
                class="w-full px-3 py-2 pr-10 bg-white dark:bg-[#323248] border border-gray-300 dark:border-gray-600 rounded-md shadow-sm text-sm focus:outline-none focus:ring-2 focus:ring-[#4f46e5]"
                spellcheck="false"
                autocomplete="off"
              />
              <button
                type="button"
                @click="showApiKey = !showApiKey"
                class="absolute right-2 top-1/2 -translate-y-1/2 text-gray-400 hover:text-gray-600 dark:hover:text-gray-300"
              >
                <i :class="['bx', showApiKey ? 'bx-hide' : 'bx-show']"></i>
              </button>
            </div>
            <p class="text-xs text-gray-500 dark:text-gray-400 mt-1">
              {{ t('settings.ai.api_key_hint', 'Your Anthropic API key for authentication.') }}
            </p>
          </div>

          <!-- Base URL -->
          <div>
            <label class="block text-xs font-medium text-gray-600 dark:text-gray-400 mb-1">
              {{ t('settings.ai.base_url', 'Base URL') }}
              <span class="text-gray-400 font-normal">(ANTHROPIC_BASE_URL)</span>
            </label>
            <input
              type="text"
              v-model="baseURL"
              :placeholder="t('settings.ai.base_url_placeholder', 'https://api.anthropic.com')"
              class="w-full px-3 py-2 bg-white dark:bg-[#323248] border border-gray-300 dark:border-gray-600 rounded-md shadow-sm text-sm focus:outline-none focus:ring-2 focus:ring-[#4f46e5]"
              spellcheck="false"
            />
            <p class="text-xs text-gray-500 dark:text-gray-400 mt-1">
              {{ t('settings.ai.base_url_hint', 'Custom API endpoint URL. Leave empty for default.') }}
            </p>
          </div>

          <!-- Temperature -->
          <div>
            <label class="block text-xs font-medium text-gray-600 dark:text-gray-400 mb-1">
              {{ t('settings.ai.temperature', 'Temperature') }}
              <span class="text-gray-400 font-normal">(AI_TEMPERATURE)</span>
            </label>
            <div class="flex items-center space-x-3">
              <input
                type="range"
                v-model.number="temperature"
                min="0"
                max="1"
                step="0.1"
                class="flex-1 h-2 bg-gray-200 dark:bg-gray-600 rounded-lg appearance-none cursor-pointer accent-[#4f46e5]"
              />
              <input
                type="number"
                v-model.number="temperature"
                min="0"
                max="1"
                step="0.1"
                class="w-20 px-2 py-1 bg-white dark:bg-[#323248] border border-gray-300 dark:border-gray-600 rounded-md shadow-sm text-sm text-center focus:outline-none focus:ring-2 focus:ring-[#4f46e5]"
              />
            </div>
            <p class="text-xs text-gray-500 dark:text-gray-400 mt-1">
              {{ t('settings.ai.temperature_hint', 'Controls randomness (0.0 = deterministic, 1.0 = creative).') }}
            </p>
          </div>
        </div>
      </div>

      <!-- Max Turns -->
      <div class="bg-white dark:bg-[#282838] rounded-lg p-4 shadow-sm border border-gray-100 dark:border-gray-700">
        <h3 class="text-sm font-medium mb-3 text-gray-700 dark:text-gray-300 flex items-center">
          <i class="bx bx-refresh mr-2 text-gray-400"></i>
          {{ t('settings.ai.max_turns', 'Max Turns') }}
        </h3>

        <div class="flex items-center space-x-2">
          <input
            type="number"
            v-model.number="maxTurns"
            min="0"
            max="100"
            step="1"
            class="w-32 px-3 py-2 bg-white dark:bg-[#323248] border border-gray-300 dark:border-gray-600 rounded-md shadow-sm text-sm focus:outline-none focus:ring-2 focus:ring-[#4f46e5]"
          />
          <span class="text-sm text-gray-500 dark:text-gray-400">
            {{ maxTurns === 0 ? t('settings.ai.unlimited', '(Unlimited)') : 'turns' }}
          </span>
        </div>

        <p class="text-xs text-gray-500 dark:text-gray-400 mt-2">
          {{ t('settings.ai.max_turns_hint', 'Maximum conversation turns per session. 0 = unlimited.') }}
        </p>
      </div>

      <!-- Permission Mode -->
      <div class="bg-white dark:bg-[#282838] rounded-lg p-4 shadow-sm border border-gray-100 dark:border-gray-700">
        <h3 class="text-sm font-medium mb-3 text-gray-700 dark:text-gray-300 flex items-center">
          <i class="bx bx-lock-alt mr-2 text-gray-400"></i>
          {{ t('settings.ai.permission_mode', 'Permission Mode') }}
        </h3>

        <select
          v-model="permissionMode"
          class="w-full px-3 py-2 bg-white dark:bg-[#323248] border border-gray-300 dark:border-gray-600 rounded-md shadow-sm text-sm focus:outline-none focus:ring-2 focus:ring-[#4f46e5]"
        >
          <option v-for="mode in permissionModes" :key="mode.value" :value="mode.value">
            {{ mode.label }} - {{ mode.description }}
          </option>
        </select>

        <p class="text-xs text-gray-500 dark:text-gray-400 mt-2">
          {{ t('settings.ai.permission_mode_hint', 'Controls how Claude Code handles tool permissions.') }}
        </p>
      </div>

      <!-- System Prompt -->
      <div class="bg-white dark:bg-[#282838] rounded-lg p-4 shadow-sm border border-gray-100 dark:border-gray-700">
        <h3 class="text-sm font-medium mb-3 text-gray-700 dark:text-gray-300 flex items-center">
          <i class="bx bx-message-square-detail mr-2 text-gray-400"></i>
          {{ t('settings.ai.system_prompt', 'System Prompt') }}
          <span class="text-gray-400 text-xs ml-2">({{ t('common.optional', 'Optional') }})</span>
        </h3>

        <textarea
          v-model="systemPrompt"
          :placeholder="t('settings.ai.system_prompt_placeholder', 'Custom system prompt for Claude...')"
          rows="4"
          class="w-full px-3 py-2 bg-white dark:bg-[#323248] border border-gray-300 dark:border-gray-600 rounded-md shadow-sm text-sm focus:outline-none focus:ring-2 focus:ring-[#4f46e5] resize-y"
          spellcheck="false"
        ></textarea>

        <p class="text-xs text-gray-500 dark:text-gray-400 mt-2">
          {{ t('settings.ai.system_prompt_hint', 'Additional instructions for Claude. Security context is automatically included.') }}
        </p>
      </div>

      <!-- Tool Confirmation -->
      <div class="bg-white dark:bg-[#282838] rounded-lg p-4 shadow-sm border border-gray-100 dark:border-gray-700">
        <h3 class="text-sm font-medium mb-3 text-gray-700 dark:text-gray-300 flex items-center">
          <i class="bx bx-shield-quarter mr-2 text-gray-400"></i>
          {{ t('settings.ai.tool_confirm', 'Tool Confirmation') }}
        </h3>

        <div class="flex items-center">
          <input
            id="requireToolConfirm"
            v-model="requireToolConfirm"
            type="checkbox"
            class="w-4 h-4 text-[#4f46e5] border-gray-300 rounded focus:ring-[#4f46e5] dark:focus:ring-[#4f46e5] dark:ring-offset-gray-800 focus:ring-2 dark:bg-gray-700 dark:border-gray-600"
          >
          <label for="requireToolConfirm" class="ml-2 text-sm text-gray-700 dark:text-gray-300">
            {{ t('settings.ai.require_confirm', 'Require confirmation for dangerous operations') }}
          </label>
        </div>

        <p class="text-xs text-gray-500 dark:text-gray-400 mt-2 ml-6">
          {{ t('settings.ai.tool_confirm_hint', 'When enabled, you will be asked to confirm before executing scans or sending requests.') }}
        </p>
      </div>

      <!-- MCP Server Configuration -->
      <div class="bg-white dark:bg-[#282838] rounded-lg p-4 shadow-sm border border-gray-100 dark:border-gray-700">
        <h3 class="text-sm font-medium mb-3 text-gray-700 dark:text-gray-300 flex items-center">
          <i class="bx bx-server mr-2 text-gray-400"></i>
          {{ t('settings.ai.mcp_config', 'MCP Server') }}
        </h3>

        <!-- MCP Enable Toggle -->
        <div class="flex items-center mb-4">
          <input
            id="mcpEnabled"
            v-model="mcpEnabled"
            type="checkbox"
            class="w-4 h-4 text-[#4f46e5] border-gray-300 rounded focus:ring-[#4f46e5] dark:focus:ring-[#4f46e5] dark:ring-offset-gray-800 focus:ring-2 dark:bg-gray-700 dark:border-gray-600"
          >
          <label for="mcpEnabled" class="ml-2 text-sm text-gray-700 dark:text-gray-300">
            {{ t('settings.ai.mcp_enabled', 'Enable built-in MCP Server') }}
          </label>
        </div>

        <!-- MCP Settings (shown when enabled) -->
        <div v-if="mcpEnabled" class="space-y-4 pl-6 border-l-2 border-gray-200 dark:border-gray-600">
          <!-- MCP Mode -->
          <div>
            <label class="block text-xs font-medium text-gray-600 dark:text-gray-400 mb-1">
              {{ t('settings.ai.mcp_mode', 'Transport Mode') }}
            </label>
            <select
              v-model="mcpMode"
              class="w-full px-3 py-2 bg-white dark:bg-[#323248] border border-gray-300 dark:border-gray-600 rounded-md shadow-sm text-sm focus:outline-none focus:ring-2 focus:ring-[#4f46e5]"
            >
              <option value="sse">SSE (Server-Sent Events)</option>
              <option value="stdio">STDIO</option>
            </select>
            <p class="text-xs text-gray-500 dark:text-gray-400 mt-1">
              {{ t('settings.ai.mcp_mode_hint', 'SSE for HTTP-based communication, STDIO for process-based.') }}
            </p>
          </div>

          <!-- MCP Port (only for SSE mode) -->
          <div v-if="mcpMode === 'sse'">
            <label class="block text-xs font-medium text-gray-600 dark:text-gray-400 mb-1">
              {{ t('settings.ai.mcp_port', 'Port') }}
            </label>
            <div class="flex items-center space-x-2">
              <input
                type="number"
                v-model.number="mcpPort"
                min="0"
                max="65535"
                :placeholder="t('settings.ai.mcp_port_placeholder', '0 = auto')"
                class="w-32 px-3 py-2 bg-white dark:bg-[#323248] border border-gray-300 dark:border-gray-600 rounded-md shadow-sm text-sm focus:outline-none focus:ring-2 focus:ring-[#4f46e5]"
              />
              <span class="text-xs text-gray-500 dark:text-gray-400">
                {{ mcpPort === 0 ? t('settings.ai.mcp_port_auto', '(Auto-select)') : '' }}
              </span>
            </div>
            <p class="text-xs text-gray-500 dark:text-gray-400 mt-1">
              {{ t('settings.ai.mcp_port_hint', 'Set to 0 for automatic port selection.') }}
            </p>
          </div>

          <!-- Advanced Settings Toggle -->
          <button
            type="button"
            @click="showMCPAdvanced = !showMCPAdvanced"
            class="text-xs text-indigo-600 dark:text-indigo-400 hover:text-indigo-800 dark:hover:text-indigo-300 flex items-center gap-1"
          >
            <i :class="['bx', showMCPAdvanced ? 'bx-chevron-down' : 'bx-chevron-right']"></i>
            {{ t('settings.ai.mcp_advanced', 'Advanced Tool Settings') }}
          </button>

          <!-- Advanced Settings - Tool Management -->
          <div v-if="showMCPAdvanced" class="space-y-3 pt-2">
            <!-- Quick Actions -->
            <div class="flex items-center gap-2 mb-3">
              <button
                type="button"
                @click="enableAllTools"
                class="px-2 py-1 text-xs bg-green-100 dark:bg-green-900/30 text-green-700 dark:text-green-400 rounded hover:bg-green-200 dark:hover:bg-green-900/50 transition-colors"
              >
                <i class="bx bx-check-double mr-1"></i>
                {{ t('settings.ai.mcp_enable_all', 'Enable All') }}
              </button>
              <button
                type="button"
                @click="disableAllTools"
                class="px-2 py-1 text-xs bg-red-100 dark:bg-red-900/30 text-red-700 dark:text-red-400 rounded hover:bg-red-200 dark:hover:bg-red-900/50 transition-colors"
              >
                <i class="bx bx-x mr-1"></i>
                {{ t('settings.ai.mcp_disable_all', 'Disable All') }}
              </button>
            </div>

            <!-- Tool List -->
            <div class="space-y-1">
              <!-- Read Tools -->
              <div class="mb-3">
                <div class="text-xs font-medium text-gray-500 dark:text-gray-400 mb-2 flex items-center">
                  <i class="bx bx-book-reader mr-1"></i>
                  {{ t('settings.ai.mcp_tools_read', 'Read Tools') }}
                </div>
                <div class="space-y-1">
                  <div
                    v-for="tool in availableMCPTools.filter(t => t.category === 'read')"
                    :key="tool.name"
                    class="flex items-center p-2 rounded-md bg-gray-50 dark:bg-[#323248] hover:bg-gray-100 dark:hover:bg-[#3a3a4a] transition-colors"
                  >
                    <input
                      :id="'tool-' + tool.name"
                      type="checkbox"
                      :checked="isToolEnabled(tool.name)"
                      @change="toggleTool(tool.name)"
                      class="w-4 h-4 text-[#4f46e5] border-gray-300 rounded focus:ring-[#4f46e5] dark:focus:ring-[#4f46e5] dark:ring-offset-gray-800 focus:ring-2 dark:bg-gray-700 dark:border-gray-600"
                    />
                    <label :for="'tool-' + tool.name" class="ml-2 flex-1 cursor-pointer">
                      <span class="text-sm font-mono text-gray-700 dark:text-gray-300">{{ tool.name }}</span>
                      <span class="text-xs text-gray-500 dark:text-gray-400 ml-2">{{ t(tool.descKey) }}</span>
                    </label>
                  </div>
                </div>
              </div>

              <!-- Analyze Tools -->
              <div class="mb-3">
                <div class="text-xs font-medium text-gray-500 dark:text-gray-400 mb-2 flex items-center">
                  <i class="bx bx-analyse mr-1"></i>
                  {{ t('settings.ai.mcp_tools_analyze', 'Analyze Tools') }}
                </div>
                <div class="space-y-1">
                  <div
                    v-for="tool in availableMCPTools.filter(t => t.category === 'analyze')"
                    :key="tool.name"
                    class="flex items-center p-2 rounded-md bg-gray-50 dark:bg-[#323248] hover:bg-gray-100 dark:hover:bg-[#3a3a4a] transition-colors"
                  >
                    <input
                      :id="'tool-' + tool.name"
                      type="checkbox"
                      :checked="isToolEnabled(tool.name)"
                      @change="toggleTool(tool.name)"
                      class="w-4 h-4 text-[#4f46e5] border-gray-300 rounded focus:ring-[#4f46e5] dark:focus:ring-[#4f46e5] dark:ring-offset-gray-800 focus:ring-2 dark:bg-gray-700 dark:border-gray-600"
                    />
                    <label :for="'tool-' + tool.name" class="ml-2 flex-1 cursor-pointer">
                      <span class="text-sm font-mono text-gray-700 dark:text-gray-300">{{ tool.name }}</span>
                      <span class="text-xs text-gray-500 dark:text-gray-400 ml-2">{{ t(tool.descKey) }}</span>
                    </label>
                  </div>
                </div>
              </div>

              <!-- Action Tools -->
              <div class="mb-3">
                <div class="text-xs font-medium text-gray-500 dark:text-gray-400 mb-2 flex items-center">
                  <i class="bx bx-play mr-1"></i>
                  {{ t('settings.ai.mcp_tools_action', 'Action Tools') }}
                  <span class="ml-2 px-1.5 py-0.5 text-xs bg-yellow-100 dark:bg-yellow-900/30 text-yellow-700 dark:text-yellow-400 rounded">
                    {{ t('settings.ai.mcp_tools_caution', 'Caution') }}
                  </span>
                </div>
                <div class="space-y-1">
                  <div
                    v-for="tool in availableMCPTools.filter(t => t.category === 'action')"
                    :key="tool.name"
                    class="flex items-center p-2 rounded-md bg-gray-50 dark:bg-[#323248] hover:bg-gray-100 dark:hover:bg-[#3a3a4a] transition-colors border-l-2 border-yellow-400 dark:border-yellow-600"
                  >
                    <input
                      :id="'tool-' + tool.name"
                      type="checkbox"
                      :checked="isToolEnabled(tool.name)"
                      @change="toggleTool(tool.name)"
                      class="w-4 h-4 text-[#4f46e5] border-gray-300 rounded focus:ring-[#4f46e5] dark:focus:ring-[#4f46e5] dark:ring-offset-gray-800 focus:ring-2 dark:bg-gray-700 dark:border-gray-600"
                    />
                    <label :for="'tool-' + tool.name" class="ml-2 flex-1 cursor-pointer">
                      <span class="text-sm font-mono text-gray-700 dark:text-gray-300">{{ tool.name }}</span>
                      <span class="text-xs text-gray-500 dark:text-gray-400 ml-2">{{ t(tool.descKey) }}</span>
                    </label>
                  </div>
                </div>
              </div>
            </div>

            <!-- Tool Status Summary -->
            <div class="text-xs text-gray-500 dark:text-gray-400 pt-2 border-t border-gray-200 dark:border-gray-600">
              {{ t('settings.ai.mcp_tools_summary', 'Enabled:') }}
              <span class="font-medium text-green-600 dark:text-green-400">
                {{ availableMCPTools.filter(t => isToolEnabled(t.name)).length }}
              </span>
              / {{ availableMCPTools.length }}
              {{ t('settings.ai.mcp_tools_total', 'tools') }}
            </div>
          </div>
        </div>

        <p class="text-xs text-gray-500 dark:text-gray-400 mt-3">
          {{ t('settings.ai.mcp_hint', 'MCP Server provides security tools to Claude Code for vulnerability scanning and analysis.') }}
        </p>
      </div>

      <!-- External MCP Servers -->
      <div class="bg-white dark:bg-[#282838] rounded-lg p-4 shadow-sm border border-gray-100 dark:border-gray-700">
        <div class="flex items-center justify-between mb-3">
          <h3 class="text-sm font-medium text-gray-700 dark:text-gray-300 flex items-center">
            <i class="bx bx-plug mr-2 text-gray-400"></i>
            {{ t('settings.ai.external_mcp', 'External MCP Servers') }}
          </h3>
          <button
            type="button"
            @click="showExternalMCP = !showExternalMCP"
            class="text-xs text-indigo-600 dark:text-indigo-400 hover:text-indigo-800 dark:hover:text-indigo-300 flex items-center gap-1"
          >
            <i :class="['bx', showExternalMCP ? 'bx-chevron-up' : 'bx-chevron-down']"></i>
            {{ showExternalMCP ? t('common.collapse', 'Collapse') : t('common.expand', 'Expand') }}
          </button>
        </div>

        <p class="text-xs text-gray-500 dark:text-gray-400 mb-3">
          {{ t('settings.ai.external_mcp_desc', 'Configure external MCP servers using Claude Desktop compatible JSON format.') }}
        </p>

        <div v-if="showExternalMCP" class="space-y-3">
          <!-- JSON 配置文本框 -->
          <div>
            <label class="block text-xs font-medium text-gray-600 dark:text-gray-400 mb-2">
              {{ t('settings.ai.mcp_json_config', 'MCP Servers Configuration (JSON)') }}
            </label>
            <textarea
              v-model="mcpServersJsonText"
              @blur="validateAndUpdateMcpJson"
              :placeholder='`{
  "mcpServers": {
    "context7": {
      "url": "https://mcp.context7.com/mcp",
      "headers": {
        "CONTEXT7_API_KEY": "YOUR_API_KEY"
      }
    },
    "local-server": {
      "command": "/usr/local/bin/mcp-server",
      "args": ["--port", "3000"],
      "env": {
        "DEBUG": "true"
      }
    }
  }
}`'
              rows="12"
              class="w-full px-3 py-2 bg-white dark:bg-[#323248] border rounded-md shadow-sm text-sm font-mono focus:outline-none focus:ring-2 focus:ring-[#4f46e5] resize-y"
              :class="mcpServersJsonError 
                ? 'border-red-300 dark:border-red-600' 
                : 'border-gray-300 dark:border-gray-600'"
              spellcheck="false"
            ></textarea>
            
            <!-- JSON 解析错误提示 -->
            <div v-if="mcpServersJsonError" class="mt-2 text-xs text-red-600 dark:text-red-400 flex items-center gap-1">
              <i class="bx bx-error-circle"></i>
              {{ mcpServersJsonError }}
            </div>
            
            <!-- 格式说明 -->
            <div class="mt-2 text-xs text-gray-500 dark:text-gray-400">
              <p class="mb-1">{{ t('settings.ai.mcp_json_hint', 'Compatible with Claude Desktop mcpServers format:') }}</p>
              <ul class="list-disc list-inside space-y-0.5 ml-2">
                <li><code class="bg-gray-100 dark:bg-gray-700 px-1 rounded">url</code> - SSE server URL</li>
                <li><code class="bg-gray-100 dark:bg-gray-700 px-1 rounded">headers</code> - Custom HTTP headers (for SSE)</li>
                <li><code class="bg-gray-100 dark:bg-gray-700 px-1 rounded">command</code> - Executable path (for STDIO)</li>
                <li><code class="bg-gray-100 dark:bg-gray-700 px-1 rounded">args</code> - Command arguments (for STDIO)</li>
                <li><code class="bg-gray-100 dark:bg-gray-700 px-1 rounded">env</code> - Environment variables (for STDIO)</li>
              </ul>
            </div>
          </div>

          <!-- 已解析的服务器预览 -->
          <div v-if="mcpExternalServers.length > 0 && !mcpServersJsonError" class="pt-3 border-t border-gray-200 dark:border-gray-600">
            <div class="text-xs font-medium text-gray-600 dark:text-gray-400 mb-2">
              {{ t('settings.ai.parsed_servers', 'Parsed Servers') }} ({{ mcpExternalServers.length }})
            </div>
            <div class="space-y-1">
              <div
                v-for="server in mcpExternalServers"
                :key="server.id"
                class="flex items-center gap-2 p-2 rounded bg-gray-50 dark:bg-[#323248] text-xs"
              >
                <span
                  class="px-1.5 py-0.5 rounded"
                  :class="server.type === 'sse'
                    ? 'bg-blue-100 dark:bg-blue-900/30 text-blue-700 dark:text-blue-400'
                    : 'bg-purple-100 dark:bg-purple-900/30 text-purple-700 dark:text-purple-400'"
                >
                  {{ server.type.toUpperCase() }}
                </span>
                <span class="font-medium text-gray-700 dark:text-gray-300">{{ server.name }}</span>
                <span class="text-gray-500 dark:text-gray-400 truncate flex-1">
                  {{ server.type === 'sse' ? server.url : server.command }}
                </span>
                <!-- 测试按钮 -->
                <button
                  type="button"
                  @click="testExternalMCPServer(server)"
                  :disabled="testingMCPServerId === server.id"
                  class="px-2 py-1 text-xs rounded transition-colors flex items-center gap-1"
                  :class="testingMCPServerId === server.id
                    ? 'bg-gray-200 dark:bg-gray-600 text-gray-400 dark:text-gray-500 cursor-not-allowed'
                    : 'bg-indigo-100 dark:bg-indigo-900/30 text-indigo-600 dark:text-indigo-400 hover:bg-indigo-200 dark:hover:bg-indigo-900/50'"
                >
                  <i :class="['bx', testingMCPServerId === server.id ? 'bx-loader-alt animate-spin' : 'bx-plug']"></i>
                  {{ testingMCPServerId === server.id ? t('common.testing', 'Testing...') : t('common.test', 'Test') }}
                </button>
              </div>
            </div>
          </div>
        </div>
      </div>

      </template>
      <!-- End of Claude Code CLI 配置 -->

      <!-- 保存消息 -->
      <div
        v-if="saveMessage"
        class="rounded-lg p-3 flex items-center gap-2"
        :class="saveMessageType === 'success'
          ? 'bg-green-50 dark:bg-green-900/20 text-green-700 dark:text-green-400 border border-green-200 dark:border-green-800'
          : 'bg-red-50 dark:bg-red-900/20 text-red-700 dark:text-red-400 border border-red-200 dark:border-red-800'"
      >
        <i :class="['bx', saveMessageType === 'success' ? 'bx-check-circle' : 'bx-error-circle']"></i>
        {{ saveMessage }}
      </div>

      <!-- 操作按钮 -->
      <div class="flex justify-end gap-3 pt-4 border-t border-gray-200 dark:border-gray-700">
        <button
          class="px-4 py-2 bg-gray-100 dark:bg-gray-700 hover:bg-gray-200 dark:hover:bg-gray-600 text-gray-700 dark:text-gray-300 rounded-md text-sm transition-colors flex items-center gap-2"
          @click="testConnection"
          :disabled="isSaving"
        >
          <i class="bx bx-play-circle"></i>
          {{ t('settings.ai.initialize', 'Initialize') }}
        </button>
        <button
          class="px-4 py-2 bg-emerald-500 hover:bg-emerald-600 text-white rounded-md text-sm transition-colors flex items-center gap-2 disabled:opacity-50 disabled:cursor-not-allowed"
          @click="testLLMConnection"
          :disabled="isTestingLLM || isSaving"
        >
          <i :class="['bx', isTestingLLM ? 'bx-loader-alt bx-spin' : 'bx-check-shield']"></i>
          {{ t('settings.ai.test_llm', 'Test LLM') }}
        </button>
        <button
          class="px-4 py-2 bg-[#4f46e5] hover:bg-[#4338ca] text-white rounded-md text-sm transition-colors flex items-center gap-2 disabled:opacity-50 disabled:cursor-not-allowed"
          @click="saveConfig"
          :disabled="isSaving"
        >
          <i :class="['bx', isSaving ? 'bx-loader-alt bx-spin' : 'bx-save']"></i>
          {{ t('common.save', 'Save') }}
        </button>
      </div>
    </div>
  </div>
</template>

<style scoped>
/* 禁用状态 */
button:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}
</style>
