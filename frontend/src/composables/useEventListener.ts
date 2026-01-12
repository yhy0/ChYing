import { onMounted, onBeforeUnmount } from 'vue';

/**
 * 通用事件监听组合式函数
 * 在组件挂载时添加事件监听，在组件卸载时自动移除事件监听
 * 
 * @param target 事件目标（如window, document等）
 * @param event 事件名称（如'click', 'keydown'等）
 * @param handler 事件处理函数
 * @param options 事件监听选项
 */
export function useEventListener(
  target: EventTarget,
  event: string, 
  handler: Function, 
  options?: AddEventListenerOptions
) {
  onMounted(() => {
    target.addEventListener(event, handler as EventListener, options);
  });
  
  onBeforeUnmount(() => {
    target.removeEventListener(event, handler as EventListener, options);
  });
} 