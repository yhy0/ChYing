<script setup lang="ts">
import { ref, reactive, onMounted, nextTick } from 'vue';
import { useI18n } from 'vue-i18n';
import { message } from '../../../utils/message';
// @ts-ignore
import { GetProxyConfigure, ConfigureScope, GetJieConfigureFileContent, ModifyJieConfigureFileContent } from "../../../../bindings/github.com/yhy0/ChYing/app.js";
import { type ScopeRule, type ProxyConfig, type CompleteConfig } from '../../../types';
import yaml from 'js-yaml';

// 引入子组件
import ProxySettingsTab from './ProxySettingsTab.vue';
import PluginsSettingsTab from './PluginsSettingsTab.vue';
import CollectionSettingsTab from './CollectionSettingsTab.vue';

const { t } = useI18n();

// 代理配置数据
const configValue = ref<ProxyConfig>({
  port: 0,
  exclude: [],
  include: [],
  filterSuffix: '',
});

// 完整配置数据
const completeConfig = ref<CompleteConfig | null>(null);
const fileBody = ref<string>('');
const activeTab = ref('proxy'); // 'proxy', 'plugins', 'collection'

// 状态
const isLoading = ref(false);

// 模态框状态
const showModal = ref(false);
const modalTitle = ref('');
const actionType = ref(''); // add 或 edit
const scopeType = ref(''); // exclude 或 include

// 当前编辑的规则
const currentRule = reactive<ScopeRule>({
  id: '',
  type: '',
  prefix: '',
  enabled: true,
  regexp: false
});

// 正则编辑模态框状态
const showRegexModal = ref(false);
const currentRegexType = ref('');
const currentRegexValue = ref('');
const currentRegexIndex = ref(-1);

// 子组件引用
const proxySettingsTabRef = ref(null);

// 初始化：加载代理配置
onMounted(async () => {
  await loadConfig();
  await loadCompleteConfig();
  
  nextTick(() => {
    if (proxySettingsTabRef.value) {
      // @ts-ignore
      proxySettingsTabRef.value.initResizableColumns();
    }
  });
});

// 加载完整配置
const loadCompleteConfig = async () => {
  try {
    const result = await GetJieConfigureFileContent();
    fileBody.value = result;
    // 尝试解析YAML为对象
    try {
      const parsed = yaml.load(result);
      completeConfig.value = parsed;
    } catch (e) {
      console.error('Failed to parse config file:', e);
    }
  } catch (error) {
    message.error(t('common.status.load_failed'));
    console.error('Failed to load complete configuration:', error);
  }
};

// 保存完整配置
const saveCompleteConfig = async () => {
  try {
    if (completeConfig.value) {
      const configStr = yaml.dump(completeConfig.value);
      await ModifyJieConfigureFileContent(configStr);
      message.success(t('common.status.saved'));
    }
  } catch (error) {
    message.error(t('common.status.save_failed'));
    console.error('Failed to save complete configuration:', error);
  }
};

// 加载配置
const loadConfig = async () => {
  isLoading.value = true;
  try {
    const res = await GetProxyConfigure();
    if (res) {
    configValue.value.port = res.port || 0;
      configValue.value.exclude = (res.exclude || []).filter(item => item !== null).map(item => ({
        ...item,
        id: String(item.id)
      }));
      configValue.value.include = (res.include || []).filter(item => item !== null).map(item => ({
        ...item,
        id: String(item.id)
      }));
    configValue.value.filterSuffix = res.filterSuffix?.join(',') || '';
    }
  } catch (error) {
    message.error(t('common.status.load_failed'));
    console.error('Failed to load proxy configuration:', error);
  } finally {
    isLoading.value = false;
  }
};

// 保存后缀过滤器
const saveFilterSuffix = async () => {
  try {
    // 分割过滤后缀并去除空值
    const suffixArray = configValue.value.filterSuffix
      .split(',')
      .map(s => s.trim())
      .filter(Boolean);
      
    // 更新完整配置
    if (completeConfig.value) {
      completeConfig.value.mitmproxy.filterSuffix = suffixArray;
      await saveCompleteConfig();
    }
      
    // 调用后端API保存
    message.success(t('common.status.saved'));
  } catch (error) {
    message.error(t('common.status.save_failed'));
    console.error('Failed to save filter suffix:', error);
  }
};

