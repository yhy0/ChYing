<script setup lang="ts">
import { computed } from 'vue';
import { useI18n } from 'vue-i18n';
import type { ChatMessage } from '../../types/claude';
import ClaudeToolUse from './ClaudeToolUse.vue';

const props = defineProps<{
  message: ChatMessage;
}>();

const { t } = useI18n();

const isUser = computed(() => props.message.role === 'user');
const isAssistant = computed(() => props.message.role === 'assistant');
const isSystem = computed(() => props.message.role === 'system');

const hasToolUses = computed(() =>
  props.message.toolUses && props.message.toolUses.length > 0
);

const formattedTime = computed(() => {
  const date = new Date(props.message.timestamp);
  return date.toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' });
});

// 简单的Markdown渲染（代码块、粗体、斜体、链接）
const renderedContent = computed(() => {
  let content = props.message.content;

  // 转义HTML
  content = content
    .replace(/&/g, '&amp;')
    .replace(/</g, '&lt;')
    .replace(/>/g, '&gt;');

  // 代码块 ```code```
  content = content.replace(/```(\w*)\n?([\s\S]*?)```/g, (_, lang, code) => {
    return `<pre class="code-block"><code class="language-${lang}">${code.trim()}</code></pre>`;
  });

  // 行内代码 `code`
  content = content.replace(/`([^`]+)`/g, '<code class="inline-code">$1</code>');

  // 粗体 **text**
  content = content.replace(/\*\*([^*]+)\*\*/g, '<strong>$1</strong>');

  // 斜体 *text*
  content = content.replace(/\*([^*]+)\*/g, '<em>$1</em>');

  // 链接 [text](url)
  content = content.replace(/\[([^\]]+)\]\(([^)]+)\)/g, '<a href="$2" target="_blank" rel="noopener">$1</a>');

  // 换行
  content = content.replace(/\n/g, '<br>');

  return content;
});
</script>

<template>
  <div
    class="claude-message"
    :class="{
      'message-user': isUser,
      'message-assistant': isAssistant,
      'message-system': isSystem
    }"
  >
    <!-- 头像 -->
    <div class="message-avatar">
      <i v-if="isUser" class="bx bx-user"></i>
      <i v-else-if="isAssistant" class="bx bx-bot"></i>
      <i v-else class="bx bx-info-circle"></i>
    </div>

    <!-- 消息内容 -->
    <div class="message-content">
      <div class="message-header">
        <span class="message-role">
          {{ isUser ? t('claude.message.you', 'You') : isAssistant ? t('claude.message.assistant', 'Assistant') : t('claude.message.system', 'System') }}
        </span>
        <span class="message-time">{{ formattedTime }}</span>
      </div>

      <!-- 文本内容 -->
      <div
        v-if="message.content"
        class="message-text"
        v-html="renderedContent"
      ></div>

      <!-- 工具调用 -->
      <div v-if="hasToolUses" class="message-tools">
        <ClaudeToolUse
          v-for="toolUse in message.toolUses"
          :key="toolUse.id"
          :toolUse="toolUse"
        />
      </div>
    </div>
  </div>
</template>

<style scoped>
.claude-message {
  display: flex;
  gap: 12px;
  padding: 12px 16px;
  border-radius: 8px;
  transition: background-color 0.2s ease;
}

.claude-message:hover {
  background: var(--color-bg-secondary);
}

/* 用户消息 */
.message-user {
  flex-direction: row-reverse;
}

.message-user .message-content {
  align-items: flex-end;
}

.message-user .message-text {
  background: var(--color-primary);
  color: white;
}

.message-user .message-text :deep(a) {
  color: rgba(255, 255, 255, 0.9);
}

.message-user .message-text :deep(.inline-code) {
  background: rgba(255, 255, 255, 0.2);
  color: white;
}

/* 助手消息 */
.message-assistant .message-text {
  background: var(--color-bg-tertiary);
  color: var(--color-text-primary);
}

/* 系统消息 */
.message-system {
  opacity: 0.8;
}

.message-system .message-text {
  background: var(--color-warning-bg, rgba(245, 158, 11, 0.1));
  color: var(--color-text-primary);
  font-style: italic;
}

/* 头像 */
.message-avatar {
  width: 36px;
  height: 36px;
  border-radius: 50%;
  background: var(--color-bg-tertiary);
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
}

.message-avatar i {
  font-size: 1.25rem;
  color: var(--color-text-secondary);
}

.message-user .message-avatar {
  background: var(--color-primary-bg, rgba(59, 130, 246, 0.1));
}

.message-user .message-avatar i {
  color: var(--color-primary);
}

.message-assistant .message-avatar {
  background: var(--color-success-bg, rgba(34, 197, 94, 0.1));
}

.message-assistant .message-avatar i {
  color: var(--color-success, #22c55e);
}

/* 消息内容 */
.message-content {
  display: flex;
  flex-direction: column;
  gap: 4px;
  max-width: 80%;
  min-width: 100px;
}

.message-header {
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: 0.75rem;
}

.message-role {
  font-weight: 600;
  color: var(--color-text-primary);
}

.message-time {
  color: var(--color-text-tertiary);
}

.message-text {
  padding: 10px 14px;
  border-radius: 12px;
  font-size: 0.9375rem;
  line-height: 1.6;
  word-wrap: break-word;
}

/* Markdown 样式 */
.message-text :deep(.code-block) {
  background: var(--color-bg-primary);
  border: 1px solid var(--color-border);
  border-radius: 6px;
  padding: 12px;
  margin: 8px 0;
  overflow-x: auto;
  font-family: 'Monaco', 'Menlo', 'Ubuntu Mono', monospace;
  font-size: 0.875rem;
  line-height: 1.5;
}

.message-text :deep(.inline-code) {
  background: var(--color-bg-secondary);
  padding: 2px 6px;
  border-radius: 4px;
  font-family: 'Monaco', 'Menlo', 'Ubuntu Mono', monospace;
  font-size: 0.875em;
}

.message-text :deep(a) {
  color: var(--color-primary);
  text-decoration: none;
}

.message-text :deep(a:hover) {
  text-decoration: underline;
}

.message-text :deep(strong) {
  font-weight: 600;
}

/* 工具调用区域 */
.message-tools {
  display: flex;
  flex-direction: column;
  gap: 8px;
  margin-top: 8px;
}
</style>
