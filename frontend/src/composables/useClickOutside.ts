/**
 * 点击外部区域检测组合式函数
 * 
 * 提供可靠的外部点击检测功能，支持多种配置选项
 * 解决了直接使用 setTimeout 可能导致的不稳定问题
 */

import { ref, onMounted, onUnmounted, readonly, type Ref } from 'vue';

export interface ClickOutsideOptions {
  /** 是否启用检测 */
  enabled?: Ref<boolean> | boolean;
  /** 延迟时间（毫秒），用于避免立即触发 */
  delay?: number;
  /** 要忽略的元素选择器或元素引用 */
  ignore?: string[] | HTMLElement[] | Ref<HTMLElement | null>[];
  /** 是否检测 Escape 键 */
  detectEscape?: boolean;
  /** 事件类型 */
  eventType?: 'click' | 'mousedown' | 'mouseup';
}

/**
 * 点击外部区域检测钩子
 */
export function useClickOutside(
  target: Ref<HTMLElement | null>,
  callback: (event: Event) => void,
  options: ClickOutsideOptions = {}
) {
  const {
    enabled = true,
    delay = 0,
    ignore = [],
    detectEscape = true,
    eventType = 'click'
  } = options;

  const isEnabled = ref(typeof enabled === 'boolean' ? enabled : enabled.value);
  let timeoutId: number | null = null;
  let isInitialized = false;

  /**
   * 检查元素是否应该被忽略
   */
  const shouldIgnore = (element: HTMLElement): boolean => {
    return ignore.some(item => {
      if (typeof item === 'string') {
        // 选择器字符串
        return element.closest(item) !== null;
      } else if (item instanceof HTMLElement) {
        // HTML 元素
        return element === item || item.contains(element);
      } else if (item && 'value' in item) {
        // Ref<HTMLElement>
        const refElement = item.value;
        return refElement && (element === refElement || refElement.contains(element));
      }
      return false;
    });
  };

  /**
   * 处理点击事件
   */
  const handleClick = (event: Event) => {
    // 如果未启用或未初始化，直接返回
    if (!isEnabled.value || !isInitialized) return;

    const targetElement = target.value;
    if (!targetElement) return;

    const clickedElement = event.target as HTMLElement;
    if (!clickedElement) return;

    // 检查是否点击在目标元素内部
    if (targetElement.contains(clickedElement)) return;

    // 检查是否点击在忽略的元素上
    if (shouldIgnore(clickedElement)) return;

    // 如果有延迟，使用 setTimeout
    if (delay > 0) {
      if (timeoutId) {
        clearTimeout(timeoutId);
      }
      timeoutId = window.setTimeout(() => {
        callback(event);
        timeoutId = null;
      }, delay);
    } else {
      callback(event);
    }
  };

  /**
   * 处理键盘事件（Escape 键）
   */
  const handleKeydown = (event: KeyboardEvent) => {
    if (!isEnabled.value || !detectEscape) return;
    
    if (event.key === 'Escape') {
      callback(event);
    }
  };

  /**
   * 启用检测
   */
  const enable = () => {
    isEnabled.value = true;
  };

  /**
   * 禁用检测
   */
  const disable = () => {
    isEnabled.value = false;
    if (timeoutId) {
      clearTimeout(timeoutId);
      timeoutId = null;
    }
  };

  /**
   * 初始化事件监听器
   */
  const initialize = () => {
    if (isInitialized) return;
    
    // 使用 setTimeout 确保在下一个事件循环中初始化
    // 这样可以避免立即触发的问题
    setTimeout(() => {
      document.addEventListener(eventType, handleClick, true);
      if (detectEscape) {
        document.addEventListener('keydown', handleKeydown, true);
      }
      isInitialized = true;
    }, 0);
  };

  /**
   * 清理事件监听器
   */
  const cleanup = () => {
    if (!isInitialized) return;
    
    document.removeEventListener(eventType, handleClick, true);
    if (detectEscape) {
      document.removeEventListener('keydown', handleKeydown, true);
    }
    
    if (timeoutId) {
      clearTimeout(timeoutId);
      timeoutId = null;
    }
    
    isInitialized = false;
  };

  // 监听 enabled 状态变化
  if (typeof enabled !== 'boolean' && 'value' in enabled) {
    // 如果 enabled 是 ref，监听其变化
    const stopWatcher = () => {
      // 这里可以添加 watch 逻辑，但为了避免依赖 Vue 的 watch
      // 我们使用简单的响应式更新
    };
  }

  // 组件挂载时初始化
  onMounted(() => {
    initialize();
  });

  // 组件卸载时清理
  onUnmounted(() => {
    cleanup();
  });

  return {
    isEnabled: readonly(isEnabled),
    enable,
    disable,
    initialize,
    cleanup
  };
}

/**
 * 简化版本的点击外部检测
 * 适用于大多数常见场景
 */
export function useSimpleClickOutside(
  target: Ref<HTMLElement | null>,
  callback: () => void,
  enabled: Ref<boolean> | boolean = true
) {
  return useClickOutside(target, callback, {
    enabled,
    delay: 0,
    detectEscape: true
  });
}
