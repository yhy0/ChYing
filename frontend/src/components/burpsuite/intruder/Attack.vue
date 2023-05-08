<template>
    <n-card >
        <n-data-table
                size="small"
                :columns="columns"
                :data="data"
                :row-props="rowProps"
                :max-height="300"
                :rowClassName="rowClassName"
                striped
        >
        </n-data-table>
    </n-card>

    <n-card style="margin-bottom: 16px; margin-top: 10px">
        <n-grid :x-gap="12" :cols="2">
            <n-gi>
                <n-tabs type="line" animated >
                    <n-tab-pane name="request" style="width: 100%; overflow-x: auto;">
                        <n-code language="http" :code="request" show-line-numbers style="white-space: pre-wrap; text-align: left;" />
                    </n-tab-pane>
                </n-tabs>
            </n-gi>

            <n-gi>
                <n-tabs type="line" animated>
                    <n-tab-pane name="response" style="width: 100%; overflow-x: auto;">
                        <n-code language="http" :code="response" show-line-numbers style="white-space: pre-wrap; text-align: left; " />
                    </n-tab-pane>
                </n-tabs>
            </n-gi>
        </n-grid>
    </n-card>
</template>

<script setup>
import {NCard, NDataTable} from "naive-ui";
import {GetAttackDump} from "../../../../wailsjs/go/main/App.js";
import {ref, defineProps} from "vue";
import {EventsOn} from "../../../../wailsjs/runtime/runtime.js";

const data = ref([]);
const request = ref('');
const response = ref('');
const rowId = ref('');

const props = defineProps({
    uuid: {
        type: String,
        required: true
    },
    len: {
        type: Number,
        required: true
    }
})

console.log(props.uuid);

const rowProps = (row) => {
    return {
        style: "cursor: pointer;",
        onClick: () => {
            rowId.value = row.Id;
            GetAttackDump(props.uuid, row.Id).then(HTTPBody => {
                request.value = HTTPBody["request"];
                response.value = HTTPBody["response"];
            });
        },
    };
};

const Payload = [""];

const columns = [
    {
        title: "Request",
        key: "Id",
        resizable: true,
        width: 100,
        sorter: (row1, row2) => row1.Id - row2.Id
    },
    // 动态生成列定义对象
    ...Array.from({ length: props.len }, (_, index) => ({
        title: `Payload ${index + 1}`,
        key: `Payload${index}`,
        resizable: true,
        width: 100,
        ellipsis: true,
    })),
    {
        title: "Status",
        key: "Status",
        width: 80,
        resizable: true
    },
    {
        title: "Length",
        key: "Length",
        width: 80,
        resizable: true
    },
]

// todo 点击之后要等一会才会高亮，待优化
const rowClassName = (row) => {
    return row.Id === rowId.value ? 'table-tr-row-id' : '';
}

EventsOn(props.uuid, IntruderRes => {
    console.log(props.len);
    console.log(IntruderRes["payload"].length)
    data.value.push({
        Id: IntruderRes["id"],
        ...Array.from({ length: IntruderRes["payload"].length }, (_, index) => ({
            [`Payload${index}`]: IntruderRes["payload"][index],
        })),
        Status: IntruderRes["status"],
        Length: IntruderRes["length"],
    });
});

</script>

<style scoped>
:deep(.table-tr-row-id) td {
    background-color: #fffa65;
}
</style>