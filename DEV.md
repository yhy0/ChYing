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
