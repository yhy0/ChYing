<script setup lang="ts">
import HttpViewer from './HttpViewer.vue';

// 修改类型定义为字符串类型
type ResponseData = string;

const props = defineProps<{
  data: ResponseData;
  readOnly?: boolean;
}>();

const emit = defineEmits<{
  (e: 'update:data', data: ResponseData): void;
}>();

</script>

<template>
  <div class="http-response-viewer h-full">
    <HttpViewer 
      :data="props.data"
      :read-only="props.readOnly"
      :is-response="true"
      @update:data="emit('update:data', $event)"
    />
  </div>
</template> 