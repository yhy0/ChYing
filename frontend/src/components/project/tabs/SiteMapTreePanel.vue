<script setup lang="ts">
import { ref, computed, onMounted, watch, nextTick, onBeforeUnmount } from 'vue';
import { useI18n } from 'vue-i18n';
import { useProjectStore } from '../../../store/project';
import type { SiteMapNode } from '../../../types/project';
import {Events} from "@wailsio/runtime";
// @ts-ignore - 忽略模块导入错误
import {
  GetHistoryDumpIndex,
  Dashboard
} from "../../../../bindings/github.com/yhy0/ChYing/app.js";

const { t } = useI18n();
const projectStore = useProjectStore();

// 站点地图数据 - 使用store中的数据
const siteMapData = computed(() => projectStore.siteMapData);
// 当前选中的主机
const host = ref<string>('');
// 选中的节点
const selectedNode = ref<SiteMapNode | null>(null);
// 搜索查询
const searchQuery = ref('');

// 添加面板宽度监听
const containerRef = ref<HTMLElement | null>(null);
const containerWidth = ref(0);
const nodeLabelWidth = computed(() => Math.max(160, containerWidth.value * 0.7 - 50));
const childLabelWidth = computed(() => Math.max(140, containerWidth.value * 0.7 - 80));
const grandchildLabelWidth = computed(() => Math.max(120, containerWidth.value * 0.7 - 110));

// 监听容器大小变化
const resizeObserver = typeof ResizeObserver !== 'undefined' 
  ? new ResizeObserver((entries) => {
      if (entries[0]) {
        containerWidth.value = entries[0].contentRect.width;
      }
    })
  : null;

onMounted(() => {
  if (containerRef.value && resizeObserver) {
    resizeObserver.observe(containerRef.value);
    containerWidth.value = containerRef.value.offsetWidth;
  }

  // 如果store中已经有数据，不需要重新加载
  if (siteMapData.value.length > 0) {
    return;
  }

  // 主动获取初始数据
  console.log('主动获取Dashboard数据');
  Dashboard()
    .then((result: any) => {
      if (result && Array.isArray(result)) {
        processDashboardData(result);
      } else {
        console.log('获取到的数据格式不正确:', result);
      }
    })
    .catch((error: any) => {
      console.error('获取Dashboard数据失败:', error);
    });

  // 继续监听Dashboard事件以获取实时更新
  Events.On('Dashboard', (result: any) => {
    // 处理事件数据
    let targets: any[] = [];
    
    // 基于日志分析，数据结构为[目标数组]
    if (Array.isArray(result) && result.length > 0 && Array.isArray(result[0])) {
      // 提取目标数组
      targets = result[0];
    } else if (result && result.data && Array.isArray(result.data) && result.data.length > 0) {
      // 处理带有data属性的格式
      const res = result.data;
      if (Array.isArray(res[0])) {
        targets = res[0];
      } else if (res[0] && res[0].target) {
        targets = res;
      }
    } else if (Array.isArray(result) && result.length > 0 && result[0] && result[0].target) {
      // 直接是目标数组
      targets = result;
    }
    
    // 如果没有有效目标，则不更新
    if (targets.length === 0) {
      console.log('无效的目标数据');
      return;
    }
    
    // 检查是否与当前数据相同
    if (siteMapData.value.length > 0) {
      const currentHosts = siteMapData.value.map(node => node.name);
      const newHosts = targets.map(t => t.target);
      
      // 检查主机列表是否相同
      const hostsMatch = currentHosts.length === newHosts.length && 
                         currentHosts.every(host => newHosts.includes(host));
      
      if (hostsMatch) {
        // 如果主机列表相同，检查是否有站点地图更新
        let mapChanged = false;
        for (let i = 0; i < targets.length; i++) {
          const target = targets[i];
          const node = siteMapData.value.find(n => n.name === target.target || 
                                               n.name.startsWith(target.target + ' '));
          
          if (node && target.site_map && Array.isArray(target.site_map)) {
            // 检查站点地图URL数量是否变化
            const currentUrlCount = countUrls(node);
            if (currentUrlCount !== target.site_map.length) {
              mapChanged = true;
              break;
            }
          }
        }
        
        if (!mapChanged) {
          return;
        }
      }
    }
    
    // 处理数据更新
    processDashboardData(targets);
  });
});

