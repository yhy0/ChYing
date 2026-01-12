/**
 * Inspector面板相关工具函数
 */

import { ref } from 'vue';

// Inspector面板部分类型定义
export type InspectorSection = 
  | 'requestAttributes' 
  | 'queryParameters' 
  | 'cookies' 
  | 'requestHeaders' 
  | 'responseHeaders';

/**
 * 创建Inspector面板的展开/折叠状态管理
 */
export function useInspectorSections() {
  // 各部分的展开状态
  const expandedSections = ref<Record<InspectorSection, boolean>>({
    requestAttributes: false,
    queryParameters: false,
    cookies: false,
    requestHeaders: false,
    responseHeaders: false
  });

  /**
   * 切换展开折叠状态 (手风琴效果)
   */
  const toggleSection = (section: InspectorSection) => {
    // 关闭其他已展开的部分，实现手风琴效果
    Object.keys(expandedSections.value).forEach(key => {
      if (key !== section) {
        expandedSections.value[key as InspectorSection] = false;
      }
    });
    // 切换当前部分的展开/折叠状态
    expandedSections.value[section] = !expandedSections.value[section];
  };

  return {
    expandedSections,
    toggleSection
  };
}

/**
 * 创建Inspector面板的显示/隐藏状态管理
 */
export function useInspectorVisibility() {
  // Inspector 是否显示
  const showInspector = ref(false);

  // 切换显示/隐藏状态
  const toggleInspector = () => {
    showInspector.value = !showInspector.value;
  };

  // 显示 Inspector
  const showInspectorPanel = () => {
    showInspector.value = true;
  };

  // 隐藏 Inspector
  const hideInspectorPanel = () => {
    showInspector.value = false;
  };

  return {
    showInspector,
    toggleInspector,
    showInspectorPanel,
    hideInspectorPanel
  };
}

/**
 * 面板宽度调整相关逻辑
 */
export function usePanelResize() {
  // 请求面板宽度
  const requestWidth = ref(50); // 初始宽度50%
  
  // 调整状态变量
  let isResizingRequest = false;
  let startX = 0;
  let startRequestWidth = 0;
  let resizeContainerWidth = 0;

  /**
   * 开始调整大小
   */
  const startResizeRequest = (e: MouseEvent) => {
    isResizingRequest = true;
    startX = e.clientX;
    startRequestWidth = requestWidth.value;
    
    // 容器宽度
    const container = document.querySelector('.request-response-container') as HTMLElement;
    if (container) {
      resizeContainerWidth = container.offsetWidth;
    }
    
    document.addEventListener('mousemove', handleRequestResize);
    document.addEventListener('mouseup', stopResizeRequest);
    
    // 添加cursor类到body以确保光标在整个拖动过程中保持一致
    document.body.classList.add('cursor-ew-resize');
  };

  /**
   * 处理调整过程
   */
  const handleRequestResize = (e: MouseEvent) => {
    if (!isResizingRequest) return;
    
    // 计算鼠标移动的像素数
    const delta = e.clientX - startX;
    
    // 将像素转换为百分比 (相对于容器宽度)
    const percentDelta = (delta / resizeContainerWidth) * 100;
    
    // 更新宽度百分比
    const newWidth = startRequestWidth + percentDelta;
    
    // 限制最小和最大宽度
    requestWidth.value = Math.max(20, Math.min(80, newWidth));
  };

  /**
   * 停止调整
   */
  const stopResizeRequest = () => {
    isResizingRequest = false;
    document.removeEventListener('mousemove', handleRequestResize);
    document.removeEventListener('mouseup', stopResizeRequest);
    
    // 移除cursor类
    document.body.classList.remove('cursor-ew-resize');
  };

  return {
    requestWidth,
    startResizeRequest,
    // 这些方法一般不需要暴露，但如果需要自定义调整过程，也可以选择暴露
    handleRequestResize,
    stopResizeRequest
  };
} 