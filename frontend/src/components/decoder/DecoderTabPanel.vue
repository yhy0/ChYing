<script setup lang="ts">
import { ref, computed } from 'vue';
import CryptoJS from 'crypto-js';
import { useModulesStore } from '../../store';
import type { DecoderTab, DecoderMethod, DecoderOperation, DecoderStep } from '../../types';
import { useI18n } from 'vue-i18n';
import { HttpRequestViewer } from '../common/codemirror';
import { generateUUID } from '../../utils';

// Props
const props = defineProps<{ tab: DecoderTab }>();

// Store
const store = useModulesStore();

// I18n
const { t } = useI18n();

// --- State --- 
const selectedStepId = ref<string>('initial');

// 添加换行状态管理
const wordWrap = ref<Record<string, boolean>>({
  initial: true
});

// 更新标签状态的便捷方法
const updateTab = (updates: Partial<DecoderTab>) => {
  store.updateDecoderTab(props.tab.id, updates);
};

// --- Computed Properties --- 
// 计算选中的步骤对象
const selectedStep = computed(() => {
  if (selectedStepId.value === 'initial' || !props.tab) {
    return null;
  }
  return props.tab.steps.find(step => step.id === selectedStepId.value);
});

// 计算当前选中的卡片的输出 (作为下一个操作的输入)
const currentInputForNewStep = computed(() => {
  if (selectedStepId.value === 'initial' || !props.tab) {
    return props.tab?.initialInput || '';
  }
  const step = selectedStep.value;
  return step && !step.error ? step.outputText : '';
});

// 计算最后一个有效步骤的输出 (用于复制按钮)
const lastValidOutput = computed(() => {
  if (!props.tab || props.tab.steps.length === 0) {
    return props.tab?.initialInput || '';
  }
  // 从后往前找第一个没有错误的步骤
  for (let i = props.tab.steps.length - 1; i >= 0; i--) {
    if (!props.tab.steps[i].error) {
      return props.tab.steps[i].outputText;
    }
  }
  return props.tab.initialInput;
});

// --- Methods --- 
// 设置选中的步骤
const selectStep = (id: string) => {
  selectedStepId.value = id;
};

// 定义所有方法及其属性
const availableMethods: { id: DecoderMethod; name: string; encode: boolean; decode: boolean; isHash?: boolean }[] = [
  { id: 'URL', name: 'URL', encode: true, decode: true },
  { id: 'HTML', name: 'HTML', encode: true, decode: true }, 
  { id: 'Base64', name: 'Base64', encode: true, decode: true },
  { id: 'Hex', name: 'Hex', encode: true, decode: true }, 
  { id: 'Unicode', name: 'Unicode', encode: true, decode: true },
  { id: 'MD5', name: 'MD5', encode: true, decode: false, isHash: true },
  { id: 'SHA1', name: 'SHA-1', encode: true, decode: false, isHash: true },
  { id: 'SHA256', name: 'SHA-256', encode: true, decode: false, isHash: true },
];

