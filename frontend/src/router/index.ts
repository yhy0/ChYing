import { createRouter, createWebHashHistory } from 'vue-router';
import type { RouteRecordRaw } from 'vue-router';

// 导入视图组件
import ProjectSelection from '../views/ProjectSelection.vue';
import ScanLogView from '../views/ScanLogView.vue';
import VulnerabilityView from '../views/VulnerabilityView.vue';
import ClaudeAgentView from '../views/ClaudeAgentView.vue';

// 导入模块组件
import ProjectView from '../components/project/ProjectView.vue';
import ProxyView from '../components/proxy/ProxyView.vue';
import RepeaterView from '../components/repeater/RepeaterView.vue';
import IntruderView from '../components/intruder/IntruderView.vue';
import DecoderView from '../components/decoder/DecoderView.vue';
import PluginsView from '../components/plugins/PluginsView.vue';
import SettingsView from '../components/settings/SettingsView.vue';

// 导入布局组件
import MainLayout from '../components/layout/MainLayout.vue';

const routes: RouteRecordRaw[] = [
  // 项目选择页面（无布局）
  {
    path: '/',
    name: 'ProjectSelection',
    component: ProjectSelection
  },
  
  // 扫描日志窗口（无布局，用于新窗口）
  {
    path: '/scanLog',
    name: 'ScanLog',
    component: ScanLogView
  },
  
  // 漏洞列表窗口（无布局，用于新窗口）
  {
    path: '/vulnerability',
    name: 'Vulnerability',
    component: VulnerabilityView
  },

  // Claude AI Agent窗口（无布局，用于新窗口）
  {
    path: '/claude-agent',
    name: 'ClaudeAgent',
    component: ClaudeAgentView
  },
  
  // 主应用布局及其子路由
  {
    path: '/app',
    component: MainLayout,
    children: [
      {
        path: '',
        redirect: '/app/project'
      },
      {
        path: 'project',
        name: 'Project',
        component: ProjectView,
        meta: {
          title: 'modules.project.title',
          description: 'modules.project.description',
          icon: 'bx-folder'
        }
      },
      {
        path: 'proxy',
        name: 'Proxy',
        component: ProxyView,
        meta: {
          title: 'modules.proxy.title',
          description: 'modules.proxy.description',
          icon: 'bx-globe'
        }
      },
      {
        path: 'repeater',
        name: 'Repeater',
        component: RepeaterView,
        meta: {
          title: 'modules.repeater.title',
          description: 'modules.repeater.description',
          icon: 'bx-repeat'
        }
      },
      {
        path: 'intruder',
        name: 'Intruder',
        component: IntruderView,
        meta: {
          title: 'modules.intruder.title',
          description: 'modules.intruder.description',
          icon: 'bx-target-lock'
        }
      },
      {
        path: 'decoder',
        name: 'Decoder',
        component: DecoderView,
        meta: {
          title: 'modules.decoder.title',
          description: 'modules.decoder.description',
          icon: 'bx-code-alt'
        }
      },
      {
        path: 'plugins',
        name: 'Plugins',
        component: PluginsView,
        meta: {
          title: 'modules.plugins.title',
          description: 'modules.plugins.description',
          icon: 'bx-plug'
        }
      },
      {
        path: 'settings',
        name: 'Settings',
        component: SettingsView,
        meta: {
          title: 'common.settings',
          description: '',
          icon: 'bx-cog'
        }
      }
    ]
  }
];

const router = createRouter({
  history: createWebHashHistory(),
  routes
});

// 添加路由守卫来处理未匹配的路由
router.beforeEach((to, from, next) => {
  console.log("---1---", to, from)
  console.log("---2---", to.matched)
  // 如果路由不存在，重定向到首页
  if (!to.matched.length) {
    console.warn(`Route not found: ${to.path}, redirecting to /`);
    next('/');
  } else {
    next();
  }
});

export default router;
