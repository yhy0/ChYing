<script setup lang="ts">
import { ref, computed, watch, nextTick, onMounted, onBeforeUnmount } from 'vue';
import { 
  normalizeData, 
  useRequestViewManager,
  useResponseViewManager
} from '../../utils';
import { RequestPanel, ResponsePanel, FunctionModulesBar, InspectorPanel, NotesPanel, ExtractorPanel } from './requestResponse';
// @ts-ignore
import { Jsluice } from "../../../bindings/github.com/yhy0/ChYing/app.js";

// 定义组件props
const props = defineProps<{
  requestData: string;
  responseData: string;
  requestReadOnly?: boolean;
  responseReadOnly?: boolean;
  title?: {
    request?: string;
    response?: string;
  };
  hideEmptyResponse?: boolean; // 是否隐藏空响应
  uuid?: string; // 请求/响应的唯一标识符
  serverDurationMs: number;
}>();

// 定义组件事件
const emit = defineEmits<{
  (e: 'update:requestData', data: string): void;
  (e: 'update:responseData', data: string): void;
}>();

// 编辑器引用
const requestViewer = ref<any>(null);
const responseViewer = ref<any>(null);
const containerRef = ref<HTMLElement | null>(null);

// 请求和响应的标题
const requestTitle = computed(() => props.title?.request || 'Request');
const responseTitle = computed(() => props.title?.response || 'Response');
const serverDurationMs = computed(() => props.serverDurationMs || 0);

// 转换请求和响应数据为普通对象
const normalizedRequestData = computed(() => normalizeData(props.requestData || ''));
const normalizedResponseData = computed(() => normalizeData(props.responseData || ''));

// 使用视图管理器
const { 
  requestViewType, 
  setRequestViewType
} = useRequestViewManager(normalizedRequestData.value);

const { 
  responseViewType, 
  setResponseViewType,
  isResponseEmpty
} = useResponseViewManager(normalizedResponseData.value);

// 修改方法以适应类型要求
const handleRequestViewTypeChange = (type: string) => {
  setRequestViewType(type as any);
};

const handleResponseViewTypeChange = (type: string) => {
  setResponseViewType(type as any);
};

// 是否显示响应
const shouldShowResponse = computed(() => {
  if (props.hideEmptyResponse && isResponseEmpty.value) {
    return false;
  }
  return true;
});

// 拖拽调整面板大小相关变量
const requestWidth = ref(50); // 初始宽度50%
let isDragging = false;
let initialX = 0;
let initialLeftWidth = 0;

// 计算请求部分的宽度
const computedRequestWidth = computed(() => {
  return shouldShowResponse.value ? requestWidth.value : 100;
});

// 计算响应部分的宽度
const computedResponseWidth = computed(() => {
  return 100 - computedRequestWidth.value;
});

// 更新CSS变量
watch(computedRequestWidth, (newWidth) => {
  if (containerRef.value) {
    containerRef.value.style.setProperty('--request-width', `${newWidth}%`);
  }
}, { immediate: true });

// 开始拖拽
const startResizeRequest = (e: MouseEvent) => {
  e.preventDefault();
  isDragging = true;
  initialX = e.clientX;
  
  // 获取初始请求面板宽度
  if (containerRef.value) {
    const containerRect = containerRef.value.getBoundingClientRect();
    initialLeftWidth = (requestWidth.value / 100) * containerRect.width;
  }
  
  // 添加全局事件监听器
  document.addEventListener('mousemove', handleMouseMove);
  document.addEventListener('mouseup', stopResize);
  
  // 添加拖拽中的样式
  document.body.classList.add('dragging');
};

// 拖拽过程
const handleMouseMove = (e: MouseEvent) => {
  if (!isDragging || !containerRef.value) return;
  
  const containerRect = containerRef.value.getBoundingClientRect();
  const deltaX = e.clientX - initialX;
  const newLeftPanelWidth = initialLeftWidth + deltaX;
  
  // 计算百分比宽度
  const percentWidth = (newLeftPanelWidth / containerRect.width) * 100;
  
  // 限制宽度范围（20% - 80%）
  requestWidth.value = Math.max(20, Math.min(80, percentWidth));
};

// 停止拖拽
const stopResize = () => {
  isDragging = false;
  
  // 移除全局事件监听器
  document.removeEventListener('mousemove', handleMouseMove);
  document.removeEventListener('mouseup', stopResize);
  
  // 移除拖拽中的样式
  document.body.classList.remove('dragging');
};

