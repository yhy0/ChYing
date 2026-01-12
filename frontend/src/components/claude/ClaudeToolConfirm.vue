<script setup lang="ts">
import { computed } from 'vue';
import { useI18n } from 'vue-i18n';
import type { ToolUse } from '../../types/claude';
import { getToolInfo } from '../../types/claude';

const props = defineProps<{
  toolUses: ToolUse[];
}>();

const emit = defineEmits<{
  (e: 'confirm', toolUseId: string, confirmed: boolean): void;
}>();

const { t } = useI18n();

// 获取待确认的工具
const pendingTools = computed(() =>
  props.toolUses.filter(tool => tool.status === 'pending')
);

// 当前工具
const currentTool = computed(() => pendingTools.value[0]);
const currentToolInfo = computed(() =>
  currentTool.value ? getToolInfo(currentTool.value.name) : null
);

// 格式化输入
const formattedInput = computed(() => {
  if (!currentTool.value) return '';
  try {
    return JSON.stringify(currentTool.value.input, null, 2);
  } catch {
    return String(currentTool.value.input);
  }
});

// 确认执行
const confirmTool = () => {
  if (currentTool.value) {
    emit('confirm', currentTool.value.id, true);
  }
};

// 拒绝执行
const rejectTool = () => {
  if (currentTool.value) {
    emit('confirm', currentTool.value.id, false);
  }
};

// 确认所有
const confirmAll = () => {
  pendingTools.value.forEach(tool => {
    emit('confirm', tool.id, true);
  });
};

// 拒绝所有
const rejectAll = () => {
  pendingTools.value.forEach(tool => {
    emit('confirm', tool.id, false);
  });
};
</script>

<template>
  <div class="tool-confirm-overlay">
    <div class="tool-confirm-modal">
      <!-- 头部 -->
      <div class="modal-header">
        <div class="header-icon warning">
          <i class="bx bx-shield-x"></i>
        </div>
        <h3>{{ t('claude.toolConfirm.title', 'Confirm Tool Execution') }}</h3>
        <p class="header-description">
          {{ t('claude.toolConfirm.description', 'The AI wants to execute the following operation. Please review and confirm.') }}
        </p>
      </div>

      <!-- 工具信息 -->
      <div v-if="currentTool && currentToolInfo" class="modal-content scrollbar-thin">
        <!-- 工具名称 -->
        <div class="tool-info-card">
          <div class="tool-header">
            <i :class="['bx', currentToolInfo.icon]"></i>
            <span class="tool-name">{{ currentToolInfo.displayName }}</span>
            <span v-if="currentToolInfo.dangerous" class="danger-badge">
              <i class="bx bx-error"></i>
              {{ t('claude.tool.dangerous', 'Dangerous') }}
            </span>
          </div>
          <p class="tool-description">{{ currentToolInfo.description }}</p>
        </div>

        <!-- 输入参数 -->
        <div class="params-section">
          <h4>
            <i class="bx bx-code-alt"></i>
            {{ t('claude.toolConfirm.parameters', 'Parameters') }}
          </h4>
          <pre class="params-code">{{ formattedInput }}</pre>
        </div>

        <!-- 警告信息 -->
        <div v-if="currentToolInfo.dangerous" class="warning-section">
          <i class="bx bx-info-circle"></i>
          <div class="warning-content">
            <strong>{{ t('claude.toolConfirm.warningTitle', 'Warning') }}</strong>
            <p>{{ t('claude.toolConfirm.warningText', 'This operation may send requests to external servers or modify data. Make sure you trust the target before proceeding.') }}</p>
          </div>
        </div>

        <!-- 待处理数量 -->
        <div v-if="pendingTools.length > 1" class="pending-count">
          <i class="bx bx-list-ul"></i>
          {{ t('claude.toolConfirm.pendingCount', { count: pendingTools.length - 1 }) }}
        </div>
      </div>

      <!-- 底部按钮 -->
      <div class="modal-footer">
        <div class="footer-left">
          <button
            v-if="pendingTools.length > 1"
            class="btn btn-sm btn-secondary"
            @click="rejectAll"
          >
            {{ t('claude.toolConfirm.rejectAll', 'Reject All') }}
          </button>
          <button
            v-if="pendingTools.length > 1"
            class="btn btn-sm btn-warning"
            @click="confirmAll"
          >
            {{ t('claude.toolConfirm.confirmAll', 'Confirm All') }}
          </button>
        </div>
        <div class="footer-right">
          <button class="btn btn-secondary" @click="rejectTool">
            <i class="bx bx-x"></i>
            {{ t('claude.toolConfirm.reject', 'Reject') }}
          </button>
          <button class="btn btn-primary" @click="confirmTool">
            <i class="bx bx-check"></i>
            {{ t('claude.toolConfirm.confirm', 'Confirm') }}
          </button>
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
.tool-confirm-overlay {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background: rgba(0, 0, 0, 0.6);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 1100;
}

