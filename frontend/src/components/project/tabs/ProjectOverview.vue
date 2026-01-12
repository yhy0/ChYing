<script setup lang="ts">
import { ref, computed } from 'vue';
import { useI18n } from 'vue-i18n';
import { useVulnerabilityStore } from '../../../store';

const { t } = useI18n();
const vulnerabilityStore = useVulnerabilityStore();

// 活动记录从后端获取，初始为空数组
interface ActivityItem {
  id: number;
  method: string;
  url: string;
  status: number;
  statusText: string;
  timestamp: string;
}

const recentActivities = ref<ActivityItem[]>([]);

// 从 store 获取漏洞统计
const issuesSummary = computed(() => ({
  high: vulnerabilityStore.highCount,
  medium: vulnerabilityStore.mediumCount,
  low: vulnerabilityStore.lowCount,
  info: vulnerabilityStore.infoCount
}));

// 获取状态码对应的颜色
const getStatusColor = (status: number) => {
  if (status >= 200 && status < 300) return { dot: 'bg-green-500', badge: 'bg-green-500/20 text-green-500' };
  if (status >= 300 && status < 400) return { dot: 'bg-yellow-500', badge: 'bg-yellow-500/20 text-yellow-500' };
  if (status >= 400 && status < 500) return { dot: 'bg-red-500', badge: 'bg-red-500/20 text-red-500' };
  if (status >= 500) return { dot: 'bg-purple-500', badge: 'bg-purple-500/20 text-purple-500' };
  return { dot: 'bg-gray-500', badge: 'bg-gray-500/20 text-gray-500' };
};
</script>

<template>
  <div class="grid grid-cols-3 gap-6">
    <!-- Recent Activity Card -->
    <div class="col-span-2 bg-[#f3f4f6] dark:bg-[#282838] rounded-md border border-gray-200 dark:border-gray-700 p-4">
      <h3 class="font-medium flex items-center">
        <i class="bx bx-history mr-2 text-gray-400"></i> {{ t('modules.project.activity.recent_activity') }}
      </h3>
      <div class="mt-4 space-y-4">
        <!-- 空状态 -->
        <div v-if="recentActivities.length === 0" class="text-center py-8 text-gray-400">
          <i class="bx bx-inbox text-4xl mb-2"></i>
          <p class="text-sm">{{ t('modules.project.activity.no_activity', 'No recent activity') }}</p>
        </div>
        <!-- 活动列表 -->
        <div v-for="activity in recentActivities" :key="activity.id" class="flex">
          <div class="w-8 flex-shrink-0">
            <div class="w-2 h-2 rounded-full mt-1.5 mx-auto" :class="getStatusColor(activity.status).dot"></div>
          </div>
          <div class="flex-1">
            <p class="text-sm">{{ activity.method }} request to <span class="text-[#4f46e5]">{{ activity.url }}</span></p>
            <p class="text-xs text-gray-400 mt-1">{{ activity.timestamp }}</p>
          </div>
          <div class="text-xs font-medium px-2 py-0.5 rounded h-fit" :class="getStatusColor(activity.status).badge">
            {{ activity.status }} {{ activity.statusText }}
          </div>
        </div>
      </div>
      <button v-if="recentActivities.length > 0" class="w-full mt-4 text-xs text-gray-400 hover:text-gray-900 dark:hover:text-white py-2 border-t border-gray-200 dark:border-gray-700">
        {{ t('modules.project.activity.view_all') }}
      </button>
    </div>
    
    <!-- Issues Summary Card -->
    <div class="bg-[#f3f4f6] dark:bg-[#282838] rounded-md border border-gray-200 dark:border-gray-700 p-4">
      <h3 class="font-medium flex items-center">
        <i class="bx bx-bug mr-2 text-gray-400"></i> {{ t('modules.project.issues.summary') }}
      </h3>
      <div class="mt-4 space-y-3">
        <div class="flex items-center justify-between">
          <div class="flex items-center">
            <div class="w-2 h-2 rounded-full bg-red-500 mr-2"></div>
            <span class="text-sm">{{ t('modules.project.issues.high_risk') }}</span>
          </div>
          <span class="text-sm font-medium">{{ issuesSummary.high }}</span>
        </div>
        <div class="flex items-center justify-between">
          <div class="flex items-center">
            <div class="w-2 h-2 rounded-full bg-yellow-500 mr-2"></div>
            <span class="text-sm">{{ t('modules.project.issues.medium_risk') }}</span>
          </div>
          <span class="text-sm font-medium">{{ issuesSummary.medium }}</span>
        </div>
        <div class="flex items-center justify-between">
          <div class="flex items-center">
            <div class="w-2 h-2 rounded-full bg-blue-500 mr-2"></div>
            <span class="text-sm">{{ t('modules.project.issues.low_risk') }}</span>
          </div>
          <span class="text-sm font-medium">{{ issuesSummary.low }}</span>
        </div>
        <div class="flex items-center justify-between">
          <div class="flex items-center">
            <div class="w-2 h-2 rounded-full bg-gray-500 mr-2"></div>
            <span class="text-sm">{{ t('modules.project.issues.information') }}</span>
          </div>
          <span class="text-sm font-medium">{{ issuesSummary.info }}</span>
        </div>
      </div>
    </div>
  </div>
</template> 