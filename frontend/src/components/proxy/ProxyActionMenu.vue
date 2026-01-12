<script setup lang="ts">
import { ref, computed, onMounted, onUnmounted, nextTick, watch } from 'vue';
import { useI18n } from 'vue-i18n';
import { getThemeColors } from '../../utils';
import type { ProxyHistoryItem } from '../../store';
import type { FilterOptions } from './ProxyFilter.vue';

const props = defineProps<{
  request: ProxyHistoryItem | null;
  x: number;
  y: number;
  visible: boolean;
  filterOptions: FilterOptions;
}>();

const emit = defineEmits<{
  close: [];
  'send-to-repeater': [request: ProxyHistoryItem];
  'send-to-intruder': [request: ProxyHistoryItem];
  'set-color': [request: ProxyHistoryItem, color: string];
  'clear-all-history': [];
  'clean-memory-only': [];
  'filter-by-host': [host: string];
  'filter-by-method': [method: string];
  'clear-filters': [];
}>();

const { t } = useI18n();
const menuRef = ref<HTMLDivElement | null>(null);

// 当前模式下的颜色
const availableColors = computed(() => getThemeColors());

// 二级菜单状态
const showColorSubmenu = ref(false);
const colorSubmenuPosition = ref({ x: 0, y: 0 });

// 计算当前筛选状态
const hasActiveFilters = computed(() => {
  return !!(props.filterOptions.method ||
           props.filterOptions.host ||
           props.filterOptions.path ||
           props.filterOptions.status ||
           props.filterOptions.contentType);
});

const isHostFiltered = computed(() => !!props.filterOptions.host);
const isMethodFiltered = computed(() => !!props.filterOptions.method);

// 设置颜色
const setRowColor = (color: string) => {
  if (props.request) {
  emit('set-color', props.request, color);
  }
  emit('close');
};

// 切换颜色子菜单显示状态
const toggleColorMenu = (event: MouseEvent) => {
  event.stopPropagation(); // 防止事件冒泡

  if (showColorSubmenu.value) {
    showColorSubmenu.value = false;
    return;
  }

  const target = event.currentTarget as HTMLElement;
  const rect = target.getBoundingClientRect();

  // 计算子菜单位置，确保不超出视口
  let x = rect.right + 5;
  let y = rect.top;

  // 检查是否会超出右边界
  const submenuWidth = 200; // 预估子菜单宽度
  if (x + submenuWidth > window.innerWidth) {
    x = rect.left - submenuWidth - 5; // 显示在左侧
  }

  // 检查是否会超出下边界
  const submenuHeight = 150; // 预估子菜单高度
  if (y + submenuHeight > window.innerHeight) {
    y = window.innerHeight - submenuHeight - 10;
  }

  colorSubmenuPosition.value = { x, y };
  showColorSubmenu.value = true;
};

const handleClickOutside = (event: MouseEvent) => {
  const target = event.target as Node;

  // 检查是否点击在主菜单或颜色子菜单外
  const clickedInMainMenu = menuRef.value && menuRef.value.contains(target);
  const colorSubmenu = document.querySelector('.color-submenu');
  const clickedInColorSubmenu = colorSubmenu && colorSubmenu.contains(target);

  if (!clickedInMainMenu && !clickedInColorSubmenu) {
    emit('close');
  } else if (!clickedInColorSubmenu && showColorSubmenu.value) {
    // 如果点击在主菜单内但不在颜色子菜单内，关闭颜色子菜单
    showColorSubmenu.value = false;
  }
};

const sendToRepeater = () => {
  if (props.request) {
  emit('send-to-repeater', props.request);
  }
  emit('close');
};

const sendToIntruder = () => {
  if (props.request) {
  emit('send-to-intruder', props.request);
  }
  emit('close');
};

const clearAllHistory = () => {
  emit('clear-all-history');
  emit('close');
};

// 复制URL到剪贴板
const copyUrl = async () => {
  if (props.request?.url) {
    try {
      await navigator.clipboard.writeText(props.request.url);
      // 这里可以显示复制成功的提示
    } catch (err) {
      console.error('Failed to copy URL:', err);
    }
  }
  emit('close');
};

