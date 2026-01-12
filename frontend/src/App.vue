<script setup lang="ts">
import { onMounted, onUnmounted } from 'vue';
import { useRouter } from 'vue-router';
import 'boxicons/css/boxicons.min.css';
import { useModulesStore } from './store/modules';
import eventBus, { 
  SEND_TO_REPEATER_REQUESTED, 
  SEND_TO_INTRUDER_FROM_PROXY_REQUESTED, 
  SEND_TO_INTRUDER_FROM_REPEATER_REQUESTED 
} from './utils/eventBus';

const router = useRouter();
const store = useModulesStore();

// Event Bus Handler Functions
const handleSendToRepeater = (payload: { sourceItem: import('./types').ProxyHistoryItem }) => {
  console.log('App.vue: Event SEND_TO_REPEATER_REQUESTED received', payload);
  const newTabId = store.addRepeaterTabFromEventPayload(payload); 
  if (newTabId) {
    router.push('/app/repeater');
  }
};

const handleSendToIntruder = (payload: { sourceItem: import('./types').ProxyHistoryItem | import('./types').RepeaterTab }) => {
  console.log('App.vue: Event SEND_TO_INTRUDER_REQUESTED received', payload);
  const newTabId = store.addIntruderTabFromEventPayload(payload); 
  if (newTabId) {
    router.push('/app/intruder');
  }
};

// Event Bus Listeners Setup
onMounted(() => {
  // 初始化 store 中的通知状态
  store.setUnreadCount?.(0);

  console.log('App.vue: Mounting and setting up event listeners.');
  eventBus.on(SEND_TO_REPEATER_REQUESTED, handleSendToRepeater);
  eventBus.on(SEND_TO_INTRUDER_FROM_PROXY_REQUESTED, handleSendToIntruder);
  eventBus.on(SEND_TO_INTRUDER_FROM_REPEATER_REQUESTED, handleSendToIntruder); 
});

onUnmounted(() => {
  console.log('App.vue: Unmounting and tearing down event listeners.');
  eventBus.off(SEND_TO_REPEATER_REQUESTED, handleSendToRepeater);
  eventBus.off(SEND_TO_INTRUDER_FROM_PROXY_REQUESTED, handleSendToIntruder);
  eventBus.off(SEND_TO_INTRUDER_FROM_REPEATER_REQUESTED, handleSendToIntruder);
});
</script>

<template>
  <div class="app-container">
    <RouterView />
  </div>
</template>

<style scoped>
.app-container {
  height: 100vh;
  width: 100vw;
  overflow: auto; /* 允许滚动 */
}
</style>
