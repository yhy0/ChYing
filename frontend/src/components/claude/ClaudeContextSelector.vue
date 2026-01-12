<script setup lang="ts">
import { ref, computed, onMounted } from 'vue';
import { useI18n } from 'vue-i18n';
import type { AgentContext } from '../../types/claude';

const props = defineProps<{
  context: AgentContext;
}>();

const emit = defineEmits<{
  (e: 'update:context', context: AgentContext): void;
}>();

const { t } = useI18n();

// 本地状态
const autoCollect = ref(props.context.autoCollect);
const customData = ref(props.context.customData || '');
const selectedTrafficIds = ref<string[]>(props.context.selectedTrafficIds || []);
const selectedVulnIds = ref<string[]>(props.context.selectedVulnIds || []);
const selectedFingerprints = ref<string[]>(props.context.selectedFingerprints || []);

// 展开状态
const expandedSections = ref({
  auto: true,
  traffic: false,
  vulns: false,
  fingerprints: false,
  custom: false
});

// 切换展开
const toggleSection = (section: keyof typeof expandedSections.value) => {
  expandedSections.value[section] = !expandedSections.value[section];
};

// 更新上下文
const updateContext = () => {
  emit('update:context', {
    ...props.context,
    autoCollect: autoCollect.value,
    customData: customData.value,
    selectedTrafficIds: selectedTrafficIds.value,
    selectedVulnIds: selectedVulnIds.value,
    selectedFingerprints: selectedFingerprints.value
  });
};

// 清除选择
const clearSelection = (type: 'traffic' | 'vulns' | 'fingerprints') => {
  switch (type) {
    case 'traffic':
      selectedTrafficIds.value = [];
      break;
    case 'vulns':
      selectedVulnIds.value = [];
      break;
    case 'fingerprints':
      selectedFingerprints.value = [];
      break;
  }
  updateContext();
};

// 统计信息
const trafficCount = computed(() => selectedTrafficIds.value.length);
const vulnCount = computed(() => selectedVulnIds.value.length);
const fingerprintCount = computed(() => selectedFingerprints.value.length);
</script>

<template>
  <div class="claude-context-selector">
    <div class="selector-header">
      <h4>
        <i class="bx bx-data"></i>
        {{ t('claude.context.title', 'Context') }}
      </h4>
      <p class="selector-description">
        {{ t('claude.context.description', 'Configure what data the AI can access') }}
      </p>
    </div>

    <!-- 自动收集 -->
    <div class="context-section">
      <div class="section-header" @click="toggleSection('auto')">
        <div class="section-title">
          <i class="bx bx-refresh"></i>
          <span>{{ t('claude.context.autoCollect', 'Auto Collect') }}</span>
        </div>
        <i :class="['bx', expandedSections.auto ? 'bx-chevron-up' : 'bx-chevron-down']"></i>
      </div>
      <div v-if="expandedSections.auto" class="section-content">
        <label class="toggle-label">
          <input
            type="checkbox"
            v-model="autoCollect"
            @change="updateContext"
          />
          <span class="toggle-text">
            {{ t('claude.context.autoCollectDesc', 'Automatically include recent traffic and vulnerabilities') }}
          </span>
        </label>
      </div>
    </div>

    <!-- HTTP流量 -->
    <div class="context-section">
      <div class="section-header" @click="toggleSection('traffic')">
        <div class="section-title">
          <i class="bx bx-transfer"></i>
          <span>{{ t('claude.context.httpTraffic', 'HTTP Traffic') }}</span>
          <span v-if="trafficCount > 0" class="count-badge">{{ trafficCount }}</span>
        </div>
        <i :class="['bx', expandedSections.traffic ? 'bx-chevron-up' : 'bx-chevron-down']"></i>
      </div>
      <div v-if="expandedSections.traffic" class="section-content">
        <div v-if="trafficCount === 0" class="empty-state">
          <i class="bx bx-info-circle"></i>
          <span>{{ t('claude.context.noTrafficSelected', 'No traffic selected. Select from Proxy view.') }}</span>
        </div>
        <div v-else class="selection-info">
          <span>{{ t('claude.context.selectedItems', { count: trafficCount }) }}</span>
          <button class="btn-clear" @click="clearSelection('traffic')">
            <i class="bx bx-x"></i>
            {{ t('common.clear', 'Clear') }}
          </button>
        </div>
      </div>
    </div>

    <!-- 漏洞 -->
    <div class="context-section">
      <div class="section-header" @click="toggleSection('vulns')">
        <div class="section-title">
          <i class="bx bx-bug"></i>
          <span>{{ t('claude.context.vulnerabilities', 'Vulnerabilities') }}</span>
          <span v-if="vulnCount > 0" class="count-badge">{{ vulnCount }}</span>
        </div>
        <i :class="['bx', expandedSections.vulns ? 'bx-chevron-up' : 'bx-chevron-down']"></i>
      </div>
      <div v-if="expandedSections.vulns" class="section-content">
        <div v-if="vulnCount === 0" class="empty-state">
          <i class="bx bx-info-circle"></i>
          <span>{{ t('claude.context.noVulnsSelected', 'No vulnerabilities selected.') }}</span>
        </div>
        <div v-else class="selection-info">
          <span>{{ t('claude.context.selectedItems', { count: vulnCount }) }}</span>
          <button class="btn-clear" @click="clearSelection('vulns')">
            <i class="bx bx-x"></i>
            {{ t('common.clear', 'Clear') }}
          </button>
        </div>
      </div>
    </div>

    <!-- 指纹 -->
    <div class="context-section">
      <div class="section-header" @click="toggleSection('fingerprints')">
        <div class="section-title">
          <i class="bx bx-fingerprint"></i>
          <span>{{ t('claude.context.fingerprints', 'Fingerprints') }}</span>
          <span v-if="fingerprintCount > 0" class="count-badge">{{ fingerprintCount }}</span>
        </div>
        <i :class="['bx', expandedSections.fingerprints ? 'bx-chevron-up' : 'bx-chevron-down']"></i>
      </div>
      <div v-if="expandedSections.fingerprints" class="section-content">
        <div v-if="fingerprintCount === 0" class="empty-state">
          <i class="bx bx-info-circle"></i>
          <span>{{ t('claude.context.noFingerprintsSelected', 'No fingerprints selected.') }}</span>
        </div>
        <div v-else class="selection-info">
          <span>{{ t('claude.context.selectedItems', { count: fingerprintCount }) }}</span>
          <button class="btn-clear" @click="clearSelection('fingerprints')">
            <i class="bx bx-x"></i>
            {{ t('common.clear', 'Clear') }}
          </button>
        </div>
      </div>
    </div>

    <!-- 自定义数据 -->
    <div class="context-section">
      <div class="section-header" @click="toggleSection('custom')">
        <div class="section-title">
          <i class="bx bx-edit"></i>
          <span>{{ t('claude.context.customData', 'Custom Data') }}</span>
        </div>
        <i :class="['bx', expandedSections.custom ? 'bx-chevron-up' : 'bx-chevron-down']"></i>
      </div>
      <div v-if="expandedSections.custom" class="section-content">
        <textarea
          spellcheck="false"
          v-model="customData"
          @blur="updateContext"
          :placeholder="t('claude.context.customDataPlaceholder', 'Add custom context data here...')"
          class="custom-data-input scrollbar-thin"
          rows="4"
        ></textarea>
      </div>
    </div>
  </div>
