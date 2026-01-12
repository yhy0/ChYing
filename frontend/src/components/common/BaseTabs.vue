<script setup lang="ts">
import { ref, computed, toRef } from 'vue';
import { useTabsManagement, type BaseTab, type TabGroup } from '../../composables';

// 使用已定义的类型而不是创建新的接口
interface TabWithGroup extends BaseTab {
  groupId?: string | null;
  isRunning: boolean;
}

// 组件props类型定义
interface Props {
  tabs: TabWithGroup[];
  groups: TabGroup[];
  prefix: string;
  colorOptions: Array<{
    id: string;
    value: string;
    label: string;
  }>;
  borderColor?: string;
  textPrimary?: string;
  textMuted?: string;
  bgPrimary?: string;
  bgSecondary?: string;
  bgTertiary?: string;
  bgHover?: string;
  borderFocus?: string;
}

const props = defineProps<Props>();

const emit = defineEmits([
  'select-tab',
  'close-tab',
  'rename-tab',
  'change-tab-color',
  'change-tab-group',
  'create-group',
  'reorder-tabs',
  'reorder-groups',
]);

// 组件CSS变量
const cssVars = computed(() => ({
  '--tab-prefix': props.prefix,
  '--tab-border': props.borderColor || `var(--${props.prefix}-border, #e5e7eb)`,
  '--tab-text-primary': props.textPrimary || `var(--${props.prefix}-text-primary, #111827)`,
  '--tab-text-muted': props.textMuted || `var(--${props.prefix}-text-muted, #6b7280)`,
  '--tab-bg-primary': props.bgPrimary || `var(--${props.prefix}-bg-primary, #ffffff)`,
  '--tab-bg-secondary': props.bgSecondary || `var(--${props.prefix}-bg-secondary, #f3f4f6)`,
  '--tab-bg-tertiary': props.bgTertiary || `var(--${props.prefix}-bg-tertiary, #f3f4f6)`,
  '--tab-bg-hover': props.bgHover || `var(--${props.prefix}-bg-secondary-hover, #e5e7eb)`,
  '--tab-border-focus': props.borderFocus || `var(--${props.prefix}-btn-primary, #6366f1)`
}));

// 跟踪折叠的分组
const collapsedGroups = ref<Set<string>>(new Set());

// 切换分组折叠状态
const toggleGroupCollapse = (groupId: string, event: MouseEvent) => {
  event.stopPropagation(); // 防止触发拖拽
  if (collapsedGroups.value.has(groupId)) {
    collapsedGroups.value.delete(groupId);
  } else {
    collapsedGroups.value.add(groupId);
  }
};

// 使用tabs管理composable
const {
  // 状态
  editingTabId,
  tabNameInput,
  editingTabName,
  draggedTabId,
  draggedGroupId,
  dragOverTabId,
  dragOverGroupId,
  isDragging,
  isGroupDragging,
  
  // 菜单相关
  showContextMenu,
  
  // 重命名相关
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
} = useTabsManagement(
  props.tabs,
  toRef(props, 'groups'), // 将 props.groups 转换为响应式引用
  {
    enableGroups: true,
    classPrefix: props.prefix,
    colorOptions: props.colorOptions
  }
);

// 处理Tab点击
const handleTabClick = (tabId: string) => {
  emit('select-tab', tabId);
};

// 处理Tab关闭
const handleCloseTab = (event: MouseEvent, tabId: string) => {
  event.stopPropagation();
  emit('close-tab', tabId);
};

// 处理右键菜单显示
const handleContextMenu = (event: MouseEvent, tabId: string) => {
  showContextMenu(
    event, 
    tabId,
    () => emit('create-group'),
    (tabId, groupId) => emit('change-tab-group', tabId, groupId),
    (tabId, color) => emit('change-tab-color', tabId, color)
  );
};

// 处理Tab重命名完成
const completeRenaming = (tabId: string, newName: string) => {
  emit('rename-tab', tabId, newName);
};

// 处理按键事件 
const handleTabKeyDown = (event: KeyboardEvent) => {
  handleKeyDown(event, completeRenaming);
};

// 增强版的handleDragLeave，用于清除所有拖放样式
const clearDragStyles = (event: DragEvent) => {
  handleDragLeave(event);
  
  if (event.currentTarget instanceof HTMLElement) {
    event.currentTarget.classList.remove('group-dragover');
    event.currentTarget.classList.remove('receiving-from-group');
    event.currentTarget.classList.remove('receiving-from-other-group');
    event.currentTarget.classList.remove('droppable-hover');
    event.currentTarget.classList.remove('cross-group-target');
  }
};

