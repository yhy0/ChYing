import type { HttpTrafficItem } from './http';
import type { VNode } from 'vue';

/**
 * 表格列基本配置接口
 */
export interface TableColumn {
  id: string;
  name: string;
  width: number;
}

/**
 * HTTP流量表格列配置
 * 使用泛型 T 来表示行数据的类型
 */
export interface HttpTrafficColumn<T = HttpTrafficItem> extends TableColumn {
  cellRenderer?: ({ item }: { item: T }) => VNode;
}

/**
 * 表格排序方向类型
 */
export type SortDirection = 'asc' | 'desc' | false;

/**
 * 表格排序状态
 */
export interface TableSortState {
  id: string;
  desc: boolean;
}

/**
 * 颜色选项接口
 */
export interface ColorOption {
  id: string;
  color: string;
} 