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
                    <n-checkbox :checked="DecodeUnicode[index]"  @update:checked="handleDecodeUnicode(index)">Unicode</n-checkbox>
                </n-col>
                <n-col span="2">
                    <n-checkbox :checked="DecodeURL[index]" @update:checked="handleDecodeURL(index)">URL</n-checkbox>
                </n-col>
                <n-col span="3">
                    <n-checkbox :checked="DecodeBase64[index]" @update:checked="handleDecodeBase64(index)">Base64</n-checkbox>
                </n-col>
                <n-col span="2">
                    <n-checkbox :checked="DecodeHex[index]" @update:checked="handleDecodeHex(index)">Hex</n-checkbox>
                </n-col>
                <n-col span="2">
                    <n-checkbox :checked="DecodeMD5[index]" @update:checked="handleMD5(index)">MD5</n-checkbox>
                </n-col>
            </n-row>
            <n-row span="20">
                <n-col span="2">
                    Encode
                </n-col>
                <n-col span="3">
                    <n-checkbox :checked="EncodeUnicode[index]" @update:checked="handleEncodeUnicode(index)">Unicode</n-checkbox>
                </n-col>
                <n-col span="2">
                    <n-checkbox :checked="EncodeURL[index]" @update:checked="handleEncodeURL(index)">URL</n-checkbox>
                </n-col>
                <n-col span="3">
                    <n-checkbox :checked="EncodeBase64[index]" @update:checked="handleEncodeBase64(index)">Base64</n-checkbox>
                </n-col>
                <n-col span="2">
                    <n-checkbox :checked="EncodeHex[index]" @update:checked="handleEncodeHex(index)">Hex</n-checkbox>
                </n-col>
            </n-row>
        </n-row>
    </n-space>
</template>

<script setup>
import { ref } from 'vue'
import {Decoder} from '../../wailsjs/go/main/App'
const values = ref([''])

// 这些是为了选中一个，取消其他的选中，更美观
const DecodeUnicode = ref([false])
const DecodeURL = ref([false])
const DecodeBase64 = ref([false])
const DecodeHex = ref([false])
const DecodeMD5 = ref([false])
const EncodeUnicode = ref([false])
const EncodeURL = ref([false])
const EncodeBase64 = ref([false])
const EncodeHex = ref([false])


function addInput(index) {
    const lastIndex = values.value.length - 1
    const lastValue = values.value[lastIndex]
    if (lastValue !== '') {
        values.value.push('')
        DecodeUnicode.value.push(false)
        DecodeURL.value.push(false)
        DecodeBase64.value.push(false)
        DecodeHex.value.push(false)
        DecodeMD5.value.push(false)
        EncodeUnicode.value.push(false)
        EncodeURL.value.push(false)
        EncodeBase64.value.push(false)
        EncodeHex.value.push(false)
    }
    if (index >= 0) {
        values.value.splice(index + 2)
        DecodeUnicode.value.splice(index + 2)
        DecodeURL.value.splice(index + 2)
        DecodeBase64.value.splice(index + 2)
        DecodeHex.value.splice(index + 2)
        DecodeMD5.value.splice(index + 2)
        EncodeUnicode.value.splice(index + 2)
        EncodeURL.value.splice(index + 2)
        EncodeBase64.value.splice(index + 2)
        EncodeHex.value.splice(index + 2)
    }
}

function handleDecodeUnicode(index, event) {
    DecodeUnicode.value[index] = true
    Decoder(values.value[index],"DecodeUnicode").then(result => {
        values.value[index+1] = result
    })
    values.value.splice(index + 2)
    clearOtherChecks("DecodeUnicode",index)
}

function handleEncodeUnicode(index) {
    EncodeUnicode.value[index] = true
    Decoder(values.value[index],"EncodeUnicode").then(result => {
        values.value[index+1] = result
    })
    values.value.splice(index + 2)
    clearOtherChecks("EncodeUnicode",index)
    console.log(DecodeURL.value[index])
    console.log(DecodeURL)
}

function handleDecodeURL(index) {
    DecodeURL.value[index] = true
    Decoder(values.value[index],"DecodeURL").then(result => {
        values.value[index+1] = result
    })
    values.value.splice(index + 2)
    clearOtherChecks("DecodeURL",index)
    console.log(EncodeUnicode.value[index])
}

function handleEncodeURL(index) {
    EncodeURL.value[index] = true
    Decoder(values.value[index],"EncodeURL").then(result => {
        values.value[index+1] = result
    })
    values.value.splice(index + 2)
    clearOtherChecks("EncodeURL",index)
}

function handleDecodeBase64(index) {
    DecodeBase64.value[index] = true
    clearOtherChecks("DecodeBase64",index)
    Decoder(values.value[index],"DecodeBase64").then(result => {
        values.value[index+1] = result
    })
    values.value.splice(index + 2)
}

function handleEncodeBase64(index) {
    EncodeBase64.value[index] = true
    clearOtherChecks("EncodeBase64",index)
    Decoder(values.value[index],"EncodeBase64").then(result => {
        values.value[index+1] = result
    })
    values.value.splice(index + 2)
}

function handleDecodeHex(index) {
    DecodeHex.value[index] = true
    clearOtherChecks("DecodeHex",index)
    Decoder(values.value[index],"DecodeHex").then(result => {
        values.value[index+1] = result
    })
    values.value.splice(index + 2)
}

function handleEncodeHex(index) {
    EncodeHex.value[index] = true
    clearOtherChecks("EncodeHex",index)
    Decoder(values.value[index],"EncodeHex").then(result => {
        values.value[index+1] = result
    })
    values.value.splice(index + 2)
}

function handleMD5(index) {
    DecodeMD5.value[index] = true
    clearOtherChecks("DecodeMD5",index)
    Decoder(values.value[index],"MD5").then(result => {
        values.value[index+1] = result
    })
    values.value.splice(index + 2)
}

// 去除其他的选中效果
function clearOtherChecks(exclude, index) {
    const checks = [
        'DecodeUnicode',
        'DecodeURL',
        'DecodeBase64',
        'DecodeHex',
        'EncodeUnicode',
        'EncodeURL',
        'EncodeBase64',
        'EncodeHex',
        'DecodeMD5'
    ]

    for (const check of checks) {
        if (check !== exclude) {
            eval(check + '.value[index] = false')
            eval(check + '.value.splice(index + 2)')
        }
    }
}


</script>

<style scoped>
</style>