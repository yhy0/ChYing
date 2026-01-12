<script setup lang="ts">
import { ref } from 'vue';
import MessageComponent from './MessageComponent.vue';

interface MessageItem {
  id: string;
  type: 'info' | 'success' | 'warning' | 'error';
  message: string;
  duration: number;
}

const messages = ref<MessageItem[]>([]);
let messageIdCounter = 0;

const addMessage = (type: 'info' | 'success' | 'warning' | 'error', message: string, duration: number = 3) => {
  const id = `message_${Date.now()}_${messageIdCounter++}`;
  messages.value.push({ id, type, message, duration });
  return id;
};

const removeMessage = (id: string) => {
  const index = messages.value.findIndex(msg => msg.id === id);
  if (index !== -1) {
    messages.value.splice(index, 1);
  }
};

defineExpose({
  addMessage,
  removeMessage
});
</script>

<template>
  <div class="message-container">
    <message-component
      v-for="msg in messages"
      :key="msg.id"
      :id="msg.id"
      :type="msg.type"
      :message="msg.message"
      :duration="msg.duration"
      :onClose="removeMessage"
    />
  </div>
</template>

<style scoped>
.message-container {
  position: fixed;
  top: 20px;
  left: 50%;
  transform: translateX(-50%);
  z-index: 9999;
  display: flex;
  flex-direction: column;
  align-items: center;
  pointer-events: none;
}

.message-container :deep(.message-component) {
  pointer-events: auto;
}
</style> 