<script setup lang="ts">
import { ref, onMounted, onBeforeUnmount } from 'vue';

const props = defineProps({
  id: {
    type: String,
    required: true
  },
  type: {
    type: String,
    default: 'info',
    validator: (value: string) => ['info', 'success', 'warning', 'error'].includes(value)
  },
  message: {
    type: String,
    required: true
  },
  duration: {
    type: Number,
    default: 3 // 默认3秒后消失，-1表示不自动消失
  },
  onClose: {
    type: Function,
    required: true
  }
});

const visible = ref(true);
let timer: number | null = null;

const close = () => {
  visible.value = false;
  setTimeout(() => {
    props.onClose(props.id);
  }, 300); // 动画结束后再移除DOM
};

const getIconClass = () => {
  switch (props.type) {
    case 'success': return 'bx bx-check-circle';
    case 'warning': return 'bx bx-error';
    case 'error': return 'bx bx-x-circle';
    default: return 'bx bx-info-circle';
  }
};

onMounted(() => {
  if (props.duration > 0) {
    timer = window.setTimeout(() => {
      close();
    }, props.duration * 1000);
  }
});

onBeforeUnmount(() => {
  if (timer) {
    clearTimeout(timer);
  }
});
</script>

<template>
  <transition name="message-fade">
    <div v-if="visible" :class="['message-component', `message-${type}`]">
      <i :class="getIconClass()" class="message-icon"></i>
      <span class="message-content">{{ message }}</span>
      <i class="message-close bx bx-x" @click="close"></i>
    </div>
  </transition>
</template>

<!-- 样式已移至 notification.css 统一管理 -->