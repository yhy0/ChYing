/**
 * Decoder 编码/解码方法类型
 */
export type DecoderMethod = 'URL' | 'Base64' | 'HTML' | 'Unicode' | 'Hex' | 'JWT' | 'MD5' | 'SHA1' | 'SHA256';

/**
 * Decoder 操作类型
 */
export type DecoderOperation = 'encode' | 'decode';

/**
 * Decoder 输入类型 - 这可能不再需要在每个标签级别定义
 */
// export type DecoderInputType = 'text' | 'hex' | 'binary'; // 注释掉，因为每个步骤可能有自己的格式

/**
 * 单个解码/编码步骤的状态接口
 */
export interface DecoderStep {
  id: string; // 步骤的唯一ID
  method: DecoderMethod;
  operation: DecoderOperation;
  inputText: string; // 此步骤的输入
  outputText: string; // 此步骤的输出
  error?: string; // 此步骤执行时的错误信息
}

/**
 * Decoder 模块的标签页状态接口（修改后）
 */
export interface DecoderTab {
  id: string;
  name: string;
  initialInput: string; // 初始输入文本
  steps: DecoderStep[]; // 应用的步骤列表
  isActive: boolean;
} 