import { createApp } from 'vue'
import App from './App.vue'
import store from './store'
import router from './router'
import vueLib from '@starport/vue'

const app = createApp(App)
app.config.globalProperties._depsLoaded = true
app.config.globalProperties.window = window
app.use(store).use(router).use(vueLib).mount('#app')
