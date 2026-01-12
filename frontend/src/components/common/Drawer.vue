<script setup>
import { ref, watch, computed } from 'vue';
import { useI18n } from 'vue-i18n';
import { useBodyScrollLock } from '@/composables/useBodyScrollLock';

const { t } = useI18n();

const props = defineProps({
  show: {
    type: Boolean,
    default: false
  },
  title: {
    type: String,
    default: ''
  },
  defaultWidth: {
    type: [Number, String],
    default: '75%'
  },
  placement: {
    type: String,
    default: 'right',
    validator: (value) => ['left', 'right', 'top', 'bottom'].includes(value)
  }
});

const emit = defineEmits(['update:show']);

// 使用身体滚动锁定
const { lock, unlock } = useBodyScrollLock();

// 计算样式
const drawerStyle = computed(() => {
  if (props.placement === 'left' || props.placement === 'right') {
    return { width: typeof props.defaultWidth === 'number' ? `${props.defaultWidth}px` : props.defaultWidth };
  } else {
    return { height: typeof props.defaultWidth === 'number' ? `${props.defaultWidth}px` : props.defaultWidth };
  }
});

// 计算变换样式
const transformClass = computed(() => {
  switch (props.placement) {
    case 'left':
      return props.show ? 'translate-x-0' : '-translate-x-full';
    case 'right':
      return props.show ? 'translate-x-0' : 'translate-x-full';
    case 'top':
      return props.show ? 'translate-y-0' : '-translate-y-full';
    case 'bottom':
      return props.show ? 'translate-y-0' : 'translate-y-full';
    default:
      return '';
  }
});

// 计算位置样式
const placementClass = computed(() => {
  switch (props.placement) {
    case 'left':
      return 'top-0 left-0 h-full';
    case 'right':
      return 'top-0 right-0 h-full';
    case 'top':
      return 'top-0 left-0 w-full';
    case 'bottom':
      return 'bottom-0 left-0 w-full';
    default:
      return '';
  }
});

// 监听props.show的变化，控制滚动锁定
watch(() => props.show, (newVal) => {
  if (newVal) {
    lock();
  } else {
    unlock();
  }
});

// 关闭抽屉
const closeDrawer = () => {
  emit('update:show', false);
};
</script>

<template>
  <teleport to="body">
    <!-- 背景蒙层 -->
    <div 
      v-if="show"
      class="fixed inset-0 bg-gray-900 bg-opacity-50 dark:bg-opacity-70 z-40 transition-opacity duration-300 pointer-events-auto"
      @click="closeDrawer"
    ></div>

    <!-- 抽屉容器 -->
    <div 
      v-if="show"
      :class="[
        'fixed z-50 bg-white dark:bg-[#1e1e2e] shadow-xl transform transition-transform duration-300 ease-in-out flex flex-col',
        placementClass,
        transformClass
      ]"
      :style="drawerStyle"
    >
      <!-- 抽屉标题栏 -->
      <div class="sticky top-0 left-0 right-0 z-[60] bg-white dark:bg-[#1e1e36] border-b border-gray-200 dark:border-gray-700 flex-shrink-0 shadow-sm">
        <div class="p-4 flex items-center justify-between">
          <h2 class="text-lg font-medium text-gray-800 dark:text-gray-200">
            {{ title }}
          </h2>
          <button 
            @click="closeDrawer"
            class="p-1.5 rounded-full text-gray-500 hover:text-red-500 hover:bg-gray-100 dark:hover:bg-gray-800 transition-colors"
          >
            <i class="bx bx-x text-xl"></i>
          </button>
        </div>
      </div>

      <!-- 抽屉内容 -->
      <div class="flex-grow overflow-auto relative">
        <div class="p-4 pt-5">
          <slot></slot>
        </div>
      </div>
    </div>
  </teleport>
</template>

<style scoped>
/* 滚动锁定样式已移至 useBodyScrollLock 组合式函数中统一管理 */

/* 抽屉打开时阻止主内容区域的交互 */
.fixed.inset-0 {
  pointer-events: auto;
  user-select: none;
}
</style> 