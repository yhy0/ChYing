<script setup lang="ts">
import { ref, watch } from 'vue';
import { RequestPanel } from '../common/requestResponse';
import type { RequestViewType } from '../../utils/viewerUtils';
import { useI18n } from 'vue-i18n';
import type { PayloadPosition } from '../../types/intruder';
import { useIntruderUtils } from './components/IntruderUtils';

// 初始化工具函数
const utils = useIntruderUtils();

// 初始化i18n
const { t } = useI18n();

// 定义props
const props = defineProps<{
  request: string;
  payloadMarker?: string;
}>();

// 定义emits
const emit = defineEmits<{
  (e: 'update:request', value: string): void;
  (e: 'payload-positions-changed', positions: PayloadPosition[]): void;
}>();

// 请求视图类型
const requestViewType = ref<RequestViewType>('pretty');

// 载荷标记符
const payloadMarker = props.payloadMarker || '§';

// 处理请求数据更新
const handleRequestUpdate = (data: string) => {
  // 直接发出更新整个请求的事件
  emit('update:request', data);
  
  // 提取并发出payload位置
  const positions = utils.extractPayloadPositions(data, payloadMarker);
  emit('payload-positions-changed', positions);
};

// 监听请求数据变化，提取payload位置
watch(() => props.request, (newData) => {
  const positions = utils.extractPayloadPositions(newData, payloadMarker);
  emit('payload-positions-changed', positions);
}, { immediate: true });
</script>

<template>
  <div class="w-full h-full border border-gray-200 dark:border-gray-700 rounded-md overflow-hidden" style="min-height: 350px; display: flex; flex-direction: column;">
    <RequestPanel
      :normalized-request-data="request"
      :request-width="100"
      :request-view-type="requestViewType"
      :request-title="t('common.ui.request')"
      @set-request-view-type="(type: string) => requestViewType = type as RequestViewType"
      @update:request-data="handleRequestUpdate"
      class="flex-1"
    />
  </div>
</template> 