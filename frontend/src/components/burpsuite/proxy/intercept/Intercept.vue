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

            <n-dropdown trigger="hover" :show-arrow="true" :options="options" @select="handleSelect">
                <n-button>Action</n-button>
            </n-dropdown>
        </div>

        <n-card>
            <n-code id="myCode" contenteditable language="http" :code="body" style="white-space: pre-wrap; text-align: left;" />
        </n-card>
    </n-space>

</template>

<script setup>
import {ref, h} from "vue";
import {NCard, useMessage} from "naive-ui";
import {Intercept, InterceptSend} from "../../../../../wailsjs/go/main/App.js";
import {EventsOn} from "../../../../../wailsjs/runtime/runtime.js";

const message = useMessage();

const options = [
    {
        label: () => h("span", { style: { color: "green" } }, "Repeater"),
        key: "repeater",
    },
    {
        label: () => h("span", { style: { color: "green" } }, "Intruder"),
        key: "intruder",
    },
];
const handleSelect = (option) => {
    if (option === "repeater") {

        if(body.value !== "") {
            message.info("Send To Repeater Success");
            InterceptSend("RepeaterBody")
        }

    } else if (option === "intruder") {
        if(body.value !== "") {
            message.info("Send To Intruder Success");
            InterceptSend("IntruderBody")
        }

    }
};


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