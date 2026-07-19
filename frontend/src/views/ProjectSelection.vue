<script setup lang="ts">
import { ref, onMounted, onUnmounted, computed } from 'vue';
import { useRouter } from 'vue-router';
import { useI18n } from 'vue-i18n';
import { Events } from '@wailsio/runtime';
// @ts-ignore
import {
  GetProject,
  GetDesktopRuntimeState,
  OpenExistingProject,
  StartInitialization,
  StepBasicInitialization,
  StepConfigurationLoad,
  StepDatabaseConnection,
  StepSchemaValidation,
  StepProxyServerStart,
  StepProjectDataLoad,
  StepInitializationComplete,
  GetLocalProjects,
  CreateLocalProject,
  DeleteLocalProject,
} from "../../bindings/github.com/yhy0/ChYing/app.js";

import { useProjectStore } from '../store/project';
import type { ProjectInfo } from '../types/project';

const { t } = useI18n();
const router = useRouter();
const projectStore = useProjectStore();

// 状态管理
const notificationMessage = ref('');
const notificationType = ref<'success' | 'error'>('error');
const showNotification = ref(false);
const isLoading = ref(false);

// 项目状态
const projectAction = ref<'open' | 'new' | 'temp' | null>(null);
const projectName = ref('');
const selectedProject = ref<ProjectInfo | null>(null);
const localProjects = ref<ProjectInfo[]>([]);

// 加载进度状态
const loadingProgress = ref(0);
const loadingStep = ref('');

// 删除确认对话框状态
const showDeleteConfirm = ref(false);
const projectToDelete = ref<ProjectInfo | null>(null);
const isDeleting = ref(false);

// 计算属性
const canProceed = computed(() => {
  if (isLoading.value) return false;
  if (!projectAction.value) return false;
  
  if (projectAction.value === 'temp') return true;
  if (projectAction.value === 'new') return projectName.value.trim() !== '';
  if (projectAction.value === 'open') return selectedProject.value !== null;
  
  return false;
});

// 显示通知
const showMessage = (message: string, type: 'success' | 'error' = 'error') => {
  notificationMessage.value = message;
  notificationType.value = type;
  showNotification.value = true;
  setTimeout(() => {
    showNotification.value = false;
  }, 3000);
};

// 加载本地项目
const loadLocalProjects = async () => {
  try {
    const response = await GetLocalProjects();
    if (response.error) throw new Error(response.error);

    localProjects.value = (response.data?.projects || []).map((p: any) => ({
      ...p,
      source: 'local'
    }));

    // 如果有本地项目，自动选择打开模式
    if (localProjects.value.length > 0) {
      projectAction.value = 'open';
    } else {
      projectAction.value = 'new';
    }
  } catch (err) {
    console.error('加载本地项目失败:', err);
    showMessage(t('pages.project.load_failed') + ': ' + (err as Error).message, 'error');
  }
};

// 打开删除确认对话框
const openDeleteConfirm = (project: ProjectInfo, event: Event) => {
  event.stopPropagation(); // 阻止事件冒泡，避免选中项目
  projectToDelete.value = project;
  showDeleteConfirm.value = true;
};

// 关闭删除确认对话框
const closeDeleteConfirm = () => {
  showDeleteConfirm.value = false;
  projectToDelete.value = null;
};

// 确认删除项目
const confirmDeleteProject = async () => {
  if (!projectToDelete.value || isDeleting.value) return;

  try {
    isDeleting.value = true;
    const response = await DeleteLocalProject(projectToDelete.value.name);

    if (response.error) {
      throw new Error(response.error);
    }

    // 如果删除的是当前选中的项目，清除选中状态
    if (selectedProject.value?.id === projectToDelete.value.id) {
      selectedProject.value = null;
    }

    showMessage(t('pages.project.delete_success'), 'success');
    closeDeleteConfirm();

    // 重新加载项目列表
    await loadLocalProjects();
  } catch (err) {
    console.error('删除项目失败:', err);
    showMessage(t('pages.project.delete_failed') + ': ' + (err as Error).message, 'error');
  } finally {
    isDeleting.value = false;
  }
};

