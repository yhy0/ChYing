<script setup lang="ts">
import { ref, onMounted, computed, h, watch } from 'vue';
import { useI18n } from 'vue-i18n';
import HttpTrafficTable from '../../common/HttpTrafficTable.vue';
import type { HttpTrafficColumn } from '../../../types';
// @ts-ignore - 忽略类型检查，因为缺少声明文件
import { GetProxyListeners, SaveProxyListener, DeleteProxyListener, ToggleProxyListener, GetProxyStatus, CheckPortAvailable, RestartProxyServer } from '../../../../bindings/github.com/yhy0/ChYing/app.js';

// 使用国际化
const { t } = useI18n();

// 发送通知事件
const emit = defineEmits(['notify']);

// 监听器数据结构
interface ProxyListener {
  id: string;
  host: string;
  port: number;
  enabled: boolean;
  running: boolean;
}

// 当前监听器列表
const listeners = ref<ProxyListener[]>([]);

// 加载状态
const isLoading = ref(false);

// 是否显示添加/编辑对话框
const showListenerDialog = ref(false);

// 当前编辑的监听器
const currentListener = ref<ProxyListener>({
  id: '',
  host: '127.0.0.1',
  port: 9080,
  enabled: true,
  running: false
});

// 是否是编辑模式
const isEditMode = ref(false);

// 端口检测状态
const portCheckStatus = ref<'idle' | 'checking' | 'available' | 'occupied'>('idle');
const portCheckMessage = ref('');

// 检测端口是否可用
const checkPort = async (port: number) => {
  if (!port || port < 1 || port > 65535) {
    portCheckStatus.value = 'idle';
    portCheckMessage.value = '';
    return;
  }

  // 如果是编辑模式且端口未改变，跳过检测
  if (isEditMode.value) {
    const originalListener = listeners.value.find(l => l.id === currentListener.value.id);
    if (originalListener && originalListener.port === port) {
      portCheckStatus.value = 'available';
      portCheckMessage.value = '';
      return;
    }
  }

  portCheckStatus.value = 'checking';
  portCheckMessage.value = t('modules.listeners.checkingPort');

  try {
    const response = await CheckPortAvailable(port);
    if (response.error) {
      throw new Error(response.error);
    }

    if (response.data?.available) {
      portCheckStatus.value = 'available';
      portCheckMessage.value = t('modules.listeners.portAvailable');
    } else {
      portCheckStatus.value = 'occupied';
      portCheckMessage.value = t('modules.listeners.portOccupied');
    }
  } catch (error) {
    console.error('Failed to check port:', error);
    portCheckStatus.value = 'idle';
    portCheckMessage.value = '';
  }
};

// 监听端口变化，自动检测
let portCheckTimeout: ReturnType<typeof setTimeout> | null = null;
watch(() => currentListener.value.port, (newPort) => {
  if (portCheckTimeout) {
    clearTimeout(portCheckTimeout);
  }
  portCheckStatus.value = 'idle';
  portCheckMessage.value = '';

  // 延迟检测，避免频繁调用
  portCheckTimeout = setTimeout(() => {
    checkPort(newPort);
  }, 500);
});

// 加载监听器列表
const loadListeners = async () => {
  isLoading.value = true;
  try {
    const response = await GetProxyListeners();

    if (response.error) {
      throw new Error(response.error);
    }

    listeners.value = response.data || [];

    // 获取代理状态以更新 running 状态
    const statusResponse = await GetProxyStatus();
    if (!statusResponse.error && statusResponse.data) {
      // 更新监听器的运行状态
      const status = statusResponse.data;
      listeners.value.forEach(listener => {
        // 如果代理正在运行且监听器已启用，则标记为运行中
        listener.running = status.running && listener.enabled;
      });
    }

    emit('notify', t('modules.listeners.listenersLoaded', { count: listeners.value.length }));
  } catch (error) {
    console.error('Failed to load proxy listeners:', error);
    emit('notify', t('modules.listeners.loadError', { error: String(error) }));
  } finally {
    isLoading.value = false;
  }
};

