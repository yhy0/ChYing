## 安装 wails3

注意 wails 应该和 ChYing 目录在一级，也就是 ls
...
ChYing
wails
...

go.mod 中是这样写的
replace github.com/wailsapp/wails/v3 => ../wails/v3


```azure
git clone https://github.com/wailsapp/wails.git
cd wails
git checkout v3-alpha
cd v3/cmd/wails3
go install
```

如果执行 wails3 失败，则需要看 go 的 bin 目录是否已经加入到环境变量 GOPATH
还需要 npm 环境

## Windows 本地开发环境要求

本项目依赖 CGO（go-sqlite3、tree-sitter-javascript 等），Windows 本地开发需要安装 MinGW：

### 安装 MinGW（推荐使用 MSYS2）

1. 下载并安装 MSYS2：https://www.msys2.org/
2. 打开 MSYS2 UCRT64 终端，执行：
   ```bash
   pacman -S mingw-w64-ucrt-x86_64-gcc
   ```
3. 将 MinGW bin 目录添加到系统 PATH 环境变量：
   ```
   C:\msys64\ucrt64\bin
   ```
4. 验证安装：
   ```bash
   gcc --version
   ```

### 或使用 Chocolatey 安装

```powershell
choco install mingw
```

### 确保 CGO 启用

编译时需要确保 `CGO_ENABLED=1`：
```bash
set CGO_ENABLED=1
wails3 task windows:build PRODUCTION=true
```

**注意**：如果编译时 `CGO_ENABLED=0`，会导致 sqlite 数据库无法工作，程序运行时会崩溃。

### PRODUCTION 参数说明

`PRODUCTION=true` 参数用于区分开发构建和生产构建：

| 参数 | 构建标志 | 用途 |
|------|---------|------|
| 不设置或 `PRODUCTION=false` | `-buildvcs=false -gcflags=all="-l"` | 开发调试，保留调试信息 |
| `PRODUCTION=true` | `-tags production -trimpath -buildvcs=false -ldflags="-w -s"` | 生产发布，优化体积，移除调试信息 |

**Windows 本地构建必须使用 `PRODUCTION=true`**，否则可能无法正常运行。

https://github.com/yhy0/ChYing-Inside
https://github.com/wailsapp/wails.git



运行会在 /Users/你的用户名/.config/ChYing 下生成一个文件夹

~/.config/ChYing/proxify_data/cacert.pem  双击安装证书，用于捕获 https 流量

debug
```bash
ulimit -c unlimited
export GOTRACEBACK=crash
go install github.com/go-delve/delve/cmd/dlv@latest
```

# 一次性多版本
wails3 task darwin:package:universal


package
```shell
 wails3 task package
 
 or 

 wails3 task darwin:package
 wails3 task windows:package
 wails3 task linux:package
```
mac 下编译 Windows 平台参考

```bash
brew install mingw-w64
CGO_ENABLED=1
wails3 task windows:package
```

https://icon-sets.iconify.design/icon-park-outline/?category=General

# 前端依赖更新
pnpm install -g npm-check-updates
下载后使用 

cd frontend

ncu -u
