// src/utils/eventBus.ts
import mitt from 'mitt';
import type { ProxyHistoryItem, RepeaterTab } from '../types'; // 根据实际类型调整路径

// 定义应用中会用到的事件及其负载类型
export interface SendToRepeaterPayload {
  sourceModule: string;
  requestData: string;
}

export interface SendToIntruderPayload {
  sourceModule: string;
  requestData: string;
  initialAttackType?: string; // More specific than generic targetDetails
}

export type ApplicationEvents = {
  sendToRepeaterRequested: { sourceItem: ProxyHistoryItem };
  sendToIntruderFromProxyRequested: { sourceItem: ProxyHistoryItem };
  sendToIntruderFromRepeaterRequested: { sourceItem: RepeaterTab };
  // 未来可以根据需要添加其他事件
};

// 创建并导出事件总线实例
const eventBus = mitt<ApplicationEvents>();
export default eventBus;

// 定义事件名称常量，以便在派发和监听时保持一致性，避免拼写错误
export const SEND_TO_REPEATER_REQUESTED = 'sendToRepeaterRequested';
export const SEND_TO_INTRUDER_FROM_PROXY_REQUESTED = 'sendToIntruderFromProxyRequested';
export const SEND_TO_INTRUDER_FROM_REPEATER_REQUESTED = 'sendToIntruderFromRepeaterRequested';