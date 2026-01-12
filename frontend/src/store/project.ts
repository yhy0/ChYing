import { defineStore } from 'pinia';
import { ref } from 'vue';
import type { SiteMapNode } from '../types';

export const useProjectStore = defineStore('project', () => {
  // 站点地图数据
  const siteMapData = ref<SiteMapNode[]>([]);

  // 更新站点地图数据
  const updateSiteMapData = (nodes: SiteMapNode[]) => {
    siteMapData.value = nodes;
  };

  // 清空站点地图数据（切换项目时调用）
  const clearSiteMapData = () => {
    siteMapData.value = [];
  };

  return {
    siteMapData,
    updateSiteMapData,
    clearSiteMapData,
  };
});