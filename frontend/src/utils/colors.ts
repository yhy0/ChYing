// 颜色常量定义
// 参考 Windi CSS 颜色方案 https://cn.windicss.org/utilities/general/colors.html

// 预定义的浅色模式颜色选项 - 优化透明度，增强可见性
export const LIGHT_COLORS = [
  { id: 'none', color: '' },
  { id: 'red', color: 'rgba(239, 68, 68, 0.75)' },       // red-500 - 提高饱和度和透明度
  { id: 'orange', color: 'rgba(249, 115, 22, 0.75)' },   // orange-500
  { id: 'amber', color: 'rgba(245, 158, 11, 0.75)' },    // amber-500
  { id: 'yellow', color: 'rgba(234, 179, 8, 0.7)' },     // yellow-500
  { id: 'lime', color: 'rgba(132, 204, 22, 0.75)' },     // lime-500
  { id: 'green', color: 'rgba(34, 197, 94, 0.75)' },     // green-500
  { id: 'emerald', color: 'rgba(16, 185, 129, 0.75)' },  // emerald-500
  { id: 'teal', color: 'rgba(20, 184, 166, 0.75)' },     // teal-500
  { id: 'cyan', color: 'rgba(6, 182, 212, 0.75)' },      // cyan-500
  { id: 'sky', color: 'rgba(14, 165, 233, 0.75)' },      // sky-500
  { id: 'blue', color: 'rgba(59, 130, 246, 0.75)' },     // blue-500
  { id: 'indigo', color: 'rgba(99, 102, 241, 0.75)' },   // indigo-500
  { id: 'violet', color: 'rgba(139, 92, 246, 0.75)' },   // violet-500
  { id: 'purple', color: 'rgba(147, 51, 234, 0.75)' },   // purple-500
  { id: 'fuchsia', color: 'rgba(217, 70, 239, 0.75)' },  // fuchsia-500
  { id: 'pink', color: 'rgba(236, 72, 153, 0.75)' },     // pink-500
  { id: 'rose', color: 'rgba(244, 63, 94, 0.75)' },      // rose-500
  { id: 'gray', color: 'rgba(107, 114, 128, 0.7)' },     // gray-500
];

// 预定义的深色模式颜色选项 - 优化透明度，增强可见性
export const DARK_COLORS = [
  { id: 'none', color: '' },
  { id: 'red', color: 'rgba(248, 113, 113, 0.65)' },     // red-400 - 使用更亮的色调和更高透明度
  { id: 'orange', color: 'rgba(251, 146, 60, 0.65)' },   // orange-400
  { id: 'amber', color: 'rgba(251, 191, 36, 0.65)' },    // amber-400
  { id: 'yellow', color: 'rgba(250, 204, 21, 0.6)' },    // yellow-400
  { id: 'lime', color: 'rgba(163, 230, 53, 0.65)' },     // lime-400
  { id: 'green', color: 'rgba(74, 222, 128, 0.65)' },    // green-400
  { id: 'emerald', color: 'rgba(52, 211, 153, 0.65)' },  // emerald-400
  { id: 'teal', color: 'rgba(45, 212, 191, 0.65)' },     // teal-400
  { id: 'cyan', color: 'rgba(34, 211, 238, 0.65)' },     // cyan-400
  { id: 'sky', color: 'rgba(56, 189, 248, 0.65)' },      // sky-400
  { id: 'blue', color: 'rgba(96, 165, 250, 0.65)' },     // blue-400
  { id: 'indigo', color: 'rgba(129, 140, 248, 0.65)' },  // indigo-400
  { id: 'violet', color: 'rgba(167, 139, 250, 0.65)' },  // violet-400
  { id: 'purple', color: 'rgba(196, 181, 253, 0.65)' },  // purple-400
  { id: 'fuchsia', color: 'rgba(232, 121, 249, 0.65)' }, // fuchsia-400
  { id: 'pink', color: 'rgba(244, 114, 182, 0.65)' },    // pink-400
  { id: 'rose', color: 'rgba(251, 113, 133, 0.65)' },    // rose-400
  { id: 'gray', color: 'rgba(156, 163, 175, 0.6)' },     // gray-400
];

/**
 * 根据当前主题返回对应的颜色数组
 * @returns 当前主题应使用的颜色数组
 */
export const getThemeColors = () => {
  if (typeof document !== 'undefined' && 
      document.documentElement && 
      document.documentElement.classList.contains('dark')) {
    return DARK_COLORS;
  }
  return LIGHT_COLORS;
};

// 模块名称专用颜色方案（去掉 none 和 gray，提供更丰富的颜色）
const MODULE_COLORS = [
  'red', 'orange', 'amber', 'yellow', 'lime', 'green', 'emerald', 
  'teal', 'cyan', 'sky', 'blue', 'indigo', 'violet', 'purple', 
  'fuchsia', 'pink', 'rose'
];

/**
 * 为模块名称分配一致的颜色
 * 使用简单的哈希算法确保相同的模块名总是得到相同的颜色
 * @param moduleName 模块名称
 * @returns 颜色对象，包含id和对应主题下的颜色值
 */
export const getModuleColor = (moduleName: string) => {
  // 简单的字符串哈希函数
  let hash = 0;
  for (let i = 0; i < moduleName.length; i++) {
    const char = moduleName.charCodeAt(i);
    hash = ((hash << 5) - hash) + char;
    hash = hash & hash; // 转为32位整数
  }
  
  // 取绝对值并映射到颜色数组索引
  const colorIndex = Math.abs(hash) % MODULE_COLORS.length;
  const colorId = MODULE_COLORS[colorIndex];
  
  // 根据当前主题返回对应的颜色
  const themeColors = getThemeColors();
  const colorItem = themeColors.find(c => c.id === colorId);
  
  return {
    id: colorId,
    color: colorItem?.color || '',
    textClass: `text-${colorId}-600` // CSS类名，可用于文本颜色
  };
};

/**
 * 为插件名称分配一致的颜色
 * 使用与模块颜色相同的算法，但使用不同的颜色方案以避免冲突
 * @param pluginName 插件名称
 * @returns 颜色对象，包含id和对应主题下的颜色值
 */
export const getPluginColor = (pluginName: string) => {
  // 简单的字符串哈希函数
  let hash = 0;
  for (let i = 0; i < pluginName.length; i++) {
    const char = pluginName.charCodeAt(i);
    hash = ((hash << 5) - hash) + char;
    hash = hash & hash; // 转为32位整数
  }
  
  // 取绝对值并映射到颜色数组索引
  const colorIndex = Math.abs(hash) % MODULE_COLORS.length;
  const colorId = MODULE_COLORS[colorIndex];
  
  // 根据当前主题返回对应的颜色
  const themeColors = getThemeColors();
  const colorItem = themeColors.find(c => c.id === colorId);
  
  return {
    id: colorId,
    color: colorItem?.color || '',
    textClass: `text-${colorId}-600` // CSS类名，可用于文本颜色
  };
}; 