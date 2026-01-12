<script setup lang="ts">
import { ref, computed } from 'vue';
import { useI18n } from 'vue-i18n';

const { t } = useI18n();

// Props定义
const props = defineProps<{
  rules: Array<{
    id: number;
    type: string;
    config: Record<string, any>;
  }>;
}>();

// Emits定义
const emit = defineEmits<{
  (e: 'update:rules', value: any[]): void;
  (e: 'add-rule', ruleType: string, config: Record<string, any>): void;
  (e: 'remove-rule', index: number): void;
  (e: 'move-rule-up', index: number): void;
  (e: 'move-rule-down', index: number): void;
}>();

// 处理规则弹窗状态
const showRuleModal = ref(false);
const selectedRuleType = ref('');
const selectedEncodeType = ref('');
const selectedDecodeType = ref('');
const ruleConfig = ref<Record<string, any>>({});
const selectedRuleIndex = ref(-1);

// 面板折叠状态
const processingPanelOpen = ref(true);

// 处理规则类型列表 - 使用 computed 确保 i18n 已初始化
const processingRuleTypes = computed(() => [
  { value: '', label: t('common.ui.none') },
  { value: 'add-prefix', label: t('modules.intruder.prefix') },
  { value: 'add-suffix', label: t('modules.intruder.suffix') },
  { value: 'match-replace', label: t('modules.intruder.match') + '/' + t('modules.intruder.replace') },
  { value: 'substring', label: 'Substring' },
  { value: 'reverse-substring', label: 'Reverse substring' },
  { value: 'modify-case', label: 'Modify case' },
  { value: 'encode', label: t('modules.decoder.encode') },
  { value: 'decode', label: t('modules.decoder.decode') },
  { value: 'hash', label: 'Hash' }
]);

// 编码类型选项 - 使用 computed 确保 i18n 已初始化
const encodeTypes = computed(() => [
  { value: '', label: t('common.ui.none') },
  { value: 'url-key', label: 'URL-encode key characters' },
  { value: 'url-all', label: 'URL-encode all characters' },
  { value: 'url-unicode', label: 'URL-encode all characters (Unicode)' },
  { value: 'html-key', label: 'HTML-encode key characters' },
  { value: 'html-all', label: 'HTML-encode all characters' },
  { value: 'html-numeric', label: 'HTML-encode all characters (numeric entities)' },
  { value: 'html-hex', label: 'HTML-encode all characters (hex entities)' },
  { value: 'base64', label: t('modules.decoder.methods.base64') },
  { value: 'ascii-hex', label: t('modules.decoder.methods.hex') },
  { value: 'unicode-escape', label: 'Unicode escape sequences' }
]);

// 解码类型选项 - 使用 computed 确保 i18n 已初始化
const decodeTypes = computed(() => [
  { value: '', label: t('common.ui.none') },
  { value: 'url', label: t('modules.decoder.methods.url') },
  { value: 'html', label: t('modules.decoder.methods.html') },
  { value: 'base64', label: t('modules.decoder.methods.base64') },
  { value: 'ascii-hex', label: t('modules.decoder.methods.hex') },
  { value: 'unicode-unescape', label: 'Unicode unescape' }
]);

// 切换面板折叠状态
const togglePanel = () => {
  processingPanelOpen.value = !processingPanelOpen.value;
};

// 打开新规则弹窗
const openAddRuleModal = () => {
  selectedRuleType.value = '';
  selectedEncodeType.value = '';
  selectedDecodeType.value = '';
  ruleConfig.value = {};
  selectedRuleIndex.value = -1;
  showRuleModal.value = true;
};

// 选择规则进行编辑
const selectRule = (index: number) => {
  selectedRuleIndex.value = index;
};

// 关闭规则弹窗
const closeRuleModal = () => {
  showRuleModal.value = false;
};