// 创建新项目
const createNewProject = async (name: string) => {
  try {
    const projectId = name.toLowerCase().replace(/[^a-z0-9]/g, '_') + `_${Date.now()}`;
    await CreateLocalProject(projectId, name);
    return { projectId, projectName: name };
  } catch (error) {
    console.error('创建项目失败:', error);
    throw error;
  }
};

// 创建临时项目
const createTempProject = async () => {
  try {
    const now = new Date();
    const timestamp = now.getFullYear() + '-' + 
                     String(now.getMonth() + 1).padStart(2, '0') + '-' + 
                     String(now.getDate()).padStart(2, '0') + '-' + 
                     String(now.getHours()).padStart(2, '0') + '-' + 
                     String(now.getMinutes()).padStart(2, '0');
    
    const tempDbName = `temp-${timestamp}`;
    const tempProjectName = `temp-${timestamp}`;

    await CreateLocalProject(tempDbName, tempProjectName);
    
    return { projectId: tempDbName, projectName: tempProjectName };
  } catch (error) {
    console.error('创建临时项目失败:', error);
    throw error;
  }
};

// 更新进度辅助函数
const updateProgressStep = (progress: number, step: string) => {
  loadingProgress.value = progress;
  loadingStep.value = step;
};

// 辅助函数：调用初始化步骤
const callInitStep = async (stepFunction: Function, args: any[], description: string, progress: number) => {
  updateProgressStep(progress, description);
  const result = await stepFunction(...args);
  if (result && result.error) {
    throw new Error(result.error);
  }
  await new Promise(resolve => setTimeout(resolve, 200));
};

const handleDesktopProjectProgress = (event: any) => {
  const state = event?.data;
  if (!state) return;

  if (state.status === 'opening') {
    isLoading.value = true;
    updateProgressStep(state.progress || 0, state.message || t('pages.project.init_preparing'));
  } else if (state.status === 'failed') {
    isLoading.value = false;
    loadingProgress.value = 0;
    loadingStep.value = '';
    showMessage(state.error || t('pages.project.operation_failed'), 'error');
  } else if (state.status === 'ready') {
    updateProgressStep(100, t('pages.project.init_done'));
  }
};

// 处理下一步
async function handleNext() {
  if (!canProceed.value || isLoading.value) return;
  
  projectStore.clearSiteMapData();
  
  let finalProjectType = '';
  let finalProjectName = '';
  
  try {
    isLoading.value = true;
    updateProgressStep(0, t('pages.project.init_preparing'));

    switch (projectAction.value) {
      case 'temp':
        const tempProject = await createTempProject();
        finalProjectType = 'Temporary project';
        finalProjectName = tempProject.projectName;
        break;
      case 'new':
        finalProjectType = 'New project';
        finalProjectName = projectName.value.trim();
        await createNewProject(finalProjectName);
        break;
      case 'open':
        finalProjectType = 'Open existing project';
        finalProjectName = selectedProject.value!.name;
        break;
    }

    if (projectAction.value === 'open') {
      const result = await OpenExistingProject(finalProjectName);
      if (result?.error) {
        throw new Error(result.error);
      }
      await GetProject();
      updateProgressStep(100, t('pages.project.init_done'));
      isLoading.value = false;
      await router.push('/app/project');
      return;
    }

    await callInitStep(StartInitialization, [finalProjectType, finalProjectName], t('pages.project.init_starting'), 15);
    await callInitStep(StepBasicInitialization, [], t('pages.project.init_base_components'), 25);
    await callInitStep(StepConfigurationLoad, [], t('pages.project.init_loading_config'), 35);
    await callInitStep(StepDatabaseConnection, [finalProjectName], t('pages.project.init_connecting_db'), 50);
    await callInitStep(StepSchemaValidation, [], t('pages.project.init_verifying_schema'), 65);
    await callInitStep(StepProxyServerStart, [], t('pages.project.init_starting_proxy'), 80);
    await callInitStep(StepProjectDataLoad, [finalProjectType, finalProjectName], t('pages.project.init_loading_data'), 90);
    await callInitStep(StepInitializationComplete, [], t('pages.project.init_completing'), 100);
    
    await GetProject();

    setTimeout(() => {
      updateProgressStep(100, t('pages.project.init_done'));
      setTimeout(() => {
        isLoading.value = false;
        router.push('/app/project');
      }, 500);
    }, 500);

  } catch (error) {
    isLoading.value = false;
    loadingProgress.value = 0;
    loadingStep.value = '';
    console.error('项目加载失败:', error);
    showMessage(`${t('pages.project.operation_failed')}: ${error}`, 'error');
  }
};

