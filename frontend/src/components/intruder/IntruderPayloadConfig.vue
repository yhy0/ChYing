<script setup lang="ts">
import { ref, computed, watch } from 'vue';
import { useI18n } from 'vue-i18n';
import PayloadSetConfig from './components/PayloadSetConfig.vue';
import PayloadProcessingRules from './components/PayloadProcessingRules.vue';
import PayloadEncodingConfig from './components/PayloadEncodingConfig.vue';

const { t } = useI18n();

// 定义PayloadSet的类型 - 统一使用一种类型，避免兼容性问题
type PayloadSetType = 'simple-list' | 'numbers' | 'brute-force' | 'custom';

// 更新接口定义
interface ProcessingRule {
  type: string;
  config: Record<string, any>;
}

interface EncodingConfig {
  enabled: boolean;
  urlEncode: boolean;
  characterSet: string;
}

interface ProcessingConfig {
  rules: ProcessingRule[];
  encoding: EncodingConfig;
}

// 单一 PayloadSet 接口
interface PayloadSet {
  id: number;
  type: PayloadSetType;
  items: string[];
  processing: ProcessingConfig;
}

// 定义属性
const props = defineProps<{
  payloadSets: PayloadSet[];
  payloadPositions: Array<{
    start: number;
    end: number;
    value: string;
    paramName?: string;
  }>;
  attackType: string;
}>();

// 定义事件
const emit = defineEmits<{
  (e: 'update:payloadSets', value: PayloadSet[]): void;
  (e: 'add-payload-item', value: string, index: number): void;
  (e: 'paste-payload-items', items: string[], index: number): void;
  (e: 'load-payload-items', items: string[], index: number): void;
  (e: 'remove-payload-item', index: number, payloadIndex: number): void;
  (e: 'clear-payload-items', index: number): void;
}>();

// 处理规则状态
const processingRules = ref<Array<ProcessingRule>>([]);

// 修改当前标签页的 processing rules 和 encoding 设置
const currentPayloadSetProcessing = computed({
  get: () => {
    // 确保在 activePayloadSetIndex 无效或 payloadSets 为空时返回默认值
    if (!props.payloadSets || 
        activePayloadSetIndex.value === undefined || 
        activePayloadSetIndex.value < 0 || 
        !props.payloadSets[activePayloadSetIndex.value]) {
      return {
        rules: [],
        encoding: {
          enabled: false,
          urlEncode: false,
          characterSet: 'UTF-8'
        }
      };
    }
    
    // 确保 processing 对象存在
    const processing = props.payloadSets[activePayloadSetIndex.value].processing;
    if (!processing) {
      return {
        rules: [],
        encoding: {
          enabled: false,
          urlEncode: false,
          characterSet: 'UTF-8'
        }
      };
    }
    
    // 确保 rules 数组存在
    return {
      rules: Array.isArray(processing.rules) ? processing.rules : [],
      encoding: processing.encoding || {
        enabled: false,
        urlEncode: false,
        characterSet: 'UTF-8'
      }
    };
  },
  set: (value) => {
    // 确保 value 不为 null 或 undefined
    if (!value) return;
    
    // 创建新的 PayloadSets 数组，避免直接修改 props
    const newPayloadSets = Array.isArray(props.payloadSets) ? [...props.payloadSets] : [];
    
    // 确保数组长度足够
    while (newPayloadSets.length <= activePayloadSetIndex.value) {
      newPayloadSets.push({
        id: newPayloadSets.length + 1,
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
      });
    }
    
    // 安全地更新 processing 字段
    if (newPayloadSets[activePayloadSetIndex.value]) {
      newPayloadSets[activePayloadSetIndex.value].processing = {
        rules: Array.isArray(value.rules) ? value.rules : [],
        encoding: value.encoding || {
          enabled: false,
          urlEncode: false,
          characterSet: 'UTF-8'
        }
      };
    }
    
    // 更新 processingRules 引用，保持同步
    processingRules.value = Array.isArray(value.rules) ? [...value.rules] : [];
    
    // 发出更新事件
    emit('update:payloadSets', newPayloadSets);
  }
});

// 处理载荷项添加
const handleAddPayloadItem = (value: string) => {
  emit('add-payload-item', value, activePayloadSetIndex.value);
};

// 处理载荷项粘贴
const handlePastePayloadItems = (items: string[]) => {
  emit('paste-payload-items', items, activePayloadSetIndex.value);
};

