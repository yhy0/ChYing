<script setup lang="ts">
import { ref, onMounted, computed, h } from 'vue';
import { useI18n } from 'vue-i18n';
import HttpTrafficTable from '../../common/HttpTrafficTable.vue';
import type { HttpTrafficColumn } from '../../../types';
// @ts-ignore - 忽略类型检查，因为缺少声明文件
import { GetMatchReplaceRules, SaveMatchReplaceRule, DeleteMatchReplaceRule, ApplyMatchReplaceRules } from '../../../../bindings/github.com/yhy0/ChYing/app.js';

// 使用国际化
const { t } = useI18n();

// 发送通知事件
const emit = defineEmits(['notify']);

// 匹配替换规则类型
type RuleType = 'request_header' | 'request_body' | 'response_header' | 'response_body' | 'request_first_line' | 'request_param_name' | 'request_param_value';

// 单条规则的数据结构
interface MatchReplaceRule {
  id: number;
  enabled: boolean;
  type: RuleType;
  name: string;
  match: string;
  replace: string;
  comment: string;
}

// 当前规则列表
const rules = ref<MatchReplaceRule[]>([]);

// 加载状态
const isLoading = ref(false);

// 是否显示添加/编辑规则对话框
const showRuleDialog = ref(false);

// 当前编辑的规则
const currentRule = ref<MatchReplaceRule>({
  id: 0,
  enabled: true,
  type: 'request_header',
  name: '',
  match: '',
  replace: '',
  comment: ''
});

// 是否是编辑模式
const isEditMode = ref(false);

// 规则类型列表
const ruleTypes = [
  { value: 'request_header', label: t('modules.matchReplace.ruleTypes.requestHeader') },
  { value: 'request_body', label: t('modules.matchReplace.ruleTypes.requestBody') },
  { value: 'response_header', label: t('modules.matchReplace.ruleTypes.responseHeader') },
  { value: 'response_body', label: t('modules.matchReplace.ruleTypes.responseBody') },
  { value: 'request_first_line', label: t('modules.matchReplace.ruleTypes.requestFirstLine') },
  { value: 'request_param_name', label: t('modules.matchReplace.ruleTypes.requestParamName') },
  { value: 'request_param_value', label: t('modules.matchReplace.ruleTypes.requestParamValue') }
];

// 获取规则类型显示名称
const getRuleTypeLabel = (type: RuleType) => {
  const ruleType = ruleTypes.find(t => t.value === type);
  return ruleType ? ruleType.label : type;
};

// 加载规则列表
const loadRules = async () => {
  isLoading.value = true;
  try {
    // 调用后端接口获取规则列表
    const response = await GetMatchReplaceRules();

    // 检查是否有错误
    if (response.error) {
      throw new Error(response.error);
    }

    // 从data字段中获取规则数据
    const rulesData = response.data || { rules: [], enabled: false, onlyScopedItems: false };
    rules.value = rulesData.rules || [];

    emit('notify', t('modules.matchReplace.rulesLoaded', { count: rules.value.length }));
  } catch (error) {
    console.error('Failed to load match replace rules:', error);
    emit('notify', t('modules.matchReplace.loadError', { error: String(error) }));
  } finally {
    isLoading.value = false;
  }
};

// 保存规则
const saveRule = async () => {
  isLoading.value = true;
  try {
    // 调用后端接口保存规则
    const response = await SaveMatchReplaceRule(currentRule.value);

    // 检查是否有错误
    if (response.error) {
      throw new Error(response.error);
    }

    // 更新本地规则列表
    if (isEditMode.value) {
      const index = rules.value.findIndex(r => r.id === currentRule.value.id);
      if (index !== -1) {
        rules.value[index] = { ...currentRule.value };
      }
    } else {
      // 获取新ID
      const newId = Math.max(0, ...rules.value.map(r => r.id)) + 1;
      currentRule.value.id = newId;
      rules.value.push({ ...currentRule.value });
    }

    showRuleDialog.value = false;
    emit('notify', t(`modules.matchReplace.${isEditMode.value ? 'ruleUpdated' : 'ruleAdded'}`));

    // 应用规则
    applyRules();
  } catch (error) {
    console.error('Failed to save match replace rule:', error);
    emit('notify', t('modules.matchReplace.saveError', { error: String(error) }));
  } finally {
    isLoading.value = false;
  }
};

