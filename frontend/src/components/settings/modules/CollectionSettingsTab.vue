<script setup lang="ts">
import { ref } from 'vue';
import { useI18n } from 'vue-i18n';
import { type CompleteConfig } from '../../../types';

const props = defineProps<{
  completeConfig: CompleteConfig | null;
}>();

const emit = defineEmits<{
  (e: 'saveCompleteConfig'): void;
  (e: 'addNewRegex', type: string): void;
  (e: 'openRegexModal', type: string, index: number): void;
  (e: 'deleteRegex', type: string, index: number): void;
}>();

const { t } = useI18n();
</script>

<template>
  <div v-if="completeConfig" class="bg-white dark:bg-[#282838] rounded-lg p-5 shadow-sm border border-gray-100 dark:border-gray-700">
    <div class="flex items-center justify-between mb-4">
      <h3 class="text-lg font-medium">{{ t('settings.collection_settings') }}</h3>
      <button 
        @click="emit('saveCompleteConfig')"
        class="px-3 py-1.5 bg-amber-500 hover:bg-amber-600 text-white rounded-md text-sm font-medium transition-colors"
      >
        {{ t('common.actions.save') }}
      </button>
    </div>
    
    <div class="space-y-6">
      <!-- 域名正则 -->
      <div class="p-4 border border-gray-200 dark:border-gray-700 rounded-lg">
        <div class="flex items-center justify-between mb-3">
          <h4 class="font-medium">{{ t('settings.domain_regex') }}</h4>
          <button 
            @click="emit('addNewRegex', 'domain')"
            class="flex items-center px-2 py-1 bg-green-500 hover:bg-green-600 text-white rounded-md text-xs font-medium transition-colors"
          >
            <i class="bx bx-plus mr-1"></i>
            {{ t('settings.add_regex') }}
          </button>
        </div>
        
        <div class="space-y-2 max-h-48 overflow-auto">
          <div 
            v-for="(regex, index) in completeConfig.collection.domain" 
            :key="`domain-${index}`"
            class="flex items-center justify-between bg-gray-50 dark:bg-gray-700 p-2 rounded"
          >
            <div class="text-sm text-gray-800 dark:text-gray-200 truncate max-w-xl">
              {{ regex }}
            </div>
            <div class="flex space-x-2">
              <button 
                @click="emit('openRegexModal', 'domain', index)"
                class="text-blue-500 hover:text-blue-700 focus:outline-none"
              >
                <i class="bx bx-edit text-lg"></i>
              </button>
              <button 
                @click="emit('deleteRegex', 'domain', index)"
                class="text-red-500 hover:text-red-700 focus:outline-none"
              >
                <i class="bx bx-trash text-lg"></i>
              </button>
            </div>
          </div>
        </div>
      </div>
      
      <!-- IP正则 -->
      <div class="p-4 border border-gray-200 dark:border-gray-700 rounded-lg">
        <div class="flex items-center justify-between mb-3">
          <h4 class="font-medium">{{ t('settings.ip_regex') }}</h4>
          <button 
            @click="emit('addNewRegex', 'ip')"
            class="flex items-center px-2 py-1 bg-green-500 hover:bg-green-600 text-white rounded-md text-xs font-medium transition-colors"
          >
            <i class="bx bx-plus mr-1"></i>
            {{ t('settings.add_regex') }}
          </button>
        </div>
        
        <div class="space-y-2 max-h-48 overflow-auto">
          <div 
            v-for="(regex, index) in completeConfig.collection.ip" 
            :key="`ip-${index}`"
            class="flex items-center justify-between bg-gray-50 dark:bg-gray-700 p-2 rounded"
          >
            <div class="text-sm text-gray-800 dark:text-gray-200 truncate max-w-xl">
              {{ regex }}
            </div>
            <div class="flex space-x-2">
              <button 
                @click="emit('openRegexModal', 'ip', index)"
                class="text-blue-500 hover:text-blue-700 focus:outline-none"
              >
                <i class="bx bx-edit text-lg"></i>
              </button>
              <button 
                @click="emit('deleteRegex', 'ip', index)"
                class="text-red-500 hover:text-red-700 focus:outline-none"
              >
                <i class="bx bx-trash text-lg"></i>
              </button>
            </div>
          </div>
        </div>
      </div>
      
      <!-- 手机号正则 -->
      <div class="p-4 border border-gray-200 dark:border-gray-700 rounded-lg">
        <div class="flex items-center justify-between mb-3">
          <h4 class="font-medium">{{ t('settings.phone_regex') }}</h4>
          <button 
            @click="emit('addNewRegex', 'phone')"
            class="flex items-center px-2 py-1 bg-green-500 hover:bg-green-600 text-white rounded-md text-xs font-medium transition-colors"
          >
            <i class="bx bx-plus mr-1"></i>
            {{ t('settings.add_regex') }}
          </button>
        </div>
        
        <div class="space-y-2 max-h-48 overflow-auto">
          <div 
            v-for="(regex, index) in completeConfig.collection.phone" 
            :key="`phone-${index}`"
            class="flex items-center justify-between bg-gray-50 dark:bg-gray-700 p-2 rounded"
          >
            <div class="text-sm text-gray-800 dark:text-gray-200 truncate max-w-xl">
              {{ regex }}
            </div>
            <div class="flex space-x-2">
              <button 
                @click="emit('openRegexModal', 'phone', index)"
                class="text-blue-500 hover:text-blue-700 focus:outline-none"
              >
                <i class="bx bx-edit text-lg"></i>
              </button>
              <button 
                @click="emit('deleteRegex', 'phone', index)"
                class="text-red-500 hover:text-red-700 focus:outline-none"
              >
                <i class="bx bx-trash text-lg"></i>
              </button>
            </div>
          </div>
        </div>
      </div>
      
      <!-- 邮箱正则 -->
      <div class="p-4 border border-gray-200 dark:border-gray-700 rounded-lg">
        <div class="flex items-center justify-between mb-3">
          <h4 class="font-medium">{{ t('settings.email_regex') }}</h4>
          <button 
            @click="emit('addNewRegex', 'email')"
            class="flex items-center px-2 py-1 bg-green-500 hover:bg-green-600 text-white rounded-md text-xs font-medium transition-colors"
          >
            <i class="bx bx-plus mr-1"></i>
            {{ t('settings.add_regex') }}
          </button>
        </div>
        
        <div class="space-y-2 max-h-48 overflow-auto">
          <div 
            v-for="(regex, index) in completeConfig.collection.email" 
            :key="`email-${index}`"
            class="flex items-center justify-between bg-gray-50 dark:bg-gray-700 p-2 rounded"
          >
            <div class="text-sm text-gray-800 dark:text-gray-200 truncate max-w-xl">
              {{ regex }}
            </div>
            <div class="flex space-x-2">
              <button 
                @click="emit('openRegexModal', 'email', index)"
                class="text-blue-500 hover:text-blue-700 focus:outline-none"
              >
                <i class="bx bx-edit text-lg"></i>
              </button>
              <button 
                @click="emit('deleteRegex', 'email', index)"
                class="text-red-500 hover:text-red-700 focus:outline-none"
              >
                <i class="bx bx-trash text-lg"></i>
              </button>
            </div>
          </div>
        </div>
      </div>
      
      <!-- API正则 -->
      <div class="p-4 border border-gray-200 dark:border-gray-700 rounded-lg">
        <div class="flex items-center justify-between mb-3">
          <h4 class="font-medium">{{ t('settings.api_regex') }}</h4>
          <button 
            @click="emit('addNewRegex', 'api')"
            class="flex items-center px-2 py-1 bg-green-500 hover:bg-green-600 text-white rounded-md text-xs font-medium transition-colors"
          >
            <i class="bx bx-plus mr-1"></i>
            {{ t('settings.add_regex') }}
          </button>
        </div>
        
        <div class="space-y-2 max-h-48 overflow-auto">
          <div 
            v-for="(regex, index) in completeConfig.collection.api" 
            :key="`api-${index}`"
            class="flex items-center justify-between bg-gray-50 dark:bg-gray-700 p-2 rounded"
          >
            <div class="text-sm text-gray-800 dark:text-gray-200 truncate max-w-xl">
              {{ regex }}
            </div>
            <div class="flex space-x-2">
              <button 
                @click="emit('openRegexModal', 'api', index)"
                class="text-blue-500 hover:text-blue-700 focus:outline-none"
              >
                <i class="bx bx-edit text-lg"></i>
              </button>
              <button 
                @click="emit('deleteRegex', 'api', index)"
                class="text-red-500 hover:text-red-700 focus:outline-none"
              >
                <i class="bx bx-trash text-lg"></i>
              </button>
            </div>
          </div>
        </div>
      </div>
      
      <!-- URL正则 -->
      <div class="p-4 border border-gray-200 dark:border-gray-700 rounded-lg">
        <div class="flex items-center justify-between mb-3">
          <h4 class="font-medium">{{ t('settings.url_regex') }}</h4>
          <button 
            @click="emit('addNewRegex', 'url')"
            class="flex items-center px-2 py-1 bg-green-500 hover:bg-green-600 text-white rounded-md text-xs font-medium transition-colors"
          >
            <i class="bx bx-plus mr-1"></i>
            {{ t('settings.add_regex') }}
          </button>
        </div>
        
        <div class="space-y-2 max-h-48 overflow-auto">
          <div 
            v-for="(regex, index) in completeConfig.collection.url" 
            :key="`url-${index}`"
            class="flex items-center justify-between bg-gray-50 dark:bg-gray-700 p-2 rounded"
          >
            <div class="text-sm text-gray-800 dark:text-gray-200 truncate max-w-xl">
              {{ regex }}
            </div>
            <div class="flex space-x-2">
              <button 
                @click="emit('openRegexModal', 'url', index)"
                class="text-blue-500 hover:text-blue-700 focus:outline-none"
              >
                <i class="bx bx-edit text-lg"></i>
              </button>
              <button 
                @click="emit('deleteRegex', 'url', index)"
                class="text-red-500 hover:text-red-700 focus:outline-none"
              >
                <i class="bx bx-trash text-lg"></i>
              </button>
            </div>
          </div>
        </div>
      </div>
      
      <!-- URL过滤器规则 -->
      <div class="p-4 border border-gray-200 dark:border-gray-700 rounded-lg">
        <div class="flex items-center justify-between mb-3">
          <h4 class="font-medium">{{ t('settings.url_filter_regex') }}</h4>
          <button 
            @click="emit('addNewRegex', 'urlFilter')"
            class="flex items-center px-2 py-1 bg-green-500 hover:bg-green-600 text-white rounded-md text-xs font-medium transition-colors"
          >
            <i class="bx bx-plus mr-1"></i>
            {{ t('settings.add_regex') }}
          </button>
        </div>
        
        <div class="space-y-2 max-h-48 overflow-auto">
          <div 
            v-for="(regex, index) in completeConfig.collection.urlFilter || []" 
            :key="`urlFilter-${index}`"
            class="flex items-center justify-between bg-gray-50 dark:bg-gray-700 p-2 rounded"
          >
            <div class="text-sm text-gray-800 dark:text-gray-200 truncate max-w-xl">
              {{ regex }}
            </div>
            <div class="flex space-x-2">
              <button 
                @click="emit('openRegexModal', 'urlFilter', index)"
                class="text-blue-500 hover:text-blue-700 focus:outline-none"
              >
                <i class="bx bx-edit text-lg"></i>
              </button>
              <button 
                @click="emit('deleteRegex', 'urlFilter', index)"
                class="text-red-500 hover:text-red-700 focus:outline-none"
              >
                <i class="bx bx-trash text-lg"></i>
              </button>
            </div>
          </div>
          <div v-if="!completeConfig.collection.urlFilter || completeConfig.collection.urlFilter.length === 0" 
               class="text-center text-sm text-gray-500 dark:text-gray-400 py-2">
            {{ t('common.status.no_data') }}
          </div>
        </div>
      </div>
      
      <!-- 身份证规则 -->
      <div class="p-4 border border-gray-200 dark:border-gray-700 rounded-lg">
        <div class="flex items-center justify-between mb-3">
          <h4 class="font-medium">{{ t('settings.id_card_regex') }}</h4>
          <button 
            @click="emit('addNewRegex', 'idCard')"
            class="flex items-center px-2 py-1 bg-green-500 hover:bg-green-600 text-white rounded-md text-xs font-medium transition-colors"
          >
            <i class="bx bx-plus mr-1"></i>
            {{ t('settings.add_regex') }}
          </button>
        </div>
        
        <div class="space-y-2 max-h-48 overflow-auto">
          <div 
            v-for="(regex, index) in completeConfig.collection.idCard || []" 
            :key="`idCard-${index}`"
            class="flex items-center justify-between bg-gray-50 dark:bg-gray-700 p-2 rounded"
          >
            <div class="text-sm text-gray-800 dark:text-gray-200 truncate max-w-xl">
              {{ regex }}
            </div>
            <div class="flex space-x-2">
              <button 
                @click="emit('openRegexModal', 'idCard', index)"
                class="text-blue-500 hover:text-blue-700 focus:outline-none"
              >
                <i class="bx bx-edit text-lg"></i>
              </button>
              <button 
                @click="emit('deleteRegex', 'idCard', index)"
                class="text-red-500 hover:text-red-700 focus:outline-none"
              >
                <i class="bx bx-trash text-lg"></i>
              </button>
            </div>
          </div>
          <div v-if="!completeConfig.collection.idCard || completeConfig.collection.idCard.length === 0" 
               class="text-center text-sm text-gray-500 dark:text-gray-400 py-2">
            {{ t('common.status.no_data') }}
          </div>
        </div>
      </div>
      
      <!-- 敏感参数列表 -->
      <div class="p-4 border border-gray-200 dark:border-gray-700 rounded-lg">
        <div class="flex items-center justify-between mb-3">
          <h4 class="font-medium">{{ t('settings.sensitive_parameters') }}</h4>
          <button 
            @click="emit('addNewRegex', 'sensitiveParameters')"
            class="flex items-center px-2 py-1 bg-green-500 hover:bg-green-600 text-white rounded-md text-xs font-medium transition-colors"
          >
            <i class="bx bx-plus mr-1"></i>
            {{ t('settings.add_regex') }}
          </button>
        </div>
        
        <div class="space-y-2 max-h-48 overflow-auto">
          <div 
            v-for="(param, index) in completeConfig.collection.sensitiveParameters || []" 
            :key="`sensitiveParam-${index}`"
            class="flex items-center justify-between bg-gray-50 dark:bg-gray-700 p-2 rounded"
          >
            <div class="text-sm text-gray-800 dark:text-gray-200 truncate max-w-xl">
              {{ param }}
            </div>
            <div class="flex space-x-2">
              <button 
                @click="emit('openRegexModal', 'sensitiveParameters', index)"
                class="text-blue-500 hover:text-blue-700 focus:outline-none"
              >
                <i class="bx bx-edit text-lg"></i>
              </button>
              <button 
                @click="emit('deleteRegex', 'sensitiveParameters', index)"
                class="text-red-500 hover:text-red-700 focus:outline-none"
              >
                <i class="bx bx-trash text-lg"></i>
              </button>
            </div>
          </div>
          <div v-if="!completeConfig.collection.sensitiveParameters || completeConfig.collection.sensitiveParameters.length === 0" 
               class="text-center text-sm text-gray-500 dark:text-gray-400 py-2">
            {{ t('common.status.no_data') }}
          </div>
        </div>
      </div>
      
      <!-- 其他正则 -->
      <div class="p-4 border border-gray-200 dark:border-gray-700 rounded-lg">
        <div class="flex items-center justify-between mb-3">
          <h4 class="font-medium">{{ t('settings.other_regex') }}</h4>
          <button 
            @click="emit('addNewRegex', 'other')"
            class="flex items-center px-2 py-1 bg-green-500 hover:bg-green-600 text-white rounded-md text-xs font-medium transition-colors"
          >
            <i class="bx bx-plus mr-1"></i>
            {{ t('settings.add_regex') }}
          </button>
        </div>
        
        <div class="space-y-2 max-h-48 overflow-auto">
          <div 
            v-for="(regex, index) in completeConfig.collection.other" 
            :key="`other-${index}`"
            class="flex items-center justify-between bg-gray-50 dark:bg-gray-700 p-2 rounded"
          >
            <div class="text-sm text-gray-800 dark:text-gray-200 truncate max-w-xl">
              {{ regex }}
            </div>
            <div class="flex space-x-2">
              <button 
                @click="emit('openRegexModal', 'other', index)"
                class="text-blue-500 hover:text-blue-700 focus:outline-none"
              >
                <i class="bx bx-edit text-lg"></i>
              </button>
              <button 
                @click="emit('deleteRegex', 'other', index)"
                class="text-red-500 hover:text-red-700 focus:outline-none"
              >
                <i class="bx bx-trash text-lg"></i>
              </button>
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
/* 滚动条样式 */
::-webkit-scrollbar {
  width: 6px;
  height: 6px;
}

::-webkit-scrollbar-track {
  background: transparent;
}

::-webkit-scrollbar-thumb {
  background-color: rgba(156, 163, 175, 0.5);
  border-radius: 3px;
}

::-webkit-scrollbar-thumb:hover {
  background-color: rgba(156, 163, 175, 0.7);
}

/* 截断过长文本 */
.truncate {
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}
</style> 