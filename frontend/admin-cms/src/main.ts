import { createApp } from 'vue'
import { createPinia } from 'pinia'
import App from './App.vue'
import { router } from './router'
import BladeIcon from './components/common/BladeIcon.vue'

const app = createApp(App)
const pinia = createPinia()

app.component('BladeIcon', BladeIcon)
app.component('blade-icon', BladeIcon)

app.use(pinia)
app.use(router)

app.mount('#app')