onBeforeUnmount(() => {
  if (resizeObserver && containerRef.value) {
    resizeObserver.unobserve(containerRef.value);
  }
});

// 添加子节点
const addSiteMapChildNodes = (rootNode: SiteMapNode, targetHost: string, urls: string[], startId: number) => {
  let idCounter = startId;
  
  urls.forEach((url: string) => {
    let processedUrl = url;
    
    // 移除协议和主机名
    processedUrl = processedUrl.replace(`http://${targetHost}`, '');
    processedUrl = processedUrl.replace(`https://${targetHost}`, '');
    
    if (processedUrl === '') {
      processedUrl = '/';
      return; // 根路径已经在root节点中
    }
    
    // 分割路径部分
    const parts = processedUrl.split('/').filter(Boolean);
    let currentNode = rootNode;
    let currentPath = '';
    
    // 对每个路径部分创建节点
    parts.forEach((part: string, index: number) => {
      currentPath += `/${part}`;
      
      // 检查当前节点的子节点中是否已存在此路径部分
      let found = currentNode.children.find(child => child.name === part);
      
      if (!found) {
        const isLastPart = index === parts.length - 1;
        const nodeType = isLastPart ? 'file' : 'directory';
        const icon = isLastPart ? 'bx-file' : 'bx-folder';
        
        // 创建新节点
        const newNode: SiteMapNode = {
          id: idCounter++,
          name: part,
          path: currentPath,
          fullUrl: url.includes('https://') ? `https://${targetHost}${currentPath}` : `http://${targetHost}${currentPath}`,
          nodeType: nodeType,
          icon: icon,
          isExpanded: false,
          children: []
        };
        
        // 添加到当前节点的子节点中
        currentNode.children.push(newNode);
        found = newNode;
      }
      
      // 移动到下一级
      currentNode = found;
    });
  });
};

// 处理Dashboard数据并转换为站点地图格式
const processDashboardData = (result: any) => {
  console.log('处理Dashboard数据:', result);
  
  // 检查数据结构
  if (!result || result.length === 0) {
    console.log('Dashboard数据为空');
    return;
  }
  
  // 处理数据前保存当前节点状态，用于后续恢复
  const expandedStates = new Map();
  if (siteMapData.value.length > 0) {
    // 保存所有节点的展开状态
    const saveExpandState = (nodes: SiteMapNode[]) => {
      nodes.forEach(node => {
        expandedStates.set(node.fullUrl, node.isExpanded);
        if (node.children && node.children.length > 0) {
          saveExpandState(node.children);
        }
      });
    };
    saveExpandState(siteMapData.value);
  }
  
  // 创建新的站点地图数据
  const newSiteMapData: SiteMapNode[] = [];
  let idCounter = 1;
  
  // 处理每个目标
  result.forEach((target: any) => {
    // 创建主机节点
    const hostNode: SiteMapNode = {
      id: idCounter++,
      name: target.target,
      path: '/',
      fullUrl: `https://${target.target}/`,
      nodeType: 'host',
      icon: 'bx-globe',
      isExpanded: true,
      children: [],
      rawData: target
    };
    
    // 如果有站点地图数据，添加子节点
    if (target.site_map && Array.isArray(target.site_map)) {
      addSiteMapChildNodes(hostNode, target.target, target.site_map, idCounter);
      idCounter += target.site_map.length * 2; // 预估子节点数量
    } else {
      // 如果没有站点地图数据，可以添加一些常见目录
      const commonPaths = ['admin', 'login', 'api', 'static', 'images'];
      commonPaths.forEach(path => {
        hostNode.children.push({
          id: idCounter++,
          name: path,
          path: `/${path}`,
          fullUrl: `https://${target.target}/${path}`,
          nodeType: 'directory',
          icon: 'bx-folder',
          isExpanded: false,
          children: []
        });
      });
    }
    
    // 添加到站点地图数据
    newSiteMapData.push(hostNode);
  });
  
  // 恢复节点展开状态
  const restoreExpandState = (nodes: SiteMapNode[]) => {
    nodes.forEach(node => {
      const savedState = expandedStates.get(node.fullUrl);
      if (savedState !== undefined) {
        node.isExpanded = savedState;
      }
      if (node.children && node.children.length > 0) {
        restoreExpandState(node.children);
      }
    });
  };
  restoreExpandState(newSiteMapData);
  
  // 使用store更新站点地图数据
  projectStore.updateSiteMapData(newSiteMapData);
  
  // 如果有目标，设置第一个目标为当前主机
  if (result[0] && result[0].target) {
    host.value = result[0].target;
  }
};