// 组件加载时初始化
onMounted(async () => {
  Events.On('DesktopProjectOpenProgress', handleDesktopProjectProgress);
  await loadLocalProjects();

  try {
    const response = await GetDesktopRuntimeState();
    const state: any = response?.data;
    if (state?.status === 'opening') {
      isLoading.value = true;
      updateProgressStep(state.progress || 0, state.message || t('pages.project.init_preparing'));
    } else if (state?.status === 'ready') {
      await router.push('/app/project');
    } else if (state?.status === 'failed') {
      showMessage(state.error || t('pages.project.operation_failed'), 'error');
    }
  } catch (error) {
    console.error('读取桌面运行状态失败:', error);
  }
});

onUnmounted(() => {
  Events.Off('DesktopProjectOpenProgress');
  isLoading.value = false;
});
</script>

<template>
  <!-- 加载遮罩 -->
  <Transition name="fade">
    <div v-if="isLoading" class="fixed inset-0 bg-black/50 z-50 flex items-center justify-center backdrop-blur-sm">
      <div class="loading-container">
        <!-- 主标题 -->
        <div class="loading-title">
          <h3 class="text-xl font-semibold text-white mb-2">{{ t('pages.project.app_starting') }}</h3>
          <p class="text-gray-300 text-sm mb-6">{{ t('pages.project.app_starting_desc') }}</p>
        </div>
        
        <!-- 进度条容器 -->
        <div class="progress-container">
          <div class="progress-bar-bg">
            <div 
              class="progress-bar-fill" 
              :style="{ width: `${loadingProgress}%` }"
            ></div>
            <div class="progress-glow" :style="{ left: `${loadingProgress}%` }"></div>
          </div>
          
          <!-- 进度百分比 -->
          <div class="progress-info">
            <span class="progress-percentage">{{ loadingProgress }}%</span>
          </div>
        </div>
        
        <!-- 当前步骤 -->
        <div class="step-container">
          <div class="step-icon">
            <i class="bx bx-loader-alt animate-spin text-primary"></i>
          </div>
          <span class="step-text">{{ loadingStep }}</span>
        </div>
        
        <!-- 装饰元素 -->
        <div class="loading-decoration">
          <div class="decoration-dot"></div>
          <div class="decoration-dot"></div>
          <div class="decoration-dot"></div>
        </div>
      </div>
    </div>
  </Transition>
  
  <!-- 错误提示 -->
  <Transition name="slide-down">
    <div v-if="showNotification" class="fixed top-4 left-1/2 transform -translate-x-1/2 z-50">
      <div class="notification-badge px-4 py-2 rounded-lg shadow-lg" :class="notificationType === 'success' ? 'bg-green-500 text-white' : 'bg-red-500 text-white'">
        {{ notificationMessage }}
      </div>
    </div>
  </Transition>

  <!-- 删除确认对话框 -->
  <Transition name="fade">
    <div v-if="showDeleteConfirm" class="dialog-overlay">
      <div class="delete-confirm-dialog">
        <div class="dialog-header">
          <div class="dialog-icon warning">
            <i class="fas fa-exclamation-triangle"></i>
          </div>
          <h3>{{ t('pages.project.delete_confirm_title') }}</h3>
        </div>
        <div class="dialog-content">
          <p>{{ t('pages.project.delete_confirm_message', { name: projectToDelete?.name }) }}</p>
          <p class="warning-text">{{ t('pages.project.delete_warning') }}</p>
        </div>
        <div class="dialog-actions">
          <button class="btn btn-secondary btn-sm cancel-btn" type="button" @click="closeDeleteConfirm" :disabled="isDeleting">
            {{ t('common.actions.cancel') }}
          </button>
          <button class="btn btn-danger btn-sm confirm-btn danger" type="button" @click="confirmDeleteProject" :disabled="isDeleting">
            <i v-if="isDeleting" class="fas fa-spinner fa-spin"></i>
            <span v-else>{{ t('common.actions.delete') }}</span>
          </button>
        </div>
      </div>
    </div>
  </Transition>
  
  <!-- 主容器 -->
  <div class="project-selection-container">
    <!-- 背景装饰 - 承影剑光 (底层) -->
    <div class="background-decoration">
      <!-- 剑形光效 - 承影剑 -->
      <div class="sword-blade sword-blade-1"></div>
      <div class="sword-blade sword-blade-2"></div>
      <div class="sword-blade sword-blade-3"></div>
      <div class="sword-blade sword-blade-4"></div>
      <!-- 古剑形态 -->
      <div class="sword-ancient sword-ancient-1"></div>
      <div class="sword-ancient sword-ancient-2"></div>
      <div class="sword-ancient sword-ancient-3"></div>
      <!-- 飞剑形态 -->
      <div class="sword-flying sword-flying-1"></div>
      <div class="sword-flying sword-flying-2"></div>
      <div class="sword-flying sword-flying-3"></div>
      <div class="sword-flying sword-flying-4"></div>
      <div class="sword-flying sword-flying-5"></div>
      <!-- 垂直剑光 (保留部分作为辅助效果) -->
      <div class="sword-light sword-light-1"></div>
      <div class="sword-light sword-light-2"></div>
      <div class="sword-light sword-light-3"></div>
      <!-- 斜向剑光 -->
      <div class="sword-diagonal sword-diagonal-1"></div>
      <div class="sword-diagonal sword-diagonal-2"></div>
      <!-- 剑气光晕 -->
      <div class="sword-glow sword-glow-1"></div>
      <div class="sword-glow sword-glow-2"></div>
      <div class="sword-glow sword-glow-3"></div>
      <div class="sword-glow sword-glow-4"></div>
      <!-- 剑锋闪烁 -->
      <div class="sword-tip-flash sword-tip-flash-1"></div>
      <div class="sword-tip-flash sword-tip-flash-2"></div>
      <div class="sword-tip-flash sword-tip-flash-3"></div>
    </div>

    <!-- 主内容区域 -->
    <div class="main-content-wrapper">
      <!-- 紧凑的页面标题 -->
      <div class="compact-header">
        <div class="brand-container">
          <h1 class="brand-title">{{ t('layout.app.title') }}</h1>
        </div>

        <!-- 引言卡片 -->
        <div class="quote-card-compact">
          <div class="quote-icon">
            <i class="fas fa-quote-left"></i>
          </div>
          <p class="quote-text">{{ t('pages.project.literary_quote') }}</p>
          <div class="quote-author">{{ t('pages.project.literary_source') }}</div>
        </div>
      </div>

      <!-- 项目操作区域 - 左右布局 -->
      <div class="layout-container">
        <!-- 左侧面板：项目操作 -->
        <div class="left-panel">
          <div class="panel-header">
            <h3>{{ t('pages.project.project_actions') }}</h3>
            <p>{{ t('pages.project.select_action') }}</p>
          </div>

          <div class="action-selector-compact">
            <button
              type="button"
              class="action-option-compact"
              @click="projectAction = 'open'"
              :class="{ active: projectAction === 'open' }"
              :aria-pressed="projectAction === 'open'"
            >
              <div class="action-sword-icon">
                <svg viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg">
                  <path d="M4 20L7 17M20 4L11 13M11 13L8 16L4 20M11 13L14 10" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round"/>
                  <path d="M15 5L19 9" stroke="currentColor" stroke-width="1.5" stroke-linecap="round"/>
                </svg>
              </div>
              <div class="action-info-compact">
                <h4>{{ t('pages.project.open_existing') }}</h4>
                <p>{{ t('pages.project.open_existing_desc') }}</p>
              </div>
              <div class="action-radio">
                <div class="radio-button" :class="{ checked: projectAction === 'open' }" aria-hidden="true"></div>
              </div>
            </button>

            <button
              type="button"
              class="action-option-compact"
              @click="projectAction = 'new'"
              :class="{ active: projectAction === 'new' }"
              :aria-pressed="projectAction === 'new'"
            >
              <div class="action-sword-icon">
                <svg viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg">
                  <path d="M12 3V21M3 12H21" stroke="currentColor" stroke-width="1.5" stroke-linecap="round"/>
                  <path d="M7 7L17 17M17 7L7 17" stroke="currentColor" stroke-width="1" stroke-linecap="round" opacity="0.3"/>
                </svg>
              </div>
              <div class="action-info-compact">
                <h4>{{ t('pages.project.create_new') }}</h4>
                <p>{{ t('pages.project.create_new_desc') }}</p>
              </div>
              <div class="action-radio">
                <div class="radio-button" :class="{ checked: projectAction === 'new' }" aria-hidden="true"></div>
              </div>
            </button>

            <button
              type="button"
              class="action-option-compact"
              @click="projectAction = 'temp'"
              :class="{ active: projectAction === 'temp' }"
              :aria-pressed="projectAction === 'temp'"
            >
              <div class="action-sword-icon">
                <svg viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg">
                  <path d="M13 3L4 14H12L11 21L20 10H12L13 3Z" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round"/>
                </svg>
              </div>
              <div class="action-info-compact">
                <h4>{{ t('pages.project.start_temp') }}</h4>
                <p>{{ t('pages.project.start_temp_desc') }}</p>
              </div>
              <div class="action-radio">
                <div class="radio-button" :class="{ checked: projectAction === 'temp' }" aria-hidden="true"></div>
              </div>
            </button>
          </div>
        </div>

        <!-- 右侧面板：操作详情 -->
        <div class="right-panel">
          <Transition name="fade-slide" mode="out-in">
            <!-- 打开现有项目 -->
            <div v-if="projectAction === 'open'" key="open" class="project-list-compact">
              <div class="details-header">
                <h3>{{ t('pages.project.select_project') }}</h3>
                <div class="project-count">{{ t('pages.project.project_count', { count: localProjects.length }) }}</div>
              </div>

              <div
                class="project-list-ultra-compact"
                v-if="localProjects.length > 0"
                role="listbox"
                :aria-label="t('pages.project.select_project')"
              >
                <div v-for="project in localProjects" :key="project.id"
                     class="project-item-compact"
                     :class="{ selected: selectedProject?.id === project.id }"
                     role="option"
                     tabindex="0"
                     :aria-selected="selectedProject?.id === project.id"
                     @click="selectedProject = project"
                     @keydown.enter="selectedProject = project"
                     @keydown.space.prevent="selectedProject = project">
                  <div class="project-sword-icon">
                    <svg viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg">
                      <path d="M4 20L7 17M20 4L11 13M11 13L8 16L4 20M11 13L14 10" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round"/>
                      <path d="M15 5L19 9" stroke="currentColor" stroke-width="1.5" stroke-linecap="round"/>
                    </svg>
                  </div>
                  <div class="project-item-info">
                    <div class="project-item-name">{{ project.name }}</div>
                    <div class="project-item-meta">
                      <span class="project-source local">{{ t('pages.project.local_tag') }}</span>
                      <span class="project-requests">{{ t('pages.project.request_count', { count: project.requests?.toLocaleString() || 0 }) }}</span>
                      <span v-if="project.size_formatted" class="project-size">{{ project.size_formatted }}</span>
                    </div>
                  </div>
                  <div class="project-item-actions">
                    <button
                      class="delete-btn"
                      type="button"
                      @click="openDeleteConfirm(project, $event)"
                      :title="t('common.actions.delete')"
                      :aria-label="t('common.actions.delete')"
                    >
                      <i class="bx bx-trash"></i>
                    </button>
                  </div>
                  <div class="project-item-selector">
                    <div class="radio-button" :class="{ checked: selectedProject?.id === project.id }" aria-hidden="true"></div>
                  </div>
                </div>
              </div>

              <div v-else class="empty-state-compact">
                <i class="fas fa-folder-open"></i>
                <h4>{{ t('pages.project.no_projects_found') }}</h4>
                <p>{{ t('pages.project.create_first') }}</p>
              </div>
            </div>

            <!-- 创建新项目 -->
            <div v-else-if="projectAction === 'new'" key="new" class="new-project-compact">
              <div class="details-header">
                <h3>{{ t('pages.project.create_new_project') }}</h3>
              </div>
              <div class="input-card-compact">
                <div class="input-header">
                  <i class="fas fa-edit"></i>
                  <span>{{ t('pages.project.project_name_label') }}</span>
                </div>
                <input v-model="projectName"
                       type="text"
                       class="project-name-input-compact"
                       :placeholder="t('pages.project.project_name_input_placeholder')"
                       @keyup.enter="handleNext" spellcheck="false">
                <div class="input-hint">
                  {{ t('pages.project.project_name_hint') }}
                </div>
              </div>
            </div>

            <!-- 临时项目 -->
            <div v-else-if="projectAction === 'temp'" key="temp" class="temp-project-compact">
              <div class="details-header">
                <h3>{{ t('pages.project.temp_project_info') }}</h3>
              </div>
              <div class="temp-info-card-compact">
                <div class="temp-icon">
                  <i class="fas fa-info-circle"></i>
                </div>
                <div class="temp-content">
                  <p>{{ t('pages.project.temp_project_desc') }}</p>
                  <div class="temp-features">
                    <div class="feature-item">
                      <i class="fas fa-bolt"></i>
                      <span>{{ t('pages.project.quick_start') }}</span>
                    </div>
                    <div class="feature-item">
                      <i class="fas fa-save"></i>
                      <span>{{ t('pages.project.auto_save') }}</span>
                    </div>
                  </div>
                </div>
              </div>
            </div>

            <!-- 未选择操作时的提示 -->
            <div v-else key="empty" class="empty-state-compact">
              <i class="fas fa-hand-pointer"></i>
              <h4>{{ t('pages.project.select_action_prompt') }}</h4>
              <p>{{ t('pages.project.select_action_hint') }}</p>
            </div>
          </Transition>
        </div>
      </div>

      <!-- 底部操作区域 -->
      <div class="bottom-actions-compact">
        <button @click="handleNext"
                :disabled="!canProceed"
                class="next-button-compact"
                :class="{ 'button-ready': canProceed }">
          <span class="button-text">{{ t('common.actions.next') }}</span>
          <i class="fas fa-arrow-right button-icon"></i>
        </button>
      </div>
    </div>
  </div>
</template>

<style scoped>
/* 所有样式已移至 frontend/src/styles/modules/project-selection.css */
</style>