// 保存监听器
const saveListener = async () => {
  isLoading.value = true;
  try {
    // 如果是新增，生成 ID
    if (!isEditMode.value) {
      currentListener.value.id = `listener-${Date.now()}`;
    }

    const response = await SaveProxyListener(currentListener.value);

    if (response.error) {
      throw new Error(response.error);
    }

    // 更新本地列表
    if (isEditMode.value) {
      const index = listeners.value.findIndex(l => l.id === currentListener.value.id);
      if (index !== -1) {
        listeners.value[index] = { ...currentListener.value };
      }
    } else {
      listeners.value.push({ ...currentListener.value });
    }

    showListenerDialog.value = false;
    emit('notify', t(`modules.listeners.${isEditMode.value ? 'listenerUpdated' : 'listenerAdded'}`));

    // 保存后自动重启代理服务器以应用新配置
    await restartProxyAfterSave();
  } catch (error) {
    console.error('Failed to save proxy listener:', error);
    emit('notify', t('modules.listeners.saveError', { error: String(error) }));
  } finally {
    isLoading.value = false;
  }
};

// 保存后自动重启代理（静默重启，不显示加载状态）
const restartProxyAfterSave = async () => {
  try {
    const response = await RestartProxyServer();
    if (response.error) {
      console.error('Failed to restart proxy after save:', response.error);
      return;
    }
    // 等待一段时间后刷新监听器列表
    setTimeout(() => {
      loadListeners();
    }, 1000);
  } catch (error) {
    console.error('Failed to restart proxy after save:', error);
  }
};

// 删除监听器
const deleteListener = async (listenerId: string) => {
  isLoading.value = true;
  try {
    const response = await DeleteProxyListener(listenerId);

    if (response.error) {
      throw new Error(response.error);
    }

    listeners.value = listeners.value.filter(l => l.id !== listenerId);
    emit('notify', t('modules.listeners.listenerDeleted'));
  } catch (error) {
    console.error('Failed to delete proxy listener:', error);
    emit('notify', t('modules.listeners.deleteError', { error: String(error) }));
  } finally {
    isLoading.value = false;
  }
};

// 切换监听器启用状态
const toggleListenerEnabled = async (listener: ProxyListener) => {
  const newEnabled = !listener.enabled;

  try {
    const response = await ToggleProxyListener(listener.id, newEnabled);

    if (response.error) {
      throw new Error(response.error);
    }

    listener.enabled = newEnabled;
    emit('notify', t('modules.listeners.listenerToggled', {
      enabled: newEnabled ? t('modules.listeners.enabled') : t('modules.listeners.disabled')
    }));
  } catch (error) {
    console.error('Failed to toggle proxy listener:', error);
    emit('notify', t('modules.listeners.toggleError', { error: String(error) }));
  }
};

// 添加监听器
const addListener = () => {
  isEditMode.value = false;
  currentListener.value = {
    id: '',
    host: '127.0.0.1',
    port: 9080,
    enabled: true,
    running: false
  };
  portCheckStatus.value = 'idle';
  portCheckMessage.value = '';
  showListenerDialog.value = true;
  // 打开对话框后检测默认端口
  setTimeout(() => checkPort(currentListener.value.port), 100);
};

// 编辑监听器
const editListener = (listener: ProxyListener) => {
  isEditMode.value = true;
  currentListener.value = { ...listener };
  portCheckStatus.value = 'available'; // 编辑时默认端口是可用的
  portCheckMessage.value = '';
  showListenerDialog.value = true;
};

// 取消对话框
const cancelListenerDialog = () => {
  showListenerDialog.value = false;
  portCheckStatus.value = 'idle';
  portCheckMessage.value = '';
};

// 计算保存按钮是否可用
const canSave = computed(() => {
  return currentListener.value.host &&
    currentListener.value.port &&
    currentListener.value.port >= 1 &&
    currentListener.value.port <= 65535 &&
    portCheckStatus.value !== 'occupied' &&
    portCheckStatus.value !== 'checking';
});

// 组件加载后获取监听器列表
onMounted(() => {
  loadListeners();
});

// 数据转换函数：将监听器转换为表格格式（与 ProxyMatchReplacePanel 保持一致）
const transformedListeners = computed(() => {
  return listeners.value.map((listener, index) => ({
    // 使用数字索引作为 id，因为 HttpTrafficItem.id 需要是 number 类型
    id: index + 1,
    originalId: listener.id, // 保留原始字符串 ID
    index: index,
    interface: `${listener.host}:${listener.port}`,
    enabled: listener.enabled,
    running: listener.running,
    port: listener.port,
    // HttpTrafficItem 必需字段的默认值
    method: 'LISTENER' as const,
    url: `${listener.host}:${listener.port}`,
    host: listener.host,
    path: listener.id,
    status: listener.running ? 200 : 0,
    length: 0,
    size: 0,
    mimeType: '',
    extension: '',
    title: '',
    ip: '',
    note: '',
    timestamp: Date.now(), // 与 ProxyMatchReplacePanel 保持一致，使用数字类型
  }));
});

