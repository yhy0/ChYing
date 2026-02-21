<script setup lang="ts">
import { ref, onMounted } from 'vue';
import { useI18n } from 'vue-i18n';
import { useOastStore, type OASTProvider } from '../../../store/oast';

const { t } = useI18n();
const store = useOastStore();
const emit = defineEmits<{ (e: 'close'): void }>();

// 编辑表单
const showForm = ref(false);
const editingId = ref<string | null>(null);
const form = ref({
  name: '',
  type: 'interactsh',
  url: '',
  token: '',
});

onMounted(() => {
  store.loadProviders();
});

const openAddForm = () => {
  editingId.value = null;
  form.value = { name: '', type: 'interactsh', url: '', token: '' };
  showForm.value = true;
};

const openEditForm = (provider: OASTProvider) => {
  editingId.value = provider.id;
  form.value = {
    name: provider.name,
    type: provider.type,
    url: provider.url,
    token: provider.token,
  };
  showForm.value = true;
};

const saveForm = async () => {
  if (!form.value.name || !form.value.type || !form.value.url) return;

  if (editingId.value) {
    await store.updateProvider(editingId.value, form.value);
  } else {
    await store.createProvider({
      ...form.value,
      enabled: true,
    });
  }

  showForm.value = false;
};

const deleteProvider = async (id: string) => {
  await store.deleteProvider(id);
};

const toggleProvider = async (id: string, enabled: boolean) => {
  await store.toggleProvider(id, enabled);
};
</script>

<template>
  <div class="modal-overlay" @click.self="emit('close')">
    <div class="modal-content">
      <!-- Header -->
      <div class="modal-header">
        <h3>{{ t('modules.plugins.oast.provider_manager', 'OAST Provider Manager') }}</h3>
        <button class="btn-icon" @click="emit('close')">
          <i class="bx bx-x text-xl"></i>
        </button>
      </div>

      <!-- Provider 列表 -->
      <div class="provider-list">
        <table class="provider-table">
          <thead>
            <tr>
              <th>{{ t('modules.plugins.oast.name', 'Name') }}</th>
              <th>{{ t('modules.plugins.oast.type', 'Type') }}</th>
              <th>URL</th>
              <th class="w-20">{{ t('modules.plugins.oast.enabled', 'Enabled') }}</th>
              <th class="w-24">{{ t('modules.plugins.oast.actions', 'Actions') }}</th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="provider in store.providers" :key="provider.id">
              <td>{{ provider.name }}</td>
              <td>
                <span class="type-badge">{{ provider.type }}</span>
                <span v-if="provider.builtin" class="builtin-badge">{{ t('modules.plugins.oast.builtin', 'Built-in') }}</span>
              </td>
              <td class="truncate max-w-48">{{ provider.url }}</td>
              <td>
                <label class="toggle-switch">
                  <input
                    type="checkbox"
                    :checked="provider.enabled"
                    @change="toggleProvider(provider.id, !provider.enabled)"
                  />
                  <span class="toggle-slider"></span>
                </label>
              </td>
              <td>
                <div class="action-btns">
                  <button class="btn-icon-sm" @click="openEditForm(provider)" :title="t('modules.plugins.oast.edit', 'Edit')">
                    <i class="bx bx-edit"></i>
                  </button>
                  <button v-if="!provider.builtin" class="btn-icon-sm text-red-500" @click="deleteProvider(provider.id)" :title="t('modules.plugins.oast.delete', 'Delete')">
                    <i class="bx bx-trash"></i>
                  </button>
                </div>
              </td>
            </tr>
            <tr v-if="store.providers.length === 0">
              <td colspan="5" class="text-center text-gray-400 py-6">
                {{ t('modules.plugins.oast.no_providers', 'No providers configured.') }}
              </td>
            </tr>
          </tbody>
        </table>
      </div>

      <!-- 添加/编辑按钮 -->
      <div class="modal-footer">
        <button class="btn btn-primary btn-sm" @click="openAddForm">
          <i class="bx bx-plus mr-1"></i>
          {{ t('modules.plugins.oast.add_provider', 'Add Provider') }}
        </button>
      </div>

      <!-- 编辑表单弹窗 -->
      <div v-if="showForm" class="form-overlay" @click.self="showForm = false">
        <div class="form-content">
          <h4>{{ editingId ? t('modules.plugins.oast.edit_provider', 'Edit Provider') : t('modules.plugins.oast.add_provider', 'Add Provider') }}</h4>

          <div class="form-group">
            <label>{{ t('modules.plugins.oast.name', 'Name') }}</label>
            <input v-model="form.name" type="text" :placeholder="t('modules.plugins.oast.name_placeholder', 'e.g., My Interactsh')" />
          </div>

          <div class="form-group">
            <label>{{ t('modules.plugins.oast.type', 'Type') }}</label>
            <select v-model="form.type">
              <option value="interactsh">Interactsh</option>
              <option value="boast">BOAST</option>
              <option value="webhooksite">Webhook.site</option>
              <option value="postbin">PostBin</option>
              <option value="digpm">dig.pm</option>
              <option value="dnslogcn">dnslog.cn</option>
            </select>
          </div>

          <div class="form-group">
            <label>URL</label>
            <input v-model="form.url" type="text" placeholder="https://oast.pro" />
          </div>

          <div class="form-group">
            <label>Token ({{ t('modules.plugins.oast.optional', 'Optional') }})</label>
            <input v-model="form.token" type="text" :placeholder="t('modules.plugins.oast.token_placeholder', 'API key or auth token')" />
          </div>

          <div class="form-actions">
            <button class="btn btn-secondary btn-sm" @click="showForm = false">
              {{ t('modules.plugins.oast.cancel', 'Cancel') }}
            </button>
            <button class="btn btn-primary btn-sm" @click="saveForm" :disabled="!form.name || !form.url">
              {{ t('modules.plugins.oast.save', 'Save') }}
            </button>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
