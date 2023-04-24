<script setup>
import Home from './components/Home.vue'

import hljs from 'highlight.js/lib/core'  // 代码高亮
import accesslog from 'highlight.js/lib/languages/accesslog'
import json from 'highlight.js/lib/languages/json'
import html from 'highlight.js/lib/languages/http'

hljs.registerLanguage("accesslog", accesslog)
hljs.registerLanguage("json", json)
hljs.registerLanguage("html", html)

</script>

<template>
    <n-config-provider :hljs="hljs">
      <Home/>
    </n-config-provider>
</template>

