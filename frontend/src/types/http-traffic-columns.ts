import type { HttpTrafficColumn } from './table';

/**
 * 默认的HTTP流量表格列配置
 */
export const defaultHttpTrafficColumns: HttpTrafficColumn[] = [
  { id: 'id', name: '#', width: 80 },
  { id: 'method', name: 'Method', width: 80 },
  { id: 'host', name: 'Host', width: 200 },
  { id: 'path', name: 'Path', width: 300 },
  { id: 'status', name: 'Status', width: 80 },
  { id: 'length', name: 'Length', width: 100 },
  { id: 'mimeType', name: 'MIME Type', width: 80 },
  { id: 'extension', name: 'Extension', width: 20 },
  { id: 'title', name: 'Title', width: 200 },
  { id: 'ip', name: 'IP', width: 120 },
  { id: 'note', name: 'Note', width: 100 },
  { id: 'time', name: 'Time', width: 100 }
]; 