// 删除规则
const deleteRule = async (ruleId: number) => {
  isLoading.value = true;
  try {
    // 调用后端接口删除规则
    const response = await DeleteMatchReplaceRule(ruleId);

    // 检查是否有错误
    if (response.error) {
      throw new Error(response.error);
    }

    // 更新本地规则列表
    rules.value = rules.value.filter(r => r.id !== ruleId);

    emit('notify', t('modules.matchReplace.ruleDeleted'));

    // 应用规则
    applyRules();
  } catch (error) {
    console.error('Failed to delete match replace rule:', error);
    emit('notify', t('modules.matchReplace.deleteError', { error: String(error) }));
  } finally {
    isLoading.value = false;
  }
};

// 应用规则
const applyRules = async () => {
  try {
    // 调用后端接口应用规则
    const response = await ApplyMatchReplaceRules({
      rules: rules.value
    });

    // 检查是否有错误
    if (response.error) {
      console.error("ApplyMatchReplaceRules返回错误:", response.error);
      throw new Error(response.error);
    }

    return true;
  } catch (error) {
    console.error('Failed to apply match replace rules:', error);
    emit('notify', t('modules.matchReplace.applyError', { error: String(error) }));

    return false;
  }
};

// 添加规则
const addRule = () => {
  isEditMode.value = false;
  currentRule.value = {
    id: 0,
    enabled: true,
    type: 'request_header',
    name: '',
    match: '',
    replace: '',
    comment: ''
  };
  showRuleDialog.value = true;
};

// 编辑规则
const editRule = (rule: MatchReplaceRule) => {
  isEditMode.value = true;
  currentRule.value = { ...rule };
  showRuleDialog.value = true;
};

// 上移规则
const moveRuleUp = async (index: number) => {
  if (index <= 0) return;

  // 交换位置
  const temp = rules.value[index];
  rules.value[index] = rules.value[index - 1];
  rules.value[index - 1] = temp;

  // 应用规则
  await applyRules();
};

// 下移规则
const moveRuleDown = async (index: number) => {
  if (index >= rules.value.length - 1) return;

  // 交换位置
  const temp = rules.value[index];
  rules.value[index] = rules.value[index + 1];
  rules.value[index + 1] = temp;

  // 应用规则
  await applyRules();
};

// 取消添加/编辑规则
const cancelRuleDialog = () => {
  showRuleDialog.value = false;
};

// 组件加载后获取规则列表
onMounted(() => {
  loadRules();
});

// 根据规则类型获取匹配示例
const getMatchExampleForType = (type: RuleType) => {
  switch (type) {
    case 'request_header':
      return 'User-Agent.*';
    case 'request_body':
      return 'password":"([^"]+)"';
    case 'response_header':
      return 'Server: (.*)';
    case 'response_body':
      return '<title>(.*?)</title>';
    case 'request_first_line':
      return 'POST /api/login';
    case 'request_param_name':
      return 'token';
    case 'request_param_value':
      return 'old_value';
    default:
      return '';
  }
};

// 根据规则类型获取替换示例
const getReplaceExampleForType = (type: RuleType) => {
  switch (type) {
    case 'request_header':
      return 'User-Agent: Mozilla/5.0 (ChYing-Inside)';
    case 'request_body':
      return 'password":"替换后的密码"';
    case 'response_header':
      return 'Server: Hidden';
    case 'response_body':
      return '<title>Modified Title</title>';
    case 'request_first_line':
      return 'POST /api/login-modified';
    case 'request_param_name':
      return 'modified_token';
    case 'request_param_value':
      return 'new_value';
    default:
      return '';
  }
};

