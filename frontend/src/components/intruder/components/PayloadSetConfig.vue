<script setup lang="ts">
import { ref, computed } from 'vue';
import { useI18n } from 'vue-i18n';

const { t } = useI18n();

// Props定义
const props = defineProps<{
  // 替换原有的 payloadSets 数组为单个 payloadSet
  payloadSet: {
    id: number;
    type: 'simple-list' | 'numbers' | 'brute-force' | 'custom';
    items: string[];
    processing: {
      rules: Array<{
        type: string;
        config: Record<string, any>;
      }>;
      encoding: {
        enabled: boolean;
        urlEncode: boolean;
        characterSet: string;
      };
    };
  };
  // 添加对应的 payload 位置信息
  payloadPosition: {
    start: number;
    end: number;
    value: string;
    paramName?: string;
    index: number;  // 添加 index 属性
  };
}>();

// Emits定义
const emit = defineEmits<{
  (e: 'update:payloadSet', value: any): void;
  (e: 'add-payload-item', value: string, index: number): void;
  (e: 'paste-payload-items', items: string[], index: number): void;
  (e: 'load-payload-items', items: string[], index: number): void;
  (e: 'remove-payload-item', index: number, payloadIndex: number): void;
  (e: 'clear-payload-items', index: number): void;
}>();

// 载荷输入
const newPayloadItem = ref('');
const selectedPayloadIndex = ref(-1);

// 面板状态
const configPanelOpen = ref(true);

// 切换面板折叠状态
const togglePanel = () => {
  configPanelOpen.value = !configPanelOpen.value;
};

// 载荷类型选项 - 使用 computed 确保 i18n 已初始化
const payloadTypes = computed(() => [
  { value: 'simple-list', label: t('modules.intruder.payload_type_simple_list') },
  { value: 'numbers', label: t('modules.intruder.payload_type_numbers') },
  { value: 'brute-force', label: t('modules.intruder.payload_type_brute_force') },
  { value: 'custom', label: t('modules.intruder.payload_type_custom') }
]);

// 修改获取当前载荷集的computed
const currentPayloadSet = computed(() => {
  return props.payloadSet || {
    id: Date.now(),
    type: 'simple-list',
    items: [],
    processing: {
      rules: [],
      encoding: {
        enabled: false,
        urlEncode: false,
        characterSet: 'UTF-8'
      }
    }
  };
});

// 更新当前载荷集类型
const updatePayloadType = (event: Event) => {
  const value = (event.target as HTMLSelectElement).value;
  emit('update:payloadSet', { 
    ...currentPayloadSet.value,
    type: value as any 
  });
};

// 添加新的载荷项
const addPayloadItem = () => {
  if (newPayloadItem.value.trim()) {
    emit('add-payload-item', newPayloadItem.value.trim(), props.payloadPosition.index);
    newPayloadItem.value = '';
  }
};

// 粘贴载荷项
const pastePayloadItems = async () => {
  try {
    const clipboardText = await navigator.clipboard.readText();
    const items = clipboardText.split('\n')
      .map(line => line.trim())
      .filter(line => line.length > 0);
    
    if (items.length > 0) {
      emit('paste-payload-items', items, props.payloadPosition.index);
    }
  } catch (error) {
    console.error('Failed to read clipboard: ', error);
  }
};

// 加载载荷项
const loadPayloadItems = () => {
  // 创建一个隐藏的文件输入元素
  const fileInput = document.createElement('input');
  fileInput.type = 'file';
  fileInput.accept = '.txt';
  fileInput.style.display = 'none';
  
  fileInput.onchange = (event) => {
    const file = (event.target as HTMLInputElement).files?.[0];
    if (file) {
      const reader = new FileReader();
      reader.onload = (e) => {
        const content = e.target?.result as string;
        const items = content.split('\n')
          .map(line => line.trim())
          .filter(line => line.length > 0);
        
        if (items.length > 0) {
          emit('load-payload-items', items, props.payloadPosition.index);
        }
      };
      reader.readAsText(file);
    }
    
    // 移除元素
    document.body.removeChild(fileInput);
  };
  
  document.body.appendChild(fileInput);
  fileInput.click();
};

// 移除选中的载荷项
const removeSelectedPayloadItem = () => {
  if (selectedPayloadIndex.value >= 0) {
    emit('remove-payload-item', selectedPayloadIndex.value, props.payloadPosition.index);
    selectedPayloadIndex.value = -1;
  }
};

