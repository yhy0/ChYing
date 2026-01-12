/**
 * Jsluice 数据提取组合式函数
 * 
 * 将 Jsluice 调用和结果处理逻辑从组件中分离出来
 * 提供可复用的数据提取功能
 */

import { ref, watch, computed, readonly, type Ref } from 'vue';
// @ts-ignore
import { ExtractJsluiceData } from "../../bindings/github.com/yhy0/ChYing/app.js";

export interface JsluiceResult {
  urls?: string[];
  secrets?: string[];
  [key: string]: any;
}

export interface JsluiceOptions {
  /** 是否自动提取 */
  autoExtract?: boolean;
  /** 提取延迟（毫秒） */
  delay?: number;
  /** 是否启用缓存 */
  enableCache?: boolean;
  /** 缓存过期时间（毫秒） */
  cacheExpiry?: number;
}

interface CacheEntry {
  result: JsluiceResult;
  timestamp: number;
}

// 全局缓存
const cache = new Map<string, CacheEntry>();

/**
 * Jsluice 数据提取钩子
 */
export function useJsluiceExtractor(
  responseData: Ref<string>,
  uuid: Ref<string>,
  options: JsluiceOptions = {}
) {
  const {
    autoExtract = true,
    delay = 300,
    enableCache = true,
    cacheExpiry = 5 * 60 * 1000 // 5分钟
  } = options;

  // 提取结果
  const result = ref<JsluiceResult>({});
  const isLoading = ref(false);
  const error = ref<string | null>(null);

  // 计算缓存键
  const cacheKey = computed(() => {
    if (!responseData.value) return '';
    // 使用响应数据的哈希作为缓存键
    return `jsluice_${uuid.value}_${hashString(responseData.value)}`;
  });

  /**
   * 简单的字符串哈希函数
   */
  function hashString(str: string): string {
    let hash = 0;
    if (str.length === 0) return hash.toString();
    
    for (let i = 0; i < str.length; i++) {
      const char = str.charCodeAt(i);
      hash = ((hash << 5) - hash) + char;
      hash = hash & hash; // 转换为32位整数
    }
    
    return Math.abs(hash).toString(36);
  }

  /**
   * 从缓存获取结果
   */
  const getFromCache = (key: string): JsluiceResult | null => {
    if (!enableCache) return null;
    
    const entry = cache.get(key);
    if (!entry) return null;
    
    // 检查是否过期
    if (Date.now() - entry.timestamp > cacheExpiry) {
      cache.delete(key);
      return null;
    }
    
    return entry.result;
  };

  /**
   * 保存到缓存
   */
  const saveToCache = (key: string, data: JsluiceResult) => {
    if (!enableCache) return;
    
    cache.set(key, {
      result: data,
      timestamp: Date.now()
    });
  };

  /**
   * 执行 Jsluice 提取
   */
  const extract = async (): Promise<JsluiceResult> => {
    if (!responseData.value) {
      return {};
    }

    const key = cacheKey.value;
    
    // 尝试从缓存获取
    const cached = getFromCache(key);
    if (cached) {
      result.value = cached;
      return cached;
    }

    isLoading.value = true;
    error.value = null;

    try {
      // 调用 Wails 后端函数
      const extractedData = await ExtractJsluiceData(responseData.value) as JsluiceResult;

      result.value = extractedData;
      
      // 保存到缓存
      saveToCache(key, extractedData);
      
      return extractedData;
    } catch (err) {
      const errorMessage = err instanceof Error ? err.message : '提取失败';
      error.value = errorMessage;
      console.error('Jsluice 提取失败:', err);
      return {};
    } finally {
      isLoading.value = false;
    }
  };

  /**
   * 清除结果
   */
  const clear = () => {
    result.value = {};
    error.value = null;
  };

  /**
   * 清除缓存
   */
  const clearCache = () => {
    cache.clear();
  };

  /**
   * 获取特定类型的提取结果
   */
  const getUrls = computed(() => result.value.urls || []);
  const getSecrets = computed(() => result.value.secrets || []);
  
  /**
   * 检查是否有提取结果
   */
  const hasResults = computed(() => {
    return Object.keys(result.value).length > 0;
  });

  /**
   * 获取结果统计
   */
  const stats = computed(() => {
    return {
      urlCount: getUrls.value.length,
      secretCount: getSecrets.value.length,
      totalCount: Object.values(result.value).flat().length
    };
  });

  // 自动提取逻辑
  let timeoutId: number | null = null;

  if (autoExtract) {
    watch(
      [responseData, uuid],
      () => {
        // 清除之前的定时器
        if (timeoutId) {
          clearTimeout(timeoutId);
        }

        // 如果没有响应数据，清除结果
        if (!responseData.value) {
          clear();
          return;
        }

        // 延迟提取，避免频繁调用
        timeoutId = window.setTimeout(() => {
          extract();
          timeoutId = null;
        }, delay);
      },
      { immediate: true }
    );
  }

  // 清理定时器
  const cleanup = () => {
    if (timeoutId) {
      clearTimeout(timeoutId);
      timeoutId = null;
    }
  };

  return {
    // 状态
    result: readonly(result),
    isLoading: readonly(isLoading),
    error: readonly(error),
    
    // 计算属性
    hasResults,
    stats,
    getUrls,
    getSecrets,
    
    // 方法
    extract,
    clear,
    clearCache,
    cleanup
  };
}

/**
 * 批量 Jsluice 提取钩子
 * 用于同时处理多个响应数据
 */
export function useBatchJsluiceExtractor(
  responseDataList: Ref<Array<{ uuid: string; data: string }>>,
  options: JsluiceOptions = {}
) {
  const extractors = computed(() => {
    return responseDataList.value.map(item => {
      const dataRef = ref(item.data);
      const uuidRef = ref(item.uuid);
      
      return {
        uuid: item.uuid,
        extractor: useJsluiceExtractor(dataRef, uuidRef, {
          ...options,
          autoExtract: false // 批量模式下手动控制提取
        })
      };
    });
  });

  /**
   * 提取所有数据
   */
  const extractAll = async () => {
    const promises = extractors.value.map(({ extractor }) => extractor.extract());
    return Promise.all(promises);
  };

  /**
   * 清除所有结果
   */
  const clearAll = () => {
    extractors.value.forEach(({ extractor }) => extractor.clear());
  };

  /**
   * 获取所有结果
   */
  const allResults = computed(() => {
    return extractors.value.map(({ uuid, extractor }) => ({
      uuid,
      result: extractor.result,
      isLoading: extractor.isLoading,
      error: extractor.error
    }));
  });

  return {
    extractors,
    allResults,
    extractAll,
    clearAll
  };
}
