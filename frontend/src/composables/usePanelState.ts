import { ref, computed } from 'vue';

/**
 * 管理请求/响应面板的状态
 * 根据UUID隔离不同请求的状态
 */
export function usePanelState(defaultUuid: string = 'default') {
  // 存储活跃模块与UUID的映射
  const activeModules = ref<Record<string, string | null>>({});

  // 面板宽度 - 根据UUID存储不同请求的面板宽度
  const panelWidthsStore = ref<Record<string, Record<string, number>>>({});

  // 笔记内容 - 根据UUID存储不同请求的笔记
  const notesContent = ref<Record<string, string>>({});

  // 展开状态 - 根据UUID存储不同请求的展开状态
  const expandedSectionsStore = ref<Record<string, Record<string, boolean>>>({});

  /**
   * 获取当前请求对应的活跃模块
   */
  const getActiveModule = (uuid: string = defaultUuid) => {
    return computed(() => activeModules.value[uuid] || null);
  };

  /**
   * 切换模块激活状态
   */
  const toggleModule = (moduleName: string, uuid: string = defaultUuid) => {
    if (activeModules.value[uuid] === moduleName) {
      activeModules.value[uuid] = null; // 如果点击已激活的模块，则关闭它
    } else {
      activeModules.value[uuid] = moduleName; // 否则激活点击的模块
    }
  };

  /**
   * 获取当前请求的面板宽度
   */
  const getPanelWidths = (uuid: string = defaultUuid) => {
    if (!panelWidthsStore.value[uuid]) {
      // 初始化该请求的面板宽度
      panelWidthsStore.value[uuid] = {
        inspector: 320,
        notes: 320
      };
    }
    
    return panelWidthsStore.value[uuid];
  };

  /**
   * 更新面板宽度
   */
  const updatePanelWidth = (panel: string, width: number, uuid: string = defaultUuid) => {
    const widths = getPanelWidths(uuid);
    panelWidthsStore.value[uuid][panel] = width;
  };

  /**
   * 获取/设置当前请求的笔记内容
   */
  const getNotes = (uuid: string = defaultUuid) => {
    return computed({
      get: () => notesContent.value[uuid] || '',
      set: (value: string) => {
        notesContent.value[uuid] = value;
      }
    });
  };

  /**
   * 获取当前请求的展开状态
   */
  const getExpandedSections = (uuid: string = defaultUuid) => {
    if (!expandedSectionsStore.value[uuid]) {
      // 初始化该请求的展开状态
      expandedSectionsStore.value[uuid] = {
        requestAttributes: false,
        queryParameters: false,
        cookies: false,
        requestHeaders: false,
        responseHeaders: false
      };
    }
    
    return expandedSectionsStore.value[uuid];
  };

  /**
   * 切换部分的展开/折叠状态
   */
  const toggleSection = (section: string, uuid: string = defaultUuid) => {
    const sections = getExpandedSections(uuid);
    
    // 关闭其他已展开的部分，实现手风琴效果
    Object.keys(sections).forEach(key => {
      if (key !== section) {
        expandedSectionsStore.value[uuid][key] = false;
      }
    });
    
    // 切换当前部分的状态
    expandedSectionsStore.value[uuid][section] = !expandedSectionsStore.value[uuid][section];
  };

  return {
    getActiveModule,
    toggleModule,
    getPanelWidths,
    updatePanelWidth,
    getNotes,
    getExpandedSections,
    toggleSection
  };
} 