// 清空所有载荷项
const clearPayloadItems = () => {
  emit('clear-payload-items', props.payloadPosition.index);
  selectedPayloadIndex.value = -1;
};

// 去重载荷项
const deduplicatePayloadItems = () => {
  if (props.payloadSet.items.length > 0) {
    const uniqueItems = [...new Set(props.payloadSet.items)];
    emit('update:payloadSet', {
      ...props.payloadSet,
      items: uniqueItems
    });
  }
};

// 定位信息文本
const positionsInfo = computed(() => {
  if (!props.payloadPosition) {
    return t('modules.intruder.no_payload_positions');
  }
  
  // 获取位置索引（从 payloadPosition 中获取）
  const positionNumber = props.payloadPosition.paramName 
    ? `${t('modules.intruder.position_number', { number: props.payloadPosition.index + 1 })} - ${props.payloadPosition.paramName}`
    : t('modules.intruder.position_number', { number: props.payloadPosition.index + 1 });

  return `${positionNumber} (${props.payloadPosition.value})`;
});

// 载荷项计数
const payloadItemsCount = computed(() => {
  return currentPayloadSet.value?.items?.length || 0;
});

// 检查是否有载荷项
const hasPayloadItems = computed(() => {
  return (currentPayloadSet.value?.items?.length || 0) > 0;
});
</script>

