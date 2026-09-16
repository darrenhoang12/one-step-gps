import { createApp } from "vue";
import { VueQueryPlugin } from "@tanstack/vue-query";
import App from "./App.vue";
import "./style.css";
import { initializeTheme } from "./lib/theme";

initializeTheme();
createApp(App).use(VueQueryPlugin).mount("#app");
