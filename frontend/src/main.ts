import { createApp } from 'vue';
import { createPinia } from 'pinia';
import 'uno.css'; // UnoCSS样式
import './style.css';
import './styles/index.css';

import App from './App.vue';
import { i18n } from './i18n';
import { initTheme } from './theme';
import MessagePlugin from './utils/message';

import router from './router'; // 引入定义好的路由
// Create Vue app
const app = createApp(App);

// 创建Pinia实例
const pinia = createPinia();

// Add plugins
app.use(i18n);
app.use(pinia);
app.use(MessagePlugin); // 注册消息插件

// 初始化主题
initTheme();

app.use(router);  // 注册路由
// Mount app
app.mount('#app');