<script setup lang="ts">
import { ref, computed, inject, onMounted } from 'vue';
import { useI18n } from 'vue-i18n';
import { useVulnerabilityStore } from '../../../store/vulnerability';
import type { VulnerabilityItem } from '../../../types/vulnerability';

const { t } = useI18n();
const vulnerabilityStore = useVulnerabilityStore();

// 注入获取严重性颜色的函数
const getSeverityColorClass = inject('getSeverityColorClass', (severity: string) => '') as (severity: string) => string;
const getSeverityText = inject('getSeverityText', (severity: string) => '') as (severity: string) => string;

// 直接从 store 获取数据，不再维护本地状态
const vulnerabilities = computed(() => vulnerabilityStore.vulnerabilities);

// 加载状态 - 基于 store 中是否有数据来判断
const isLoading = ref(true);

// 选中的漏洞
const selectedVulnerability = ref<VulnerabilityItem | null>(null);

// 漏洞级别统计 - 直接从 store 获取
const vulnerabilityStats = computed(() => {
  const stats = vulnerabilityStore.levelStatistics;
  return {
    high: stats.high + stats.critical, // 将 critical 和 high 合并显示
    medium: stats.medium,
    low: stats.low,
    info: 0 // store 中没有 info 级别，如果需要可以扩展
  };
});

onMounted(() => {
  // 设置超时，如果 store 中有数据或者等待一段时间后停止加载状态
  setTimeout(() => {
    isLoading.value = false;
  }, 500);
  
  // 如果 store 中已经有数据，立即停止加载状态
  if (vulnerabilities.value.length > 0) {
    isLoading.value = false;
  }
});

// 选择漏洞详情
const selectVulnerability = (vuln: VulnerabilityItem) => {
  selectedVulnerability.value = vuln;
};

// 根据级别获取样式类
const getLevelClass = (level: string) => {
  return getSeverityColorClass(level);
};

// 获取漏洞级别翻译
const getLevelText = (level: string) => {
  return getSeverityText(level);
};

// 漏洞类型的图标
const getVulnerabilityTypeIcon = (type: string) => {
  switch (type.toLowerCase()) {
    case 'sql':
    case 'sqli':
      return 'bx-data';
    case 'xss':
      return 'bx-code-alt';
    case 'rce':
      return 'bx-terminal';
    case 'ssrf':
      return 'bx-link-external';
    case 'upload':
      return 'bx-upload';
    case 'lfi':
    case 'rfi':
      return 'bx-file';
    case 'xxe':
      return 'bx-file-xml';
    default:
      return 'bx-bug-alt';
  }
};
</script>

