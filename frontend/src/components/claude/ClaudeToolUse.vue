<script setup lang="ts">
import { computed } from 'vue';
import { useI18n } from 'vue-i18n';
import type { ToolUse } from '../../types/claude';
import { getToolInfo } from '../../types/claude';

const props = defineProps<{
  toolUse: ToolUse;
}>();

const { t } = useI18n();

const toolInfo = computed(() => getToolInfo(props.toolUse.name));

const statusClass = computed(() => {
  switch (props.toolUse.status) {
    case 'pending':
      return 'status-pending';
    case 'confirmed':
      return 'status-confirmed';
    case 'rejected':
      return 'status-rejected';
    case 'completed':
      return 'status-completed';
    case 'error':
      return 'status-error';
    default:
      return '';
  }
});

const statusIcon = computed(() => {
  switch (props.toolUse.status) {
    case 'pending':
      return 'bx-time';
    case 'confirmed':
      return 'bx-loader-alt bx-spin';
    case 'rejected':
      return 'bx-x';
    case 'completed':
      return 'bx-check';
    case 'error':
      return 'bx-error';
    default:
      return 'bx-cog';
  }
});

const statusText = computed(() => {
  switch (props.toolUse.status) {
    case 'pending':
      return t('claude.tool.status.pending', 'Pending');
    case 'confirmed':
      return t('claude.tool.status.running', 'Running');
    case 'rejected':
      return t('claude.tool.status.rejected', 'Rejected');
    case 'completed':
      return t('claude.tool.status.completed', 'Completed');
    case 'error':
      return t('claude.tool.status.error', 'Error');
    default:
      return '';
  }
});

const formattedInput = computed(() => {
  try {
    return JSON.stringify(props.toolUse.input, null, 2);
  } catch {
    return String(props.toolUse.input);
  }
});

const hasResult = computed(() => !!props.toolUse.result);
const hasError = computed(() => !!props.toolUse.error);
</script>

<template>
  <div class="claude-tool-use" :class="[statusClass, { 'dangerous': toolInfo.dangerous }]">
    <!-- 工具头部 -->
    <div class="tool-header">
      <div class="tool-info">
        <i :class="['bx', toolInfo.icon]"></i>
        <span class="tool-name">{{ toolInfo.displayName }}</span>
        <span v-if="toolInfo.dangerous" class="danger-badge">
          <i class="bx bx-shield-x"></i>
          {{ t('claude.tool.dangerous', 'Dangerous') }}
        </span>
      </div>
      <div class="tool-status">
        <i :class="['bx', statusIcon]"></i>
        <span>{{ statusText }}</span>
      </div>
    </div>

    <!-- 工具描述 -->
    <div class="tool-description">
      {{ toolInfo.description }}
    </div>

    <!-- 输入参数 -->
    <div class="tool-section">
      <div class="section-label">
        <i class="bx bx-code-alt"></i>
        {{ t('claude.tool.input', 'Input') }}
      </div>
      <pre class="tool-code">{{ formattedInput }}</pre>
    </div>

    <!-- 执行结果 -->
    <div v-if="hasResult" class="tool-section">
      <div class="section-label">
        <i class="bx bx-check-circle"></i>
        {{ t('claude.tool.result', 'Result') }}
      </div>
      <pre class="tool-code tool-result">{{ toolUse.result }}</pre>
    </div>

    <!-- 错误信息 -->
    <div v-if="hasError" class="tool-section">
      <div class="section-label error-label">
        <i class="bx bx-error-circle"></i>
        {{ t('claude.tool.error', 'Error') }}
      </div>
      <pre class="tool-code tool-error">{{ toolUse.error }}</pre>
    </div>
  </div>
</template>

<style scoped>
.claude-tool-use {
  background: var(--color-bg-secondary);
  border: 1px solid var(--color-border);
  border-radius: 8px;
  padding: 12px;
  font-size: 0.875rem;
}

.claude-tool-use.dangerous {
  border-color: var(--color-warning, #f59e0b);
}

/* 状态样式 */
.status-pending {
  border-left: 3px solid var(--color-text-tertiary);
}

.status-confirmed {
  border-left: 3px solid var(--color-primary);
}

.status-rejected {
  border-left: 3px solid var(--color-danger, #ef4444);
  opacity: 0.7;
}

.status-completed {
  border-left: 3px solid var(--color-success, #22c55e);
}

.status-error {
  border-left: 3px solid var(--color-danger, #ef4444);
}

/* 头部 */
.tool-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 8px;
}

.tool-info {
  display: flex;
  align-items: center;
  gap: 8px;
}

.tool-info > i {
  font-size: 1.125rem;
  color: var(--color-primary);
}

.tool-name {
  font-weight: 600;
  color: var(--color-text-primary);
}

.danger-badge {
  display: flex;
  align-items: center;
  gap: 4px;
  padding: 2px 6px;
  background: var(--color-warning-bg, rgba(245, 158, 11, 0.1));
  color: var(--color-warning, #f59e0b);
  border-radius: 4px;
  font-size: 0.75rem;
  font-weight: 500;
}

.tool-status {
  display: flex;
  align-items: center;
  gap: 4px;
  font-size: 0.75rem;
  color: var(--color-text-secondary);
}

.status-completed .tool-status {
  color: var(--color-success, #22c55e);
}

.status-error .tool-status {
  color: var(--color-danger, #ef4444);
}

.status-rejected .tool-status {
  color: var(--color-danger, #ef4444);
}

/* 描述 */
.tool-description {
  color: var(--color-text-secondary);
  font-size: 0.8125rem;
  margin-bottom: 12px;
}

/* 区块 */
.tool-section {
  margin-top: 8px;
}

.section-label {
  display: flex;
  align-items: center;
  gap: 4px;
  font-size: 0.75rem;
  font-weight: 500;
  color: var(--color-text-secondary);
  margin-bottom: 4px;
}

.error-label {
  color: var(--color-danger, #ef4444);
}

.tool-code {
  background: var(--color-bg-primary);
  border: 1px solid var(--color-border);
  border-radius: 4px;
  padding: 8px;
  font-family: 'Monaco', 'Menlo', 'Ubuntu Mono', monospace;
  font-size: 0.8125rem;
  line-height: 1.5;
  overflow-x: auto;
  white-space: pre-wrap;
  word-break: break-word;
  max-height: 200px;
  overflow-y: auto;
  margin: 0;
}

.tool-result {
  border-color: var(--color-success, #22c55e);
  background: var(--color-success-bg, rgba(34, 197, 94, 0.05));
}

.tool-error {
  border-color: var(--color-danger, #ef4444);
  background: var(--color-danger-bg, rgba(239, 68, 68, 0.05));
  color: var(--color-danger, #ef4444);
}
</style>
