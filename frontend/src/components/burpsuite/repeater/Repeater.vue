<script setup>
import {computed, ref} from "vue";
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
    value.value = panels.value[Math.min(index, panels.value.length - 1)].key;
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

function send(panel) {
    // console.log(document.getElementById("myCodeR"))
    // request.value = document.getElementById("myCodeR").textContent;
    console.log(request.value)
    if (request.value === "") {
        request.value = panel.req;
    }

    Raw(request.value, panel.url, panel.id).then(result=>{
        panel.req = result.request;
        panel.res = result.response;
        panel.id = result.uuid;
    })
}

// import Prism Editor
import { PrismEditor } from 'vue-prism-editor';
import 'vue-prism-editor/dist/prismeditor.min.css'; // import the styles somewhere

// import highlighting library (you can use any library you want just return html string)
import { highlight, languages } from 'prismjs/components/prism-core';
import 'prismjs/components/prism-clike';
import 'prismjs/components/prism-http.js';
import 'prismjs/themes/prism.css'; // import syntax highlighting styles


function highlighter(code) {
    request.value = code.toString()
    return highlight(code, Prism.languages.http,'http'); // languages.<insert language> to return html with markup
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
                            <n-tab-pane name="Request" style="width: 100%; overflow-x: auto;">
                                <prism-editor class="my-editor" id="myCodeR" v-model="panel.req" :highlight="highlighter" line-numbers></prism-editor>
                            </n-tab-pane>
                        </n-tabs>
                    </n-gi>

                    <n-gi>
                        <n-tabs type="line" animated>
                            <n-tab-pane name="Response" style="width: 100%; overflow-x: auto;">
                                <n-code language="http" word-wrap :code="panel.res" show-line-numbers style="white-space: pre-wrap; text-align: left; " />
                            </n-tab-pane>
                        </n-tabs>
                    </n-gi>
                </n-grid>
            </n-card>
        </n-tab-pane>
    </n-tabs>
</template>

<style scoped>
.my-editor {
    line-height: 1.8;
    padding: 5px;
    white-space: nowrap;
}
/* optional class for removing the outline */
.prism-editor__textarea:focus {
    outline: none;
}
</style>