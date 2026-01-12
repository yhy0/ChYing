import { markRaw, isRef, type Ref } from 'vue';

/**
 * 对大量数据进行响应式优化
 * @param data 数据数组
 * @param threshold 阈值，超过此数量的数据将被优化
 * @returns 优化后的数据
 */
export function optimizeDataForRendering<T extends Record<string, any>>(data: T[], threshold = 1000): T[] {
  if (data.length <= threshold) {
    return data;
  }

  // 对大数据量使用markRaw减少响应式开销
  return data.map(item => markRaw(item) as T);
}

/**
 * 安全地获取数组数据，处理ref和普通数组
 * @param itemsRaw 可能是ref或普通数组的数据
 * @returns 数组数据
 */
export function safeGetArrayData<T>(itemsRaw: T[] | Ref<T[]>): T[] {
  const rawItems = isRef(itemsRaw) ? itemsRaw.value : itemsRaw;
  return rawItems || [];
}

/**
 * 性能监控装饰器
 * @param name 监控名称
 * @param fn 要监控的函数
 * @returns 装饰后的函数
 */
export function withPerformanceMonitor<T extends (...args: any[]) => any>(
  name: string,
  fn: T
): T {
  return ((...args: any[]) => {
    const start = performance.now();
    const result = fn(...args);
    const end = performance.now();
    
    if (end - start > 16) { // 超过一帧的时间
      console.warn(`[表格性能] ${name} 耗时: ${(end - start).toFixed(2)}ms`);
    }
    
    return result;
  }) as T;
}

/**
 * 节流函数，用于高频事件
 * @param fn 要节流的函数
 * @param delay 延迟时间
 * @returns 节流后的函数
 */
export function throttle<T extends (...args: any[]) => any>(
  fn: T,
  delay: number
): T {
  let lastCall = 0;
  return ((...args: any[]) => {
    const now = Date.now();
    if (now - lastCall >= delay) {
      lastCall = now;
      return fn(...args);
    }
  }) as T;
}

/**
 * 防抖函数，用于延迟执行
 * @param fn 要防抖的函数
 * @param delay 延迟时间
 * @returns 防抖后的函数
 */
export function debounce<T extends (...args: any[]) => any>(
  fn: T,
  delay: number
): T {
  let timeoutId: ReturnType<typeof setTimeout>;
  return ((...args: any[]) => {
    clearTimeout(timeoutId);
    timeoutId = setTimeout(() => fn(...args), delay);
  }) as T;
}

/**
 * 批量更新优化
 * @param updates 更新函数数组
 */
export function batchUpdates(updates: (() => void)[]) {
  // 使用 requestAnimationFrame 确保在下一帧执行
  requestAnimationFrame(() => {
    updates.forEach(update => update());
  });
}

/**
 * 内存使用监控
 */
export function logMemoryUsage(context: string) {
  if ('memory' in performance) {
    const memory = (performance as any).memory;
    console.log(`[内存监控] ${context}:`, {
      used: `${(memory.usedJSHeapSize / 1024 / 1024).toFixed(2)}MB`,
      total: `${(memory.totalJSHeapSize / 1024 / 1024).toFixed(2)}MB`,
      limit: `${(memory.jsHeapSizeLimit / 1024 / 1024).toFixed(2)}MB`
    });
  }
} 