// 保留规则级别的开关功能
// 切换规则启用状态
const toggleRuleEnabled = async (rule: MatchReplaceRule) => {
  rule.enabled = !rule.enabled;

  try {
    // 调用后端接口保存规则
    const response = await SaveMatchReplaceRule(rule);

    // 检查是否有错误
    if (response.error) {
      // 恢复状态
      rule.enabled = !rule.enabled;
      throw new Error(response.error);
    }

    emit('notify', t('modules.matchReplace.ruleToggled', {
      enabled: rule.enabled ? t('modules.matchReplace.enabled') : t('modules.matchReplace.disabled')
    }));

    // 应用规则
    applyRules();
  } catch (error) {
    // 恢复状态
    rule.enabled = !rule.enabled;

    console.error('Failed to toggle match replace rule:', error);
    emit('notify', t('modules.matchReplace.toggleError', { error: String(error) }));
  }
};

// 数据转换函数：将规则转换为 HttpTrafficItem 格式
const transformedRules = computed(() => {
  return rules.value.map((rule, index) => ({
    id: rule.id,
    enabled: rule.enabled,
    type: rule.type,
    name: rule.name,
    match: rule.match,
    replace: rule.replace,
    comment: rule.comment,
    index: index, // 添加索引用于排序操作
    // HttpTrafficItem 必需字段的默认值
    method: 'RULE' as const,
    url: rule.name || rule.match,
    status: rule.enabled ? 200 : 0,
    timestamp: Date.now(),
    size: 0,
    host: 'rule',
    path: rule.name || rule.match
  }));
});

// 当前选中的规则
const selectedRule = ref<any>(null);

// 列定义
const matchReplaceColumns = computed<HttpTrafficColumn<any>[]>(() => [
  {
    id: 'enabled',
    name: t('modules.matchReplace.columns.enabled'),
    width: 80,
    cellRenderer: ({ item }) => {
      // 找到原始规则对象以确保响应式
      const originalRule = rules.value.find(r => r.id === item.id);
      return h('label', {
        class: 'toggle-switch',
        onClick: (e: Event) => {
          e.preventDefault();
          e.stopPropagation();
          if (originalRule) {
            toggleRuleEnabled(originalRule);
          }
        },
        title: '启用或禁用此规则'
      }, [
        h('input', {
          type: 'checkbox',
          checked: originalRule?.enabled || false
        }),
        h('span', { class: 'toggle-slider' })
      ]);
    }
  },
  {
    id: 'type',
    name: t('modules.matchReplace.columns.item'),
    width: 120,
    cellRenderer: ({ item }) => h('span', {}, getRuleTypeLabel(item.type))
  },
  {
    id: 'name',
    name: t('modules.matchReplace.columns.name'),
    width: 150
  },
  {
    id: 'match',
    name: t('modules.matchReplace.columns.match'),
    width: 200,
    cellRenderer: ({ item }) => h('code', { class: 'code-snippet' }, item.match)
  },
  {
    id: 'replace',
    name: t('modules.matchReplace.columns.replace'),
    width: 200,
    cellRenderer: ({ item }) => h('code', { class: 'code-snippet' }, item.replace)
  },
  {
    id: 'regex',
    name: t('modules.matchReplace.columns.type'),
    width: 80,
    cellRenderer: () => h('span', {}, 'Regex')
  },
  {
    id: 'comment',
    name: t('modules.matchReplace.columns.comment'),
    width: 150
  },
  {
    id: 'actions',
    name: t('modules.matchReplace.columns.actions'),
    width: 160,
    cellRenderer: ({ item }) => h('div', { class: 'rule-actions' }, [
      h('button', {
        class: 'btn-icon edit-button',
        title: 'Edit',
        onClick: (e: Event) => {
          e.stopPropagation();
          editRule(item);
        }
      }, [h('i', { class: 'bx bx-edit' })]),
      h('button', {
        class: 'btn-icon up-button',
        title: 'Move Up',
        disabled: item.index === 0,
        onClick: (e: Event) => {
          e.stopPropagation();
          moveRuleUp(item.index);
        }
      }, [h('i', { class: 'bx bx-up-arrow-alt' })]),
      h('button', {
        class: 'btn-icon down-button',
        title: 'Move Down',
        disabled: item.index === rules.value.length - 1,
        onClick: (e: Event) => {
          e.stopPropagation();
          moveRuleDown(item.index);
        }
      }, [h('i', { class: 'bx bx-down-arrow-alt' })]),
      h('button', {
        class: 'btn-icon delete-button',
        title: 'Delete',
        onClick: (e: Event) => {
          e.stopPropagation();
          deleteRule(item.id);
        }
      }, [h('i', { class: 'bx bx-trash' })])
    ])
  }
]);

