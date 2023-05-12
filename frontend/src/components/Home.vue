<script setup>
import {ref} from "vue";
import Twj from "./Twj.vue";
import Swagger from "./Swagger.vue";
import Fuzz from "./fuzz/Fuzz.vue";
import Decoder from "./Decoder.vue";
import BurpSuite from "./burpsuite/BurpSuite.vue";
import About from "./About.vue";
import Tools from "./tools/Tools.vue";

const activeTabs = ref(['Fuzz', 'Swagger', 'TWJ', 'Decoder', 'BurpSuite', 'About'])

</script>

<template>
    <n-card>
        <n-tabs type="line" animated v-model="activeTabs">
            <!-- display-directive="show:lazy" ，切换不会情况数据，并且切换时懒加载 -->
            <n-tab-pane name="Fuzz" display-directive="show:lazy" tab="Fuzz">
                <n-message-provider>
                    <Fuzz/>
                </n-message-provider>
            </n-tab-pane>

            <n-tab-pane name="Swagger" display-directive="show:lazy" tab="Swagger">
                <n-message-provider>
                    <Swagger/>
                </n-message-provider>
            </n-tab-pane>

            <n-tab-pane name="TWJ" display-directive="show:lazy" tab="TWJ">
                    <n-message-provider>
                        <Twj/>
                    </n-message-provider>
            </n-tab-pane>

            <n-tab-pane name="BurpSuite" display-directive="show:lazy" tab="BurpSuite">
                <n-message-provider>
                    <BurpSuite/>
                </n-message-provider>
            </n-tab-pane>

            <n-tab-pane name="Decoder" display-directive="show:lazy" tab="Decoder">
                <n-message-provider>
                    <Decoder/>
                </n-message-provider>
            </n-tab-pane>

            <n-tab-pane name="Tools" display-directive="show:lazy" tab="Tools">
                <n-message-provider>
                    <Tools/>
                </n-message-provider>
            </n-tab-pane>

            <n-tab-pane name="About" display-directive="show:lazy" tab="About">
                <n-message-provider>
                    <About/>
                </n-message-provider>
            </n-tab-pane>
        </n-tabs>
    </n-card>
</template>