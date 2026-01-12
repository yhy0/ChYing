import { ref, onBeforeUnmount, onMounted } from 'vue';

// 面板调整相关的组合式API
export function useIntruderPanelResizer() {
  // 拖拽调整面板大小相关状态
  const isDragging = ref(false);
  const initialX = ref(0);
  const initialLeftWidth = ref(0);
  const leftPanelWidth = ref(50); // 初始宽度50%
  const leftPanel = ref<HTMLElement | null>(null);
  const rightPanel = ref<HTMLElement | null>(null);
  const containerRef = ref<HTMLElement | null>(null);
  const hasInitialized = ref(false);

  // 初始化函数，可以在onMounted中调用
  const initializePanels = () => {
    if (!containerRef.value) {
      containerRef.value = document.querySelector('.intruder-resizable-panels');
    }
    
    if (!leftPanel.value) {
      leftPanel.value = document.querySelector('.intruder-panel-left');
    }
    
    if (!rightPanel.value) {
      rightPanel.value = document.querySelector('.intruder-panel-right');
    }
    
    hasInitialized.value = !!(containerRef.value && leftPanel.value && rightPanel.value);
    
    return hasInitialized.value;
  };

  // 开始拖拽
  const startResize = (e: MouseEvent) => {
    e.preventDefault();
    e.stopPropagation();
    
    // 如果面板引用未初始化，尝试初始化
    if (!hasInitialized.value) {
      if (!initializePanels()) {
        console.error('无法初始化面板引用，拖拽操作被取消');
        return;
      }
    }
    
    isDragging.value = true;
    initialX.value = e.clientX;
    
    if (leftPanel.value) {
      const leftPanelRect = leftPanel.value.getBoundingClientRect();
      initialLeftWidth.value = leftPanelRect.width;
    }
    
    // 添加全局事件监听器
    document.addEventListener('mousemove', handleMouseMove, { capture: true });
    document.addEventListener('mouseup', stopResize, { capture: true });
    
    // 添加拖拽中的样式
    document.body.classList.add('dragging');
    
    console.log('开始拖拽操作：', {
      initialX: initialX.value,
      initialLeftWidth: initialLeftWidth.value,
      isDragging: isDragging.value,
      panels: {
        container: containerRef.value,
        left: leftPanel.value,
        right: rightPanel.value
      }
    });
  };

  // 拖拽过程
  const handleMouseMove = (e: MouseEvent) => {
    if (!isDragging.value || !containerRef.value || !leftPanel.value || !rightPanel.value) {
      return;
    }
    
    e.preventDefault();
    e.stopPropagation();
    
    const containerRect = containerRef.value.getBoundingClientRect();
    const deltaX = e.clientX - initialX.value;
    const newLeftPanelWidth = initialLeftWidth.value + deltaX;
    
    // 计算百分比宽度
    const percentWidth = (newLeftPanelWidth / containerRect.width) * 100;
    
    // 限制宽度范围（20% - 80%）
    leftPanelWidth.value = Math.max(20, Math.min(80, percentWidth));
    
    // 应用新宽度
    leftPanel.value.style.width = `${leftPanelWidth.value}%`;
    rightPanel.value.style.width = `${100 - leftPanelWidth.value}%`;
    
    // 防止文本选择
    window.getSelection()?.removeAllRanges();
  };

  // 停止拖拽
  const stopResize = (e?: MouseEvent) => {
    if (e) {
      e.preventDefault();
      e.stopPropagation();
    }
    
    isDragging.value = false;
    
    // 移除全局事件监听器
    document.removeEventListener('mousemove', handleMouseMove, { capture: true });
    document.removeEventListener('mouseup', stopResize, { capture: true });
    
    // 恢复正常鼠标样式
    document.body.classList.remove('dragging');
    
    console.log('拖拽操作结束，最终宽度：', {
      leftWidth: `${leftPanelWidth.value}%`,
      rightWidth: `${100 - leftPanelWidth.value}%`
    });
  };

  // 清理事件监听器
  const cleanupResizer = () => {
    document.removeEventListener('mousemove', handleMouseMove, { capture: true });
    document.removeEventListener('mouseup', stopResize, { capture: true });
    document.body.classList.remove('dragging');
    
    console.log('已清理拖拽事件监听器');
  };

  // 组件卸载时自动清理
  onBeforeUnmount(() => {
    cleanupResizer();
  });

  // 自动尝试初始化
  onMounted(() => {
    // 延迟初始化，确保DOM已经渲染
    setTimeout(() => {
      initializePanels();
    }, 100);
  });

  return {
    isDragging,
    leftPanelWidth,
    leftPanel,
    rightPanel,
    containerRef,
    startResize,
    stopResize,
    cleanupResizer,
    initializePanels,
    hasInitialized
  };
} 