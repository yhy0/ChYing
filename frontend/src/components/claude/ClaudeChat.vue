<script setup lang="ts">
import { ref, computed, watch, nextTick, onMounted } from 'vue';
import { useI18n } from 'vue-i18n';
import { useClaudeStore } from '../../store/claude';
import ClaudeMessage from './ClaudeMessage.vue';
import { message } from '../../utils/message';

const props = defineProps<{
  initialMessage?: string;
  selectedTrafficIds?: number[];
}>();

const emit = defineEmits<{
  (e: 'clearTrafficIds'): void;
}>();

const { t } = useI18n();
const claudeStore = useClaudeStore();

// 统计信息
const costInfo = computed(() => claudeStore.costInfo);
const currentModel = computed(() => claudeStore.config?.model || 'claude');

// 输入状态
const inputMessage = ref('');
const inputRef = ref<HTMLTextAreaElement | null>(null);
const messagesContainerRef = ref<HTMLDivElement | null>(null);

// 计算属性
const messages = computed(() => claudeStore.messages);
const isStreaming = computed(() => claudeStore.streaming);
const isLoading = computed(() => claudeStore.loading);
const isInitialized = computed(() => claudeStore.initialized);
const streamingContent = computed(() => claudeStore.streamingContent);
const errorMessage = computed(() => claudeStore.error);

// 发送消息
const sendMessage = async () => {
  const message = inputMessage.value.trim();
  if (!message || isStreaming.value || isLoading.value) return;

  inputMessage.value = '';
  await claudeStore.sendMessage(message);
};

// 处理键盘事件
const handleKeydown = (e: KeyboardEvent) => {
  if (e.key === 'Enter' && !e.shiftKey) {
    e.preventDefault();
    sendMessage();
  }
};

// 自动调整输入框高度
const adjustTextareaHeight = () => {
  if (inputRef.value) {
    inputRef.value.style.height = 'auto';
    inputRef.value.style.height = Math.min(inputRef.value.scrollHeight, 200) + 'px';
  }
};

// 滚动到底部
const scrollToBottom = () => {
  nextTick(() => {
    if (messagesContainerRef.value) {
      messagesContainerRef.value.scrollTop = messagesContainerRef.value.scrollHeight;
    }
  });
};

// 监听消息变化，自动滚动
watch(messages, scrollToBottom, { deep: true });
watch(streamingContent, scrollToBottom);

// 监听错误消息，显示通知
watch(errorMessage, (newError) => {
  if (newError) {
    message.error(newError, 5);
    // 清除错误状态，避免重复显示
    claudeStore.clearError();
  }
});

// 组件挂载时聚焦输入框
onMounted(() => {
  // 如果有初始消息，设置到输入框
  if (props.initialMessage) {
    inputMessage.value = props.initialMessage;
    nextTick(() => {
      adjustTextareaHeight();
    });
  }
  if (inputRef.value) {
    inputRef.value.focus();
  }
});

// 监听 initialMessage 变化
watch(() => props.initialMessage, (newMessage) => {
  if (newMessage) {
    inputMessage.value = newMessage;
    nextTick(() => {
      adjustTextareaHeight();
      if (inputRef.value) {
        inputRef.value.focus();
        // 将光标移到末尾
        inputRef.value.setSelectionRange(newMessage.length, newMessage.length);
      }
    });
  }
});

// 暴露方法供父组件调用
defineExpose({
  setMessage: (msg: string) => {
    inputMessage.value = msg;
    nextTick(() => {
      adjustTextareaHeight();
      if (inputRef.value) {
        inputRef.value.focus();
        inputRef.value.setSelectionRange(msg.length, msg.length);
      }
    });
  },
  focusInput: () => {
    if (inputRef.value) {
      inputRef.value.focus();
    }
  }
});
</script>