// 处理规则选择
const handleRuleSelect = (rule: any) => {
  selectedRule.value = rule;
};
</script>

<template>
  <div class="plugin-container">
    <!-- 顶部控制栏 -->
    <div class="section-header">
      <h3>{{ t('modules.matchReplace.title') }}</h3>
      <div class="header-actions">
        <button class="btn btn-secondary" @click="addRule" :disabled="isLoading">
          <i class="bx bx-plus"></i>
          {{ t('modules.matchReplace.addRule') }}
        </button>
      </div>
    </div>

    <!-- 规则表格 -->
    <div class="plugin-table-wrapper" v-if="!isLoading">
      <HttpTrafficTable
        :items="transformedRules"
        :selectedItem="selectedRule"
        :customColumns="matchReplaceColumns"
        :tableClass="'compact-table'"
        :containerHeight="'400px'"
        tableId="match-replace-rules-table"
        @select-item="handleRuleSelect"
      />
    </div>

    <!-- 全局加载指示器 -->
    <div v-if="isLoading" class="loading full-height">
      <div class="loading-spinner"></div>
      <span>{{ t('modules.matchReplace.loading') }}</span>
    </div>

    <!-- 添加/编辑规则对话框 -->
    <div class="rule-dialog-overlay" v-if="showRuleDialog">
      <div class="rule-dialog">
        <div class="dialog-header">
          <h3>{{ isEditMode ? t('modules.matchReplace.editRule') : t('modules.matchReplace.addRule') }}</h3>
          <button class="dialog-close" @click="cancelRuleDialog">
            <i class="bx bx-x"></i>
          </button>
        </div>

        <div class="dialog-body">
          <div class="form-group">
            <label for="rule-type">{{ t('modules.matchReplace.ruleType') }}</label>
            <select id="rule-type" v-model="currentRule.type">
              <option v-for="type in ruleTypes" :key="type.value" :value="type.value">{{ type.label }}</option>
            </select>
          </div>

          <div class="form-group">
            <label for="rule-name">{{ t('modules.matchReplace.ruleName') }}</label>
            <input id="rule-name" type="text" v-model="currentRule.name"
              :placeholder="t('modules.matchReplace.ruleNamePlaceholder')" />
          </div>

          <div class="form-group">
            <label for="rule-match">{{ t('modules.matchReplace.matchPattern') }}</label>
            <div class="input-with-example">
              <input id="rule-match" type="text" v-model="currentRule.match"
                :placeholder="t('modules.matchReplace.matchPlaceholder')" />
              <div class="input-example">{{ t('modules.matchReplace.example') }}: {{
                getMatchExampleForType(currentRule.type) }}</div>
            </div>
          </div>

          <div class="form-group">
            <label for="rule-replace">{{ t('modules.matchReplace.replacePattern') }}</label>
            <div class="input-with-example">
              <input id="rule-replace" type="text" v-model="currentRule.replace"
                :placeholder="t('modules.matchReplace.replacePlaceholder')" />
              <div class="input-example">{{ t('modules.matchReplace.example') }}: {{
                getReplaceExampleForType(currentRule.type) }}</div>
            </div>
          </div>

          <div class="form-group">
            <label for="rule-comment">{{ t('modules.matchReplace.comment') }}</label>
            <input id="rule-comment" type="text" v-model="currentRule.comment"
              :placeholder="t('modules.matchReplace.commentPlaceholder')" />
          </div>

          <div class="form-group checkbox-group">
            <label class="checkbox-label">
              <input type="checkbox" v-model="currentRule.enabled" />
              <span>{{ t('modules.matchReplace.enableThisRule') }}</span>
            </label>
          </div>
        </div>

        <div class="dialog-footer">
          <button class="cancel-button" @click="cancelRuleDialog">{{ t('modules.matchReplace.cancel') }}</button>
          <button class="save-button" @click="saveRule" :disabled="!currentRule.match">{{ t('modules.matchReplace.save')
            }}</button>
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
/* 规则表格样式现在由 compact-table.css 统一管理 */
</style>