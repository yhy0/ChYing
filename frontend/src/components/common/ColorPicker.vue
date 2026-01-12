<script setup lang="ts">
import { onMounted, onUnmounted, nextTick, watch, ref, computed } from 'vue';
import { useI18n } from 'vue-i18n';
import { useClickOutside } from '@/composables/useClickOutside';

// 优化颜色选项接口
interface ColorOption {
  id: string;
  color: string;
  name?: string; // 可选的显示名称
}

// 预设组类型
interface ColorPreset {
  name: string;
  colors: ColorOption[];
}

const props = defineProps<{
  colors: ColorOption[];
  title?: string;
  position: { x: number; y: number };
  shown: boolean;
  maxHistory?: number; // 历史记录最大数量
  presets?: ColorPreset[]; // 预设颜色组
}>();

const emit = defineEmits<{
  (e: 'select', color: string): void;
  (e: 'close'): void;
}>();

const { t } = useI18n();

// 默认预设
const defaultPresets = [
  {
    name: '安全工具',
    colors: [
      { id: 'danger', color: '#ef4444', name: '危险' },
      { id: 'warning', color: '#f59e0b', name: '警告' },
      { id: 'success', color: '#10b981', name: '安全' },
      { id: 'info', color: '#3b82f6', name: '信息' }
    ]
  },
  {
    name: '流量标记',
    colors: [
      { id: 'auth', color: '#6366f1', name: '认证' },
      { id: 'api', color: '#8b5cf6', name: 'API' },
      { id: 'static', color: '#a3a3a3', name: '静态' },
      { id: 'sensitive', color: '#f43f5e', name: '敏感' }
    ]
  }
];

// 使用传入的预设或默认预设
const effectivePresets = computed(() => props.presets || defaultPresets);

// 历史记录
const colorHistory = ref<ColorOption[]>([]);
const maxHistory = computed(() => props.maxHistory || 5);

// 活动标签
const activeTab = ref('custom'); // 'custom', 'preset', 'history'

// 当前选择的预设组
const activePresetIndex = ref(0);

// 菜单引用
const menuRef = ref<HTMLElement | null>(null);

// 使用点击外部检测
const { enable, disable } = useClickOutside(
  menuRef,
  () => emit('close'),
  {
    enabled: computed(() => props.shown),
    delay: 100 // 延迟100ms避免立即触发
  }
);

// 点击外部检测已由 useClickOutside 组合式函数处理

// 选择颜色前先阻止事件冒泡并添加到历史记录
const selectColor = (event: MouseEvent, colorOption: ColorOption) => {
  event.stopPropagation();
  
  // 添加到历史记录
  addToHistory(colorOption);
  
  emit('select', colorOption.color);
};

// 添加颜色到历史记录
const addToHistory = (colorOption: ColorOption) => {
  // 如果已在历史记录中，先移除
  colorHistory.value = colorHistory.value.filter(item => item.color !== colorOption.color);
  
  // 添加到历史记录开头
  colorHistory.value.unshift({...colorOption});
  
  // 限制历史记录数量
  if (colorHistory.value.length > maxHistory.value) {
    colorHistory.value = colorHistory.value.slice(0, maxHistory.value);
  }
  
  // 尝试保存到localStorage
  try {
    localStorage.setItem('colorPickerHistory', JSON.stringify(colorHistory.value));
  } catch (e) {
    console.error('Failed to save color history:', e);
  }
};

// 从localStorage加载历史记录
const loadHistory = () => {
  try {
    const savedHistory = localStorage.getItem('colorPickerHistory');
    if (savedHistory) {
      colorHistory.value = JSON.parse(savedHistory);
    }
  } catch (e) {
    console.error('Failed to load color history:', e);
  }
};

// 阻止点击颜色菜单时的事件冒泡
const handleMenuClick = (event: MouseEvent) => {
  event.stopPropagation();
};

// 切换标签
const setActiveTab = (tab: string) => {
  activeTab.value = tab;
};

// 选择预设组
const selectPreset = (index: number) => {
  activePresetIndex.value = index;
};

// 初始化时加载历史记录
onMounted(() => {
  loadHistory();
});

// 事件清理已由 useClickOutside 组合式函数处理
</script>

<template>
  <teleport to="body">
    <div v-if="shown"
      ref="menuRef"
      class="color-menu-container"
      :style="{
        left: `${position.x}px`,
        top: `${position.y}px`,
        zIndex: 1000,
      }"
      @click="handleMenuClick"
    >
      <div class="color-menu-header">
        <span class="color-menu-title">{{ title || t('modules.proxy.colors.title') }}</span>
        <!-- 标签切换 -->
        <div class="color-menu-tabs">
          <button 
            @click="setActiveTab('custom')" 
            class="color-menu-tab"
            :class="{ active: activeTab === 'custom' }"
          >
            {{ t('common.custom') }}
          </button>
          <button 
            @click="setActiveTab('preset')" 
            class="color-menu-tab"
            :class="{ active: activeTab === 'preset' }"
          >
            {{ t('common.presets') }}
          </button>
          <button 
            @click="setActiveTab('history')" 
            class="color-menu-tab"
            :class="{ active: activeTab === 'history' }"
            v-if="colorHistory.length > 0"
          >
            {{ t('common.history') }}
          </button>
        </div>
      </div>
      
      <!-- 自定义颜色 -->
      <div v-if="activeTab === 'custom'" class="color-menu-content">
        <div class="color-picker">
          <button
            v-for="colorOption in colors"
            :key="colorOption.id"
            @click="(event) => selectColor(event, colorOption)"
            class="color-option"
            :class="{ none: colorOption.id === 'none' }"
            :style="{ backgroundColor: colorOption.color || 'transparent' }"
            :title="colorOption.id === 'none' ? t('common.ui.none') : (colorOption.name || t(`colors.${colorOption.id}`))"
          >
            <i v-if="colorOption.id === 'none'" class="bx bx-x"></i>
          </button>
        </div>
      </div>
      
      <!-- 预设颜色 -->
      <div v-else-if="activeTab === 'preset'" class="color-menu-content">
        <!-- 预设组选择器 -->
        <div class="preset-selector" v-if="effectivePresets.length > 1">
          <button 
            v-for="(preset, index) in effectivePresets" 
            :key="index"
            @click="selectPreset(index)"
            class="preset-button"
            :class="{ active: activePresetIndex === index }"
          >
            {{ preset.name }}
          </button>
        </div>
        
        <!-- 预设颜色 -->
        <div class="color-picker" v-if="effectivePresets.length > 0">
          <button
            v-for="colorOption in effectivePresets[activePresetIndex].colors"
            :key="colorOption.id"
            @click="(event) => selectColor(event, colorOption)"
            class="color-option"
            :style="{ backgroundColor: colorOption.color || 'transparent' }"
            :title="colorOption.name || t(`colors.${colorOption.id}`)"
          ></button>
        </div>
      </div>
      
      <!-- 历史记录 -->
      <div v-else-if="activeTab === 'history'" class="color-menu-content">
        <div class="color-picker color-history">
          <button
            v-for="(colorOption, index) in colorHistory"
            :key="`history-${index}`"
            @click="(event) => selectColor(event, colorOption)"
            class="color-option"
            :style="{ backgroundColor: colorOption.color || 'transparent' }"
            :title="colorOption.name || t(`colors.${colorOption.id}`)"
          ></button>
        </div>
      </div>
    </div>
  </teleport>
</template>

<!-- 样式已迁移到 styles/components/color-picker.css 统一管理 -->
