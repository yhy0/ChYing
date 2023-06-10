<script setup>
import Home from './components/Home.vue'

import hljs from 'highlight.js/lib/core'  // 代码高亮
import accesslog from 'highlight.js/lib/languages/accesslog'
import json from 'highlight.js/lib/languages/json'
import http from 'highlight.js/lib/languages/http'
import xml from 'highlight.js/lib/languages/xml'
import java from 'highlight.js/lib/languages/java'

hljs.registerLanguage("accesslog", accesslog)
hljs.registerLanguage("json", json)
hljs.registerLanguage("http", http)
hljs.registerLanguage("xml", xml)
hljs.registerLanguage("java", java)

import { useOsTheme, darkTheme } from "naive-ui";
import {computed} from "vue";
const osThemeRef = useOsTheme();
const theme = computed(() => osThemeRef.value === "dark" ? darkTheme : null);

</script>

<template>
    <n-config-provider :hljs="hljs" :theme="theme">
        <!-- https://www.naiveui.com/zh-CN/os-theme/components/global-style -->
        <n-global-style />
        <Home/>
    </n-config-provider>
</template>

