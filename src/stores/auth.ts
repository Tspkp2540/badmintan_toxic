import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import type { User } from '@/models/User'
import { authService } from '@/services/authService'
import type { LoginRequest, RegisterRequest } from '@/models/User'

export const useAuthStore = defineStore('auth', () => {
  const user = ref<User | null>(null)
  const token = ref<string | null>(localStorage.getItem('auth_token'))
  const loading = ref(false)
  const error = ref<string | null>(null)
  const sessionReady = ref(false)

  const isAuthenticated = computed(() => !!token.value)
  const playerLevel = computed(() => user.value?.level ?? 1)
  const playerRank = computed(() => user.value?.rank ?? 'Bronze')

  /**
   * เรียกตอนเริ่มต้น app เพื่อ restore session
   * ถ้ามี token แต่ยังไม่มี user data จะ fetch profile อัตโนมัติ
   */
  async function initSession() {
    if (token.value && !user.value) {
      try {
        await fetchProfile()
      } catch {
        // token หมดอายุหรือใช้ไม่ได้ ให้ logout
        logout()
      }
    }
    sessionReady.value = true
  }

  async function login(data: LoginRequest) {
    loading.value = true
    error.value = null
    try {
      const res = await authService.login(data)
      token.value = res.token
      user.value = res.user
    } catch (e: any) {
      error.value = e.response?.data?.message || 'เข้าสู่ระบบไม่สำเร็จ'
      throw e
    } finally {
      loading.value = false
    }
  }

  async function register(data: RegisterRequest) {
    loading.value = true
    error.value = null
    try {
      const res = await authService.register(data)
      token.value = res.token
      user.value = res.user
    } catch (e: any) {
      error.value = e.response?.data?.message || 'ลงทะเบียนไม่สำเร็จ'
      throw e
    } finally {
      loading.value = false
    }
  }

  async function fetchProfile() {
    if (!token.value) return
    loading.value = true
    try {
      user.value = await authService.getProfile()
    } catch (e: any) {
      error.value = e.response?.data?.message || 'โหลดข้อมูลไม่สำเร็จ'
    } finally {
      loading.value = false
    }
  }

  function logout() {
    authService.logout()
    token.value = null
    user.value = null
  }

  return {
    user,
    token,
    loading,
    error,
    sessionReady,
    isAuthenticated,
    playerLevel,
    playerRank,
    initSession,
    login,
    register,
    fetchProfile,
    logout,
  }
})
