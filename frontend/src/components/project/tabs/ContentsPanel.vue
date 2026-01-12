<script setup lang="ts">
import { ref, computed, watch, onMounted, onBeforeUnmount } from 'vue';
import { useI18n } from 'vue-i18n';
import type { SiteMapNode } from '../../../types/project';
import RequestResponsePanel from '../../common/RequestResponsePanel.vue';
import { useProjectStore } from '../../../store/project';

const { t } = useI18n();
const projectStore = useProjectStore();

// 接收props
const props = defineProps({
  selectedNode: {
    type: Object as () => SiteMapNode | null,
    default: null
  },
  formatBytes: {
    type: Function as unknown as () => (bytes: number, decimals?: number) => string,
    required: true
  },
  getMethodColor: {
    type: Function as unknown as () => (method: string) => string,
    required: true
  },
  getStatusColor: {
    type: Function as unknown as () => (status: number) => string,
    required: true
  },
  // 添加请求和响应数据props
  requestData: {
    type: String,
    default: ''
  },
  responseData: {
    type: String,
    default: ''
  }
});

// 本地状态
const request = ref('');
const response = ref('');
const serverDurationMs = ref(0);
const nodeInfoHeight = ref(300); // 节点信息面板的默认高度
const containerRef = ref<HTMLElement | null>(null);

// 计算当前选中节点对应的主机节点
const hostNode = computed(() => {
  if (!props.selectedNode) return null;

  // 如果当前节点就是主机节点，直接返回
  if (props.selectedNode.nodeType === 'host') {
    return props.selectedNode;
  }

  // 在站点地图数据中查找当前节点所属的主机
  const findHostForNode = (node: SiteMapNode): SiteMapNode | null => {
    // 从节点的fullUrl提取主机名
    let url = node.fullUrl || '';
    try {
      const urlObj = new URL(url);
      const hostname = urlObj.hostname;

      // 在siteMapData中查找匹配的主机节点
      const siteMapData = projectStore.siteMapData;
      for (const hostNode of siteMapData) {
        if (hostNode.nodeType === 'host' &&
          (hostNode.name === hostname || hostNode.name.startsWith(hostname + ' '))) {
          return hostNode;
        }
      }
    } catch (e) {
      console.error('解析URL失败:', url, e);
    }
    return null;
  };

  return findHostForNode(props.selectedNode);
});

// 监听请求和响应数据变化
watch(() => [props.requestData, props.responseData], ([newRequest, newResponse]) => {
  console.log('接收到请求数据:', newRequest?.substring(0, 50));
  console.log('接收到响应数据:', newResponse?.substring(0, 50));

  request.value = newRequest || '';
  response.value = newResponse || '';
}, { immediate: true });

// 监听节点变化
watch(() => props.selectedNode, (node) => {
  console.log('节点变化:', node?.name, node?.path);
  if (!node) {
    request.value = '';
    response.value = '';
  }
}, { immediate: true });

// 计算是否有数据显示
const hasData = computed(() => {
  return !!(request.value || response.value);
});

// 面板调整变量
let isDragging = false;
let initialY = 0;
let initialHeight = 0;

// 开始调整面板高度
const startResize = (e: MouseEvent) => {
  e.preventDefault();
  isDragging = true;
  initialY = e.clientY;

  if (containerRef.value) {
    initialHeight = nodeInfoHeight.value;
  }

  document.addEventListener('mousemove', handleMouseMove);
  document.addEventListener('mouseup', stopResize);
  document.body.classList.add('dragging');
  document.body.classList.add('cursor-ns-resize');
};

// 处理鼠标移动
const handleMouseMove = (e: MouseEvent) => {
  if (!isDragging) return;

  const deltaY = e.clientY - initialY;
  const newHeight = Math.max(100, Math.min(500, initialHeight + deltaY));
  nodeInfoHeight.value = newHeight;
};

// 停止拖拽
const stopResize = () => {
  isDragging = false;
  document.removeEventListener('mousemove', handleMouseMove);
  document.removeEventListener('mouseup', stopResize);
  document.body.classList.remove('dragging');
  document.body.classList.remove('cursor-ns-resize');
};

// 组件挂载时清理可能存在的事件监听器
onMounted(() => {
  document.removeEventListener('mousemove', handleMouseMove);
  document.removeEventListener('mouseup', stopResize);
  document.body.classList.remove('dragging');
  document.body.classList.remove('cursor-ns-resize');
});

// 组件卸载时清理事件监听器
onBeforeUnmount(() => {
  document.removeEventListener('mousemove', handleMouseMove);
  document.removeEventListener('mouseup', stopResize);
  document.body.classList.remove('dragging');
  document.body.classList.remove('cursor-ns-resize');
});
</script>

