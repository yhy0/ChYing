<script setup lang="ts">
import { useI18n } from 'vue-i18n';
import type { ProjectDetails } from '../../types';

const { t } = useI18n();

defineProps<{
  selectedProject: ProjectDetails;
}>();

defineEmits<{
  'open-export-modal': [];
}>();

</script>

<template>
  <div class="project-header">
    <div class="flex justify-between items-center mb-2">
      <div>
        <h1 class="project-title">
          {{ selectedProject.name }}
        </h1>
        <p class="project-date">
          <i class="bx bx-calendar mr-1"></i>
          {{ t('modules.project.created_on', { date: selectedProject.createdDate }) }}
        </p>
      </div>
    </div>
    
    <div class="grid grid-cols-4 gap-4">
      <div class="project-stat-card">
        <div class="flex items-center justify-between mb-1.5">
          <span class="project-stat-label">{{ t('modules.project.stats.requests') }}</span>
          <i class="bx bx-globe text-base project-icon project-icon-primary"></i>
        </div>
        <div class="project-stat-value">
          {{ selectedProject.totalRequests }}
        </div>
      </div>
      
      <div class="project-stat-card">
        <div class="flex items-center justify-between mb-1.5">
          <span class="project-stat-label">{{ t('modules.project.stats.issues') }}</span>
          <i class="bx bx-bug text-base project-icon project-icon-error"></i>
        </div>
        <div class="project-stat-value">
          {{ selectedProject.issuesFound }}
        </div>
      </div>
      
      <div class="project-stat-card">
        <div class="flex items-center justify-between mb-1.5">
          <span class="project-stat-label">{{ t('modules.project.stats.hosts') }}</span>
          <i class="bx bx-server text-base project-icon project-icon-success"></i>
        </div>
        <div class="project-stat-value">
          {{ selectedProject.hosts }}
        </div>
      </div>
      
      <div class="project-stat-card">
        <div class="flex items-center justify-between mb-1.5">
          <span class="project-stat-label">{{ t('modules.project.stats.scan_progress') }}</span>
          <i class="bx bx-loader-alt text-base project-icon project-icon-primary"></i>
        </div>
        <div class="flex items-center space-x-2">
          <div class="flex-1 progress-bar">
            <div class="progress-bar-fill" :style="{ width: `${selectedProject.scanProgress}%` }"></div>
          </div>
          <span class="text-sm font-medium text-gray-700 dark:text-gray-300">{{ selectedProject.scanProgress }}%</span>
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
.hover\:shadow-md:hover {
  box-shadow: 0 6px 12px -2px rgba(0, 0, 0, 0.08), 0 3px 6px -2px rgba(0, 0, 0, 0.05);
}

.export-btn {
  position: relative;
  overflow: hidden;
}

.export-btn::after {
  content: '';
  position: absolute;
  top: 50%;
  left: 50%;
  width: 120%;
  height: 0;
  padding-bottom: 120%;
  transform: translate(-50%, -50%) scale(0);
  background: rgba(255, 255, 255, 0.1);
  border-radius: 50%;
  opacity: 0;
  transition: transform 0.4s ease, opacity 0.3s ease;
}

.export-btn:active::after {
  transform: translate(-50%, -50%) scale(1);
  opacity: 1;
  transition: 0s;
}

.export-btn:hover .animate-icon {
  animation: bounce 0.6s ease;
}

@keyframes bounce {
  0%, 100% {
    transform: translateX(0);
  }
  40% {
    transform: translateX(-3px);
  }
  60% {
    transform: translateX(3px);
  }
  80% {
    transform: translateX(-2px);
  }
}
</style> 