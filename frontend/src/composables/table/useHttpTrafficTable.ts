import { computed, ref, markRaw, watch, type Ref, nextTick, type ComponentPublicInstance } from 'vue';
import {
  createColumnHelper,
  getCoreRowModel,
  getSortedRowModel,
  useVueTable,
  type ColumnDef,
  type SortingState as TanStackSortingState,
  type Row,
} from '@tanstack/vue-table';
import { useVirtualizer } from '@tanstack/vue-virtual';
import { formatDateTime, formatSize } from '../../utils/formatters';
import { defaultHttpTrafficColumns } from '../../types/http-traffic-columns';
import type { HttpTrafficItem } from '../../types/http';
import type { HttpTrafficColumn } from '../../types/table';
import { useTableSort } from './useTableSort';
import { useTableColumnResize } from './useTableColumnResize';
import { optimizeDataForRendering, safeGetArrayData } from '../../utils/tableOptimization';

interface UseHttpTrafficTableOptions<T> {
  tableId?: string;
  customColumns?: HttpTrafficColumn<T>[];
  i18nTranslate: (key: string) => string;
  estimateRowHeight?: number;
}

export function useHttpTrafficTable<T extends HttpTrafficItem>(
  itemsRaw: T[] | Ref<T[]>,
  options: UseHttpTrafficTableOptions<T>
) {
  const { 
    tableId = 'http-traffic-table', 
    customColumns, 
    i18nTranslate: t,
    estimateRowHeight = 33
  } = options;
  
  // 优化：使用工具函数处理数据和响应式优化
  const items = computed<T[]>(() => {
    const rawItems = safeGetArrayData(itemsRaw);
    return optimizeDataForRendering(rawItems, 1000);
  });
  
  const { 
    sorting, 
    resetSorting, 
    savedSortingExists 
  } = useTableSort({
    localStorageKey: `${tableId}-sorting`
  });
  
  const sortingState = computed<TanStackSortingState>(() => {
    return Array.isArray(sorting.value) ? sorting.value : [];
  });
  
  const hasSavedSorting = computed(() => savedSortingExists());
  
  const columns = computed<HttpTrafficColumn<T>[]>(() => {
    return customColumns?.length 
      ? customColumns 
      : (defaultHttpTrafficColumns as HttpTrafficColumn<T>[]);
  });

  const getDefaultWidths = computed(() => {
    const widths: Record<string, number> = {};
    columns.value.forEach(col => {
      widths[col.id] = col.width;
    });
    return widths;
  });

  const columnHelper = markRaw(createColumnHelper<T>());

  const tableColumns = computed(() => {
    return columns.value.map(column => {
      if (column.cellRenderer) {
        return columnHelper.display({
          id: column.id,
          header: column.name,
          size: column.width,
          minSize: 20,
          enableResizing: true,
          enableSorting: false,
          cell: ({ row }) => column.cellRenderer!({ item: row.original as T })
        });
      }

      const colConfig = {
        header: column.name,
        size: column.width,
        minSize: 20,
        enableResizing: true,
        enableSorting: true,
      };

      switch (column.id as keyof HttpTrafficItem) {
        case 'id':
          return columnHelper.accessor('id' as any, {
            ...colConfig,
            header: '#',
            cell: ({ row }) => row.original.id
          });
        case 'method':
          return columnHelper.accessor('method' as any, {
            ...colConfig,
            cell: ({ getValue }) => getValue()
          });
        case 'host':
          return columnHelper.accessor('host' as any, {
            ...colConfig,
            cell: ({ getValue }) => getValue()
          });
        case 'path':
          return columnHelper.accessor('path' as any, {
            ...colConfig,
            cell: ({ getValue }) => {
              const value = getValue();
              return value || '';
            },
            meta: {
              truncate: true
            }
          });
        case 'status':
          return columnHelper.accessor('status' as any, {
            ...colConfig,
            cell: ({ getValue }) => getValue()
          });
        case 'length':
          return columnHelper.accessor('length' as any, {
            ...colConfig,
            cell: ({ getValue }) => formatSize(getValue() as number)
          });
        case 'mimeType':
          return columnHelper.accessor('mimeType' as any, {
            ...colConfig,
            cell: ({ getValue }) => {
              const value = getValue() as string | undefined;
              return value || '';
            }
          });
        case 'extension':
          return columnHelper.accessor('extension' as any, {
            ...colConfig,
            size: column.width,
            maxSize: 100,
            cell: ({ getValue }) => {
              const value = getValue() as string | undefined;
              return value && value.length > 10 ? value.slice(0, 10) + '...' : (value || '');
            },
            meta: {
              truncate: true
            }
          });
        case 'title':
          return columnHelper.accessor('title' as any, {
            ...colConfig,
            cell: ({ getValue }) => {
              const value = getValue() as string | undefined;
              return value || '';
            }
          });
        case 'ip':
          return columnHelper.accessor('ip' as any, {
            ...colConfig,
            cell: ({ getValue }) => {
              const value = getValue() as string | undefined;
              return value || '';
            }
          });
        case 'note':
          return columnHelper.accessor('note' as any, {
            ...colConfig,
            cell: ({ getValue }) => {
              const value = getValue() as string | undefined;
              return value || '';
            }
          });
        case 'time':
          return columnHelper.accessor('timestamp' as any, {
            ...colConfig,
            header: t('common.ui.time'),
            cell: ({ getValue }) => formatDateTime(getValue() as string)
          });
        case 'timeMs':
          return columnHelper.accessor('timeMs' as any, {
            ...colConfig,
            header: t('common.ui.time_ms'),
            cell: ({ getValue }) => {
              const value = getValue() as number | undefined;
              return value ? `${value} ms` : '';
            }
          });
        case 'payload':
          return columnHelper.accessor('payload' as any, {
            ...colConfig,
            cell: ({ getValue }) => {
              const value = getValue() as string | undefined;
              return value || '';
            }
          });
        case 'url':
          return columnHelper.accessor('url' as any, {
            ...colConfig,
            cell: ({ getValue }) => {
              const value = getValue() as string;
              return value || '';
            },
            meta: {
              truncate: true
            }
          });
        case 'timestamp':
          return columnHelper.accessor('timestamp' as any, {
            ...colConfig,
            cell: ({ getValue }) => {
              const value = getValue() as string | undefined;
              return value ? formatDateTime(value) : '';
            },
            meta: {
              truncate: true
            }
          });
        case 'ruleDescription':
          return columnHelper.accessor('ruleDescription' as any, {
            ...colConfig,
            cell: ({ getValue }) => {
              const value = getValue() as string;
              return value || '';
            },
            meta: {
              truncate: true
            }
          });
        default:
          return columnHelper.accessor(column.id as any, {
            ...colConfig,
            header: column.name,
            cell: ({ row }) => {
              const value = (row.original as any)[column.id as string];
              return value !== undefined ? String(value) : '';
            }
          });
      }
    });
  });

  const { 
    columnWidths, 
    isResizing, 
    resizingColumnId, 
    autoAdjustColumnWidth, 
    handleResizeStart 
  } = useTableColumnResize({
    tableId,
    defaultWidths: getDefaultWidths.value
  });

  const table = useVueTable({
    get data() {
      return items.value;
    },
    columns: tableColumns.value as ColumnDef<T>[],
    state: {
      get sorting() {
        return sortingState.value;
      },
      get columnSizing() {
        return columnWidths.value;
      }
    },
    getCoreRowModel: getCoreRowModel(),
    getSortedRowModel: getSortedRowModel(),
    onSortingChange: (updater) => {
      const newSorting = typeof updater === 'function' 
        ? updater(sortingState.value) 
        : (Array.isArray(updater) ? updater : []);
      sorting.value = newSorting;
    },
    onColumnSizingChange: (sizing) => {
      Object.entries(sizing).forEach(([colId, width]) => {
        columnWidths.value[colId] = width as number;
      });
    },
    enableColumnResizing: true,
    columnResizeMode: 'onChange',
    debugTable: false,
  });

  const rows = computed(() => table.getRowModel().rows);

  const tableContainerRef = ref<HTMLDivElement | null>(null);

  // 虚拟滚动配置和创建
  const rowVirtualizerOptions = computed(() => ({
    count: rows.value.length,
    estimateSize: () => estimateRowHeight,
    getScrollElement: () => tableContainerRef.value,
    overscan: 5,
  }));
  
  const rowVirtualizer = useVirtualizer(rowVirtualizerOptions);

  const virtualRows = computed(() => {
    if (!rowVirtualizer?.value) {
      return null;
    }
    return rowVirtualizer.value.getVirtualItems();
  });

  const totalSize = computed(() => {
    if (!rowVirtualizer?.value) {
      return 0;
    }
    return rowVirtualizer.value.getTotalSize();
  });

  const visibleRows = computed<Row<T>[]>(() => {
    if (!virtualRows.value) {
      return rows.value;
    }
    
    return virtualRows.value
      .map((vRow: any) => rows.value[vRow.index])
      .filter(Boolean);
  });

  const measureElement = (el: Element | ComponentPublicInstance | null) => {
    if (!rowVirtualizer?.value || !el) {
      return;
    }
    // 如果是组件实例，获取其 $el
    const element = (el as any)?.$el || el;
    if (element instanceof Element) {
      rowVirtualizer.value.measureElement(element);
    }
  };

  const getFlatHeaders = () => {
    return table.getFlatHeaders();
  };

  const getVisibleRows = (): Row<T>[] => {
    return table.getRowModel().rows;
  };

  const getRowStyle = (item: T) => {
    return {
      backgroundColor: item.color || 'inherit'
    };
  };
  
  const resetTableSorting = () => {
    resetSorting();
  };

  return {
    table,
    rows,
    visibleRows,
    virtualRows,
    totalSize,
    tableContainerRef,
    measureElement,
    enableVirtualization: true,
    columnWidths,
    isResizing,
    resizingColumnId,
    getRowStyle,
    autoAdjustColumnWidth,
    handleResizeStart,
    getFlatHeaders,
    getVisibleRows,
    hasSavedSorting,
    resetTableSorting
  };
} 