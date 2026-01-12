/**
 * 面板拖拽调整大小组合式函数
 * 
 * 提供可复用的面板拖拽调整大小功能
 * 支持水平和垂直方向的拖拽，以及最小/最大尺寸限制
 */

import { ref, reactive, onMounted, onUnmounted, readonly, type Ref } from 'vue';

export interface PanelResizerOptions {
  /** 拖拽方向 */
  direction?: 'horizontal' | 'vertical';
  /** 最小尺寸（像素或百分比） */
  minSize?: number;
  /** 最大尺寸（像素或百分比） */
  maxSize?: number;
  /** 是否使用百分比 */
  usePercentage?: boolean;
  /** 容器引用（用于计算百分比） */
  containerRef?: Ref<HTMLElement | null>;
  /** 拖拽时的回调 */
  onResize?: (size: number) => void;
  /** 拖拽开始的回调 */
  onResizeStart?: () => void;
  /** 拖拽结束的回调 */
  onResizeEnd?: (size: number) => void;
}

/**
 * 面板拖拽调整大小钩子
 */
export function usePanelResizer(
  initialSize: number,
  options: PanelResizerOptions = {}
) {
  const {
    direction = 'horizontal',
    minSize = 200,
    maxSize = Infinity,
    usePercentage = false,
    containerRef,
    onResize,
    onResizeStart,
    onResizeEnd
  } = options;

  // 当前尺寸
  const size = ref(initialSize);
  
  // 拖拽状态
  const isDragging = ref(false);
  const dragStartPos = ref(0);
  const dragStartSize = ref(0);

  /**
   * 获取容器尺寸
   */
  const getContainerSize = (): number => {
    if (!containerRef?.value) return 1000; // 默认容器大小
    
    const rect = containerRef.value.getBoundingClientRect();
    return direction === 'horizontal' ? rect.width : rect.height;
  };

  /**
   * 将像素值转换为百分比
   */
  const pixelsToPercentage = (pixels: number): number => {
    const containerSize = getContainerSize();
    return (pixels / containerSize) * 100;
  };

  /**
   * 将百分比转换为像素值
   */
  const percentageToPixels = (percentage: number): number => {
    const containerSize = getContainerSize();
    return (percentage / 100) * containerSize;
  };

  /**
   * 限制尺寸在有效范围内
   */
  const clampSize = (newSize: number): number => {
    let min = minSize;
    let max = maxSize;

    if (usePercentage && containerRef?.value) {
      // 如果使用百分比，需要转换限制值
      if (minSize < 1) {
        min = minSize * 100; // 假设传入的是 0-1 的比例
      }
      if (maxSize < 1) {
        max = maxSize * 100;
      }
    }

    return Math.max(min, Math.min(max, newSize));
  };

  /**
   * 开始拖拽
   */
  const startResize = (event: MouseEvent) => {
    event.preventDefault();
    
    isDragging.value = true;
    dragStartPos.value = direction === 'horizontal' ? event.clientX : event.clientY;
    dragStartSize.value = size.value;
    
    // 添加全局事件监听器
    document.addEventListener('mousemove', handleResize);
    document.addEventListener('mouseup', stopResize);
    document.body.style.cursor = direction === 'horizontal' ? 'col-resize' : 'row-resize';
    document.body.style.userSelect = 'none';
    
    onResizeStart?.();
  };

  /**
   * 处理拖拽
   */
  const handleResize = (event: MouseEvent) => {
    if (!isDragging.value) return;

    const currentPos = direction === 'horizontal' ? event.clientX : event.clientY;
    const delta = currentPos - dragStartPos.value;
    
    let newSize: number;
    
    if (usePercentage) {
      // 百分比模式
      const deltaPercentage = pixelsToPercentage(delta);
      newSize = clampSize(dragStartSize.value + deltaPercentage);
    } else {
      // 像素模式
      newSize = clampSize(dragStartSize.value + delta);
    }
    
    size.value = newSize;
    onResize?.(newSize);
  };

  /**
   * 停止拖拽
   */
  const stopResize = () => {
    if (!isDragging.value) return;
    
    isDragging.value = false;
    
    // 移除全局事件监听器
    document.removeEventListener('mousemove', handleResize);
    document.removeEventListener('mouseup', stopResize);
    document.body.style.cursor = '';
    document.body.style.userSelect = '';
    
    onResizeEnd?.(size.value);
  };

  /**
   * 设置尺寸
   */
  const setSize = (newSize: number) => {
    size.value = clampSize(newSize);
    onResize?.(size.value);
  };

  /**
   * 重置为初始尺寸
   */
  const reset = () => {
    setSize(initialSize);
  };

  // 清理事件监听器
  onUnmounted(() => {
    if (isDragging.value) {
      stopResize();
    }
  });

  return {
    size: readonly(size),
    isDragging: readonly(isDragging),
    startResize,
    setSize,
    reset,
    // 工具函数
    pixelsToPercentage,
    percentageToPixels,
    getContainerSize
  };
}

/**
 * 多面板拖拽调整大小钩子
 * 用于管理多个相互关联的面板
 */
export function useMultiPanelResizer(
  panels: Array<{ initialSize: number; minSize?: number; maxSize?: number }>,
  options: Omit<PanelResizerOptions, 'minSize' | 'maxSize'> = {}
) {
  const resizers = panels.map((panel, index) => {
    return usePanelResizer(panel.initialSize, {
      ...options,
      minSize: panel.minSize,
      maxSize: panel.maxSize,
      onResize: (size) => {
        options.onResize?.(size);
        // 可以在这里添加多面板联动逻辑
      }
    });
  });

  return {
    resizers,
    sizes: resizers.map(r => r.size),
    isDragging: resizers.some(r => r.isDragging),
    reset: () => resizers.forEach(r => r.reset())
  };
}