// 确认添加规则
const confirmAddRule = () => {
  if (!selectedRuleType.value) {
    // 显示错误信息：必须选择规则类型
    return;
  }
  
  let config: Record<string, any> = {};
  
  // 根据不同的规则类型设置配置
  switch (selectedRuleType.value) {
    case 'add-prefix':
    case 'add-suffix':
      config = { text: ruleConfig.value.text || '' };
      break;
    case 'match-replace':
      config = { 
        match: ruleConfig.value.match || '',
        replace: ruleConfig.value.replace || '',
        regex: ruleConfig.value.regex || false
      };
      break;
    case 'substring':
    case 'reverse-substring':
      config = {
        startIndex: ruleConfig.value.startIndex || 0,
        length: ruleConfig.value.length || 1
      };
      break;
    case 'modify-case':
      config = { type: ruleConfig.value.type || 'lowercase' };
      break;
    case 'encode':
      config = { type: selectedEncodeType.value };
      break;
    case 'decode':
      config = { type: selectedDecodeType.value };
      break;
    case 'hash':
      config = { type: ruleConfig.value.type || 'md5' };
      break;
  }
  
  emit('add-rule', selectedRuleType.value, config);
  closeRuleModal();
};

// 删除规则
const removeRule = (index: number) => {
  emit('remove-rule', index);
  if (selectedRuleIndex.value === index) {
    selectedRuleIndex.value = -1;
  }
};

// 上移规则
const moveRuleUp = (index: number) => {
  if (index > 0) {
    emit('move-rule-up', index);
    if (selectedRuleIndex.value === index) {
      selectedRuleIndex.value = index - 1;
    } else if (selectedRuleIndex.value === index - 1) {
      selectedRuleIndex.value = index;
    }
  }
};

// 下移规则
const moveRuleDown = (index: number) => {
  if (index < props.rules.length - 1) {
    emit('move-rule-down', index);
    if (selectedRuleIndex.value === index) {
      selectedRuleIndex.value = index + 1;
    } else if (selectedRuleIndex.value === index + 1) {
      selectedRuleIndex.value = index;
    }
  }
};

// 规则描述
const getRuleDescription = (rule: { type: string, config: Record<string, any> }): string => {
  switch (rule.type) {
    case 'add-prefix':
      return `${t('modules.intruder.prefix')}: "${rule.config.text}"`;
    case 'add-suffix':
      return `${t('modules.intruder.suffix')}: "${rule.config.text}"`;
    case 'match-replace':
      return `${t('modules.intruder.replace')} ${rule.config.regex ? 'regex' : 'string'} "${rule.config.match}" ${t('common.with')} "${rule.config.replace}"`;
    case 'substring':
      return `子字符串: 从索引 ${rule.config.startIndex} 取 ${rule.config.length} 个字符`;
    case 'reverse-substring':
      return `反向子字符串: 从索引 ${rule.config.startIndex} 取 ${rule.config.length} 个字符并反转`;
    case 'modify-case':
      const caseTypeMap: Record<string, string> = {
        'lowercase': '转小写',
        'uppercase': '转大写',
        'capitalize': '首字母大写'
      };
      return `修改大小写: ${caseTypeMap[rule.config.type] || rule.config.type}`;
    case 'encode':
      const encodeType = encodeTypes.value.find(item => item.value === rule.config.type);
      return `${t('modules.decoder.encode')}: ${encodeType?.label || rule.config.type}`;
    case 'decode':
      const decodeType = decodeTypes.value.find(item => item.value === rule.config.type);
      return `${t('modules.decoder.decode')}: ${decodeType?.label || rule.config.type}`;
    case 'hash':
      return `哈希: ${rule.config.type?.toUpperCase() || 'MD5'}`;
    default:
      return `${t('modules.intruder.rule')}: ${rule.type}`;
  }
};

