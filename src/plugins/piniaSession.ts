import type { PiniaPluginContext } from 'pinia'
import { watch } from 'vue'

const SESSION_KEY = 'badminton_session'

export interface SessionData {
  token: string | null
  user: any | null
  lastActivity: number
}

/**
 * Pinia plugin ที่จะ persist auth state ลง localStorage
 * และ restore กลับมาเมื่อ reload หน้า
 */
export function piniaSessionPlugin({ store }: PiniaPluginContext) {
  // เฉพาะ auth store เท่านั้น
  if (store.$id !== 'auth') return

  // Restore session เมื่อ store ถูกสร้าง
  const saved = loadSession()
  if (saved && saved.token) {
    // ตรวจสอบว่า session หมดอายุหรือยัง (24 ชั่วโมง)
    const SESSION_EXPIRY_MS = 24 * 60 * 60 * 1000
    const isExpired = Date.now() - saved.lastActivity > SESSION_EXPIRY_MS

    if (isExpired) {
      clearSession()
    } else {
      store.token = saved.token
      store.user = saved.user
    }
  }

  // Watch และ sync ลง localStorage ทุกครั้งที่ state เปลี่ยน
  watch(
    () => ({ token: store.token, user: store.user }),
    (newState) => {
      if (newState.token) {
        saveSession({
          token: newState.token,
          user: newState.user,
          lastActivity: Date.now(),
        })
      } else {
        clearSession()
      }
    },
    { deep: true }
  )
}

function saveSession(data: SessionData): void {
  try {
    localStorage.setItem(SESSION_KEY, JSON.stringify(data))
    localStorage.setItem('auth_token', data.token ?? '')
  } catch {
    // localStorage อาจเต็มหรือถูก block
  }
}

function loadSession(): SessionData | null {
  try {
    const raw = localStorage.getItem(SESSION_KEY)
    if (!raw) return null
    return JSON.parse(raw) as SessionData
  } catch {
    return null
  }
}

function clearSession(): void {
  localStorage.removeItem(SESSION_KEY)
  localStorage.removeItem('auth_token')
}

/**
 * อัปเดต lastActivity ทุกครั้งที่ user มี interaction
 * เรียกจาก App.vue หรือ main.ts
 */
export function refreshSessionActivity(): void {
  try {
    const raw = localStorage.getItem(SESSION_KEY)
    if (!raw) return
    const data = JSON.parse(raw) as SessionData
    data.lastActivity = Date.now()
    localStorage.setItem(SESSION_KEY, JSON.stringify(data))
  } catch {
    // ignore
  }
}
