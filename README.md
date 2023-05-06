## 承影

将旦昧爽之交，日夕昏明之际，北面而察之，淡淡焉若有物存，莫识其状。其所触也，窃窃然有声，经物而物不疾也。

<video src="./images/ChYing.mp4"></video>

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

上述两个功能会自动进行 403 bypass

https://github.com/devploit/dontgo403

https://infosecwriteups.com/403-bypass-lyncdiscover-microsoft-com-db2778458c33

### JWT

- JWT token 解析，[jwt.io](https://jwt.io/) 样式显示
- JWT 秘钥爆破

### BurpSuite

使用 [go-mitmproxy](https://github.com/lqqyt2423/go-mitmproxy) 项目实现 BurpSuite 的 功能

-   [ ] Proxy 模块 						HTTP history 部分实现，其它未实现
-   [x] Repeater 模块               
-   [ ] Intruder 模块                

### 字典可配置

用到的各种字典文件, 第一次运行会将内置字典释放到用户目录的`.config/ChYing`目录下，后续每次运行都会先读取一遍

## 问题

- 现在各个 tabs 页面，不点进去不会激活，导致 BurpSuite 用之前必须点击一遍每个页面

## License

This code is distributed under the [MIT license](https://github.com/yhy0/ChYing/blob/main/LICENSE). See [LICENSE](https://github.com/yhy0/ChYing/blob/main/LICENSE) in this directory.

## 鸣谢

感谢 [JetBrains](https://www.jetbrains.com/) 提供的一系列好用的 IDE 和对本项目的支持。

![JetBrains Logo (Main) logo](https://resources.jetbrains.com/storage/products/company/brand/logos/jb_beam.svg)

https://github.com/lijiejie/bbscan

https://github.com/lqqyt2423/go-mitmproxy

## Star History

[![Star History Chart](https://api.star-history.com/svg?repos=yhy0/ChYing&type=Date)](https://star-history.com/#yhy0/ChYing&Date)