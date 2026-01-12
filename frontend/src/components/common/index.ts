// 从codemirror目录导入组件及类型
import { 
  HttpRequestViewer, 
  HttpResponseViewer, 
  HttpViewer,
  type RequestData,
  type ResponseData,
} from './codemirror';

// 导入各个组件
import RequestResponsePanel from './RequestResponsePanel.vue';
import HttpTrafficTable from './HttpTrafficTable.vue';
import BaseTabs from './BaseTabs.vue';
import ColorPicker from './ColorPicker.vue';
import ErrorBoundary from './ErrorBoundary.vue';
import LoadingState from './LoadingState.vue';

// 导入和重新导出表格相关类型
import type { HttpTrafficItem, HttpTrafficColumn } from '../../types';
export type { HttpTrafficItem, HttpTrafficColumn };

// 将codemirror作为命名空间导出
import * as cm from './codemirror';
export const codemirror = cm;

// 导出所有组件和类型
export {
  HttpRequestViewer,
  HttpResponseViewer, 
  HttpViewer,
  RequestResponsePanel,
  HttpTrafficTable,
  BaseTabs,
  ColorPicker,
  ErrorBoundary,
  LoadingState,
  type RequestData,
  type ResponseData,
}; 