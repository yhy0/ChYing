<script setup lang="ts" generic="T extends HttpTrafficItem">
import { ref, computed, onMounted, watch, toRef } from 'vue';
import { useI18n } from 'vue-i18n';
import { getThemeColors } from '../../utils';
import { FlexRender, type Cell } from '@tanstack/vue-table';
import { useHttpTrafficTable } from '../../composables';
import ColorPicker from './ColorPicker.vue';
import type { HttpTrafficItem, HttpTrafficColumn } from '../../types';
import { logMemoryUsage } from '../../utils/tableOptimization';

const props = defineProps<{
  items: T[]; // 使用泛型 T
  selectedItem: T | null; // 使用泛型 T
  tableId?: string; // 表格ID，用于保存不同表格的设置，如列宽
  customColumns?: HttpTrafficColumn<T>[]; // 使用泛型 T
  tableClass?: string; // 额外的表格CSS类
  containerHeight?: string; // 容器高度
  enableMultiSelect?: boolean; // 是否启用多选模式
  checkedItems?: T[]; // 已选中的项目列表
}>();

const emit = defineEmits<{
  (e: 'select-item', item: T): void; // 使用泛型 T
  (e: 'context-menu', event: MouseEvent, item: T): void; // 使用泛型 T
  (e: 'set-color', item: T, color: string): void; // 使用泛型 T
  (e: 'update:checkedItems', items: T[]): void; // 多选项目更新
}>();

const { t } = useI18n();

// 获取表格ID，如果未提供则使用默认值
const effectiveTableId = computed(() => props.tableId || 'http-traffic-table');

// 颜色选择器状态
const showColorMenu = ref(false);
const colorMenuPosition = ref({ x: 0, y: 0 });
const colorMenuItemId = ref<number | null>(null);

// 当前模式下的颜色
const availableColors = computed(() => getThemeColors());

// 将 props.items 转换为响应式引用，确保数据变化时能正确更新
const itemsRef = toRef(props, 'items');

// 使用HTTP流量表格组合式API
const {
  table,
  rows,                    // 所有行数据
  visibleRows,
  virtualRows,             // 虚拟行数据
  totalSize,               // 总高度
  tableContainerRef,       // 容器引用
  measureElement,          // 测量函数
  columnWidths,
  isResizing,
  resizingColumnId,
  getRowStyle,
  handleResizeStart,
  getFlatHeaders,
  hasSavedSorting,
  resetTableSorting
} = useHttpTrafficTable(itemsRef, {
  tableId: effectiveTableId.value,
  customColumns: props.customColumns,
  i18nTranslate: t,
  estimateRowHeight: 33        // 行高33px
});

// 是否显示排序重置按钮
const showSortReset = computed(() => hasSavedSorting.value);

// 确保表格初始化完成
onMounted(() => {
  // 性能监控：记录组件挂载时的内存使用
  if (props.items.length > 1000) {
    logMemoryUsage(`HttpTrafficTable 挂载 - ${props.items.length} 条数据`);
  }
});

// 监听数据变化，进行性能日志
watch(() => props.items.length, (newCount, oldCount) => {
  if (newCount > 1000 && newCount !== oldCount) {
    logMemoryUsage(`数据更新 - 从 ${oldCount} 到 ${newCount} 条`);
  }
}, { flush: 'post' });

// 行点击事件
const handleRowClick = (item: T) => { // 使用泛型 T
  emit('select-item', item);
};

// 右键菜单事件
const handleContextMenu = (event: MouseEvent, item: T) => { // 使用泛型 T
  event.preventDefault();
  emit('select-item', item);
  emit('context-menu', event, item);
};