// 打开添加规则模态框
const openAddModal = (type: string) => {
  // 重置当前规则
  Object.assign(currentRule, {
    id: Date.now().toString(),
    type: type,
    prefix: '',
    enabled: true,
    regexp: false
  });
  
  actionType.value = 'add';
  scopeType.value = type;
  modalTitle.value = t('modules.proxy.filter.add_rule');
  showModal.value = true;
};

// 打开编辑规则模态框
const openEditModal = (rule: ScopeRule, type: string) => {
  // 复制规则数据到当前规则
  Object.assign(currentRule, rule);
  
  actionType.value = 'edit';
  scopeType.value = type;
  modalTitle.value = t('modules.proxy.filter.edit_rule');
  showModal.value = true;
};

// 关闭模态框
const closeModal = () => {
  showModal.value = false;
};

// 保存规则
const saveRule = async () => {
  if (!currentRule.prefix.trim()) {
    message.warning(t('modules.proxy.filter.prefix_required'));
    return;
  }
  
  try {
    if (actionType.value === 'add') {
      // 添加新规则
      if (scopeType.value === 'exclude') {
        configValue.value.exclude.push({...currentRule});
      } else {
        configValue.value.include.push({...currentRule});
      }
    } else {
      // 更新现有规则
      const rules = scopeType.value === 'exclude' ? configValue.value.exclude : configValue.value.include;
      const index = rules.findIndex(r => r.id === currentRule.id);
      if (index !== -1) {
        rules[index] = {...currentRule};
      }
    }
    
    // 更新完整配置
    if (completeConfig.value) {
      if (scopeType.value === 'exclude') {
        completeConfig.value.mitmproxy.exclude = configValue.value.exclude.map(r => r.prefix);
      } else {
        completeConfig.value.mitmproxy.include = configValue.value.include.map(r => r.prefix);
      }
      await saveCompleteConfig();
    }
    
    // 保存到后端
    await saveRules(scopeType.value);
    message.success(t('common.status.saved'));
    closeModal();
  } catch (error) {
    message.error(t('common.status.save_failed'));
    console.error('Failed to save rule:', error);
  }
};

// 删除规则
const deleteRule = async (rule: ScopeRule, type: string) => {
  try {
    if (type === 'exclude') {
      configValue.value.exclude = configValue.value.exclude.filter(item => item.id !== rule.id);
      
      // 更新完整配置
      if (completeConfig.value) {
        completeConfig.value.mitmproxy.exclude = configValue.value.exclude.map(r => r.prefix);
      }
    } else {
      configValue.value.include = configValue.value.include.filter(item => item.id !== rule.id);
      
      // 更新完整配置
      if (completeConfig.value) {
        completeConfig.value.mitmproxy.include = configValue.value.include.map(r => r.prefix);
      }
    }
    
    await saveCompleteConfig();
    await saveRules(type);
    message.success(t('common.status.deleted'));
  } catch (error) {
    message.error(t('common.status.delete_failed'));
    console.error('Failed to delete rule:', error);
  }
};

// 更新规则切换状态
const updateRuleToggle = async (rule: ScopeRule, type: string, field: string, event: Event) => {
  const checked = (event.target as HTMLInputElement).checked;
  try {
    const rules = type === 'exclude' ? configValue.value.exclude : configValue.value.include;
    const index = rules.findIndex(item => item.id === rule.id);
    
    if (index !== -1) {
      if (field === 'enabled') {
        rules[index].enabled = checked;
      } else if (field === 'regexp') {
        rules[index].regexp = checked;
      }
    }
    
    // 保存到后端
    await saveRules(type);
  } catch (error) {
    message.error(t('common.status.save_failed'));
    console.error('Failed to update rule:', error);
  }
};

// 保存规则列表
const saveRules = async (type: string) => {
  try {
    const rules = type === 'exclude' ? configValue.value.exclude : configValue.value.include;
    // 这里调用ConfigureScope更新所有规则
    // 注：模拟保存，实际应调用后端API
    // const result = await ConfigureScope("update", type, rules);
    message.success(t('common.status.saved'));
    return true;
  } catch (error) {
    message.error(t('common.status.save_failed'));
    console.error('Failed to save rules:', error);
    return false;
  }
};

