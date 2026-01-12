<script setup lang="ts">
import { ref, nextTick } from 'vue';
import { useModulesStore } from '../../store';
import type { DecoderTab } from '../../types';
import { useI18n } from 'vue-i18n';

const store = useModulesStore();
const { t } = useI18n();

// 重命名相关状态
const editingTabId = ref<string | null>(null);
const editingTabName = ref('');
const tabNameInput = ref<HTMLInputElement | null>(null);

// 拖动相关状态
const draggedTabId = ref<string | null>(null);
const dragOverTabId = ref<string | null>(null);
const isDragging = ref(false);

// 处理标签选择
const handleTabSelect = (tabId: string) => {
  store.setActiveDecoderTab(tabId);
};

// 处理标签关闭
const handleTabClose = (tabId: string, event: Event) => {
  event.stopPropagation(); // 防止触发选择事件
  store.closeDecoderTab(tabId);
};

// 开始重命名
const startRenaming = (tab: DecoderTab, event: Event) => {
  event.stopPropagation(); // 防止触发选择事件
  editingTabId.value = tab.id;
  editingTabName.value = tab.name;
  
  // 在下一个周期聚焦输入框
  nextTick(() => {
    if (tabNameInput.value) {
      tabNameInput.value.focus();
      tabNameInput.value.select();
    }
  });
};

// 完成重命名
const finishRenaming = () => {
  if (editingTabId.value && editingTabName.value.trim()) {
    store.updateDecoderTab(editingTabId.value, { name: editingTabName.value.trim() });
  }
  editingTabId.value = null;
};

// 处理按键事件
const handleKeyDown = (event: KeyboardEvent) => {
  if (event.key === 'Enter') {
    finishRenaming();
  } else if (event.key === 'Escape') {
    editingTabId.value = null;
  }
};

// 双击重命名
const handleDoubleClick = (tab: DecoderTab) => {
  startRenaming(tab, new Event('dblclick'));
};

// 判断是否显示边框
const shouldShowBorder = (index: number, tabs: DecoderTab[]) => {
  // 如果是最后一个标签或者下一个标签是活跃的，不显示边框
  if (index === tabs.length - 1) return false;
  // 优化多行情况，每行最后一个标签也不显示边框
  if ((index + 1) % 5 === 0) return false;
  
  const nextTab = tabs[index + 1];
  return !nextTab.isActive;
};

// 开始拖动标签
const handleDragStart = (event: DragEvent, tab: DecoderTab) => {
  if (event.dataTransfer) {
    event.dataTransfer.effectAllowed = 'move';
    event.dataTransfer.setData('text/plain', tab.id);
    draggedTabId.value = tab.id;
    isDragging.value = true;
    
    // 设置拖动视觉效果
    const element = event.target as HTMLElement;
    if (element) {
      setTimeout(() => {
        element.classList.add('decoder-tab-dragging');
      }, 0);
    }
  }
};

// 拖动结束
const handleDragEnd = () => {
  isDragging.value = false;
  draggedTabId.value = null;
  dragOverTabId.value = null;
  
  // 移除拖动样式
  document.querySelectorAll('.decoder-tab-dragging').forEach(el => {
    el.classList.remove('decoder-tab-dragging');
  });
  document.querySelectorAll('.decoder-tab-dragover').forEach(el => {
    el.classList.remove('decoder-tab-dragover');
  });
};

// 拖动进入目标区域
const handleDragEnter = (event: DragEvent, tab: DecoderTab) => {
  if (draggedTabId.value !== tab.id) {
    dragOverTabId.value = tab.id;
    
    // 添加视觉反馈
    if (event.currentTarget instanceof HTMLElement) {
      event.currentTarget.classList.add('decoder-tab-dragover');
    }
  }
};

// 拖动离开目标区域
const handleDragLeave = (event: DragEvent) => {
  // 移除视觉反馈
  if (event.currentTarget instanceof HTMLElement) {
    event.currentTarget.classList.remove('decoder-tab-dragover');
  }
};

