## 敏感信息检测

~~使用[secrets-patterns-db](https://github.com/mazen160/secrets-patterns-db/)规则~~

误报过多，使用 nuclei[https://github.com/projectdiscovery/nuclei-templates/tree/main/file/keys]  ，从中提取相关正则

检测流程
```bash
无头浏览器爬虫扫描获取body -> 使用规则正则匹配
```

## TODO
上报过多，需要优化