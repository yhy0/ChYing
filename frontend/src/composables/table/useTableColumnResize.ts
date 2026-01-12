import { ref, onMounted, onBeforeUnmount } from 'vue';

interface UseTableColumnResizeOptions {
  tableId: string;
  defaultWidths: Record<string, number>;
}

export function useTableColumnResize(options: UseTableColumnResizeOptions) {
  const { tableId, defaultWidths } = options;
  
  // 列宽设置
  const columnWidths = ref<Record<string, number>>({});
  
  // 处理列调整逻辑状态
  const isResizing = ref(false);
  const resizingColumnId = ref<string | null>(null);
  const startX = ref(0);
  const initialColumnWidth = ref(0); // 存储调整开始时的初始列宽

  // 双击自动调整列宽
  const autoAdjustColumnWidth = (table?: any, columnId?: string) => {
    if (columnId) {
      // 单列调整
      const defaultWidth = defaultWidths[columnId] || 100;
      columnWidths.value[columnId] = defaultWidth;
    } else {
      // 全部列调整为默认宽度
      Object.assign(columnWidths.value, defaultWidths);
    }
    
    // 如果提供了表格实例，则使用它的setColumnSizing方法更新表格状态
    if (table && typeof table.setColumnSizing === 'function') {
      table.setColumnSizing({ ...columnWidths.value });
    }
    
    // 保存列宽到本地存储
    saveColumnWidths();
  };

  // 保存列宽到本地存储
  const saveColumnWidths = () => {
    localStorage.setItem(`${tableId}-column-widths`, JSON.stringify(columnWidths.value));
  };
  
  // 处理列调整开始
  const handleResizeStart = (event: MouseEvent, columnId: string) => {
    event.preventDefault();
    event.stopPropagation();
    console.log('handleResizeStart', event, columnId);
    isResizing.value = true;
    resizingColumnId.value = columnId;
    startX.value = event.clientX;
    initialColumnWidth.value = columnWidths.value[columnId] || defaultWidths[columnId] || 100;
    
    // 添加调整中的视觉指示
    document.body.classList.add('column-resizing');
    
    // 添加全局事件监听
    document.addEventListener('mousemove', handleResize);
    document.addEventListener('mouseup', handleResizeEnd);
    
    // 阻止点击排序
    return false;
  };

  // 处理列调整过程
  const handleResize = (event: MouseEvent) => {
    if (!isResizing.value || !resizingColumnId.value) return;
    
    // 计算拖动距离（相对于初始点击位置）
    const moveX = event.clientX - startX.value;
    
    // 当前列宽
    const columnId = resizingColumnId.value;
    
    // 计算新列宽（最小宽度限制为20px）
    const newWidth = Math.max(20, initialColumnWidth.value + moveX);
    
    // 更新列宽
    columnWidths.value[columnId] = newWidth;
    
    // 实时更新DOM以提供即时反馈
    // updateColumnWidthInDOM(columnId, newWidth);
    
    // 保持拖动时的鼠标样式
    document.body.style.cursor = 'col-resize';
  };

  // 处理列调整结束
  const handleResizeEnd = () => {
    console.log('handleResizeEnd triggered');
    if (!isResizing.value) return;
    
    isResizing.value = false;
    const columnId = resizingColumnId.value;
    resizingColumnId.value = null;
    
    // 移除全局事件监听
    document.removeEventListener('mousemove', handleResize);
    document.removeEventListener('mouseup', handleResizeEnd);
    
    // 移除调整中的视觉指示
    document.body.classList.remove('column-resizing');
    document.body.style.cursor = '';
    
    // 保存列宽到本地存储
    saveColumnWidths();
    
    // 触发一个自定义事件，通知表格组件列宽已更新
    if (columnId) {
      const event = new CustomEvent('column-resize-end', { 
        detail: { columnId, width: columnWidths.value[columnId] } 
      });
      document.dispatchEvent(event);
    }
  };

  // 初始化时从本地存储加载列宽
  onMounted(() => {
    // 先设置默认列宽
    columnWidths.value = { ...defaultWidths };
    
    // 尝试从本地存储加载
    const savedWidths = localStorage.getItem(`${tableId}-column-widths`);
    if (savedWidths) {
      try {
        columnWidths.value = { ...columnWidths.value, ...JSON.parse(savedWidths) };
      } catch(e) {
        console.error('Failed to parse saved column widths', e);
      }
    }
    console.log(`[${tableId}] Initial columnWidths after onMounted:`, JSON.parse(JSON.stringify(columnWidths.value)));
  });

  // 在组件销毁时清理
  onBeforeUnmount(() => {
    // 移除全局事件监听器（如果存在）
    if (isResizing.value) {
      document.removeEventListener('mousemove', handleResize);
      document.removeEventListener('mouseup', handleResizeEnd);
    }
  });

  return {
    columnWidths,
    isResizing,
    resizingColumnId,
    autoAdjustColumnWidth,
    handleResizeStart,
    saveColumnWidths
  };
} 