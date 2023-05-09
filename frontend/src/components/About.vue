<script setup>
import {BrowserOpenURL} from "../../wailsjs/runtime/runtime.js";

function openURL(url) {
    BrowserOpenURL(url);
}
</script>

<template>
    <n-collapse :default-expanded-names="['3', '4']">
        <n-collapse-item title="已有功能" name="1">
            <div>
                <n-card style="text-align: left;">
                    <n-h3> 目录扫描 </n-h3>
                    <n-p> 提取 <n-button @click="openURL('https://github.com/maurosoria/dirsearch')" quaternary type="primary">Dirsearch</n-button> 的字典规则进行扫描，目前只会进行一层目录扫描，后期考虑根据找到的目录，进行多层目录遍历</n-p>
                    <n-p><n-button @click="openURL('https://github.com/lijiejie/bbscan')" quaternary type="primary">BBscan</n-button> 规则扫描</n-p>
                </n-card>

                <n-card style="text-align: left;">
                    <n-h3> Swagger 测试 </n-h3>
                    <n-p> 对 `swagger api` 进行未授权、ssrf、注入等测试</n-p>
                </n-card>

                <n-card style="text-align: left;">
                    <n-h3> 403 bypass </n-h3>
                    <n-p> 上述两个功能会自动进行 403 bypass</n-p>
                    <n-p><n-button @click="openURL('https://github.com/devploit/dontgo403')" quaternary type="primary">dontgo403</n-button> </n-p>
                    <n-p><n-button @click="openURL('https://infosecwriteups.com/403-bypass-lyncdiscover-microsoft-com-db2778458c33')" quaternary type="primary">403-bypass-lyncdiscover-microsoft-com-db2778458c33</n-button> </n-p>
                </n-card>
                <n-card style="text-align: left;">
                    <n-h3> JWT 测试 </n-h3>
                    <n-p> JWT token 解析，<n-button @click="openURL('https://jwt.io/')" quaternary type="primary">jwt.io</n-button> 样式显示</n-p>
                    <n-p> JWT 秘钥爆破 </n-p>
                </n-card>

                <n-card style="text-align: left;">
                    <n-h3> BurpSuite </n-h3>
                    <n-p> Proxy 模块 HTTP history 部分实现，其它未实现 </n-p>
                    <n-p> Repeater 模块</n-p>
                    <n-p> Intruder 模块 四种Fuzz模式已实现， payload部分各种处理函数未实现，现在只有MD5加密</n-p>
                </n-card>
            </div>
        </n-collapse-item>
        <n-collapse-item title="字典可配置" name="2">
            <n-card style="text-align: left;">
                用到的各种字典文件, 第一次运行会将内置字典释放到用户目录的`.config/ChYing`目录下，后续每次运行都会先读取一遍
            </n-card>
        </n-collapse-item>
        <n-collapse-item title="问题" name="3">
            <n-card style="text-align: left;">
                <n-gradient-text type="error">
                    现在各个 tabs 页面，不点进去不会激活，导致 BurpSuite 用之前必须点击一遍每个页面
                </n-gradient-text>
                <n-gradient-text type="error">
                    Intruder 模块 手动修改request包内容时，每写一个字符，要重新点击鼠标，将光标重新定位到要修改的位置
                </n-gradient-text>
                <n-gradient-text type="error">
                    Attack 显示不能切换别的 Intruder tab页，不然结果就不显示了，前端数据绑定问题，太菜了，还没想好怎么写
                </n-gradient-text>
            </n-card>
        </n-collapse-item>

        <n-collapse-item title="yhy" name="4">
            <n-card >
                <div style="display: flex; justify-content: center;">
                    <n-button @click="openURL('https://github.com/yhy0/')" quaternary type="primary" size="large">yhy</n-button>
                    <n-button @click="openURL('https://github.com/yhy0/ChYing')" quaternary type="primary" size="large">承影</n-button>
                </div>
                <img src="../assets/images/yhy.jpeg" alt="yhy"/>
            </n-card>
        </n-collapse-item>
    </n-collapse>
</template>

<style scoped>

</style>