<script setup lang="ts">
import { inject, getCurrentInstance } from 'vue';
import { message } from '../../utils/message';

// 方法1: 通过message直接调用
const showInfoDirectly = () => {
  message.info('这是一条信息', 3);
};

const showSuccessDirectly = () => {
  message.success('操作成功', 3);
};

const showWarningDirectly = () => {
  message.warning('注意这个操作', 3);
};

const showErrorDirectly = () => {
  message.error('操作失败', 3);
};

const showPersistentMessage = () => {
  message.warning('这条消息需要手动关闭', -1);
};

// 方法2: 通过inject使用
const injectedMessage = inject('$message') as typeof message;

const showInfoInjected = () => {
  injectedMessage.info('这是通过inject调用的信息');
};

// 方法3: 通过实例属性使用
const instance = getCurrentInstance();
const showInfoInstance = () => {
  if (instance) {
    instance.proxy?.$message.info('这是通过实例调用的信息');
  }
};
</script>

<template>
  <div class="p-4 space-y-4">
    <h2 class="text-lg font-medium">消息通知示例</h2>
    
    <div class="space-x-2">
      <button class="btn btn-primary" @click="showInfoDirectly">显示信息</button>
      <button class="btn btn-success" @click="showSuccessDirectly">显示成功</button>
      <button class="btn btn-warning" @click="showWarningDirectly">显示警告</button>
      <button class="btn btn-danger" @click="showErrorDirectly">显示错误</button>
      <button class="btn btn-secondary" @click="showPersistentMessage">显示持续消息</button>
    </div>
    
    <div class="space-x-2 mt-4">
      <button class="btn btn-outline" @click="showInfoInjected">通过inject调用</button>
      <button class="btn btn-outline" @click="showInfoInstance">通过实例调用</button>
    </div>
  </div>
</template> 