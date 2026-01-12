<script setup lang="ts">
import BaseTabs from '../common/BaseTabs.vue';
import type { IntruderTab, IntruderGroup } from '../../types';

defineProps<{
  tabs: IntruderTab[];
  groups: IntruderGroup[];
}>();

const emit = defineEmits<{
  (e: 'select-tab', tabId: string): void;
  (e: 'close-tab', tabId: string): void;
  (e: 'rename-tab', tabId: string, newName: string): void;
  (e: 'change-tab-color', tabId: string, color: string): void;
  (e: 'change-tab-group', tabId: string, groupId: string | null): void;
  (e: 'create-group'): void;
  (e: 'reorder-tabs', tabs: IntruderTab[]): void;
  (e: 'reorder-groups', groups: IntruderGroup[]): void;
}>();

// 预定义的颜色选项
const colorOptions = [
  { id: 'default', value: '#4f46e5', label: 'Default (Purple)' },
  { id: 'red', value: '#ef4444', label: 'Red' },
  { id: 'green', value: '#10b981', label: 'Green' },
  { id: 'blue', value: '#3b82f6', label: 'Blue' },
  { id: 'yellow', value: '#f59e0b', label: 'Yellow' },
  { id: 'orange', value: '#f97316', label: 'Orange' },
  { id: 'teal', value: '#14b8a6', label: 'Teal' },
];
</script>

<template>
  <BaseTabs
    :tabs="tabs"
    :groups="groups"
    prefix="intruder"
    :colorOptions="colorOptions"
    @select-tab="tabId => emit('select-tab', tabId)"
    @close-tab="tabId => emit('close-tab', tabId)"
    @rename-tab="(tabId, newName) => emit('rename-tab', tabId, newName)"
    @change-tab-color="(tabId, color) => emit('change-tab-color', tabId, color)"
    @change-tab-group="(tabId, groupId) => emit('change-tab-group', tabId, groupId)"
    @create-group="() => emit('create-group')"
    @reorder-tabs="tabs => emit('reorder-tabs', tabs)"
    @reorder-groups="groups => emit('reorder-groups', groups)"
  >
    <template #tab-extra="{ tab }">
      <span v-if="tab.isRunning" class="running-indicator"></span>
    </template>
    <template #group-collapsed-indicator="{ tabs }">
      <span v-if="tabs.some(tab => tab.isActive)" class="group-active-indicator"></span>
    </template>
  </BaseTabs>
</template>

<style scoped>
</style> 