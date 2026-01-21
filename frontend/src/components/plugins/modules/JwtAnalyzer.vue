<script setup lang="ts">
import { ref, computed, watch, onMounted, onBeforeUnmount } from 'vue';
import { useI18n } from 'vue-i18n';
import { useDebounceFn } from '@vueuse/core';
// @ts-ignore
import { GetConfig, Verify, Brute, FileSelection, Sign } from "../../../../bindings/github.com/yhy0/ChYing/app.js";
// @ts-ignore
import { Events } from "@wailsio/runtime";

const { t } = useI18n();

// JWT 解析器状态
const jwtInput = ref('');
const selectedAlgorithm = ref('HS256');
const secretKey = ref('');
const secretKeyPath = ref('');
const verificationResult = ref<string | null>(null);
const isVerifying = ref(false);
const showTooltip = ref<string | null>(null);

// 爆破进度相关
const bruteProgress = ref(0);
const showProgress = ref(false);

// 编辑状态
const headerJson = ref('');
const payloadJson = ref('');
const jsonError = {
  header: ref<string | undefined>(undefined),
  payload: ref<string | undefined>(undefined)
};

// 防止循环更新的标志
const isUpdatingFromJwt = ref(false);
const isUpdatingFromEditor = ref(false);

// 算法选项
const algorithms = [
  { value: 'HS256', label: 'HS256' },
  { value: 'HS384', label: 'HS384' },
  { value: 'HS512', label: 'HS512' },
  { value: 'RS256', label: 'RS256' },
  { value: 'RS384', label: 'RS384' },
  { value: 'RS512', label: 'RS512' },
  { value: 'ES256', label: 'ES256' },
  { value: 'ES384', label: 'ES384' },
  { value: 'ES512', label: 'ES512' },
];

onMounted(() => {
  // 获取默认的爆破字典文件路径
  GetConfig().then((result: any) => {
    console.log('获取配置:', result);
    try {
      // 解析JSON
      const config = JSON.parse(result);
      
      // 提取jwt_file字段
      secretKeyPath.value = config.jwt_file;
    } catch (error) {
      console.error('解析配置失败:', error);
    }
  }).catch((err: any) => {
    console.error('获取JWTBruteFile失败:', err);
  });

  // 监听爆破进度
  // Wails v3: percentage 是 WailsEvent 对象，percentage.data 是后端发送的 float64 值
  Events.On("Percentage", (percentage: any) => {
    if (percentage && percentage.data !== undefined) {
      bruteProgress.value = percentage.data;
      console.log('爆破进度:', bruteProgress.value);
    }
  });
});

// 组件卸载前清理事件监听
onBeforeUnmount(() => {
  Events.Off("Percentage");
});

// JWT 解析计算属性
const jwtParts = computed(() => {
  if (!jwtInput.value) return { header: null, payload: null, signature: null };

  try {
    const parts = jwtInput.value.split('.');
    if (parts.length !== 3) return { header: null, payload: null, signature: null };

    const header = JSON.parse(atob(parts[0]));
    const payload = JSON.parse(atob(parts[1]));

    return {
      header,
      payload,
      signature: parts[2]
    };
  } catch (e) {
    return { header: null, payload: null, signature: null };
  }
});

// 监听JWT解析结果变化，更新JSON编辑区域
watch(() => jwtParts.value.header, (newHeader) => {
  if (newHeader && !isUpdatingFromEditor.value) {
    isUpdatingFromJwt.value = true;
    headerJson.value = JSON.stringify(newHeader, null, 2);
    isUpdatingFromJwt.value = false;
  }
}, { immediate: true });

watch(() => jwtParts.value.payload, (newPayload) => {
  if (newPayload && !isUpdatingFromEditor.value) {
    isUpdatingFromJwt.value = true;
    payloadJson.value = JSON.stringify(newPayload, null, 2);
    isUpdatingFromJwt.value = false;
  }
}, { immediate: true });