// 处理ID单元格点击事件
const handleIdCellClick = (event: MouseEvent, item: T) => { // 使用泛型 T
  event.stopPropagation(); 
  event.preventDefault();
  
  emit('select-item', item);
  
  // 如果已显示同一项的颜色菜单，则关闭
  if (showColorMenu.value && colorMenuItemId.value === item.id) {
    showColorMenu.value = false;
    return;
  }
  
  // 显示颜色菜单
  colorMenuItemId.value = item.id;
  
  // 计算颜色菜单位置
  const cellElement = event.currentTarget as HTMLElement;
  const rect = cellElement.getBoundingClientRect();
  
  colorMenuPosition.value = {
    x: rect.left + rect.width,
    y: rect.top
  };
  
  // 立即显示菜单
  showColorMenu.value = true;
};

// 处理颜色选择
const handleColorSelect = (color: string) => {
  if (colorMenuItemId.value !== null) {
    const item = props.items.find(item => item.id === colorMenuItemId.value);
    if (item) {
      emit('set-color', item, color);
    }
  }
  closeColorMenu();
};

// 关闭颜色菜单
const closeColorMenu = () => {
  showColorMenu.value = false;
};

// 处理列排序
const handleColumnSort = (columnId: string) => {
  const column = table.getColumn(columnId);
  if (column && column.getCanSort()) {
    column.toggleSorting();
  }
};

// 重置表格排序
const handleResetSorting = () => {
  resetTableSorting();
};

// Helper function to get value for title attribute
const getCellTitleValue = (cell: Cell<T, unknown>): string | undefined => {
  const value = cell.getValue();
  // Return undefined if value is null or undefined, so title attribute isn't set
  return value !== null && value !== undefined ? String(value) : undefined;
};

// 多选相关
const checkedItemIds = computed(() => {
  return new Set((props.checkedItems || []).map(item => item.id));
});

const isAllChecked = computed(() => {
  if (!props.enableMultiSelect || props.items.length === 0) return false;
  return props.items.every(item => checkedItemIds.value.has(item.id));
});

const isIndeterminate = computed(() => {
  if (!props.enableMultiSelect || props.items.length === 0) return false;
  const checkedCount = props.items.filter(item => checkedItemIds.value.has(item.id)).length;
  return checkedCount > 0 && checkedCount < props.items.length;
});

// 切换单个项目的选中状态
const toggleItemCheck = (item: T, event: Event) => {
  event.stopPropagation();
  const currentChecked = props.checkedItems || [];
  const isChecked = checkedItemIds.value.has(item.id);

  let newChecked: T[];
  if (isChecked) {
    newChecked = currentChecked.filter(i => i.id !== item.id);
  } else {
    newChecked = [...currentChecked, item];
  }

  emit('update:checkedItems', newChecked);
};

// 切换全选状态
const toggleAllCheck = (event: Event) => {
  event.stopPropagation();
  if (isAllChecked.value) {
    // 取消全选
    emit('update:checkedItems', []);
  } else {
    // 全选当前显示的所有项目
    emit('update:checkedItems', [...props.items]);
  }
};

</script>