// 更新HTTP配置
const updateHttpConfig = async () => {
  if (!completeConfig.value) return;
  
  try {
    await saveCompleteConfig();
    message.success(t('common.status.saved'));
  } catch (error) {
    message.error(t('common.status.save_failed'));
    console.error('Failed to update HTTP config:', error);
  }
};

// 更新插件配置
const updatePluginConfig = async () => {
  if (!completeConfig.value) return;
  
  try {
    await saveCompleteConfig();
    message.success(t('common.status.saved'));
  } catch (error) {
    message.error(t('common.status.save_failed'));
    console.error('Failed to update plugin config:', error);
  }
};

// 打开正则编辑模态框
const openRegexModal = (type: string, index: number) => {
  if (!completeConfig.value || !completeConfig.value.collection[type]) return;
  
  currentRegexType.value = type;
  currentRegexIndex.value = index;
  currentRegexValue.value = completeConfig.value.collection[type][index] || '';
  showRegexModal.value = true;
};

// 关闭正则编辑模态框
const closeRegexModal = () => {
  showRegexModal.value = false;
};

// 添加新正则
const addNewRegex = (type: string) => {
  if (!completeConfig.value) return;
  
  if (!completeConfig.value.collection[type]) {
    completeConfig.value.collection[type] = [];
  }
  
  completeConfig.value.collection[type].push('');
  
  // 打开编辑模态框
  openRegexModal(type, completeConfig.value.collection[type].length - 1);
};

// 删除正则
const deleteRegex = async (type: string, index: number) => {
  if (!completeConfig.value || !completeConfig.value.collection[type]) return;
  
  completeConfig.value.collection[type].splice(index, 1);
  
  try {
    await saveCompleteConfig();
    message.success(t('common.status.deleted'));
  } catch (error) {
    message.error(t('common.status.delete_failed'));
    console.error('Failed to delete regex:', error);
  }
};

// 保存正则
const saveRegex = async () => {
  if (!completeConfig.value || !completeConfig.value.collection[currentRegexType.value]) return;
  
  if (currentRegexIndex.value >= 0 && currentRegexIndex.value < completeConfig.value.collection[currentRegexType.value].length) {
    completeConfig.value.collection[currentRegexType.value][currentRegexIndex.value] = currentRegexValue.value;
    
    try {
      await saveCompleteConfig();
      message.success(t('common.status.saved'));
      closeRegexModal();
    } catch (error) {
      message.error(t('common.status.save_failed'));
      console.error('Failed to save regex:', error);
    }
  }
};
</script>

