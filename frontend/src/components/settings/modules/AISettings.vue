<script setup lang="ts">
import { ref, onMounted } from 'vue';
import { useI18n } from 'vue-i18n';
// @ts-ignore
import { ClaudeUpdateConfig, ClaudeInitialize, ClaudeGetConfig, ClaudeIsInitialized, ClaudeTestConnection, ClaudeGetCLISettings, ClaudeUpdateCLISettings, A2AUpdateConfig, A2ATestConnection } from "../../../../bindings/github.com/yhy0/ChYing/app.js";

// Props
const props = defineProps<{
  compact?: boolean;  // 紧凑模式（用于模态框）
}>();

// Emits
const emit = defineEmits<{
  (e: 'saved'): void;  // 保存成功后触发
}>();

const { t } = useI18n();

// ==================== Claude Code CLI 配置状态 ====================
// Agent 模式
const agentMode = ref<'claude-code' | 'a2a'>('claude-code');

// ChYing 配置（简化版）
const cliPath = ref('');
const model = ref('claude-sonnet-4');
const maxTurns = ref(0);
const systemPrompt = ref('');
const permissionMode = ref('default');

// Claude CLI settings.json 编辑
const cliSettingsPath = ref('');
const cliSettingsJson = ref('');
const cliSettingsJsonError = ref('');
const cliSettingsExists = ref(false);
const showCLISettings = ref(false);

// ==================== A2A 配置状态 ====================
const a2aAgentURL = ref('');
const a2aHeaders = ref<Record<string, string>>({});
const a2aTimeout = ref(300);
const a2aEnableSSE = ref(true);
const a2aHeadersJson = ref('{}');
const a2aHeadersJsonError = ref('');
const isTestingA2A = ref(false);
const a2aAgentInfo = ref<any>(null);

// ==================== 常量定义 ====================
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
const isSavingCLISettings = ref(false);
const saveMessage = ref('');
const saveMessageType = ref<'success' | 'error'>('success');
const isInitialized = ref(false);
const initError = ref('');
let messageTimer: ReturnType<typeof setTimeout> | null = null;

// ==================== 消息显示函数 ====================
const showMessage = (message: string, type: 'success' | 'error', duration = 3000) => {
  if (messageTimer) {
    clearTimeout(messageTimer);
  }

  saveMessage.value = message;
  saveMessageType.value = type;

  const timeout = type === 'error' ? Math.max(duration, 5000) : duration;
  messageTimer = setTimeout(() => {
    saveMessage.value = '';
  }, timeout);
};

// ==================== 加载配置 ====================
const loadConfig = async () => {
  isLoading.value = true;
  initError.value = '';
  try {
    // 检查初始化状态
    const initResult = await ClaudeIsInitialized();
    isInitialized.value = initResult?.data === true;

    // 获取 ChYing 配置
    const result = await ClaudeGetConfig();
    if (result?.data) {
      const config = typeof result.data === 'string' ? JSON.parse(result.data) : result.data;
      cliPath.value = config.cli_path || config.cliPath || '';
      model.value = config.model || 'claude-sonnet-4';
      maxTurns.value = config.max_turns || config.maxTurns || 0;
      systemPrompt.value = config.system_prompt || config.systemPrompt || '';
      permissionMode.value = config.permission_mode || config.permissionMode || 'default';

      // 加载 A2A 配置
      agentMode.value = config.agent_mode || 'claude-code';
      if (config.a2a) {
        a2aAgentURL.value = config.a2a.agent_url || '';
        a2aHeaders.value = config.a2a.headers || {};
        a2aTimeout.value = config.a2a.timeout || 300;
        a2aEnableSSE.value = config.a2a.enable_sse ?? true;
        a2aHeadersJson.value = JSON.stringify(a2aHeaders.value, null, 2);
      }
    }

    // 加载 Claude CLI settings.json
    await loadCLISettings();

  } catch (error) {
    console.error('加载 AI 配置失败:', error);
  } finally {
    isLoading.value = false;
  }
};

