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
        <n-form-item style="display: flex;">
            <n-card title="Exclude" :bordered="false">
                <n-input v-model:value="formValue.exclude" type="textarea" status="error" placeholder="可以清除" round clearable style="width: 200px; height: 200px;"/>
            </n-card>

            <n-card title="Include" :bordered="false">
                <n-input v-model:value="formValue.include" type="textarea" status="warning" placeholder="可以清除" round clearable style="width: 200px; height: 200px;"/>
            </n-card>
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
import {GetBurpSettings, Settings} from "../../../wailsjs/go/main/App.js";
import {EventsOn} from "../../../wailsjs/runtime/runtime.js";

const message = useMessage();
const formRef =ref()
const formValue = ref({
    port: 9080,
    exclude: '',
    include: '',
});

EventsOn("ProxyPort", result => {
    formValue.value.port = result
});

EventsOn("Exclude", result => {
    formValue.value.exclude = result
    console.log(formValue.value.exclude)
});
EventsOn("Include", result => {
    formValue.value.include = result
});

GetBurpSettings().then((res)=> {
    formValue.value.port = res.port;
    for (let i = 0; i < res.exclude.length; i++) {
        formValue.value.exclude += res.exclude[i] + "\r\n"
    }

    for (let i = 0; i < res.include.length; i++) {
        formValue.value.include += res.include[i] + "\r\n"
    }

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
            console.log(formValue.value)
            Settings(formValue.value).then((result) => {
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