// 当前选中的监听器
const selectedListener = ref<any>(null);

// 列定义 - 调整列宽以适应内容
const listenerColumns = computed<HttpTrafficColumn<any>[]>(() => [
  {
    id: 'running',
    name: t('modules.listeners.columns.running'),
    width: 70,
    cellRenderer: ({ item }) => {
      // 使用 originalId 查找原始监听器（因为 item.id 现在是数字索引）
      const originalListener = listeners.value.find(l => l.id === item.originalId);
      const isRunning = originalListener?.running || false;
      return h('div', {
        class: 'status-cell',
        title: isRunning ? t('modules.listeners.runningStatus') : t('modules.listeners.stoppedStatus')
      }, [
        h('span', {
          class: ['status-indicator', isRunning ? 'status-running' : 'status-stopped']
        })
      ]);
    }
  },
  {
    id: 'interface',
    name: t('modules.listeners.columns.interface'),
    width: 180,
    cellRenderer: ({ item }) => h('code', { class: 'interface-code' }, item.interface)
  },
  {
    id: 'enabled',
    name: t('modules.listeners.columns.enabled'),
    width: 70,
    cellRenderer: ({ item }) => {
      // 使用 originalId 查找原始监听器
      const originalListener = listeners.value.find(l => l.id === item.originalId);
      return h('label', {
        class: 'toggle-switch',
        onClick: (e: Event) => {
          e.preventDefault();
          e.stopPropagation();
          if (originalListener) {
            toggleListenerEnabled(originalListener);
          }
        },
        title: t('modules.listeners.toggleEnabled')
      }, [
        h('input', {
          type: 'checkbox',
          checked: originalListener?.enabled || false
        }),
        h('span', { class: 'toggle-slider' })
      ]);
    }
  },
  {
    id: 'actions',
    name: t('modules.listeners.columns.actions'),
    width: 80,
    cellRenderer: ({ item }) => h('div', { class: 'listener-actions' }, [
      h('button', {
        class: 'btn-icon edit-button',
        title: t('modules.listeners.edit'),
        onClick: (e: Event) => {
          e.stopPropagation();
          // 使用 originalId 查找原始监听器进行编辑
          const originalListener = listeners.value.find(l => l.id === item.originalId);
          if (originalListener) {
            editListener(originalListener);
          }
        }
      }, [h('i', { class: 'bx bx-edit' })]),
      h('button', {
        class: 'btn-icon delete-button',
        title: t('modules.listeners.delete'),
        onClick: (e: Event) => {
          e.stopPropagation();
          // 使用 originalId 删除监听器
          deleteListener(item.originalId);
        }
      }, [h('i', { class: 'bx bx-trash' })])
    ])
  }
]);

// 处理监听器选择
const handleListenerSelect = (listener: any) => {
  selectedListener.value = listener;
};

// 重启代理服务器
const isRestarting = ref(false);
const restartProxy = async () => {
  isRestarting.value = true;
  try {
    const response = await RestartProxyServer();
    if (response.error) {
      throw new Error(response.error);
    }
    emit('notify', t('modules.listeners.restartSuccess'));
    // 等待一段时间后刷新监听器列表
    setTimeout(() => {
      loadListeners();
    }, 1000);
  } catch (error) {
    console.error('Failed to restart proxy:', error);
    emit('notify', t('modules.listeners.restartError', { error: String(error) }));
  } finally {
    isRestarting.value = false;
  }
};
</script>

