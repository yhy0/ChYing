<script setup lang="ts">
import { ref, computed, watch } from 'vue';
import { useI18n } from 'vue-i18n';
import type { SecurityIssue, SiteMapNode } from '../../../types/project';
import Drawer from '../../common/Drawer.vue';
import Info from '../../tools/Info.vue';

const { t } = useI18n();

const currentHost = ref('');
// 控制抽屉显示状态
const showInfoDrawer = ref(false);
// 接收props
const props = defineProps({
  securityIssues: {
    type: Array as () => SecurityIssue[],
    required: true
  },
  selectedNode: {
    type: Object as () => SiteMapNode | null,
    default: null
  },
  getSeverityColorClass: {
    type: Function as unknown as () => (severity: string) => string,
    required: true
  },
  getSeverityText: {
    type: Function as unknown as () => (severity: string) => string,
    required: true
  }
});

// 本地状态
const selectedIssue = ref<SecurityIssue | null>(null);
const issuesTableHeight = ref('45%');

// 筛选安全问题
const filteredIssues = computed(() => {
  if (!props.selectedNode) return props.securityIssues;

  // 如果选中的是主机节点，筛选该主机的所有安全问题
  if (props.selectedNode.nodeType === 'host') {
    currentHost.value = props.selectedNode?.name;
    return props.securityIssues.filter(item =>
      item.host === props.selectedNode?.name
    );
  }

  // 如果选中的是目录或文件，筛选该路径下的安全问题
  return props.securityIssues.filter(item =>
    item.host === props.selectedNode?.fullUrl.split('/')[2] &&
    item.path.startsWith(props.selectedNode.path)
  );
});

// 选择安全问题
const selectIssue = (issue: SecurityIssue) => {
  selectedIssue.value = issue;
  // 发射选择事件
  emit('issue-selected', issue);
};

// 监听选中的安全问题变化，调整表格高度
watch(() => selectedIssue.value, (newValue) => {
  if (newValue) {
    issuesTableHeight.value = '35%';
  } else {
    issuesTableHeight.value = '45%';
  }
}, { immediate: true });

// 清除选择的安全问题
const clearSelectedIssue = () => {
  selectedIssue.value = null;
};

// 监听节点变化，清除已选择的安全问题
watch(() => props.selectedNode, (newNode) => {
  clearSelectedIssue();

  // 更新当前主机值
  if (newNode && newNode.nodeType === 'host') {
    currentHost.value = newNode.name;
    console.log(currentHost.value);
  } else if (newNode) {
    // 如果是其他类型节点，尝试从URL中提取主机名
    const hostFromUrl = newNode.fullUrl?.split('/')[2];
    if (hostFromUrl) {
      currentHost.value = hostFromUrl;
    }
  } else {
    // 如果没有选中节点，清空currentHost
    currentHost.value = '';
  }
});

// 定义事件
const emit = defineEmits(['issue-selected']);

// 导出方法给父组件
defineExpose({
  clearSelectedIssue
});
</script>