// 监听传入数据的变化
watch(
  () => [props.requestData, props.responseData],
  () => {
    nextTick(() => {
      if (requestViewer.value?.$el || responseViewer.value?.$el) {
        setTimeout(() => {
          requestViewer.value?.$el?.click?.();
          responseViewer.value?.$el?.click?.();
        }, 50);
      }
    });
  },
  { deep: true }
);

// 处理请求更新
const handleRequestUpdate = (data: string) => {
  emit('update:requestData', data);
};

// 处理响应更新
const handleResponseUpdate = (data: string) => {
  emit('update:responseData', data);
};

// 功能模块管理 - 存储活跃模块与UUID的映射
const activeModules = ref<Record<string, string | null>>({});

// 获取当前请求对应的活跃模块
const activeModule = computed(() => {
  const requestId = props.uuid || 'default';
  return activeModules.value[requestId] || null;
});

// 切换模块
const toggleModule = (moduleName: string) => {
  const requestId = props.uuid || 'default';
  
  if (activeModules.value[requestId] === moduleName) {
    activeModules.value[requestId] = null; // 如果点击已激活的模块，则关闭它
  } else {
    activeModules.value[requestId] = moduleName; // 否则激活点击的模块
  }
};

// 面板宽度 - 根据UUID存储不同请求的面板宽度
const panelWidthsStore = ref<Record<string, Record<string, number>>>({});

// 获取当前请求的面板宽度
const panelWidths = computed(() => {
  const requestId = props.uuid || 'default';
  
  if (!panelWidthsStore.value[requestId]) {
    // 初始化该请求的面板宽度
    panelWidthsStore.value[requestId] = {
      inspector: 320,
      notes: 320,
      extractor: 350 // 新增 extractor 面板默认宽度
    };
  } else if (!panelWidthsStore.value[requestId].extractor) {
    // 如果已存在 requestId 但没有 extractor，则添加
    panelWidthsStore.value[requestId].extractor = 350;
  }
  
  return panelWidthsStore.value[requestId];
});

// 拖动调整功能面板大小
let isResizingPanel = false;
let startResizeX = 0;
let currentPanelWidth = 0;
let currentPanel = '';

const startResizePanel = (e: MouseEvent, panel: string) => {
  isResizingPanel = true;
  currentPanel = panel;
  startResizeX = e.clientX;
  
  const requestId = props.uuid || 'default';
  currentPanelWidth = panelWidthsStore.value[requestId][panel];
  
  document.addEventListener('mousemove', handlePanelResize);
  document.addEventListener('mouseup', stopPanelResize);
  
  // 添加cursor类
  document.body.classList.add('cursor-ew-resize');
};

const handlePanelResize = (e: MouseEvent) => {
  if (!isResizingPanel) return;
  
  // 计算鼠标移动的距离
  const delta = startResizeX - e.clientX; // 注意方向，左移增加宽度
  
  // 更新面板宽度
  let newWidth = currentPanelWidth + delta;
  
  // 限制最小和最大宽度
  newWidth = Math.max(280, Math.min(600, newWidth)); // 调整最大宽度
  
  // 更新宽度
  const requestId = props.uuid || 'default';
  panelWidthsStore.value[requestId][currentPanel] = newWidth;
};

const stopPanelResize = () => {
  isResizingPanel = false;
  document.removeEventListener('mousemove', handlePanelResize);
  document.removeEventListener('mouseup', stopPanelResize);
  
  // 移除cursor类
  document.body.classList.remove('cursor-ew-resize');
};

// 笔记内容 - 根据UUID存储不同请求的笔记
const notesContent = ref<Record<string, string>>({});

// 获取当前请求的笔记内容
const currentNotes = computed({
  get: () => {
    const requestId = props.uuid || 'default';
    return notesContent.value[requestId] || '';
  },
  set: (value: string) => {
    const requestId = props.uuid || 'default';
    notesContent.value[requestId] = value;
  }
});

// 更新笔记内容
const updateNotes = (value: string) => {
  currentNotes.value = value;
};

// 使用Inspector面板状态管理 - 根据UUID存储不同请求的展开状态
const expandedSectionsStore = ref<Record<string, Record<string, boolean>>>({});

// 获取当前请求的展开状态
const expandedSections = computed(() => {
  const requestId = props.uuid || 'default';
  
  if (!expandedSectionsStore.value[requestId]) {
    // 初始化该请求的展开状态
    expandedSectionsStore.value[requestId] = {
      requestAttributes: false,
      queryParameters: false,
      cookies: false,
      requestHeaders: false,
      responseHeaders: false
    };
  }
  
  return expandedSectionsStore.value[requestId];
});

