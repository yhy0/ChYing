<script setup lang="ts">
import { computed } from 'vue';
import { useI18n } from 'vue-i18n';
import type { ScanStatistics } from '../../types/scanLog';
import { getModuleColor } from '../../utils/colors';

const { t } = useI18n();

const props = defineProps<{
  statistics: ScanStatistics;
}>();

// 计算模块统计数据
const moduleStats = computed(() => {
  return Object.entries(props.statistics.byModule)
    .sort((a, b) => b[1] - a[1]) // 按数量降序排列
    .map(([moduleName, count]) => {
      const moduleColor = getModuleColor(moduleName);
      return {
        key: moduleName,
        label: moduleName,
        value: count,
        icon: getModuleIcon(moduleName),
        colorClass: 'status-info',
        color: moduleColor.color,
        textClass: moduleColor.textClass,
        percentage: props.statistics.total > 0 ? (count / props.statistics.total * 100).toFixed(1) : '0'
      };
    });
});

// 根据模块名称获取合适的图标
const getModuleIcon = (moduleName: string): string => {
  const name = moduleName.toLowerCase();
  if (name.includes('xss')) return 'bx-code-alt';
  if (name.includes('sql')) return 'bx-data';
  if (name.includes('auth')) return 'bx-shield-quarter';
  if (name.includes('upload')) return 'bx-upload';
  if (name.includes('xxe')) return 'bx-file-blank';
  if (name.includes('ssrf')) return 'bx-link-external';
  if (name.includes('cmd') || name.includes('inject')) return 'bx-terminal';
  if (name.includes('sensitive')) return 'bx-shield-alt-2';
  if (name.includes('waf')) return 'bx-shield-x';
  if (name.includes('brute')) return 'bx-key';
  if (name.includes('finger')) return 'bx-fingerprint';
  if (name.includes('port') || name.includes('scan')) return 'bx-search';
  if (name.includes('subdomain') || name.includes('domain')) return 'bx-network-chart';
  if (name.includes('directory') || name.includes('bbscan')) return 'bx-folder';
  if (name.includes('json')) return 'bx-bug';
  if (name.includes('swagger')) return 'bx-file-doc';
  if (name.includes('raw')) return 'bx-code';
  if (name.includes('http')) return 'bx-transfer-alt';
  return 'bx-cog';
};
</script>

<template>
  <div class="statistics-container scrollbar-thin">
    <!-- 总览统计 -->
    <div class="stats-section">
      <div class="section-title">
        <i class="bx bx-bar-chart-alt-2"></i>
        <span>{{ t('scanLog.statistics.overview', '总览') }}</span>
      </div>

      <div class="overview-card">
        <div class="overview-icon">
          <i class="bx bx-list-ul"></i>
        </div>
        <div class="overview-content">
          <div class="overview-number">{{ statistics.total.toLocaleString() }}</div>
          <div class="overview-label">{{ t('scanLog.statistics.totalLogs', '扫描记录') }}</div>
        </div>
      </div>
    </div>

    <!-- 模块分布 -->
    <div class="stats-section" v-if="moduleStats.length > 0">
      <div class="section-title">
        <i class="bx bx-pie-chart-alt-2"></i>
        <span>{{ t('scanLog.statistics.modules', '模块分布') }}</span>
      </div>

      <div class="module-list">
        <div
          v-for="stat in moduleStats"
          :key="stat.key"
          class="module-item"
        >
          <div class="module-header">
            <div class="module-info">
              <i :class="['bx', stat.icon]" class="module-icon" :style="{ color: stat.color }"></i>
              <span class="module-name">{{ stat.label }}</span>
            </div>
            <div class="module-stats">
              <span class="module-count">{{ stat.value }}</span>
              <span class="module-percentage">{{ stat.percentage }}%</span>
            </div>
          </div>
          <div class="progress-bar-container">
            <div
              class="progress-bar-fill"
              :style="{
                width: stat.percentage + '%',
                backgroundColor: stat.color
              }"
            ></div>
          </div>
        </div>
      </div>
    </div>

    <!-- 空状态 -->
    <div v-else class="empty-stats">
      <i class="bx bx-bar-chart"></i>
      <span>{{ t('scanLog.statistics.noData', '暂无统计数据') }}</span>
    </div>
  </div>