// 解码JWT
const decodeJwt = () => {
  // 基本的解码已经在计算属性中完成了
  // 这里可以添加更多逻辑，如提示解码成功等
  console.log('JWT解码完成');
};

// 重新签名 JWT（当 header、payload 或 secret 变化时调用）
const resignJwtImpl = async () => {
  if (!headerJson.value || !payloadJson.value) return;

  // 验证 JSON 格式
  try {
    JSON.parse(headerJson.value);
    JSON.parse(payloadJson.value);
  } catch (e) {
    return; // JSON 格式错误，不进行签名
  }

  // 如果没有 secret，只更新 header 和 payload 部分（不重新计算签名）
  if (!secretKey.value) {
    regenerateJwtWithoutSign();
    return;
  }

  // 只支持 HS 系列算法的签名
  if (!selectedAlgorithm.value.startsWith('HS')) {
    regenerateJwtWithoutSign();
    return;
  }

  try {
    isUpdatingFromEditor.value = true;
    const result = await Sign(headerJson.value, payloadJson.value, secretKey.value, selectedAlgorithm.value);
    if (result.error === "") {
      jwtInput.value = result.data;
    }
    isUpdatingFromEditor.value = false;
  } catch (error) {
    console.error('签名失败:', error);
    isUpdatingFromEditor.value = false;
  }
};

// 使用防抖包装，避免频繁调用后端
const resignJwt = useDebounceFn(resignJwtImpl, 300);

