<template>
    <n-space vertical>
        <n-input
            v-model:value="values[0]"
            type="textarea"
            placeholder="基本的 Textarea"
            @input="addInput(0)"
        />
        <n-input
            v-for="(value, index) in values.slice(1)"
            :key="index"
            v-model:value="values[index + 1]"
            type="textarea"
            :placeholder="`第 ${index + 1} 个 Textarea`"
            @input="addInput(index + 1)"
        />
    </n-space>
</template>

<script setup>
import { ref } from 'vue'

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
</script>