<template>
  <div class="space-y-6 max-w-4xl mx-auto">
    <!-- 标签导航 -->
    <div class="border-b border-gray-200 dark:border-gray-700">
      <nav class="flex -mb-px">
        <button 
          @click="activeTab = 'proxy'" 
          class="py-3 px-5 font-medium text-sm border-b-2 transition-colors"
          :class="activeTab === 'proxy' ? 'border-amber-500 text-amber-500 dark:text-amber-400' : 'border-transparent text-gray-500 hover:text-gray-700 dark:text-gray-400 dark:hover:text-gray-300'"
        >
          {{ t('modules.proxy.settings') }}
        </button>
        <button 
          @click="activeTab = 'plugins'" 
          class="py-3 px-5 font-medium text-sm border-b-2 transition-colors"
          :class="activeTab === 'plugins' ? 'border-amber-500 text-amber-500 dark:text-amber-400' : 'border-transparent text-gray-500 hover:text-gray-700 dark:text-gray-400 dark:hover:text-gray-300'"
        >
          {{ t('modules.plugins.settings') }}
        </button>
        <button 
          @click="activeTab = 'collection'" 
          class="py-3 px-5 font-medium text-sm border-b-2 transition-colors"
          :class="activeTab === 'collection' ? 'border-amber-500 text-amber-500 dark:text-amber-400' : 'border-transparent text-gray-500 hover:text-gray-700 dark:text-gray-400 dark:hover:text-gray-300'"
        >
          {{ t('modules.collection.settings') }}
        </button>
      </nav>
    </div>
    
    <!-- 代理设置 -->
    <ProxySettingsTab 
      v-if="activeTab === 'proxy'" 
      ref="proxySettingsTabRef"
      :config-value="configValue"
      :complete-config="completeConfig"
      @save-filter-suffix="saveFilterSuffix"
      @update-http-config="updateHttpConfig"
      @open-add-modal="openAddModal"
      @open-edit-modal="openEditModal"
      @delete-rule="deleteRule"
      @update-rule-toggle="updateRuleToggle"
    />
    
    <!-- 插件设置部分 -->
    <PluginsSettingsTab 
      v-else-if="activeTab === 'plugins'" 
      :complete-config="completeConfig"
      @update-plugin-config="updatePluginConfig"
    />
    
    <!-- 收集设置部分 -->
    <CollectionSettingsTab 
      v-else-if="activeTab === 'collection'" 
      :complete-config="completeConfig"
      @save-complete-config="saveCompleteConfig"
      @add-new-regex="addNewRegex"
      @open-regex-modal="openRegexModal"
      @delete-regex="deleteRegex"
    />
    
    <!-- 规则编辑模态框 -->
    <div v-if="showModal" class="fixed inset-0 bg-black bg-opacity-50 z-50 flex items-center justify-center">
      <div class="bg-white dark:bg-[#282838] rounded-lg shadow-xl w-full max-w-md overflow-hidden" @click.stop>
        <!-- 模态框标题 -->
        <div class="px-6 py-4 border-b border-gray-200 dark:border-gray-700 flex items-center justify-between">
          <h3 class="text-lg font-medium text-gray-800 dark:text-gray-200">{{ modalTitle }}</h3>
          <button 
            @click="closeModal"
            class="text-gray-500 hover:text-red-500 focus:outline-none"
          >
            <i class="bx bx-x text-xl"></i>
          </button>
        </div>
        
        <!-- 模态框内容 -->
        <div class="px-6 py-5 space-y-4">
          <!-- 前缀输入 -->
          <div>
            <label class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">
              {{ t('modules.proxy.filter.prefix') }}
            </label>
            <input
              v-model="currentRule.prefix"
              type="text"
              class="w-full rounded-md border-gray-300 dark:border-gray-600 bg-white dark:bg-[#32324c] shadow-sm px-3 py-2 text-gray-800 dark:text-gray-200"
              :placeholder="t('modules.proxy.filter.prefix_placeholder')"
            />
            <p class="text-xs text-gray-500 dark:text-gray-400 mt-1">
              {{ t('modules.proxy.filter.prefix_hint') }}
            </p>
          </div>
          
          <!-- 启用状态切换 -->
          <div class="flex items-center space-x-2">
            <input
              id="rule-enabled"
              v-model="currentRule.enabled"
              type="checkbox"
              class="rounded border-gray-300 text-green-500 focus:ring-green-500"
            />
            <label for="rule-enabled" class="text-sm font-medium text-gray-700 dark:text-gray-300">
              {{ t('common.ui.enabled') }}
            </label>
          </div>
          
          <!-- 正则表达式切换 -->
          <div class="flex items-center space-x-2">
            <input
              id="rule-regexp"
              v-model="currentRule.regexp"
              type="checkbox"
              class="rounded border-gray-300 text-blue-500 focus:ring-blue-500"
            />
            <label for="rule-regexp" class="text-sm font-medium text-gray-700 dark:text-gray-300">
              {{ t('common.ui.regex') }}
            </label>
          </div>
        </div>
        
        <!-- 模态框底部按钮 -->
        <div class="px-6 py-4 bg-gray-50 dark:bg-[#1e1e2e] border-t border-gray-200 dark:border-gray-700 flex justify-end space-x-3">
          <button 
            @click="closeModal"
            class="px-4 py-2 border border-gray-300 dark:border-gray-600 rounded-md text-sm font-medium text-gray-700 dark:text-gray-300 bg-white dark:bg-[#282838] hover:bg-gray-50 dark:hover:bg-gray-700 transition-colors"
          >
            {{ t('common.actions.cancel') }}
          </button>
          <button 
            @click="saveRule"
            class="px-4 py-2 bg-amber-500 hover:bg-amber-600 text-white rounded-md text-sm font-medium transition-colors"
          >
            {{ t('common.actions.save') }}
          </button>
        </div>
      </div>
    </div>
    
    <!-- 正则编辑模态框 -->
    <div v-if="showRegexModal" class="fixed inset-0 bg-black bg-opacity-50 z-50 flex items-center justify-center">
      <div class="bg-white dark:bg-[#282838] rounded-lg shadow-xl w-full max-w-2xl overflow-hidden" @click.stop>
        <!-- 模态框标题 -->
        <div class="px-6 py-4 border-b border-gray-200 dark:border-gray-700 flex items-center justify-between">
          <h3 class="text-lg font-medium text-gray-800 dark:text-gray-200">{{ t('settings.edit_regex') }}</h3>
          <button 
            @click="closeRegexModal"
            class="text-gray-500 hover:text-red-500 focus:outline-none"
          >
            <i class="bx bx-x text-xl"></i>
          </button>
        </div>
        
        <!-- 模态框内容 -->
        <div class="px-6 py-5 space-y-4">
          <!-- 正则输入 -->
          <div>
            <label class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">
              {{ t('common.ui.regex') }}
            </label>
            <textarea
              spellcheck="false"
              v-model="currentRegexValue"
              class="w-full h-48 rounded-md border-gray-300 dark:border-gray-600 bg-white dark:bg-[#32324c] shadow-sm px-3 py-2 text-gray-800 dark:text-gray-200 font-mono"
              placeholder="输入正则表达式"
            ></textarea>
            <p class="text-xs text-gray-500 dark:text-gray-400 mt-1">
              请输入有效的正则表达式
            </p>
          </div>
        </div>
        
        <!-- 模态框底部按钮 -->
        <div class="px-6 py-4 bg-gray-50 dark:bg-[#1e1e2e] border-t border-gray-200 dark:border-gray-700 flex justify-end space-x-3">
          <button 
            @click="closeRegexModal"
            class="px-4 py-2 border border-gray-300 dark:border-gray-600 rounded-md text-sm font-medium text-gray-700 dark:text-gray-300 bg-white dark:bg-[#282838] hover:bg-gray-50 dark:hover:bg-gray-700 transition-colors"
          >
            {{ t('common.actions.cancel') }}
          </button>
          <button 
            @click="saveRegex"
            class="px-4 py-2 bg-amber-500 hover:bg-amber-600 text-white rounded-md text-sm font-medium transition-colors"
          >
            {{ t('common.actions.save') }}
          </button>
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
/* 美化滚动条 */
::-webkit-scrollbar {
  width: 6px;
  height: 6px;
}

