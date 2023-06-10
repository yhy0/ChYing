<script setup>
import {reactive} from 'vue'
import { useMessage} from 'naive-ui'
import {ShiroDecrypt} from "../../../wailsjs/go/main/App.js";

const message = useMessage()    // 使用这个组件外面一层必须使用包裹   <n-message-provider> </n-message-provider>

const shiro = reactive({
  data: "",
  key: "",
  iv: "",
  type: "",
  decrypt: "",
  decryptB64: "",
  serialization: "",
})

function decrypt() {
  ShiroDecrypt(shiro.key, shiro.data).then(result => {
    if(result === null) {
      message.error("失败");
      shiro.key = ""
      shiro.iv = ""
      shiro.type = ""
      shiro.decrypt = ""
      shiro.decryptB64 = ""
      shiro.serialization = ""
      return
    }
    message.success("Shiro Decrypt Success")
    shiro.key = result.key
    shiro.iv = result.iv
    shiro.type = result.type
    shiro.decrypt = result.decrypt
    shiro.decryptB64 = result.decryptB64
    shiro.serialization = result.serialization
  })
}

function copyText(id) {
  let text = ""
  if(id === 1) {
    text = shiro.decrypt
  } else if(id===2){
    text = shiro.decryptB64
  } else if(id===3){
    text = shiro.serialization
  }



  navigator.clipboard.writeText(text)
      .then(() => {
        message.info('Text copied to clipboard');
      })
      .catch((error) => {
        message.error('Failed to copy text: ', error);
      });
}
</script>