.tool-confirm-modal {
  width: 520px;
  max-height: 80vh;
  background: var(--color-bg-primary);
  border-radius: 12px;
  box-shadow: 0 20px 40px rgba(0, 0, 0, 0.4);
  display: flex;
  flex-direction: column;
  overflow: hidden;
}

/* 头部 */
.modal-header {
  padding: 24px 24px 16px;
  text-align: center;
  border-bottom: 1px solid var(--color-border);
}

.header-icon {
  width: 56px;
  height: 56px;
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  margin: 0 auto 16px;
}

.header-icon.warning {
  background: var(--color-warning-bg, rgba(245, 158, 11, 0.1));
}

.header-icon i {
  font-size: 28px;
  color: var(--color-warning, #f59e0b);
}

.modal-header h3 {
  margin: 0 0 8px 0;
  font-size: 1.25rem;
  color: var(--color-text-primary);
}

.header-description {
  margin: 0;
  font-size: 0.875rem;
  color: var(--color-text-secondary);
  line-height: 1.5;
}

/* 内容 */
.modal-content {
  flex: 1;
  padding: 20px 24px;
  overflow-y: auto;
}

/* 工具信息卡片 */
.tool-info-card {
  background: var(--color-bg-secondary);
  border: 1px solid var(--color-border);
  border-radius: 8px;
  padding: 16px;
  margin-bottom: 16px;
}

.tool-header {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-bottom: 8px;
}

.tool-header > i {
  font-size: 1.25rem;
  color: var(--color-primary);
}

.tool-name {
  font-weight: 600;
  font-size: 1rem;
  color: var(--color-text-primary);
}

.danger-badge {
  display: flex;
  align-items: center;
  gap: 4px;
  padding: 2px 8px;
  background: var(--color-warning-bg, rgba(245, 158, 11, 0.1));
  color: var(--color-warning, #f59e0b);
  border-radius: 4px;
  font-size: 0.75rem;
  font-weight: 500;
  margin-left: auto;
}

.tool-description {
  margin: 0;
  font-size: 0.875rem;
  color: var(--color-text-secondary);
}

/* 参数区域 */
.params-section {
  margin-bottom: 16px;
}

.params-section h4 {
  display: flex;
  align-items: center;
  gap: 6px;
  margin: 0 0 8px 0;
  font-size: 0.875rem;
  font-weight: 600;
  color: var(--color-text-primary);
}

.params-code {
  background: var(--color-bg-secondary);
  border: 1px solid var(--color-border);
  border-radius: 6px;
  padding: 12px;
  font-family: 'Monaco', 'Menlo', 'Ubuntu Mono', monospace;
  font-size: 0.8125rem;
  line-height: 1.5;
  overflow-x: auto;
  white-space: pre-wrap;
  word-break: break-word;
  max-height: 150px;
  overflow-y: auto;
  margin: 0;
}

/* 警告区域 */
.warning-section {
  display: flex;
  gap: 12px;
  padding: 12px;
  background: var(--color-warning-bg, rgba(245, 158, 11, 0.1));
  border: 1px solid var(--color-warning, #f59e0b);
  border-radius: 8px;
  margin-bottom: 16px;
}

.warning-section > i {
  font-size: 1.25rem;
  color: var(--color-warning, #f59e0b);
  flex-shrink: 0;
}

.warning-content strong {
  display: block;
  margin-bottom: 4px;
  color: var(--color-warning, #f59e0b);
  font-size: 0.875rem;
}

.warning-content p {
  margin: 0;
  font-size: 0.8125rem;
  color: var(--color-text-secondary);
  line-height: 1.5;
}

/* 待处理数量 */
.pending-count {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 10px 12px;
  background: var(--color-bg-tertiary);
  border-radius: 6px;
  font-size: 0.8125rem;
  color: var(--color-text-secondary);
}

/* 底部 */
.modal-footer {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 16px 24px;
  border-top: 1px solid var(--color-border);
  background: var(--color-bg-secondary);
}

.footer-left,
.footer-right {
  display: flex;
  gap: 8px;
}

.modal-footer .btn {
  display: flex;
  align-items: center;
  gap: 6px;
}
</style>