</template>

<style scoped>
.claude-context-selector {
  padding: 16px;
}

.selector-header {
  margin-bottom: 16px;
}

.selector-header h4 {
  display: flex;
  align-items: center;
  gap: 8px;
  margin: 0 0 4px 0;
  font-size: 1rem;
  color: var(--color-text-primary);
}

.selector-header h4 i {
  color: var(--color-primary);
}

.selector-description {
  margin: 0;
  font-size: 0.8125rem;
  color: var(--color-text-secondary);
}

/* 区块 */
.context-section {
  background: var(--color-bg-primary);
  border: 1px solid var(--color-border);
  border-radius: 8px;
  margin-bottom: 8px;
  overflow: hidden;
}

.section-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 12px;
  cursor: pointer;
  transition: background-color 0.2s ease;
}

.section-header:hover {
  background: var(--color-bg-tertiary);
}

.section-title {
  display: flex;
  align-items: center;
  gap: 8px;
  font-weight: 500;
  color: var(--color-text-primary);
}

.section-title i {
  font-size: 1.125rem;
  color: var(--color-text-secondary);
}

.count-badge {
  padding: 2px 8px;
  background: var(--color-primary);
  color: white;
  border-radius: 10px;
  font-size: 0.75rem;
  font-weight: 600;
}

.section-content {
  padding: 12px;
  border-top: 1px solid var(--color-border);
  background: var(--color-bg-secondary);
}

/* 开关 */
.toggle-label {
  display: flex;
  align-items: flex-start;
  gap: 10px;
  cursor: pointer;
}

.toggle-label input[type="checkbox"] {
  width: 18px;
  height: 18px;
  margin-top: 2px;
  cursor: pointer;
}

.toggle-text {
  font-size: 0.875rem;
  color: var(--color-text-secondary);
  line-height: 1.5;
}

/* 空状态 */
.empty-state {
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: 0.8125rem;
  color: var(--color-text-tertiary);
}

/* 选择信息 */
.selection-info {
  display: flex;
  justify-content: space-between;
  align-items: center;
  font-size: 0.875rem;
  color: var(--color-text-secondary);
}

.btn-clear {
  display: flex;
  align-items: center;
  gap: 4px;
  padding: 4px 8px;
  background: transparent;
  border: 1px solid var(--color-border);
  border-radius: 4px;
  color: var(--color-text-secondary);
  font-size: 0.75rem;
  cursor: pointer;
  transition: all 0.2s ease;
}

.btn-clear:hover {
  background: var(--color-danger-bg, rgba(239, 68, 68, 0.1));
  border-color: var(--color-danger, #ef4444);
  color: var(--color-danger, #ef4444);
}

/* 自定义数据输入 */
.custom-data-input {
  width: 100%;
  padding: 10px;
  background: var(--color-bg-primary);
  border: 1px solid var(--color-border);
  border-radius: 6px;
  color: var(--color-text-primary);
  font-size: 0.875rem;
  font-family: inherit;
  resize: vertical;
  min-height: 80px;
}

.custom-data-input:focus {
  outline: none;
  border-color: var(--color-primary);
}

.custom-data-input::placeholder {
  color: var(--color-text-tertiary);
}
</style>
