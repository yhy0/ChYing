// 代理规则类型
export interface ScopeRule {
  id: string;
  type: string;
  prefix: string;
  enabled: boolean;
  regexp: boolean;
}

// 代理配置类型
export interface ProxyConfig {
  port: number;
  exclude: ScopeRule[];
  include: ScopeRule[];
  filterSuffix: string;
}

// 插件配置基础类型
export interface PluginConfig {
  enabled: boolean;
  [key: string]: any;
}

// 收集配置类型
export interface CollectionConfig {
  [key: string]: string[];
}

// HTTP配置类型
export interface HttpConfig {
  proxy: string;
  timeout: number;
  maxConnsPerHost: number;
  retryTimes: number;
  allowRedirect: number;
  verifySSL: boolean;
  maxQps: number;
  headers: Record<string, string>;
  forceHTTP1: boolean;
}

// 完整配置类型
export interface CompleteConfig {
  version: string;
  parallel: number;
  http: HttpConfig;
  plugins: {
    bruteForce: {
      web: boolean;
      service: boolean;
      usernameDict: string;
      passwordDict: string;
    };
    cmdInjection: PluginConfig;
    crlfInjection: PluginConfig;
    xss: {
      enabled: boolean;
      detectXssInCookie: boolean;
    };
    sql: {
      enabled: boolean;
      booleanBasedDetection: boolean;
      errorBasedDetection: boolean;
      timeBasedDetection: boolean;
      detectInCookie: boolean;
    };
    sqlmapApi: {
      enabled: boolean;
      url: string;
      username: string;
      password: string;
    };
    xxe: PluginConfig;
    ssrf: PluginConfig;
    bbscan: PluginConfig;
    jsonp: PluginConfig;
    log4j: PluginConfig;
    bypass403: PluginConfig;
    fastjson: PluginConfig;
    archive: PluginConfig;
    iis: PluginConfig;
    nginxAliasTraversal: PluginConfig;
    poc: PluginConfig;
    nuclei: PluginConfig;
    portScan: PluginConfig;
    [key: string]: any;
  };
  mitmproxy: {
    caCert: string;
    caKey: string;
    basicAuth: {
      header: string;
      username: string;
      password: string;
    };
    exclude: string[];
    include: string[];
    filterSuffix: string[];
    maxLength: number;
  };
  collection: CollectionConfig;
} 