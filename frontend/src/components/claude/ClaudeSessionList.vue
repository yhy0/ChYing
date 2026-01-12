<script setup lang="ts">
import { ref, computed, onMounted, watch } from 'vue';
import { useI18n } from 'vue-i18n';
import { useClaudeStore } from '../../store/claude';
import type { Session } from '../../types/claude';

const { t } = useI18n();
const claudeStore = useClaudeStore();

// 会话列表
const sessionList = ref<Session[]>([]);
const isLoadingSessions = ref(false);
const searchQuery = ref('');

// 项目筛选
const projectList = ref<string[]>([]);
const selectedProjectFilter = ref<string>('current'); // 'current' | 'all' | 具体项目ID
const showProjectDropdown = ref(false);

// 当前项目ID（从props或store获取）
const props = defineProps<{
  currentProjectId?: string;
}>();

const currentProjectId = computed(() => props.currentProjectId || 'default');

// 计算属性
const currentSessionId = computed(() => claudeStore.currentSessionId);
const isLoading = computed(() => claudeStore.loading);
const isStreaming = computed(() => claudeStore.streaming);

// 筛选器显示文本
const filterDisplayText = computed(() => {
  if (selectedProjectFilter.value === 'current') {
    return t('claude.sessions.filter.currentProject', 'Current Project');
  } else if (selectedProjectFilter.value === 'all') {
    return t('claude.sessions.filter.allProjects', 'All Projects');
  } else {
    return selectedProjectFilter.value;
  }
});

// 过滤后的会话列表
const filteredSessions = computed(() => {
  if (!searchQuery.value.trim()) {
    return sessionList.value;
  }
  const query = searchQuery.value.toLowerCase();
  return sessionList.value.filter(session => {
    const title = getSessionTitle(session).toLowerCase();
    return title.includes(query);
  });
});

// 按日期分组的会话
const groupedSessions = computed(() => {
  const groups: { label: string; sessions: Session[] }[] = [];
  const today = new Date();
  today.setHours(0, 0, 0, 0);
  const yesterday = new Date(today);
  yesterday.setDate(yesterday.getDate() - 1);
  const lastWeek = new Date(today);
  lastWeek.setDate(lastWeek.getDate() - 7);
  const lastMonth = new Date(today);
  lastMonth.setMonth(lastMonth.getMonth() - 1);

  const todaySessions: Session[] = [];
  const yesterdaySessions: Session[] = [];
  const lastWeekSessions: Session[] = [];
  const lastMonthSessions: Session[] = [];
  const olderSessions: Session[] = [];

  filteredSessions.value.forEach(session => {
    const sessionDate = new Date(session.updatedAt || session.createdAt);
    sessionDate.setHours(0, 0, 0, 0);

    if (sessionDate >= today) {
      todaySessions.push(session);
    } else if (sessionDate >= yesterday) {
      yesterdaySessions.push(session);
    } else if (sessionDate >= lastWeek) {
      lastWeekSessions.push(session);
    } else if (sessionDate >= lastMonth) {
      lastMonthSessions.push(session);
    } else {
      olderSessions.push(session);
    }
  });

  if (todaySessions.length > 0) {
    groups.push({ label: t('claude.sessions.today', 'Today'), sessions: todaySessions });
  }
  if (yesterdaySessions.length > 0) {
    groups.push({ label: t('claude.sessions.yesterday', 'Yesterday'), sessions: yesterdaySessions });
  }
  if (lastWeekSessions.length > 0) {
    groups.push({ label: t('claude.sessions.lastWeek', 'Last 7 Days'), sessions: lastWeekSessions });
  }
  if (lastMonthSessions.length > 0) {
    groups.push({ label: t('claude.sessions.lastMonth', 'Last 30 Days'), sessions: lastMonthSessions });
  }
  if (olderSessions.length > 0) {
    groups.push({ label: t('claude.sessions.older', 'Older'), sessions: olderSessions });
  }

  return groups;
});

