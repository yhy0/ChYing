<script setup>
import {ref} from "vue";
import {TaskList} from "../../../wailsjs/go/main/App.js";

const columns = [
    {
        title: "进程名字",
        key: "process"
    },
    {
        title: "杀软名称",
        key: "av"
    },
];

const data = ref([]);
const value = ref('')

function submit() {
    data.value = [];
    TaskList(value.value).then((res) => {
        for (const k in res) {
            data.value.push({
                process:k,
                av: res[k],
            })
        }
    })
}

</script>

<template>
    <n-space vertical>
        <div v-if="data.length">
        <n-data-table :columns="columns" :data="data" striped/>
        </div>
        <n-card title="tasklist /SVC" header-style="color:'blue';" hoverable content-style="height: 600px; overflow-y: auto;">
                <n-input
                    v-model:value="value"
                    type="textarea"
                    size="large"
                    clearable
                    style="min-height: 50%"
                />
        </n-card>
        <n-button ghost type="error" @click="submit">
            Submit
        </n-button>
    </n-space>
</template>

<style scoped>

</style>