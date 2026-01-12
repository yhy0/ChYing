/**
 * 身体滚动锁定组合式函数
 * 
 * 用于安全地管理多个组件对 body 滚动的控制
 * 使用引用计数确保只有在所有组件都解锁时才真正解锁滚动
 */

import { ref, onUnmounted, readonly } from 'vue';

// 全局引用计数器
let lockCount = 0;
const isLocked = ref(false);

/**
 * 身体滚动锁定钩子
 * @returns 锁定和解锁函数
 */
export function useBodyScrollLock() {
  let isCurrentlyLocked = false;

  /**
   * 锁定身体滚动
   */
  const lock = () => {
    if (isCurrentlyLocked) return;
    
    lockCount++;
    isCurrentlyLocked = true;
    
    if (lockCount === 1) {
      // 第一次锁定时，保存当前滚动位置并锁定
      const scrollY = window.scrollY;
      document.body.style.position = 'fixed';
      document.body.style.top = `-${scrollY}px`;
      document.body.style.width = '100%';
      document.body.classList.add('overflow-hidden');
      isLocked.value = true;
    }
  };

  /**
   * 解锁身体滚动
   */
  const unlock = () => {
    if (!isCurrentlyLocked) return;
    
    lockCount--;
    isCurrentlyLocked = false;
    
    if (lockCount === 0) {
      // 最后一个解锁时，恢复滚动位置
      const scrollY = document.body.style.top;
      document.body.style.position = '';
      document.body.style.top = '';
      document.body.style.width = '';
      document.body.classList.remove('overflow-hidden');
      
      if (scrollY) {
        window.scrollTo(0, parseInt(scrollY || '0') * -1);
      }
      
      isLocked.value = false;
    }
  };

  /**
   * 切换锁定状态
   */
  const toggle = () => {
    if (isCurrentlyLocked) {
      unlock();
    } else {
      lock();
    }
  };

  // 组件卸载时自动解锁
  onUnmounted(() => {
    if (isCurrentlyLocked) {
      unlock();
    }
  });

  return {
    lock,
    unlock,
    toggle,
    isLocked: readonly(isLocked),
    isCurrentlyLocked: () => isCurrentlyLocked
  };
}

/**
 * 获取全局锁定状态
 */
export function useBodyScrollLockState() {
  return {
    isLocked: readonly(isLocked),
    lockCount: () => lockCount
  };
}
