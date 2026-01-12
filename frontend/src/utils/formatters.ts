/**
 * 格式化日期时间，返回时:分:秒格式
 * @param timestamp 时间戳字符串
 * @returns 格式化后的时间字符串
 */
export const formatDateTime = (timestamp: string): string => {
  const date = new Date(timestamp);
  return date.toLocaleTimeString(undefined, {
    hour: '2-digit',
    minute: '2-digit',
    second: '2-digit'
  });
};

/**
 * 格式化字节大小，将字节数转换为更友好的单位
 * @param bytes 字节数
 * @returns 格式化后的大小字符串
 */
export const formatSize = (bytes: number): string => {
  if (bytes < 1024) {
    return bytes + ' B';
  } else if (bytes < 1024 * 1024) {
    return (bytes / 1024).toFixed(1) + ' KB';
  } else {
    return (bytes / (1024 * 1024)).toFixed(1) + ' MB';
  }
}; 