// 处理载荷项加载
const handleLoadPayloadItems = (items: string[]) => {
  emit('load-payload-items', items, activePayloadSetIndex.value);
};

// 处理载荷项移除
const handleRemovePayloadItem = (index: number) => {
  emit('remove-payload-item', index, activePayloadSetIndex.value);
};

// 处理清空载荷项
const handleClearPayloadItems = () => {
  emit('clear-payload-items', activePayloadSetIndex.value);
};

// 处理编码设置更新
const handleEncodingUpdate = (newSettings: EncodingConfig) => {
  // 直接更新 payloadSets，避免通过 computed setter 可能导致的循环
  if (!props.payloadSets || activePayloadSetIndex.value < 0) return;

  // 创建新的 PayloadSets 数组
  const newPayloadSets = [...props.payloadSets];

  // 确保数组长度足够
  while (newPayloadSets.length <= activePayloadSetIndex.value) {
    newPayloadSets.push({
      id: newPayloadSets.length + 1,
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
    });
  }

  // 获取当前的 processing
  const currentProcessing = newPayloadSets[activePayloadSetIndex.value]?.processing || {
    rules: [],
    encoding: { enabled: false, urlEncode: false, characterSet: 'UTF-8' }
  };

  // 更新 encoding 设置
  newPayloadSets[activePayloadSetIndex.value] = {
    ...newPayloadSets[activePayloadSetIndex.value],
    processing: {
      rules: Array.isArray(currentProcessing.rules) ? currentProcessing.rules : [],
      encoding: newSettings || {
        enabled: false,
        urlEncode: false,
        characterSet: 'UTF-8'
      }
    }
  };

  // 发出更新事件
  emit('update:payloadSets', newPayloadSets);
};

// 处理规则相关
const handleAddRule = (ruleType: string, config: Record<string, any>) => {
  // 创建新规则
  const newRule = {
    id: Date.now(),
    type: ruleType,
    config: config || {}
  };
  
  console.log('添加新规则:', newRule);
  
  // 避免在 currentPayloadSetProcessing.value 可能为 null 时使用展开语法
  if (!currentPayloadSetProcessing.value) return;
  
  // 获取当前规则数组，确保它是数组
  const currentRules = Array.isArray(currentPayloadSetProcessing.value.rules) 
                        ? currentPayloadSetProcessing.value.rules 
                        : [];
  
  // 创建更新后的规则数组
  const updatedRules = [...currentRules, newRule];
  console.log('更新后的规则列表:', updatedRules);
  
  // 更新计算属性
  currentPayloadSetProcessing.value = {
    rules: updatedRules,
    encoding: currentPayloadSetProcessing.value.encoding || {
      enabled: false,
      urlEncode: false,
      characterSet: 'UTF-8'
    }
  };
};

const handleRemoveRule = (index: number) => {
  // 避免在 currentPayloadSetProcessing.value 可能为 null 时使用展开语法
  if (!currentPayloadSetProcessing.value) return;
  
  // 获取当前规则数组，确保它是数组
  const currentRules = Array.isArray(currentPayloadSetProcessing.value.rules) 
                        ? currentPayloadSetProcessing.value.rules 
                        : [];
  
  // 创建新数组并移除指定项
  const newRules = [...currentRules];
  if (index >= 0 && index < newRules.length) {
    newRules.splice(index, 1);
  }
  
  // 更新计算属性
  currentPayloadSetProcessing.value = {
    rules: newRules,
    encoding: currentPayloadSetProcessing.value.encoding || {
      enabled: false,
      urlEncode: false,
      characterSet: 'UTF-8'
    }
  };
};

const handleMoveRuleUp = (index: number) => {
  // 避免在 currentPayloadSetProcessing.value 可能为 null 时使用展开语法
  if (!currentPayloadSetProcessing.value) return;
  
  // 获取当前规则数组，确保它是数组
  const currentRules = Array.isArray(currentPayloadSetProcessing.value.rules) 
                        ? currentPayloadSetProcessing.value.rules 
                        : [];
  
  // 只有在索引有效且不是第一项时才进行上移操作
  if (index > 0 && index < currentRules.length) {
    // 创建新数组并交换位置
    const newRules = [...currentRules];
    const rule = newRules[index];
    newRules.splice(index, 1);
    newRules.splice(index - 1, 0, rule);
    
    // 更新计算属性
    currentPayloadSetProcessing.value = {
      rules: newRules,
      encoding: currentPayloadSetProcessing.value.encoding || {
        enabled: false,
        urlEncode: false,
        characterSet: 'UTF-8'
      }
    };
  }
};