.modal-overlay {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background: rgba(0, 0, 0, 0.4);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 100;
}

.modal-content {
  background: var(--color-bg-primary, white);
  border-radius: 12px;
  width: 680px;
  max-height: 80vh;
  display: flex;
  flex-direction: column;
  box-shadow: 0 20px 60px rgba(0, 0, 0, 0.15);
}

.modal-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 16px 20px;
  border-bottom: 1px solid var(--glass-border-light, #e5e7eb);
}

.modal-header h3 {
  font-size: 16px;
  font-weight: 600;
  margin: 0;
}

.provider-list {
  flex: 1;
  overflow: auto;
  padding: 0;
}

.provider-table {
  width: 100%;
  border-collapse: collapse;
  font-size: 13px;
}

.provider-table th {
  padding: 8px 16px;
  text-align: left;
  font-weight: 600;
  color: var(--color-text-secondary);
  background: var(--glass-bg-secondary, #f9fafb);
  border-bottom: 1px solid var(--glass-border-light, #e5e7eb);
}

.provider-table td {
  padding: 8px 16px;
  border-bottom: 1px solid var(--glass-border-light, #f3f4f6);
}

.type-badge {
  display: inline-block;
  padding: 2px 8px;
  border-radius: 4px;
  font-size: 11px;
  font-weight: 600;
  background: #e0e7ff;
  color: #4338ca;
}

.dark .type-badge {
  background: rgba(99, 102, 241, 0.2);
  color: #a5b4fc;
}

.builtin-badge {
  display: inline-block;
  padding: 2px 6px;
  border-radius: 4px;
  font-size: 10px;
  font-weight: 500;
  background: #dcfce7;
  color: #166534;
  margin-left: 4px;
}

.dark .builtin-badge {
  background: rgba(34, 197, 94, 0.2);
  color: #86efac;
}

.toggle-switch {
  position: relative;
  display: inline-block;
  width: 36px;
  height: 20px;
}

.toggle-switch input {
  opacity: 0;
  width: 0;
  height: 0;
}

.toggle-slider {
  position: absolute;
  cursor: pointer;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background-color: #ccc;
  transition: 0.2s;
  border-radius: 20px;
}

.toggle-slider:before {
  position: absolute;
  content: "";
  height: 16px;
  width: 16px;
  left: 2px;
  bottom: 2px;
  background-color: white;
  transition: 0.2s;
  border-radius: 50%;
}

input:checked + .toggle-slider {
  background-color: #3b82f6;
}

input:checked + .toggle-slider:before {
  transform: translateX(16px);
}

.action-btns {
  display: flex;
  gap: 4px;
}

.btn-icon-sm {
  background: none;
  border: none;
  cursor: pointer;
  color: var(--color-text-secondary);
  padding: 4px;
  font-size: 16px;
  border-radius: 4px;
  transition: all 0.15s;
}

.btn-icon-sm:hover {
  background: var(--glass-bg-secondary, rgba(0,0,0,0.05));
}

.modal-footer {
  padding: 12px 20px;
  border-top: 1px solid var(--glass-border-light, #e5e7eb);
  display: flex;
  justify-content: flex-end;
}

.form-overlay {
  position: absolute;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background: rgba(0, 0, 0, 0.3);
  display: flex;
  align-items: center;
  justify-content: center;
  border-radius: 12px;
}

.form-content {
  background: var(--color-bg-primary, white);
  padding: 20px;
  border-radius: 10px;
  width: 400px;
  box-shadow: 0 10px 40px rgba(0, 0, 0, 0.15);
}

.form-content h4 {
  margin: 0 0 16px 0;
  font-size: 15px;
  font-weight: 600;
}

.form-group {
  margin-bottom: 12px;
}

.form-group label {
  display: block;
  font-size: 12px;
  font-weight: 500;
  color: var(--color-text-secondary);
  margin-bottom: 4px;
}

.form-group input,
.form-group select {
  width: 100%;
  padding: 6px 10px;
  border: 1px solid var(--glass-border-light, #d1d5db);
  border-radius: 6px;
  font-size: 13px;
  background: var(--color-bg-primary);
  color: var(--color-text-primary);
}

.form-actions {
  display: flex;
  justify-content: flex-end;
  gap: 8px;
  margin-top: 16px;
}

.btn-icon {
  background: none;
  border: none;
  cursor: pointer;
  color: var(--color-text-secondary);
  padding: 2px;
}
</style>
