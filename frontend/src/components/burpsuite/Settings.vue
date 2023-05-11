<template>
    <n-form
        :model="formValue"
        :rules="rules"
        ref="formRef"
    >
        <n-form-item label="正在监听" path="port" >
            <n-input-group>
                <n-button type="primary">
                    Port
                </n-button>
                <n-input-number v-model:value="formValue.port">
                </n-input-number>
            </n-input-group>


        </n-form-item>
        <n-form-item>
            <n-button ghost type="error" @click="submitForm">
                Submit
            </n-button>
        </n-form-item>
    </n-form>

</template>

<script setup>
import {ref} from "vue";
import {useMessage} from "naive-ui";
import {GetProxyPort, Settings} from "../../../wailsjs/go/main/App.js";
import {EventsOn} from "../../../wailsjs/runtime/runtime.js";

const message = useMessage();
const formRef =ref()
const formValue = ref({
    port: 9080,
});

EventsOn("ProxyPort", result => {
    formValue.value.port = result
});

GetProxyPort().then((res)=> {
     formValue.value.port = res
})

const rules = {
    port: [
        { required: true, message: '请输入端口号' },
        {
            validator: (rule, value) => {
                if (isNaN(value)) {
                    return Promise.reject('端口号必须是数字');
                } else if (!Number.isInteger(value)) {
                    return Promise.reject('端口号必须是整数');
                } else if (value < 1 || value > 65535) {
                    return Promise.reject('端口号必须是1到65535之间的整数');
                } else {
                    return Promise.resolve();
                }
            },
            trigger: 'blur'
        }
    ]
};

const submitForm = () => {
    formRef.value?.validate((valid) => {
        if (!valid) {
            // 表单验证通过
            Settings(formValue.value.port.toString()).then((result) => {
                if(result === "") {
                    message.success('设置成功');
                } else {
                    message.error(result);
                }
            })

        } else {
            // 表单验证不通过
            message.error('请检查表单输入');
        }
    });
};

</script>

<style scoped>

</style>