// 辅助函数：计算节点包含的URL总数
const countUrls = (node: SiteMapNode): number => {
  let count = 0;
  if (node.nodeType === 'file') {
    count = 1;
  }
  if (node.children && node.children.length > 0) {
    count += node.children.reduce((sum, child) => sum + countUrls(child), 0);
  }
  return count;
};

// 搜索功能的计算属性
const flattenedNodes = computed(() => {
  const result: SiteMapNode[] = [];
  
  function flatten(nodes: SiteMapNode[]) {
    nodes.forEach(node => {
      result.push(node);
      if (node.children && node.children.length > 0) {
        flatten(node.children);
      }
    });
  }
  
  flatten(siteMapData.value);
  return result;
});

const filteredNodes = computed(() => {
  if (!searchQuery.value.trim()) {
    return siteMapData.value;
  }
  
  const query = searchQuery.value.toLowerCase();
  const matches = flattenedNodes.value.filter(node => 
    node.fullUrl.toLowerCase().includes(query)
  );
  
  // 父节点ID集合，用于确保树形结构完整
  const parentIds = new Set<number>();
  
  // 递归查找所有父节点
  function findParents(node: SiteMapNode) {
    const parent = flattenedNodes.value.find(n => 
      n.children.some(child => child.id === node.id)
    );
    
    if (parent) {
      parentIds.add(parent.id);
      findParents(parent);
    }
  }
  
  // 为所有匹配节点找到其父节点
  matches.forEach(node => findParents(node));
  
  // 构建新的树结构，只包含匹配节点及其父节点
  function filterTree(nodes: SiteMapNode[]): SiteMapNode[] {
    return nodes
      .filter(node => matches.some(m => m.id === node.id) || parentIds.has(node.id))
      .map(node => ({
        ...node,
        isExpanded: true, // 自动展开搜索结果
        children: filterTree(node.children)
      }));
  }
  
  return filterTree(siteMapData.value);
});

// 工具函数
const toggleExpand = (node: SiteMapNode) => {
  node.isExpanded = !node.isExpanded;
};

// 节点选择函数
const selectNode = (node: SiteMapNode) => {
  selectedNode.value = node;
  
  // 首先发送节点选择事件，传递空数据，让UI立即更新
  emit('node-selected', node, "", "");
  
  console.log('选择节点:', node.name, node.path, node.fullUrl);
  
  // 如果是文件节点或者有明确路径的节点，尝试获取请求响应数据
  if (node.nodeType === 'file' || (node.path !== '/' && node.path.length > 1)) {
    // 准备请求的参数，使用节点的fullUrl作为key
    const url = node.fullUrl;
    
    console.log('尝试获取请求响应数据:', url);
    
    // 调用GetHistoryDumpIndex获取请求响应数据
    GetHistoryDumpIndex(url).then((HTTPBody: any) => {
      console.log('获取到请求响应:', HTTPBody);
      
      if (HTTPBody === null) {
        // 如果没有数据，发送空数据
        console.log('没有找到请求响应数据');
        emit('node-selected', node, "", "");
        return;
      }
      
      // 确保我们获取到了正确的数据格式
      // 后端返回的字段名是 request_raw 和 response_raw
      let request = HTTPBody["request_raw"] || HTTPBody["request"] || "";
      let response = HTTPBody["response_raw"] || HTTPBody["response"] || "";
      
      // 如果数据是对象，则转为JSON字符串
      if (typeof request === 'object') request = JSON.stringify(request, null, 2);
      if (typeof response === 'object') response = JSON.stringify(response, null, 2);
      
      // 发送节点和请求响应数据
      console.log('发送请求响应数据');
      emit('node-selected', node, request, response);
    }).catch((error: Error) => {
      console.error('获取HTTP请求响应失败:', error);
      emit('node-selected', node, "", "");
    });
  } else {
    // 如果是根节点或目录节点，只发送节点信息
    console.log('根节点或目录节点，不获取请求响应');
    emit('node-selected', node, "", "");
  }
};

