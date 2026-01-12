<script setup>
import { ref, watch } from "vue";
// @ts-ignore
import { GetCollectionMsg } from "../../../bindings/github.com/yhy0/ChYing/app.js";
import { copyList } from "../../utils/copy";

// 定义 props
const props = defineProps({
  host: String,
});

const info = ref({
  Subdomain: [],
  OtherDomain: [],
  PublicIp: [],
  InnerIp: [],
  Phone: [],
  Email: [],
  IdCard: [],
  Others: [],
  Urls: [],
  Api: [],
  Parameters: [],
});

GetCollectionMsg(props.host).then((result) => {
  info.value.Subdomain = result.subdomains;
  info.value.OtherDomain = result.other_domains;
  info.value.PublicIp = result.public_ip;
  info.value.InnerIp = result.inner_ip;
  info.value.Phone = result.phone;
  info.value.Email = result.email;
  info.value.IdCard = result.id_card;
  info.value.Others = result.others;
  info.value.Urls = result.urls;
  info.value.Api = result.api;
  if (result.parameters) {
    info.value.Parameters = Object.entries(result.parameters).map(([key, value]) => ({ key, value }));
  }
})

</script>

<template>
  <div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4 gap-4">
    <!-- API 卡片 -->
    <div class="bg-white dark:bg-[#1e1e2e] rounded-lg border border-gray-200 dark:border-gray-700 shadow-sm transition-all duration-300 hover:shadow-md">
      <div class="p-4 border-b border-gray-200 dark:border-gray-700 flex justify-between items-center">
        <h3 class="text-base font-medium text-gray-800 dark:text-gray-200">API</h3>
        <button 
          class="p-1.5 rounded text-gray-500 hover:text-indigo-600 hover:bg-gray-100 dark:hover:bg-gray-800 transition-colors"
          @click="copyList(info.Api)"
        >
          <i class="bx bx-copy text-lg"></i>
        </button>
      </div>
      <div class="p-2 max-h-[200px] overflow-y-auto custom-scrollbar">
        <ul v-if="info.Api.length > 0" class="divide-y divide-gray-200 dark:divide-gray-700">
          <li 
            v-for="(item, index) in info.Api" 
            :key="index"
            class="py-2 px-3 text-sm text-gray-700 dark:text-gray-300 hover:bg-gray-50 dark:hover:bg-gray-800 cursor-pointer"
          >
            {{ item }}
          </li>
        </ul>
        <p v-else class="text-sm text-gray-500 dark:text-gray-400 p-2 text-center">
          暂无数据
        </p>
      </div>
    </div>

    <!-- Parameters 卡片 -->
    <div class="bg-white dark:bg-[#1e1e2e] rounded-lg border border-gray-200 dark:border-gray-700 shadow-sm transition-all duration-300 hover:shadow-md">
      <div class="p-4 border-b border-gray-200 dark:border-gray-700 flex justify-between items-center">
        <h3 class="text-base font-medium text-gray-800 dark:text-gray-200">Parameters</h3>
        <button 
          class="p-1.5 rounded text-gray-500 hover:text-indigo-600 hover:bg-gray-100 dark:hover:bg-gray-800 transition-colors"
          @click="copyList(info.Parameters)"
        >
          <i class="bx bx-copy text-lg"></i>
        </button>
      </div>
      <div class="p-2 max-h-[200px] overflow-y-auto custom-scrollbar">
        <ul v-if="info.Parameters.length > 0" class="divide-y divide-gray-200 dark:divide-gray-700">
          <li 
            v-for="(item, index) in info.Parameters" 
            :key="index"
            class="py-2 px-3 text-sm text-gray-700 dark:text-gray-300 hover:bg-gray-50 dark:hover:bg-gray-800 cursor-pointer flex justify-between"
          >
            <span>{{ item.key }}</span>
            <span class="text-gray-500 dark:text-gray-400">{{ item.value }}</span>
          </li>
        </ul>
        <p v-else class="text-sm text-gray-500 dark:text-gray-400 p-2 text-center">
          暂无数据
        </p>
      </div>
    </div>

    <!-- Subdomain 卡片 -->
    <div class="bg-white dark:bg-[#1e1e2e] rounded-lg border border-gray-200 dark:border-gray-700 shadow-sm transition-all duration-300 hover:shadow-md">
      <div class="p-4 border-b border-gray-200 dark:border-gray-700 flex justify-between items-center">
        <h3 class="text-base font-medium text-gray-800 dark:text-gray-200">Subdomain</h3>
        <button 
          class="p-1.5 rounded text-gray-500 hover:text-indigo-600 hover:bg-gray-100 dark:hover:bg-gray-800 transition-colors"
          @click="copyList(info.Subdomain)"
        >
          <i class="bx bx-copy text-lg"></i>
        </button>
      </div>
      <div class="p-2 max-h-[200px] overflow-y-auto custom-scrollbar">
        <ul v-if="info.Subdomain.length > 0" class="divide-y divide-gray-200 dark:divide-gray-700">
          <li 
            v-for="(item, index) in info.Subdomain" 
            :key="index"
            class="py-2 px-3 text-sm text-gray-700 dark:text-gray-300 hover:bg-gray-50 dark:hover:bg-gray-800 cursor-pointer"
          >
            {{ item }}
          </li>
        </ul>
        <p v-else class="text-sm text-gray-500 dark:text-gray-400 p-2 text-center">
          暂无数据
        </p>
      </div>
    </div>

    <!-- Other Domain 卡片 -->
    <div class="bg-white dark:bg-[#1e1e2e] rounded-lg border border-gray-200 dark:border-gray-700 shadow-sm transition-all duration-300 hover:shadow-md">
      <div class="p-4 border-b border-gray-200 dark:border-gray-700 flex justify-between items-center">
        <h3 class="text-base font-medium text-gray-800 dark:text-gray-200">Other Domain</h3>
        <button 
          class="p-1.5 rounded text-gray-500 hover:text-indigo-600 hover:bg-gray-100 dark:hover:bg-gray-800 transition-colors"
          @click="copyList(info.OtherDomain)"
        >
          <i class="bx bx-copy text-lg"></i>
        </button>
      </div>
      <div class="p-2 max-h-[200px] overflow-y-auto custom-scrollbar">
        <ul v-if="info.OtherDomain.length > 0" class="divide-y divide-gray-200 dark:divide-gray-700">
          <li 
            v-for="(item, index) in info.OtherDomain" 
            :key="index"
            class="py-2 px-3 text-sm text-gray-700 dark:text-gray-300 hover:bg-gray-50 dark:hover:bg-gray-800 cursor-pointer"
          >
            {{ item }}
          </li>
        </ul>
        <p v-else class="text-sm text-gray-500 dark:text-gray-400 p-2 text-center">
          暂无数据
        </p>
      </div>
    </div>

    <!-- Public IP 卡片 -->
    <div class="bg-white dark:bg-[#1e1e2e] rounded-lg border border-gray-200 dark:border-gray-700 shadow-sm transition-all duration-300 hover:shadow-md">
      <div class="p-4 border-b border-gray-200 dark:border-gray-700 flex justify-between items-center">
        <h3 class="text-base font-medium text-gray-800 dark:text-gray-200">Public IP</h3>
        <button 
          class="p-1.5 rounded text-gray-500 hover:text-indigo-600 hover:bg-gray-100 dark:hover:bg-gray-800 transition-colors"
          @click="copyList(info.PublicIp)"
        >
          <i class="bx bx-copy text-lg"></i>
        </button>
      </div>
      <div class="p-2 max-h-[200px] overflow-y-auto custom-scrollbar">
        <ul v-if="info.PublicIp.length > 0" class="divide-y divide-gray-200 dark:divide-gray-700">
          <li 
            v-for="(item, index) in info.PublicIp" 
            :key="index"
            class="py-2 px-3 text-sm text-gray-700 dark:text-gray-300 hover:bg-gray-50 dark:hover:bg-gray-800 cursor-pointer"
          >
            {{ item }}
          </li>
        </ul>
        <p v-else class="text-sm text-gray-500 dark:text-gray-400 p-2 text-center">
          暂无数据
        </p>
      </div>
    </div>

    <!-- Inner IP 卡片 -->
    <div class="bg-white dark:bg-[#1e1e2e] rounded-lg border border-gray-200 dark:border-gray-700 shadow-sm transition-all duration-300 hover:shadow-md">
      <div class="p-4 border-b border-gray-200 dark:border-gray-700 flex justify-between items-center">
        <h3 class="text-base font-medium text-gray-800 dark:text-gray-200">Inner IP</h3>
        <button 
          class="p-1.5 rounded text-gray-500 hover:text-indigo-600 hover:bg-gray-100 dark:hover:bg-gray-800 transition-colors"
          @click="copyList(info.InnerIp)"
        >
          <i class="bx bx-copy text-lg"></i>
        </button>
      </div>
      <div class="p-2 max-h-[200px] overflow-y-auto custom-scrollbar">
        <ul v-if="info.InnerIp.length > 0" class="divide-y divide-gray-200 dark:divide-gray-700">
          <li 
            v-for="(item, index) in info.InnerIp" 
            :key="index"
            class="py-2 px-3 text-sm text-gray-700 dark:text-gray-300 hover:bg-gray-50 dark:hover:bg-gray-800 cursor-pointer"
          >
            {{ item }}
          </li>
        </ul>
        <p v-else class="text-sm text-gray-500 dark:text-gray-400 p-2 text-center">
          暂无数据
        </p>
      </div>
    </div>

    <!-- Phone 卡片 -->
    <div class="bg-white dark:bg-[#1e1e2e] rounded-lg border border-gray-200 dark:border-gray-700 shadow-sm transition-all duration-300 hover:shadow-md">
      <div class="p-4 border-b border-gray-200 dark:border-gray-700 flex justify-between items-center">
        <h3 class="text-base font-medium text-gray-800 dark:text-gray-200">Phone</h3>
        <button 
          class="p-1.5 rounded text-gray-500 hover:text-indigo-600 hover:bg-gray-100 dark:hover:bg-gray-800 transition-colors"
          @click="copyList(info.Phone)"
        >
          <i class="bx bx-copy text-lg"></i>
        </button>
      </div>
      <div class="p-2 max-h-[200px] overflow-y-auto custom-scrollbar">
        <ul v-if="info.Phone.length > 0" class="divide-y divide-gray-200 dark:divide-gray-700">
          <li 
            v-for="(item, index) in info.Phone" 
            :key="index"
            class="py-2 px-3 text-sm text-gray-700 dark:text-gray-300 hover:bg-gray-50 dark:hover:bg-gray-800 cursor-pointer"
          >
            {{ item }}
          </li>
        </ul>
        <p v-else class="text-sm text-gray-500 dark:text-gray-400 p-2 text-center">
          暂无数据
        </p>
      </div>
    </div>

    <!-- Email 卡片 -->
    <div class="bg-white dark:bg-[#1e1e2e] rounded-lg border border-gray-200 dark:border-gray-700 shadow-sm transition-all duration-300 hover:shadow-md">
      <div class="p-4 border-b border-gray-200 dark:border-gray-700 flex justify-between items-center">
        <h3 class="text-base font-medium text-gray-800 dark:text-gray-200">Email</h3>
        <button 
          class="p-1.5 rounded text-gray-500 hover:text-indigo-600 hover:bg-gray-100 dark:hover:bg-gray-800 transition-colors"
          @click="copyList(info.Email)"
        >
          <i class="bx bx-copy text-lg"></i>
        </button>
      </div>
      <div class="p-2 max-h-[200px] overflow-y-auto custom-scrollbar">
        <ul v-if="info.Email.length > 0" class="divide-y divide-gray-200 dark:divide-gray-700">
          <li 
            v-for="(item, index) in info.Email" 
            :key="index"
            class="py-2 px-3 text-sm text-gray-700 dark:text-gray-300 hover:bg-gray-50 dark:hover:bg-gray-800 cursor-pointer"
          >
            {{ item }}
          </li>
        </ul>
        <p v-else class="text-sm text-gray-500 dark:text-gray-400 p-2 text-center">
          暂无数据
        </p>
      </div>
    </div>

    <!-- ID Card 卡片 -->
    <div class="bg-white dark:bg-[#1e1e2e] rounded-lg border border-gray-200 dark:border-gray-700 shadow-sm transition-all duration-300 hover:shadow-md">
      <div class="p-4 border-b border-gray-200 dark:border-gray-700 flex justify-between items-center">
        <h3 class="text-base font-medium text-gray-800 dark:text-gray-200">ID Card</h3>
        <button 
          class="p-1.5 rounded text-gray-500 hover:text-indigo-600 hover:bg-gray-100 dark:hover:bg-gray-800 transition-colors"
          @click="copyList(info.IdCard)"
        >
          <i class="bx bx-copy text-lg"></i>
        </button>
      </div>
      <div class="p-2 max-h-[200px] overflow-y-auto custom-scrollbar">
        <ul v-if="info.IdCard.length > 0" class="divide-y divide-gray-200 dark:divide-gray-700">
          <li 
            v-for="(item, index) in info.IdCard" 
            :key="index"
            class="py-2 px-3 text-sm text-gray-700 dark:text-gray-300 hover:bg-gray-50 dark:hover:bg-gray-800 cursor-pointer"
          >
            {{ item }}
          </li>
        </ul>
        <p v-else class="text-sm text-gray-500 dark:text-gray-400 p-2 text-center">
          暂无数据
        </p>
      </div>
    </div>

    <!-- Others 卡片 -->
    <div class="bg-white dark:bg-[#1e1e2e] rounded-lg border border-gray-200 dark:border-gray-700 shadow-sm transition-all duration-300 hover:shadow-md">
      <div class="p-4 border-b border-gray-200 dark:border-gray-700 flex justify-between items-center">
        <h3 class="text-base font-medium text-gray-800 dark:text-gray-200">Others</h3>
        <button 
          class="p-1.5 rounded text-gray-500 hover:text-indigo-600 hover:bg-gray-100 dark:hover:bg-gray-800 transition-colors"
          @click="copyList(info.Others)"
        >
          <i class="bx bx-copy text-lg"></i>
        </button>
      </div>
      <div class="p-2 max-h-[200px] overflow-y-auto custom-scrollbar">
        <ul v-if="info.Others.length > 0" class="divide-y divide-gray-200 dark:divide-gray-700">
          <li 
            v-for="(item, index) in info.Others" 
            :key="index"
            class="py-2 px-3 text-sm text-gray-700 dark:text-gray-300 hover:bg-gray-50 dark:hover:bg-gray-800 cursor-pointer"
          >
            {{ item }}
          </li>
        </ul>
        <p v-else class="text-sm text-gray-500 dark:text-gray-400 p-2 text-center">
          暂无数据
        </p>
      </div>
    </div>

    <!-- URLs 卡片 -->
    <div class="bg-white dark:bg-[#1e1e2e] rounded-lg border border-gray-200 dark:border-gray-700 shadow-sm transition-all duration-300 hover:shadow-md">
      <div class="p-4 border-b border-gray-200 dark:border-gray-700 flex justify-between items-center">
        <h3 class="text-base font-medium text-gray-800 dark:text-gray-200">URLs</h3>
        <button 
          class="p-1.5 rounded text-gray-500 hover:text-indigo-600 hover:bg-gray-100 dark:hover:bg-gray-800 transition-colors"
          @click="copyList(info.Urls)"
        >
          <i class="bx bx-copy text-lg"></i>
        </button>
      </div>
      <div class="p-2 max-h-[200px] overflow-y-auto custom-scrollbar">
        <ul v-if="info.Urls.length > 0" class="divide-y divide-gray-200 dark:divide-gray-700">
          <li 
            v-for="(item, index) in info.Urls" 
            :key="index"
            class="py-2 px-3 text-sm text-gray-700 dark:text-gray-300 hover:bg-gray-50 dark:hover:bg-gray-800 cursor-pointer"
          >
            {{ item }}
          </li>
        </ul>
        <p v-else class="text-sm text-gray-500 dark:text-gray-400 p-2 text-center">
          暂无数据
        </p>
      </div>
    </div>
  </div>
</template>

<style scoped>
.custom-scrollbar::-webkit-scrollbar {
  width: 6px;
  height: 6px;
}

.custom-scrollbar::-webkit-scrollbar-track {
  background: transparent;
}

.custom-scrollbar::-webkit-scrollbar-thumb {
  background: #d1d5db;
  border-radius: 3px;
}

.dark .custom-scrollbar::-webkit-scrollbar-thumb {
  background: #4b5563;
}

.custom-scrollbar::-webkit-scrollbar-thumb:hover {
  background: #9ca3af;
}

.dark .custom-scrollbar::-webkit-scrollbar-thumb:hover {
  background: #6b7280;
}
</style> 