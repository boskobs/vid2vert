import './assets/main.css';
import mitt from 'mitt';
import { createApp } from 'vue';
import App from './App.vue';
import vuex from './conf/store.js';
const app = createApp(App);
// Vex
import 'vex-js/dist/css/vex.css';
import 'vex-js/dist/css/vex-theme-os.css';
import vex from 'vex-js/dist/js/vex.combined.js';
vex.defaultOptions.className = 'vex-theme-os';
app.config.globalProperties.$prompt = (options) => new Promise(callback => vex.dialog.prompt({ ...options, callback }));
app.config.globalProperties.$alert = (message) => new Promise(callback => vex.dialog.alert({ message, callback }));
app.config.globalProperties.$confirm = (message) => new Promise(callback => vex.dialog.confirm({ message, callback }));
//
const bus = mitt();
app.config.globalProperties.$bus = bus;
app.use(vuex);
app.mount('#app');