::-webkit-scrollbar-track {
  background: transparent;
}

::-webkit-scrollbar-thumb {
  background-color: rgba(156, 163, 175, 0.5);
  border-radius: 3px;
}

::-webkit-scrollbar-thumb:hover {
  background-color: rgba(156, 163, 175, 0.7);
}

/* 自定义单选框和复选框样式 */
input[type="checkbox"] {
  color-scheme: light dark;
  cursor: pointer;
}

/* 确保暗色模式下勾选框可见 */
.dark input[type="checkbox"] {
  background-color: #32324c;
  border-color: #4b5563;
}

/* 设置选中状态样式 */
input[type="checkbox"]:checked {
  background-image: url("data:image/svg+xml,%3csvg viewBox='0 0 16 16' fill='white' xmlns='http://www.w3.org/2000/svg'%3e%3cpath d='M5.707 7.293a1 1 0 0 0-1.414 1.414l2 2a1 1 0 0 0 1.414 0l4-4a1 1 0 0 0-1.414-1.414L7 8.586 5.707 7.293z'/%3e%3c/svg%3e");
  background-size: 100% 100%;
  background-position: center;
  background-repeat: no-repeat;
}

input.text-green-500:checked {
  background-color: #10b981;
  border-color: #10b981;
}

input.text-blue-500:checked {
  background-color: #3b82f6;
  border-color: #3b82f6;
}

input:focus-visible {
  outline: none;
}
</style>