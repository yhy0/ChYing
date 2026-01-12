import { createApp } from 'vue';
import type { App } from 'vue';
import MessageContainer from '../components/common/MessageContainer.vue';

// 创建消息容器实例
let messageInstance: any = null;

// 获取消息容器实例
const getInstance = () => {
  if (!messageInstance) {
    // 创建一个容器元素
    const container = document.createElement('div');
    container.setAttribute('id', 'message-container');
    document.body.appendChild(container);
    
    // 创建Vue应用并挂载
    const app = createApp(MessageContainer);
    messageInstance = app.mount('#message-container');
  }
  return messageInstance;
};

// 消息对象定义
const message = {
  info(content: string, duration: number = 3) {
    return getInstance().addMessage('info', content, duration);
  },
  success(content: string, duration: number = 3) {
    return getInstance().addMessage('success', content, duration);
  },
  warning(content: string, duration: number = 3) {
    return getInstance().addMessage('warning', content, duration);
  },
  error(content: string, duration: number = 3) {
    return getInstance().addMessage('error', content, duration);
  }
};

// Vue插件
export default {
  install(app: App) {
    // 添加全局属性
    app.config.globalProperties.$message = message;
    
    // 添加全局变量，方便在setup中使用
    app.provide('$message', message);
  }
};

// 导出消息函数
export { message }; 