// 获取会话标题（基于第一条用户消息或默认标题）
const getSessionTitle = (session: Session): string => {
  if (session.history && session.history.length > 0) {
    const firstUserMessage = session.history.find(msg => msg.role === 'user');
    if (firstUserMessage) {
      const content = firstUserMessage.content.trim();
      // 使用展开运算符安全截取 UTF-8 字符串
      const chars = [...content];
      return chars.length > 50 ? chars.slice(0, 50).join('') + '...' : content;
    }
  }
  // 默认标题：使用创建时间
  const date = new Date(session.createdAt);
  return t('claude.sessions.newSession', 'New Session') + ' - ' + formatTime(date);
};

// 格式化时间
const formatTime = (date: Date): string => {
  const now = new Date();
  const diff = now.getTime() - date.getTime();
  const minutes = Math.floor(diff / 60000);
  const hours = Math.floor(diff / 3600000);
  const days = Math.floor(diff / 86400000);

  if (minutes < 1) {
    return t('claude.sessions.justNow', 'Just now');
  } else if (minutes < 60) {
    return t('claude.sessions.minutesAgo', { n: minutes });
  } else if (hours < 24) {
    return t('claude.sessions.hoursAgo', { n: hours });
  } else if (days < 7) {
    return t('claude.sessions.daysAgo', { n: days });
  } else {
    return date.toLocaleDateString();
  }
};

// 加载项目列表
const loadProjects = async () => {
  try {
    projectList.value = await claudeStore.listProjects();
  } catch (e) {
    console.error('Failed to load projects:', e);
  }
};

// 加载会话列表
const loadSessions = async () => {
  isLoadingSessions.value = true;
  try {
    let sessions: Session[];
    if (selectedProjectFilter.value === 'current') {
      sessions = await claudeStore.listSessionsByProject(currentProjectId.value);
    } else if (selectedProjectFilter.value === 'all') {
      sessions = await claudeStore.listSessions();
    } else {
      // 特定项目
      sessions = await claudeStore.listSessionsByProject(selectedProjectFilter.value);
    }
    // 按更新时间倒序排列
    sessionList.value = sessions.sort((a, b) => {
      const dateA = new Date(a.updatedAt || a.createdAt).getTime();
      const dateB = new Date(b.updatedAt || b.createdAt).getTime();
      return dateB - dateA;
    });
  } catch (e) {
    console.error('Failed to load sessions:', e);
  } finally {
    isLoadingSessions.value = false;
  }
};

// 选择项目筛选
const selectProjectFilter = (filter: string) => {
  selectedProjectFilter.value = filter;
  showProjectDropdown.value = false;
  loadSessions();
};

// 切换下拉菜单
const toggleProjectDropdown = () => {
  showProjectDropdown.value = !showProjectDropdown.value;
  if (showProjectDropdown.value) {
    loadProjects();
  }
};

// 切换会话
const switchToSession = async (sessionId: string) => {
  if (sessionId === currentSessionId.value || isStreaming.value) return;
  await claudeStore.switchSession(sessionId);
};

// 删除会话
const deleteSession = async (sessionId: string, event: Event) => {
  event.stopPropagation();
  
  if (!confirm(t('claude.sessions.confirmDelete', 'Are you sure you want to delete this session?'))) {
    return;
  }
  
  const isDeletingCurrentSession = sessionId === currentSessionId.value;
  
  await claudeStore.deleteSession(sessionId);
  // 从本地列表中移除
  sessionList.value = sessionList.value.filter(s => s.id !== sessionId);
  
  // 如果删除的是当前会话，切换到另一个会话或创建新会话
  if (isDeletingCurrentSession) {
    if (sessionList.value.length > 0) {
      // 切换到列表中的第一个会话
      await claudeStore.switchSession(sessionList.value[0].id);
    } else {
      // 没有其他会话，触发创建新会话
      emit('newSession');
    }
  }
};

// 创建新会话
const emit = defineEmits<{
  (e: 'newSession'): void;
}>();

const createNewSession = () => {
  emit('newSession');
};

// 监听当前会话变化，刷新列表
watch(currentSessionId, () => {
  loadSessions();
});

// 组件挂载时加载会话列表
onMounted(() => {
  loadSessions();
});

// 暴露刷新方法
defineExpose({
  refresh: loadSessions
});
</script>

