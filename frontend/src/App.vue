<script setup>
import Home from './components/Home.vue'

import hljs from 'highlight.js/lib/core'  // 代码高亮
import accesslog from 'highlight.js/lib/languages/accesslog'
import json from 'highlight.js/lib/languages/json'
import http from 'highlight.js/lib/languages/http'
import xml from 'highlight.js/lib/languages/xml'

hljs.registerLanguage("accesslog", accesslog)
hljs.registerLanguage("json", json)
hljs.registerLanguage("http", http)
hljs.registerLanguage("xml", xml)

</script>

<template>
    <n-config-provider :hljs="hljs">
      <Home/>
    </n-config-provider>
</template>