// 图标样式函数
const getNodeIconClass = (node: SiteMapNode) => {
  switch (node.nodeType) {
    case 'host':
      return 'bx bx-globe text-blue-500';
    case 'directory':
      return 'bx bx-folder text-yellow-500';
    case 'file':
      if (node.path.endsWith('.js')) return 'bx bx-code-alt text-green-500';
      if (node.path.endsWith('.css')) return 'bx bx-code-curly text-purple-500';
      if (node.path.endsWith('.png') || node.path.endsWith('.jpg') || node.path.endsWith('.gif')) 
        return 'bx bx-image text-orange-500';
      if (node.path.includes('admin')) return 'bx bx-lock text-red-500';
      return 'bx bx-file text-gray-500';
    default:
      return 'bx bx-file text-gray-500';
  }
};

// 定义事件
const emit = defineEmits<{
  (e: 'node-selected', node: SiteMapNode, request?: string, response?: string): void
}>();

// 导出需要暴露给父组件的方法
defineExpose({
  host
});
</script>

<template>
  <!-- 主容器，设置为100%高度，防止滚动区域被压缩 -->
  <div ref="containerRef" class="site-map-container">
    <!-- 固定高度的搜索框区域 -->
    <div class="site-map-search">
      <div class="relative">
        <input 
          v-model="searchQuery"
          type="text" 
          class="w-full bg-[#f3f4f6] dark:bg-[#282838] border border-gray-200 dark:border-gray-700 rounded-md py-2 pl-9 pr-3 text-sm" 
          :placeholder="t('modules.project.site_map.search')"
          spellcheck="false"
        >
        <i class="bx bx-search absolute left-3 top-2.5 text-gray-500"></i>
        <button 
          v-if="searchQuery" 
          @click="searchQuery = ''" 
          class="absolute right-3 top-2.5 text-gray-500 hover:text-gray-700"
        >
          <i class="bx bx-x"></i>
        </button>
      </div>
    </div>
    
    <!-- 滚动区域，使用absolute定位占满剩余空间 -->
    <div class="site-map-scroll-container">
      <div class="p-2">
        <template v-for="node in filteredNodes" :key="node.id">
          <div class="pl-2">
            <div 
              class="flex items-center py-2 px-2 rounded cursor-pointer"
              :class="{'bg-[#f9fafb] dark:bg-[#28283e]': selectedNode?.id === node.id, 'hover:bg-[#f9fafb] dark:hover:bg-[#28283e]': selectedNode?.id !== node.id}"
              @click="selectNode(node)"
            >
              <i 
                v-if="node.children && node.children.length > 0"
                class="bx mr-1 text-gray-500 cursor-pointer flex-shrink-0" 
                :class="node.isExpanded ? 'bx-chevron-down' : 'bx-chevron-right'"
                @click.stop="toggleExpand(node)"
              ></i>
              <span v-else class="w-4 mr-1 flex-shrink-0"></span>
              <i :class="[getNodeIconClass(node), 'mr-2 flex-shrink-0']"></i>
              <span class="text-sm truncate flex-1" :style="{ maxWidth: `${nodeLabelWidth}px` }" :title="node.name">{{ node.name }}</span>
            </div>
            
            <div v-if="node.isExpanded && node.children && node.children.length > 0" class="border-l border-gray-200 dark:border-gray-700 ml-3">
              <template v-for="child in node.children" :key="child.id">
                <div class="pl-2">
                  <div 
                    class="flex items-center py-2 px-2 rounded cursor-pointer"
                    :class="{'bg-[#f9fafb] dark:bg-[#28283e]': selectedNode?.id === child.id, 'hover:bg-[#f9fafb] dark:hover:bg-[#28283e]': selectedNode?.id !== child.id}"
                    @click="selectNode(child)"
                  >
                    <i 
                      v-if="child.children && child.children.length > 0"
                      class="bx mr-1 text-gray-500 cursor-pointer flex-shrink-0" 
                      :class="child.isExpanded ? 'bx-chevron-down' : 'bx-chevron-right'"
                      @click.stop="toggleExpand(child)"
                    ></i>
                    <span v-else class="w-4 mr-1 flex-shrink-0"></span>
                    <i :class="[getNodeIconClass(child), 'mr-2 flex-shrink-0']"></i>
                    <span class="text-sm truncate flex-1" :style="{ maxWidth: `${childLabelWidth}px` }" :title="child.name">{{ child.name }}</span>
                  </div>
                  
                  <div v-if="child.isExpanded && child.children && child.children.length > 0" class="border-l border-gray-200 dark:border-gray-700 ml-3">
                    <template v-for="grandchild in child.children" :key="grandchild.id">
                      <div class="pl-2">
                        <div 
                          class="flex items-center py-2 px-2 rounded cursor-pointer"
                          :class="{'bg-[#f9fafb] dark:bg-[#28283e]': selectedNode?.id === grandchild.id, 'hover:bg-[#f9fafb] dark:hover:bg-[#28283e]': selectedNode?.id !== grandchild.id}"
                          @click="selectNode(grandchild)"
                        >
                          <span class="w-4 mr-1 flex-shrink-0"></span>
                          <i :class="[getNodeIconClass(grandchild), 'mr-2 flex-shrink-0']"></i>
                          <span class="text-sm truncate flex-1" :style="{ maxWidth: `${grandchildLabelWidth}px` }" :title="grandchild.name">{{ grandchild.name }}</span>
                        </div>
                      </div>
                    </template>
                  </div>
                </div>
              </template>
            </div>
          </div>
        </template>
      </div>
    </div>
  </div>