<template>
  <div class="overflow-hidden flex flex-col h-full">
    <div
      class="p-3 bg-white dark:bg-[#1e1e36] border-b border-gray-200 dark:border-gray-700 flex items-center justify-between">
      <h3 class="text-sm font-medium">
        {{ t('modules.project.issues.title') }}
        <span v-if="selectedNode" class="text-xs text-gray-500 ml-2">
          ({{ filteredIssues.length }} {{ t('modules.project.issues.found') }})
        </span>
      </h3>
      <button 
        v-if="currentHost"
        @click="showInfoDrawer = true"
        class="flex items-center justify-center px-2 py-1 text-xs bg-indigo-600 hover:bg-indigo-700 text-white rounded-md transition-colors duration-200 shadow-sm"
      >
        <i class="bx bx-info-circle mr-1"></i>
        <span>{{ t('modules.tools.info_collection') }}</span>
      </button>
      
      <!-- 使用Drawer组件 -->
      <Drawer 
        v-if="currentHost"
        v-model:show="showInfoDrawer" 
        :title="t('modules.tools.info_collection_results') + ' (' + currentHost + ')'"
        :default-width="860"
        placement="right"
      >
        <Info :host="currentHost" />
      </Drawer>
    </div>

    <!-- Security Issues Table with dynamic height -->
    <div
      class="bg-white dark:bg-[#1e1e36] rounded-lg border border-gray-200 dark:border-gray-700 overflow-hidden transition-all duration-300 ease-in-out shadow-sm"
      :style="{ height: issuesTableHeight }">
      <div class="h-full overflow-auto scrollbar-thin scrollbar-thumb-gray-300 dark:scrollbar-thumb-gray-600">
        <table class="min-w-full divide-y divide-gray-200 dark:divide-gray-700">
          <thead class="bg-gray-50 dark:bg-[#292945] sticky top-0 z-10">
            <tr>
              <th scope="col"
                class="px-3 py-2 text-left text-xs font-medium text-gray-500 dark:text-gray-400 uppercase tracking-wider">
                {{ t('modules.project.issues.name') }}
              </th>
              <th scope="col"
                class="px-3 py-2 text-left text-xs font-medium text-gray-500 dark:text-gray-400 uppercase tracking-wider w-24">
                {{ t('modules.project.issues.severity') }}
              </th>
            </tr>
          </thead>
          <tbody class="bg-white dark:bg-[#1e1e36] divide-y divide-gray-200 dark:divide-gray-700">
            <tr v-for="issue in filteredIssues" :key="issue.id" @click="selectIssue(issue)"
              class="hover:bg-gray-50 dark:hover:bg-[#252545] cursor-pointer transition-colors duration-150"
              :class="{ 'bg-indigo-50/50 dark:bg-indigo-900/10': selectedIssue?.id === issue.id }">
              <td class="px-3 py-2">
                <div class="text-xs font-medium text-gray-800 dark:text-gray-200">{{ issue.name }}</div>
                <div class="text-xs text-gray-500 dark:text-gray-400">{{ issue.path }}</div>
              </td>
              <td class="px-3 py-2 whitespace-nowrap">
                <span
                  :class="['px-1.5 py-0.5 text-2xs rounded-full font-medium', getSeverityColorClass(issue.severity)]">
                  {{ getSeverityText(issue.severity) }}
                </span>
              </td>
            </tr>
          </tbody>
        </table>
      </div>
    </div>

    <!-- Security Issue Details (if any selected) -->
    <div v-if="selectedIssue"
      class="mt-3 bg-white dark:bg-[#1e1e36] rounded-lg border border-gray-200 dark:border-gray-700 p-4 flex-1 overflow-auto transition-all duration-300 ease-in-out transform animate-slideUp shadow-sm">
      <div class="flex justify-between mb-4">
        <h3 class="text-sm font-semibold text-gray-800 dark:text-gray-100">{{ selectedIssue.name }}</h3>
        <span
          :class="['px-1.5 py-0.5 text-2xs rounded-full font-medium', getSeverityColorClass(selectedIssue.severity)]">
          {{ getSeverityText(selectedIssue.severity) }}
        </span>
      </div>

      <div class="grid grid-cols-2 gap-x-6 gap-y-3 mb-4 bg-gray-50 dark:bg-[#252545] p-3 rounded-md">
        <div>
          <label class="block text-2xs text-gray-500 dark:text-gray-400 uppercase mb-0.5 font-medium">{{
            t('modules.project.issues.host') }}</label>
          <div class="text-xs text-gray-800 dark:text-gray-200">
            {{ selectedIssue.host }}{{ selectedIssue.path }}
          </div>
        </div>
        <div>
          <label class="block text-2xs text-gray-500 dark:text-gray-400 uppercase mb-0.5 font-medium">{{
            t('modules.project.issues.discovered') }}</label>
          <div class="text-xs text-gray-800 dark:text-gray-200">
            {{ selectedIssue.timestamp }}
          </div>
        </div>
      </div>

      <div class="mb-4 border-l-2 border-indigo-400 dark:border-indigo-700 pl-3">
        <label class="block text-2xs text-gray-500 dark:text-gray-400 uppercase mb-0.5 font-medium">{{
          t('modules.project.issues.description') }}</label>
        <div class="text-xs text-gray-800 dark:text-gray-200">
          {{ selectedIssue.description }}
        </div>
      </div>
    </div>

    <!-- No issue selected message -->
    <div v-else-if="!selectedNode || selectedNode.nodeType !== 'host'"
      class="mt-3 bg-white dark:bg-[#1e1e36] rounded-lg border border-gray-200 dark:border-gray-700 p-4 text-center flex-none transition-all duration-300 ease-in-out shadow-sm">
      <div class="flex flex-col items-center py-4">
        <i class="bx bx-flag text-xl text-gray-400 dark:text-gray-600 mb-2"></i>
        <p class="text-xs text-gray-500 dark:text-gray-400">{{ t('modules.project.issues.select_issue') }}</p>
      </div>
    </div>
  </div>
</template>

<style scoped>
.text-2xs {
  font-size: 0.65rem;
}

.animate-slideUp {
  animation: slideUp 0.3s ease-out;
}

@keyframes slideUp {
  from {
    opacity: 0;
    transform: translateY(10px);
  }

  to {
    opacity: 1;
    transform: translateY(0);
  }
}

/* 滚动条样式已移至 scrollbar.css 统一管理 */
</style>