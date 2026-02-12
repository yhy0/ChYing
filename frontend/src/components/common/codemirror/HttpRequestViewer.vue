<script setup lang="ts">
import HttpViewer from './HttpViewer.vue';

// 修改类型定义为字符串类型
type RequestData = string;

const props = defineProps<{
  data: RequestData;
  readOnly?: boolean;
  bodyFormatted?: boolean;
}>();

const emit = defineEmits<{
  (e: 'update:data', data: RequestData): void;
}>();
</script>

<template>
  <div class="http-request-viewer h-full">
    <HttpViewer 
      :data="props.data"
      :read-only="props.readOnly"
      :is-response="false"
      :body-formatted="props.bodyFormatted"
      @update:data="emit('update:data', $event)"
    />
  </div>
</template> 