// 加载 Claude CLI settings.json
const loadCLISettings = async () => {
  try {
    const result = await ClaudeGetCLISettings();
    if (result?.data) {
      cliSettingsPath.value = result.data.path || '';
      cliSettingsExists.value = result.data.exists || false;
      if (result.data.raw) {
        cliSettingsJson.value = result.data.raw;
      } else {
        cliSettingsJson.value = JSON.stringify(result.data.settings || {}, null, 2);
      }
    }
  } catch (error) {
    console.error('加载 Claude CLI settings 失败:', error);
  }
};

// 验证 CLI settings JSON
const validateCLISettingsJson = () => {
  cliSettingsJsonError.value = '';
  if (!cliSettingsJson.value.trim()) {
    return true;
  }
  try {
    JSON.parse(cliSettingsJson.value);
    return true;
  } catch (e) {
    cliSettingsJsonError.value = e instanceof Error ? e.message : 'Invalid JSON format';
    return false;
  }
};

// 保存 Claude CLI settings.json
const saveCLISettings = async () => {
  if (!validateCLISettingsJson()) {
    showMessage('Invalid JSON format', 'error');
    return;
  }

  isSavingCLISettings.value = true;
  try {
    const result = await ClaudeUpdateCLISettings(cliSettingsJson.value);
    if (result?.error) {
      throw new Error(result.error);
    }
    showMessage('Claude CLI settings saved successfully', 'success');
    cliSettingsExists.value = true;
  } catch (error: any) {
    console.error('保存 Claude CLI settings 失败:', error);
    showMessage(error.message || 'Failed to save Claude CLI settings', 'error');
  } finally {
    isSavingCLISettings.value = false;
  }
};

// ==================== 保存 ChYing 配置 ====================
const saveConfig = async () => {
  isSaving.value = true;
  saveMessage.value = '';

  try {
    // 保存 Claude Code 配置（简化版）
    const result = await ClaudeUpdateConfig(
      cliPath.value,
      model.value,
      maxTurns.value,
      systemPrompt.value,
      permissionMode.value
    );

    if (result?.error) {
      throw new Error(result.error);
    }

    // 保存 A2A 配置
    const a2aConfig = {
      enabled: agentMode.value === 'a2a',
      agent_url: a2aAgentURL.value,
      headers: a2aHeaders.value,
      timeout: a2aTimeout.value,
      enable_sse: a2aEnableSSE.value
    };

    const a2aResult = await A2AUpdateConfig(agentMode.value, a2aConfig);
    if (a2aResult?.error) {
      throw new Error(a2aResult.error);
    }

    showMessage(t('settings.ai.save_success', 'Settings saved successfully'), 'success');
    emit('saved');
  } catch (error: any) {
    console.error('保存 AI 配置失败:', error);
    showMessage(error.message || t('settings.ai.save_error', 'Failed to save settings'), 'error');
  } finally {
    isSaving.value = false;
  }
};

// 测试连接/初始化
const testConnection = async () => {
  isSaving.value = true;
  saveMessage.value = '';

  try {
    await saveConfig();
    const result = await ClaudeInitialize();

    if (result?.error) {
      throw new Error(result.error);
    }

    isInitialized.value = true;
    showMessage(t('settings.ai.connection_success', 'Claude Code initialized successfully'), 'success');
  } catch (error: any) {
    console.error('初始化失败:', error);
    showMessage(error.message || t('settings.ai.connection_error', 'Initialization failed'), 'error');
  } finally {
    isSaving.value = false;
  }
};

// 测试 LLM 连接
const testLLMConnection = async () => {
  isTestingLLM.value = true;
  saveMessage.value = '';

  try {
    await saveConfig();
    const result = await ClaudeTestConnection();

    if (result?.error) {
      throw new Error(result.error);
    }

    showMessage(t('settings.ai.llm_test_success', 'LLM connection test successful'), 'success');
  } catch (error: any) {
    console.error('LLM 连接测试失败:', error);
    showMessage(error.message || t('settings.ai.llm_test_error', 'LLM connection test failed'), 'error');
  } finally {
    isTestingLLM.value = false;
  }
};

