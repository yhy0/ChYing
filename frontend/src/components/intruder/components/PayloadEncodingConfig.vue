<script setup lang="ts">
import { ref, computed } from 'vue';
import { useI18n } from 'vue-i18n';

const { t } = useI18n();

// 默认编码配置
const defaultEncoding = {
  enabled: false,
  urlEncode: false,
  characterSet: 'UTF-8'
};

// 更新 props 定义
const props = defineProps<{
  encoding?: {
    enabled: boolean;
    urlEncode: boolean;
    characterSet: string;
  };
}>();

// 更新事件定义
const emit = defineEmits<{
  (e: 'update:encoding', value: { enabled: boolean; urlEncode: boolean; characterSet: string }): void;
}>();

// 安全获取 encoding 属性，确保不会为 undefined
const safeEncoding = computed(() => {
  return props.encoding || defaultEncoding;
});

// URL 编码复选框的双向绑定
const urlEncodeChecked = computed({
  get: () => safeEncoding.value.urlEncode,
  set: (value: boolean) => {
    emit('update:encoding', {
      ...safeEncoding.value,
      urlEncode: value
    });
  }
});

// 面板折叠状态
const encodingPanelOpen = ref(true);

// 切换面板折叠状态
const togglePanel = () => {
  encodingPanelOpen.value = !encodingPanelOpen.value;
};

// 处理字符集变化
const handleCharacterSetChange = (characterSet: string) => {
  emit('update:encoding', {
    ...safeEncoding.value,
    characterSet
  });
};
</script>

<template>
  <div class="border border-gray-200 dark:border-gray-700 rounded-md mb-2">
    <!-- 编码设置面板标题 -->
    <div 
      class="flex items-center justify-between p-2 bg-[#f3f4f6] dark:bg-[#282838] cursor-pointer" 
      @click="togglePanel"
    >
      <h4 class="text-sm font-medium text-gray-700 dark:text-gray-300">
        {{ t('modules.intruder.payload_encoding') }}
      </h4>
      <i :class="[encodingPanelOpen ? 'bx bx-chevron-up' : 'bx bx-chevron-down']" class="text-gray-500"></i>
    </div>
    
    <!-- 编码设置内容 -->
    <div class="p-2" v-show="encodingPanelOpen">
      <p class="mb-2 text-xs text-gray-600 dark:text-gray-400">
        {{ t('modules.intruder.payload_encoding_description') }}
      </p>
      
      <div class="flex items-start mb-2">
        <div class="flex items-center h-5">
          <input
            spellcheck="false"
            type="checkbox"
            id="url-encode-checkbox"
            v-model="urlEncodeChecked"
            class="focus:ring-indigo-500 h-4 w-4 text-indigo-600 border-gray-300 rounded"
          >
        </div>
        <div class="ml-2">
          <label for="url-encode-checkbox" class="text-xs text-gray-700 dark:text-gray-300">
            {{ t('modules.intruder.url_encode_chars') }}
          </label>
        </div>
      </div>
      
      <div class="mt-1">
        <input 
          type="text" 
          :value="safeEncoding.characterSet" 
          :disabled="!safeEncoding.urlEncode"
          @input="handleCharacterSetChange(($event.target as HTMLInputElement).value)"
          class="w-full text-xs border border-gray-300 dark:border-gray-600 rounded bg-white dark:bg-[#1e1e2e] px-2 py-1"
          :class="{ 'opacity-50': !safeEncoding.urlEncode }"
          :placeholder="t('modules.intruder.url_encode_chars_placeholder')"
          spellcheck="false"
        >
        <p v-if="safeEncoding.urlEncode" class="mt-1 text-xs text-gray-500 dark:text-gray-400">
          {{ t('modules.intruder.url_encode_chars_help') }}
        </p>
      </div>
    </div>
  </div>
</template>