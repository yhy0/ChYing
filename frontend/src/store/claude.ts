import { defineStore } from 'pinia';
import { ref, computed } from 'vue';
import type {
  ChatMessage,
  ToolUse,
  Session,
  AgentContext,
  StreamEvent,
  ClaudeConfig,
  ClaudeState
} from '../types/claude';
import { getToolInfo } from '../types/claude';

// Wails bindings - 使用 App 的方法
// @ts-ignore
import {
  ClaudeInitialize,
  ClaudeIsInitialized,
  ClaudeGetConfig,
  ClaudeUpdateConfig,
  ClaudeCreateSession,
  ClaudeCreateSessionWithContext,
  ClaudeGetSession,
  ClaudeGetSessionHistory,
  ClaudeDeleteSession,
  ClaudeClearSession,
  ClaudeListSessions,
  ClaudeListSessionsByProject,
  ClaudeListProjects,
  ClaudeSendMessage,
  ClaudeConfirmToolExecution,
  ClaudeUpdateContext,
  ClaudeStopSession
} from "../../bindings/github.com/yhy0/ChYing/app.js";
import { Events } from "@wailsio/runtime";

/**
 * Claude AI Agent Store
 * 管理 Claude AI 对话状态和交互
 * 基于 Claude Code CLI 实现
 */
