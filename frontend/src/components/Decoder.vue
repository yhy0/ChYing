<template>
    <n-space vertical>
        <n-row v-for="(input, index) in values" :key="index">
            <n-col span="24">
                <n-input
                    v-model:value="values[index]"
                    type="textarea"
                    @input="addInput(index)"
                />
            </n-col>
            <n-row span="20">
                <n-col span="2">
                    Decode
                </n-col>
                <n-col span="3">
                    <n-checkbox @update:checked="handleDecodeUnicode(index)">Unicode</n-checkbox>
                </n-col>
                <n-col span="2">
                    <n-checkbox @update:checked="handleDecodeURL(index)">URL</n-checkbox>
                </n-col>
                <n-col span="3">
                    <n-checkbox @update:checked="handleDecodeBase64(index)">Base64</n-checkbox>
                </n-col>
                <n-col span="2">
                    <n-checkbox @update:checked="handleMD5(index)">MD5</n-checkbox>
                </n-col>
            </n-row>
            <n-row span="20">
                <n-col span="2">
                    Encode
                </n-col>
                <n-col span="3">
                    <n-checkbox @update:checked="handleEncodeUnicode(index)">Unicode</n-checkbox>
                </n-col>
                <n-col span="2">
                    <n-checkbox @update:checked="handleEncodeURL(index)">URL</n-checkbox>
                </n-col>
                <n-col span="3">
                    <n-checkbox @update:checked="handleEncodeBase64(index)">Base64</n-checkbox>
                </n-col>
            </n-row>
        </n-row>
    </n-space>
</template>

<script setup>
import { ref } from 'vue'
import {Decoder} from '../../wailsjs/go/main/App'

const values = ref([''])

function addInput(index) {
    const lastIndex = values.value.length - 1
    const lastValue = values.value[lastIndex]
    if (lastValue !== '') {
        values.value.push('')
    }
    if (index >= 0) {
        values.value.splice(index + 2)
    }
}

function handleDecodeUnicode(index) {
    Decoder(values.value[index],"DecodeUnicode").then(result => {
        values.value[index+1] = result
    })
}

function handleEncodeUnicode(index) {
    Decoder(values.value[index],"EncodeUnicode").then(result => {
        values.value[index+1] = result
    })
}

function handleDecodeURL(index) {
    Decoder(values.value[index],"DecodeURL").then(result => {
        values.value[index+1] = result
    })
}

function handleEncodeURL(index) {
    Decoder(values.value[index],"EncodeURL").then(result => {
        values.value[index+1] = result
    })
}

function handleDecodeBase64(index) {
    Decoder(values.value[index],"DecodeBase64").then(result => {
        values.value[index+1] = result
    })
}

function handleEncodeBase64(index) {
    Decoder(values.value[index],"EncodeBase64").then(result => {
        values.value[index+1] = result
    })
}

function handleMD5(index) {
    Decoder(values.value[index],"MD5").then(result => {
        values.value[index+1] = result
    })
}
</script>