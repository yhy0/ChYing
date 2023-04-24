<script setup>
import {reactive, watch} from 'vue'
import { useMessage} from 'naive-ui'
import {Parser, Verify, Brute} from '../../wailsjs/go/main/App'
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
    twj.header = result.header
    twj.payload = result.payload
    twj.message = result.message
    twj.signature = result.signatureStr
  })
}

function verify() {
  Verify(twj.jwt, twj.secret).then(result => {
    if(result.Error !== "") {
      message.error(result.Error)
      return
    }
    message.success("Signature Verified")
    twj.signature = result.Msg
  })
}


function brute() {
    Brute().then(result => {
        if(result !== "") {
            message.success("brute success")
            twj.secret = result
        }
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

      <n-card :bordered="false" title="Header" size="small">
        <n-input v-model:value="twj.header" type="textarea" :autosize="{minRows: 1}"/>
      </n-card>

      <n-card :bordered="false" title="Payload" size="small">
          <n-input v-model:value="twj.payload" type="textarea" :autosize="{minRows: 3}"/>
      </n-card>

      <n-card :bordered="false" title="Verify" size="small">
        <n-input v-model:value="twj.signature" type="textarea" :autosize="{minRows: 3}"/>
      </n-card>

    </n-gi>
  </n-grid>

    <n-button type="primary" @click="brute">
        Brute
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

