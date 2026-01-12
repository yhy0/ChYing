/**
 * 攻击类型枚举
 */
export type AttackType = 'sniper' | 'battering-ram' | 'pitchfork' | 'cluster-bomb';

/**
 * 载荷位置接口
 */
export interface PayloadPosition {
  start: number;
  end: number;
  value: string;
  paramName?: string;
  index: number;  // 添加 index 属性来标识位置的序号
}

/**
 * 处理规则接口
 */
export interface ProcessingRule {
  id: number;
  type: string;
  config: Record<string, any>;
}

/**
 * 编码配置接口
 */
export interface EncodingConfig {
  enabled: boolean;
  urlEncode: boolean;
  characterSet: string;
}

/**
 * 载荷集接口
 */
export interface PayloadSet {
  id: number;
  type: 'simple-list' | 'numbers' | 'brute-force' | 'custom';
  items: string[];
  processing: {
    rules: ProcessingRule[];
    encoding: EncodingConfig;
  };
}

/**
 * 攻击结果接口
 */
export interface AttackResult {
  id: string;
  status: number;
  length: number;
  payload: string[];
  timeMs: number;
  request: string;
  response: string;
  timestamp: number;
}

/**
 * Intruder 结果项接口
 */
export interface IntruderResult {
  id: number;
  payload: string[];
  status: number;
  length: number;
  timeMs: number;
  timestamp: string;
  request: string;
  response: string;
  selected?: boolean;
  color?: string;
  [key: string]: any;
}

/**
 * Intruder 标签页接口
 */
export interface IntruderTab {
  id: string;
  name: string;
  color: string;
  groupId: string | null;
  target: {
    url: string;
    method: string;
    headers: string;
    body: string;
    fullRequest: string;
  };
  attackType: AttackType;
  payloadPositions: PayloadPosition[];
  payloadSets: PayloadSet[];
  results: IntruderResult[];
  isActive: boolean;
  isRunning: boolean;
  progress: {
    total: number;
    current: number;
    startTime: number | null;
    endTime: number | null;
  };
}

/**
 * Intruder 分组接口
 */
export interface IntruderGroup {
  id: string;
  name: string;
  color: string;
}

/**
 * 请求编辑器实例接口
 */
export interface RequestEditorInstance {
  wrapSelectionWithMarker?: () => void;
  clearAllMarkers?: () => void;
}