<template>
  <div class="session-list-container">
    <!-- 头部 -->
    <div class="session-list-header">
      <h4>{{ t('claude.sessions.title', 'Sessions') }}</h4>
      <button
        @click="createNewSession"
        class="new-session-btn"
        :disabled="isLoading || isStreaming"
        :title="t('claude.actions.newSession', 'New Session')"
      >
        <i class="bx bx-plus"></i>
      </button>
    </div>

    <!-- 项目筛选器 -->
    <div class="project-filter">
      <button
        class="filter-trigger"
        @click="toggleProjectDropdown"
      >
        <i class="bx bx-filter-alt"></i>
        <span class="filter-text">{{ filterDisplayText }}</span>
        <i class="bx" :class="showProjectDropdown ? 'bx-chevron-up' : 'bx-chevron-down'"></i>
      </button>

      <!-- 下拉菜单 -->
      <div v-if="showProjectDropdown" class="filter-dropdown">
        <div
          class="filter-option"
          :class="{ active: selectedProjectFilter === 'current' }"
          @click="selectProjectFilter('current')"
        >
          <i class="bx bx-folder"></i>
          <span>{{ t('claude.sessions.filter.currentProject', 'Current Project') }}</span>
          <i v-if="selectedProjectFilter === 'current'" class="bx bx-check"></i>
        </div>
        <div
          class="filter-option"
          :class="{ active: selectedProjectFilter === 'all' }"
          @click="selectProjectFilter('all')"
        >
          <i class="bx bx-globe"></i>
          <span>{{ t('claude.sessions.filter.allProjects', 'All Projects') }}</span>
          <i v-if="selectedProjectFilter === 'all'" class="bx bx-check"></i>
        </div>

        <!-- 分隔线 -->
        <div v-if="projectList.length > 0" class="filter-divider"></div>

        <!-- 具体项目列表 -->
        <div
          v-for="project in projectList"
          :key="project"
          class="filter-option"
          :class="{ active: selectedProjectFilter === project }"
          @click="selectProjectFilter(project)"
        >
          <i class="bx bx-folder-open"></i>
          <span class="project-name">{{ project }}</span>
          <i v-if="selectedProjectFilter === project" class="bx bx-check"></i>
        </div>
      </div>

      <!-- 点击外部关闭 -->
      <div
        v-if="showProjectDropdown"
        class="filter-backdrop"
        @click="showProjectDropdown = false"
      ></div>
    </div>

    <!-- 搜索框 -->
    <div class="session-search">
      <i class="bx bx-search search-icon"></i>
      <input
        v-model="searchQuery"
        type="text"
        :placeholder="t('claude.sessions.search', 'Search sessions...')"
        class="search-input"
      />
      <button
        v-if="searchQuery"
        @click="searchQuery = ''"
        class="clear-search-btn"
      >
        <i class="bx bx-x"></i>
      </button>
    </div>

    <!-- 会话列表 -->
    <div class="session-list scrollbar-thin">
      <!-- 加载状态 -->
      <div v-if="isLoadingSessions" class="loading-state">
        <i class="bx bx-loader-alt bx-spin"></i>
        <span>{{ t('common.loading', 'Loading...') }}</span>
      </div>

      <!-- 空状态 -->
      <div v-else-if="filteredSessions.length === 0" class="empty-state">
        <i class="bx bx-chat"></i>
        <span v-if="searchQuery">{{ t('claude.sessions.noResults', 'No sessions found') }}</span>
        <span v-else>{{ t('claude.sessions.empty', 'No sessions yet') }}</span>
      </div>

      <!-- 分组会话列表 -->
      <template v-else>
        <div
          v-for="group in groupedSessions"
          :key="group.label"
          class="session-group"
        >
          <div class="group-label">{{ group.label }}</div>
          <div
            v-for="session in group.sessions"
            :key="session.id"
            class="session-item"
            :class="{ 
              'active': session.id === currentSessionId,
              'disabled': isStreaming && session.id !== currentSessionId
            }"
            @click="switchToSession(session.id)"
          >
            <div class="session-icon">
              <i class="bx bx-message-square-dots"></i>
            </div>
            <div class="session-info">
              <div class="session-title">{{ getSessionTitle(session) }}</div>
              <div class="session-meta">
                <span class="session-time">{{ formatTime(new Date(session.updatedAt || session.createdAt)) }}</span>
                <span v-if="session.history?.length" class="session-count">
                  {{ session.history.length }} {{ t('claude.sessions.messages', 'msgs') }}
                </span>
              </div>
            </div>
            <button
              class="delete-btn"
              @click="deleteSession(session.id, $event)"
              :title="t('common.delete', 'Delete')"
            >
              <i class="bx bx-trash"></i>
            </button>
          </div>
        </div>
      </template>
    </div>

    <!-- 底部刷新按钮 -->
    <div class="session-list-footer">
      <button
        @click="loadSessions"
        class="refresh-btn"
        :disabled="isLoadingSessions"
      >
        <i class="bx bx-refresh" :class="{ 'bx-spin': isLoadingSessions }"></i>
        {{ t('common.refresh', 'Refresh') }}
      </button>
    </div>
  </div>
