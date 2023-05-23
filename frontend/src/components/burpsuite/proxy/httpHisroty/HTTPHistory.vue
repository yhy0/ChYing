<template>
    <n-card >
        <n-collapse>
            <n-collapse-item title="Filter">
                <n-grid x-gap="12" :cols="2">
                    <n-gi>
                        <n-card title="Filter by MIME type">
                            <n-checkbox-group v-model:value="filter.mime" @update:value="handleCheckedMime">
                                <n-space >
                                    <n-checkbox value="HTML">HTML</n-checkbox>
                                    <n-checkbox value="Script">Script</n-checkbox>
                                    <n-checkbox value="XML">XML</n-checkbox>
                                    <n-checkbox value="CSS">CSS</n-checkbox>
                                    <n-checkbox value="Other text">Other text</n-checkbox>
                                    <n-checkbox value="Images">Images</n-checkbox>
                                    <n-checkbox value="Flash">Flash</n-checkbox>
                                    <n-checkbox value="Other binary">Other binary</n-checkbox>
                                </n-space>
                            </n-checkbox-group>
                        </n-card>
                    </n-gi>
                    <n-gi>
                        <n-card title="Filter by status type">
                            <n-checkbox-group v-model:value="filter.status" @update:value="handleCheckedStatus">
                                <n-space>
                                    <n-checkbox value="2xx">2xx</n-checkbox>
                                    <n-checkbox value="3xx">3xx</n-checkbox>
                                    <n-checkbox value="4xx">4xx</n-checkbox>
                                    <n-checkbox value="5xx">5xx</n-checkbox>
                                </n-space>
                            </n-checkbox-group>
                        </n-card>
                    </n-gi>
                </n-grid>

                <n-space align="center" style="margin-top: 10px">
                    <n-input v-model:value="filter.ext" placeholder="Filter by file extension" @input="filterExt" />
                    <n-input v-model:value="filter.term" placeholder="Filter by search term" @input="filterTerm" />
                    <n-tag type="success" size="small">
                        多个过滤条件以空格分开
                    </n-tag>
                </n-space>

            </n-collapse-item>
        </n-collapse>

        <n-data-table
                size="small"
                :columns="columns"
                :data="data"
                :row-props="rowProps"
                :max-height="300"
                :scroll-x="1800"
                :rowClassName="rowClassName"
                @update:filters="handleUpdateFilter"
                striped
                style="margin-bottom: 16px; margin-top: 10px"
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
import {ref, h, nextTick, reactive} from "vue";
import {EventsOn} from "../../../../../wailsjs/runtime/runtime.js";
import {GetHistoryDump, SendToRepeater, SendToIntruder} from "../../../../../wailsjs/go/main/App.js";

const message = useMessage();

// 右键
const x = ref(0);
const y = ref(0);
const showDropdown = ref(false);

const options = [
    {
        label: () => h("span", { style: { color: "green" } }, "Repeater"),
        key: "repeater",
    },
    {
        label: () => h("span", { style: { color: "green" } }, "Intruder"),
        key: "intruder",
    },
    {
        label: () => h("span", { style: { color: "red" } }, "Clear all"),
        key: "clear"
    }
];

const rid = ref(0);
const data = ref([]);
const request = ref('');
const response = ref('');
const rowId = ref('');

const handleSelect = (option) => {
    showDropdown.value = false;
    if (option === "repeater") {
        SendToRepeater(rid.value);
        message.info("Send To Repeater Success");
    } else if (option === "intruder") {
        SendToIntruder(rid.value);
        message.info("Send To Intruder Success");
    } else if (option === "clear") {
        data.value = []
        request.value = ''
        response.value = ''
        message.info("Clear All Success");
    }
};

const onClickOutSide = () => {
    showDropdown.value = false;
};

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

const StatusColumn = reactive({
    title: "Status",
    key: "Status",
    filterMultiple: true,
    filterOptionValue: null,
    resizable: true,
    sorter: "default",
    filter(value, row) {
        const status = parseInt(row.Status)
        if(filter.value.status.includes('2xx')) {
            if(status >= 200 && status < 300) {
                return true;
            }
        } else if(filter.value.status.includes('3xx')) {
            if(status >= 300 && status < 400) {
                return true;
            }
        } else if(filter.value.status.includes('4xx')) {
            if(status >= 400 && status < 500) {
                return true;
            }
        } else if(filter.value.status.includes('5xx')) {
            if(status >= 500) {
                return true;
            }
        }
        return false;
    }
});

