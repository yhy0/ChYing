import { ref, nextTick, toRef, type Ref } from 'vue';

// 通用Tab类型，包含Repeater和Intruder标签页的共同属性
export interface BaseTab {
  id: string;
  name: string;
  color: string;
  isActive: boolean;
}

// 通用的分组接口
export interface TabGroup {
  id: string;
  name: string;
  color: string;
}

// 配置选项接口
export interface TabsManagementOptions<T extends BaseTab> {
  // 是否启用分组功能
  enableGroups?: boolean;
  // 样式前缀，用于CSS类名
  classPrefix: string;
  // 预定义的颜色选项
  colorOptions: Array<{
    id: string;
    value: string;
    label: string;
  }>;
}

// 默认配置
const defaultOptions: Partial<TabsManagementOptions<BaseTab>> = {
  enableGroups: false,
  classPrefix: 'tab',
  colorOptions: [
    { id: 'default', value: '#4f46e5', label: 'Default (Purple)' },
    { id: 'red', value: '#ef4444', label: 'Red' },
    { id: 'green', value: '#10b981', label: 'Green' },
    { id: 'blue', value: '#3b82f6', label: 'Blue' },
    { id: 'yellow', value: '#f59e0b', label: 'Yellow' },
    { id: 'orange', value: '#f97316', label: 'Orange' },
    { id: 'teal', value: '#14b8a6', label: 'Teal' },
  ]
};

// 菜单相关常量
const MENU_WIDTH = 190;
const MENU_HEIGHT = 400;
const SAFETY_MARGIN = 20;
const MAX_GROUP_NAME_LENGTH = 12;
const Z_INDEX = 10000;

