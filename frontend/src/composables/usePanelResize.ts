import { ref, onMounted, onBeforeUnmount } from 'vue';

/**
 * 通用面板大小调整组合式函数
 * 用于处理表格和面板的拖拽调整
 */
interface UsePanelResizeOptions {
  panelId: string;
  initialHeight?: number;
  minHeight?: number;
  maxHeightOffset?: number;
}

export function usePanelResize(options: UsePanelResizeOptions) {
  const { 
    panelId, 
    initialHeight = 300, 
    minHeight = 100, 
    maxHeightOffset = 200 
  } = options;

  const panelHeight = ref(initialHeight);
  
  let isResizing = false;
  let startY = 0;
  let startHeight = 0;

  const localStorageKey = `${panelId}-height`;

  const saveHeight = () => {
    localStorage.setItem(localStorageKey, panelHeight.value.toString());
  };

  onMounted(() => {
    const savedHeight = localStorage.getItem(localStorageKey);
    if (savedHeight !== null) {
      const parsedHeight = parseInt(savedHeight, 10);
      if (!isNaN(parsedHeight)) {
        // Ensure loaded height is within min/max bounds if needed, 
        // or simply apply it and let user adjust if it's out of current dynamic bounds.
        // For now, we'll just apply it.
        panelHeight.value = Math.max(minHeight, Math.min(window.innerHeight - maxHeightOffset, parsedHeight));
      }
    } else {
      // If no saved height, ensure initialHeight is within current dynamic bounds too
      panelHeight.value = Math.max(minHeight, Math.min(window.innerHeight - maxHeightOffset, initialHeight));
    }
  });
  
  const startResize = (e: MouseEvent) => {
    isResizing = true;
    startY = e.clientY;
    startHeight = panelHeight.value;
    
    document.addEventListener('mousemove', handleResize);
    document.addEventListener('mouseup', stopResize);
    document.body.classList.add('cursor-ns-resize');
  };
  
  const handleResize = (e: MouseEvent) => {
    if (!isResizing) return;
    
    const delta = e.clientY - startY;
    panelHeight.value = Math.max(
      minHeight,
      Math.min(window.innerHeight - maxHeightOffset, startHeight + delta)
    );
  };
  
  const stopResize = () => {
    if (!isResizing) return;
    isResizing = false;
    
    document.removeEventListener('mousemove', handleResize);
    document.removeEventListener('mouseup', stopResize);
    document.body.classList.remove('cursor-ns-resize');
    saveHeight(); // Save height on resize end
  };
  
  onBeforeUnmount(() => {
    if (isResizing) {
      stopResize();
    }
  });
  
  return {
    panelHeight,
    startResize,
    handleResize,
    stopResize
  };
} 