<template>
    <n-grid x-gap="12" :cols="2">
      <n-gi>
        <n-space vertical>
          <n-card title="Encoded" size="small">
            <n-input
                v-model:value="shiro.data"
                type="textarea"
                @input="decrypt"
                :autosize="{
                  minRows: 10,
                  maxRows: 10,
                }"
            />
          </n-card>
          <n-card title="key" size="small">
            <n-input v-model:value="shiro.key" type="text" placeholder="secret" @input="decrypt"/>
          </n-card>
          <n-card title="type" size="small">
            <n-code :code="shiro.type" type="text" placeholder="type"/>
          </n-card>
        </n-space>
      </n-gi>
      <n-gi>
        <n-card title="decrypt" size="small" style="margin-bottom: 16px; margin-top: 10px; height: 300px;overflow: auto;">
            <template #header-extra>
            <n-button
                class="edit-button"
                text
                @click="copyText(1)"
            >
              <template #icon>
                <n-icon size="18">
                  <svg
                      viewBox="0 0 544 560"
                      fill="none"
                      xmlns="http://www.w3.org/2000/svg"
                  >
                    <path
                        d="M399.503 143.667C399.319 152.501 406.332 159.812 415.167 159.997C424.001 160.181 431.312 153.168 431.497 144.333L399.503 143.667ZM416 120L431.997 120.333C431.999 120.207 432 120.081 432 119.954L416 120ZM360 64L360.046 48.0001C360.03 48 360.015 48 360 48L360 64ZM144 64V48C143.984 48 143.968 48 143.953 48.0001L144 64ZM80 128L64.0001 127.953C64 127.968 64 127.984 64 128L80 128ZM80 344H64C64 344.015 64 344.03 64.0001 344.046L80 344ZM136 400L135.954 416C135.97 416 135.985 416 136 416L136 400ZM160 416C168.837 416 176 408.837 176 400C176 391.163 168.837 384 160 384V416ZM217 160H439V128H217V160ZM439 160C461.644 160 480 178.356 480 201H512C512 160.683 479.317 128 439 128V160ZM480 201V423H512V201H480ZM480 423C480 445.644 461.644 464 439 464V496C479.317 496 512 463.317 512 423H480ZM439 464H217V496H439V464ZM217 464C194.356 464 176 445.644 176 423H144C144 463.317 176.683 496 217 496V464ZM176 423V201H144V423H176ZM176 201C176 178.356 194.356 160 217 160V128C176.683 128 144 160.683 144 201H176ZM431.497 144.333L431.997 120.333L400.003 119.667L399.503 143.667L431.497 144.333ZM432 119.954C431.946 100.888 424.347 82.6173 410.865 69.1349L388.238 91.7624C395.741 99.2658 399.97 109.434 400 120.046L432 119.954ZM410.865 69.1349C397.383 55.6526 379.112 48.0543 360.046 48.0001L359.954 79.9999C370.566 80.0301 380.734 84.2589 388.238 91.7624L410.865 69.1349ZM360 48H144V80H360V48ZM143.953 48.0001C122.767 48.0627 102.467 56.5064 87.4868 71.4868L110.114 94.1142C119.117 85.1118 131.316 80.0376 144.047 79.9999L143.953 48.0001ZM87.4868 71.4868C72.5064 86.4673 64.0627 106.767 64.0001 127.953L95.9999 128.047C96.0376 115.316 101.112 103.117 110.114 94.1142L87.4868 71.4868ZM64 128V344H96V128H64ZM64.0001 344.046C64.0543 363.112 71.6526 381.383 85.1349 394.865L107.762 372.238C100.259 364.734 96.0301 354.566 95.9999 343.954L64.0001 344.046ZM85.1349 394.865C98.6173 408.347 116.888 415.946 135.954 416L136.046 384C125.434 383.97 115.266 379.741 107.762 372.238L85.1349 394.865ZM136 416H160V384H136V416Z"
                        fill="currentColor"
                    />
                  </svg>
                </n-icon>
              </template>
            </n-button>
          </template>
            <n-code language="xml" :code="shiro.decrypt" word-wrap style="white-space: pre-wrap; text-align: left;"/>
        </n-card>

        <n-card title="decryptBase64" size="small" style="margin-bottom: 16px; margin-top: 10px; height: 120px;overflow: auto;">
          <template #header-extra>
            <n-button
                class="edit-button"
                text
                @click="copyText(2)"
            >
              <template #icon>
                <n-icon size="18">
                  <svg
                      viewBox="0 0 544 560"
                      fill="none"
                      xmlns="http://www.w3.org/2000/svg"
                  >
                    <path
                        d="M399.503 143.667C399.319 152.501 406.332 159.812 415.167 159.997C424.001 160.181 431.312 153.168 431.497 144.333L399.503 143.667ZM416 120L431.997 120.333C431.999 120.207 432 120.081 432 119.954L416 120ZM360 64L360.046 48.0001C360.03 48 360.015 48 360 48L360 64ZM144 64V48C143.984 48 143.968 48 143.953 48.0001L144 64ZM80 128L64.0001 127.953C64 127.968 64 127.984 64 128L80 128ZM80 344H64C64 344.015 64 344.03 64.0001 344.046L80 344ZM136 400L135.954 416C135.97 416 135.985 416 136 416L136 400ZM160 416C168.837 416 176 408.837 176 400C176 391.163 168.837 384 160 384V416ZM217 160H439V128H217V160ZM439 160C461.644 160 480 178.356 480 201H512C512 160.683 479.317 128 439 128V160ZM480 201V423H512V201H480ZM480 423C480 445.644 461.644 464 439 464V496C479.317 496 512 463.317 512 423H480ZM439 464H217V496H439V464ZM217 464C194.356 464 176 445.644 176 423H144C144 463.317 176.683 496 217 496V464ZM176 423V201H144V423H176ZM176 201C176 178.356 194.356 160 217 160V128C176.683 128 144 160.683 144 201H176ZM431.497 144.333L431.997 120.333L400.003 119.667L399.503 143.667L431.497 144.333ZM432 119.954C431.946 100.888 424.347 82.6173 410.865 69.1349L388.238 91.7624C395.741 99.2658 399.97 109.434 400 120.046L432 119.954ZM410.865 69.1349C397.383 55.6526 379.112 48.0543 360.046 48.0001L359.954 79.9999C370.566 80.0301 380.734 84.2589 388.238 91.7624L410.865 69.1349ZM360 48H144V80H360V48ZM143.953 48.0001C122.767 48.0627 102.467 56.5064 87.4868 71.4868L110.114 94.1142C119.117 85.1118 131.316 80.0376 144.047 79.9999L143.953 48.0001ZM87.4868 71.4868C72.5064 86.4673 64.0627 106.767 64.0001 127.953L95.9999 128.047C96.0376 115.316 101.112 103.117 110.114 94.1142L87.4868 71.4868ZM64 128V344H96V128H64ZM64.0001 344.046C64.0543 363.112 71.6526 381.383 85.1349 394.865L107.762 372.238C100.259 364.734 96.0301 354.566 95.9999 343.954L64.0001 344.046ZM85.1349 394.865C98.6173 408.347 116.888 415.946 135.954 416L136.046 384C125.434 383.97 115.266 379.741 107.762 372.238L85.1349 394.865ZM136 416H160V384H136V416Z"
                        fill="currentColor"
                    />
                  </svg>
                </n-icon>
              </template>
            </n-button>
          </template>
          <n-code language="java" :code="shiro.decryptB64" word-wrap style="white-space: pre-wrap; text-align: left;"/>
        </n-card>

        <n-card title="serialization" size="small" style="margin-bottom: 16px; margin-top: 10px; height: 300px;overflow: auto;">
          <template #header-extra>
            <n-button
                class="edit-button"
                text
                @click="copyText(3)"
            >
              <template #icon>
                <n-icon size="18">
                  <svg
                      viewBox="0 0 544 560"
                      fill="none"
                      xmlns="http://www.w3.org/2000/svg"
                  >
                    <path
                        d="M399.503 143.667C399.319 152.501 406.332 159.812 415.167 159.997C424.001 160.181 431.312 153.168 431.497 144.333L399.503 143.667ZM416 120L431.997 120.333C431.999 120.207 432 120.081 432 119.954L416 120ZM360 64L360.046 48.0001C360.03 48 360.015 48 360 48L360 64ZM144 64V48C143.984 48 143.968 48 143.953 48.0001L144 64ZM80 128L64.0001 127.953C64 127.968 64 127.984 64 128L80 128ZM80 344H64C64 344.015 64 344.03 64.0001 344.046L80 344ZM136 400L135.954 416C135.97 416 135.985 416 136 416L136 400ZM160 416C168.837 416 176 408.837 176 400C176 391.163 168.837 384 160 384V416ZM217 160H439V128H217V160ZM439 160C461.644 160 480 178.356 480 201H512C512 160.683 479.317 128 439 128V160ZM480 201V423H512V201H480ZM480 423C480 445.644 461.644 464 439 464V496C479.317 496 512 463.317 512 423H480ZM439 464H217V496H439V464ZM217 464C194.356 464 176 445.644 176 423H144C144 463.317 176.683 496 217 496V464ZM176 423V201H144V423H176ZM176 201C176 178.356 194.356 160 217 160V128C176.683 128 144 160.683 144 201H176ZM431.497 144.333L431.997 120.333L400.003 119.667L399.503 143.667L431.497 144.333ZM432 119.954C431.946 100.888 424.347 82.6173 410.865 69.1349L388.238 91.7624C395.741 99.2658 399.97 109.434 400 120.046L432 119.954ZM410.865 69.1349C397.383 55.6526 379.112 48.0543 360.046 48.0001L359.954 79.9999C370.566 80.0301 380.734 84.2589 388.238 91.7624L410.865 69.1349ZM360 48H144V80H360V48ZM143.953 48.0001C122.767 48.0627 102.467 56.5064 87.4868 71.4868L110.114 94.1142C119.117 85.1118 131.316 80.0376 144.047 79.9999L143.953 48.0001ZM87.4868 71.4868C72.5064 86.4673 64.0627 106.767 64.0001 127.953L95.9999 128.047C96.0376 115.316 101.112 103.117 110.114 94.1142L87.4868 71.4868ZM64 128V344H96V128H64ZM64.0001 344.046C64.0543 363.112 71.6526 381.383 85.1349 394.865L107.762 372.238C100.259 364.734 96.0301 354.566 95.9999 343.954L64.0001 344.046ZM85.1349 394.865C98.6173 408.347 116.888 415.946 135.954 416L136.046 384C125.434 383.97 115.266 379.741 107.762 372.238L85.1349 394.865ZM136 416H160V384H136V416Z"
                        fill="currentColor"
                    />
                  </svg>
                </n-icon>
              </template>
            </n-button>
          </template>
          <n-code language="xml" :code="shiro.serialization" word-wrap style="white-space: pre-wrap; text-align: left;"/>
        </n-card>
      </n-gi>
    </n-grid>

</template>