<template>
  <div class="claude-chat">
    <!-- 消息列表 -->
    <div class="messages-container scrollbar-thin" ref="messagesContainerRef">
      <!-- 欢迎消息 -->
      <div v-if="messages.length === 0" class="welcome-message">
        <div class="welcome-icon">
          <i class="bx bx-bot"></i>
        </div>
        <h3>{{ t('claude.welcome.title', 'AI Security Assistant') }}</h3>
        <p>{{ t('claude.welcome.description', 'I can help you analyze HTTP traffic, identify vulnerabilities, and provide security testing recommendations.') }}</p>
        <div class="welcome-suggestions">
          <button class="suggestion-btn" @click="inputMessage = t('claude.suggestions.analyzeTraffic', 'Analyze recent HTTP traffic')">
            <i class="bx bx-history"></i>
            {{ t('claude.suggestions.analyzeTraffic', 'Analyze recent HTTP traffic') }}
          </button>
          <button class="suggestion-btn" @click="inputMessage = t('claude.suggestions.listVulns', 'Show discovered vulnerabilities')">
            <i class="bx bx-bug"></i>
            {{ t('claude.suggestions.listVulns', 'Show discovered vulnerabilities') }}
          </button>
          <button class="suggestion-btn" @click="inputMessage = t('claude.suggestions.fingerprints', 'What technologies are detected?')">
            <i class="bx bx-fingerprint"></i>
            {{ t('claude.suggestions.fingerprints', 'What technologies are detected?') }}
          </button>
        </div>
      </div>

      <!-- 消息列表 -->
      <template v-for="message in messages" :key="message.id">
        <ClaudeMessage :message="message" />
      </template>

      <!-- 流式加载指示器 -->
      <div v-if="isStreaming" class="streaming-indicator">
        <div class="typing-dots">
          <span></span>
          <span></span>
          <span></span>
        </div>
      </div>
    </div>

    <!-- 输入区域 -->
    <div class="input-container">
      <!-- 选中流量提示 -->
      <div
        v-if="selectedTrafficIds && selectedTrafficIds.length > 0"
        class="traffic-banner"
      >
        <i class="bx bx-transfer-alt"></i>
        <span>{{ t('claude.prefill.selectedTraffic', { count: selectedTrafficIds.length }) }}</span>
        <span class="traffic-ids">ID: {{ selectedTrafficIds.join(', ') }}</span>
        <button @click="emit('clearTrafficIds')" class="btn btn-sm btn-ghost">
          <i class="bx bx-x"></i>
        </button>
      </div>
      <!-- 统计信息面板 -->
      <div v-if="costInfo || messages.length > 0" class="stats-panel">
        <div class="stats-item" :title="t('claude.stats.model', 'Model')">
          <i class="bx bx-chip"></i>
          <span>{{ currentModel }}</span>
        </div>
        <div class="stats-item" v-if="costInfo" :title="t('claude.stats.inputTokens', 'Input Tokens')">
          <i class="bx bx-log-in"></i>
          <span>{{ costInfo.inputTokens.toLocaleString() }}</span>
        </div>
        <div class="stats-item" v-if="costInfo" :title="t('claude.stats.outputTokens', 'Output Tokens')">
          <i class="bx bx-log-out"></i>
          <span>{{ costInfo.outputTokens.toLocaleString() }}</span>
        </div>
        <div class="stats-item" v-if="costInfo" :title="t('claude.stats.totalTokens', 'Total Tokens')">
          <i class="bx bx-transfer"></i>
          <span>{{ (costInfo.inputTokens + costInfo.outputTokens).toLocaleString() }}</span>
        </div>
        <div class="stats-item cost" v-if="costInfo" :title="t('claude.stats.cost', 'Cost')">
          <i class="bx bx-dollar"></i>
          <span>${{ costInfo.costUSD.toFixed(4) }}</span>
        </div>
        <div class="stats-item" :title="t('claude.stats.messages', 'Messages')">
          <i class="bx bx-message-square-dots"></i>
          <span>{{ messages.length }}</span>
        </div>
      </div>
      <div class="input-wrapper">
        <textarea
          spellcheck="false"
          ref="inputRef"
          v-model="inputMessage"
          @keydown="handleKeydown"
          @input="adjustTextareaHeight"
          :placeholder="t('claude.input.placeholder', 'Ask me about security testing...')"
          :disabled="!isInitialized || isStreaming"
          rows="1"
          class="message-input scrollbar-thin"
        ></textarea>
        <button
          @click="sendMessage"
          :disabled="!inputMessage.trim() || !isInitialized || isStreaming"
          class="send-btn"
          :class="{ 'sending': isStreaming }"
        >
          <i v-if="isStreaming" class="bx bx-loader-alt bx-spin"></i>
          <i v-else class="bx bx-send"></i>
        </button>
      </div>
      <div class="input-hint">
        <span>{{ t('claude.input.hint', 'Press Enter to send, Shift+Enter for new line') }}</span>
      </div>
    </div>
  </div>
</template>

<style scoped>
.claude-chat {
  display: flex;
  flex-direction: column;
  height: 100%;
  background: var(--color-bg-primary);
}

.messages-container {
  flex: 1;
  overflow-y: auto;
  padding: 16px;
  display: flex;
  flex-direction: column;
  gap: 16px;
}

/* 欢迎消息 */
.welcome-message {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  text-align: center;
  padding: 48px 24px;
  margin: auto;
  max-width: 600px;
}

.welcome-icon {
  width: 80px;
  height: 80px;
  border-radius: 50%;
  background: var(--color-primary-bg, rgba(59, 130, 246, 0.1));
  display: flex;
  align-items: center;
  justify-content: center;
  margin-bottom: 24px;
}

.welcome-icon i {
  font-size: 40px;
  color: var(--color-primary);
}