// 复制Host到剪贴板
const copyHost = async () => {
  if (props.request?.host) {
    try {
      await navigator.clipboard.writeText(props.request.host);
      // 这里可以显示复制成功的提示
    } catch (err) {
      console.error('Failed to copy host:', err);
    }
  }
  emit('close');
};

// 按主机筛选
const filterByHost = () => {
  if (props.request?.host) {
    emit('filter-by-host', props.request.host);
  }
  emit('close');
};

// 按方法筛选
const filterByMethod = () => {
  if (props.request?.method) {
    emit('filter-by-method', props.request.method);
  }
  emit('close');
};

// 仅清理内存数据
const cleanMemoryOnly = () => {
  emit('clean-memory-only');
  emit('close');
};

// 清除所有筛选
const clearFilters = () => {
  emit('clear-filters');
  emit('close');
};

// Position the menu within viewport bounds
const adjustPosition = () => {
  if (!menuRef.value) return;
  
  nextTick(() => {
    if (!menuRef.value) return;
    
    const menu = menuRef.value;
    const rect = menu.getBoundingClientRect();
    const viewportWidth = window.innerWidth;
    const viewportHeight = window.innerHeight;
    
    let left = props.x;
    let top = props.y;
    
    // Check if menu would go off the right edge
    if (left + rect.width > viewportWidth) {
      left = viewportWidth - rect.width - 5;
    }
    
    // Check if menu would go off the bottom edge
    if (top + rect.height > viewportHeight) {
      top = viewportHeight - rect.height - 5;
    }
    
    menu.style.left = `${left}px`;
    menu.style.top = `${top}px`;
  });
};

onMounted(() => {
  document.addEventListener('click', handleClickOutside);
  window.addEventListener('resize', adjustPosition);
  adjustPosition();
});

onUnmounted(() => {
  document.removeEventListener('click', handleClickOutside);
  window.removeEventListener('resize', adjustPosition);
});

// Reposition when visibility changes
watch(() => props.visible, () => {
  if (props.visible) {
    adjustPosition();
  }
});
</script>

