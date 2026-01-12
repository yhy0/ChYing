<script setup lang="ts">
import { useI18n } from 'vue-i18n';
import { type CompleteConfig } from '../../../types';

defineProps<{
  completeConfig: CompleteConfig | null;
}>();

const emit = defineEmits<{
  (e: 'updatePluginConfig'): void;
}>();

const { t } = useI18n();
</script>

<template>
  <div v-if="completeConfig && completeConfig.plugins" class="settings-container">
    <div class="flex items-center justify-between mb-4">
      <h3 class="text-lg font-medium">{{ t('modules.plugins.title') }}</h3>
      <button 
        @click="emit('updatePluginConfig')"
        class="px-3 py-1.5 bg-amber-500 hover:bg-amber-600 text-white rounded-md text-sm font-medium transition-colors"
      >
        {{ t('common.actions.save') }}
      </button>
    </div>
    
    <div class="space-y-6">
      <!-- 暴力破解模块设置 -->
      <div v-if="completeConfig.plugins.bruteForce" class="settings-section">
        <h4 class="settings-section-title">{{ t('modules.plugins.bruteforce.title') }}</h4>

        <div class="settings-grid">
          <div class="settings-item">
            <input
              id="bruteforce-web"
              v-model="completeConfig.plugins.bruteForce.web"
              type="checkbox"
            />
            <label for="bruteforce-web">
              {{ t('modules.plugins.bruteforce.web') }}
            </label>
          </div>
          
          <div class="settings-item">
            <input
              id="bruteforce-service"
              v-model="completeConfig.plugins.bruteForce.service"
              type="checkbox"
            />
            <label for="bruteforce-service">
              {{ t('modules.plugins.bruteforce.service') }}
            </label>
          </div>
          
          <div>
            <label class="settings-label">{{ t('modules.plugins.bruteforce.username_dict') }}</label>
            <input
              v-model="completeConfig.plugins.bruteForce.usernameDict"
              type="text"
              class="settings-input"
              :placeholder="t('modules.plugins.bruteforce.use_built_in')"
              spellcheck="false"
            />
          </div>

          <div>
            <label class="settings-label">{{ t('modules.plugins.bruteforce.password_dict') }}</label>
            <input
              v-model="completeConfig.plugins.bruteForce.passwordDict"
              type="text"
              class="settings-input"
              :placeholder="t('modules.plugins.bruteforce.use_built_in')"
              spellcheck="false"
            />
          </div>
        </div>
      </div>
      
      <!-- 命令注入模块设置 -->
      <div v-if="completeConfig.plugins.cmdInjection" class="settings-section">
        <div class="settings-item">
          <input
            id="cmdinjection-enabled"
            v-model="completeConfig.plugins.cmdInjection.enabled"
            type="checkbox"
          />
          <label for="cmdinjection-enabled">
            {{ t('modules.plugins.cmdinjection.title') }}
          </label>
        </div>
      </div>
      
      <!-- CRLF注入模块设置 -->
      <div v-if="completeConfig.plugins.crlfInjection" class="settings-section">
        <div class="settings-item">
          <input
            id="crlfinjection-enabled"
            v-model="completeConfig.plugins.crlfInjection.enabled"
            type="checkbox"
          />
          <label for="crlfinjection-enabled">
            {{ t('modules.plugins.crlfinjection.title') }}
          </label>
        </div>
      </div>
      
      <!-- XSS模块设置 -->
      <div v-if="completeConfig.plugins.xss" class="settings-section">
        <h4 class="settings-section-title">{{ t('modules.plugins.xss.title') }}</h4>

        <div class="settings-grid">
          <div class="flex items-center space-x-2">
            <input
              id="xss-enabled"
              v-model="completeConfig.plugins.xss.enabled"
              type="checkbox"
              class="rounded border-gray-300 text-blue-500 focus:ring-blue-500"
            />
            <label for="xss-enabled" class="text-sm font-medium text-gray-700 dark:text-gray-300">
              {{ t('modules.plugins.xss.enable') }}
            </label>
          </div>
          
          <div class="flex items-center space-x-2">
            <input
              id="xss-cookie"
              v-model="completeConfig.plugins.xss.detectXssInCookie"
              type="checkbox"
              class="rounded border-gray-300 text-blue-500 focus:ring-blue-500"
            />
            <label for="xss-cookie" class="text-sm font-medium text-gray-700 dark:text-gray-300">
              {{ t('modules.plugins.xss.detect_in_cookie') }}
            </label>
          </div>
        </div>
      </div>
      
      <!-- SQL注入模块设置 -->
      <div v-if="completeConfig.plugins.sql" class="p-4 border border-gray-200 dark:border-gray-700 rounded-lg">
        <div class="flex items-center justify-between mb-3">
          <h4 class="font-medium">{{ t('modules.plugins.sqlinjection.title') }}</h4>
        </div>
        
        <div class="grid grid-cols-1 md:grid-cols-2 gap-4">
          <div class="flex items-center space-x-2">
            <input
              id="sql-enabled"
              v-model="completeConfig.plugins.sql.enabled"
              type="checkbox"
              class="rounded border-gray-300 text-blue-500 focus:ring-blue-500"
            />
            <label for="sql-enabled" class="text-sm font-medium text-gray-700 dark:text-gray-300">
              {{ t('modules.plugins.sqlinjection.enable') }}
            </label>
          </div>
          
          <div class="flex items-center space-x-2">
            <input
              id="sql-boolean"
              v-model="completeConfig.plugins.sql.booleanBasedDetection"
              type="checkbox"
              class="rounded border-gray-300 text-blue-500 focus:ring-blue-500"
            />
            <label for="sql-boolean" class="text-sm font-medium text-gray-700 dark:text-gray-300">
              {{ t('modules.plugins.sqlinjection.boolean_based') }}
            </label>
          </div>
          
          <div class="flex items-center space-x-2">
            <input
              id="sql-error"
              v-model="completeConfig.plugins.sql.errorBasedDetection"
              type="checkbox"
              class="rounded border-gray-300 text-blue-500 focus:ring-blue-500"
            />
            <label for="sql-error" class="text-sm font-medium text-gray-700 dark:text-gray-300">
              {{ t('modules.plugins.sqlinjection.error_based') }}
            </label>
          </div>
          
          <div class="flex items-center space-x-2">
            <input
              id="sql-time"
              v-model="completeConfig.plugins.sql.timeBasedDetection"
              type="checkbox"
              class="rounded border-gray-300 text-blue-500 focus:ring-blue-500"
            />
            <label for="sql-time" class="text-sm font-medium text-gray-700 dark:text-gray-300">
              {{ t('modules.plugins.sqlinjection.time_based') }}
            </label>
          </div>
          
          <div class="flex items-center space-x-2">
            <input
              id="sql-cookie"
              v-model="completeConfig.plugins.sql.detectInCookie"
              type="checkbox"
              class="rounded border-gray-300 text-blue-500 focus:ring-blue-500"
            />
            <label for="sql-cookie" class="text-sm font-medium text-gray-700 dark:text-gray-300">
              {{ t('modules.plugins.sqlinjection.detect_in_cookie') }}
            </label>
          </div>
        </div>
      </div>
      
      <!-- 其他插件开关 -->
      <div class="p-4 border border-gray-200 dark:border-gray-700 rounded-lg">
        <h4 class="font-medium mb-3">{{ t('modules.plugins.other.title') }}</h4>
        
        <div class="grid grid-cols-1 md:grid-cols-3 gap-4">
          <div v-if="completeConfig.plugins.xxe" class="flex items-center space-x-2">
            <input
              id="xxe-enabled"
              v-model="completeConfig.plugins.xxe.enabled"
              type="checkbox"
              class="rounded border-gray-300 text-blue-500 focus:ring-blue-500"
            />
            <label for="xxe-enabled" class="text-sm font-medium text-gray-700 dark:text-gray-300">
              {{ t('modules.plugins.xxe.title') }}
            </label>
          </div>
          
          <div v-if="completeConfig.plugins.ssrf" class="flex items-center space-x-2">
            <input
              id="ssrf-enabled"
              v-model="completeConfig.plugins.ssrf.enabled"
              type="checkbox"
              class="rounded border-gray-300 text-blue-500 focus:ring-blue-500"
            />
            <label for="ssrf-enabled" class="text-sm font-medium text-gray-700 dark:text-gray-300">
              {{ t('modules.plugins.ssrf.title') }}
            </label>
          </div>
          
          <div v-if="completeConfig.plugins.bbscan" class="flex items-center space-x-2">
            <input
              id="bbscan-enabled"
              v-model="completeConfig.plugins.bbscan.enabled"
              type="checkbox"
              class="rounded border-gray-300 text-blue-500 focus:ring-blue-500"
            />
            <label for="bbscan-enabled" class="text-sm font-medium text-gray-700 dark:text-gray-300">
              {{ t('modules.plugins.bbscan.title') }}
            </label>
          </div>
          
          <div v-if="completeConfig.plugins.jsonp" class="flex items-center space-x-2">
            <input
              id="jsonp-enabled"
              v-model="completeConfig.plugins.jsonp.enabled"
              type="checkbox"
              class="rounded border-gray-300 text-blue-500 focus:ring-blue-500"
            />
            <label for="jsonp-enabled" class="text-sm font-medium text-gray-700 dark:text-gray-300">
              {{ t('modules.plugins.jsonp.title') }}
            </label>
          </div>
          
          <div v-if="completeConfig.plugins.log4j" class="flex items-center space-x-2">
            <input
              id="log4j-enabled"
              v-model="completeConfig.plugins.log4j.enabled"
              type="checkbox"
              class="rounded border-gray-300 text-blue-500 focus:ring-blue-500"
            />
            <label for="log4j-enabled" class="text-sm font-medium text-gray-700 dark:text-gray-300">
              {{ t('modules.plugins.log4j.title') }}
            </label>
          </div>
          
          <div v-if="completeConfig.plugins.bypass403" class="flex items-center space-x-2">
            <input
              id="bypass403-enabled"
              v-model="completeConfig.plugins.bypass403.enabled"
              type="checkbox"
              class="rounded border-gray-300 text-blue-500 focus:ring-blue-500"
            />
            <label for="bypass403-enabled" class="text-sm font-medium text-gray-700 dark:text-gray-300">
              {{ t('modules.plugins.bypass403.title') }}
            </label>
          </div>
          
          <div v-if="completeConfig.plugins.fastjson" class="flex items-center space-x-2">
            <input
              id="fastjson-enabled"
              v-model="completeConfig.plugins.fastjson.enabled"
              type="checkbox"
              class="rounded border-gray-300 text-blue-500 focus:ring-blue-500"
            />
            <label for="fastjson-enabled" class="text-sm font-medium text-gray-700 dark:text-gray-300">
              {{ t('modules.plugins.fastjson.title') }}
            </label>
          </div>
          
          <div v-if="completeConfig.plugins.archive" class="flex items-center space-x-2">
            <input
              id="archive-enabled"
              v-model="completeConfig.plugins.archive.enabled"
              type="checkbox"
              class="rounded border-gray-300 text-blue-500 focus:ring-blue-500"
            />
            <label for="archive-enabled" class="text-sm font-medium text-gray-700 dark:text-gray-300">
              {{ t('modules.plugins.archive.title') }}
            </label>
          </div>
          
          <div v-if="completeConfig.plugins.iis" class="flex items-center space-x-2">
            <input
              id="iis-enabled"
              v-model="completeConfig.plugins.iis.enabled"
              type="checkbox"
              class="rounded border-gray-300 text-blue-500 focus:ring-blue-500"
            />
            <label for="iis-enabled" class="text-sm font-medium text-gray-700 dark:text-gray-300">
              {{ t('modules.plugins.iis.title') }}
            </label>
          </div>
          
          <div v-if="completeConfig.plugins.nginxAliasTraversal" class="flex items-center space-x-2">
            <input
              id="nginx-alias-enabled"
              v-model="completeConfig.plugins.nginxAliasTraversal.enabled"
              type="checkbox"
              class="rounded border-gray-300 text-blue-500 focus:ring-blue-500"
            />
            <label for="nginx-alias-enabled" class="text-sm font-medium text-gray-700 dark:text-gray-300">
              {{ t('modules.plugins.nginxAlias.title') }}
            </label>
          </div>
          
          <div v-if="completeConfig.plugins.poc" class="flex items-center space-x-2">
            <input
              id="poc-enabled"
              v-model="completeConfig.plugins.poc.enabled"
              type="checkbox"
              class="rounded border-gray-300 text-blue-500 focus:ring-blue-500"
            />
            <label for="poc-enabled" class="text-sm font-medium text-gray-700 dark:text-gray-300">
              {{ t('modules.plugins.poc.title') }}
            </label>
          </div>
          
          <div v-if="completeConfig.plugins.nuclei" class="flex items-center space-x-2">
            <input
              id="nuclei-enabled"
              v-model="completeConfig.plugins.nuclei.enabled"
              type="checkbox"
              class="rounded border-gray-300 text-blue-500 focus:ring-blue-500"
            />
            <label for="nuclei-enabled" class="text-sm font-medium text-gray-700 dark:text-gray-300">
              {{ t('modules.plugins.nuclei.title') }}
            </label>
          </div>
          
          <div v-if="completeConfig.plugins.portScan" class="flex items-center space-x-2">
            <input
              id="port-scan-enabled"
              v-model="completeConfig.plugins.portScan.enabled"
              type="checkbox"
              class="rounded border-gray-300 text-blue-500 focus:ring-blue-500"
            />
            <label for="port-scan-enabled" class="text-sm font-medium text-gray-700 dark:text-gray-300">
              {{ t('modules.plugins.portScan.title') }}
            </label>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
/* 复选框样式已移至 form.css 统一管理 */
</style>