// 优化后的Tab拖放逻辑
const handleTabDrop = (event: DragEvent, targetTab: any) => {
  // 检查是否跨分组拖拽
  if (draggedTabId.value && draggedTabId.value !== targetTab.id) {
    const draggedTab = props.tabs.find(tab => tab.id === draggedTabId.value);
    
    if (draggedTab && 'groupId' in draggedTab && 'groupId' in targetTab) {
      const sourceGroupId = (draggedTab as any).groupId;
      const targetGroupId = (targetTab as any).groupId;
      
      // 只处理跨分组的情况
      if (sourceGroupId !== targetGroupId) {
        emit('change-tab-group', draggedTabId.value, targetGroupId);
        
        // 清理拖拽样式
        if (event.currentTarget instanceof HTMLElement) {
          event.currentTarget.classList.remove('tab-dragover');
        }
        removeDragStyles('.tab-dragging');
        removeDragStyles('.tab-dragover');
        removeDragStyles('.receiving-from-other-group');
        
        // 重置状态
        draggedTabId.value = null;
        dragOverTabId.value = null;
        
        return;
      }
    }
  }
  
  // 常规拖放处理(同组内排序)
  handleDrop(
    event, 
    targetTab, 
    (tabs) => emit('reorder-tabs', tabs),
    (tabId, groupId) => emit('change-tab-group', tabId, groupId)
  );
};

// 处理Group拖放排序
const handleGroupDropEvent = (event: DragEvent, targetGroupId: string) => {
  handleGroupDrop(
    event, 
    targetGroupId, 
    (groups) => emit('reorder-groups', groups),
    (tabId, groupId) => emit('change-tab-group', tabId, groupId)
  );
};

// 处理无分组区域的拖拽进入
const handleNoGroupDragEnter = (event: DragEvent) => {
  event.preventDefault();
  
  if (isDragging.value && draggedTabId.value) {
    dragOverGroupId.value = null;
    
    const draggedTab = props.tabs.find(tab => tab.id === draggedTabId.value);
    if (!draggedTab) return;
    
    if (event.currentTarget instanceof HTMLElement) {
      event.currentTarget.classList.add('group-dragover');
      
      if ('groupId' in draggedTab && (draggedTab as any).groupId !== null) {
        event.currentTarget.classList.add('receiving-from-group');
      }
    }
  }
};

// 处理无分组区域的拖放
const handleNoGroupDrop = (event: DragEvent) => {
  // 使用工具函数处理标签从分组拖到无分组区域
  const handled = handleTabToNoGroup(
    event, 
    (tabId, groupId) => emit('change-tab-group', tabId, groupId)
  );
  
  if (handled) {
    if (event.currentTarget instanceof HTMLElement) {
      event.currentTarget.classList.remove('group-dragover');
      event.currentTarget.classList.remove('receiving-from-group');
    }
    return;
  }
  
  event.preventDefault();
  
  const type = event.dataTransfer?.getData('type');
  
  if (type === 'tab' && draggedTabId.value) {
    const draggedTab = props.tabs.find(tab => tab.id === draggedTabId.value);
    if (!draggedTab) return;
    
    if ('groupId' in draggedTab && (draggedTab as any).groupId !== null) {
      emit('change-tab-group', draggedTabId.value, null);
    }
    
    if (event.currentTarget instanceof HTMLElement) {
      event.currentTarget.classList.remove('group-dragover');
      event.currentTarget.classList.remove('receiving-from-group');
    }
  }
  
  // 重置状态
  draggedTabId.value = null;
  dragOverGroupId.value = null;
};

// 暴露给插槽使用的方法和数据
defineExpose({
  groupedTabs
});
</script>

