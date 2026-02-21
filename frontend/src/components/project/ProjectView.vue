<script setup lang="ts">
import { ref, provide, computed, watch, onMounted, onBeforeUnmount } from 'vue';
import { useI18n } from 'vue-i18n';
import { useVulnerabilityStore } from '../../store';
import { Events } from "@wailsio/runtime";
// @ts-ignore
import { GetTotalRequests } from "../../../bindings/github.com/yhy0/ChYing/app.js";
import ProjectHeader from './ProjectHeader.vue';
import ProjectSiteMapIntegrated from './tabs/ProjectSiteMapIntegrated.vue';
import ProjectVulnerabilitiesTab from './tabs/ProjectVulnerabilitiesTab.vue';
import type { Project, ProjectDetails, SecurityIssue, HttpHistoryItem, SiteMapNode } from '../../types';

const { t } = useI18n();
const vulnerabilityStore = useVulnerabilityStore();

const activeTab = ref('siteMap');

const setActiveTab = (tabName: string) => {
  activeTab.value = tabName;
};

// 站点地图中的主机数据
const uniqueHosts = ref<string[]>([]);

// 更新主机数据的函数，由SiteMap组件调用
const updateSiteMapHosts = (hosts: string[]) => {
  if (hosts && hosts.length > 0) {
    // 直接设置为新的主机列表
    uniqueHosts.value = hosts;
  }
};

// 提供给子组件
provide('updateSiteMapHosts', updateSiteMapHosts);

// 计算漏洞数量和主机数量（从 store 获取）
const vulnerabilitiesCount = computed(() => vulnerabilityStore.totalCount);
const hostsCount = computed(() => uniqueHosts.value.length);

// 项目详情数据对象
const selectedProject = ref<ProjectDetails>({
  name: '',
  createdDate: '',
  totalRequests: 0,
  issuesFound: 0,
  hosts: 0,
  scanProgress: 0
});

// 使用 watch 监听计算属性的变化，更新 selectedProject
watch([vulnerabilitiesCount, hostsCount], ([vulCount, hostCount]) => {
  selectedProject.value.issuesFound = vulCount;
  selectedProject.value.hosts = hostCount;
}, { immediate: true });

// 获取请求总数并监听实时更新
onMounted(() => {
  // 从后端获取初始请求总数
  GetTotalRequests()
    .then((count: number) => {
      selectedProject.value.totalRequests = count || 0;
    })
    .catch((err: any) => {
      console.error('获取请求总数失败:', err);
    });

  // 监听 HttpHistory 事件，每有新请求就递增计数
  Events.On("HttpHistory", () => {
    selectedProject.value.totalRequests++;
  });
});

onBeforeUnmount(() => {
  Events.Off("HttpHistory");
});

// 安全问题数据从后端获取，初始为空数组
const securityIssues = ref<SecurityIssue[]>([]);

const getSeverityColorClass = (severity: string) => {
  switch (severity) {
    case 'high':
      return 'bg-red-500/20 text-red-500';
    case 'medium':
      return 'bg-yellow-500/20 text-yellow-500';
    case 'low':
      return 'bg-blue-500/20 text-blue-500';
    case 'info':
      return 'bg-gray-500/20 text-gray-500';
    default:
      return 'bg-gray-500/20 text-gray-500';
  }
};

const getSeverityText = (severity: string) => {
  switch (severity) {
    case 'high':
      return t('modules.project.issues.high_risk');
    case 'medium':
      return t('modules.project.issues.medium_risk');
    case 'low':
      return t('modules.project.issues.low_risk');
    case 'info':
      return t('modules.project.issues.information');
    default:
      return severity;
  }
};

const getStatusColor = (status: number) => {
  if (status >= 200 && status < 300) {
    return 'bg-green-500/20 text-green-500';
  } else if (status >= 300 && status < 400) {
    return 'bg-yellow-500/20 text-yellow-500';
  } else if (status >= 400 && status < 500) {
    return 'bg-red-500/20 text-red-500';
  } else if (status >= 500) {
    return 'bg-purple-500/20 text-purple-500';
  } else {
    return 'bg-gray-500/20 text-gray-500';
  }
};

