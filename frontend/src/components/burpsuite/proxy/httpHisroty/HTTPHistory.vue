<template>
    <n-card >
        <n-data-table
                size="small"
                :columns="columns"
                :data="data"
                :row-props="rowProps"
                :max-height="300"
                :scroll-x="1800"
                :rowClassName="rowClassName"
                striped
        >
        </n-data-table>

        <n-dropdown
            placement="bottom-start"
            trigger="manual"
            :x="x"
            :y="y"
            :options="options"
            :show="showDropdown"
            :on-clickoutside="onClickOutSide"
            @select="handleSelect"
        />
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
import {NCard, NDataTable, useMessage} from "naive-ui";
import {ref, h, nextTick } from "vue";
import {EventsOn} from "../../../../../wailsjs/runtime/runtime.js";
import {GetHistoryDump, SendToRepeater} from "../../../../../wailsjs/go/main/App.js";


const message = useMessage();


// 右键

const x = ref(0);
const y = ref(0);
const showDropdown = ref(false);

const options = [
    {
        label: () => h("span", { style: { color: "green" } }, "Repeater"),
        key: "sr",
    },
    {
        label: () => h("span", { style: { color: "red" } }, "删除"),
        key: "delete"
    }
];

const rid = ref(0);

const handleSelect = (option) => {
    showDropdown.value = false;
    if (option === "sr") {
        SendToRepeater(rid.value);
        message.info("success");

    } else if (option === "edit") {
        // 打开编辑对话框
        // ...
    }
};

const onClickOutSide = () => {
    showDropdown.value = false;
};



const data = ref([]);

const request = ref('');
const response = ref('');

const rowId = ref('');

const rowProps = (row) => {
    return {
        style: "cursor: pointer;",
        onClick: () => {
            rowId.value = row.Id;
            GetHistoryDump(row.Id).then(HTTPBody => {
                request.value = HTTPBody["request"];
                response.value = HTTPBody["response"];
            });
        },
        onContextmenu: (e) => {
            e.preventDefault();
            showDropdown.value = false;
            nextTick().then(() => {
                showDropdown.value = true;
                x.value = e.clientX;
                y.value = e.clientY;
            });
            rid.value = row.Id;
        },
    };
};

const columns = [
    {
        title: "Id",
        key: "Id",
        resizable: true,
        width: 60,
        sorter: (row1, row2) => row1.Id - row2.Id
    },
    {
        title: "Host",
        key: "Host",
        resizable: true,
        width: 200,
        ellipsis: true
    },
    {
        title: "Method",
        key: "Method",
        resizable: true
    },
    {
        title: "Url",
        key: "Url",
        width: 200,
        resizable: true,
        ellipsis: true
    },
    {
        title: "Params",
        key: "Params",
        resizable: true,
        ellipsis: true
    },
    {
        title: "Edited",
        key: "Edited",
        resizable: true
    },
    {
        title: "Status",
        key: "Status",
        resizable: true
    },
    {
        title: "Length",
        key: "Length",
        resizable: true
    },
    {
        title: "MIMEType",
        key: "MIMEType",
        resizable: true,
        ellipsis: true
    },
    {
        title: "Extension",
        key: "Extension",
        resizable: true,
        ellipsis: true
    },
    {
        title: "Title",
        key: "Title",
        resizable: true,
        ellipsis: true
    },
    {
        title: "Comment",
        key: "Comment",
        resizable: true,
        ellipsis: true
    },
    {
        title: "TLS",
        key: "TLS",
        resizable: true,
        ellipsis: true
    },
    {
        title: "Ip",
        key: "Ip",
        resizable: true,
        ellipsis: true
    },
    {
        title: "Cookies",
        key: "Cookies",
        resizable: true,
        ellipsis: true
    },
    {
        title: "Time",
        key: "Time",
        resizable: true,
        ellipsis: true
    },
];


EventsOn("HttpHistory", HttpHistory => {
    data.value.push({
        Id: HttpHistory["id"],
        Host: HttpHistory["host"],
        Method: HttpHistory["method"],
        Url: HttpHistory["url"],
        Params: HttpHistory["params"],
        Edited: HttpHistory["edited"],
        Status: HttpHistory["status"],
        Length: HttpHistory["length"],
        MIMEType: HttpHistory["mime_type"],
        Extension: HttpHistory["extension"],
        Title : HttpHistory["title"],
        Comment: HttpHistory["comment"],
        TLS: HttpHistory["tls"],
        IP: HttpHistory["ip"],
        Cookies: HttpHistory["cookies"],
        Time: HttpHistory["time"],
    });
});


// todo 点击之后要等一会才会高亮，待优化
const rowClassName = (row) => {
    return row.Id === rowId.value ? 'table-tr-row-id' : '';
}

</script>

<style scoped>
:deep(.table-tr-row-id) td {
    background-color: #fffa65;
}
</style>