// 不重新计算签名，只更新 header 和 payload
const regenerateJwtWithoutSign = () => {
  try {
    const header = JSON.parse(headerJson.value);
    const payload = JSON.parse(payloadJson.value);

    const headerBase64 = btoa(JSON.stringify(header)).replace(/=/g, '').replace(/\+/g, '-').replace(/\//g, '_');
    const payloadBase64 = btoa(JSON.stringify(payload)).replace(/=/g, '').replace(/\+/g, '-').replace(/\//g, '_');

    const parts = jwtInput.value.split('.');
    const signature = parts.length === 3 ? parts[2] : '';

    isUpdatingFromEditor.value = true;
    jwtInput.value = `${headerBase64}.${payloadBase64}.${signature}`;
    isUpdatingFromEditor.value = false;
  } catch (e) {
    console.error('更新JWT失败:', e);
  }
};

// 监听编辑区域变化，尝试更新JWT
watch(headerJson, (newValue) => {
  if (isUpdatingFromJwt.value) return;

  try {
    if (!newValue) return;
    JSON.parse(newValue);
    jsonError.header.value = undefined;
    resignJwt();
  } catch (e) {
    jsonError.header.value = t('modules.plugins.jwt_analyzer.invalid_json', '无效的JSON格式');
  }
}, { deep: true });

watch(payloadJson, (newValue) => {
  if (isUpdatingFromJwt.value) return;

  try {
    if (!newValue) return;
    JSON.parse(newValue);
    jsonError.payload.value = undefined;
    resignJwt();
  } catch (e) {
    jsonError.payload.value = t('modules.plugins.jwt_analyzer.invalid_json', '无效的JSON格式');
  }
}, { deep: true });

// 监听 secret 变化，重新签名
watch(secretKey, () => {
  if (headerJson.value && payloadJson.value) {
    resignJwt();
  }
});

// 监听算法变化，重新签名
watch(selectedAlgorithm, () => {
  if (headerJson.value && payloadJson.value) {
    resignJwt();
  }
});

// 从编辑区域生成JWT（手动触发签名）
const encodeJwt = () => {
  resignJwtImpl();
};

// 验证结果类型
type VerificationStatus = 'success' | 'error' | 'warning' | null;
const verificationStatus = ref<VerificationStatus>(null);

// 验证JWT签名
const verifyJwt = () => {
  if (!jwtInput.value || !secretKey.value) {
    verificationStatus.value = 'warning';
    verificationResult.value = t('modules.plugins.jwt_analyzer.enter_jwt_key');
    return;
  }

  isVerifying.value = true;

  try {
    Verify(jwtInput.value, secretKey.value).then((result: any) => {
      if (result.error !== "") {
        verificationStatus.value = 'error';
        verificationResult.value = t('modules.plugins.jwt_analyzer.verify_error');
        isVerifying.value = false;
        return
      }
      isVerifying.value = false;
      verificationStatus.value = 'success';
      verificationResult.value = t('modules.plugins.jwt_analyzer.verify_success');
    })
  } catch (error) {
    verificationStatus.value = 'error';
    verificationResult.value = t('modules.plugins.jwt_analyzer.verify_error');
    isVerifying.value = false;
  }
};

// 爆破JWT签名
const bruteJwt = () => {
  if (!jwtInput.value) {
    return;
  }
  isVerifying.value = true;
  showProgress.value = true;
  bruteProgress.value = 0;
  
  Brute(jwtInput.value, secretKeyPath.value).then((result: any) => {
    showProgress.value = false;
    if (result !== "") {
      secretKey.value = result
      Verify(jwtInput.value, secretKey.value).then((result: any) => {
        if (result.error !== "") {
          verificationStatus.value = 'error';
          verificationResult.value = t('modules.plugins.jwt_analyzer.verify_error');
          isVerifying.value = false;
          return
        }
        isVerifying.value = false;
        verificationStatus.value = 'success';
        verificationResult.value = t('modules.plugins.jwt_analyzer.verify_success');
      })
    } else {
      verificationStatus.value = 'error';
      verificationResult.value = t('modules.plugins.jwt_analyzer.brute_error');
      isVerifying.value = false;
    }
  }).catch((err: any) => {
    console.error('爆破失败:', err);
    showProgress.value = false;
    isVerifying.value = false;
    verificationStatus.value = 'error';
    verificationResult.value = t('modules.plugins.jwt_analyzer.brute_error');
  });
};

// 停止验证
const stopVerification = () => {
  isVerifying.value = false;
  showProgress.value = false;
  verificationStatus.value = 'warning';
  verificationResult.value = t('modules.plugins.jwt_analyzer.verify_cancelled');
};

// 复制结果
const copyResult = () => {
  if (!jwtParts.value.header && !jwtParts.value.payload) return;

  const resultText = JSON.stringify({
    header: jwtParts.value.header,
    payload: jwtParts.value.payload,
    signature: jwtParts.value.signature
  }, null, 2);

  navigator.clipboard.writeText(resultText)
    .then(() => console.log('JWT解析结果已复制到剪贴板'))
    .catch(err => console.error('复制失败:', err));
};

// 浏览文件按钮处理
const handleBrowse = () => {
  try {
    // 调用后端的文件选择对话框
    FileSelection().then((result: any) => {
      secretKeyPath.value = result;
      console.log('选择的文件路径:', result);
    }).catch((err: any) => {
      console.error('打开文件对话框失败:', err);
    });
  } catch (error) {
    console.error('尝试打开文件选择器时出错:', error);
  }
};

// 显示tooltip
const toggleTooltip = (id: string | null) => {
  showTooltip.value = id;
};
</script>

<template>
  <div class="flex flex-col lg:flex-row gap-2">
    <!-- 左侧 JWT 输入区域 -->
    <div
      class="lg:w-[40%] bg-white dark:bg-[#1e1e2e] border border-gray-200 dark:border-gray-700 rounded-lg p-5 shadow-sm">
      <div class="mb-5">
        <h3 class="text-base font-medium text-gray-800 dark:text-gray-200 mb-3 flex items-center">
          <i class="bx bx-code-alt text-blue-500 mr-2 text-xl"></i>
          {{ t('modules.plugins.jwt_analyzer.jwt_token') }}
        </h3>

        <textarea spellcheck="false" v-model="jwtInput"
          class="w-full h-32 p-3 border border-gray-200 dark:border-gray-700 rounded-lg bg-white dark:bg-[#1e1e2e] focus:outline-none focus:ring-2 focus:ring-indigo-400 transition-all duration-300 text-sm font-mono shadow-inner resize-none"
          :placeholder="t('modules.plugins.jwt_analyzer.paste_token')"></textarea>
      </div>

      <!-- 验证签名区域 -->
      <div class="mb-4">
        <h3 class="text-base font-medium text-gray-800 dark:text-gray-200 mb-3 flex items-center">
          <i class="bx bx-check-shield text-purple-500 mr-2 text-xl"></i>
          {{ t('modules.plugins.jwt_analyzer.verify') }}
        </h3>

        <div class="border border-gray-200 dark:border-gray-700 rounded-lg p-4 bg-white dark:bg-[#1e1e2e]">
          <p class="text-sm text-gray-600 dark:text-gray-400 mb-3">
            {{ t('modules.plugins.jwt_analyzer.verification_description') }}
          </p>

          <textarea spellcheck="false" v-model="secretKey"
            class="w-full h-20 p-3 border border-gray-200 dark:border-gray-700 rounded-lg bg-white dark:bg-[#282838] focus:outline-none focus:ring-2 focus:ring-indigo-400 transition-all duration-300 text-sm font-mono shadow-inner resize-none mb-3"
            :placeholder="t('modules.plugins.jwt_analyzer.enter_secret')"></textarea>

          <div class="flex items-center gap-2 flex-1 min-w-[220px] mb-4">
            <input v-model="secretKeyPath" type="text"
              class="flex-1 border border-gray-200 dark:border-gray-700 rounded-lg bg-white dark:bg-[#282838] px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-indigo-400 transition-all duration-300 shadow-inner"
              :placeholder="t('modules.plugins.jwt_analyzer.key_file_path')" />
            <button @click="handleBrowse" class="btn btn-secondary">
              {{ t('common.actions.browse') }}
            </button>
          </div>

          <!-- 爆破进度条 -->
          <div v-if="showProgress" class="mb-4">
            <div class="flex items-center justify-between mb-1">
              <span class="text-xs font-medium text-gray-600 dark:text-gray-400">爆破进度</span>
              <span class="text-xs font-medium text-gray-600 dark:text-gray-400">{{ Math.round(bruteProgress) }}%</span>
            </div>
            <div class="w-full bg-gray-200 dark:bg-gray-700 rounded-full h-2.5">
              <div class="bg-indigo-600 dark:bg-indigo-500 h-2.5 rounded-full transition-all duration-300" 
                   :style="{ width: `${bruteProgress}%` }"></div>
            </div>
          </div>

          <div class="flex flex-wrap gap-3">
            <button v-if="!isVerifying" @click="bruteJwt" class="btn btn-primary" :disabled="!jwtInput"
              :class="{ 'opacity-50 cursor-not-allowed': !jwtInput }">
              <i class="bx bx-lock-alt mr-1.5"></i> {{ t('modules.plugins.jwt_analyzer.verify_jwt') }}
            </button>
            <button v-else @click="stopVerification" class="btn btn-danger">
              <i class="bx bx-stop mr-1.5"></i> {{ t('modules.plugins.jwt_analyzer.stop') }}
            </button>

            <button @click="copyResult" class="btn btn-secondary" :disabled="!jwtParts.header && !jwtParts.payload"
              :class="{ 'opacity-50 cursor-not-allowed': !jwtParts.header && !jwtParts.payload }">
              <i class="bx bx-copy mr-1.5"></i> {{ t('common.actions.copy') }}
            </button>
          </div>

          <div v-if="verificationResult" class="mt-4 p-3 rounded-lg text-sm transition-all duration-300 flex items-center gap-2" :class="{
            'bg-green-100/80 dark:bg-green-900/30 text-green-800 dark:text-green-200 border border-green-200 dark:border-green-800': verificationStatus === 'success',
            'bg-red-100/80 dark:bg-red-900/30 text-red-800 dark:text-red-200 border border-red-200 dark:border-red-800': verificationStatus === 'error',
            'bg-amber-100/80 dark:bg-amber-900/30 text-amber-800 dark:text-amber-200 border border-amber-200 dark:border-amber-800': verificationStatus === 'warning'
          }">
            <i v-if="verificationStatus === 'success'" class="bx bx-check-circle text-lg text-green-600 dark:text-green-400"></i>
            <i v-else-if="verificationStatus === 'error'" class="bx bx-x-circle text-lg text-red-600 dark:text-red-400"></i>
            <i v-else-if="verificationStatus === 'warning'" class="bx bx-error text-lg text-amber-600 dark:text-amber-400"></i>
            <span>{{ verificationResult }}</span>
          </div>
        </div>
      </div>
    </div>

    <!-- 中间控制面板 -->
    <div class="lg:w-[15%] flex flex-col justify-start items-center">
      <div
        class="bg-white dark:bg-[#1e1e2e] border border-gray-200 dark:border-gray-700 rounded-lg p-4 shadow-sm w-full">
        <h4 class="text-sm font-medium text-gray-700 dark:text-gray-300 mb-4 text-center">{{
          t('modules.plugins.jwt_analyzer.sign_algorithm', '签名算法') }}</h4>

        <div class="relative w-full mb-6">
          <select v-model="selectedAlgorithm"
            class="w-full appearance-none border border-gray-200 dark:border-gray-700 rounded-lg bg-white dark:bg-[#282838] px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-indigo-400 transition-all duration-300 shadow-inner">
            <option v-for="algo in algorithms" :key="algo.value" :value="algo.value">
              {{ algo.label }}
            </option>
          </select>
          <div class="absolute inset-y-0 right-0 flex items-center px-2 pointer-events-none">
            <i class="bx bx-chevron-down text-gray-500"></i>
          </div>
        </div>

        <div class="flex flex-col gap-3 items-center">
          <button @click="decodeJwt" class="w-full btn btn-success">
            <i class="bx bx-code mr-1.5"></i> {{ t('modules.plugins.jwt_analyzer.decode') }} =>
          </button>

          <button @click="encodeJwt" class="w-full btn btn-primary">
            <i class="bx bx-lock mr-1.5"></i>
            <= {{ t('modules.plugins.jwt_analyzer.encode') }} </button>

              <button v-if="!isVerifying" @click="verifyJwt" class="w-full btn btn-warning" :disabled="!jwtInput || !secretKey"
              :class="{ 'opacity-50 cursor-not-allowed': !jwtInput || !secretKey }">
                <i class="bx bx-check-shield mr-1.5"></i> {{ t('modules.plugins.jwt_analyzer.verify_short') }}
              </button>
        </div>
      </div>
    </div>

    <!-- 右侧 JWT 解析结果显示区域 -->
    <div class="lg:w-[45%] flex flex-col space-y-5">
      <!-- Header 部分 -->
      <div class="bg-white dark:bg-[#1e1e2e] border border-gray-200 dark:border-gray-700 rounded-lg p-5 shadow-sm">
        <h3 class="text-base font-medium text-gray-800 dark:text-gray-200 mb-3 flex items-center justify-between">
          <div class="flex items-center">
            <i class="bx bx-code-curly text-blue-500 mr-2 text-xl"></i>
            {{ t('modules.plugins.jwt_analyzer.header') }}
            <div class="relative ml-2">
              <button @mouseenter="toggleTooltip('header')" @mouseleave="toggleTooltip(null)" type="button"
                class="inline-flex items-center justify-center w-5 h-5 text-gray-400 hover:text-gray-600 dark:hover:text-gray-300 rounded-full border border-gray-400 hover:border-gray-600 dark:hover:border-gray-300 focus:outline-none transition-colors duration-300">
                <i class="bx bx-question-mark text-sm"></i>
              </button>
              <div v-if="showTooltip === 'header'"
                class="absolute left-0 top-full mt-2 w-64 bg-white dark:bg-[#282838] shadow-lg rounded-lg p-3 text-xs text-gray-600 dark:text-gray-300 border border-gray-200 dark:border-gray-700 z-10 transform origin-top transition-all duration-300">
                {{ t('modules.plugins.jwt_analyzer.header_tooltip') }}
                <br>{{ t('modules.plugins.jwt_analyzer.header_alg') }}
                <br>{{ t('modules.plugins.jwt_analyzer.header_typ') }}
              </div>
            </div>
          </div>
        </h3>

        <div class="border border-gray-200 dark:border-gray-700 rounded-lg p-4 bg-white dark:bg-[#1e1e2e]">
          <div v-if="jwtParts.header" class="bg-white dark:bg-[#1e1e2e]">
            <textarea spellcheck="false" v-model="headerJson"
              class="w-full h-48 p-3 border border-gray-200 dark:border-gray-700 rounded-lg bg-white dark:bg-[#282838] focus:outline-none focus:ring-2 focus:ring-indigo-400 transition-all duration-300 text-sm font-mono shadow-inner resize-none"
              :class="{ 'border-red-500 dark:border-red-500 focus:ring-red-400': jsonError.header.value }"></textarea>
            <p v-if="jsonError.header.value" class="mt-1 text-xs text-red-500">{{ jsonError.header.value }}</p>
          </div>
          <div v-else class="text-sm text-gray-500 dark:text-gray-400 italic flex items-center justify-center py-6">
            <i class="bx bx-info-circle mr-2 text-lg"></i>
            {{ t('modules.plugins.jwt_analyzer.no_header_data') }}
          </div>
        </div>
      </div>

      <!-- Payload 部分 -->
      <div class="bg-white dark:bg-[#1e1e2e] border border-gray-200 dark:border-gray-700 rounded-lg p-5 shadow-sm">
        <h3 class="text-base font-medium text-gray-800 dark:text-gray-200 mb-3 flex items-center justify-between">
          <div class="flex items-center">
            <i class="bx bx-box text-purple-500 mr-2 text-xl"></i>
            {{ t('modules.plugins.jwt_analyzer.payload') }}
            <div class="relative ml-2">
              <button @mouseenter="toggleTooltip('payload')" @mouseleave="toggleTooltip(null)" type="button"
                class="inline-flex items-center justify-center w-5 h-5 text-gray-400 hover:text-gray-600 dark:hover:text-gray-300 rounded-full border border-gray-400 hover:border-gray-600 dark:hover:border-gray-300 focus:outline-none transition-colors duration-300">
                <i class="bx bx-question-mark text-sm"></i>
              </button>
              <div v-if="showTooltip === 'payload'"
                class="absolute left-0 top-full mt-2 w-64 bg-white dark:bg-[#282838] shadow-lg rounded-lg p-3 text-xs text-gray-600 dark:text-gray-300 border border-gray-200 dark:border-gray-700 z-10 transform origin-top transition-all duration-300">
                {{ t('modules.plugins.jwt_analyzer.payload_tooltip') }}
                <br>{{ t('modules.plugins.jwt_analyzer.payload_iss') }}
                <br>{{ t('modules.plugins.jwt_analyzer.payload_sub') }}
                <br>{{ t('modules.plugins.jwt_analyzer.payload_aud') }}
                <br>{{ t('modules.plugins.jwt_analyzer.payload_exp') }}
                <br>{{ t('modules.plugins.jwt_analyzer.payload_nbf') }}
                <br>{{ t('modules.plugins.jwt_analyzer.payload_iat') }}
                <br>{{ t('modules.plugins.jwt_analyzer.payload_jti') }}
              </div>
            </div>
          </div>
        </h3>

        <div class="border border-gray-200 dark:border-gray-700 rounded-lg p-4 bg-white dark:bg-[#1e1e2e]">
          <div v-if="jwtParts.payload" class="bg-white dark:bg-[#1e1e2e]">
            <textarea spellcheck="false" v-model="payloadJson"
              class="w-full h-48 p-3 border border-gray-200 dark:border-gray-700 rounded-lg bg-white dark:bg-[#282838] focus:outline-none focus:ring-2 focus:ring-indigo-400 transition-all duration-300 text-sm font-mono shadow-inner resize-none"
              :class="{ 'border-red-500 dark:border-red-500 focus:ring-red-400': jsonError.payload.value }"></textarea>
            <p v-if="jsonError.payload.value" class="mt-1 text-xs text-red-500">{{ jsonError.payload.value }}</p>
          </div>
          <div v-else class="text-sm text-gray-500 dark:text-gray-400 italic flex items-center justify-center py-6">
            <i class="bx bx-info-circle mr-2 text-lg"></i>
            {{ t('modules.plugins.jwt_analyzer.no_payload_data') }}
          </div>
        </div>
      </div>

      <!-- Signature 部分 -->
      <div class="bg-white dark:bg-[#1e1e2e] border border-gray-200 dark:border-gray-700 rounded-lg p-5 shadow-sm">
        <h3 class="text-base font-medium text-gray-800 dark:text-gray-200 mb-3 flex items-center justify-between">
          <div class="flex items-center">
            <i class="bx bx-lock text-red-500 mr-2 text-xl"></i>
            {{ t('modules.plugins.jwt_analyzer.signature') }}
            <div class="relative ml-2">
              <button @mouseenter="toggleTooltip('signature')" @mouseleave="toggleTooltip(null)" type="button"
                class="inline-flex items-center justify-center w-5 h-5 text-gray-400 hover:text-gray-600 dark:hover:text-gray-300 rounded-full border border-gray-400 hover:border-gray-600 dark:hover:border-gray-300 focus:outline-none transition-colors duration-300">
                <i class="bx bx-question-mark text-sm"></i>
              </button>
              <div v-if="showTooltip === 'signature'"
                class="absolute left-0 top-full mt-2 w-64 bg-white dark:bg-[#282838] shadow-lg rounded-lg p-3 text-xs text-gray-600 dark:text-gray-300 border border-gray-200 dark:border-gray-700 z-10 transform origin-top transition-all duration-300">
                {{ t('modules.plugins.jwt_analyzer.signature_tooltip') }}
              </div>
            </div>
          </div>
        </h3>

        <div class="border border-gray-200 dark:border-gray-700 rounded-lg p-4 bg-white dark:bg-[#1e1e2e]">
          <div v-if="jwtParts.signature" class="bg-white dark:bg-[#1e1e2e]">
            <code class="text-sm text-gray-800 dark:text-gray-200 font-mono break-all">{{ jwtParts.signature }}</code>
            <p class="mt-3 text-xs text-gray-500 dark:text-gray-400">
              <i class="bx bx-info-circle mr-1"></i> {{ t('modules.plugins.jwt_analyzer.signature_hint') }}
            </p>
          </div>
          <div v-else class="text-sm text-gray-500 dark:text-gray-400 italic flex items-center justify-center py-6">
            <i class="bx bx-info-circle mr-2 text-lg"></i>
            {{ t('modules.plugins.jwt_analyzer.no_signature_data') }}
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
pre {
  white-space: pre-wrap;
  word-break: break-word;
}

/* 添加过渡动画效果 */
.fade-enter-active,
.fade-leave-active {
  transition: opacity 0.3s ease;
}

.fade-enter-from,
.fade-leave-to {
  opacity: 0;
}

/* 滚动条样式已移至 scrollbar.css 统一管理 */
</style>