// 允许放置
const handleDragOver = (event: DragEvent) => {
  event.preventDefault();
};

// 处理放置
const handleDrop = (event: DragEvent, targetTab: DecoderTab) => {
  event.preventDefault();
  
  if (!draggedTabId.value || draggedTabId.value === targetTab.id) {
    return;
  }
  
  // 移除拖放样式
  document.querySelectorAll('.decoder-tab-dragging').forEach(el => {
    el.classList.remove('decoder-tab-dragging');
  });
  document.querySelectorAll('.decoder-tab-dragover').forEach(el => {
    el.classList.remove('decoder-tab-dragover');
  });
  
  // 重新排序标签
  const decoderTabs = [...store.decoderTabs];
  const draggedIndex = decoderTabs.findIndex(tab => tab.id === draggedTabId.value);
  const targetIndex = decoderTabs.findIndex(tab => tab.id === targetTab.id);
  
  if (draggedIndex !== -1 && targetIndex !== -1) {
    // 从数组中移除拖动的标签
    const [removed] = decoderTabs.splice(draggedIndex, 1);
    // 插入到目标位置
    decoderTabs.splice(targetIndex, 0, removed);
    
    // 更新store
    store.decoderTabs = decoderTabs;
  }
  
  // 重置状态
  draggedTabId.value = null;
  dragOverTabId.value = null;
};
</script>

<template>
  <div class="decoder-tabs-container">
    <div class="flex flex-wrap items-center py-1 min-h-[32px]">
      <!-- 标签列表 -->
      <div class="flex flex-wrap">
        <div 
          v-for="(tab, index) in store.decoderTabs" 
          :key="tab.id"
          class="h-7 flex items-center relative my-0.5"
          :class="{ 
            'border-r border-gray-200 dark:border-gray-700': shouldShowBorder(index, store.decoderTabs)
          }"
        >
          <div 
            class="decoder-tab"
            :class="{ 
              'decoder-tab-active': tab.isActive,
              'decoder-tab-dragging': draggedTabId === tab.id
            }"
            @click="handleTabSelect(tab.id)"
            @dblclick="handleDoubleClick(tab)"
            draggable="true"
            @dragstart="handleDragStart($event, tab)"
            @dragend="handleDragEnd"
            @dragenter="handleDragEnter($event, tab)"
            @dragleave="handleDragLeave"
            @dragover="handleDragOver"
            @drop="handleDrop($event, tab)"
          >
            <!-- 颜色指示条 -->
            <div class="absolute left-0 top-0 bottom-0 w-0.5 bg-[#4f46e5]"></div>
            
            <!-- 标签名称（编辑时显示输入框） -->
            <div class="flex items-center ml-1 max-w-[150px]">
              <div v-if="editingTabId === tab.id" class="flex items-center">
                <input 
                  spellcheck="false"
                  ref="tabNameInput"
                  v-model="editingTabName"
                  @blur="finishRenaming"
                  @keydown="handleKeyDown"
                  class="text-xs w-32 px-1 py-0.5 bg-transparent border border-gray-300 dark:border-gray-600 rounded focus:ring-1 focus:ring-indigo-500 focus:border-indigo-500 dark:focus:ring-indigo-400 dark:focus:border-indigo-400 outline-none"
                />
              </div>
              <div 
                v-else 
                class="decoder-tab-name"
                :class="{ 
                  'font-medium text-gray-900 dark:text-white': tab.isActive,
                  'text-gray-500 dark:text-gray-400': !tab.isActive
                }"
              >
                {{ tab.name }}
              </div>
            </div>
            
            <!-- 关闭按钮 -->
            <button 
              class="decoder-tab-close"
              :class="{ 'ml-2 text-gray-400 hover:text-gray-700 dark:hover:text-gray-300': true }"
              @click.stop="handleTabClose(tab.id, $event)"
              :aria-label="t('modules.decoder.close_tab')"
            >
              <i class="bx bx-x text-sm"></i>
            </button>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
/* 滚动条样式已移至 scrollbar.css 统一管理 */
</style>