// 切换展开折叠状态
const toggleSection = (section: string) => {
  const requestId = props.uuid || 'default';
  
  // 关闭其他已展开的部分，实现手风琴效果
  Object.keys(expandedSectionsStore.value[requestId]).forEach(key => {
    if (key !== section) {
      expandedSectionsStore.value[requestId][key] = false;
    }
  });
  
  // 切换当前部分的展开/折叠状态
  expandedSectionsStore.value[requestId][section] = !expandedSectionsStore.value[requestId][section];
};

// 关闭功能面板
const closePanel = () => {
  const requestId = props.uuid || 'default';
  activeModules.value[requestId] = null;
};

// Jsluice 提取结果
interface UrlItem { url: string; type?: string; source?: string; }
interface SecretItem { type: string; value: string; source?: string; }
interface JsluiceResultState {
  urls: UrlItem[];
  secrets: SecretItem[];
  loading: boolean;
  error: string | null;
}
const jsluiceResultsStore = ref<Record<string, JsluiceResultState>>({});

// 获取当前请求的Jsluice结果
const currentJsluiceResults = computed<JsluiceResultState>(() => {
  const requestId = props.uuid || 'default';
  if (!jsluiceResultsStore.value[requestId]) {
    jsluiceResultsStore.value[requestId] = { urls: [], secrets: [], loading: false, error: null };
  }
  return jsluiceResultsStore.value[requestId];
});

// 监听 Extractor 模块激活和响应数据变化
watch(
  () => ({
    active: activeModule.value,
    responseData: props.responseData,
    uuid: props.uuid
  }),
  async (newVal, oldVal) => {
    const requestId = newVal.uuid || 'default';
    if (newVal.active === 'extractor') {
      const responseBodyChanged = newVal.responseData !== oldVal?.responseData;
      const justActivated = newVal.active !== oldVal?.active && newVal.active === 'extractor';
      const uuidChangedAndIsExtractor = newVal.uuid !== oldVal?.uuid && newVal.active === 'extractor';

      if (justActivated || responseBodyChanged || uuidChangedAndIsExtractor || 
          (!jsluiceResultsStore.value[requestId]?.urls?.length && 
           !jsluiceResultsStore.value[requestId]?.secrets?.length && 
           !jsluiceResultsStore.value[requestId]?.loading)) {
        
        if (!jsluiceResultsStore.value[requestId]) {
          jsluiceResultsStore.value[requestId] = { urls: [], secrets: [], loading: false, error: null };
        }
        jsluiceResultsStore.value[requestId].loading = true;
        jsluiceResultsStore.value[requestId].error = null;
        
        try {
          let body = '';
          if (newVal.responseData) {
            const parts = newVal.responseData.split("\r\n\r\n");
            if (parts.length > 1) {
              body = parts.slice(1).join("\r\n\r\n");
            } else {
              const trimmedData = newVal.responseData.trim();
              if ((trimmedData.startsWith('{') && trimmedData.endsWith('}')) || (trimmedData.startsWith('[') && trimmedData.endsWith(']'))) {
                body = trimmedData;
              } else {
                body = newVal.responseData;
                console.warn("Extractor: responseData might not be a full HTTP response or simple JSON/Array. Using raw data for Jsluice.");
              }
            }
          }

          if (!body) {
            console.log("Extractor: Response body is empty. No data to process.");
            jsluiceResultsStore.value[requestId].urls = [];
            jsluiceResultsStore.value[requestId].secrets = [];
            jsluiceResultsStore.value[requestId].loading = false;
            return;
          }
          
          const rawResults = await Jsluice(body); // 使用真实的 Jsluice 函数

          const parsedUrls: UrlItem[] = [];
          if (rawResults && rawResults.urls && Array.isArray(rawResults.urls)) {
            rawResults.urls.forEach((item: any) => { // item 是 JSON 字符串
              try {
                if (typeof item === 'string') {
                  parsedUrls.push(JSON.parse(item));
                } else if (typeof item === 'object' && item !== null) {
                  // 如果 Jsluice 已经返回了解析好的对象，直接使用
                  parsedUrls.push(item as UrlItem);
                } else {
                  console.warn("Jsluice returned a URL item in an unexpected format:", item);
                }
              } catch (e) {
                console.error("Failed to parse URL item from Jsluice:", item, e);
              }
            });
          }

          const parsedSecrets: SecretItem[] = [];
          if (rawResults && rawResults.secrets && Array.isArray(rawResults.secrets)) {
            rawResults.secrets.forEach((item: any) => { // item 是 JSON 字符串
              try {
                if (typeof item === 'string') {
                  parsedSecrets.push(JSON.parse(item));
                } else if (typeof item === 'object' && item !== null) {
                   // 如果 Jsluice 已经返回了解析好的对象，直接使用
                  parsedSecrets.push(item as SecretItem);
                } else {
                  console.warn("Jsluice returned a secret item in an unexpected format:", item);
                }
              } catch (e) {
                console.error("Failed to parse secret item from Jsluice:", item, e);
              }
            });
          }
          
          jsluiceResultsStore.value[requestId].urls = parsedUrls;
          jsluiceResultsStore.value[requestId].secrets = parsedSecrets;
        } catch (err: any) {
          console.error("Error running Jsluice or processing results:", err);
          jsluiceResultsStore.value[requestId].error = err.message || "Failed to extract data.";
          jsluiceResultsStore.value[requestId].urls = [];
          jsluiceResultsStore.value[requestId].secrets = [];
        } finally {
          jsluiceResultsStore.value[requestId].loading = false;
        }
      }
    }
  },
  { deep: true, immediate: true }
);