const getMethodColor = (method: string) => {
  switch (method.toUpperCase()) {
    case 'GET':
      return 'bg-blue-500/20 text-blue-500';
    case 'POST':
      return 'bg-green-500/20 text-green-500';
    case 'PUT':
      return 'bg-yellow-500/20 text-yellow-500';
    case 'DELETE':
      return 'bg-red-500/20 text-red-500';
    default:
      return 'bg-gray-500/20 text-gray-500';
  }
};

const formatBytes = (bytes: number, decimals = 2) => {
  if (bytes === 0) return '0 B';
  
  const k = 1024;
  const dm = decimals < 0 ? 0 : decimals;
  const sizes = ['B', 'KB', 'MB', 'GB', 'TB', 'PB', 'EB', 'ZB', 'YB'];
  
  const i = Math.floor(Math.log(bytes) / Math.log(k));
  
  return parseFloat((bytes / Math.pow(k, i)).toFixed(dm)) + ' ' + sizes[i];
};

const showExportModal = ref(false);

provide('securityIssues', securityIssues);
provide('getSeverityColorClass', getSeverityColorClass);
provide('getSeverityText', getSeverityText);
provide('getStatusColor', getStatusColor);
provide('getMethodColor', getMethodColor);
provide('formatBytes', formatBytes);
provide('Project', {} as Project);
provide('ProjectDetails', {} as ProjectDetails);
provide('SecurityIssue', {} as SecurityIssue);
provide('HttpHistoryItem', {} as HttpHistoryItem);
provide('SiteMapNode', {} as SiteMapNode);
</script>

<template>
  <!-- 项目视图容器：占满整个视图高度 -->
  <div class="h-screen flex flex-col bg-white dark:bg-gray-900">
    <!-- 项目头部：固定高度 -->
    <div class="flex-none">
      <ProjectHeader
        :selected-project="selectedProject"
        @open-export-modal="showExportModal = true"
      />
    </div>

    <!-- 主要内容区域，包含Tab导航和Tab内容 -->
    <div class="flex-1 flex flex-col overflow-hidden">
      <!-- Tab 导航 (参考 ProxyView.vue 样式) -->
      <div class="flex border-b border-gray-200 dark:border-gray-700">
        <button
          @click="setActiveTab('siteMap')"
          class="px-4 py-2 text-sm font-medium flex items-center"
          :class="[
            activeTab === 'siteMap'
              ? 'border-b-2 border-indigo-500 text-indigo-600 dark:text-indigo-400'
              : 'text-gray-500 dark:text-gray-400 hover:text-gray-700 dark:hover:text-gray-300'
          ]"
        >
          <i class="bx bx-sitemap mr-1.5"></i>
          {{ t('project.tabs.siteMap', 'Site Map') }}
        </button>
        <button
          @click="setActiveTab('vulnerabilities')"
          class="px-4 py-2 text-sm font-medium flex items-center"
          :class="[
            activeTab === 'vulnerabilities'
              ? 'border-b-2 border-indigo-500 text-indigo-600 dark:text-indigo-400'
              : 'text-gray-500 dark:text-gray-400 hover:text-gray-700 dark:hover:text-gray-300'
          ]"
        >
          <i class="bx bx-shield-quarter mr-1.5"></i>
          {{ t('project.tabs.vulnerabilities', 'Vulnerabilities') }}
        </button>
      </div>

      <!-- Tab 内容区域 -->
      <div class="flex-1 min-h-0 overflow-y-auto">
        <ProjectSiteMapIntegrated v-if="activeTab === 'siteMap'" />
        <ProjectVulnerabilitiesTab v-if="activeTab === 'vulnerabilities'" />
      </div>
    </div>
  </div>
</template>

<style scoped>
/* 移除之前的 .project-tab, .project-tab-active, .project-tab-inactive 样式 */
/* 如果需要特定的额外样式可以在这里添加，但大部分样式已通过Tailwind类直接应用 */

/* 标签切换过渡动画 */
.fade-enter-active, .fade-leave-active {
  transition: opacity 0.3s ease;
}
.fade-enter-from, .fade-leave-to {
  opacity: 0;
}
</style> 