<template>
  <div class="h-full flex flex-col bg-white dark:bg-gray-900 overflow-hidden">
    <!-- 标题和统计信息 -->
    <div class="px-4 pt-4 pb-3 border-b border-gray-200 dark:border-gray-700">
      <h2 class="text-lg font-semibold text-gray-800 dark:text-gray-200 mb-2">
        {{ t('project.vulnerabilities.title', '漏洞信息') }}
      </h2>
      
      <!-- 漏洞统计 -->
      <div class="flex flex-wrap gap-2">
        <div class="py-1 px-2 rounded-md bg-red-500/10 text-red-600 dark:text-red-400 text-xs font-medium" v-if="vulnerabilityStats.high > 0">
          {{ t('project.vulnerabilities.high', '高危') }}: {{ vulnerabilityStats.high }}
        </div>
        <div class="py-1 px-2 rounded-md bg-yellow-500/10 text-yellow-600 dark:text-yellow-400 text-xs font-medium" v-if="vulnerabilityStats.medium > 0">
          {{ t('project.vulnerabilities.medium', '中危') }}: {{ vulnerabilityStats.medium }}
        </div>
        <div class="py-1 px-2 rounded-md bg-blue-500/10 text-blue-600 dark:text-blue-400 text-xs font-medium" v-if="vulnerabilityStats.low > 0">
          {{ t('project.vulnerabilities.low', '低危') }}: {{ vulnerabilityStats.low }}
        </div>
        <div class="py-1 px-2 rounded-md bg-gray-500/10 text-gray-600 dark:text-gray-400 text-xs font-medium" v-if="vulnerabilityStats.info > 0">
          {{ t('project.vulnerabilities.info', '信息') }}: {{ vulnerabilityStats.info }}
        </div>
      </div>
    </div>
    
    <!-- 加载状态 -->
    <div v-if="isLoading" class="flex-1 flex items-center justify-center p-4">
      <div class="text-center">
        <div class="inline-block animate-spin rounded-full h-8 w-8 border-t-2 border-b-2 border-indigo-500 mb-2"></div>
        <p class="text-gray-500 dark:text-gray-400">{{ t('project.vulnerabilities.loading', '加载漏洞数据...') }}</p>
      </div>
    </div>
    
    <!-- 无数据状态 -->
    <div v-else-if="vulnerabilities.length === 0" class="flex-1 flex items-center justify-center p-4">
      <div class="text-center">
        <i class="bx bx-shield-quarter text-5xl text-gray-300 dark:text-gray-600 mb-2"></i>
        <p class="text-gray-500 dark:text-gray-400">{{ t('project.vulnerabilities.noData', '暂无漏洞数据') }}</p>
      </div>
    </div>
    
    <!-- 漏洞列表和详情 -->
    <div v-else class="flex-1 flex overflow-hidden">
      <!-- 漏洞列表 -->
      <div class="w-1/3 border-r border-gray-200 dark:border-gray-700 overflow-y-auto">
        <div class="p-2">
          <div 
            v-for="vuln in vulnerabilities" 
            :key="vuln.id"
            class="p-3 mb-2 border rounded-md cursor-pointer transition-colors duration-150 hover:bg-gray-50 dark:hover:bg-gray-800"
            :class="[
              'border-gray-200 dark:border-gray-700',
              selectedVulnerability === vuln ? 'bg-indigo-50 dark:bg-indigo-900/20' : ''
            ]"
            @click="selectVulnerability(vuln)"
          >
            <div class="flex items-start justify-between">
              <div class="flex items-start space-x-2">
                <i :class="['bx text-lg mt-0.5', getVulnerabilityTypeIcon(vuln.vulnType)]"></i>
                <div>
                  <div class="font-medium text-gray-800 dark:text-gray-200">{{ vuln.plugin }}</div>
                  <div class="text-xs text-gray-500 dark:text-gray-400">{{ vuln.vulnType }}</div>
                </div>
              </div>
              <span 
                :class="['px-1.5 py-0.5 text-xs font-medium rounded-full', getLevelClass(vuln.level)]"
              >
                {{ getLevelText(vuln.level) }}
              </span>
            </div>
          </div>
        </div>
      </div>
      
      <!-- 漏洞详情 -->
      <div class="w-2/3 overflow-y-auto p-4">
        <div v-if="selectedVulnerability" class="space-y-4">
          <div class="flex justify-between items-start">
            <h3 class="text-xl font-semibold text-gray-900 dark:text-gray-100">
              {{ selectedVulnerability.plugin }}
            </h3>
            <span 
              :class="['px-2 py-1 text-sm font-semibold rounded-full', getLevelClass(selectedVulnerability.level)]"
            >
              {{ getLevelText(selectedVulnerability.level) }}
            </span>
          </div>
          
          <div>
            <div class="text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">
              {{ t('project.vulnerabilities.typeLabel', '漏洞类型') }}
            </div>
            <div class="text-gray-800 dark:text-gray-200">
              {{ selectedVulnerability.vulnType }}
            </div>
          </div>
          
          <div v-if="selectedVulnerability.target">
            <div class="text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">
              {{ t('project.vulnerabilities.targetLabel', '目标地址') }}
            </div>
            <div class="text-gray-800 dark:text-gray-200 font-mono text-sm">
              {{ selectedVulnerability.target }}
            </div>
          </div>
          
          <div v-if="selectedVulnerability.description">
            <div class="text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">
              {{ t('project.vulnerabilities.descriptionLabel', '漏洞描述') }}
            </div>
            <div class="text-gray-800 dark:text-gray-200">
              {{ selectedVulnerability.description }}
            </div>
          </div>
          
          <div v-if="selectedVulnerability.payload">
            <div class="text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">
              {{ t('project.vulnerabilities.payloadLabel', 'Payload') }}
            </div>
            <pre class="whitespace-pre-wrap bg-gray-50 dark:bg-gray-800 p-3 rounded-md text-sm border border-gray-200 dark:border-gray-700 font-mono">{{ selectedVulnerability.payload }}</pre>
          </div>
          
          <div v-if="selectedVulnerability.request">
            <div class="text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">
              {{ t('project.vulnerabilities.requestLabel', '请求内容') }}
            </div>
            <pre class="whitespace-pre-wrap bg-gray-50 dark:bg-gray-800 p-3 rounded-md text-sm overflow-auto max-h-[30vh] border border-gray-200 dark:border-gray-700 font-mono">{{ selectedVulnerability.request }}</pre>
          </div>
          
          <div v-if="selectedVulnerability.response">
            <div class="text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">
              {{ t('project.vulnerabilities.responseLabel', '响应内容') }}
            </div>
            <pre class="whitespace-pre-wrap bg-gray-50 dark:bg-gray-800 p-3 rounded-md text-sm overflow-auto max-h-[30vh] border border-gray-200 dark:border-gray-700 font-mono">{{ selectedVulnerability.response }}</pre>
          </div>
          
          <div v-if="selectedVulnerability.curlCommand">
            <div class="text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">
              {{ t('project.vulnerabilities.curlLabel', 'cURL 命令') }}
            </div>
            <pre class="whitespace-pre-wrap bg-gray-50 dark:bg-gray-800 p-3 rounded-md text-sm border border-gray-200 dark:border-gray-700 font-mono">{{ selectedVulnerability.curlCommand }}</pre>
          </div>
        </div>
        
        <div v-else class="h-full flex items-center justify-center">
          <p class="text-gray-500 dark:text-gray-400">{{ t('project.vulnerabilities.selectPrompt', '请选择左侧漏洞查看详情') }}</p>
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
/* 可以添加特定于该组件的样式 */
</style>