const handleMoveRuleDown = (index: number) => {
  // 避免在 currentPayloadSetProcessing.value 可能为 null 时使用展开语法
  if (!currentPayloadSetProcessing.value) return;
  
  // 获取当前规则数组，确保它是数组
  const currentRules = Array.isArray(currentPayloadSetProcessing.value.rules) 
                        ? currentPayloadSetProcessing.value.rules 
                        : [];
  
  // 只有在索引有效且不是最后一项时才进行下移操作
  if (index >= 0 && index < currentRules.length - 1) {
    // 创建新数组并交换位置
    const newRules = [...currentRules];
    const rule = newRules[index];
    newRules.splice(index, 1);
    newRules.splice(index + 1, 0, rule);
    
    // 更新计算属性
    currentPayloadSetProcessing.value = {
      rules: newRules,
      encoding: currentPayloadSetProcessing.value.encoding || {
        enabled: false,
        urlEncode: false,
        characterSet: 'UTF-8'
      }
    };
  }
};

// 处理规则更新
const handleProcessingRulesUpdate = (rules: Array<ProcessingRule>) => {
  console.log('收到规则更新:', rules);
  
  // 确保 rules 是数组
  const safeRules = Array.isArray(rules) ? rules : [];
  
  // 更新 processingRules 引用
  processingRules.value = [...safeRules];
  
  // 避免在 currentPayloadSetProcessing.value 可能为 null 时使用展开语法
  if (!currentPayloadSetProcessing.value) return;
  
  // 更新计算属性
  currentPayloadSetProcessing.value = {
    rules: safeRules,
    encoding: currentPayloadSetProcessing.value.encoding || {
      enabled: false,
      urlEncode: false,
      characterSet: 'UTF-8'
    }
  };
  
  // 安全地记录日志
  console.log('PayloadSet 处理后:', 
    props.payloadSets && 
    activePayloadSetIndex.value >= 0 && 
    activePayloadSetIndex.value < props.payloadSets.length ?
    props.payloadSets[activePayloadSetIndex.value]?.processing :
    'No valid payload set');
};

// 更新payload sets，确保类型兼容性
const handleUpdatePayloadSets = (updatedSets: PayloadSet[]) => {
  // 确保处理规则存在
  const convertedSets = updatedSets.map(set => {
    const newSet = { ...set };
    
    // 确保处理规则存在
    if (!newSet.processing) {
      newSet.processing = {
        rules: [],
        encoding: {
          enabled: false,
          urlEncode: false,
          characterSet: 'UTF-8'
        }
      };
    }
    
    return newSet;
  });

  // 确保当前索引的 payload set 存在
  while (convertedSets.length <= activePayloadSetIndex.value) {
    convertedSets.push({
      id: convertedSets.length + 1, // 使用连续的数字id，从1开始
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
    });
  }
  
  emit('update:payloadSets', convertedSets);
};

// 处理单个 payload set 更新
const handlePayloadSetUpdate = (updatedPayloadSet: PayloadSet) => {
  if (!props.payloadSets) return;
  
  const newPayloadSets = [...props.payloadSets];
  // 确保数组长度足够
  while (newPayloadSets.length <= activePayloadSetIndex.value) {
    newPayloadSets.push({
      id: newPayloadSets.length + 1, // 使用连续的数字id，从1开始
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
    });
  }
  // 更新指定索引的 payload set
  newPayloadSets[activePayloadSetIndex.value] = updatedPayloadSet;
  // 转换并发出更新事件
  handleUpdatePayloadSets(newPayloadSets);
};

// 当前激活的 payload set 标签页索引
const activePayloadSetIndex = ref(0);

// 计算显示标签页数量
const tabCount = computed(() => {
  if (props.attackType === 'sniper' || props.attackType === 'battering-ram') {
    return 1;
  }
  return props.payloadPositions.length;
});

// 标签页标题
const getTabTitle = (index: number) => {
  const position = props.payloadPositions[index];
  if (position?.paramName) {
    return `${t('modules.intruder.payload')} ${index + 1} - ${position.paramName}`;
  }
  return `${t('modules.intruder.payload')} ${index + 1}`;
};