</template>

<style scoped>
.statistics-container {
  padding: var(--spacing-md);
  display: flex;
  flex-direction: column;
  gap: var(--spacing-lg);
}

.stats-section {
  display: flex;
  flex-direction: column;
  gap: var(--spacing-sm);
}

.section-title {
  display: flex;
  align-items: center;
  gap: var(--spacing-xs);
  font-size: var(--text-sm);
  font-weight: var(--font-weight-semibold);
  color: var(--color-text-secondary);
  text-transform: uppercase;
  letter-spacing: 0.05em;
}

.section-title i {
  font-size: 16px;
  color: var(--color-primary);
}

/* 总览卡片 */
.overview-card {
  display: flex;
  align-items: center;
  gap: var(--spacing-md);
  padding: var(--spacing-md);
  background: var(--glass-bg-card);
  backdrop-filter: var(--glass-blur-subtle);
  -webkit-backdrop-filter: var(--glass-blur-subtle);
  border: 1px solid var(--glass-border-light);
  border-radius: var(--radius-lg);
  transition: all var(--glass-transition-fast);
}

.overview-card:hover {
  background: var(--glass-bg-hover);
  border-color: var(--glass-border);
  box-shadow: var(--glass-shadow-subtle);
}

.overview-icon {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 48px;
  height: 48px;
  background: linear-gradient(135deg, var(--color-primary), var(--color-primary-dark));
  border-radius: var(--radius-md);
  color: white;
  font-size: 24px;
}

.overview-content {
  display: flex;
  flex-direction: column;
  gap: 2px;
}

.overview-number {
  font-size: var(--text-xl);
  font-weight: var(--font-weight-bold);
  color: var(--color-text-primary);
  line-height: 1.2;
}

.overview-label {
  font-size: var(--text-xs);
  color: var(--color-text-secondary);
}

/* 模块列表 */
.module-list {
  display: flex;
  flex-direction: column;
  gap: var(--spacing-sm);
}

.module-item {
  display: flex;
  flex-direction: column;
  gap: var(--spacing-xs);
  padding: var(--spacing-sm);
  background: var(--glass-bg-tertiary);
  border: 1px solid var(--glass-border-light);
  border-radius: var(--radius-sm);
  transition: all var(--glass-transition-fast);
  cursor: pointer;
}

.module-item:hover {
  background: var(--glass-bg-hover);
  border-color: var(--glass-border);
}

.module-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
}

.module-info {
  display: flex;
  align-items: center;
  gap: var(--spacing-xs);
}

.module-icon {
  font-size: 14px;
}

.module-name {
  font-size: var(--text-sm);
  font-weight: var(--font-weight-medium);
  color: var(--color-text-primary);
}

.module-stats {
  display: flex;
  align-items: center;
  gap: var(--spacing-xs);
}

.module-count {
  font-size: var(--text-sm);
  font-weight: var(--font-weight-semibold);
  color: var(--color-text-primary);
}

.module-percentage {
  font-size: var(--text-xs);
  color: var(--color-text-tertiary);
}

/* 进度条 */
.progress-bar-container {
  height: 4px;
  background: var(--glass-border-light);
  border-radius: var(--radius-full);
  overflow: hidden;
}

.progress-bar-fill {
  height: 100%;
  border-radius: var(--radius-full);
  transition: width var(--glass-transition-normal);
}

/* 空状态 */
.empty-stats {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: var(--spacing-sm);
  padding: var(--spacing-xl);
  color: var(--color-text-tertiary);
}

.empty-stats i {
  font-size: 32px;
  opacity: 0.5;
}

.empty-stats span {
  font-size: var(--text-sm);
}
</style> 