const MIMEColumn = reactive({
    title: "MIMEType",
    key: "MIMEType",
    filterMultiple: true,
    filterOptionValue: null,
    resizable: true,
    ellipsis: {
        tooltip: true
    },
    sorter: "default",
    filter(value, row) {
        const mime = parseInt(row.MIMEType)
        if(filter.value.mime.includes('HTML')) {
            if(mime === "html") {
                return true;
            }
        } else if(filter.value.mime.includes('Script')) {
            if(mime === "script" || mime === "json") {
                return true;
            }
        } else if(filter.value.mime.includes('XML')) {
            if(mime === "xml") {
                return true;
            }
        } else if(filter.value.mime.includes('css')) {
            if(mime === "CSS") {
                return true;
            }
        } else if(filter.value.mime.includes('Other text')) {
            if(mime === "text" || mime === "") {
                return true;
            }
        } else if(filter.value.mime.includes('Images')) {
            if(mime === "image") {
                return true;
            }
        } else if(filter.value.mime.includes('Flash')) {
            if(mime === "Flash") {
                return true;
            }
        } else if(filter.value.mime.includes('Other binary')) {
            if(mime === "Other binary") {
                return true;
            }
        }
        return false;
    }
});

const ExtColumn = reactive({
    title: "Extension",
    key: "Extension",
    filterMultiple: true,
    filterOptionValue: null,
    resizable: true,
    ellipsis: {
        tooltip: true
    },
    sorter: "default",
    filter(value, row) {
        const exts = filter.value.ext.split(' ');
        for (const ext of exts) {
            if (row.Extension === ext) {
                return true;
            }
        }
        return false;
    }
});


// filter
const filter = ref({
    status: ["3xx", "4xx", "5xx", "2xx"],
    mime: ["HTML", "Script", "XML", "Other text"],
    ext: null,
    term: null,
});


function handleCheckedStatus(checked) {
    filter.value.status = checked;
    StatusColumn.filterOptionValue = checked;
}

function handleCheckedMime(checked) {
    filter.value.mime = checked;
    MIMEColumn.filterOptionValue = checked;
}

const filterExt = () => {
    ExtColumn.filterOptionValue = filter.value.ext;
};

const filterTerm = () => {
    ExtColumn.filterOptionValue = filter.value.term;
};

const handleUpdateFilter = (filters, sourceColumn) => {
    StatusColumn.filterOptionValue = filters[sourceColumn.key];
    filter.value.status = filters[sourceColumn.key];

    MIMEColumn.filterOptionValue = filters[sourceColumn.key];
    filter.value.mime = filters[sourceColumn.key];

    ExtColumn.filterOptionValue = filters[sourceColumn.key];
    filter.value.ext = filters[sourceColumn.key];
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
        ellipsis: {
            tooltip: true
        }
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
        ellipsis: {
            tooltip: true
        }
    },
    {
        title: "Params",
        key: "Params",
        resizable: true,
        ellipsis: {
            tooltip: true
        }
    },
    {
        title: "Edited",
        key: "Edited",
        resizable: true
    },
    StatusColumn,
    {
        title: "Length",
        key: "Length",
        resizable: true
    },
    MIMEColumn,
    ExtColumn,
    {
        title: "Title",
        key: "Title",
        resizable: true,
        ellipsis: {
            tooltip: true
        }
    },
    {
        title: "Comment",
        key: "Comment",
        resizable: true,
        ellipsis: {
            tooltip: true
        }
    },
    {
        title: "TLS",
        key: "TLS",
        resizable: true,
        ellipsis: {
            tooltip: true
        }
    },
    {
        title: "Ip",
        key: "Ip",
        resizable: true,
        ellipsis: {
            tooltip: true
        }
    },
    {
        title: "Cookies",
        key: "Cookies",
        resizable: true,
        ellipsis: {
            tooltip: true
        }
    },
    {
        title: "Time",
        key: "Time",
        resizable: true,
        ellipsis: {
            tooltip: true
        }
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
    background-color: rgba(255, 165, 101, 0.98);
}
</style>