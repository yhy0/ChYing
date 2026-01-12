/**
 * Repeater 标签页接口
 */
export interface RepeaterTab {
  id: string;
  name: string;
  color: string;
  groupId: string | null;
  request: string;
  response: string | null;
  isActive: boolean;
  isRunning: boolean;
  modified: boolean;
  serverDurationMs: number;
  method?: string;
  url?: string;
}

/**
 * Repeater 分组接口
 */
export interface RepeaterGroup {
  id: string;
  name: string;
  color: string;
} 