</template>

<style scoped>
.session-list-container {
  display: flex;
  flex-direction: column;
  height: 100%;
  background: var(--color-bg-secondary);
}

/* 头部 */
.session-list-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 12px 16px;
  border-bottom: 1px solid var(--color-border);
}

.session-list-header h4 {
  margin: 0;
  font-size: 0.875rem;
  font-weight: 600;
  color: var(--color-text-primary);
}

.new-session-btn {
  width: 28px;
  height: 28px;
  border-radius: 6px;
  border: 1px solid var(--color-border);
  background: var(--color-bg-primary);
  color: var(--color-text-primary);
  cursor: pointer;
  display: flex;
  align-items: center;
  justify-content: center;
  transition: all 0.2s ease;
}

.new-session-btn:hover:not(:disabled) {
  background: var(--color-primary);
  border-color: var(--color-primary);
  color: white;
}

.new-session-btn:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

/* 搜索框 */
.session-search {
  position: relative;
  padding: 8px 12px;
  border-bottom: 1px solid var(--color-border);
}

.search-icon {
  position: absolute;
  left: 20px;
  top: 50%;
  transform: translateY(-50%);
  color: var(--color-text-tertiary);
  font-size: 1rem;
}

.search-input {
  width: 100%;
  padding: 8px 32px;
  border: 1px solid var(--color-border);
  border-radius: 6px;
  background: var(--color-bg-primary);
  color: var(--color-text-primary);
  font-size: 0.8125rem;
  outline: none;
  transition: border-color 0.2s ease;
}

.search-input:focus {
  border-color: var(--color-primary);
}

.search-input::placeholder {
  color: var(--color-text-tertiary);
}

.clear-search-btn {
  position: absolute;
  right: 20px;
  top: 50%;
  transform: translateY(-50%);
  width: 20px;
  height: 20px;
  border: none;
  background: transparent;
  color: var(--color-text-tertiary);
  cursor: pointer;
  display: flex;
  align-items: center;
  justify-content: center;
  border-radius: 4px;
  transition: all 0.2s ease;
}

.clear-search-btn:hover {
  background: var(--color-bg-tertiary);
  color: var(--color-text-primary);
}

/* 会话列表 */
.session-list {
  flex: 1;
  overflow-y: auto;
  padding: 8px;
}

/* 加载状态 */
.loading-state,
.empty-state {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: 32px 16px;
  color: var(--color-text-tertiary);
  gap: 8px;
}

.loading-state i,
.empty-state i {
  font-size: 2rem;
}

/* 分组 */
.session-group {
  margin-bottom: 16px;
}

.group-label {
  padding: 4px 8px;
  font-size: 0.6875rem;
  font-weight: 600;
  color: var(--color-text-tertiary);
  text-transform: uppercase;
  letter-spacing: 0.5px;
}

/* 会话项 */
.session-item {
  display: flex;
  align-items: flex-start;
  gap: 10px;
  padding: 10px 12px;
  border-radius: 8px;
  cursor: pointer;
  transition: all 0.2s ease;
  position: relative;
}

.session-item:hover {
  background: var(--color-bg-tertiary);
}

.session-item.active {
  background: var(--color-primary-bg, rgba(59, 130, 246, 0.1));
}