<template>
  <div class="plugin-container">
    <!-- 顶部控制栏 -->
    <div class="section-header">
      <h3>{{ t('modules.listeners.title') }}</h3>
      <div class="header-actions">
        <button class="btn btn-secondary" @click="restartProxy" :disabled="isLoading || isRestarting">
          <i class="bx" :class="isRestarting ? 'bx-loader-alt bx-spin' : 'bx-refresh'"></i>
          {{ t('modules.listeners.restart') }}
        </button>
        <button class="btn btn-secondary" @click="addListener" :disabled="isLoading">
          <i class="bx bx-plus"></i>
          {{ t('modules.listeners.add') }}
        </button>
      </div>
    </div>

    <!-- 监听器表格 -->
    <div class="plugin-table-wrapper listeners-table-wrapper" v-if="!isLoading">
      <HttpTrafficTable
        :items="transformedListeners"
        :selectedItem="selectedListener"
        :customColumns="listenerColumns"
        :tableClass="'compact-table listeners-table'"
        :containerHeight="'300px'"
        tableId="proxy-listeners-table"
        @select-item="handleListenerSelect"
      />
    </div>

    <!-- 全局加载指示器 -->
    <div v-if="isLoading" class="loading full-height">
      <div class="loading-spinner"></div>
      <span>{{ t('modules.listeners.loading') }}</span>
    </div>

    <!-- 添加/编辑监听器对话框 -->
    <div class="rule-dialog-overlay" v-if="showListenerDialog">
      <div class="rule-dialog">
        <div class="dialog-header">
          <h3>{{ isEditMode ? t('modules.listeners.editListener') : t('modules.listeners.addListener') }}</h3>
          <button class="dialog-close" @click="cancelListenerDialog">
            <i class="bx bx-x"></i>
          </button>
        </div>

        <div class="dialog-body">
          <div class="form-group">
            <label for="listener-host">{{ t('modules.listeners.host') }}</label>
            <input id="listener-host" type="text" v-model="currentListener.host"
              :placeholder="t('modules.listeners.hostPlaceholder')" />
          </div>

          <div class="form-group">
            <label for="listener-port">{{ t('modules.listeners.port') }}</label>
            <div class="port-input-wrapper">
              <input id="listener-port" type="number" v-model.number="currentListener.port"
                :placeholder="t('modules.listeners.portPlaceholder')" min="1" max="65535"
                :class="{ 'port-error': portCheckStatus === 'occupied' }" />
              <span v-if="portCheckStatus === 'checking'" class="port-status checking">
                <i class="bx bx-loader-alt bx-spin"></i>
              </span>
              <span v-else-if="portCheckStatus === 'available'" class="port-status available">
                <i class="bx bx-check-circle"></i>
              </span>
              <span v-else-if="portCheckStatus === 'occupied'" class="port-status occupied">
                <i class="bx bx-x-circle"></i>
              </span>
            </div>
            <p v-if="portCheckMessage" class="port-message" :class="portCheckStatus">{{ portCheckMessage }}</p>
          </div>

          <div class="form-group checkbox-group">
            <label class="checkbox-label">
              <input type="checkbox" v-model="currentListener.enabled" />
              <span>{{ t('modules.listeners.enableThisListener') }}</span>
            </label>
          </div>
        </div>

        <div class="dialog-footer">
          <button class="cancel-button" @click="cancelListenerDialog">{{ t('modules.listeners.cancel') }}</button>
          <button class="save-button" @click="saveListener" :disabled="!canSave">
            {{ t('modules.listeners.save') }}
          </button>
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
/* 表格容器 - 与 ProxyMatchReplacePanel 保持一致 */
.listeners-table-wrapper {
  flex: 1;
  overflow-y: auto;
  min-height: 200px;
  background: var(--glass-bg-tertiary);
  backdrop-filter: var(--glass-blur-sm);
  -webkit-backdrop-filter: var(--glass-blur-sm);
  border: 1px solid var(--glass-border-light);
  border-radius: var(--glass-radius-md);
}

/* 端口输入框样式 */
.port-input-wrapper {
  position: relative;
  display: flex;
  align-items: center;
}

.port-input-wrapper input {
  flex: 1;
  padding-right: 32px;
}

.port-input-wrapper input.port-error {
  border-color: #ef4444;
}

.port-status {
  position: absolute;
  right: 8px;
  display: flex;
  align-items: center;
  font-size: 16px;
}

.port-status.checking {
  color: #6b7280;
}

.port-status.available {
  color: #22c55e;
}

.port-status.occupied {
  color: #ef4444;
}

.port-message {
  margin-top: 4px;
  font-size: 12px;
}

.port-message.checking {
  color: #6b7280;
}

.port-message.available {
  color: #22c55e;
}

.port-message.occupied {
  color: #ef4444;
}

/* 复选框组样式 */
.checkbox-group {
  display: flex;
  align-items: center;
}

.checkbox-label {
  display: flex;
  align-items: center;
  gap: 8px;
  cursor: pointer;
  font-size: 14px;
  color: var(--color-text-primary);
}

.checkbox-label input[type="checkbox"] {
  width: 16px;
  height: 16px;
  cursor: pointer;
}
</style>