// 单步处理函数
const processSingleStep = (methodId: DecoderMethod, operation: DecoderOperation, inputText: string): { outputText: string; error?: string } => {
  let outputText = '';
  let error: string | undefined = undefined;

  if (!inputText && methodId !== 'MD5' && methodId !== 'SHA1' && methodId !== 'SHA256') {
    return { outputText: '', error: 'Input is empty' };
  }

  try {
    switch (methodId) {
      case 'URL': outputText = operation === 'encode' ? encodeURIComponent(inputText) : decodeURIComponent(inputText); break;
      case 'Base64': outputText = operation === 'encode' ? btoa(unescape(encodeURIComponent(inputText))) : decodeURIComponent(escape(atob(inputText))); break;
      case 'HTML':
        if (operation === 'encode') {
          outputText = inputText.replace(/&/g, '&amp;').replace(/</g, '&lt;').replace(/>/g, '&gt;').replace(/"/g, '&quot;').replace(/'/g, '&#039;');
        } else {
          const ta = document.createElement('textarea'); 
          ta.innerHTML = inputText; 
          outputText = ta.value; 
        }
        break;
      case 'Unicode': outputText = operation === 'encode' ? 
        inputText.split('').map(char => '\\u' + char.charCodeAt(0).toString(16).padStart(4, '0')).join('') : 
        inputText.replace(/\\u([0-9a-fA-F]{4})/g, (_, hex) => String.fromCharCode(parseInt(hex, 16))); 
        break;
      case 'Hex': outputText = operation === 'encode' ? 
        inputText.split('').map(char => char.charCodeAt(0).toString(16).padStart(2, '0')).join('') : 
        inputText.match(/.{1,2}/g)?.map(hex => String.fromCharCode(parseInt(hex, 16))).join('') || ''; 
        break;
      case 'MD5': if (operation === 'encode') outputText = CryptoJS.MD5(inputText).toString(); else throw new Error('MD5 decode not supported'); break;
      case 'SHA1': if (operation === 'encode') outputText = CryptoJS.SHA1(inputText).toString(); else throw new Error('SHA1 decode not supported'); break;
      case 'SHA256': if (operation === 'encode') outputText = CryptoJS.SHA256(inputText).toString(); else throw new Error('SHA256 decode not supported'); break;
      default: throw new Error('Unsupported method');
    }
          } catch (e) {
    error = e instanceof Error ? e.message : 'Processing error';
    outputText = `Error: ${error}`;
  }
  return { outputText, error };
};

// 重新计算从指定索引开始的所有步骤
const recalculateSteps = (startIndex: number, currentSteps: DecoderStep[]): DecoderStep[] => {
  const recalculatedSteps = [...currentSteps];
  for (let i = startIndex; i < recalculatedSteps.length; i++) {
    const previousOutput = i === 0 ? (props.tab?.initialInput || '') : (recalculatedSteps[i - 1].error ? '' : recalculatedSteps[i - 1].outputText);
    const { outputText, error } = processSingleStep(recalculatedSteps[i].method, recalculatedSteps[i].operation, previousOutput);
    recalculatedSteps[i] = { ...recalculatedSteps[i], inputText: previousOutput, outputText, error };
  }
  return recalculatedSteps;
};

// 插入新步骤并重新计算后续步骤
const insertAndRecalculate = (methodId: DecoderMethod, operation: DecoderOperation) => {
  if (!props.tab) return;
  const methodInfo = availableMethods.find(m => m.id === methodId);
  if (!methodInfo) return;

  const isValidOp = operation === 'encode' ? methodInfo.encode : methodInfo.decode;
  if (!isValidOp) return;

  const currentInput = currentInputForNewStep.value;
  const { outputText, error } = processSingleStep(methodId, operation, currentInput);

  const newStep: DecoderStep = {
    id: generateUUID(),
    method: methodId,
    operation: operation,
    inputText: currentInput,
    outputText: outputText,
    error: error,
  };

  // 查找选中步骤的索引
  const selectedIndex = selectedStepId.value === 'initial' 
    ? -1  // 初始输入框
    : props.tab.steps.findIndex(step => step.id === selectedStepId.value);
  
  // 插入点位置
  const insertionIndex = selectedIndex === -1 ? 0 : selectedIndex + 1;
  
  // 创建新的步骤数组，但只保留选中步骤及之前的步骤
  const newSteps = selectedIndex === -1 
    ? [] // 如果选中的是初始输入，则清空所有步骤
    : [...props.tab.steps.slice(0, insertionIndex)]; // 否则只保留选中步骤之前(含)的部分
  
  // 添加新步骤
  newSteps.push(newStep);
  
  // 更新tab中的步骤
  updateTab({ steps: newSteps });
  
  // 选中新添加的步骤
  selectStep(newStep.id);
};

// 处理步骤输出更新
const handleStepOutputUpdate = (data: string, stepId: string) => {
  if (!props.tab) return;
  
  const stepIndex = props.tab.steps.findIndex(step => step.id === stepId);
  if (stepIndex === -1) return;
  
  const updatedSteps = [...props.tab.steps];
  updatedSteps[stepIndex] = {
    ...updatedSteps[stepIndex],
    outputText: data,
    error: undefined,
  };
  
  const finalSteps = recalculateSteps(stepIndex + 1, updatedSteps);
  updateTab({ steps: finalSteps });
  selectStep(stepId);
};

// 检查步骤是否被手动编辑过
const isManuallyEdited = (step: DecoderStep, index: number): boolean => {
  if (index === 0) {
    return step.inputText !== props.tab?.initialInput;
  } else if (index > 0 && props.tab && props.tab.steps.length > index - 1) {
    const prevStep = props.tab.steps[index - 1];
    return step.inputText !== prevStep.outputText;
  }
  return false;
};

// 处理初始输入更新
const handleInitialInputUpdate = (data: string) => {
  if (!props.tab) return;
  const recalculatedSteps = recalculateSteps(0, props.tab.steps);
  updateTab({ initialInput: data, steps: recalculatedSteps });
};

// 清空所有
const clearAll = () => {
  if (props.tab) {
    updateTab({ initialInput: '', steps: [] });
    selectStep('initial');
  }
};

// 复制最后一个有效输出
const copyLastOutput = async () => {
  const outputToCopy = lastValidOutput.value;
  if (!outputToCopy) return;
  try {
    await navigator.clipboard.writeText(outputToCopy);
    console.log('Copied!'); 
  } catch (err) {
    console.error('Failed to copy text: ', err);
  }
};

// 切换换行状态
const toggleWordWrap = (id: string) => {
  if (!wordWrap.value[id]) {
    wordWrap.value[id] = true;
  } else {
    wordWrap.value[id] = false;
  }
};

// 检查指定ID是否启用了自动换行
const isWordWrapEnabled = (id: string): boolean => {
  // 默认启用自动换行
  if (wordWrap.value[id] === undefined) return true;
  return !!wordWrap.value[id];
};
</script>

<template>
  <div class="decoder-panel-container-vertical" v-if="props.tab">
    <!-- 操作按钮区域 -->
    <div class="decoder-actions-top-rows">
      <div class="button-rows-container">
        <div class="buttons-row">
          <template v-for="method in availableMethods" :key="`${method.id}-encode`">
            <button 
              v-if="method.encode" 
              class="btn btn-action encode" 
              :class="{ 'hash': method.isHash }" 
              @click="insertAndRecalculate(method.id, 'encode')"
              :title="`${method.name} ${method.isHash ? 'Hash' : 'Encode'}`"
              :disabled="!props.tab"
            >
              {{ method.name }} {{ method.isHash ? '' : t('modules.decoder.encode') }}
            </button>
            <div v-else-if="method.decode && !method.isHash" class="btn-placeholder"></div>
          </template>
        </div>
        <div class="buttons-row">
          <template v-for="method in availableMethods" :key="`${method.id}-decode`">
            <button 
              v-if="method.decode" 
              class="btn btn-action decode" 
              @click="insertAndRecalculate(method.id, 'decode')"
              :title="`${method.name} Decode`"
              :disabled="!props.tab"
            >
              {{ method.name }} {{ t('modules.decoder.decode') }}
            </button>
            <div v-else class="btn-placeholder"></div> 
          </template>
        </div>
      </div>
      <div class="global-actions-top">
            <button 
          class="btn btn-secondary btn-copy"
          @click="copyLastOutput"
          :disabled="!lastValidOutput || !props.tab"
          :title="t('modules.decoder.copy_output')"
        >
          <i class="bx bx-copy"></i>
            </button>
            <button 
              class="btn btn-secondary btn-clear"
              @click="clearAll"
          :disabled="!props.tab"
          :title="t('modules.decoder.clear_all')"
            >
          <i class="bx bx-trash"></i>
            </button>
      </div>
    </div>
    
    <!-- 初始输入面板 -->
    <div 
      class="decoder-step-card initial-input-card"
      :class="{ 'selected': selectedStepId === 'initial' }"
      @click="selectStep('initial')"
    >
      <div class="decoder-step-header">
        <h3 class="step-title">{{ t('modules.decoder.initial_input') }}</h3>
        <div class="step-actions">
          <button 
            class="btn-icon tooltip-container"
            @click.stop="toggleWordWrap('initial')"
            :class="{ 'active': isWordWrapEnabled('initial') }"
          >
            <i class="bx bx-text"></i>
            <span class="tooltip-text">{{ t('modules.decoder.toggle_word_wrap') }}</span>
          </button>
        <span class="char-count">{{ props.tab?.initialInput?.length || 0 }} {{ t('modules.decoder.chars') }}</span>
        </div>
              </div>
      <div class="editor-wrapper">
            <HttpRequestViewer
          :data="props.tab?.initialInput || ''"
              :read-only="false"
          @update:data="handleInitialInputUpdate"
          :class="{ 'word-wrap': isWordWrapEnabled('initial') }"
            />
          </div>
        </div>
        
    <!-- 步骤列表区域 -->
    <div class="decoder-steps-scrollable-area">
      <div 
        v-if="props.tab.steps"
        v-for="(step, index) in props.tab.steps" 
        :key="step.id" 
        class="decoder-step-card"
        :class="{ 'selected': selectedStepId === step.id }"
        @click="selectStep(step.id)"
      >
        <div class="decoder-step-header">
          <h3 class="step-title">
            <span class="step-number">{{ index + 1 }}.</span>
            {{ step.method }} - {{ step.operation === 'encode' ? t('modules.decoder.encode') : t('modules.decoder.decode') }}
            <span v-if="isManuallyEdited(step, index)" class="manual-edit-badge">{{ t('modules.decoder.manual_edit') }}</span>
          </h3>
          <div class="step-actions">
            <button 
              class="btn-icon tooltip-container"
              @click.stop="toggleWordWrap(step.id)"
              :class="{ 'active': isWordWrapEnabled(step.id) }"
            >
              <i class="bx bx-text"></i>
              <span class="tooltip-text">{{ t('modules.decoder.toggle_word_wrap') }}</span>
            </button>
          <span v-if="!step.error" class="char-count">{{ step.outputText.length }} {{ t('modules.decoder.chars') }}</span>
          <span v-else class="error-badge">{{ t('common.error') }}</span>
          </div>
            </div>
        <div class="editor-wrapper">
          <div v-if="step.error" class="error-message">
            {{ step.error }}
            </div>
            <HttpRequestViewer
            :data="step.outputText" 
            :read-only="false" 
            @update:data="(data) => handleStepOutputUpdate(data, step.id)"
            :class="{ 'word-wrap': isWordWrapEnabled(step.id) }"
          />
        </div>
      </div>
    </div>
  </div>

  <!-- 错误提示 -->
  <div v-else class="decoder-error-container">
    <p>{{ t('modules.decoder.no_tab_selected') }}</p>
  </div>
</template> 

<style scoped>
.decoder-panel-container-vertical {
  display: flex;
  flex-direction: column;
  height: 100%;
  padding: 1rem;
  background-color: var(--bg-color, #ffffff);
  color: var(--text-color, #1f2937);
}

.decoder-actions-top-rows {
  display: flex; 
  justify-content: space-between; 
  padding-bottom: 1rem; 
  border-bottom: 1px solid var(--border-color-light, #f3f4f6); 
  margin-bottom: 1rem; 
  flex-shrink: 0;
}

.button-rows-container {
  display: flex;
  flex-direction: column; 
  gap: 0.5rem; 
  flex-grow: 1; 
}

.buttons-row {
  display: flex;
  flex-wrap: nowrap; 
  gap: 0.5rem;
}

.btn-action, .btn-placeholder {
  flex: 1 1 0px; 
  min-width: 80px; 
  max-width: 150px; 
  padding: 0.3rem 0.6rem;
  font-size: 0.75rem;
  border-radius: 4px;
  text-align: center;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.btn-placeholder {
  visibility: hidden; 
  border: 1px solid transparent; 
}

.global-actions-top {
  display: flex;
  flex-direction: column; 
  gap: 0.5rem;
  flex-shrink: 0; 
  align-items: flex-start; 
}

.initial-input-card {
  flex-shrink: 0; 
}

.decoder-steps-scrollable-area {
  flex-grow: 1; 
  overflow-y: auto; 
}

.decoder-step-card {
  background-color: var(--surface-color, #f9fafb);
  border: 1px solid var(--border-color, #e5e7eb);
  border-radius: 6px;
  margin-bottom: 1rem;
  overflow: hidden;
  cursor: pointer;
  transition: border-color 0.2s ease, box-shadow 0.2s ease;
}

.decoder-step-card:hover {
  border-color: var(--primary-color-light, rgba(79, 70, 229, 0.5));
}

.decoder-step-card.selected {
  border-color: var(--primary-color, #4f46e5);
  box-shadow: 0 0 0 2px var(--primary-color-light, rgba(79, 70, 229, 0.3));
}

.decoder-step-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 0.5rem 0.75rem;
  border-bottom: 1px solid var(--border-color, #e5e7eb);
  background-color: var(--header-bg-color, #f3f4f6);
}

.step-title {
  font-size: 0.875rem; 
  font-weight: 500;
  color: var(--text-color-medium, #4b5563);
}

.step-number {
  color: var(--primary-color, #4f46e5);
  font-weight: 600;
  margin-right: 0.25rem;
}

.char-count, .error-badge, .manual-edit-badge {
  font-size: 0.75rem; 
  color: var(--text-color-light, #6b7280);
  background-color: var(--badge-bg-color, #e5e7eb);
  padding: 0.125rem 0.5rem;
  border-radius: 9999px;
}

.error-badge {
  background-color: var(--error-bg-light, #fee2e2);
  color: var(--error-color, #dc2626);
}

.manual-edit-badge {
  background-color: var(--warning-bg-light, #fef3c7);
  color: var(--warning-color, #d97706);
  margin-left: 0.5rem;
}

.editor-wrapper {
  min-height: 100px;
  height: auto;
  max-height: 300px; 
  overflow-y: auto;
  position: relative;
}

.editor-wrapper :deep(.cm-editor) {
  height: 100%;
}

.editor-wrapper :deep(.cm-scroller) {
  overflow-y: auto;
}

.error-message {
  padding: 1rem;
  font-size: 0.875rem;
  color: var(--error-color, #dc2626);
  background-color: var(--error-bg-light, #fee2e2);
  white-space: pre-wrap;
}

.btn-action {
  border: 1px solid transparent;
  cursor: pointer;
  transition: all 0.2s ease;
}

.btn-action.encode {
  background-color: var(--info-bg-light, #dbeafe);
  color: var(--info-color, #2563eb);
  border-color: var(--info-border, #bfdbfe);
}

.btn-action.encode.hash {
  background-color: var(--warning-bg-light, #fef3c7);
  color: var(--warning-color, #d97706);
  border-color: var(--warning-border, #fde68a);
}

.btn-action.encode:hover {
  background-color: var(--info-bg-hover, #bfdbfe);
}

.btn-action.encode.hash:hover {
  background-color: var(--warning-bg-hover, #fde047);
}

.btn-action.decode {
  background-color: var(--success-bg-light, #d1fae5);
  color: var(--success-color, #059669);
  border-color: var(--success-border, #a7f3d0);
}

.btn-action.decode:hover {
  background-color: var(--success-bg-hover, #6ee7b7);
}

.btn-action:disabled, .btn-secondary:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.btn-secondary {
  padding: 0.5rem;
  background-color: var(--secondary-btn-bg, #e5e7eb);
  color: var(--secondary-btn-text, #374151);
  border: 1px solid var(--secondary-btn-border, #d1d5db);
  border-radius: 4px;
  cursor: pointer;
  transition: background-color 0.2s ease;
}

.btn-secondary:hover {
  background-color: var(--secondary-btn-bg-hover, #d1d5db);
}

.btn-secondary i {
  font-size: 1.1rem;
}

.decoder-error-container {
  display: flex;
  justify-content: center;
  align-items: center;
  height: 100%;
  color: var(--text-color-medium, #4b5563);
  font-size: 1rem;
}

/* 深色模式变量 */
.dark .decoder-panel-container-vertical {
  --bg-color: #111827; 
  --text-color: #d1d5db;
  --surface-color: #1f2937;
  --border-color: #374151;
  --header-bg-color: #374151;
  --text-color-medium: #9ca3af;
  --text-color-light: #6b7280;
  --badge-bg-color: #374151;
  --error-bg-light: rgba(220, 38, 38, 0.2);
  --error-color: #f87171;
  --info-bg-light: rgba(59, 130, 246, 0.2);
  --info-color: #93c5fd;
  --info-border: rgba(59, 130, 246, 0.4);
  --info-bg-hover: rgba(59, 130, 246, 0.3);
  --success-bg-light: rgba(16, 185, 129, 0.2);
  --success-color: #6ee7b7;
  --success-border: rgba(16, 185, 129, 0.4);
  --success-bg-hover: rgba(16, 185, 129, 0.3);
  --warning-bg-light: rgba(245, 158, 11, 0.15);
  --warning-color: #fcd34d;
  --warning-border: rgba(245, 158, 11, 0.3);
  --warning-bg-hover: rgba(245, 158, 11, 0.25);
  --secondary-btn-bg: #374151;
  --secondary-btn-text: #d1d5db;
  --secondary-btn-border: #4b5563;
  --secondary-btn-bg-hover: #4b5563;
  --border-color-light: #2b364a; 
  --primary-color: #6366f1;
  --primary-color-light: rgba(99, 102, 241, 0.3);
}

.step-actions {
  display: flex;
  align-items: center;
  gap: 0.5rem;
}

.btn-icon {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 24px;
  height: 24px;
  border-radius: 4px;
  color: var(--text-color-light, #6b7280);
  background: transparent;
  border: none;
  cursor: pointer;
  transition: all 0.2s ease;
}

.btn-icon:hover {
  color: var(--primary-color, #4f46e5);
  background-color: var(--bg-hover, rgba(79, 70, 229, 0.1));
}

.btn-icon.active {
  color: var(--primary-color, #4f46e5);
  background-color: var(--bg-hover, rgba(79, 70, 229, 0.1));
}

.btn-icon i {
  font-size: 0.95rem;
}

.tooltip-container {
  position: relative;
}

.tooltip-text {
  visibility: hidden;
  position: absolute;
  bottom: -30px;
  left: 50%;
  transform: translateX(-50%);
  background-color: rgba(0, 0, 0, 0.75);
  color: #fff;
  text-align: center;
  border-radius: 4px;
  padding: 4px 8px;
  font-size: 0.7rem;
  white-space: nowrap;
  z-index: 100;
  pointer-events: none;
  opacity: 0;
  transition: opacity 0.2s;
}

.tooltip-container:hover .tooltip-text {
  visibility: visible;
  opacity: 1;
}
</style> 