export function useTabsManagement<T extends BaseTab, G extends TabGroup = TabGroup>(
  tabs: T[],
  groups: Ref<G[]> | G[],
  options: TabsManagementOptions<T>
) {
  // 合并选项
  const opts = { ...defaultOptions, ...options };
  const classPrefix = opts.classPrefix;

  // 确保 groups 是响应式的
  const reactiveGroups = Array.isArray(groups) ? ref(groups) : groups;
  
  // ===== 标签编辑状态 =====
  const editingTabId = ref<string | null>(null);
  const tabNameInput = ref<HTMLInputElement | null>(null);
  const editingTabName = ref('');

  // ===== 右键菜单状态 =====
  const menuVisible = ref(false);
  const activeTabId = ref<string | null>(null);
  const menuElement = ref<HTMLElement | null>(null);

  // ===== 拖动相关状态 =====
  const draggedTabId = ref<string | null>(null);
  const draggedGroupId = ref<string | null>(null);
  const dragOverTabId = ref<string | null>(null);
  const dragOverGroupId = ref<string | null>(null);
  const isDragging = ref(false);
  const isGroupDragging = ref(false);
  const dragStartGroup = ref<string | null>(null);

  // ===== 菜单样式 =====
  const getMenuStyles = () => `
    .menu-section {
      padding: 0.25rem 0;
    }
    .menu-header {
      padding: 0.25rem 0.5rem;
      font-size: 0.75rem;
      font-weight: 500;
      color: var(--${classPrefix}-text-muted, #6b7280);
    }
    .menu-item {
      padding: 0.25rem 0.5rem;
      display: flex;
      align-items: center;
      font-size: 0.75rem;
      color: var(--${classPrefix}-text-primary, #111827);
      cursor: pointer;
      white-space: nowrap;
      overflow: hidden;
      text-overflow: ellipsis;
    }
    .menu-item:hover {
      background-color: var(--${classPrefix}-bg-secondary, #f3f4f6);
    }
    .menu-divider {
      height: 1px;
      background-color: var(--${classPrefix}-border, #e5e7eb);
      margin: 0.25rem 0;
    }
    .color-dot {
      width: 8px;
      height: 8px;
      border-radius: 50%;
      margin-right: 0.3rem;
      border: 1px solid var(--${classPrefix}-border, #e5e7eb);
      flex-shrink: 0;
    }
    .menu-icon {
      margin-right: 0.3rem;
      font-size: 0.9rem;
      width: 0.9rem;
      text-align: center;
      flex-shrink: 0;
    }
    .menu-text {
      white-space: nowrap;
      overflow: hidden;
      text-overflow: ellipsis;
      font-size: 0.7rem;
    }
  `;

  // ===== 事件监听器管理 =====
  const setupMenuEventListeners = () => {
    document.addEventListener('click', closeContextMenu);
    document.addEventListener('keydown', handleMenuKeyDown);
  };

  const removeMenuEventListeners = () => {
    document.removeEventListener('click', closeContextMenu);
    document.removeEventListener('keydown', handleMenuKeyDown);
  };

  // ===== 菜单位置计算 =====
  const calculateMenuPosition = (event: MouseEvent) => {
    const viewportWidth = window.innerWidth;
    const viewportHeight = window.innerHeight;
    
    let x = event.clientX;
    let y = event.clientY;
    
    // 检查水平空间
    if (x + MENU_WIDTH + SAFETY_MARGIN > viewportWidth) {
      x = Math.max(SAFETY_MARGIN, x - MENU_WIDTH);
    }
    
    // 检查垂直空间
    if (y + MENU_HEIGHT + SAFETY_MARGIN > viewportHeight) {
      y = Math.max(SAFETY_MARGIN, viewportHeight - MENU_HEIGHT - SAFETY_MARGIN);
    }
    
    return { x, y };
  };

  // ===== 创建和移除菜单元素 =====
  const createMenuElement = (x: number, y: number) => {
    const menu = document.createElement('div');
    menu.className = 'analysis-menu-container';
    menu.style.position = 'fixed';
    menu.style.zIndex = Z_INDEX.toString();
    menu.style.top = '0';
    menu.style.left = '0';
    menu.style.width = '100vw';
    menu.style.height = '100vh';
    menu.style.pointerEvents = 'none';
    
    const viewportHeight = window.innerHeight;
    const actualMenuHeight = Math.min(MENU_HEIGHT, viewportHeight - 20);
    
    menu.innerHTML = `
      <div class="context-menu" style="
        position: absolute;
        top: ${y}px;
        left: ${x}px;
        background-color: var(--${classPrefix}-bg-primary, #ffffff);
        border: 1px solid var(--${classPrefix}-border, #e5e7eb);
        border-radius: 0.375rem;
        box-shadow: 0 4px 12px -1px rgba(0, 0, 0, 0.2), 0 2px 6px -1px rgba(0, 0, 0, 0.1);
        width: ${MENU_WIDTH - 20}px;
        max-height: ${actualMenuHeight}px;
        overflow-y: auto;
        pointer-events: auto;
      ">
      </div>
    `;
    
    return menu;
  };

  const removeMenuElement = () => {
    if (menuElement.value) {
      try {
        document.body.removeChild(menuElement.value);
      } catch (e) {
        console.error('Menu already removed');
      }
      menuElement.value = null;
    }
  };

  // ===== 处理菜单键盘事件 =====
  const handleMenuKeyDown = (event: KeyboardEvent) => {
    if (event.key === 'Escape') {
      closeContextMenu();
    }
  };

  // ===== 关闭右键菜单 =====
  const closeContextMenu = () => {
    removeMenuElement();
    menuVisible.value = false;
    removeMenuEventListeners();
  };

  // ===== HTML生成函数 =====
  const generateGroupItem = (group: G) => {
    const displayName = group.name.length > MAX_GROUP_NAME_LENGTH 
      ? `${group.name.substring(0, MAX_GROUP_NAME_LENGTH)}...` 
      : group.name;
    
    return `
      <div class="menu-item change-group" data-group="${group.id}">
        <i class="bx bx-folder menu-icon" style="color: var(--color-text-tertiary, #6b7280);"></i>
        <div class="color-dot" style="background-color: ${group.color};"></div>
        <span class="menu-text" title="${group.name}">${displayName}</span>
      </div>
    `;
  };

  const generateColorItem = (color: { id: string, value: string, label: string }) => {
    return `
      <div class="menu-item change-color" data-color="${color.value}">
        <div class="color-dot" style="background-color: ${color.value};"></div>
        <span class="menu-text">${color.label}</span>
      </div>
    `;
  };

  // ===== 生成菜单内容 =====
  const generateMenuContent = (
    emitCreateGroup: (() => void) | null = null,
    emitChangeTabGroup: ((tabId: string, groupId: string | null) => void) | null = null,
    emitChangeTabColor: (tabId: string, color: string) => void
  ) => {
    let html = `<style>${getMenuStyles()}</style>`;

    // 分组部分 - 只有当启用分组时才显示
    if (opts.enableGroups && emitChangeTabGroup && emitCreateGroup) {
      html += `
        <div class="menu-section">
          <div class="menu-header">分组</div>
          <div class="menu-item create-group">
            <i class="bx bx-folder-plus menu-icon" style="color: var(--color-primary, #4f46e5);"></i>
            <span class="menu-text">创建新分组</span>
          </div>
          <div class="menu-divider"></div>
          <div class="menu-item change-group" data-group="">
            <i class="bx bx-x-circle menu-icon" style="color: var(--color-text-tertiary, #6b7280);"></i>
            <span class="menu-text">无分组</span>
          </div>
      `;

      // 添加所有分组 - 使用响应式的 groups 数据
      reactiveGroups.value.forEach(group => {
        html += generateGroupItem(group);
      });
      
      html += `</div>`;
    }
    
    // 颜色部分 - 始终显示
    if (!opts.enableGroups) {
      html += `<div class="menu-section">`;  
    } else {
      html += `
        <div class="menu-divider"></div>
        <div class="menu-section">
      `;
    }
    html += `<div class="menu-header">颜色</div>`;
    
    // 添加所有颜色选项
    opts.colorOptions.forEach(color => {
      html += generateColorItem(color);
    });
    
    html += `</div>`;
    
    return html;
  };

  // ===== 添加菜单项的事件监听 =====
  const addMenuEventListeners = (
    menuElement: HTMLElement,
    emitCreateGroup: (() => void) | null = null,
    emitChangeTabGroup: ((tabId: string, groupId: string | null) => void) | null = null,
    emitChangeTabColor: (tabId: string, color: string) => void
  ) => {
    // 阻止菜单内点击事件冒泡，避免关闭菜单
    menuElement.addEventListener('click', (event) => {
      event.stopPropagation();
    });
    
    // 创建新分组 - 仅当启用分组时
    if (opts.enableGroups && emitCreateGroup) {
      const createGroupButton = menuElement.querySelector('.menu-item.create-group');
      if (createGroupButton) {
        createGroupButton.addEventListener('click', () => {
          emitCreateGroup();
          closeContextMenu();
        });
      }
    }
    
    // 更改分组 - 仅当启用分组时
    if (opts.enableGroups && emitChangeTabGroup) {
      const changeGroupButtons = menuElement.querySelectorAll('.menu-item.change-group');
      changeGroupButtons.forEach(button => {
        button.addEventListener('click', () => {
          const groupId = (button as HTMLElement).dataset.group || null;
          if (activeTabId.value) {
            emitChangeTabGroup(activeTabId.value, groupId);
            closeContextMenu();
          }
        });
      });
    }
    
    // 更改颜色 - 始终启用
    const changeColorButtons = menuElement.querySelectorAll('.menu-item.change-color');
    changeColorButtons.forEach(button => {
      button.addEventListener('click', () => {
        const color = (button as HTMLElement).dataset.color;
        if (activeTabId.value && color) {
          emitChangeTabColor(activeTabId.value, color);
          closeContextMenu();
        }
      });
    });
  };

  // ===== 显示右键菜单 =====
  const showContextMenu = (
    event: MouseEvent, 
    tabId: string,
    emitCreateGroup: (() => void) | null = null,
    emitChangeTabGroup: ((tabId: string, groupId: string | null) => void) | null = null,
    emitChangeTabColor: (tabId: string, color: string) => void
  ) => {
    event.preventDefault();
    
    // 关闭之前可能打开的菜单
    closeContextMenu();
    
    // 设置活动标签ID
    activeTabId.value = tabId;
    
    // 等待下一个DOM更新周期创建菜单
    nextTick(() => {
      // 计算菜单位置
      const { x, y } = calculateMenuPosition(event);
      
      // 创建菜单元素
      const menu = createMenuElement(x, y);
      
      // 添加到body
      document.body.appendChild(menu);
      menuElement.value = menu;
      
      // 获取创建的菜单内容元素
      const menuContent = menu.querySelector('.context-menu') as HTMLElement;
      if (!menuContent) {
        console.error('无法找到菜单内容元素');
        return;
      }
      
      // 填充菜单内容
      const menuHTML = generateMenuContent(emitCreateGroup, emitChangeTabGroup, emitChangeTabColor);
      menuContent.innerHTML = menuHTML;
      
      // 添加事件处理
      addMenuEventListeners(menuContent, emitCreateGroup, emitChangeTabGroup, emitChangeTabColor);
      
      // 添加事件监听器
      setupMenuEventListeners();
      
      // 标记菜单为可见
      menuVisible.value = true;
    });
  };

  // ===== 标签重命名相关函数 =====
  const startRenaming = (tabId: string, currentName: string) => {
    editingTabId.value = tabId;
    editingTabName.value = currentName;
    
    // Focus the input after DOM update
    nextTick(() => {
      if (tabNameInput.value) {
        tabNameInput.value.focus();
        tabNameInput.value.select();
      }
    });
  };

  const finishRenaming = (emitRenameTab: (tabId: string, newName: string) => void) => {
    if (editingTabId.value) {
      emitRenameTab(editingTabId.value, editingTabName.value);
      editingTabId.value = null;
    }
  };

  const handleKeyDown = (event: KeyboardEvent, emitRenameTab: (tabId: string, newName: string) => void) => {
    if (event.key === 'Enter') {
      finishRenaming(emitRenameTab);
    } else if (event.key === 'Escape') {
      editingTabId.value = null;
    }
  };

  // ===== Tab分组和边框逻辑 =====
  const groupedTabs = () => {
    if (!opts.enableGroups) {
      return { noGroupTabs: tabs, groupedTabs: [] };
    }
    
    const noGroupTabs = tabs.filter(tab => {
      // 检查tab是否有groupId属性，且值为null
      return !('groupId' in tab) || (tab as any).groupId === null;
    });
    
    const groupedTabs = reactiveGroups.value.map(group => ({
      group,
      tabs: tabs.filter(tab => 'groupId' in tab && (tab as any).groupId === group.id)
    })).filter(group => group.tabs.length > 0);

    return {
      noGroupTabs,
      groupedTabs
    };
  };

  // 修改handleDoubleClick函数，使它接受一个可选的回调参数
  const handleDoubleClick = (tab: T, emitRenameTab?: (tabId: string, newName: string) => void) => {
    startRenaming(tab.id, tab.name);
  };

  const shouldShowBorder = (tab: T, index: number, groupTabs: T[]) => {
    if (index === groupTabs.length - 1) return false;
    if ((index + 1) % 5 === 0) return false;
    
    const nextTab = groupTabs[index + 1];
    return !nextTab.isActive;
  };

  // ===== 拖拽相关函数 =====
  const removeDragStyles = (selector: string) => {
    document.querySelectorAll(selector).forEach(el => {
      el.classList.remove(selector.substring(1));
    });
  };

  const handleDragStart = (event: DragEvent, tab: T) => {
    if (event.dataTransfer) {
      event.dataTransfer.effectAllowed = 'move';
      event.dataTransfer.setData('text/plain', tab.id);
      event.dataTransfer.setData('type', 'tab');
      // 如果tab属于分组，添加分组信息
      if ('groupId' in tab) {
        event.dataTransfer.setData('sourceGroupId', (tab as any).groupId || 'null');
      }
      
      draggedTabId.value = tab.id;
      dragStartGroup.value = 'groupId' in tab ? (tab as any).groupId : null;
      isDragging.value = true;
      
      const element = event.target as HTMLElement;
      if (element) {
        setTimeout(() => {
          element.classList.add(`${classPrefix}-tab-dragging`);
        }, 0);
      }
    }
  };

  // 工具函数：处理标签从分组拖到无分组区域
  const handleTabToNoGroup = (
    event: DragEvent,
    emitChangeTabGroup: (tabId: string, groupId: string | null) => void
  ) => {
    event.preventDefault();
    
    // 检查是否是标签拖放
    if (!draggedTabId.value) return false;
    
    // 获取被拖拽的标签
    const draggedTab = tabs.find(tab => tab.id === draggedTabId.value);
    if (!draggedTab) return false;
    
    // 如果标签有分组，则将其移动到无分组
    if ('groupId' in draggedTab && (draggedTab as any).groupId !== null) {
      // 触发分组更改
      emitChangeTabGroup(draggedTabId.value, null);
      
      // 重置状态
      draggedTabId.value = null;
      dragOverTabId.value = null;
      dragOverGroupId.value = null;
      
      // 移除所有拖放样式
      removeDragStyles(`.${classPrefix}-tab-dragging`);
      removeDragStyles(`.${classPrefix}-tab-dragover`);
      
      return true;
    }
    
    return false;
  };

  const handleGroupDragStart = (event: DragEvent, groupId: string) => {
    if (!opts.enableGroups) return;
    
    if (event.dataTransfer) {
      event.dataTransfer.effectAllowed = 'move';
      event.dataTransfer.setData('text/plain', groupId);
      event.dataTransfer.setData('type', 'group');
      draggedGroupId.value = groupId;
      isGroupDragging.value = true;
      
      const element = event.target as HTMLElement;
      if (element) {
        setTimeout(() => {
          element.classList.add(`${classPrefix}-group-dragging`);
        }, 0);
      }
    }
  };

  const handleDragEnd = (event: DragEvent) => {
    isDragging.value = false;
    isGroupDragging.value = false;
    draggedTabId.value = null;
    draggedGroupId.value = null;
    dragOverTabId.value = null;
    dragOverGroupId.value = null;
    dragStartGroup.value = null;
    
    // 移除所有拖动样式
    removeDragStyles(`.${classPrefix}-tab-dragging`);
    removeDragStyles(`.${classPrefix}-tab-dragover`);
    removeDragStyles(`.${classPrefix}-group-dragging`);
    removeDragStyles(`.${classPrefix}-group-dragover`);
    removeDragStyles(`.receiving-from-group`);
    removeDragStyles(`.receiving-from-other-group`);
    removeDragStyles(`.cross-group-target`);
    removeDragStyles(`.droppable-hover`);
  };

  const handleDragEnter = (event: DragEvent, tab: T) => {
    if (draggedTabId.value === null || draggedTabId.value === tab.id) {
      return;
    }
    
    dragOverTabId.value = tab.id;
    
    if (event.currentTarget instanceof HTMLElement) {
      event.currentTarget.classList.add(`${classPrefix}-tab-dragover`);
      
      // 检查是否是跨分组拖拽
      if (opts.enableGroups && 'groupId' in tab) {
        const targetGroupId = (tab as any).groupId;
        
        // 找到被拖拽的标签
        const draggedTab = tabs.find(t => t.id === draggedTabId.value);
        if (draggedTab && 'groupId' in draggedTab) {
          const sourceGroupId = (draggedTab as any).groupId;
          
          // 如果是跨分组拖拽，添加特殊样式
          if (sourceGroupId !== targetGroupId) {
            event.currentTarget.classList.add('cross-group-target');
          }
        }
      }
    }
  };

  const handleGroupDragEnter = (event: DragEvent, groupId: string) => {
    if (!opts.enableGroups) return;
    
    if (draggedGroupId.value !== groupId) {
      dragOverGroupId.value = groupId;
      
      if (event.currentTarget instanceof HTMLElement) {
        event.currentTarget.classList.add(`${classPrefix}-group-dragover`);
        
        // 如果是标签拖拽到分组上，添加特殊样式
        if (isDragging.value && draggedTabId.value) {
          const draggedTab = tabs.find(tab => tab.id === draggedTabId.value);
          if (draggedTab && 'groupId' in draggedTab) {
            const sourceGroupId = (draggedTab as any).groupId;
            // 如果标签来自不同分组，添加特殊标记
            if (sourceGroupId !== null && sourceGroupId !== groupId) {
              event.currentTarget.classList.add('receiving-from-other-group');
              
              // 如果当前元素是分组标题，添加额外的视觉效果
              if (event.currentTarget.classList.contains('group-header')) {
                event.currentTarget.classList.add('droppable-hover');
              }
            }
          }
        }
      }
    }
  };

  const handleDragLeave = (event: DragEvent) => {
    if (event.currentTarget instanceof HTMLElement) {
      event.currentTarget.classList.remove(`${classPrefix}-tab-dragover`);
      event.currentTarget.classList.remove(`${classPrefix}-group-dragover`);
      event.currentTarget.classList.remove('receiving-from-other-group');
      event.currentTarget.classList.remove('droppable-hover');
      event.currentTarget.classList.remove('cross-group-target');
    }
  };

  const handleDragOver = (event: DragEvent) => {
    event.preventDefault();
  };

  const handleDrop = (event: DragEvent, targetTab: T, emitReorderTabs: (tabs: T[]) => void, emitChangeTabGroup?: (tabId: string, groupId: string | null) => void) => {
    event.preventDefault();
    
    const type = event.dataTransfer?.getData('type');
    
    if (type === 'tab') {
      if (!draggedTabId.value || draggedTabId.value === targetTab.id) {
        return;
      }
      
      // 移除所有拖放样式
      removeDragStyles(`.${classPrefix}-tab-dragging`);
      removeDragStyles(`.${classPrefix}-tab-dragover`);
      
      // 获取源和目标标签
      const draggedTab = tabs.find(tab => tab.id === draggedTabId.value);
      if (!draggedTab) return;
      
      // 检查是否涉及跨分组操作
      if (opts.enableGroups && 'groupId' in draggedTab && 'groupId' in targetTab) {
        const draggedGroupId = (draggedTab as any).groupId;
        const targetGroupId = (targetTab as any).groupId;
        
        // 如果分组不同，且提供了改变分组的回调，则执行跨分组拖放
        if (draggedGroupId !== targetGroupId && emitChangeTabGroup) {
          // 先更改分组
          emitChangeTabGroup(draggedTabId.value, targetGroupId);
          
          // 创建新的标签数组并重新排序 - 确保在分组变更后调整顺序
          const newTabs = [...tabs];
          const draggedIndex = newTabs.findIndex(tab => tab.id === draggedTabId.value);
          const targetIndex = newTabs.findIndex(tab => tab.id === targetTab.id);
          
          if (draggedIndex !== -1 && targetIndex !== -1) {
            // 从数组中移除拖动的标签
            const [removed] = newTabs.splice(draggedIndex, 1);
            // 更新分组信息以确保顺序正确
            (removed as any).groupId = targetGroupId;
            // 插入到目标位置
            newTabs.splice(targetIndex, 0, removed);
            
            // 发出重新排序事件
            emitReorderTabs(newTabs);
          }
          
          // 重置状态
          draggedTabId.value = null;
          dragOverTabId.value = null;
          return;
        }
      }
      
      // 如果是同分组排序或未启用分组，执行常规排序
      // 创建新的标签数组并重新排序
      const newTabs = [...tabs];
      const draggedIndex = newTabs.findIndex(tab => tab.id === draggedTabId.value);
      const targetIndex = newTabs.findIndex(tab => tab.id === targetTab.id);
      
      if (draggedIndex !== -1 && targetIndex !== -1) {
        // 从数组中移除拖动的标签
        const [removed] = newTabs.splice(draggedIndex, 1);
        // 插入到目标位置
        newTabs.splice(targetIndex, 0, removed);
        
        // 发出重新排序事件
        emitReorderTabs(newTabs);
      }
    }
    
    // 重置状态
    draggedTabId.value = null;
    dragOverTabId.value = null;
  };

  const handleGroupDrop = (event: DragEvent, targetGroupId: string, emitReorderGroups: (groups: G[]) => void, emitChangeTabGroup?: (tabId: string, groupId: string | null) => void) => {
    if (!opts.enableGroups) return;
    
    event.preventDefault();
    
    const type = event.dataTransfer?.getData('type');
    
    // 处理标签拖到分组的情况
    if (type === 'tab' && emitChangeTabGroup && draggedTabId.value) {
      // 获取被拖拽的标签
      const draggedTab = tabs.find(tab => tab.id === draggedTabId.value);
      if (!draggedTab) return;
      
      // 如果标签没有分组或分组不同，则更改分组
      if ('groupId' in draggedTab) {
        const sourceGroupId = (draggedTab as any).groupId;
        // 只有当源分组和目标分组不同时才更改
        if (sourceGroupId !== targetGroupId) {
          // 触发分组更改
          emitChangeTabGroup(draggedTabId.value, targetGroupId);
          
          // 重置状态
          draggedTabId.value = null;
          dragOverTabId.value = null;
          dragOverGroupId.value = null;
          
          // 移除所有拖放样式
          removeDragStyles(`.${classPrefix}-tab-dragging`);
          removeDragStyles(`.${classPrefix}-tab-dragover`);
          removeDragStyles(`.receiving-from-other-group`);
          
          // 移除当前元素上的样式
          if (event.currentTarget instanceof HTMLElement) {
            event.currentTarget.classList.remove(`${classPrefix}-group-dragover`);
            event.currentTarget.classList.remove('receiving-from-other-group');
          }
          
          return;
        }
      }
    }
    
    // 处理分组拖到分组的情况
    if (type === 'group') {
      if (!draggedGroupId.value || draggedGroupId.value === targetGroupId) {
        return;
      }
      
      // 移除所有拖放样式
      removeDragStyles(`.${classPrefix}-group-dragging`);
      removeDragStyles(`.${classPrefix}-group-dragover`);
      
      // 创建新的分组数组并重新排序
      const newGroups = [...reactiveGroups.value];
      const draggedIndex = newGroups.findIndex(group => group.id === draggedGroupId.value);
      const targetIndex = newGroups.findIndex(group => group.id === targetGroupId);
      
      if (draggedIndex !== -1 && targetIndex !== -1) {
        // 从数组中移除拖动的分组
        const [removed] = newGroups.splice(draggedIndex, 1);
        // 插入到目标位置
        newGroups.splice(targetIndex, 0, removed);
        
        // 发出重新排序事件
        emitReorderGroups(newGroups as G[]);
      }
    }
    
    // 重置状态
    draggedGroupId.value = null;
    dragOverGroupId.value = null;
  };

  // 返回所有需要的状态和方法
  return {
    // 状态
    editingTabId,
    tabNameInput,
    editingTabName,
    menuVisible,
    activeTabId,
    draggedTabId,
    draggedGroupId,
    dragOverTabId,
    dragOverGroupId,
    isDragging,
    isGroupDragging,
    
    // 菜单相关
    showContextMenu,
    closeContextMenu,
    
    // 重命名相关
    startRenaming,
    finishRenaming,
    handleKeyDown,
    
    // 分组和布局相关
    groupedTabs,
    handleDoubleClick,
    shouldShowBorder,
    
    // 拖拽相关
    handleDragStart,
    handleGroupDragStart,
    handleDragEnd,
    handleDragEnter,
    handleGroupDragEnter,
    handleDragLeave,
    handleDragOver,
    handleDrop,
    handleGroupDrop,
    handleTabToNoGroup,
    removeDragStyles
  };
} 