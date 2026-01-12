## WebUnpack

webpack 打包器扫描解析获取 API,本来想写成[Packer-Fuzzer](https://github.com/rtcatc/Packer-Fuzzer) 一样，又放弃了

最终缝合了[sourcemap](https://github.com/orsinium-labs/sourcemap) 和 [WebFinder](https://github.com/L0nm4r/WebFinder)


目前功能为解析 xxx.js.map 还原为源代码格式, 然后对 js 进行扫描提取
没有 xxx.js.map 的就无能为力了
