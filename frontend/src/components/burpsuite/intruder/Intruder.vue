<script setup>
import {computed, ref, watch} from "vue";
import {v4 as uuidv4} from 'uuid';
import {NButton, NCard, useMessage} from "naive-ui";
import {EventsOn} from "../../../../wailsjs/runtime/runtime.js";
import {Intruder} from "../../../../wailsjs/go/main/App.js";
import Attack from "./Attack.vue";
import {PrismEditor} from "vue-prism-editor";

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

const attackTypes = ref([""]); // 用于保存每个标签页对应的攻击类型选项

function handleAdd() {
    const newKey = String(+panels.value[panels.value.length - 1].key + 1);
    attackTypes.value.push('Sniper');
    panels.value.push({
        title: `Tab ${newKey}`,
        req: ``,
        res: ``,
        url: ``,
        id: ``,
        uuid:  uuidv4(),   // 生成随机 UUID
        key: newKey,
        len: 0,
    });
    value.value = newKey;
}

function handleClose(name) {
    const index = panels.value.findIndex(panel => panel.key === name);
    if (index === -1) {
        return;
    }
    panels.value.splice(index, 1);
    value.value = panels.value[Math.min(index, panels.value.length - 1)].key;
}

EventsOn("IntruderBody", result => {
    const newKey = String(panels.value.length);
    const newPanel = {
        title: `Tab ${newKey}`,
        req: result.request,
        url: result.targetUrl,
        key: newKey,
        uuid: uuidv4(),   // 生成随机 UUID
        id: "",
        len: 0,
    };
    attackTypes.value.push('Sniper');
    panels.value.push(newPanel);
    value.value = newKey;
});

const request = ref('');

const options = [
    {
        label: "Sniper",
        value: "Sniper",
    },
    {
        label: "Battering ram",
        value: "Battering ram"
    },
    {
        label: "Pitchfork",
        value: "Pitchfork"
    },
    {
        label: "Cluster bomb",
        value: "Cluster bomb",
    },
]
const payloadCount = ref(0);


// import Prism Editor
import 'vue-prism-editor/dist/prismeditor.min.css'; // import the styles somewhere
// import highlighting library (you can use any library you want just return html string)
import { highlight } from 'prismjs/components/prism-core';
import 'prismjs/components/prism-clike';
import 'prismjs/components/prism-http.js';
import 'prismjs/themes/prism.css';


function highlighter(code) {
    request.value = code.toString()
    return highlight(code, Prism.languages.http,'http'); // languages.<insert language> to return html with markup
}

const selectedText = ref('')
function  handleMouseUp() {
    selectedText.value = window.getSelection().toString();
}

// 鼠标选中 ,点击按钮，增加 §§
function Add(panel) {
    if(selectedText) {
        request.value = request.value.replace(selectedText.value, `§${selectedText.value}§`)
        const count = (request.value.match(/§/g) || []).length; // 计算 § 符号的数量
        payloadCount.value = count / 2;
        if(["Sniper", "Battering ram"].includes(attackTypes.value[Number(value.value)])){
            panel.len = 1;
        } else {
            panel.len = payloadCount.value;
        }
        panel.req = request.value;
    }
}

function Clear(panel) {
    request.value = request.value.replaceAll("§", "");
    panel.req = request.value;
}

const shouldRenderInputs = computed(() => {
    return ["Pitchfork", "Cluster bomb"].includes(attackTypes.value[Number(value.value)])
})

const optionsMapped = [
    {
        label: "",
        value: "None",
    },
    {
        label: "Encode",
        value: "Encode"
    },
    {
        label: "Decode",
        value: "Decode"
    },
    {
        label: "Hash",
        value: "Hash"
    },
]

const subOptions = {
    None :[{label:"", value:"None"}],
    Encode: [
        {label: "Base64-encode", value: "Base64-encode"},
        {label: "Unicode-encode", value: "Unicode-encode"},
        {label: "URL-encode", value: "URL-encode"},
        {label: "Hex-encode", value: "Hex-encode"},
    ],
    Decode: [
        {label: "Base64-decode", value: "Base64-decode"},
        {label: "Unicode-decode", value: "Unicode-decode"},
        {label: "URL-decode", value: "URL-decode"},
        {label: "Hex-decode", value: "Hex-decode"},
    ],
    Hash: [
        {label: "MD5", value: "MD5"},
    ]
}

const selectedOptionSingle = ref(optionsMapped[0].value)
const selectedSubOptionSingle = ref(subOptions[selectedOptionSingle.value][0].value)

watch(selectedOptionSingle, (newValue,) => {
    selectedSubOptionSingle.value = subOptions[newValue][0].value
})

const selectedOption = ref(Array(payloadCount.value).fill(optionsMapped[0].value))
const selectedSubOption = ref(Array(payloadCount.value).fill(subOptions[optionsMapped[0].value][0].value))

