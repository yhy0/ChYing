/**
 * 通知严重性等级
 */
export type NotificationSeverity = 'info' | 'warning' | 'error' | 'success';

/**
 * 通知项接口
 */
export interface NotificationItem {
  id: string;
  title: string;
  message: string;
  timestamp: string;
  read: boolean;
  severity: NotificationSeverity;
}

/**
 * 全局通知状态
 */
export interface NotificationState {
  showNotifications: boolean;
  unreadCount: number;
  items?: NotificationItem[];
} 