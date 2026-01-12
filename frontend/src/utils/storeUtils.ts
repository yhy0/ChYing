/**
 * Store 相关的工具函数
 */

/**
 * 生成 UUID
 * @returns 生成的 UUID
 */
export function generateUUID(): string {
  return 'xxxxxxxx-xxxx-4xxx-yxxx-xxxxxxxxxxxx'.replace(/[xy]/g, function(c) {
    const r = Math.random() * 16 | 0;
    const v = c === 'x' ? r : (r & 0x3 | 0x8);
    return v.toString(16);
  });
}

/**
 * 从请求头中提取方法和主机信息
 * @param headers 请求头字符串
 * @returns 提取的方法和主机信息
 */
export function extractMethodAndHost(headers: string): { method: string; host: string } {
  const method = headers.split(' ')[0] || 'GET';
  const hostMatch = headers.match(/Host:\s*([^\r\n]+)/i);
  const host = hostMatch ? hostMatch[1].trim() : '';
  return { method, host };
} 