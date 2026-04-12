import { createApp } from 'vue'
import { createPinia } from 'pinia'
import App from './App.vue'
import router from './router'
import { piniaSessionPlugin, refreshSessionActivity } from './plugins/piniaSession'
import './style.css'

const pinia = createPinia()
pinia.use(piniaSessionPlugin)

const app = createApp(App)
app.use(pinia)
app.use(router)
app.mount('#app')

// อัปเดต session activity ทุกครั้งที่ user คลิกหรือกดปุ่ม
const ACTIVITY_EVENTS = ['click', 'keydown', 'touchstart']
let activityThrottle = 0
function onUserActivity() {
  const now = Date.now()
  // throttle ทุก 60 วินาที
  if (now - activityThrottle < 60_000) return
  activityThrottle = now
  refreshSessionActivity()
}
ACTIVITY_EVENTS.forEach((evt) => window.addEventListener(evt, onUserActivity, { passive: true }))