<template>
  <div class="h-full flex flex-col bg-white dark:bg-gray-900" ref="containerRef">
    <!-- 标题栏 -->
    <div class="flex-none p-3 border-b border-gray-200 dark:border-gray-800">
      <h3 class="text-sm font-medium text-gray-900 dark:text-white">
        {{ t('modules.project.contents.title') }}
      </h3>
    </div>

    <!-- 节点信息面板 -->
    <div v-if="selectedNode"
      class="flex-none overflow-hidden bg-white dark:bg-gray-900 border-b border-gray-200 dark:border-gray-800"
      :style="{ height: `${nodeInfoHeight}px` }">
      <div class="h-full overflow-y-auto scrollbar-thin">
        <div class="p-4">
          <!-- 节点基本信息 -->
          <div class="grid grid-cols-2 gap-4 mb-3 bg-gray-50 dark:bg-gray-800/50 p-3 rounded-lg">
            <div v-if="selectedNode.nodeType">
              <label class="block text-2xs text-gray-500 dark:text-gray-400 uppercase mb-0.5 font-medium">类型</label>
              <div class="text-xs text-gray-800 dark:text-gray-200">
                {{ selectedNode.nodeType === 'host' ? '主机' : selectedNode.nodeType === 'directory' ? '目录' : '文件' }}
                <span v-if="hostNode?.rawData?.cdn"
                  class="px-1.5 py-0.5 text-2xs rounded-full font-medium bg-blue-100 dark:bg-blue-900/20 text-blue-600 dark:text-blue-400">
                  CDN
                </span>
              </div>

            </div>

            <!-- 显示主机信息 - 始终使用hostNode -->
            <div v-if="hostNode?.rawData?.api_cnt !== undefined">
              <label class="block text-2xs text-gray-500 dark:text-gray-400 uppercase mb-0.5 font-medium">API数量</label>
              <div class="text-xs text-gray-800 dark:text-gray-200">
                {{ hostNode.rawData.api_cnt || 0 }}
              </div>
            </div>
            <div v-if="hostNode?.rawData?.inner_ip_cnt !== undefined">
              <label class="block text-2xs text-gray-500 dark:text-gray-400 uppercase mb-0.5 font-medium">内部IP数量</label>
              <div class="text-xs text-gray-800 dark:text-gray-200">
                {{ hostNode.rawData.inner_ip_cnt || 0 }}
              </div>
            </div>
            <div v-if="hostNode?.rawData?.subdomain_cnt !== undefined">
              <label class="block text-2xs text-gray-500 dark:text-gray-400 uppercase mb-0.5 font-medium">子域名数量</label>
              <div class="text-xs text-gray-800 dark:text-gray-200">
                {{ hostNode.rawData.subdomain_cnt || 0 }}
              </div>
            </div>
            <div v-if="hostNode?.rawData?.params_cnt !== undefined">
              <label class="block text-2xs text-gray-500 dark:text-gray-400 uppercase mb-0.5 font-medium">参数数量</label>
              <div class="text-xs text-gray-800 dark:text-gray-200">
                {{ hostNode.rawData.params_cnt || 0 }}
              </div>
            </div>
            <!-- IP信息 - 使用hostNode -->
            <div v-if="hostNode?.rawData?.IPMsg" class="mb-3 border-l-2 border-blue-500 pl-3">
              <label class="block text-xs text-gray-500 dark:text-gray-400 font-medium mb-1">IP信息</label>
              <div class="text-sm text-gray-700 dark:text-gray-300">
                {{ hostNode.rawData.IPMsg }}
              </div>
            </div>
          </div>

          <!-- 指纹信息 - 使用hostNode -->
          <div v-if="hostNode?.rawData?.fingerprint?.length" class="mb-3">
            <label class="block text-xs text-gray-500 dark:text-gray-400 font-medium mb-1">指纹信息</label>
            <div class="flex flex-wrap gap-1">
              <span v-for="(fp, index) in hostNode.rawData.fingerprint" :key="index"
                class="px-1.5 py-0.5 text-2xs rounded-full font-medium bg-green-100 dark:bg-green-900/20 text-green-600 dark:text-green-400">
                {{ fp }}
              </span>
            </div>
          </div>
        </div>
      </div>
    </div>

    <!-- 高度调整器 - 使用panel-divider类的水平版本 -->
    <div class="panel-divider-horizontal" @mousedown="startResize"></div>

    <!-- 请求和响应面板容器 -->
    <div class="flex-1 min-h-0">
      <div v-if="hasData" class="h-full">
        <RequestResponsePanel :request-data="request" :response-data="response" :request-read-only="true"
          :response-read-only="true" :server-duration-ms="serverDurationMs"
          :uuid="selectedNode?.id.toString() || 'default'" :hide-empty-response="false" />
      </div>

      <!-- 没有请求/响应数据时显示提示信息 -->
      <div v-else
        class="h-full flex items-center justify-center bg-white dark:bg-gray-900 rounded-lg border border-gray-200 dark:border-gray-800">
        <div class="text-center">
          <i class="bx bx-network-chart text-xl text-gray-400 dark:text-gray-600 mb-2"></i>
          <p class="text-sm text-gray-500 dark:text-gray-400">
            {{ t('modules.project.contents.select_request') }}
          </p>
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
.text-2xs {
  font-size: 0.65rem;
}

.scrollbar-thin::-webkit-scrollbar {
  width: 4px;
}

.scrollbar-thin::-webkit-scrollbar-track {
  background-color: transparent;
}

.scrollbar-thin::-webkit-scrollbar-thumb {
  background-color: rgba(156, 163, 175, 0.5);
  border-radius: 2px;
}

.dark .scrollbar-thin::-webkit-scrollbar-thumb {
  background-color: rgba(75, 85, 99, 0.5);
}
</style>