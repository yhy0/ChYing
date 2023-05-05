<template>
    <n-card>
        <n-tabs type="line" animated v-model="activeTabs">
            <!-- display-directive="show:lazy" ，切换不会情况数据，并且切换时懒加载 -->
            <n-tab-pane name="Intercept" display-directive="show:lazy" tab="Intercept">
                <n-message-provider>
                    <Intercept/>
                </n-message-provider>
            </n-tab-pane>

            <n-tab-pane name="HTTPHistory" display-directive="show:lazy" tab="HTTPHistory">
                <n-message-provider>
                    <HTTPHistory/>
                </n-message-provider>
            </n-tab-pane>

        </n-tabs>
    </n-card>
</template>

<script setup>
import Intercept from "./intercept/Intercept.vue";
import HTTPHistory from "./httpHisroty/HTTPHistory.vue";
import {ref} from "vue";

const activeTabs = ref(['Intercept', 'HTTPHistory'])

</script>

<style scoped>

</style>