</template>

<style scoped>
.text-2xs {
  font-size: 0.65rem;
}

/* 站点地图主容器 */
.site-map-container {
  position: relative;
  height: 100%;
  width: 100%;
  display: flex;
  flex-direction: column;
  overflow: hidden;
  border-right: 1px solid var(--border-color, #e5e7eb);
  min-height: 0; /* 确保flex子项正确收缩 */
}

/* 暗黑模式下的边框颜色 */
.dark .site-map-container {
  border-color: var(--dark-border-color, #374151);
}

/* 搜索框部分 */
.site-map-search {
  padding: 0.75rem;
  flex: none;
  z-index: 2;
  background-color: white;
}

.dark .site-map-search {
  background-color: #1e1e36;
}

/* 滚动区域 */
.site-map-scroll-container {
  position: absolute;
  top: 4rem; /* 搜索框高度 */
  left: 0;
  right: 0;
  bottom: 0;
  overflow-y: auto;
  overflow-x: hidden;
  background-color: white;
  padding-bottom: 1rem; /* 底部留白 */
  height: calc(100% - 4rem); /* 明确设置高度为总高度减去搜索框高度 */
}

.dark .site-map-scroll-container {
  background-color: #1e1e36;
}

/* 添加滚动条样式 */
.site-map-scroll-container {
  scrollbar-width: thin;
  scrollbar-color: rgba(156, 163, 175, 0.5) transparent;
}

.site-map-scroll-container::-webkit-scrollbar {
  width: 4px;
}

.site-map-scroll-container::-webkit-scrollbar-track {
  background-color: transparent;
}

.site-map-scroll-container::-webkit-scrollbar-thumb {
  background-color: rgba(156, 163, 175, 0.5);
  border-radius: 2px;
}

.dark .site-map-scroll-container::-webkit-scrollbar-thumb {
  background-color: rgba(75, 85, 99, 0.5);
}

/* 添加文字截断提示相关样式 */
.truncate {
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}
</style> 