<template>
  <div class="http-traffic-table-container relative h-full">
    
    <!-- 虚拟滚动容器 -->
    <div
      ref="tableContainerRef"
      class="virtual-table-container"
      :style="{
        overflow: 'auto',
        position: 'relative',
        height: props.containerHeight || '600px', // 使用prop或默认值
        width: '100%',
        minWidth: 'fit-content' // 确保容器能够适应内容宽度
      }"
    >
      <div :style="{ height: `${totalSize}px` }">
        <!-- 使用CSS Grid布局的虚拟化表格 -->
        <table class="http-traffic-table virtual-table" :class="props.tableClass" :style="{
          display: 'grid',
          width: '100%',
          minWidth: 'fit-content'
        }">
          <thead
            :style="{
              display: 'grid',
              position: 'sticky',
              top: 0,
              zIndex: 10,
            }"
          >
            <!-- 重置排序按钮 - 浮动在表头左上角外侧，不遮挡表头文字 -->
            <div
              v-if="showSortReset"
              class="reset-sort-floating-btn absolute -top-2 -left-2 z-20"
            >
              <button
                class="reset-sort-btn flex items-center text-xs px-2 py-1.5 rounded-lg bg-white hover:bg-blue-50 dark:bg-gray-800 dark:hover:bg-blue-900/30 border border-gray-300 dark:border-gray-600 shadow-md hover:shadow-lg transition-all duration-200 text-blue-600 dark:text-blue-400 hover:text-blue-700 dark:hover:text-blue-300"
                @click="handleResetSorting"
                :title="t('common.ui.reset_sorting_title')"
              >
                <i class="bx bx-reset text-sm"></i>
                <span class="ml-1.5 hidden lg:inline text-xs font-medium">{{ t('common.ui.reset_sorting') }}</span>
              </button>
            </div>
            <tr
              v-for="headerGroup in table.getHeaderGroups()"
              :key="headerGroup.id"
              :style="{ display: 'flex', width: '100%' }"
            >
              <!-- 多选复选框列头 -->
              <th
                v-if="props.enableMultiSelect"
                scope="col"
                class="checkbox-header"
                :style="{
                  width: '40px',
                  minWidth: '40px',
                  maxWidth: '40px',
                  flexShrink: 0,
                  display: 'flex',
                  alignItems: 'center',
                  justifyContent: 'center'
                }"
              >
                <input
                  type="checkbox"
                  :checked="isAllChecked"
                  :indeterminate="isIndeterminate"
                  @change="toggleAllCheck"
                  class="multi-select-checkbox"
                  :aria-label="t('common.ui.accessibility.selectAll')"
                />
              </th>
              <th
                v-for="header in headerGroup.headers"
                :key="header.id"
                scope="col"
                class="group cursor-default"
                :class="{ 'text-center': header.id === 'id' }"
                :style="{
                  width: `${header.getSize()}px`,
                  minWidth: `${header.getSize()}px`,
                  maxWidth: `${header.getSize()}px`,
                  flexShrink: 0
                }"
                :data-column-id="header.id"
              >
                <div 
                  class="flex items-center" 
                  :class="{ 'justify-center': header.id === 'id' }"
                >
                  <FlexRender
                    :render="header.column.columnDef.header"
                    :props="header.getContext()"
                  />
                  <i 
                    v-if="header.column.getCanSort()"
                    :class="[
                      header.column.getIsSorted() === 'asc' ? 'bx bx-sort-up text-blue-500' : 
                      header.column.getIsSorted() === 'desc' ? 'bx bx-sort-down text-blue-500' : 
                      'bx bx-sort text-gray-400'
                    ]" 
                    class="ml-1 text-sm sort-icon"
                    @click.stop="handleColumnSort(header.id)"
                  ></i>
                </div>
                <!-- 列调整器 -->
                <div 
                  class="resizer" 
                  @mousedown.stop.prevent="handleResizeStart($event, header.id)"
                  :class="{ 'resizing': isResizing && resizingColumnId === header.id }"
                ></div>
              </th>
            </tr>
          </thead>
          <tbody
            :style="{
              display: 'grid',
              height: `${totalSize}px`,
              position: 'relative',
            }"
          >
            <tr
              v-for="vRow in virtualRows"
              :key="rows[vRow.index]?.id || vRow.index"
              :data-index="vRow.index"
              :ref="measureElement"
              :style="{
                display: 'flex',
                position: 'absolute',
                transform: `translateY(${vRow.start}px)`,
                width: '100%',
                height: '33px',
                minHeight: '33px',
                ...(rows[vRow.index]?.original.color ? { '--row-color': rows[vRow.index].original.color } : {})
              }"
              :data-row-color="rows[vRow.index]?.original.color || null"
              @click="handleRowClick(rows[vRow.index].original)"
              @contextmenu="handleContextMenu($event, rows[vRow.index].original)"
              :class="[
                'cursor-pointer virtual-row',
                props.selectedItem && props.selectedItem.id === rows[vRow.index]?.original.id ? 'selected' : '',
                props.enableMultiSelect && checkedItemIds.has(rows[vRow.index]?.original.id) ? 'checked-row' : ''
              ]"
            >
              <!-- 多选复选框单元格 -->
              <td
                v-if="props.enableMultiSelect"
                class="checkbox-cell"
                :style="{
                  display: 'flex',
                  width: '40px',
                  minWidth: '40px',
                  maxWidth: '40px',
                  flexShrink: 0,
                  alignItems: 'center',
                  justifyContent: 'center'
                }"
                @click.stop
              >
                <input
                  type="checkbox"
                  :checked="checkedItemIds.has(rows[vRow.index]?.original.id)"
                  @change="toggleItemCheck(rows[vRow.index].original, $event)"
                  class="multi-select-checkbox"
                  :aria-label="t('common.ui.accessibility.selectRow', { id: rows[vRow.index]?.original.id })"
                />
              </td>
              <td
                v-for="cell in rows[vRow.index]?.getVisibleCells() || []"
                :key="cell.id"
                :style="{
                  display: 'flex',
                  width: `${cell.column.getSize()}px`,
                  minWidth: `${cell.column.getSize()}px`,
                  maxWidth: `${cell.column.getSize()}px`,
                  flexShrink: 0
                }"
                :class="[
                  'whitespace-nowrap align-items-center',
                  cell.column.id === 'id' ? 'relative text-center' : '',
                  cell.column.id === 'length' ? 'text-right' : '',
                  cell.column.id === 'enabled' ? 'enabled-column' : '',
                  cell.column.id === 'path' ? 'truncate-cell' : '',
                  cell.column.id === 'host' ? 'truncate-cell' : '',
                  cell.column.id === 'title' ? 'truncate-cell' : '',
                  cell.column.id === 'note' ? 'truncate-cell' : '',
                  cell.column.id === 'url' ? 'truncate-cell' : '',
                  cell.column.id === 'timestamp' ? 'truncate-cell' : '',
                  cell.column.id === 'ruleDescription' ? 'truncate-cell' : ''
                ]"
                :data-id-cell="cell.column.id === 'id' ? true : undefined"
                :data-column="cell.column.id"
                :title="getCellTitleValue(cell)"
              >
                <template v-if="cell.column.id === 'id'">
                  <span class="flex items-center justify-center w-full" @click.stop="handleIdCellClick($event, rows[vRow.index].original)">
                    {{ rows[vRow.index].original.id }}
                    <i class="bx bx-chevron-down ml-1 opacity-60"></i>
                  </span>
                </template>
                <template v-else-if="cell.column.id === 'method'">
                  <div :class="['method-tag', `method-tag-${rows[vRow.index].original.method.toLowerCase()}`]">
                    {{ rows[vRow.index].original.method }}
                  </div>
                </template>
                <template v-else-if="cell.column.id === 'status' && rows[vRow.index].original.status">
                  <div :class="['status-tag', `status-${Math.floor(rows[vRow.index].original.status / 100)}xx`]">
                    {{ rows[vRow.index].original.status }}
                  </div>
                </template>
                <template v-else>
                  <FlexRender
                    :render="cell.column.columnDef.cell"
                    :props="cell.getContext()"
                  />
                </template>
              </td>
            </tr>

            <!-- 没有数据时显示空状态 -->
            <tr v-if="props.items.length === 0" class="empty-state-row">
              <td :colspan="table.getAllColumns().length" class="py-8 text-center text-gray-500 dark:text-gray-400">
                {{ t('common.status.no_data') }}
              </td>
            </tr>
          </tbody>
        </table>
      </div>
    </div>
    
    <!-- 颜色选择器组件 -->
    <ColorPicker
      :colors="availableColors"
      :position="colorMenuPosition"
      :shown="showColorMenu"
      @select="handleColorSelect"
      @close="closeColorMenu"
    />
  </div>
</template>

<!-- 样式已迁移到 styles/components/table.css 统一管理 -->