<template>
  <div class="border border-gray-200 dark:border-gray-700 rounded-md mb-2">
    <!-- 配置面板标题 -->
    <div 
      class="flex items-center justify-between p-2 bg-[#f3f4f6] dark:bg-[#282838] cursor-pointer" 
      @click="togglePanel"
    >
      <h4 class="text-sm font-medium text-gray-700 dark:text-gray-300">
        {{ payloadPosition?.paramName 
          ? `${t('modules.intruder.payload_configuration')} - ${payloadPosition.paramName}`
          : t('modules.intruder.payload_configuration')
        }}
      </h4>
      <i :class="[configPanelOpen ? 'bx bx-chevron-up' : 'bx bx-chevron-down']" class="text-gray-500"></i>
    </div>
    
    <!-- 配置面板内容 -->
    <div class="p-2" v-show="configPanelOpen">
      
     <!-- 有效载荷位置信息 -->
      <div class="mb-4">
        <h4 class="text-sm font-medium mb-1">{{ t('modules.intruder.payload_positions') }}</h4>
        <div class="text-xs text-gray-600 dark:text-gray-400">
          {{ positionsInfo }}
        </div>
      </div> 
      
      <!-- 有效载荷类型选择 -->
      <div class="mb-4">
        <h4 class="text-sm font-medium mb-1">{{ t('modules.intruder.payload_type') }}</h4>
        <select 
          :value="currentPayloadSet.type"
          @change="updatePayloadType"
          class="w-full px-3 py-1.5 text-sm border border-gray-300 dark:border-gray-600 rounded-md bg-white dark:bg-[#282838] text-gray-700 dark:text-gray-300"
        >
          <option v-for="option in payloadTypes" :key="option.value" :value="option.value">
            {{ option.label }}
          </option>
        </select>
      </div>
      
      <!-- 载荷列表 -->
      <div class="mb-4">
        <div class="flex items-center justify-between mb-1">
          <h4 class="text-sm font-medium">{{ t('modules.intruder.payloads') }}</h4>
          <div class="text-xs text-gray-500 dark:text-gray-400">
            {{ payloadItemsCount }} {{ t('modules.project.issues.issues') }}
          </div>
        </div>
        
        <div class="flex gap-2 mb-2">
          <div class="flex flex-col gap-1">
            <button 
              type="button" 
              class="px-2 py-1 text-xs border border-gray-300 dark:border-gray-600 rounded bg-white dark:bg-[#282838] text-gray-700 dark:text-gray-300 hover:bg-gray-100 dark:hover:bg-[#32324c]"
              @click="pastePayloadItems"
            >
              {{ t('modules.intruder.paste') }}
            </button>
            <button 
              type="button" 
              class="px-2 py-1 text-xs border border-gray-300 dark:border-gray-600 rounded bg-white dark:bg-[#282838] text-gray-700 dark:text-gray-300 hover:bg-gray-100 dark:hover:bg-[#32324c]"
              @click="loadPayloadItems"
            >
              {{ t('modules.intruder.load') }}
            </button>
            <button 
              type="button" 
              class="px-2 py-1 text-xs border border-gray-300 dark:border-gray-600 rounded bg-white dark:bg-[#282838] text-gray-700 dark:text-gray-300 hover:bg-gray-100 dark:hover:bg-[#32324c]"
              @click="removeSelectedPayloadItem"
              :disabled="selectedPayloadIndex < 0"
              :class="selectedPayloadIndex < 0 ? 'opacity-50 cursor-not-allowed' : ''"
            >
              {{ t('modules.intruder.remove') }}
            </button>
            <button 
              type="button" 
              class="px-2 py-1 text-xs border border-gray-300 dark:border-gray-600 rounded bg-white dark:bg-[#282838] text-gray-700 dark:text-gray-300 hover:bg-gray-100 dark:hover:bg-[#32324c]"
              @click="clearPayloadItems"
              :disabled="!hasPayloadItems"
              :class="!hasPayloadItems ? 'opacity-50 cursor-not-allowed' : ''"
            >
              {{ t('common.actions.clear') }}
            </button>
            <button 
              type="button" 
              class="px-2 py-1 text-xs border border-gray-300 dark:border-gray-600 rounded bg-white dark:bg-[#282838] text-gray-700 dark:text-gray-300 hover:bg-gray-100 dark:hover:bg-[#32324c]"
              @click="deduplicatePayloadItems"
              :disabled="!hasPayloadItems"
              :class="!hasPayloadItems ? 'opacity-50 cursor-not-allowed' : ''"
            >
              {{ t('modules.intruder.deduplicate') }}
            </button>
          </div>
          
          <div class="flex-1 border border-gray-300 dark:border-gray-600 rounded bg-white dark:bg-[#1e1e2e] overflow-auto max-h-40">
            <ul class="text-xs text-gray-700 dark:text-gray-300 divide-y divide-gray-200 dark:divide-gray-700">
              <template v-if="currentPayloadSet?.items?.length">
                <li 
                  v-for="(item, index) in currentPayloadSet.items" 
                  :key="index"
                  :class="[
                    'p-1.5 cursor-pointer hover:bg-gray-100 dark:hover:bg-[#32324c]', 
                    selectedPayloadIndex === index ? 'bg-blue-50 dark:bg-blue-900/20' : ''
                  ]"
                  @click="selectedPayloadIndex = index"
                >
                  {{ item }}
                </li>
              </template>
              <li v-else class="p-2 text-gray-500 dark:text-gray-400 italic text-center">
                {{ t('modules.intruder.no_items') }}
              </li>
            </ul>
          </div>
        </div>
        
        <div class="flex gap-2">
          <button 
            type="button" 
            class="px-2 py-1 text-xs border border-gray-300 dark:border-gray-600 rounded bg-white dark:bg-[#282838] text-gray-700 dark:text-gray-300 hover:bg-gray-100 dark:hover:bg-[#32324c]"
            @click="addPayloadItem"
            :disabled="!newPayloadItem.trim()"
            :class="!newPayloadItem.trim() ? 'opacity-50 cursor-not-allowed' : ''"
          >
            {{ t('modules.intruder.add') }}
          </button>
          <input 
            type="text" 
            v-model="newPayloadItem" 
            class="flex-1 text-xs border border-gray-300 dark:border-gray-600 rounded bg-white dark:bg-[#1e1e2e] px-2 py-1 text-gray-700 dark:text-gray-300" 
            :placeholder="t('modules.intruder.enter_new_item')"
            @keyup.enter="addPayloadItem"
            spellcheck="false"
          >
        </div>
        
        <div class="mt-2">
          <button 
            class="w-full px-2 py-1 text-xs border border-gray-300 dark:border-gray-600 rounded bg-white dark:bg-[#282838] text-gray-700 dark:text-gray-300 text-left flex justify-between items-center hover:bg-gray-100 dark:hover:bg-[#32324c]"
            @click="loadPayloadItems"
          >
            <span>{{ t('modules.intruder.add_from_list') }}</span>
            <i class="bx bx-chevron-down"></i>
          </button>
        </div>
      </div>
    </div>
  </div>
</template>