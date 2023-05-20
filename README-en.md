## ChYing

<p align="center">
将旦昧爽之交，日夕昏明之际，北面而察之，淡淡焉若有物存，莫识其状。其所触也，窃窃然有声，经物而物不疾也。
  <br/>
  <br/>
  <a href="https://github.com/yhy0/ChYing/blob/main/LICENSE">
    <img alt="Release" src="https://img.shields.io/github/license/yhy0/ChYing"/>
  </a>
  <a href="https://github.com/yhy0/ChYing">
    <img alt="Release" src="https://img.shields.io/badge/release-v0.9-brightgreen"/>
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


ChYing is a comprehensive security toolbox designed to simplify various security testing tasks. It provides a range of features and tools, including directory scanning, JWT , Swagger API testing, encoding/decoding utilities, a lightweight BurpSuite alternative, and antivirus assistance. ChYing aims to assist security professionals and developers in identifying vulnerabilities and strengthening the security of their applications.

<video controls="controls" loop="loop" autoplay="autoplay"> 
    <source src="images/ChYing.mp4" type="video/mp4">
</video>

## Project Setup

https://wails.io/docs/gettingstarted/installation/

Install **Wails**.

Then run `wails build`.

## Features

### Directory Scanning

Scanning using dictionary rules extracted from [dirsearch](https://github.com/maurosoria/dirsearch). Currently, only scans a single level of directories. Future considerations include traversing multiple levels of directories based on the discovered directories.

Scanning with [bbscan](https://github.com/lijiejie/bbscan) rules.

### Swagger Testing

Unauthenticated, SSRF, and injection testing on `swagger api`.

### 403 Bypass

Automatic 403 bypass for the Swagger features.

https://github.com/devploit/dontgo403

Not implemented: https://infosecwriteups.com/403-bypass-lyncdiscover-microsoft-com-db2778458c33

### JWT

- JWT token parsing with visual display similar to [jwt.io](https://jwt.io/).
- JWT key cracking.

### BurpSuite

Utilizing the features of the [go-mitmproxy](https://github.com/lqqyt2423/go-mitmproxy) project to replicate BurpSuite functionality.

[Certificate Installation](https://github.com/lqqyt2423/go-mitmproxy#usage):

After launching, the default HTTP proxy address is set to port 9080.

For the first launch, you need to install a certificate to decrypt HTTPS traffic. The certificate will be automatically generated after the first launch command and saved in ~/.mitmproxy/mitmproxy-ca-cert.pem. The installation steps can be found in the Python mitmproxy documentation: [Certificates](https://docs.mitmproxy.org/stable/concepts-certificates/).

-   [x] Proxy module
-   [x] Repeater module
-   [x] Intruder module

### Configurable Dictionaries

Various dictionary files are used. On the first run, the built-in dictionaries will be released to the `.config/ChYing` directory in the user's folder, and they will be read on each subsequent run.

### Encoding and Decoding
Unicode, URL, Hex, Base64 encoding/decoding.

MD5 encryption.

### Antivirus Recognition

https://github.com/gh0stkey/avList/blob/master/avlist.js

## Issues
Lack of frontend expertise; heavily reliant on ChatGPT.

- Currently, each tab page needs to be clicked to activate it, which means BurpSuite requires clicking through each page before using it.
- Intruder module
  - The Attack display cannot switch to other Intruder tab pages, otherwise the results won't be displayed. It's a frontend data binding issue. Still figuring out the best way to address it.

## License

This code is distributed under the [MIT license](https://github.com/yhy0/ChYing/blob/main/LICENSE). See [LICENSE](https://github.com/yhy0/ChYing/blob/main/LICENSE) in this directory.

## Acknowledgements

Special thanks to [JetBrains](https://www.jetbrains.com/) for providing a range of powerful IDEs and supporting this project.

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
