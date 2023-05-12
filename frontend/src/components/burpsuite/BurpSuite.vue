<template>
    <n-card>
        <n-tabs type="line" animated v-model="activeTabs">
            <!-- display-directive="show:lazy" ，切换不会清空数据，并且切换时懒加载 -->
            <n-tab-pane name="Proxy" display-directive="show:lazy" tab="Proxy">
                <n-message-provider>
                    <Proxy/>
                </n-message-provider>
            </n-tab-pane>

            <n-tab-pane name="Repeater" display-directive="show:lazy" tab="Repeater">
                <n-message-provider>
                    <Repeater/>
                </n-message-provider>
            </n-tab-pane>

            <n-tab-pane name="Intruder" display-directive="show:lazy" tab="Intruder">
                <n-message-provider>
                    <Intruder/>
                </n-message-provider>
            </n-tab-pane>
            <n-tab-pane name="Settings" tab="Settings">
                <n-message-provider>
                    <Settings/>
                </n-message-provider>
            </n-tab-pane>
        </n-tabs>
    </n-card>
</template>

<script setup>
import {ref} from "vue";
import Proxy from "./proxy/Proxy.vue";
import Repeater from "./repeater/Repeater.vue";
import Intruder from "./intruder/Intruder.vue";
import Settings from "./Settings.vue";

const activeTabs = ref(['Proxy', 'Repeater', 'Intruder'])

</script>

<style scoped>

</style>