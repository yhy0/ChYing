<script setup lang="ts">
import { ref, onMounted, computed, watch, nextTick } from 'vue';
import { useI18n } from 'vue-i18n';
// 使用any类型处理找不到的模块
// @ts-ignore
import { QueryHistoryByDSL } from '../../../../bindings/github.com/yhy0/ChYing/app.js';

// 使用国际化
const { t } = useI18n();

// 定义组件事件
const emit = defineEmits(['search-results', 'clear-search', 'notify']);

// DSL 查询表达式
const dslQuery = ref('');
// 是否显示输入提示
const showSuggestions = ref(false);
// 提示列表
const suggestions = ref<Array<{value: string, description: string, type: string}>>([]);
// 历史搜索记录
const searchHistory = ref<string[]>([]);
// 最大历史记录数量
const MAX_HISTORY = 10;

// 内置字段提示
const builtinFields = [
  { value: 'method', description: t('modules.proxy.dsl.field_method') },
  { value: 'status', description: t('modules.proxy.dsl.field_status') },
  { value: 'path', description: t('modules.proxy.dsl.field_path') },
  { value: 'host', description: t('modules.proxy.dsl.field_host') },
  { value: 'url', description: t('modules.proxy.dsl.field_url') },
  { value: 'content_type', description: t('modules.proxy.dsl.field_content_type') },
  { value: 'request_body', description: t('modules.proxy.dsl.field_request_body') },
  { value: 'response_body', description: t('modules.proxy.dsl.field_response_body') },
  { value: 'id', description: t('modules.proxy.dsl.field_id') },
  { value: 'length', description: t('modules.proxy.dsl.field_length') },
  { value: 'response', description: t('modules.proxy.dsl.field_response') },
  { value: 'request', description: t('modules.proxy.dsl.field_request') },
];

// 内置函数提示
const builtinFunctions = [
  { value: 'contains(', description: t('modules.proxy.dsl.function_contains') },
  { value: 'regex(', description: t('modules.proxy.dsl.function_regex') },
  { value: '&&', description: t('modules.proxy.dsl.function_and') },
  { value: '||', description: t('modules.proxy.dsl.function_or') },
  { value: '==', description: t('modules.proxy.dsl.function_equals') },
  { value: '!=', description: t('modules.proxy.dsl.function_not_equals') },
];

// 示例查询表达式
const dslExamples = [
  { name: t('modules.proxy.dsl.example_contains'), query: 'contains(request, "admin")' },
  { name: t('modules.proxy.dsl.example_status'), query: 'status == "200"' },
  { name: t('modules.proxy.dsl.example_content_type'), query: 'contains(content_type, "application/json")' },
  { name: t('modules.proxy.dsl.example_path'), query: 'contains(path, "/api/")' },
  { name: t('modules.proxy.dsl.example_method'), query: 'method == "POST"' },
  { name: t('modules.proxy.dsl.example_response'), query: 'contains(response_body, "error")' },
  { name: t('modules.proxy.dsl.example_combined'), query: 'method == "GET" && status == "200"' }
];

// 默认建议列表
const defaultSuggestions = [
  { value: 'status == "200"', description: '查找所有成功请求', type: 'example' },
  { value: 'contains(path, "/api/")', description: '查找API路径', type: 'example' },
  { value: 'method == "POST"', description: '查找POST请求', type: 'example' },
  { value: 'status', description: t('modules.proxy.dsl.field_status'), type: 'field' },
  { value: 'method', description: t('modules.proxy.dsl.field_method'), type: 'field' },
  { value: 'contains(', description: t('modules.proxy.dsl.function_contains'), type: 'function' }
];