export const useClaudeStore = defineStore('claude', () => {
  // ==================== 状态定义 ====================

  // 初始化状态
  const initialized = ref(false);
  const loading = ref(false);
  const streaming = ref(false);

  // 当前会话
  const currentSessionId = ref<string | null>(null);
  const sessions = ref<Map<string, Session>>(new Map());

  // 消息列表
  const messages = ref<ChatMessage[]>([]);

  // 待确认的工具调用
  const pendingToolUses = ref<ToolUse[]>([]);

  // 配置
  const config = ref<ClaudeConfig | null>(null);

  // 错误信息
  const error = ref<string | null>(null);

  // 流式内容缓冲
  const streamingContent = ref('');

  // 费用统计
  const costInfo = ref<{
    costUSD: number;
    inputTokens: number;
    outputTokens: number;
  } | null>(null);

  // ==================== 计算属性 ====================

  // 是否有待确认的工具
  const hasPendingTools = computed(() => pendingToolUses.value.length > 0);

  // 危险工具列表
  const dangerousTools = computed(() =>
    pendingToolUses.value.filter(tu => {
      const info = getToolInfo(tu.name);
      return info.dangerous;
    })
  );

  // 当前会话
  const currentSession = computed(() => {
    if (!currentSessionId.value) return null;
    return sessions.value.get(currentSessionId.value) || null;
  });

  // ==================== 初始化方法 ====================

  /**
   * 初始化 Claude Code 服务
   */
  const initializeService = async (): Promise<boolean> => {
    try {
      loading.value = true;
      error.value = null;

      const result = await ClaudeInitialize();

      // 检查错误
      if (result?.error) {
        error.value = result.error;
        console.error('Claude Code 初始化失败:', result.error);
        return false;
      }

      initialized.value = true;

      // 获取配置
      await fetchConfig();

      // 设置事件监听
      setupEventListeners();

      return true;
    } catch (e) {
      error.value = e instanceof Error ? e.message : String(e);
      console.error('Claude Code 初始化异常:', e);
      return false;
    } finally {
      loading.value = false;
    }
  };

  /**
   * 检查是否已初始化
   */
  const checkInitialized = async (): Promise<boolean> => {
    try {
      const result = await ClaudeIsInitialized();
      initialized.value = !!result?.data;
      return initialized.value;
    } catch (e) {
      console.error('检查初始化状态失败:', e);
      return false;
    }
  };

  /**
   * 获取配置
   */
  const fetchConfig = async () => {
    try {
      const result = await ClaudeGetConfig();
      if (result?.data) {
        const data = result.data;
        config.value = {
          cliPath: data.cli_path,
          model: data.model,
          maxTurns: data.max_turns,
          systemPrompt: data.system_prompt,
          permissionMode: data.permission_mode
        } as ClaudeConfig;
      }
    } catch (e) {
      console.error('获取 Claude 配置失败:', e);
    }
  };

  /**
   * 更新配置
   * 简化版：只更新 ChYing 管理的配置项
   * API Key、代理、MCP 服务器等配置请在 ~/.claude/settings.json 中设置
   */
  const updateConfig = async (newConfig: Partial<ClaudeConfig>): Promise<boolean> => {
    try {
      const currentConfig = config.value || {};
      const mergedConfig = { ...currentConfig, ...newConfig };

      const result = await ClaudeUpdateConfig(
        mergedConfig.cliPath || '',
        mergedConfig.model || '',
        mergedConfig.maxTurns || 50,
        mergedConfig.systemPrompt || '',
        mergedConfig.permissionMode || 'default'
      );

      if (result?.error) {
        error.value = result.error;
        return false;
      }

      await fetchConfig();
      return true;
    } catch (e) {
      error.value = e instanceof Error ? e.message : String(e);
      return false;
    }
  };

  // ==================== 会话管理 ====================

  /**
   * 创建新会话
   */
  const createSession = async (projectId: string, context?: AgentContext): Promise<string | null> => {
    try {
      loading.value = true;
      error.value = null;

      let result;
      if (context) {
        result = await ClaudeCreateSessionWithContext(projectId, context as any);
      } else {
        result = await ClaudeCreateSession(projectId);
      }

      // 检查错误
      if (result?.error) {
        error.value = result.error;
        console.error('创建会话失败:', result.error);
        return null;
      }

      // 检查返回数据是否有效
      if (!result?.data) {
        error.value = '创建会话失败：返回数据为空';
        console.error('Result:', result);
        return null;
      }

      const sessionData = result.data as { session_id: string; project_id: string; created_at: string };

      // 检查 session_id 是否存在
      if (!sessionData.session_id) {
        error.value = '创建会话失败：session_id 为空';
        console.error('Session data:', sessionData);
        return null;
      }

      const sessionId = sessionData.session_id;

      // 创建本地会话对象
      const session: Session = {
        id: sessionId,
        projectId: projectId,
        context: context,
        createdAt: sessionData.created_at,
        updatedAt: sessionData.created_at,
        history: []
      };

      sessions.value.set(sessionId, session);
      currentSessionId.value = sessionId;
      messages.value = [];
      pendingToolUses.value = [];
      costInfo.value = null;

      return sessionId;
    } catch (e) {
      error.value = e instanceof Error ? e.message : String(e);
      return null;
    } finally {
      loading.value = false;
    }
  };

  /**
   * 切换会话
   */
  const switchSession = async (sessionId: string): Promise<boolean> => {
    try {
      loading.value = true;

      const result = await ClaudeGetSession(sessionId);
      if (result?.error) {
        error.value = result.error;
        return false;
      }

      const sessionData = result?.data as any;
      currentSessionId.value = sessionId;
      messages.value = sessionData?.history || [];

      return true;
    } catch (e) {
      error.value = e instanceof Error ? e.message : String(e);
      return false;
    } finally {
      loading.value = false;
    }
  };

  /**
   * 获取会话历史
   */
  const getSessionHistory = async (sessionId: string): Promise<ChatMessage[]> => {
    try {
      const result = await ClaudeGetSessionHistory(sessionId);
      return result?.data || [];
    } catch (e) {
      console.error('获取会话历史失败:', e);
      return [];
    }
  };

  /**
   * 列出所有会话
   */
  const listSessions = async (): Promise<Session[]> => {
    try {
      const result = await ClaudeListSessions();
      return result?.data || [];
    } catch (e) {
      console.error('列出会话失败:', e);
      return [];
    }
  };

  /**
   * 列出项目的会话
   */
  const listSessionsByProject = async (projectId: string): Promise<Session[]> => {
    try {
      const result = await ClaudeListSessionsByProject(projectId);
      return result?.data || [];
    } catch (e) {
      console.error('列出项目会话失败:', e);
      return [];
    }
  };

  /**
   * 列出所有有会话的项目
   */
  const listProjects = async (): Promise<string[]> => {
    try {
      const result = await ClaudeListProjects();
      return result?.data || [];
    } catch (e) {
      console.error('列出项目失败:', e);
      return [];
    }
  };

  /**
   * 删除会话
   */
  const deleteSession = async (sessionId: string): Promise<boolean> => {
    try {
      await ClaudeDeleteSession(sessionId);
      sessions.value.delete(sessionId);

      if (currentSessionId.value === sessionId) {
        currentSessionId.value = null;
        messages.value = [];
      }

      return true;
    } catch (e) {
      error.value = e instanceof Error ? e.message : String(e);
      return false;
    }
  };

  /**
   * 清除当前会话
   */
  const clearCurrentSession = async (): Promise<boolean> => {
    if (!currentSessionId.value) return false;

    try {
      await ClaudeClearSession(currentSessionId.value);
      messages.value = [];
      pendingToolUses.value = [];
      costInfo.value = null;
      return true;
    } catch (e) {
      error.value = e instanceof Error ? e.message : String(e);
      return false;
    }
  };

  /**
   * 停止当前会话
   */
  const stopCurrentSession = async (): Promise<boolean> => {
    if (!currentSessionId.value) return false;

    try {
      await ClaudeStopSession(currentSessionId.value);
      streaming.value = false;
      return true;
    } catch (e) {
      error.value = e instanceof Error ? e.message : String(e);
      return false;
    }
  };

  // ==================== 消息发送 ====================

  /**
   * 发送消息（流式）
   */
  const sendMessage = async (message: string): Promise<boolean> => {
    if (!currentSessionId.value) {
      error.value = '没有活动会话';
      return false;
    }

    try {
      streaming.value = true;
      error.value = null;
      streamingContent.value = '';

      // 添加用户消息到本地
      const userMessage: ChatMessage = {
        id: `user-${Date.now()}`,
        role: 'user',
        content: message,
        timestamp: new Date().toISOString()
      };
      messages.value.push(userMessage);

      // 添加占位的助手消息
      const assistantMessage: ChatMessage = {
        id: `assistant-${Date.now()}`,
        role: 'assistant',
        content: '',
        timestamp: new Date().toISOString(),
        toolUses: []
      };
      messages.value.push(assistantMessage);

      // 发送流式请求
      const result = await ClaudeSendMessage(currentSessionId.value, message);
      if (result?.error) {
        error.value = result.error;
        streaming.value = false;
        return false;
      }

      return true;
    } catch (e) {
      error.value = e instanceof Error ? e.message : String(e);
      streaming.value = false;
      return false;
    }
  };

  /**
   * 处理流式事件
   */
  const handleStreamEvent = (event: StreamEvent) => {
    // 检查是否是当前会话的事件
    if (event.sessionId && event.sessionId !== currentSessionId.value) {
      return;
    }

    const lastMessage = messages.value[messages.value.length - 1];

    switch (event.type) {
      case 'text':
        if (lastMessage && lastMessage.role === 'assistant') {
          lastMessage.content += event.content || '';
          streamingContent.value = lastMessage.content;
        }
        break;

      case 'tool_use':
        if (event.toolUse) {
          const toolUse: ToolUse = {
            ...event.toolUse,
            status: event.toolUse.status || 'pending'
          };

          if (lastMessage && lastMessage.role === 'assistant') {
            if (!lastMessage.toolUses) {
              lastMessage.toolUses = [];
            }
            // 检查是否已存在
            const existingIndex = lastMessage.toolUses.findIndex(tu => tu.id === toolUse.id);
            if (existingIndex >= 0) {
              lastMessage.toolUses[existingIndex] = toolUse;
            } else {
              lastMessage.toolUses.push(toolUse);
            }
          }

          // Claude Code CLI 自动处理工具确认，这里只记录危险工具用于 UI 显示
          const toolInfo = getToolInfo(toolUse.name);
          if (toolInfo.dangerous) {
            if (!pendingToolUses.value.find(tu => tu.id === toolUse.id)) {
              pendingToolUses.value.push(toolUse);
            }
          }
        }
        break;

      case 'tool_result':
        // 工具执行结果
        if (event.toolUse) {
          updateToolUseStatus(event.toolUse.id, event.toolUse.status || 'completed', event.toolUse.result);
        }
        break;

      case 'cost':
        // 费用统计 - 累加而非覆盖
        if (event.costUSD !== undefined) {
          if (costInfo.value) {
            costInfo.value = {
              costUSD: costInfo.value.costUSD + event.costUSD,
              inputTokens: costInfo.value.inputTokens + (event.inputTokens || 0),
              outputTokens: costInfo.value.outputTokens + (event.outputTokens || 0)
            };
          } else {
            costInfo.value = {
              costUSD: event.costUSD,
              inputTokens: event.inputTokens || 0,
              outputTokens: event.outputTokens || 0
            };
          }
        }
        break;

      case 'error':
        error.value = event.error || '未知错误';
        streaming.value = false;
        break;

      case 'done':
        streaming.value = false;
        streamingContent.value = '';
        break;
    }
  };

  /**
   * 更新工具调用状态
   */
  const updateToolUseStatus = (toolUseId: string, status: ToolUse['status'], result?: string) => {
    // 更新消息中的工具状态
    for (const msg of messages.value) {
      if (msg.toolUses) {
        const toolUse = msg.toolUses.find(tu => tu.id === toolUseId);
        if (toolUse) {
          toolUse.status = status;
          if (result) {
            toolUse.result = result;
          }
          break;
        }
      }
    }

    // 从待确认列表中移除
    pendingToolUses.value = pendingToolUses.value.filter(tu => tu.id !== toolUseId);
  };

  // ==================== 工具执行 ====================

  /**
   * 确认工具执行
   * 注意：Claude Code CLI 模式下，工具执行由 CLI 自动处理
   * 此方法保留用于前端兼容性
   */
  const confirmToolExecution = async (toolUseId: string, confirmed: boolean): Promise<boolean> => {
    if (!currentSessionId.value) return false;

    try {
      const result = await ClaudeConfirmToolExecution(currentSessionId.value, toolUseId, confirmed);
      if (result?.error) {
        error.value = result.error;
        return false;
      }

      if (confirmed) {
        updateToolUseStatus(toolUseId, 'confirmed');
      } else {
        updateToolUseStatus(toolUseId, 'rejected');
      }

      return true;
    } catch (e) {
      error.value = e instanceof Error ? e.message : String(e);
      return false;
    }
  };

  // ==================== 上下文管理 ====================

  /**
   * 更新会话上下文
   */
  const updateContext = async (context: AgentContext): Promise<boolean> => {
    if (!currentSessionId.value) return false;

    try {
      const result = await ClaudeUpdateContext(currentSessionId.value, context as any);
      if (result?.error) {
        error.value = result.error;
        return false;
      }

      // 更新本地会话
      const session = sessions.value.get(currentSessionId.value);
      if (session) {
        session.context = context;
      }

      return true;
    } catch (e) {
      error.value = e instanceof Error ? e.message : String(e);
      return false;
    }
  };

  // ==================== 事件监听 ====================

  /**
   * 设置事件监听
   */
  const setupEventListeners = () => {
    // 监听文本事件
    Events.On('claude:text', (event: any) => {
      const data = event?.data || event;
      handleStreamEvent({ ...data, type: 'text' } as StreamEvent);
    });

    // 监听工具调用事件
    Events.On('claude:tool_use', (event: any) => {
      const data = event?.data || event;
      handleStreamEvent({ ...data, type: 'tool_use' } as StreamEvent);
    });

    // 监听工具结果事件
    Events.On('claude:tool_result', (event: any) => {
      const data = event?.data || event;
      handleStreamEvent({ ...data, type: 'tool_result' } as StreamEvent);
    });

    // 监听错误事件
    Events.On('claude:error', (event: any) => {
      const data = event?.data || event;
      handleStreamEvent({ ...data, type: 'error' } as StreamEvent);
    });

    // 监听完成事件
    Events.On('claude:done', (event: any) => {
      const data = event?.data || event;
      handleStreamEvent({ ...data, type: 'done' } as StreamEvent);
    });

    // 监听费用事件
    Events.On('claude:cost', (event: any) => {
      const data = event?.data || event;
      handleStreamEvent({ ...data, type: 'cost' } as StreamEvent);
    });

    // 监听通用流式事件（兼容）
    Events.On('claude:stream', (event: any) => {
      const data = event?.data || event;
      handleStreamEvent(data as StreamEvent);
    });
  };

  /**
   * 清理事件监听
   */
  const cleanupEventListeners = () => {
    Events.Off('claude:text');
    Events.Off('claude:tool_use');
    Events.Off('claude:tool_result');
    Events.Off('claude:error');
    Events.Off('claude:done');
    Events.Off('claude:cost');
    Events.Off('claude:stream');
  };

  // ==================== 清理方法 ====================

  /**
   * 清除错误状态
   */
  const clearError = () => {
    error.value = null;
  };

  /**
   * 重置状态
   */
  const reset = () => {
    cleanupEventListeners();
    initialized.value = false;
    loading.value = false;
    streaming.value = false;
    currentSessionId.value = null;
    sessions.value.clear();
    messages.value = [];
    pendingToolUses.value = [];
    config.value = null;
    error.value = null;
    streamingContent.value = '';
    costInfo.value = null;
  };

  // 返回所有公开的状态、计算属性和方法
  return {
    // 状态
    initialized,
    loading,
    streaming,
    currentSessionId,
    sessions,
    messages,
    pendingToolUses,
    config,
    error,
    streamingContent,
    costInfo,

    // 计算属性
    hasPendingTools,
    dangerousTools,
    currentSession,

    // 初始化方法
    initializeService,
    checkInitialized,
    fetchConfig,
    updateConfig,

    // 会话管理
    createSession,
    switchSession,
    getSessionHistory,
    listSessions,
    listSessionsByProject,
    listProjects,
    deleteSession,
    clearCurrentSession,
    stopCurrentSession,

    // 消息发送
    sendMessage,
    handleStreamEvent,

    // 工具执行
    confirmToolExecution,

    // 上下文管理
    updateContext,

    // 事件监听
    setupEventListeners,
    cleanupEventListeners,

    // 清理
    reset,
    clearError
  };
});
