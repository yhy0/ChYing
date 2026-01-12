<script setup lang="ts">
import { ref, onMounted, onUnmounted, computed } from 'vue';
import { useRouter } from 'vue-router';
import { useI18n } from 'vue-i18n';
// @ts-ignore
import {
  GetProject,
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

    console.log('本地项目加载成功:', localProjects.value);

    // 如果有本地项目，自动选择打开模式
    if (localProjects.value.length > 0) {
      projectAction.value = 'open';
    } else {
      projectAction.value = 'new';
    }
  } catch (err) {
    console.error('加载本地项目失败:', err);
    showMessage('加载本地项目失败: ' + (err as Error).message, 'error');
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
    
    console.log('创建临时项目:', tempProjectName, '数据库文件:', `${tempDbName}.db`);
    
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

// 处理下一步
async function handleNext() {
  if (!canProceed.value || isLoading.value) return;
  
  projectStore.clearSiteMapData();
  
  let finalProjectType = '';
  let finalProjectName = '';
  
  try {
    isLoading.value = true;
    updateProgressStep(0, '准备开始...');

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

    await callInitStep(StartInitialization, [finalProjectType, finalProjectName], "正在启动初始化...", 15);
    await callInitStep(StepBasicInitialization, [], "正在初始化基础组件...", 25);
    await callInitStep(StepConfigurationLoad, [], "正在加载配置文件...", 35);
    await callInitStep(StepDatabaseConnection, [finalProjectName], "正在连接数据库...", 50);
    await callInitStep(StepSchemaValidation, [], "正在验证数据库模式...", 65);
    await callInitStep(StepProxyServerStart, [], "正在启动代理服务器...", 80);
    await callInitStep(StepProjectDataLoad, [finalProjectType, finalProjectName], "正在加载项目数据...", 90);
    await callInitStep(StepInitializationComplete, [], "正在完成初始化...", 100);
    
    const projectResult = await GetProject();
    if (projectResult && Object.keys(projectResult).length > 0) {
      console.log('项目数据加载成功:', projectResult);
    }
    
    setTimeout(() => {
      updateProgressStep(100, '初始化完成，正在跳转...');
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
  loadLocalProjects();
});

onUnmounted(() => {
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
          <h3 class="text-xl font-semibold text-white mb-2">承影 正在启动</h3>
          <p class="text-gray-300 text-sm mb-6">请稍候，正在为您准备安全测试环境...</p>
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
    <div v-if="showDeleteConfirm" class="fixed inset-0 bg-black/50 z-50 flex items-center justify-center backdrop-blur-sm">
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
          <button class="cancel-btn" @click="closeDeleteConfirm" :disabled="isDeleting">
            {{ t('common.actions.cancel') }}
          </button>
          <button class="confirm-btn danger" @click="confirmDeleteProject" :disabled="isDeleting">
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
      <!-- 剑气残影 -->
      <div class="sword-afterimage sword-afterimage-1"></div>
      <div class="sword-afterimage sword-afterimage-2"></div>
      <div class="sword-afterimage sword-afterimage-3"></div>
      <div class="sword-afterimage sword-afterimage-4"></div>
      <div class="sword-afterimage sword-afterimage-5"></div>
      <!-- 垂直剑光 (保留部分作为辅助效果) -->
      <div class="sword-light sword-light-1"></div>
      <div class="sword-light sword-light-2"></div>
      <div class="sword-light sword-light-3"></div>
      <!-- 斜向剑光 -->
      <div class="sword-diagonal sword-diagonal-1"></div>
      <div class="sword-diagonal sword-diagonal-2"></div>
      <!-- 剑尖光点 -->
      <div class="sword-spark sword-spark-1"></div>
      <div class="sword-spark sword-spark-2"></div>
      <div class="sword-spark sword-spark-3"></div>
      <div class="sword-spark sword-spark-4"></div>
      <div class="sword-spark sword-spark-5"></div>
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

    <!-- 背景装饰 - 承影剑光 (顶层 - 穿透内容) -->
    <div class="background-decoration-top">
      <!-- 顶层垂直剑光 -->
      <div class="sword-light-top sword-light-top-1"></div>
      <div class="sword-light-top sword-light-top-2"></div>
      <div class="sword-light-top sword-light-top-3"></div>
      <!-- 顶层斜向剑光 -->
      <div class="sword-diagonal-top sword-diagonal-top-1"></div>
      <div class="sword-diagonal-top sword-diagonal-top-2"></div>
    </div>

    <!-- 主内容区域 -->
    <div class="main-content-wrapper">
      <!-- 紧凑的页面标题 -->
      <div class="compact-header">
        <div class="brand-container">
          <h1 class="brand-title">承影</h1>
        </div>

        <!-- 引言卡片 -->
        <div class="quote-card-compact">
          <div class="quote-icon">
            <i class="fas fa-quote-left"></i>
          </div>
          <p class="quote-text">将旦昧爽之交，日夕昏明之际，北面而察之，淡淡焉若有物存，莫识其状。其所触也，窃窃然有声，经物而物不疾也。</p>
          <div class="quote-author">—— 《列子·汤问》</div>
        </div>
      </div>

      <!-- 项目操作区域 - 左右布局 -->
      <div class="layout-container">
        <!-- 左侧面板：项目操作 -->
        <div class="left-panel">
          <div class="panel-header">
            <h3>项目操作</h3>
            <p>选择要执行的操作</p>
          </div>

          <div class="action-selector-compact">
            <div class="action-option-compact" @click="projectAction = 'open'" :class="{ active: projectAction === 'open' }">
              <div class="action-sword-icon">
                <svg viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg">
                  <path d="M4 20L7 17M20 4L11 13M11 13L8 16L4 20M11 13L14 10" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round"/>
                  <path d="M15 5L19 9" stroke="currentColor" stroke-width="1.5" stroke-linecap="round"/>
                </svg>
              </div>
              <div class="action-info-compact">
                <h4>打开现有项目</h4>
                <p>从已有项目中选择</p>
              </div>
              <div class="action-radio">
                <div class="radio-button" :class="{ checked: projectAction === 'open' }"></div>
              </div>
            </div>

            <div class="action-option-compact" @click="projectAction = 'new'" :class="{ active: projectAction === 'new' }">
              <div class="action-sword-icon">
                <svg viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg">
                  <path d="M12 3V21M3 12H21" stroke="currentColor" stroke-width="1.5" stroke-linecap="round"/>
                  <path d="M7 7L17 17M17 7L7 17" stroke="currentColor" stroke-width="1" stroke-linecap="round" opacity="0.3"/>
                </svg>
              </div>
              <div class="action-info-compact">
                <h4>创建新项目</h4>
                <p>创建全新的项目</p>
              </div>
              <div class="action-radio">
                <div class="radio-button" :class="{ checked: projectAction === 'new' }"></div>
              </div>
            </div>

            <div class="action-option-compact" @click="projectAction = 'temp'" :class="{ active: projectAction === 'temp' }">
              <div class="action-sword-icon">
                <svg viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg">
                  <path d="M13 3L4 14H12L11 21L20 10H12L13 3Z" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round"/>
                </svg>
              </div>
              <div class="action-info-compact">
                <h4>启动临时项目</h4>
                <p>快速启动，自动保存</p>
              </div>
              <div class="action-radio">
                <div class="radio-button" :class="{ checked: projectAction === 'temp' }"></div>
              </div>
            </div>
          </div>
        </div>

        <!-- 右侧面板：操作详情 -->
        <div class="right-panel">
          <Transition name="fade-slide" mode="out-in">
            <!-- 打开现有项目 -->
            <div v-if="projectAction === 'open'" key="open" class="project-list-compact">
              <div class="details-header">
                <h3>选择项目</h3>
                <div class="project-count">{{ localProjects.length }} 个项目</div>
              </div>

              <div class="project-list-ultra-compact" v-if="localProjects.length > 0">
                <div v-for="project in localProjects" :key="project.id"
                     class="project-item-compact"
                     :class="{ selected: selectedProject?.id === project.id }"
                     @click="selectedProject = project">
                  <div class="project-sword-icon">
                    <svg viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg">
                      <path d="M4 20L7 17M20 4L11 13M11 13L8 16L4 20M11 13L14 10" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round"/>
                      <path d="M15 5L19 9" stroke="currentColor" stroke-width="1.5" stroke-linecap="round"/>
                    </svg>
                  </div>
                  <div class="project-item-info">
                    <div class="project-item-name">{{ project.name }}</div>
                    <div class="project-item-meta">
                      <span class="project-source local">本地</span>
                      <span class="project-requests">{{ project.requests?.toLocaleString() || 0 }} 请求</span>
                      <span v-if="project.size_formatted" class="project-size">{{ project.size_formatted }}</span>
                    </div>
                  </div>
                  <div class="project-item-actions">
                    <button
                      class="delete-btn"
                      @click="openDeleteConfirm(project, $event)"
                      :title="t('common.actions.delete')"
                    >
                      <i class="bx bx-trash"></i>
                    </button>
                  </div>
                  <div class="project-item-selector">
                    <div class="radio-button" :class="{ checked: selectedProject?.id === project.id }"></div>
                  </div>
                </div>
              </div>

              <div v-else class="empty-state-compact">
                <i class="fas fa-folder-open"></i>
                <h4>没有找到项目</h4>
                <p>请先创建一个新项目</p>
              </div>
            </div>

            <!-- 创建新项目 -->
            <div v-else-if="projectAction === 'new'" key="new" class="new-project-compact">
              <div class="details-header">
                <h3>创建新项目</h3>
              </div>
              <div class="input-card-compact">
                <div class="input-header">
                  <i class="fas fa-edit"></i>
                  <span>项目名称</span>
                </div>
                <input v-model="projectName"
                       type="text"
                       class="project-name-input-compact"
                       placeholder="请输入新项目名称..."
                       @keyup.enter="handleNext" spellcheck="false">
                <div class="input-hint">
                  项目名称将用于标识您的测试项目
                </div>
              </div>
            </div>

            <!-- 临时项目 -->
            <div v-else-if="projectAction === 'temp'" key="temp" class="temp-project-compact">
              <div class="details-header">
                <h3>临时项目说明</h3>
              </div>
              <div class="temp-info-card-compact">
                <div class="temp-icon">
                  <i class="fas fa-info-circle"></i>
                </div>
                <div class="temp-content">
                  <p>临时项目会自动创建带时间戳的数据库文件，数据会被保存。适合快速测试使用。</p>
                  <div class="temp-features">
                    <div class="feature-item">
                      <i class="fas fa-bolt"></i>
                      <span>快速启动</span>
                    </div>
                    <div class="feature-item">
                      <i class="fas fa-save"></i>
                      <span>自动保存</span>
                    </div>
                  </div>
                </div>
              </div>
            </div>

            <!-- 未选择操作时的提示 -->
            <div v-else key="empty" class="empty-state-compact">
              <i class="fas fa-hand-pointer"></i>
              <h4>请选择操作</h4>
              <p>从左侧选择要执行的项目操作</p>
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