const payload = ref([]);
const payloadSingle = ref('');

// 通过改变 keyValue 的值来销毁该 Attack 界面
// const keyValue = ref(0);

function startAattack(panel) {
    // keyValue.value += 1
    const payloads = [];
    const rules = [];
    if(payload.value.length >0) {
        payload.value.forEach((item, index) => {
            payloads.push(item);
            rules.push(selectedSubOption.value[index])
        });
    } else {
        payloads.push(payloadSingle.value);
        rules.push(selectedSubOptionSingle.value)
    }
    if(request.value === "") {
        message.error("Fuzz 参数未设置")
        return
    }

    if(payloads.length === 0) {
        message.error("payload 参数未设置")
        return
    }

    Intruder(panel.url, request.value, payloads, rules, attackTypes.value[panel.key], panel.uuid)
}

</script>

<template>
    <n-tabs
            v-model:value="value"
            type="card"
            :addable="addable"
            :closable="closable"
            @close="handleClose"
            @add="handleAdd"
    >

        <n-tab-pane v-for="panel in panels" :key="panel.key" :name="panel.key" display-directive="show:lazy">
            <n-card style="margin-bottom: 16px; margin-top: 10px">
                <n-tabs type="line" animated display-directive="show:lazy">
                    <n-tab-pane name="Positions">
                        <n-card title="Choose an attack type" size="small">
                            <div style="display: flex; align-items: center;">
                                Attack type: <n-select v-model:value="attackTypes[panel.key]" :options="options" style="width: 80%"/>
                            </div>
                        </n-card>
                        <n-card title="Payload positions" size="large" style="margin-bottom: 16px; margin-top: 10px">
                            <n-layout has-sider sider-placement="right">
                                <n-layout-content content-style="padding: 5px; overflow-x: auto;">
                                    <PrismEditor class="my-editor" v-model="panel.req" :highlight="highlighter" line-numbers @mouseup="handleMouseUp"></PrismEditor>
                                </n-layout-content>
                                <n-layout-sider :width="90" content-style="padding: 5px;">
                                    <n-button @click="Add(panel)">Add§</n-button>
                                    <n-button @click="Clear(panel)" style="margin-top: 10px">Clear§</n-button>
                                </n-layout-sider>
                            </n-layout>

                        </n-card>
                    </n-tab-pane>
                    <n-tab-pane name="Payloads" display-directive="show:lazy">
                        <n-card title="Payload settings" size="large" >
                            <n-layout has-sider sider-placement="right">
                                <div>
                                    <div style="display: flex;" v-if="shouldRenderInputs" >
                                        <div v-for="index in payloadCount" :key="index" style="padding: 10px">
                                            <n-input v-model:value="payload[index]" type="textarea" placeholder="可以清除" round clearable style="text-align: left; width: 150px; height: 200px;"/>
                                            <div>
                                                <n-select v-model:value="selectedOption[index]" :options="optionsMapped" style="margin-top: 10px"/>
                                                <n-select v-model:value="selectedSubOption[index]" :options="subOptions[selectedOption[index]]" style="margin-top: 10px"/>
                                            </div>
                                        </div>
                                    </div>
                                    <div v-else>
                                        <n-input v-model:value="payloadSingle" type="textarea" placeholder="可以清除" round clearable style="text-align: left; width: 150px; height: 200px;"/>
                                        <n-select v-model:value="selectedOptionSingle" :options="optionsMapped" style="margin-top: 10px"/>
                                        <n-select v-model:value="selectedSubOptionSingle" :options="subOptions[selectedOptionSingle]" style="margin-top: 10px"/>
                                    </div>
                                </div>
                            </n-layout>
                        </n-card>
                    </n-tab-pane>

                    <n-tab-pane name="Attack" display-directive="show:lazy">
                        <n-button color="#ff6633" @click="startAattack(panel)" style="margin-bottom: 10px;">Start attack</n-button>
<!--                        <Attack :uuid="panel.uuid" :len="panel.len" :key="keyValue"/>-->
                        <Attack :uuid="panel.uuid" :len="panel.len"/>
                    </n-tab-pane>
                </n-tabs>
            </n-card>
        </n-tab-pane>
    </n-tabs>

</template>

<style>
.my-editor {
    font-size: 15px;
    line-height: 1.8;
    padding: 5px;
}
/* 行号不能正确处理自动换行 解决办法 https://github.com/koca/vue-prism-editor/issues/87#issuecomment-726228705 */
.prism-editor-wrapper .prism-editor__editor, .prism-editor-wrapper .prism-editor__textarea {
    white-space: pre !important;
}

.prism-editor__textarea {
    width: 999999px !important;
}
.prism-editor__editor {
    white-space: pre !important;
}
.prism-editor__container {
    overflow-x: scroll !important;
}

/* optional class for removing the outline */
.prism-editor__textarea:focus {
    outline: none;
}
</style>