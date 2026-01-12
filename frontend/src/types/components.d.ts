import { message } from '../utils/message';

// 扩展ComponentCustomProperties接口，添加$message属性
declare module 'vue' {
  interface ComponentCustomProperties {
    $message: typeof message;
  }
} 