// 处理标签页切换
const handleTabChange = (index: number) => {
  activePayloadSetIndex.value = index;
};

// 监听标签页切换，确保有对应的 payload set
watch([activePayloadSetIndex, () => props.payloadPositions.length], () => {
  if (activePayloadSetIndex.value >= 0 && activePayloadSetIndex.value < tabCount.value) {
    // 检查props.payloadSets是否为数组
    if (!Array.isArray(props.payloadSets)) {
      return;
    }
    
    // 确保当前索引的 payload set 存在
    if (!props.payloadSets[activePayloadSetIndex.value]) {
      const newPayloadSet = {
        id: props.payloadSets.length + 1, // 使用连续的数字id，从1开始
        type: 'simple-list' as const,
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
      
      // 创建一个新的数组，包含所有现有的 payload sets
      const newPayloadSets = Array.isArray(props.payloadSets) ? [...props.payloadSets] : [];
      
      // 确保数组长度足够
      while (newPayloadSets.length <= activePayloadSetIndex.value) {
        newPayloadSets.push({
          id: newPayloadSets.length + 1, // 使用连续的数字id，从1开始
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
        });
      }
      newPayloadSets[activePayloadSetIndex.value] = newPayloadSet;
      handleUpdatePayloadSets(newPayloadSets);
    }
  }
}, { immediate: true });

// 确保当 payloadPositions 改变时，activePayloadSetIndex 不会超出范围
watch(() => props.payloadPositions.length, (newLength) => {
  if (activePayloadSetIndex.value >= newLength) {
    activePayloadSetIndex.value = Math.max(0, newLength - 1);
  }
});

</script>

<template>
  <div>
    <!-- 在原有的payload配置区域之前添加tabs -->
    <div v-if="tabCount > 1" class="mb-4">
      <div class="flex border-b border-gray-200 dark:border-gray-700">
        <button v-for="index in tabCount" 
                :key="index"
                class="py-2 px-4 text-sm"
                :class="[
                  activePayloadSetIndex === index - 1
                    ? 'text-[#4f46e5] border-b-2 border-[#4f46e5] font-medium'
                    : 'text-gray-500 hover:text-gray-700 dark:text-gray-400 dark:hover:text-gray-300'
                ]"
                @click="handleTabChange(index - 1)">
          {{ getTabTitle(index - 1) }}
        </button>
      </div>
    </div>

    <!-- 可添加新的载荷集配置区域 -->
    <div class="p-3 border-t border-gray-200 dark:border-gray-700">
      <h3 class="text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">
        {{ t('modules.intruder.payload_sets') }}
      </h3>
      
      <!-- 确保传递有效的 payload set -->
      <PayloadSetConfig
        :payloadSet="payloadSets[activePayloadSetIndex] || {
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
        }"
        :payloadPosition="{
          ...payloadPositions[activePayloadSetIndex],
          index: activePayloadSetIndex
        }"
        @update:payload-set="handlePayloadSetUpdate"
        @add-payload-item="handleAddPayloadItem"
        @paste-payload-items="handlePastePayloadItems"
        @load-payload-items="handleLoadPayloadItems"
        @remove-payload-item="handleRemovePayloadItem"
        @clear-payload-items="handleClearPayloadItems"
      />
      
      <!-- 处理规则 - 通过 v-if 确保 rules 存在 -->
      <div v-if="payloadSets[activePayloadSetIndex]">
        <!-- 添加调试信息 -->
        <div class="text-xs text-gray-500 dark:text-gray-400 mb-1">
          当前规则数: {{ (currentPayloadSetProcessing.rules || []).length }}
        </div>
        
        <PayloadProcessingRules
          :rules="(currentPayloadSetProcessing.rules || []).map((rule, index) => ({
            id: (rule as any).id || Date.now() + index,
            type: rule.type,
            config: rule.config
          }))"
          @add-rule="handleAddRule"
          @remove-rule="handleRemoveRule"
          @move-rule-up="handleMoveRuleUp"
          @move-rule-down="handleMoveRuleDown"
          @update:rules="handleProcessingRulesUpdate"
        />
      </div>
      
      <!-- 编码设置 - 使用新的 encoding 对象 -->
      <PayloadEncodingConfig
        :encoding="currentPayloadSetProcessing.encoding"
        @update:encoding="handleEncodingUpdate"
      />
    </div>
  </div>
</template>