// ==================== A2A 相关函数 ====================
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
          : (isInitialized
            ? 'bg-green-50 dark:bg-green-900/20 border-green-200 dark:border-green-800'
            : 'bg-blue-50 dark:bg-blue-900/20 border-blue-200 dark:border-blue-800')"
      >
        <div class="text-3xl" :class="initError ? 'text-red-500' : (isInitialized ? 'text-green-500' : 'text-blue-500')">
          <i :class="['bx', initError ? 'bx-error-circle' : (isInitialized ? 'bx-check-circle' : 'bx-info-circle')]"></i>
        </div>
        <div class="flex-1">
          <div class="font-medium" :class="initError ? 'text-red-700 dark:text-red-400' : (isInitialized ? 'text-green-700 dark:text-green-400' : 'text-blue-700 dark:text-blue-400')">
            {{ initError
              ? t('settings.ai.init_failed', 'Initialization Failed')
              : (isInitialized
                ? t('settings.ai.configured', 'Claude Code Ready')
                : t('settings.ai.not_configured', 'Not Initialized'))
            }}
          </div>
          <div class="text-sm" :class="initError ? 'text-red-600 dark:text-red-500' : (isInitialized ? 'text-green-600 dark:text-green-500' : 'text-blue-600 dark:text-blue-500')">
            {{ initError
              ? initError
              : (isInitialized
                ? t('settings.ai.configured_desc', 'Claude Code CLI is ready to use')
                : t('settings.ai.not_configured_desc', 'Click "Initialize" to start using Claude Code'))
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
            <input type="radio" v-model="agentMode" value="claude-code" class="sr-only" />
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
            <input type="radio" v-model="agentMode" value="a2a" class="sr-only" />
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
                :class="a2aHeadersJsonError ? 'border-red-300 dark:border-red-600' : 'border-gray-300 dark:border-gray-600'"
                spellcheck="false"
              ></textarea>
              <div v-if="a2aHeadersJsonError" class="mt-1 text-xs text-red-600 dark:text-red-400 flex items-center gap-1">
                <i class="bx bx-error-circle"></i>
                {{ a2aHeadersJsonError }}
              </div>
            </div>

            <!-- Timeout & SSE -->
            <div class="flex items-center gap-4">
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
              <div class="flex items-center pt-5">
                <input id="a2aEnableSSE" v-model="a2aEnableSSE" type="checkbox" class="w-4 h-4 text-[#4f46e5] border-gray-300 rounded focus:ring-[#4f46e5]">
                <label for="a2aEnableSSE" class="ml-2 text-sm text-gray-700 dark:text-gray-300">
                  {{ t('settings.ai.a2a_enable_sse', 'Enable streaming (SSE)') }}
                </label>
              </div>
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

            <!-- Agent Info -->
            <div v-if="a2aAgentInfo" class="mt-3 p-3 bg-green-50 dark:bg-green-900/20 rounded-lg border border-green-200 dark:border-green-800">
              <div class="text-sm font-medium text-green-700 dark:text-green-400 mb-2">
                <i class="bx bx-check-circle mr-1"></i>
                {{ t('settings.ai.a2a_connected', 'Connected to Agent') }}
              </div>
              <div class="text-xs text-green-600 dark:text-green-500 space-y-1">
                <div><strong>Name:</strong> {{ a2aAgentInfo.name }}</div>
                <div v-if="a2aAgentInfo.description"><strong>Description:</strong> {{ a2aAgentInfo.description }}</div>
              </div>
            </div>
          </div>
        </div>
      </template>

      <!-- Claude Code CLI 配置 (仅在 Claude Code 模式下显示) -->
      <template v-if="agentMode === 'claude-code'">
        <!-- ChYing 基本配置 -->
        <div class="bg-white dark:bg-[#282838] rounded-lg p-4 shadow-sm border border-gray-100 dark:border-gray-700">
          <h3 class="text-sm font-medium mb-3 text-gray-700 dark:text-gray-300 flex items-center">
            <i class="bx bx-cog mr-2 text-gray-400"></i>
            {{ t('settings.ai.chying_config', 'ChYing Configuration') }}
          </h3>

          <div class="space-y-4">
            <!-- CLI Path -->
            <div>
              <label class="block text-xs font-medium text-gray-600 dark:text-gray-400 mb-1">
                {{ t('settings.ai.cli_path', 'Claude Code CLI Path') }}
                <span class="text-gray-400 font-normal ml-1">({{ t('common.optional', 'Optional') }})</span>
              </label>
              <input
                type="text"
                v-model="cliPath"
                :placeholder="t('settings.ai.cli_path_placeholder', 'Leave empty to use default (claude)')"
                class="w-full px-3 py-2 bg-white dark:bg-[#323248] border border-gray-300 dark:border-gray-600 rounded-md shadow-sm text-sm focus:outline-none focus:ring-2 focus:ring-[#4f46e5]"
                spellcheck="false"
              />
            </div>

            <!-- Model -->
            <div>
              <label class="block text-xs font-medium text-gray-600 dark:text-gray-400 mb-1">
                {{ t('settings.ai.model', 'Model') }}
              </label>
              <input
                type="text"
                v-model="model"
                :placeholder="t('settings.ai.model_placeholder', 'e.g., claude-sonnet-4')"
                class="w-full px-3 py-2 bg-white dark:bg-[#323248] border border-gray-300 dark:border-gray-600 rounded-md shadow-sm text-sm focus:outline-none focus:ring-2 focus:ring-[#4f46e5]"
                spellcheck="false"
              />
            </div>

            <!-- Max Turns & Permission Mode -->
            <div class="grid grid-cols-2 gap-4">
              <div>
                <label class="block text-xs font-medium text-gray-600 dark:text-gray-400 mb-1">
                  {{ t('settings.ai.max_turns', 'Max Turns') }}
                </label>
                <div class="flex items-center gap-2">
                  <input
                    type="number"
                    v-model.number="maxTurns"
                    min="0"
                    max="100"
                    class="w-24 px-3 py-2 bg-white dark:bg-[#323248] border border-gray-300 dark:border-gray-600 rounded-md shadow-sm text-sm focus:outline-none focus:ring-2 focus:ring-[#4f46e5]"
                  />
                  <span class="text-xs text-gray-500">{{ maxTurns === 0 ? '(Unlimited)' : '' }}</span>
                </div>
              </div>
              <div>
                <label class="block text-xs font-medium text-gray-600 dark:text-gray-400 mb-1">
                  {{ t('settings.ai.permission_mode', 'Permission Mode') }}
                </label>
                <select
                  v-model="permissionMode"
                  class="w-full px-3 py-2 bg-white dark:bg-[#323248] border border-gray-300 dark:border-gray-600 rounded-md shadow-sm text-sm focus:outline-none focus:ring-2 focus:ring-[#4f46e5]"
                >
                  <option v-for="mode in permissionModes" :key="mode.value" :value="mode.value">
                    {{ mode.label }}
                  </option>
                </select>
              </div>
            </div>

            <!-- System Prompt -->
            <div>
              <label class="block text-xs font-medium text-gray-600 dark:text-gray-400 mb-1">
                {{ t('settings.ai.system_prompt', 'System Prompt') }}
                <span class="text-gray-400 font-normal ml-1">({{ t('common.optional', 'Optional') }})</span>
              </label>
              <textarea
                v-model="systemPrompt"
                :placeholder="t('settings.ai.system_prompt_placeholder', 'Custom system prompt for Claude...')"
                rows="3"
                class="w-full px-3 py-2 bg-white dark:bg-[#323248] border border-gray-300 dark:border-gray-600 rounded-md shadow-sm text-sm focus:outline-none focus:ring-2 focus:ring-[#4f46e5] resize-y"
                spellcheck="false"
              ></textarea>
            </div>
          </div>
        </div>

        <!-- Claude CLI Settings.json 编辑器 -->
        <div class="bg-white dark:bg-[#282838] rounded-lg p-4 shadow-sm border border-gray-100 dark:border-gray-700">
          <div class="flex items-center justify-between mb-3">
            <h3 class="text-sm font-medium text-gray-700 dark:text-gray-300 flex items-center">
              <i class="bx bx-file mr-2 text-gray-400"></i>
              {{ t('settings.ai.cli_settings', 'Claude CLI Settings') }}
              <span class="ml-2 text-xs text-gray-400 font-normal">~/.claude/settings.json</span>
            </h3>
            <button
              type="button"
              @click="showCLISettings = !showCLISettings"
              class="text-xs text-indigo-600 dark:text-indigo-400 hover:text-indigo-800 dark:hover:text-indigo-300 flex items-center gap-1"
            >
              <i :class="['bx', showCLISettings ? 'bx-chevron-up' : 'bx-chevron-down']"></i>
              {{ showCLISettings ? t('common.collapse', 'Collapse') : t('common.expand', 'Expand') }}
            </button>
          </div>

          <p class="text-xs text-gray-500 dark:text-gray-400 mb-3">
            {{ t('settings.ai.cli_settings_desc', 'Edit Claude CLI settings directly. API keys, MCP servers, and other configurations are managed here.') }}
          </p>

          <div v-if="showCLISettings" class="space-y-3">
            <!-- 文件路径 -->
            <div class="flex items-center gap-2 text-xs">
              <span class="text-gray-500">{{ t('settings.ai.file_path', 'File:') }}</span>
              <code class="bg-gray-100 dark:bg-gray-700 px-2 py-1 rounded">{{ cliSettingsPath }}</code>
              <span v-if="!cliSettingsExists" class="text-yellow-600 dark:text-yellow-400">
                ({{ t('settings.ai.file_not_exists', 'Will be created') }})
              </span>
            </div>

            <!-- JSON 编辑器 -->
            <div>
              <textarea
                v-model="cliSettingsJson"
                @blur="validateCLISettingsJson"
                :placeholder='`{
  "apiKey": "sk-ant-...",
  "mcpServers": {
    "context7": {
      "url": "https://mcp.context7.com/mcp"
    }
  }
}`'
                rows="15"
                class="w-full px-3 py-2 bg-white dark:bg-[#323248] border rounded-md shadow-sm text-sm font-mono focus:outline-none focus:ring-2 focus:ring-[#4f46e5] resize-y"
                :class="cliSettingsJsonError ? 'border-red-300 dark:border-red-600' : 'border-gray-300 dark:border-gray-600'"
                spellcheck="false"
              ></textarea>

              <div v-if="cliSettingsJsonError" class="mt-2 text-xs text-red-600 dark:text-red-400 flex items-center gap-1">
                <i class="bx bx-error-circle"></i>
                {{ cliSettingsJsonError }}
              </div>
            </div>

            <!-- 保存按钮 -->
            <div class="flex justify-end">
              <button
                type="button"
                @click="saveCLISettings"
                :disabled="isSavingCLISettings || !!cliSettingsJsonError"
                class="px-4 py-2 bg-indigo-500 hover:bg-indigo-600 text-white rounded-md text-sm transition-colors flex items-center gap-2 disabled:opacity-50 disabled:cursor-not-allowed"
              >
                <i :class="['bx', isSavingCLISettings ? 'bx-loader-alt bx-spin' : 'bx-save']"></i>
                {{ t('settings.ai.save_cli_settings', 'Save CLI Settings') }}
              </button>
            </div>

            <!-- 帮助信息 -->
            <div class="mt-3 p-3 bg-blue-50 dark:bg-blue-900/20 rounded-lg border border-blue-200 dark:border-blue-800">
              <div class="text-xs text-blue-700 dark:text-blue-400">
                <p class="font-medium mb-1">{{ t('settings.ai.cli_settings_help_title', 'Common settings:') }}</p>
                <ul class="list-disc list-inside space-y-0.5 ml-2">
                  <li><code class="bg-blue-100 dark:bg-blue-800 px-1 rounded">apiKey</code> - Anthropic API key</li>
                  <li><code class="bg-blue-100 dark:bg-blue-800 px-1 rounded">mcpServers</code> - MCP server configurations</li>
                  <li><code class="bg-blue-100 dark:bg-blue-800 px-1 rounded">allowedTools</code> - Allowed tool patterns</li>
                  <li><code class="bg-blue-100 dark:bg-blue-800 px-1 rounded">env</code> - Environment variables</li>
                </ul>
              </div>
            </div>
          </div>
        </div>
      </template>

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
button:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}
</style>
