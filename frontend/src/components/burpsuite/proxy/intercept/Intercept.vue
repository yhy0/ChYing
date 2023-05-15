<template>
    <n-space vertical>
        <div style="display: flex;">
            <n-button :disabled="disabled" style="margin-right: 16px;" @click="forward">
                Forward
            </n-button>
            <n-button :disabled="disabled" style="margin-right: 16px;">
                Drop
            </n-button>
            <n-button :type="type" style="margin-right: 16px;" id="start" @click="changeButtonType()">
                {{value}}
            </n-button>
        </div>

        <n-card>
            <n-code id="myCode" contenteditable language="http" :code="body" style="white-space: pre-wrap; text-align: left;" />
        </n-card>
    </n-space>

</template>

<script setup>
import {ref} from "vue";
import {Intercept} from "../../../../../wailsjs/go/main/App.js";
import {EventsOn} from "../../../../../wailsjs/runtime/runtime.js";

const body = ref('')

EventsOn("InterceptBody", result => {
    disabled.value = false
    body.value = result
});

const value = ref('Intercept is off')
const type = ref('primary')
const disabled = ref(true)

function changeButtonType() {
    let start = document.getElementById("start");
    // 第一次点击时 type === button
    if (start.getAttribute("type") === "button" || start.getAttribute("type") === "primary") {
        type.value = 'error'
        start.setAttribute("type", "error");
        value.value = "Intercept is on";
        if (body.value !== "") {
            disabled.value = false
        }

        Intercept(true, false, "")
    } else {
        type.value = 'primary'
        start.setAttribute("type", "primary");
        value.value = "Intercept is off";
        disabled.value = true
        Intercept(false, false, "")
        body.value = ""
    }
}


function forward() {
    let request = document.getElementById("myCode").textContent;

    console.log(request)
    Intercept(true,true, request).then((res=>{
        if(res === 0) {
            disabled.value = true
        }
    }))
}

</script>

<style scoped>

</style>