.session-item.active .session-icon i {
  color: var(--color-primary);
}

.session-item.disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.session-icon {
  width: 32px;
  height: 32px;
  border-radius: 6px;
  background: var(--color-bg-primary);
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
}

.session-icon i {
  font-size: 1rem;
  color: var(--color-text-secondary);
}

.session-info {
  flex: 1;
  min-width: 0;
  overflow: hidden;
}

.session-title {
  font-size: 0.8125rem;
  font-weight: 500;
  color: var(--color-text-primary);
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
  line-height: 1.4;
}

.session-meta {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-top: 2px;
}

.session-time,
.session-count {
  font-size: 0.6875rem;
  color: var(--color-text-tertiary);
}

.session-count::before {
  content: '·';
  margin-right: 8px;
}

/* 删除按钮 */
.delete-btn {
  width: 24px;
  height: 24px;
  border: none;
  background: transparent;
  color: var(--color-text-tertiary);
  cursor: pointer;
  display: flex;
  align-items: center;
  justify-content: center;
  border-radius: 4px;
  opacity: 0;
  transition: all 0.2s ease;
  flex-shrink: 0;
}

.session-item:hover .delete-btn {
  opacity: 1;
}

.delete-btn:hover {
  background: var(--color-danger-bg, rgba(239, 68, 68, 0.1));
  color: var(--color-danger, #ef4444);
}

/* 底部 */
.session-list-footer {
  padding: 8px 12px;
  border-top: 1px solid var(--color-border);
}

.refresh-btn {
  width: 100%;
  padding: 8px 12px;
  border: 1px solid var(--color-border);
  border-radius: 6px;
  background: var(--color-bg-primary);
  color: var(--color-text-secondary);
  font-size: 0.75rem;
  cursor: pointer;
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 6px;
  transition: all 0.2s ease;
}

.refresh-btn:hover:not(:disabled) {
  background: var(--color-bg-tertiary);
  color: var(--color-text-primary);
}

.refresh-btn:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

/* 项目筛选器 */
.project-filter {
  position: relative;
  padding: 8px 12px;
  border-bottom: 1px solid var(--color-border);
}

.filter-trigger {
  width: 100%;
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 8px 12px;
  border: 1px solid var(--color-border);
  border-radius: 6px;
  background: var(--color-bg-primary);
  color: var(--color-text-primary);
  font-size: 0.8125rem;
  cursor: pointer;
  transition: all 0.2s ease;
}

.filter-trigger:hover {
  border-color: var(--color-primary);
  background: var(--color-bg-tertiary);
}

.filter-trigger i:first-child {
  color: var(--color-text-tertiary);
}

.filter-text {
  flex: 1;
  text-align: left;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.filter-trigger i:last-child {
  color: var(--color-text-tertiary);
  font-size: 1rem;
}

.filter-dropdown {
  position: absolute;
  top: calc(100% + 4px);
  left: 12px;
  right: 12px;
  background: var(--color-bg-primary);
  border: 1px solid var(--color-border);
  border-radius: 8px;
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.15);
  z-index: 100;
  max-height: 240px;
  overflow-y: auto;
}

.filter-option {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 10px 12px;
  cursor: pointer;
  transition: background 0.15s ease;
}

.filter-option:first-child {
  border-radius: 7px 7px 0 0;
}

.filter-option:last-child {
  border-radius: 0 0 7px 7px;
}

.filter-option:hover {
  background: var(--color-bg-tertiary);
}

.filter-option.active {
  background: var(--color-primary-bg, rgba(59, 130, 246, 0.1));
}

.filter-option i:first-child {
  font-size: 1rem;
  color: var(--color-text-secondary);
}

.filter-option.active i:first-child {
  color: var(--color-primary);
}

.filter-option span {
  flex: 1;
  font-size: 0.8125rem;
  color: var(--color-text-primary);
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.filter-option i.bx-check {
  color: var(--color-primary);
  font-size: 1.125rem;
}

.filter-divider {
  height: 1px;
  background: var(--color-border);
  margin: 4px 0;
}

.project-name {
  max-width: 150px;
}

.filter-backdrop {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  z-index: 99;
}
</style>
