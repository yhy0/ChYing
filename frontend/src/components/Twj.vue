<script setup>
import {reactive} from 'vue'
import {NButton, useMessage} from 'naive-ui'
import {Parser, Verify, Brute, TwjStop} from '../../wailsjs/go/main/App'
import {EventsOn} from "../../wailsjs/runtime";

const message = useMessage()    // 使用这个组件外面一层必须使用包裹   <n-message-provider> </n-message-provider>

const twj = reactive({
  jwt: "",
  secret:"",
  header: "",
  payload: "",
  message: "",
  signature: "",
  percentage: 0,
})

function parser() {
  Parser(twj.jwt).then(result => {
    if(result.header === "") {
      message.error("格式错误: " + result.message)
      return
    }
    message.success("JWT Parser Success")
    twj.header = JSON.stringify(JSON.parse(result.header),null, 2);
    twj.payload = JSON.stringify(JSON.parse(result.payload),null, 2);
    twj.message = result.message
    twj.signature = result.signature
  })
}

function verify() {
  Verify(twj.jwt, twj.secret).then(result => {
    if(result.error !== "") {
      message.error(result.error)
      return
    }
    message.success("Signature Verified")
    twj.signature = JSON.stringify(JSON.parse(result.msg),null, 2);
  })
}

function brute() {
    Brute().then(result => {
        if(result !== "") {
            message.success("brute success")
            twj.secret = result
            Verify(twj.jwt, twj.secret).then(result => {
                if(result.error !== "") {
                    message.error(result.error)
                    return
                }
                message.success("Signature Verified")
                twj.signature = JSON.stringify(JSON.parse(result.msg),null, 2);
            })
        }
    })
}

function stopBrute() {
    TwjStop().then(result => {
        message.success("爆破已停止");
    })
}


EventsOn("Percentage", Percentage => {
    twj.percentage = Percentage
});

</script>

<template>
    <n-grid x-gap="12" :cols="2">
      <n-gi>
        <n-space vertical>
          <n-card :bordered="false" title="Encoded" size="small">
            <n-input
                v-model:value="twj.jwt"
                type="textarea"
                @input="parser"
                :autosize="{
                  minRows: 10,
                  maxRows: 10,
                }"
            />
          </n-card>

          <n-card :bordered="false" title="Secret" size="small">
            <n-input v-model:value="twj.secret" type="text" placeholder="secret" @input="verify" />
          </n-card>
        </n-space>
      </n-gi>
      <n-gi>
        <n-card title="Header" size="small" style="margin-bottom: 16px; margin-top: 10px">
            <n-code id="header" language="json" :code="twj.header" word-wrap style="white-space: pre-wrap; text-align: left;"/>
        </n-card>

        <n-card title="Payload" size="small" style="margin-bottom: 16px; margin-top: 10px">
            <n-code id="payload" language="json" :code="twj.payload" word-wrap style="white-space: pre-wrap; text-align: left;"/>
        </n-card>

        <n-card title="Verify" size="small" style="margin-bottom: 16px; margin-top: 10px">
            <n-code id="signature" language="json" :code="twj.signature" word-wrap style="white-space: pre-wrap; text-align: left;"/>
        </n-card>

      </n-gi>
    </n-grid>


    <n-button type="primary" @click="brute" style="margin-right: 20px; ">
        Brute
    </n-button>

    <n-button type="error" @click="stopBrute" style="margin-right: 20px; ">
        Stop
    </n-button>


    <p></p>
    <n-progress
            type="line"
            :percentage="twj.percentage"
            :indicator-placement="'inside'"
            processing
    />

    <div style="padding: 5px">
        <span v-if="twj.secret !== ''">
            <n-tag type="success">
              {{ twj.secret }}
            </n-tag>
        </span>
    </div>
</template>