<template>
  <div class="tabs-container" :style="cssVars">
    <!-- 无分组标签 -->
    <div 
      v-if="groupedTabs().noGroupTabs.length > 0" 
      class="tabs-group no-group"
      :class="{ 'group-dragover': dragOverGroupId === null && isDragging }"
      @dragenter.prevent="handleNoGroupDragEnter"
      @dragleave.prevent="clearDragStyles"
      @dragover.prevent="handleDragOver"
      @drop.prevent="handleNoGroupDrop"
    >
      <div class="tabs-row">
        <div
          v-for="(tab, index) in groupedTabs().noGroupTabs"
          :key="tab.id"
          :class="[
            'tab',
            { 'active': tab.isActive },
            { 'tab-dragging': isDragging && draggedTabId === tab.id },
            { 'tab-dragover': dragOverTabId === tab.id }
          ]"
          :style="{ 
            borderColor: tab.color,
            borderRight: shouldShowBorder(tab, index, groupedTabs().noGroupTabs) ? '1px solid var(--tab-border)' : 'none'
          }"
          draggable="true"
          @click="handleTabClick(tab.id)"
          @dblclick="handleDoubleClick(tab, completeRenaming)"
          @contextmenu.prevent="handleContextMenu($event, tab.id)"
          @dragstart="handleDragStart($event, tab)"
          @dragend="handleDragEnd"
          @dragenter="handleDragEnter($event, tab)"
          @dragleave="handleDragLeave"
          @dragover="handleDragOver"
          @drop="handleTabDrop($event, tab)"
        >
          <div class="tab-color" :style="{ backgroundColor: tab.color }"></div>
          <div v-if="editingTabId !== tab.id" class="tab-name">
            {{ tab.name }}
            <!-- 通过插槽允许自定义Tab内容 -->
            <slot name="tab-extra" :tab="tab"></slot>
          </div>
          <div v-else class="tab-edit">
            <input
              ref="tabNameInput"
              v-model="editingTabName"
              class="tab-input"
              type="text"
              @blur="finishRenaming(completeRenaming)"
              @keydown="handleTabKeyDown"
            />
          </div>
          <button 
            class="tab-close" 
            @click="handleCloseTab($event, tab.id)"
          >
            <i class="bx bx-x"></i>
          </button>
        </div>
      </div>
    </div>

    <!-- 有分组标签容器 -->
    <div v-if="groupedTabs().groupedTabs.length > 0" class="tabs-groups-container">
      <div 
        v-for="groupData in groupedTabs().groupedTabs" 
        :key="groupData.group.id"
        class="tabs-group"
        :class="{ 
          'group-dragging': isGroupDragging && draggedGroupId === groupData.group.id,
          'group-dragover': dragOverGroupId === groupData.group.id,
          'group-collapsed': collapsedGroups.has(groupData.group.id)
        }"
        draggable="true"
        @dragstart="handleGroupDragStart($event, groupData.group.id)"
        @dragend="handleDragEnd"
        @dragenter="handleGroupDragEnter($event, groupData.group.id)"
        @dragleave="clearDragStyles"
        @dragover.prevent="handleDragOver"
        @drop="handleGroupDropEvent($event, groupData.group.id)"
      >
        <div 
          class="group-header" 
          :style="{ borderColor: groupData.group.color }"
          @dragenter.stop="handleGroupDragEnter($event, groupData.group.id)"
          @dragleave.stop="clearDragStyles"
          @dragover.stop.prevent="handleDragOver"
          @drop.stop="handleGroupDropEvent($event, groupData.group.id)"
        >
          <div class="group-color" :style="{ backgroundColor: groupData.group.color }"></div>
          <div class="group-name">{{ groupData.group.name }}</div>
          <button 
            class="group-toggle"
            @click="toggleGroupCollapse(groupData.group.id, $event)" 
            :title="collapsedGroups.has(groupData.group.id) ? '展开分组' : '折叠分组'"
          >
            <i :class="['bx', collapsedGroups.has(groupData.group.id) ? 'bx-chevron-right' : 'bx-chevron-down']"></i>
          </button>
        </div>
        <div class="tabs-row" v-show="!collapsedGroups.has(groupData.group.id)">
          <div
            v-for="(tab, index) in groupData.tabs"
            :key="tab.id"
            :class="[
              'tab',
              { 'active': tab.isActive },
              { 'tab-dragging': isDragging && draggedTabId === tab.id },
              { 'tab-dragover': dragOverTabId === tab.id }
            ]"
            :style="{ 
              borderColor: tab.color,
              borderRight: shouldShowBorder(tab, index, groupData.tabs) ? '1px solid var(--tab-border)' : 'none'
            }"
            draggable="true"
            @click="handleTabClick(tab.id)"
            @dblclick="handleDoubleClick(tab, completeRenaming)"
            @contextmenu.prevent="handleContextMenu($event, tab.id)"
            @dragstart="handleDragStart($event, tab)"
            @dragend="handleDragEnd"
            @dragenter="handleDragEnter($event, tab)"
            @dragleave="handleDragLeave"
            @dragover="handleDragOver"
            @drop="handleTabDrop($event, tab)"
            :data-group-id="(tab as any).groupId"
          >
            <div class="tab-color" :style="{ backgroundColor: tab.color }"></div>
            <div v-if="editingTabId !== tab.id" class="tab-name">
              {{ tab.name }}
              <!-- 通过插槽允许自定义Tab内容 -->
              <slot name="tab-extra" :tab="tab"></slot>
            </div>
            <div v-else class="tab-edit">
              <input
                ref="tabNameInput"
                v-model="editingTabName"
                class="tab-input"
                type="text"
                @blur="finishRenaming(completeRenaming)"
                @keydown="handleTabKeyDown"
              />
            </div>
            <button 
              class="tab-close" 
              @click="handleCloseTab($event, tab.id)"
            >
              <i class="bx bx-x"></i>
            </button>
          </div>
        </div>
        <!-- 折叠状态时显示标签计数 -->
        <div 
          v-if="collapsedGroups.has(groupData.group.id)" 
          class="group-collapsed-info"
        >
          <span class="group-tab-count">{{ groupData.tabs.length }} 个标签</span>
          <slot name="group-collapsed-indicator" :group="groupData.group" :tabs="groupData.tabs"></slot>
        </div>
      </div>
    </div>
  </div>
</template> 