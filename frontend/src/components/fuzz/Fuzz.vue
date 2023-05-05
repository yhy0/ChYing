<template>
    <n-card>
        <n-space>
            <n-space>
                <n-form ref="formRef" inline label-placement="left" :model="formValue" size="small">
                    <n-form-item label="目标" required>
                        <n-input v-model:value="formValue.targetUrl" placeholder="https://example.com/" />
                    </n-form-item>
                    <n-form-item :span="12">
                        <n-checkbox-group v-model:value="formValue.checkboxGroupValue">
                            <n-space>
                                <n-checkbox value="jsp">jsp</n-checkbox>
                                <n-checkbox value="php">php</n-checkbox>
                                <n-checkbox value="asp">asp</n-checkbox>
                                <n-checkbox value="aspx">aspx</n-checkbox>
                                <n-checkbox value="bbscan">bbscan</n-checkbox>
                            </n-space>
                        </n-checkbox-group>
<!--                        <n-input v-model:value="formValue.path" placeholder="file path" />-->
                    </n-form-item>

                </n-form>

            </n-space>
            <n-space align="center">
                <n-input ref="inputRef" default-value="http://127.0.0.1:8080" style="min-width: 50%" />
                <n-switch @update:value="handleCheckedChange" >
                    <template #checked>proxy</template>
                    <template #unchecked>unproxy</template>
                </n-switch>
                <n-button secondary style="float: right" strong @click="showAdvanced = true">高级配置</n-button>
            </n-space>

            <n-button type="primary" @click="fuzz">Fuzz</n-button>

            <n-button type="error" @click="fuzzStop">Stop</n-button>
        </n-space>

    </n-card>
    <n-progress
        type="line"
        :percentage="percentage"
        :indicator-placement="'inside'"
        processing
    />

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


    <n-modal v-model:show="showAdvanced"
             preset="dialog"
             title="高级配置"
             :show-icon="false"
             style="width: 500px"
    >
        <n-button>aaa</n-button>
    </n-modal>
</template>

<script setup>
import { useMessage, NButton, NCard, NDataTable} from "naive-ui";
import { ref, reactive } from "vue";
import {Proxy, Fuzz, FuzzStop} from '../../../wailsjs/go/main/App'
import {EventsOn} from "../../../wailsjs/runtime";

const checkedRef = ref(false);
const inputRef = ref(null);
const showAdvanced = ref(false);

const message = useMessage();

const data = ref([]);

const formValue = ref({
    targetUrl: "",
    path: "",
    checkboxGroupValue: null,
});

function fuzz() {
    const target = formValue.value.targetUrl.trim();
    if(target !== "") {
        data.value = [];
        alertType.value = "info";
        alertContent.value = target + " 正在扫描中...";
        message.success(target + " 开始扫描");
        Fuzz(target, formValue.value.checkboxGroupValue, formValue.value.path.trim()).then(result =>{
            console.log(result);
            if(result === "") {
                alertType.value = "success";
                alertContent.value = target + " 扫描完成";
                message.success(target + " 扫描完成");
            } else {
                message.error(target + result);
            }

        });
    }
}

function fuzzStop() {
    FuzzStop().then(result => {
        message.success(formValue.value.targetUrl.trim() + " 扫描已停止");
    })
}

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


const alertType = ref('warning')
const alertContent = ref("没有任务")

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

EventsOn("Fuzz", e => {
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

const percentage = ref(0);

EventsOn("FuzzPercentage", Percentage => {
    percentage.value = Percentage
});


</script>
