<script setup lang="ts">
import { computed, ref, watch } from 'vue';
import { useRouter, useRoute } from 'vue-router';
import { useI18n } from 'vue-i18n';

const router = useRouter();
const route = useRoute();
const { t } = useI18n();

// 从路由获取当前活动模块
const activeModule = computed(() => {
  const routeName = route.name as string;
  return routeName?.toLowerCase() || 'project';
});

// 用于动画的前一个活动模块
const previousModule = ref(activeModule.value);

const modules = [
  { id: 'project', icon: 'bx-folder', label: 'modules.project.title', description: 'modules.project.description', route: '/app/project' },
  { id: 'proxy', icon: 'bx-globe', label: 'modules.proxy.title', description: 'modules.proxy.description', route: '/app/proxy' },
  { id: 'repeater', icon: 'bx-repeat', label: 'modules.repeater.title', description: 'modules.repeater.description', route: '/app/repeater' },
  { id: 'intruder', icon: 'bx-target-lock', label: 'modules.intruder.title', description: 'modules.intruder.description', route: '/app/intruder' },
  { id: 'decoder', icon: 'bx-code-alt', label: 'modules.decoder.title', description: 'modules.decoder.description', route: '/app/decoder' },
  { id: 'plugins', icon: 'bx-plug', label: 'modules.plugins.title', description: 'modules.plugins.description', route: '/app/plugins' },
  { id: 'settings', icon: 'bx-cog', label: 'common.settings', description: '', route: '/app/settings' }
];

// 监听模块变化以便动画
watch(activeModule, (_, oldModule) => {
  previousModule.value = oldModule;
});

// 添加切换状态管理
const isTransitioning = ref(false);

const switchModule = (moduleId: string) => {
  if (moduleId === activeModule.value || isTransitioning.value) return;

  isTransitioning.value = true;
  const module = modules.find(m => m.id === moduleId);

  if (module) {
    // 添加短暂延迟，让动画更流畅
    setTimeout(() => {
      router.push(module.route);
      // 动画完成后重置状态
      setTimeout(() => {
        isTransitioning.value = false;
      }, 200);
    }, 50);
  }
};
</script>

<template>
  <div class="glass-sidebar">
    <!-- 品牌标志区域 -->
    <div class="glass-sidebar-brand">
      <div class="glass-sidebar-brand-icon">
        <img src="/appicon.png" alt="ChYing Inside" class="glass-sidebar-brand-logo" />
      </div>
    </div>

    <!-- 导航模块容器 -->
    <div class="glass-sidebar-nav" :class="{ 'transitioning': isTransitioning }">
      <div
        v-for="module in modules"
        :key="module.id"
        :class="[
          'glass-sidebar-module',
          {
            'active': activeModule === module.id,
            'disabled': isTransitioning
          }
        ]"
        @click="switchModule(module.id)"
      >
        <i
          :class="[
            'bx',
            module.icon,
            'glass-sidebar-module-icon'
          ]"
        ></i>

        <!-- 活动指示器 -->
        <div class="glass-sidebar-indicator"></div>

        <!-- 悬浮提示 -->
        <div class="glass-sidebar-tooltip">
          <div class="glass-sidebar-tooltip-content">
            <div class="glass-sidebar-tooltip-title">{{ t(module.label) }}</div>
            <div v-if="module.description" class="glass-sidebar-tooltip-description">{{ t(module.description) }}</div>
          </div>
          <!-- 箭头 -->
          <div class="glass-sidebar-tooltip-arrow"></div>
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
/* 侧边栏组件已使用液态玻璃样式系统 */
/* 所有样式已移至 frontend/src/styles/components/sidebar.css */
</style>