<template>
  <div 
    v-if="visible" 
    ref="menuRef"
    class="context-menu"
    style="min-width: 240px;"
  >
    <!-- 发送到工具 -->
    <div class="py-1">
      <button 
        class="context-menu-item"
        @click="sendToRepeater"
      >
        <i class="bx bx-repeat text-indigo-500"></i>
        {{ t('modules.proxy.actions.sendToRepeater') }}
        <span class="context-menu-shortcut">Ctrl+R</span>
      </button>
      <button 
        class="context-menu-item"
        @click="sendToIntruder"
      >
        <i class="bx bx-target-lock text-orange-500"></i>
        {{ t('modules.proxy.actions.send_to_intruder') }}
        <span class="context-menu-shortcut">Ctrl+I</span>
      </button>
      <button 
        class="context-menu-item"
        @click="copyUrl"
      >
        <i class="bx bx-copy text-blue-500"></i>
        复制 URL
      </button>
      <button 
        class="context-menu-item"
        @click="copyHost"
      >
        <i class="bx bx-link text-green-500"></i>
        复制 Host
      </button>
    </div>
    
    <!-- 颜色标记区 - 改为二级菜单 -->
    <div class="context-menu-divider"></div>
    <button
      class="context-menu-item submenu-item"
      @click="toggleColorMenu"
    >
      <i class="bx bx-palette text-purple-500"></i>
      {{ t('modules.proxy.actions.set_row_color') }}
      <i class="bx bx-chevron-right ml-auto text-gray-400"
         :class="{ 'rotate-90': showColorSubmenu }"></i>
    </button>
    
    <!-- 筛选操作 -->
    <div class="context-menu-divider"></div>

    <!-- 主机筛选 -->
    <button
      v-if="!isHostFiltered"
      class="context-menu-item"
      @click="filterByHost"
    >
      <i class="bx bx-filter text-purple-500"></i>
      仅显示此主机
    </button>
    <button
      v-else
      class="context-menu-item text-orange-600"
      @click="clearFilters"
    >
      <i class="bx bx-filter-alt text-orange-500"></i>
      取消主机筛选
      <span class="text-xs text-gray-400 ml-auto">{{ filterOptions.host }}</span>
    </button>

    <!-- 方法筛选 -->
    <button
      v-if="!isMethodFiltered"
      class="context-menu-item"
      @click="filterByMethod"
    >
      <i class="bx bx-code text-cyan-500"></i>
      仅显示 {{ request?.method }} 请求
    </button>
    <button
      v-else
      class="context-menu-item text-orange-600"
      @click="clearFilters"
    >
      <i class="bx bx-code-alt text-orange-500"></i>
      取消方法筛选
      <span class="text-xs text-gray-400 ml-auto">{{ filterOptions.method }}</span>
    </button>

    <!-- 清除所有筛选 -->
    <button
      v-if="hasActiveFilters && !isHostFiltered && !isMethodFiltered"
      class="context-menu-item text-red-600"
      @click="clearFilters"
    >
      <i class="bx bx-x-circle text-red-500"></i>
      清除所有筛选
    </button>

    <!-- 危险操作区 -->
    <div class="context-menu-divider"></div>
    <div class="px-3 py-2 text-xs font-medium text-red-500 dark:text-red-400">
      <i class="bx bx-trash mr-1"></i>
      清理操作
    </div>
    <button 
      class="context-menu-item text-red-600 dark:text-red-400 hover:bg-red-50 dark:hover:bg-red-900/20"
      @click="clearAllHistory"
    >
      <i class="bx bx-trash-alt text-red-600"></i>
      清空所有历史记录
      <span class="text-xs text-gray-400 ml-auto">数据库+内存</span>
    </button>
    <button 
      class="context-menu-item text-orange-600 dark:text-orange-400 hover:bg-orange-50 dark:hover:bg-orange-900/20"
      @click="cleanMemoryOnly"
    >
      <i class="bx bx-memory-card text-orange-600"></i>
      仅清理内存数据
      <span class="text-xs text-gray-400 ml-auto">保留数据库</span>
    </button>
    
    <!-- 取消 -->
    <div class="context-menu-divider"></div>
    <button
      class="context-menu-item"
      @click="emit('close')"
    >
      <i class="bx bx-x text-gray-500"></i>
      {{ t('common.actions.cancel') }}
    </button>
  </div>

  <!-- 颜色选择子菜单 -->
  <div
    v-if="showColorSubmenu"
    class="color-submenu context-menu"
    :style="{
      position: 'fixed',
      left: colorSubmenuPosition.x + 'px',
      top: colorSubmenuPosition.y + 'px',
      minWidth: '200px'
    }"
  >
    <div class="px-3 py-2 text-xs font-medium text-purple-600 dark:text-purple-400 border-b border-purple-200 dark:border-purple-700 bg-purple-50 dark:bg-purple-900/20">
      <i class="bx bx-palette mr-1 text-purple-500"></i>
      选择颜色标记
    </div>
    <div class="color-picker p-3">
      <button
        v-for="colorOption in availableColors"
        :key="colorOption.id"
        @click="setRowColor(colorOption.color)"
        class="color-option enhanced-color-option"
        :class="{ 
          'selected': request?.color === colorOption.color,
          'none': colorOption.id === 'none'
        }"
        :style="{ 
          backgroundColor: colorOption.color || 'transparent',
          '--option-color': colorOption.color || 'transparent'
        }"
        :title="colorOption.id === 'none' ? '移除颜色' : `设置为${colorOption.id}色`"
      >
        <i v-if="colorOption.id === 'none'" class="bx bx-x text-red-500 text-xl font-bold"></i>
        <span v-if="request?.color === colorOption.color" class="selected-indicator">✓</span>
      </button>
    </div>
  </div>
</template>

