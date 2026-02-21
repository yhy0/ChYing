import { defineConfig, presetUno, presetAttributify, presetTypography, presetIcons, presetWebFonts, presetTagify } from 'unocss'

export default defineConfig({
  presets: [
    presetUno(), // 基础预设 - 最重要的核心功能
    presetAttributify(), // 属性模式支持
    presetTypography(), // 排版预设
    presetIcons({
      scale: 1.2,
      warn: true,
    }), // 图标预设
    presetWebFonts({
      provider: 'none', // 不从远程获取字体
      fonts: {
        sans: [
          {
            name: 'Inter',
            provider: 'none',
          },
          {
            name: 'system-ui',
            provider: 'none',
          },
        ],
        mono: [
          {
            name: 'JetBrains Mono',
            provider: 'none',
          },
          {
            name: 'monospace',
            provider: 'none',
          },
        ],
      },
    }), // 网络字体
    presetTagify(), // 标签模式支持
  ],
  // 使用更明确的规则定义快捷方式
  rules: [
    ['theme-transition', { 'transition-property': 'color, background-color, border-color', 'transition-duration': '200ms' }],
    ['tab-active', { 'background-color': 'var(--color-bg-card, #282838)', 'color': 'white', 'border-left-width': '2px', 'border-color': 'var(--color-primary, #4f46e5)' }],
    ['font-maple', { 'font-family': 'Maple Mono, monospace' }]
  ],
  theme: {
    colors: {
      'primary': '#4f46e5',
      'secondary': '#6366f1',
      'accent': '#818cf8',
      'darkbg': '#1e1e2e',
      'darksurface': '#282838',
      'lightbg': '#ffffff',
      'lightsurface': '#f3f4f6',
      // 新增玻璃效果颜色，但不影响现有设计
      'glass': {
        'bg': {
          'primary': 'rgba(255, 255, 255, 0.1)',
          'secondary': 'rgba(255, 255, 255, 0.05)',
          'light': 'rgba(248, 250, 252, 0.85)',
        },
        'border': {
          'default': 'rgba(255, 255, 255, 0.2)',
          'light': 'rgba(148, 163, 184, 0.5)',
        }
      }
    }
  },
  safelist: [
    'scrollbar-thin',
    'theme-transition',
    'tab-active',
    'font-maple'
  ],
  // 使用preflights直接定义全局CSS
  preflights: [
    {
      layer: 'default',
      getCSS: () => `
        .scrollbar-thin::-webkit-scrollbar {
          width: 6px;
          height: 6px;
        }
        .scrollbar-thin::-webkit-scrollbar-track {
          background-color: var(--color-bg-secondary, #1e1e2e);
        }
        .scrollbar-thin::-webkit-scrollbar-thumb {
          background-color: var(--color-border, #4b5563);
          border-radius: 9999px;
        }

        /* 深色模式下的标签样式 */
        .dark .tab-active {
          background-color: var(--color-bg-card, #282838);
        }
      `
    }
  ],
  // 处理在HTML, JS, CSS中的UnoCSS指令
  content: {
    filesystem: [
      '**/*.{html,js,ts,jsx,tsx,vue,svelte,astro}',
    ]
  },
}) 