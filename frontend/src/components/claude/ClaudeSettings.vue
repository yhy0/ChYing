<script setup lang="ts">
/**
 * Claude Settings Modal
 * 复用 AISettings 组件，以模态框形式展示
 */
import { useI18n } from 'vue-i18n';
import AISettings from '../settings/modules/AISettings.vue';

const emit = defineEmits<{
  (e: 'close'): void;
}>();

const { t } = useI18n();

// 关闭面板
const close = () => {
  emit('close');
};
</script>

<template>
  <div class="claude-settings-overlay" @click.self="close">
    <div class="claude-settings-panel">
      <!-- 头部 -->
      <div class="settings-header">
        <h3>
          <i class="bx bx-cog"></i>
          {{ t('claude.settings.title', 'Claude Code Settings') }}
        </h3>
        <button class="btn-close" @click="close">
          <i class="bx bx-x"></i>
        </button>
      </div>

      <!-- 内容 - 复用 AISettings 组件 -->
      <div class="settings-content scrollbar-thin">
        <AISettings :compact="true" @saved="close" />
      </div>
    </div>
  </div>
</template>

<style scoped>
.claude-settings-overlay {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background: rgba(0, 0, 0, 0.5);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 1000;
}

.claude-settings-panel {
  width: 680px;
  max-height: 90vh;
  background: var(--color-bg-primary);
  border-radius: 12px;
  box-shadow: 0 20px 40px rgba(0, 0, 0, 0.3);
  display: flex;
  flex-direction: column;
  overflow: hidden;
}

/* 头部 */
.settings-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 16px 20px;
  border-bottom: 1px solid var(--color-border);
  flex-shrink: 0;
}

.settings-header h3 {
  display: flex;
  align-items: center;
  gap: 8px;
  margin: 0;
  font-size: 1.125rem;
  color: var(--color-text-primary);
}

.settings-header h3 i {
  color: var(--color-primary);
}

.btn-close {
  width: 32px;
  height: 32px;
  border: none;
  background: transparent;
  border-radius: 6px;
  cursor: pointer;
  display: flex;
  align-items: center;
  justify-content: center;
  color: var(--color-text-secondary);
  transition: all 0.2s ease;
}

.btn-close:hover {
  background: var(--color-bg-tertiary);
  color: var(--color-text-primary);
}

.btn-close i {
  font-size: 1.5rem;
}

/* 内容 */
.settings-content {
  flex: 1;
  padding: 20px;
  padding-bottom: 24px;
  overflow-y: auto;
  max-height: calc(90vh - 80px); /* 减去头部高度，确保内容可滚动 */
}

/* 覆盖 AISettings 组件的一些样式以适应模态框 */
.settings-content :deep(.space-y-6) {
  max-width: none;
}

.settings-content :deep(h2) {
  display: none; /* 隐藏 AISettings 的标题，因为模态框已有标题 */
}

.settings-content :deep(.text-sm.text-gray-500.-mt-4) {
  display: none; /* 隐藏描述文字 */
}
</style>
