## 承影

<p align="center">
将旦昧爽之交，日夕昏明之际，北面而察之，淡淡焉若有物存，莫识其状。其所触也，窃窃然有声，经物而物不疾也。
  <br/>
  <br/>
  <a href="https://github.com/yhy0/ChYing/blob/main/LICENSE">
    <img alt="Release" src="https://img.shields.io/github/license/yhy0/ChYing"/>
  </a>
  <a href="https://github.com/yhy0/ChYing">
    <img alt="Release" src="https://img.shields.io/badge/release-v1.1-brightgreen"/>
  </a>
  <a href="https://github.com/yhy0/ChYing">
    <img alt="GitHub Repo stars" src="https://img.shields.io/github/stars/yhy0/ChYing?color=9cf"/>
  </a>
  <a href="https://github.com/yhy0/ChYing">
    <img alt="GitHub forks" src="https://img.shields.io/github/forks/yhy0/ChYing"/>
  </a>
  <a href="https://github.com/yhy0/ChYing">
    <img alt="GitHub all release" src="https://img.shields.io/github/downloads/yhy0/ChYing/total?color=blueviolet"/>
  </a>
</p>
<div align="center">
<strong>
<samp>
  
[简体中文](./README.md) · [English](./README-en.md)
  
</samp>
</strong>
</div>

https://github.com/yhy0/ChYing/assets/31311038/54cc1130-fb95-4a8f-b90e-3479e9c5a2c7

<video controls="controls" loop="loop" autoplay="autoplay"> 
    <source src="images/ChYing.mp4" type="video/mp4">
</video>

## 构建项目

https://wails.io/zh-Hans/docs/gettingstarted/installation/

安装 **wails** 

然后 `wails build`

## 已有功能

### 目录扫描

提取 [dirsearch](https://github.com/maurosoria/dirsearch) 的字典规则进行扫描，目前只会进行一层目录扫描，后期考虑根据找到的目录，进行多层目录遍历

[bbscan](https://github.com/lijiejie/bbscan) 规则扫描

### Swagger 测试

对 `swagger api` 进行未授权、ssrf、注入等测试

### 403 bypass

Swagger 会自动进行 403 bypass

https://github.com/devploit/dontgo403

https://infosecwriteups.com/403-bypass-lyncdiscover-microsoft-com-db2778458c33

### JWT

- JWT token 解析，[jwt.io](https://jwt.io/) 样式显示
- JWT 秘钥爆破

### NucleiY

基于 nuclei 实现的重点漏洞扫描, 使用前往 https://github.com/yhy0/nucleiY 查看说明

### BurpSuite

使用 [go-mitmproxy](https://github.com/lqqyt2423/go-mitmproxy) 项目实现 BurpSuite 的 功能

[证书安装](https://github.com/lqqyt2423/go-mitmproxy#usage):

启动后HTTP代理地址默认设置为9080端口

第一次启动后需要安装证书来解析HTTPS流量。 证书会在第一次启动命令后自动生成，保存在~/.mitmproxy/mitmproxy-ca-cert.pem. 安装步骤可以在 Python mitmproxy 文档中找到：[关于证书](https://docs.mitmproxy.org/stable/concepts-certificates/)。

-   [x] Proxy 模块
-   [x] Repeater 模块
-   [x] Intruder 模块


### 字典可配置

用到的各种字典文件, 第一次运行会将内置字典释放到用户目录的`.config/ChYing`目录下，后续每次运行都会先读取一遍

### 编码、解码
Unicode 、URL、Hex、Base64 编/解码

MD5 加密
### 杀软识别

https://github.com/gh0stkey/avList/blob/master/avlist.js

## 问题
前端不会，全靠 ChatGPT 

- 现在各个 tabs 页面，不点进去不会激活，导致 BurpSuite 用之前必须点击一遍每个页面
- Intruder 模块
  - Attack 显示不能切换别的 Intruder tab页，不然结果就不显示了，前端数据绑定问题，太菜了，还没想好怎么写

## License

This code is distributed under the [AGPL-3.0 license](https://github.com/yhy0/ChYing/blob/main/LICENSE). See [LICENSE](https://github.com/yhy0/ChYing/blob/main/LICENSE) in this directory.

## 鸣谢

感谢 [JetBrains](https://www.jetbrains.com/) 提供的一系列好用的 IDE 和对本项目的支持。

![JetBrains Logo (Main) logo](https://resources.jetbrains.com/storage/products/company/brand/logos/jb_beam.svg)

https://github.com/lijiejie/bbscan

https://github.com/maurosoria/dirsearch

https://github.com/devploit/dontgo403

https://github.com/lqqyt2423/go-mitmproxy

https://github.com/gh0stkey/avList/

https://wails.io/

https://www.naiveui.com/

## Star History

[![Star History Chart](https://api.star-history.com/svg?repos=yhy0/ChYing&type=Date)](https://star-history.com/#yhy0/ChYing&Date)
