<script setup lang="ts">
import { computed, onMounted, onBeforeUnmount } from 'vue';
import DecoderTabs from './DecoderTabs.vue';
import DecoderTabPanel from './DecoderTabPanel.vue';
import { useModulesStore } from '../../store';
import type { DecoderTab } from '../../types';
import { useI18n } from 'vue-i18n';

// Store
const store = useModulesStore();

// Tabs reactive reference
const tabs = computed(() => store.decoderTabs as DecoderTab[]);
const activeTab = computed(() => tabs.value.find(tab => tab.isActive) || null);

// Key handler for keyboard shortcuts
function handleKeyDown(e: KeyboardEvent) {
  // Global keyboard shortcuts for Decoder
  if ((e.ctrlKey || e.metaKey) && e.key === 'n') { // Ctrl+N or Cmd+N
    e.preventDefault();
    addNewTab();
  }
}

// Add a new tab
function addNewTab() {
  store.createDecoderTab();
}

// Ensure at least one tab exists on mount
onMounted(() => {
  // Register key handler
  window.addEventListener('keydown', handleKeyDown);
  
  // If no tabs exist, create one
  if (tabs.value.length === 0) {
    addNewTab();
  }
});

onBeforeUnmount(() => {
  // Clean up key handler
  window.removeEventListener('keydown', handleKeyDown);
});

// I18n
const { t } = useI18n();
</script>

<template>
  <div class="h-full flex flex-col">
    <!-- Control Bar -->
    <div class="decoder-control-bar">
      <div class="flex items-center space-x-4">
        <button
          class="btn btn-primary"
          @click="addNewTab"
          title="New Tab (Ctrl+N)"
        >
          <i class="bx bx-plus mr-1"></i> {{ t('modules.decoder.new_tab') }}
        </button>
      </div>
    </div>
    
    <!-- Tabs -->
    <div class="flex-none overflow-visible">
      <DecoderTabs />
    </div>
    
    <!-- Tab Content -->
    <div v-if="activeTab" class="flex-1 overflow-hidden">
      <DecoderTabPanel 
        v-if="activeTab" 
        :tab="activeTab" 
      />
    </div>
    
    <!-- Empty State -->
    <div v-else class="empty-state">
      <i class="bx bx-code-alt empty-state-icon"></i>
      <h3 class="empty-state-text">{{ t('modules.decoder.no_tab_open') }}</h3>
      <p class="empty-state-text">{{ t('modules.decoder.create_new_tab') }}</p>
      <button 
        class="btn btn-primary mt-4"
        @click="addNewTab"
      >
        <i class="bx bx-plus mr-1"></i>
        {{ t('modules.decoder.new_tab') }}
      </button>
    </div>
  </div>
</template>

<style scoped>
.animate-fade-in-out {
  animation: fadeInOut 2s ease-in-out;
}

@keyframes fadeInOut {
  0% { opacity: 0; }
  20% { opacity: 1; }
  80% { opacity: 1; }
  100% { opacity: 0; }
}
</style> 