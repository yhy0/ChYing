<script setup>
import { ref, computed } from "vue";
import {NCard, useMessage} from "naive-ui";
import {EventsOn} from "../../../../wailsjs/runtime/runtime.js";
import {Raw} from "../../../../wailsjs/go/main/App.js";

const message = useMessage();

const value = ref("0");
const panels = ref([]);

const addable = computed(() => {
    return {
        disabled: false
    };
});

const closable = computed(() => {
    return panels.value.length > 1;
});

function handleAdd() {
    const newKey = String(+panels.value[panels.value.length - 1].key + 1);
    panels.value.push({
        title: `Tab ${newKey}`,
        req: ``,
        res: ``,
        url: ``,
        id: ``,
        key: newKey,
    });
    value.value = newKey;
}

function handleClose(name) {
    const index = panels.value.findIndex(panel => panel.key === name);
    if (index === -1) {
        return;
    }
    panels.value.splice(index, 1);
    const newKey = panels.value[Math.min(index, panels.value.length - 1)].key;
    value.value = newKey;
}

EventsOn("RepeaterBody", result => {
    const newKey = String(panels.value.length);
    const newPanel = {
        title: `Tab ${newKey}`,
        req: result.request,
        res: result.response,
        url: result.targetUrl,
        key: newKey,
        id: "",
    };
    panels.value.push(newPanel);
    value.value = newKey;
});

const request = ref('');
// 通过 id 监听 更改，应该有更优雅的实现方式，但是我不会。。。
function updateReqValue() {
    request.value = document.getElementById("myCode").textContent;
}

function send(panel) {
    console.log(panel.url);
    Raw(request.value, panel.url, panel.id).then(result=>{
        panel.req = result.request;
        panel.res = result.response;
        panel.id = result.uuid;
    })
}



</script>

<template>
    <n-tabs
        v-model:value="value"
        type="card"
        :addable="addable"
        :closable="closable"
        tab-style="min-width: 80px;"
        @close="handleClose"
        @add="handleAdd"
    >

        <n-tab-pane v-for="panel in panels" :key="panel.key" :name="panel.key">
            <n-space align="center">
                <n-button color="#ff6633" @click="send(panel)">Send</n-button>
                <n-button type="tertiary" size="small">《 </n-button>
                <n-button type="tertiary" size="small"> 》</n-button>
            </n-space>
            <n-card style="margin-bottom: 16px; margin-top: 10px">
                <n-grid :x-gap="12" :cols="2">
                    <n-gi>
                        <n-tabs type="line" animated >
                            <n-tab-pane name="request" style="width: 100%; overflow-x: auto;">
                                <!-- contenteditable 设置为可修改, 这样通过 id 获取值 不能再使用 show-line-numbers 显示行号，不然会将行号带到请求中 -->
                                <n-code id="myCode" contenteditable language="http" :code="panel.req" @input="updateReqValue" style="white-space: pre-wrap; text-align: left;" />
                            </n-tab-pane>
                        </n-tabs>
                    </n-gi>

                    <n-gi>
                        <n-tabs type="line" animated>
                            <n-tab-pane name="response" style="width: 100%; overflow-x: auto;">
                                <n-code language="http" word-wrap :code="panel.res" show-line-numbers style="white-space: pre-wrap; text-align: left; " />
                            </n-tab-pane>
                        </n-tabs>
                    </n-gi>
                </n-grid>
            </n-card>
        </n-tab-pane>
    </n-tabs>
</template>