// 检查是否可以添加规则
const canAddRule = computed(() => {
  if (!selectedRuleType.value) return false;
  
  // 对于encode和decode规则类型，需要选择具体的编码/解码方式
  if (selectedRuleType.value === 'encode') {
    return !!selectedEncodeType.value;
  } else if (selectedRuleType.value === 'decode') {
    return !!selectedDecodeType.value;
  } else if (selectedRuleType.value === 'modify-case') {
    return !!ruleConfig.value.type;
  } else if (selectedRuleType.value === 'hash') {
    return !!ruleConfig.value.type;
  } else if (selectedRuleType.value === 'substring' || selectedRuleType.value === 'reverse-substring') {
    return (ruleConfig.value.startIndex >= 0) && (ruleConfig.value.length > 0);
  } else if (selectedRuleType.value === 'add-prefix' || selectedRuleType.value === 'add-suffix') {
    return !!ruleConfig.value.text;
  } else if (selectedRuleType.value === 'match-replace') {
    return !!ruleConfig.value.match && !!ruleConfig.value.replace;
  }
  
  return true;
});
</script>

<template>
  <div class="border border-gray-200 dark:border-gray-700 rounded-md mb-2">
    <!-- 处理规则面板标题 -->
    <div 
      class="flex items-center justify-between p-2 bg-[#f3f4f6] dark:bg-[#282838] cursor-pointer" 
      @click="togglePanel"
    >
      <h4 class="text-sm font-medium text-gray-700 dark:text-gray-300">
        {{ t('modules.intruder.processing_panel') }}
      </h4>
      <i :class="[processingPanelOpen ? 'bx bx-chevron-up' : 'bx bx-chevron-down']" class="text-gray-500"></i>
    </div>
    
    <!-- 处理规则内容 -->
    <div class="p-2" v-show="processingPanelOpen">
      <p class="mb-2 text-xs text-gray-600 dark:text-gray-400">
        {{ t('modules.intruder.payload_processing_description') }}
      </p>
      
      <div class="flex gap-2 mb-2">
        <div class="flex flex-col gap-1">
          <button 
            type="button" 
            class="px-2 py-1 text-xs border border-gray-300 dark:border-gray-600 rounded bg-white dark:bg-[#282838] text-gray-700 dark:text-gray-300 hover:bg-gray-100 dark:hover:bg-[#32324c]"
            @click="openAddRuleModal"
          >
            {{ t('modules.intruder.add') }}
          </button>
          <button 
            type="button" 
            class="px-2 py-1 text-xs border border-gray-300 dark:border-gray-600 rounded bg-white dark:bg-[#282838] text-gray-700 dark:text-gray-300 hover:bg-gray-100 dark:hover:bg-[#32324c]"
            :disabled="selectedRuleIndex < 0"
            :class="selectedRuleIndex < 0 ? 'opacity-50 cursor-not-allowed' : ''"
          >
            {{ t('common.actions.edit') }}
          </button>
          <button 
            type="button" 
            class="px-2 py-1 text-xs border border-gray-300 dark:border-gray-600 rounded bg-white dark:bg-[#282838] text-gray-700 dark:text-gray-300 hover:bg-gray-100 dark:hover:bg-[#32324c]"
            @click="selectedRuleIndex >= 0 && removeRule(selectedRuleIndex)"
            :disabled="selectedRuleIndex < 0"
            :class="selectedRuleIndex < 0 ? 'opacity-50 cursor-not-allowed' : ''"
          >
            {{ t('modules.intruder.remove') }}
          </button>
          <button 
            type="button" 
            class="px-2 py-1 text-xs border border-gray-300 dark:border-gray-600 rounded bg-white dark:bg-[#282838] text-gray-700 dark:text-gray-300 hover:bg-gray-100 dark:hover:bg-[#32324c]"
            @click="selectedRuleIndex >= 0 && moveRuleUp(selectedRuleIndex)"
            :disabled="selectedRuleIndex <= 0"
            :class="selectedRuleIndex <= 0 ? 'opacity-50 cursor-not-allowed' : ''"
          >
            {{ t('modules.intruder.up') }}
          </button>
          <button 
            type="button" 
            class="px-2 py-1 text-xs border border-gray-300 dark:border-gray-600 rounded bg-white dark:bg-[#282838] text-gray-700 dark:text-gray-300 hover:bg-gray-100 dark:hover:bg-[#32324c]"
            @click="selectedRuleIndex >= 0 && moveRuleDown(selectedRuleIndex)"
            :disabled="selectedRuleIndex < 0 || selectedRuleIndex >= props.rules.length - 1"
            :class="selectedRuleIndex < 0 || selectedRuleIndex >= props.rules.length - 1 ? 'opacity-50 cursor-not-allowed' : ''"
          >
            {{ t('modules.intruder.down') }}
          </button>
        </div>
        
        <div class="flex-1 border border-gray-300 dark:border-gray-600 rounded bg-white dark:bg-[#1e1e2e]">
          <table class="w-full text-xs">
            <thead>
              <tr class="border-b border-gray-200 dark:border-gray-700">
                <th class="px-2 py-1 text-left">{{ t('modules.intruder.enabled') }}</th>
                <th class="px-2 py-1 text-left">{{ t('modules.intruder.rule') }}</th>
              </tr>
            </thead>
            <tbody>
              <tr v-if="props.rules.length === 0" class="text-gray-500 dark:text-gray-400 italic text-center">
                <td colspan="2" class="px-2 py-4">{{ t('modules.intruder.no_rules') }}</td>
              </tr>
              <tr 
                v-for="(rule, index) in props.rules" 
                :key="rule.id" 
                :class="[
                  'border-b border-gray-200 dark:border-gray-700 last:border-b-0',
                  selectedRuleIndex === index ? 'bg-blue-50 dark:bg-blue-900/20' : ''
                ]"
                @click="selectRule(index)"
                class="cursor-pointer hover:bg-gray-100 dark:hover:bg-[#32324c]"
              >
                <td class="px-2 py-1.5">
                  <input type="checkbox" checked class="rounded border-gray-300">
                </td>
                <td class="px-2 py-1.5">{{ getRuleDescription(rule) }}</td>
              </tr>
            </tbody>
          </table>
        </div>
      </div>
    </div>
    
    <!-- 添加规则弹窗 -->
    <div v-if="showRuleModal" class="fixed inset-0 bg-black/50 flex items-center justify-center z-50">
      <div class="bg-white dark:bg-[#1e1e2e] w-full max-w-md mx-auto rounded-lg shadow-xl overflow-hidden">
        <!-- 弹窗标题 -->
        <div class="bg-[#f3f4f6] dark:bg-[#282838] px-4 py-3 border-b border-gray-200 dark:border-gray-700 flex items-center justify-between">
          <h3 class="text-sm font-medium text-gray-700 dark:text-gray-300">
            {{ t('modules.intruder.add_processing_rule') }}
          </h3>
          <button 
            class="text-gray-400 hover:text-gray-600 dark:hover:text-gray-200" 
            @click="closeRuleModal"
          >
            <i class="bx bx-x text-xl"></i>
          </button>
        </div>
        
        <!-- 弹窗内容 -->
        <div class="p-4">
          <!-- 规则说明 -->
          <div class="flex mb-4">
            <div class="mr-3 text-xl text-gray-500 dark:text-gray-400">
              <i class="bx bx-question-mark rounded-full border border-gray-500 dark:border-gray-400 w-7 h-7 flex items-center justify-center"></i>
            </div>
            <p class="text-sm text-gray-600 dark:text-gray-400">
              {{ t('modules.intruder.processing_rule_description') }}
            </p>
          </div>
          
          <!-- 规则类型选择器 -->
          <div class="mb-4">
            <div class="relative">
              <select
                v-model="selectedRuleType"
                class="w-full px-3 py-2 text-sm border border-gray-300 dark:border-gray-700 rounded-md bg-white dark:bg-[#282838] text-gray-700 dark:text-gray-300 appearance-none"
              >
                <option v-for="rule in processingRuleTypes" :key="rule.value" :value="rule.value">
                  {{ rule.label }}
                </option>
              </select>
              <div class="absolute inset-y-0 right-0 flex items-center px-2 pointer-events-none">
                <i class="bx bx-chevron-down text-gray-500"></i>
              </div>
            </div>
          </div>
          
          <!-- 编码类型选择器（仅当选择Encode时显示） -->
          <div v-if="selectedRuleType === 'encode'" class="mb-4">
            <div class="relative">
              <select
                v-model="selectedEncodeType"
                class="w-full px-3 py-2 text-sm border border-gray-300 dark:border-gray-700 rounded-md bg-white dark:bg-[#282838] text-gray-700 dark:text-gray-300 appearance-none"
              >
                <option v-for="type in encodeTypes" :key="type.value" :value="type.value">
                  {{ type.label }}
                </option>
              </select>
              <div class="absolute inset-y-0 right-0 flex items-center px-2 pointer-events-none">
                <i class="bx bx-chevron-down text-gray-500"></i>
              </div>
            </div>
          </div>
          
          <!-- 解码类型选择器（仅当选择Decode时显示） -->
          <div v-if="selectedRuleType === 'decode'" class="mb-4">
            <div class="relative">
              <select
                v-model="selectedDecodeType"
                class="w-full px-3 py-2 text-sm border border-gray-300 dark:border-gray-700 rounded-md bg-white dark:bg-[#282838] text-gray-700 dark:text-gray-300 appearance-none"
              >
                <option v-for="type in decodeTypes" :key="type.value" :value="type.value">
                  {{ type.label }}
                </option>
              </select>
              <div class="absolute inset-y-0 right-0 flex items-center px-2 pointer-events-none">
                <i class="bx bx-chevron-down text-gray-500"></i>
              </div>
            </div>
          </div>
          
          <!-- 前缀/后缀配置 -->
          <div v-if="selectedRuleType === 'add-prefix' || selectedRuleType === 'add-suffix'" class="mb-4">
            <label class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">
              {{ selectedRuleType === 'add-prefix' ? t('modules.intruder.prefix') : t('modules.intruder.suffix') }}
            </label>
            <input 
              spellcheck="false"
              v-model="ruleConfig.text"
              type="text"
              class="w-full px-3 py-2 text-sm border border-gray-300 dark:border-gray-700 rounded-md bg-white dark:bg-[#282838] text-gray-700 dark:text-gray-300"
              :placeholder="selectedRuleType === 'add-prefix' ? t('modules.intruder.prefix') : t('modules.intruder.suffix')"
            >
          </div>
          
          <!-- 匹配替换配置 -->
          <div v-if="selectedRuleType === 'match-replace'" class="mb-4 space-y-2">
            <div>
              <label class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">
                {{ t('modules.intruder.match') }}
              </label>
              <input 
                spellcheck="false"
                v-model="ruleConfig.match"
                type="text"
                class="w-full px-3 py-2 text-sm border border-gray-300 dark:border-gray-700 rounded-md bg-white dark:bg-[#282838] text-gray-700 dark:text-gray-300"
                :placeholder="t('modules.intruder.match')"
              >
            </div>
            <div>
              <label class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">
                {{ t('modules.intruder.replace') }}
              </label>
              <input 
                spellcheck="false"
                v-model="ruleConfig.replace"
                type="text"
                class="w-full px-3 py-2 text-sm border border-gray-300 dark:border-gray-700 rounded-md bg-white dark:bg-[#282838] text-gray-700 dark:text-gray-300"
                :placeholder="t('modules.intruder.replace')"
              >
            </div>
            <div class="flex items-center">
              <input 
                spellcheck="false"
                v-model="ruleConfig.regex"
                type="checkbox"
                id="regex-checkbox"
                class="mr-2 rounded border-gray-300"
              >
              <label for="regex-checkbox" class="text-sm text-gray-700 dark:text-gray-300">
                {{ t('modules.intruder.use_regex') }}
              </label>
            </div>
          </div>

          <!-- 子字符串配置 -->
          <div v-if="selectedRuleType === 'substring' || selectedRuleType === 'reverse-substring'" class="mb-4 space-y-2">
            <div>
              <label class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">
                开始索引
              </label>
              <input 
                spellcheck="false"
                v-model.number="ruleConfig.startIndex"
                type="number"
                min="0"
                class="w-full px-3 py-2 text-sm border border-gray-300 dark:border-gray-700 rounded-md bg-white dark:bg-[#282838] text-gray-700 dark:text-gray-300"
                placeholder="0"
              >
            </div>
            <div>
              <label class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">
                长度
              </label>
              <input 
                spellcheck="false"  
                v-model.number="ruleConfig.length"
                type="number"
                min="1"
                class="w-full px-3 py-2 text-sm border border-gray-300 dark:border-gray-700 rounded-md bg-white dark:bg-[#282838] text-gray-700 dark:text-gray-300"
                placeholder="1"
              >
            </div>
          </div>

          <!-- 大小写修改配置 -->
          <div v-if="selectedRuleType === 'modify-case'" class="mb-4">
            <label class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">
              大小写类型
            </label>
            <div class="relative">
              <select
                v-model="ruleConfig.type"
                class="w-full px-3 py-2 text-sm border border-gray-300 dark:border-gray-700 rounded-md bg-white dark:bg-[#282838] text-gray-700 dark:text-gray-300 appearance-none"
              >
                <option value="">请选择...</option>
                <option value="lowercase">转小写</option>
                <option value="uppercase">转大写</option>
                <option value="capitalize">首字母大写</option>
              </select>
              <div class="absolute inset-y-0 right-0 flex items-center px-2 pointer-events-none">
                <i class="bx bx-chevron-down text-gray-500"></i>
              </div>
            </div>
          </div>

          <!-- 哈希配置 -->
          <div v-if="selectedRuleType === 'hash'" class="mb-4">
            <label class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">
              哈希类型
            </label>
            <div class="relative">
              <select
                v-model="ruleConfig.type"
                class="w-full px-3 py-2 text-sm border border-gray-300 dark:border-gray-700 rounded-md bg-white dark:bg-[#282838] text-gray-700 dark:text-gray-300 appearance-none"
              >
                <option value="">请选择...</option>
                <option value="md5">MD5</option>
                <option value="sha1">SHA1</option>
                <option value="sha256">SHA256</option>
              </select>
              <div class="absolute inset-y-0 right-0 flex items-center px-2 pointer-events-none">
                <i class="bx bx-chevron-down text-gray-500"></i>
              </div>
            </div>
          </div>
        </div>
        
        <!-- 弹窗按钮 -->
        <div class="bg-[#f3f4f6] dark:bg-[#282838] px-4 py-3 flex justify-end border-t border-gray-200 dark:border-gray-700">
          <button 
            class="mr-2 px-3 py-1.5 text-sm bg-gray-200 dark:bg-gray-700 hover:bg-gray-300 dark:hover:bg-gray-600 text-gray-700 dark:text-gray-300 rounded"
            @click="closeRuleModal"
          >
            {{ t('common.cancel') }}
          </button>
          <button 
            class="px-3 py-1.5 text-sm bg-[#4f46e5] hover:bg-[#4338ca] text-white rounded disabled:opacity-50"
            @click="confirmAddRule"
            :disabled="!canAddRule"
          >
            {{ t('common.add') }}
          </button>
        </div>
      </div>
    </div>
  </div>
</template> 