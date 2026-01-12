// 内存使用信息类型定义
export interface MemoryInfo {
  // 总分配的内存（字节）
  alloc: number;
  // 总分配的内存（格式化字符串）
  allocFormatted: string;
  // 从系统分配的内存（字节）
  sys: number;
  // 从系统分配的内存（格式化字符串）
  sysFormatted: string;
  // 垃圾回收次数
  numGC: number;
  // 协程数量
  numGoroutine: number;
}