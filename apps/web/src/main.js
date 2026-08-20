import { createApp } from 'vue'
import App from './App.vue'
import router from './router'
import { loadCopy } from './copy'
import './styles.css'

const app = createApp(App)
app.use(router)
loadCopy()
  .catch(() => {})
  .finally(() => {
    app.mount('#app')
  })
