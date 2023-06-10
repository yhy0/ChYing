<template>
  <n-card>
    <n-grid :x-gap="12" :y-gap="8" :cols="2">
      <n-grid-item>
        <n-space align="center">
          目标<n-input v-model:value="targetUrl" placeholder="https://example.com/" />
          Poc<n-cascader
              v-model:value="pocValue"
              :options="pocs">
          </n-cascader>
        </n-space>
        </n-grid-item>
      <n-grid-item>
        <n-space align="center">
          代理<n-input ref="inputRef" default-value="http://127.0.0.1:8080" style="min-width: 50%" />
          <n-switch @update:value="handleCheckedChange" >
            <template #checked>proxy</template>
            <template #unchecked>unproxy</template>
          </n-switch>
        </n-space>
      </n-grid-item>

      <n-grid-item>
        <n-space align="center">
          <n-button type="primary" @click="nuclei">Nuclei</n-button>
          <n-button type="warning" @click="reload">Reload</n-button>
        </n-space>
      </n-grid-item>
    </n-grid>

  </n-card>

  <n-card style="margin-top: 10px">
    <n-data-table
        size="small"
        :columns="columns"
        :data="data"
        :row-props="rowProps"
        :max-height="300"
        style="margin-top: 10px"
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
import { useMessage, NButton, NCard, NDataTable} from "naive-ui";
import {ref} from "vue";
import {NucleiY, NucleiLoad} from '../../../wailsjs/go/main/App'
import {EventsOn} from "../../../wailsjs/runtime";

const checkedRef = ref(false);
const inputRef = ref(null);

const message = useMessage();
const data = ref([]);

const targetUrl = ref("")
const proxy = ref("");
const pocValue = ref("");

const pocs = ref([]);

NucleiLoad().then(result =>{
  for (let i = 0; i < result.length; ++i) {
    const childrens = [];
    result[i].children.forEach((element) => {
      childrens.push({
        label: element,
        value: result[i].label + ":" +element,
      });
    });

    pocs.value.push({
      label: result[i].label,
      value: result[i].label + "-all",
      children: childrens,
    });
  }
})

function reload() {
  pocs.value = []
  NucleiLoad().then(result =>{
    for (let i = 0; i < result.length; ++i) {
      const childrens = [];
      result[i].children.forEach((element) => {
        childrens.push({
          label: element,
          value: result[i].label + ":" +element,
        });
      });

      pocs.value.push({
        label: result[i].label,
        value: result[i].label + "-all",
        children: childrens,
      });
    }
    message.info("reload success")
  })

}

function nuclei() {
  const target = targetUrl.value.trim();
  if(target !== "") {
    data.value = [];
    alertType.value = "info";
    alertContent.value = target + " 正在扫描中...";
    message.success(target + " 开始扫描");
    NucleiY(target.toString().trim(), pocValue.value, proxy.value.trim()).then(result =>{
      if(result.toString() === "") {
        message.success(targetUrl.value + " 扫描完成");
      } else {
        message.error(targetUrl.value + " 扫描失败 " + result.toString());
      }
    });
  }
}

function handleCheckedChange(checked) {
  checkedRef.value = checked;
  if (checked) {
    proxy.value = inputRef.value.$el.querySelector("input").value;
    message.info("代理开启");
  } else {
    proxy.value = "";
    message.warning("代理关闭");
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

const columns = [
  {
    title: "Url",
    key: "url",
  },
  {
    title: "Name",
    key: "name",
  },
]

EventsOn("nucleiYRes", e => {
  data.value.push({
    url: e.url,
    name: e.name,
    request: e.request,
    response: e.response,
  });
});

</script>
