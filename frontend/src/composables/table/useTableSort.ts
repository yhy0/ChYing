import { ref, type Ref, watch, onBeforeUnmount } from 'vue';
import type { SortingState, SortDirection } from '@tanstack/vue-table';

export interface UseTableSortOptions {
  initialSorting?: SortingState;
  localStorageKey?: string;
}

export function useTableSort(options: UseTableSortOptions = {}) {
  const {
    initialSorting = [],
    localStorageKey = 'http-traffic-table-sorting',
  } = options;

  // 从localStorage加载排序状态或使用初始值
  const loadSavedSorting = (): SortingState => {
    try {
      const saved = localStorage.getItem(localStorageKey);
      if (saved && saved !== 'undefined') {
        const parsed = JSON.parse(saved);
        return Array.isArray(parsed) ? parsed : initialSorting;
      }
      return initialSorting;
    } catch (e) {
      console.error('Error loading sorting from localStorage:', e);
      return initialSorting;
    }
  };

  // 初始化排序状态
  const sorting: Ref<SortingState> = ref(loadSavedSorting());

  // 保存排序状态到localStorage
  const saveSorting = (sortingState: SortingState) => {
    try {
      if (!Array.isArray(sortingState)) {
        sortingState = [];
      }
      localStorage.setItem(localStorageKey, JSON.stringify(sortingState));
    } catch (e) {
      console.error('Error saving sorting to localStorage:', e);
    }
  };

  // 设置排序
  const setSorting = (newSorting: SortingState) => {
    sorting.value = Array.isArray(newSorting) ? newSorting : [];
    saveSorting(sorting.value);
  };

  // 切换列的排序状态
  const toggleColumnSort = (columnId: string) => {
    const currentIndex = sorting.value.findIndex(sort => sort.id === columnId);
    
    if (currentIndex >= 0) {
      // 已有排序，切换方向或移除
      const currentSort = sorting.value[currentIndex];
      
      if (currentSort.desc) {
        // 降序 -> 移除
        setSorting(sorting.value.filter(sort => sort.id !== columnId));
      } else {
        // 升序 -> 降序
        const newSorting = [...sorting.value];
        newSorting[currentIndex] = { ...currentSort, desc: true };
        setSorting(newSorting);
      }
    } else {
      // 无排序 -> 添加升序
      setSorting([...sorting.value, { id: columnId, desc: false }]);
    }
  };

  // 获取当前列的排序方向
  const getSortDirection = (columnId: string): SortDirection | false => {
    const sort = sorting.value.find(sort => sort.id === columnId);
    return sort ? (sort.desc ? 'desc' : 'asc') : false;
  };

  // 重置排序状态
  const resetSorting = () => {
    setSorting(initialSorting);
  };

  // 监听排序变化，自动保存
  watch(sorting, (newSorting) => {
    saveSorting(newSorting);
  }, { deep: true });

  // 组件卸载前确保保存当前排序状态
  onBeforeUnmount(() => {
    saveSorting(sorting.value);
  });

  return {
    sorting,
    setSorting,
    toggleColumnSort,
    getSortDirection,
    resetSorting,
    savedSortingExists: () => !!localStorage.getItem(localStorageKey)
  };
} 