<style scoped>
  /* 右键菜单项快捷键样式 */
  .context-menu-shortcut {
    margin-left: auto;
    font-size: 0.75rem;
    color: var(--color-text-tertiary);
    opacity: 0.6;
  }

  /* 子菜单项样式 */
  .submenu-item {
    position: relative;
  }

  .submenu-item:hover {
    background-color: var(--color-bg-secondary);
  }

  /* 颜色子菜单样式 */
  .color-submenu {
    z-index: 1001; /* 确保在主菜单之上 */
    box-shadow: 0 8px 24px rgba(0, 0, 0, 0.25);
    border: 2px solid var(--color-primary);
  }
  
  /* 颜色选择器网格增强 */
  .color-picker {
    display: grid;
    grid-template-columns: repeat(auto-fit, minmax(2rem, 1fr));
    gap: 0.75rem;
    padding: 1rem;
    background: var(--glass-bg-light);
    border-radius: var(--radius-md);
  }

  /* 箭头旋转动画 */
  .bx-chevron-right {
    transition: transform 0.2s ease;
  }

  .bx-chevron-right.rotate-90 {
    transform: rotate(90deg);
  }

  /* 危险操作特殊样式 */
  .context-menu-item.text-red-600 {
    color: var(--color-error, #dc2626);
  }

  .context-menu-item.text-orange-600 {
    color: var(--color-warning-text, #ea580c);
  }

  .context-menu-item.text-red-600:hover {
    background-color: var(--color-danger-bg, #fef2f2);
  }

  .context-menu-item.text-orange-600:hover {
    background-color: var(--color-warning-bg, #fff7ed);
  }

  .dark .context-menu-item.text-red-600:hover {
    background-color: rgba(153, 27, 27, 0.2);
  }

  .dark .context-menu-item.text-orange-600:hover {
    background-color: rgba(234, 88, 12, 0.2);
  }

  /* 右侧辅助文本 */
  .context-menu-item .text-gray-400 {
    color: var(--color-text-tertiary, #9ca3af);
  }
  
  /* 增强的颜色选项样式 */
  .enhanced-color-option {
    width: 2rem;
    height: 2rem;
    border-radius: 50%;
    border: 3px solid var(--glass-border);
    cursor: pointer;
    transition: all 0.2s ease;
    display: flex;
    align-items: center;
    justify-content: center;
    position: relative;
    overflow: hidden;
    box-shadow: 0 2px 8px rgba(0, 0, 0, 0.15),
                inset 0 1px 0 rgba(255, 255, 255, 0.2);
  }
  
  .enhanced-color-option:hover {
    transform: scale(1.15);
    border-color: var(--color-primary);
    box-shadow: 0 4px 16px rgba(0, 0, 0, 0.25),
                0 0 0 3px rgba(59, 130, 246, 0.3),
                inset 0 1px 0 rgba(255, 255, 255, 0.3);
    z-index: 10;
  }
  
  .enhanced-color-option.selected {
    border-color: var(--color-primary);
    border-width: 4px;
    transform: scale(1.1);
    box-shadow: 0 0 0 4px rgba(59, 130, 246, 0.3),
                0 6px 20px rgba(59, 130, 246, 0.2),
                inset 0 1px 0 rgba(255, 255, 255, 0.4);
  }
  
  .enhanced-color-option.none {
    background: linear-gradient(45deg, 
                transparent 25%, 
                rgba(255, 0, 0, 0.1) 25%, 
                rgba(255, 0, 0, 0.1) 50%, 
                transparent 50%, 
                transparent 75%, 
                rgba(255, 0, 0, 0.1) 75%) !important;
    background-size: 6px 6px !important;
    border-style: dashed;
    border-color: var(--color-error);
  }
  
  .enhanced-color-option.none:hover {
    border-style: solid;
    background: var(--glass-bg-light) !important;
    box-shadow: 0 4px 16px rgba(239, 68, 68, 0.25),
                0 0 0 3px rgba(239, 68, 68, 0.2);
  }
  
  .selected-indicator {
    position: absolute;
    color: white;
    font-size: 1rem;
    font-weight: 900;
    text-shadow: 0 0 4px rgba(0, 0, 0, 0.9),
                 0 2px 4px rgba(0, 0, 0, 0.8);
    filter: drop-shadow(0 0 2px rgba(0, 0, 0, 0.5));
  }
</style> 