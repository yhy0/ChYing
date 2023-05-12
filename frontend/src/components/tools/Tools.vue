<script setup>
import TaskList from "./Tasklist.vue";
</script>

<template>
    <n-card>
        <n-tabs type="line" animated>
            <!-- display-directive="show:lazy" ，切换不会清空数据，并且切换时懒加载 -->
            <n-tab-pane name="TaskList" display-directive="show:lazy" tab="杀软查询">
                <n-message-provider>
                    <TaskList/>
                </n-message-provider>
            </n-tab-pane>
        </n-tabs>
    </n-card>
</template>

<style scoped>

</style>