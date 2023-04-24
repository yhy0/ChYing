<template>
    <n-card>
        <n-space vertical align="center">
            <n-space align="center">
                <n-input v-model:value="targetUrl" autosize placeholder="https://example.com/swagger-ui.html" />
                <n-button type="primary" @click="swagger">
                    解析
                </n-button>
            </n-space>
            <n-space align="center">
                <n-input ref="inputRef" default-value="http://127.0.0.1:8080" style="min-width: 50%" />
                <n-switch @update:value="handleCheckedChange" >
                    <template #checked>proxy</template>
                    <template #unchecked>unproxy</template>
                </n-switch>
                <n-tag size="small"> 使用 Burp 代理可以更好的获取展示 </n-tag>
            </n-space>
        </n-space>
    </n-card>
    <n-alert v-model:type="alertType" :bordered="false" style="margin-top: 10px">
        {{ alertContent }}
    </n-alert>

    <n-card style="margin-top: 10px">
        <n-space align="center">
            <n-input v-model:value="fcode" placeholder="Filter by code" @input="filterCode" />
            <n-input v-model:value="flength" placeholder="Filter by length" @input="filterLength" />
            <n-button @click="unfilter">
                Clear
            </n-button>
            <n-tag type="success" size="small">
                多个过滤条件以空格分开
            </n-tag>
        </n-space>

        <n-data-table
            size="small"
            :columns="columns"
            :data="data"
            @update:filters="handleUpdateFilter"
            :row-props="rowProps"
            :max-height="250"
            style="margin-top: 10px"
        >
        </n-data-table>
    </n-card>

    <n-card style="margin-bottom: 16px; margin-top: 10px">
        <n-grid :x-gap="12" :cols="2">
            <n-gi>
                <n-tabs type="line" animated >
                    <n-tab-pane name="request" style="width: 100%; overflow-x: auto;">
                        <n-code language="html" :code="request" show-line-numbers style="white-space: pre-wrap; text-align: left;" />
                    </n-tab-pane>
                </n-tabs>
            </n-gi>

            <n-gi>
                <n-tabs type="line" animated>
                    <n-tab-pane name="response" style="width: 100%; overflow-x: auto;">
                        <n-code language="html" :code="response" show-line-numbers style="white-space: pre-wrap; text-align: left; " />
                    </n-tab-pane>
                </n-tabs>
            </n-gi>
        </n-grid>
    </n-card>

</template>

<script setup>
import {NButton, NCard, NDataTable, useMessage} from "naive-ui";
import { ref, reactive } from "vue";
import {Proxy, Swagger} from '../../wailsjs/go/main/App'
import {EventsOn} from "../../wailsjs/runtime";

const checkedRef = ref(false);
const inputRef = ref(null);

const message = useMessage();

function handleCheckedChange(checked) {
    checkedRef.value = checked;
    if (checked) {
        const inputValue = inputRef.value.$el.querySelector("input").value;
        Proxy(inputValue).then(result => {
            if (result.Error !== "") {
                message.error(result.Msg + "; " + result.Error)
                return
            }
            message.success(result.Msg)
        })
    } else {
        Proxy("").then(result => {
            message.warning("代理关闭")
        })
    }
}

const targetUrl = ref('')
function swagger() {
    if(targetUrl.value !== "") {
        alertType.value = "info";
        alertContent.value = targetUrl.value + " 正在扫描中...";
        message.success(targetUrl.value + " 开始扫描");
        Swagger(targetUrl.value).then(result =>{
            alertType.value = "success";
            alertContent.value = targetUrl.value + " 扫描完成";
            message.success(targetUrl.value + " 扫描完成");
        });
    }
}

const alertType = ref('warning')
const alertContent = ref("没有任务")


const data = ref([]);

// table
const request = ref('');
const response = ref('');

const rowProps = (row) => {
    return {
        style: "cursor: pointer;",
        onClick: () => {
            request.value = row.request;
            response.value = row.response;

        }
    };
};

const fcode = ref(null);
const flength = ref(null);

const statusColumn = reactive({
    title: "Status",
    key: "statusCode",
    filterMultiple: false,
    filterOptionValue: null,
    sorter: "default",
    filter(value, row) {
        const nums = fcode.value.split(' ').map(num => parseInt(num));
        for (const num of nums) {
            if (row.statusCode == num) {
                return true;
            }
        }
        return false;
    }
});

const lengthColumn = reactive({
    title: "Length",
    key: "length",
    filterMultiple: true,
    filterOptionValue: null,
    sorter: "default",
    filter(value, row) {
        const nums = flength.value.split(' ').map(num => parseInt(num));
        for (const num of nums) {
            if (row.length == num) {
                return true;
            }
        }
        return false;
    }
});


const columns = [
    {
        title: "Url",
        key: "url",
    },
    statusColumn,
    lengthColumn,
];

EventsOn("swagger", e => {
    data.value.push({
        url: e.url,
        statusCode: e.status,
        length: e.length,
        request: e.request,
        response: e.response,
    });
});


const filterCode = () => {
    statusColumn.filterOptionValue = fcode.value;
};

const filterLength = () => {
    lengthColumn.filterOptionValue = flength.value;
};

const unfilter = () => {
    statusColumn.filterOptionValue = null;
    lengthColumn.filterOptionValue = null;
    fcode.value = "";
    flength.value = "";
};

const handleUpdateFilter = (filters, sourceColumn) => {
    statusColumn.filterOptionValue = filters[sourceColumn.key];
    fcode.value = filters[sourceColumn.key];

    lengthColumn.filterOptionValue = filters[sourceColumn.key];
    flength.value = filters[sourceColumn.key];
};

</script>