.welcome-message h3 {
  margin: 0 0 12px 0;
  font-size: 1.5rem;
  color: var(--color-text-primary);
}

.welcome-message p {
  margin: 0 0 24px 0;
  color: var(--color-text-secondary);
  line-height: 1.6;
}

.welcome-suggestions {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
  justify-content: center;
}

.suggestion-btn {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 10px 16px;
  background: var(--color-bg-secondary);
  border: 1px solid var(--color-border);
  border-radius: 8px;
  color: var(--color-text-primary);
  font-size: 0.875rem;
  cursor: pointer;
  transition: all 0.2s ease;
}

.suggestion-btn:hover {
  background: var(--color-bg-tertiary);
  border-color: var(--color-primary);
}

.suggestion-btn i {
  font-size: 1rem;
  color: var(--color-primary);
}

/* 流式加载指示器 */
.streaming-indicator {
  display: flex;
  align-items: center;
  padding: 12px 16px;
}

.typing-dots {
  display: flex;
  gap: 4px;
}

.typing-dots span {
  width: 8px;
  height: 8px;
  border-radius: 50%;
  background: var(--color-primary);
  animation: typing 1.4s infinite ease-in-out both;
}

.typing-dots span:nth-child(1) {
  animation-delay: -0.32s;
}

.typing-dots span:nth-child(2) {
  animation-delay: -0.16s;
}

@keyframes typing {
  0%, 80%, 100% {
    transform: scale(0.6);
    opacity: 0.5;
  }
  40% {
    transform: scale(1);
    opacity: 1;
  }
}

/* 输入区域 */
.input-container {
  padding: 16px;
  border-top: 1px solid var(--color-border);
  background: var(--color-bg-secondary);
}

/* 选中流量提示 */
.traffic-banner {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 10px 12px;
  margin-bottom: 12px;
  background: var(--color-info-bg, rgba(59, 130, 246, 0.1));
  border: 1px solid var(--color-info, #3b82f6);
  border-radius: 8px;
  color: var(--color-info, #3b82f6);
  font-size: 0.875rem;
}

.traffic-banner i {
  font-size: 1.25rem;
}

.traffic-banner .traffic-ids {
  font-family: var(--font-family-mono);
  font-size: 0.75rem;
  background: rgba(59, 130, 246, 0.15);
  padding: 2px 8px;
  border-radius: 4px;
}

.traffic-banner .btn-ghost {
  margin-left: auto;
  background: transparent;
  border: none;
  color: inherit;
  padding: 4px;
  cursor: pointer;
  opacity: 0.7;
  transition: opacity 0.2s;
}

.traffic-banner .btn-ghost:hover {
  opacity: 1;
}

.input-wrapper {
  display: flex;
  align-items: flex-end;
  gap: 12px;
  background: var(--color-bg-primary);
  border: 1px solid var(--color-border);
  border-radius: 12px;
  padding: 8px 12px;
  transition: border-color 0.2s ease;
}

.input-wrapper:focus-within {
  border-color: var(--color-primary);
}

.message-input {
  flex: 1;
  border: none;
  background: transparent;
  color: var(--color-text-primary);
  font-size: 0.9375rem;
  line-height: 1.5;
  resize: none;
  outline: none;
  min-height: 24px;
  max-height: 200px;
}

.message-input::placeholder {
  color: var(--color-text-tertiary);
}

.message-input:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.send-btn {
  width: 40px;
  height: 40px;
  border-radius: 8px;
  border: none;
  background: var(--color-primary);
  color: white;
  cursor: pointer;
  display: flex;
  align-items: center;
  justify-content: center;
  transition: all 0.2s ease;
  flex-shrink: 0;
}

.send-btn:hover:not(:disabled) {
  background: var(--color-primary-hover);
  transform: scale(1.05);
}

.send-btn:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.send-btn i {
  font-size: 1.25rem;
}

.input-hint {
  margin-top: 8px;
  text-align: center;
  font-size: 0.75rem;
  color: var(--color-text-tertiary);
}

/* 统计信息面板 */
.stats-panel {
  display: flex;
  flex-wrap: wrap;
  gap: 12px;
  padding: 8px 12px;
  margin-bottom: 12px;
  background: var(--color-bg-primary);
  border: 1px solid var(--color-border);
  border-radius: 8px;
  font-size: 0.75rem;
}

.stats-item {
  display: flex;
  align-items: center;
  gap: 4px;
  color: var(--color-text-secondary);
  cursor: default;
  transition: color 0.2s ease;
}

.stats-item:hover {
  color: var(--color-text-primary);
}

.stats-item i {
  font-size: 0.875rem;
  color: var(--color-primary);
}

.stats-item.cost {
  color: var(--color-success);
  font-weight: 500;
}

.stats-item.cost i {
  color: var(--color-success);
}
</style>