// 组件挂载时清理可能存在的事件监听器
onMounted(() => {
  document.removeEventListener('mousemove', handleMouseMove);
  document.removeEventListener('mouseup', stopResize);
  document.body.classList.remove('dragging');
});

// 组件卸载时清理事件监听器
onBeforeUnmount(() => {
  document.removeEventListener('mousemove', handleMouseMove);
  document.removeEventListener('mouseup', stopResize);
  document.body.classList.remove('dragging');
});

// 添加换行状态管理
const requestWordWrap = ref(false);
const responseWordWrap = ref(false);

// 切换请求换行
const toggleRequestWordWrap = () => {
  requestWordWrap.value = !requestWordWrap.value;
};

// 切换响应换行
const toggleResponseWordWrap = () => {
  responseWordWrap.value = !responseWordWrap.value;
};
</script>

<template>
  <div class="flex-1 flex overflow-hidden request-response-container" ref="containerRef">
    <!-- 主内容区域 -->
    <div class="flex flex-1 overflow-hidden">
      <!-- 请求面板 -->
      <RequestPanel 
        :normalized-request-data="normalizedRequestData"
        :request-width="computedRequestWidth"
        :request-view-type="requestViewType"
        :read-only="requestReadOnly"
        :request-title="requestTitle"
        :word-wrap="requestWordWrap"
        @set-request-view-type="handleRequestViewTypeChange"
        @update:request-data="handleRequestUpdate"
        @toggle-word-wrap="toggleRequestWordWrap"
        ref="requestViewer"
      />

      <!-- 分隔线 -->
      <div 
        v-if="shouldShowResponse"
        class="panel-divider"
        @mousedown="startResizeRequest"
      ></div>

      <!-- 响应面板 -->
      <ResponsePanel 
        v-if="shouldShowResponse"
        :normalized-response-data="normalizedResponseData"
        :response-width="computedResponseWidth"
        :response-view-type="responseViewType"
        :read-only="responseReadOnly"
        :response-title="responseTitle"
        :serverDurationMs="serverDurationMs"
        :word-wrap="responseWordWrap"
        @set-response-view-type="handleResponseViewTypeChange"
        @update:response-data="handleResponseUpdate"
        @toggle-word-wrap="toggleResponseWordWrap"
        ref="responseViewer"
      />
    </div>

    <!-- 功能模块栏 -->
    <FunctionModulesBar 
      :active-module="activeModule"
      @toggle-module="toggleModule"
    />

    <!-- Inspector面板 -->
    <InspectorPanel 
      v-if="activeModule === 'inspector'"
      :panel-width="panelWidths.inspector"
      :expanded-sections="expandedSections"
      @close="closePanel"
      @toggle-section="toggleSection"
      @start-resize-panel="startResizePanel"
    />

    <!-- 笔记面板 -->
    <NotesPanel 
      v-if="activeModule === 'notes'"
      :panel-width="panelWidths.notes"
      :notes="currentNotes"
      @close="closePanel"
      @start-resize-panel="startResizePanel"
      @update:notes="updateNotes"
    />

    <!-- Extractor 面板 -->
    <ExtractorPanel
      v-if="activeModule === 'extractor'"
      :panel-width="panelWidths.extractor"
      :results="currentJsluiceResults"
      @close="closePanel"
      @start-resize-panel="(panel, event) => startResizePanel(event, panel)"
    />
  </div>
</template>

<!-- 样式已迁移到 styles/components/editor-panel.css 统一管理 -->