// 获取当前正在输入的单词
function getCurrentWord(text: string): string {
  const cursorPosition = text.length;
  let startPos = cursorPosition;
  
  // 向前查找单词的开始位置
  while (startPos > 0) {
    const char = text.charAt(startPos - 1);
    // 如果遇到空格、括号、引号、运算符等分隔符则停止
    if (/[\s()'"=!&|,.<>?:;[\]{}+\-*/]/.test(char)) {
      break;
    }
    startPos--;
  }
  
  return text.substring(startPos, cursorPosition);
}

// 更新建议列表
const updateSuggestions = () => {
  // 先确保我们有数据可以显示
  showSuggestions.value = true;
  
  // 生成所有可能的建议
  let allSuggestions = [
    // 所有字段
    ...builtinFields.map(field => ({ 
      value: field.value, 
      description: field.description,
      type: 'field'
    })),
    // 所有函数
    ...builtinFunctions.map(func => ({ 
      value: func.value, 
      description: func.description,
      type: 'function'
    })),
    // 历史记录
    ...searchHistory.value.slice(0, 3).map(hist => ({ 
      value: hist, 
      description: t('modules.proxy.dsl.previous_search'),
      type: 'history'
    }))
  ];
  
  // 无论是否有输入，都始终显示一些默认建议
  let filteredSuggestions = allSuggestions;
  
  // 如果有输入，则过滤建议
  if (dslQuery.value.trim()) {
    const currentWord = getCurrentWord(dslQuery.value);
    const lowercaseWord = currentWord.toLowerCase();
    
    // 过滤匹配的建议
    filteredSuggestions = allSuggestions.filter(suggestion => 
      currentWord === '' || suggestion.value.toLowerCase().includes(lowercaseWord)
    );
  }
  
  // 如果过滤后没有结果，使用默认建议
  if (filteredSuggestions.length === 0) {
    filteredSuggestions = defaultSuggestions;
  }
  
  // 限制显示数量，确保不会太多
  suggestions.value = filteredSuggestions.slice(0, 8);
  
};

// 应用选中的建议
const applySuggestion = (suggestion: { value: string, type: string }) => {
  const currentWord = getCurrentWord(dslQuery.value);
  const currentPos = dslQuery.value.length;
  const startPos = currentPos - currentWord.length;
  
  if (suggestion.type === 'history') {
    // 如果是历史记录，直接替换整个查询
    dslQuery.value = suggestion.value;
  } else {
    // 替换当前单词
    dslQuery.value = dslQuery.value.substring(0, startPos) + suggestion.value + 
      (suggestion.type === 'function' ? '' : ' ');
  }
  
  showSuggestions.value = false;
  // 聚焦回输入框
  setTimeout(() => {
    const inputEl = document.querySelector('.dsl-search-input') as HTMLInputElement;
    if (inputEl) {
      inputEl.focus();
      // 如果是函数，将光标放在括号内
      if (suggestion.type === 'function') {
        inputEl.selectionStart = inputEl.selectionEnd = startPos + suggestion.value.length;
      }
    }
  }, 0);
};

// 是否显示帮助信息
const showHelp = ref(false);
// 是否正在加载
const isLoading = ref(false);

// 执行DSL查询
const executeDSLQuery = async () => {
  if (dslQuery.value.trim() === '') {
    // 如果DSL为空，清除过滤器
    emit('clear-search');
    return;
  }
  
  try {
    isLoading.value = true;
     
    // 检查引号是否平衡
    const queryString = dslQuery.value.trim();
    const doubleQuotes = (queryString.match(/"/g) || []).length;
    const singleQuotes = (queryString.match(/'/g) || []).length;
     
    if (doubleQuotes % 2 !== 0 || singleQuotes % 2 !== 0) {
      throw new Error(t('modules.proxy.dsl.unbalanced_quotes'));
    }
    
    // 将查询添加到历史记录
    addToSearchHistory(queryString);
     
    // 调用后端API执行DSL查询
    const response = await QueryHistoryByDSL(queryString);
    
    // 检查是否有错误
    if (response.error) {
      throw new Error(response.error);
    }
    
    const results = response.data || [];
     
    // 将查询结果发送给父组件
    emit('search-results', results);
     
    // 显示通知
    emit('notify', {
      message: t('modules.proxy.dsl.search_complete', { count: results.length }),
      type: 'success'
    });
  } catch (error: any) { // 使用any类型处理未知错误类型
    console.error('DSL查询错误:', error);
    // 显示错误通知
    emit('notify', {
      message: t('modules.proxy.dsl.search_error', { error: error.toString() }),
      type: 'error'
    });
  } finally {
    isLoading.value = false;
    showSuggestions.value = false;
  }
};

// 添加到搜索历史
const addToSearchHistory = (query: string) => {
  // 如果已经存在，先移除旧记录
  const index = searchHistory.value.indexOf(query);
  if (index > -1) {
    searchHistory.value.splice(index, 1);
  }
  
  // 添加到历史记录开头
  searchHistory.value.unshift(query);
  
  // 限制历史记录数量
  if (searchHistory.value.length > MAX_HISTORY) {
    searchHistory.value = searchHistory.value.slice(0, MAX_HISTORY);
  }
  
  // 保存到 localStorage
  localStorage.setItem('dsl_search_history', JSON.stringify(searchHistory.value));
};

// 清除查询
const clearQuery = () => {
  dslQuery.value = '';
  showSuggestions.value = false;
  emit('clear-search');
};

// 选择示例查询
const selectExample = (query: string) => {
  dslQuery.value = query;
  showSuggestions.value = false;
};

// 输入框焦点事件
const handleInputFocus = () => {
  showSuggestions.value = true;
  updateSuggestions();
};

// 输入事件
const handleInput = () => {
  showSuggestions.value = true;
  updateSuggestions();
};

// 输入框失去焦点事件
const handleInputBlur = () => {
  // 延迟关闭建议，以便可以点击建议
  setTimeout(() => {
    showSuggestions.value = false;
  }, 300); // 增加延迟时间，确保有足够时间点击
};

// 输入框键盘事件
const handleInputKeydown = (event: KeyboardEvent) => {
  if (event.key === 'Enter') {
    executeDSLQuery();
  } else if (event.key === 'Escape') {
    showSuggestions.value = false;
  }
};

// 测试强制更新建议
const forceUpdateSuggestions = () => {
  // 强制重置建议状态
  showSuggestions.value = true;
  suggestions.value = [];
  
  // 应用默认建议
  suggestions.value = [
    { value: 'status == "200"', description: '查找所有成功请求', type: 'example' },
    { value: 'method == "POST"', description: '查找POST请求', type: 'example' },
    { value: 'status', description: t('modules.proxy.dsl.field_status'), type: 'field' },
    { value: 'method', description: t('modules.proxy.dsl.field_method'), type: 'field' }
  ];
  
  // 确保DOM已更新
  nextTick(() => {
    const suggestionsEl = document.querySelector('.dsl-suggestions');
  
    if (!suggestionsEl) {
      console.error('建议DOM元素不存在！');
    }
  });
};

// 创建测试函数
// @ts-ignore - 忽略类型检查
window.testDSLSuggestions = forceUpdateSuggestions;

// 初始化时从localStorage加载搜索历史
onMounted(() => {
  // 从localStorage加载搜索历史
  const savedHistory = localStorage.getItem('dsl_search_history');
  if (savedHistory) {
    try {
      searchHistory.value = JSON.parse(savedHistory);
    } catch (e) {
      console.error('Failed to parse search history:', e);
    }
  }
  
  // 初始化建议列表为默认建议
  suggestions.value = defaultSuggestions;
});

// 监听输入变化
watch(dslQuery, () => {
  if (showSuggestions.value) {
    updateSuggestions();
  }
});

// 计算显示建议的条件
const shouldShowSuggestions = computed(() => {
  const result = showSuggestions.value && suggestions.value.length > 0;
  return result;
});
</script>

<template>
  <div class="dsl-search-container">
    <div class="dsl-search-input-wrapper">
      <input 
        ref="inputRef"
        type="text" 
        v-model="dslQuery"
        :placeholder="t('modules.proxy.dsl.query_placeholder')"
        class="dsl-search-input"
        @keyup.enter="executeDSLQuery"
        @input="handleInput"
        @focus="handleInputFocus"
        @blur="handleInputBlur"
        @keydown="handleInputKeydown"
        spellcheck="false"
      />
      
      <!-- 调试按钮 -->
      <button 
        class="debug-button" 
        type="button" 
        title="调试建议"
        @click="forceUpdateSuggestions"
      >
        <i class="bx bx-bug"></i>
      </button>
      
      <!-- 输入建议下拉列表 (使用计算属性) -->
      <div 
        class="dsl-suggestions" 
        v-if="shouldShowSuggestions"
      >
        <div class="dsl-suggestions-debug">
          {{ suggestions.length }}个建议可用 | 状态: {{ showSuggestions ? '显示' : '隐藏' }}
        </div>
        <div 
          v-for="suggestion in suggestions" 
          :key="suggestion.value"
          class="dsl-suggestion-item"
          @mousedown.prevent="applySuggestion(suggestion)"
        >
          <div class="dsl-suggestion-value">
            <span 
              class="dsl-suggestion-icon" 
              :class="{
                'field-icon': suggestion.type === 'field',
                'function-icon': suggestion.type === 'function',
                'history-icon': suggestion.type === 'history',
                'example-icon': suggestion.type === 'example'
              }"
            >
              <i 
                class="bx" 
                :class="{
                  'bx-box': suggestion.type === 'field',
                  'bx-code-alt': suggestion.type === 'function',
                  'bx-history': suggestion.type === 'history',
                  'bx-bulb': suggestion.type === 'example'
                }"
              ></i>
            </span>
            <span>{{ suggestion.value }}</span>
          </div>
          <div class="dsl-suggestion-description">{{ suggestion.description }}</div>
        </div>
      </div>
      
      <div class="dsl-search-actions">
        <button 
          class="help-button" 
          @click="showHelp = !showHelp"
          :title="t('modules.proxy.dsl.toggle_help')"
        >
          <i class="bx bx-help-circle"></i>
        </button>
        <button 
          class="dsl-search-button" 
          @click="executeDSLQuery"
          :disabled="isLoading"
          :title="t('modules.proxy.dsl.search')"
        >
          <span v-if="isLoading">
            <i class="bx bx-loader bx-spin"></i>
          </span>
          <span v-else>
            <i class="bx bx-search"></i>
          </span>
        </button>
        <button 
          class="dsl-clear-button" 
          @click="clearQuery"
          :disabled="isLoading || !dslQuery"
          :title="t('modules.proxy.dsl.clear')"
        >
          <i class="bx bx-x"></i>
        </button>
      </div>
    </div>
    
    <!-- DSL帮助信息悬浮窗 -->
    <div v-if="showHelp" class="dsl-help-popup">
      <div class="dsl-help-popup-header">
        <h4>{{ t('modules.proxy.dsl.help_title') }}</h4>
        <button @click="showHelp = false" class="dsl-help-close">
          <i class="bx bx-x"></i>
        </button>
      </div>
      <div class="dsl-help-popup-content">
        <p>{{ t('modules.proxy.dsl.help_description') }}</p>
        
        <h5>{{ t('modules.proxy.dsl.examples_title') }}</h5>
        <div class="dsl-examples">
          <div 
            v-for="example in dslExamples" 
            :key="example.query"
            class="dsl-example"
            @click="selectExample(example.query)"
          >
            <div class="dsl-example-name">{{ example.name }}</div>
            <div class="dsl-example-query">{{ example.query }}</div>
          </div>
        </div>
        
        <h5>{{ t('modules.proxy.dsl.available_fields') }}</h5>
        <ul class="dsl-fields-list">
          <li><code>id</code> - {{ t('modules.proxy.dsl.field_id') }}</li>
          <li><code>flow_id</code> - {{ t('modules.proxy.dsl.field_flow_id') }}</li>
          <li><code>url</code> - {{ t('modules.proxy.dsl.field_url') }}</li>
          <li><code>path</code> - {{ t('modules.proxy.dsl.field_path') }}</li>
          <li><code>method</code> - {{ t('modules.proxy.dsl.field_method') }}</li>
          <li><code>host</code> - {{ t('modules.proxy.dsl.field_host') }}</li>
          <li><code>status</code> - {{ t('modules.proxy.dsl.field_status') }}</li>
          <li><code>length</code> - {{ t('modules.proxy.dsl.field_length') }}</li>
          <li><code>content_type</code> - {{ t('modules.proxy.dsl.field_content_type') }}</li>
          <li><code>request</code> - {{ t('modules.proxy.dsl.field_request') }}</li>
          <li><code>request_body</code> - {{ t('modules.proxy.dsl.field_request_body') }}</li>
          <li><code>response</code> - {{ t('modules.proxy.dsl.field_response') }}</li>
          <li><code>response_body</code> - {{ t('modules.proxy.dsl.field_response_body') }}</li>
          <li><code>status_reason</code> - {{ t('modules.proxy.dsl.field_status_reason') }}</li>
        </ul>
        
        <h5>{{ t('modules.proxy.dsl.functions_title') }}</h5>
        <ul class="dsl-functions-list">
          <li><code>contains(field, "value")</code> - {{ t('modules.proxy.dsl.function_contains') }}</li>
          <li><code>regex(field, "pattern")</code> - {{ t('modules.proxy.dsl.function_regex') }}</li>
          <li><code>field == "value"</code> - {{ t('modules.proxy.dsl.function_equals') }}</li>
          <li><code>field != "value"</code> - {{ t('modules.proxy.dsl.function_not_equals') }}</li>
          <li><code>expression1 && expression2</code> - {{ t('modules.proxy.dsl.function_and') }}</li>
          <li><code>expression1 || expression2</code> - {{ t('modules.proxy.dsl.function_or') }}</li>
        </ul>
      </div>
    </div>
  </div>
</template>

<style scoped>
/* 使用外部样式文件定义样式 */
</style> 