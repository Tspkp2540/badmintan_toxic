<script setup lang="ts">
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import { useAuthStore } from '@/stores/auth'

const router = useRouter()
const authStore = useAuthStore()

const username = ref('')
const fullName = ref('')
const email = ref('')
const password = ref('')
const confirmPassword = ref('')
const errorMsg = ref('')

async function handleRegister() {
  errorMsg.value = ''

  if (password.value !== confirmPassword.value) {
    errorMsg.value = 'รหัสผ่านไม่ตรงกัน'
    return
  }

  if (password.value.length < 6) {
    errorMsg.value = 'รหัสผ่านต้องมีอย่างน้อย 6 ตัวอักษร'
    return
  }

  try {
    await authStore.register({
      username: username.value,
      fullName: fullName.value,
      email: email.value,
      password: password.value,
    })
    router.push('/dashboard')
  } catch {
    errorMsg.value = authStore.error || 'ลงทะเบียนไม่สำเร็จ'
  }
}
</script>

<template>
  <div class="auth-container">
    <div class="auth-card">
      <div class="auth-header">
        <h1>🏸 Badminton Hub</h1>
        <p>ลงทะเบียนนักแบดมินตัน</p>
      </div>

      <form @submit.prevent="handleRegister" class="auth-form">
        <div class="form-group">
          <label for="username">ชื่อผู้ใช้</label>
          <input
            id="username"
            v-model="username"
            type="text"
            placeholder="badminton_player"
            required
          />
        </div>

        <div class="form-group">
          <label for="fullName">ชื่อ-นามสกุล</label>
          <input
            id="fullName"
            v-model="fullName"
            type="text"
            placeholder="สมชาย ใจดี"
            required
          />
        </div>

        <div class="form-group">
          <label for="email">อีเมล</label>
          <input
            id="email"
            v-model="email"
            type="email"
            placeholder="you@example.com"
            required
          />
        </div>

        <div class="form-group">
          <label for="password">รหัสผ่าน</label>
          <input
            id="password"
            v-model="password"
            type="password"
            placeholder="อย่างน้อย 6 ตัวอักษร"
            required
          />
        </div>

        <div class="form-group">
          <label for="confirmPassword">ยืนยันรหัสผ่าน</label>
          <input
            id="confirmPassword"
            v-model="confirmPassword"
            type="password"
            placeholder="••••••••"
            required
          />
        </div>

        <p v-if="errorMsg" class="error-text">{{ errorMsg }}</p>

        <button type="submit" class="btn-primary" :disabled="authStore.loading">
          {{ authStore.loading ? 'กำลังลงทะเบียน...' : 'ลงทะเบียน' }}
        </button>
      </form>

      <p class="auth-footer">
        มีบัญชีแล้ว?
        <router-link to="/login">เข้าสู่ระบบ</router-link>
      </p>
    </div>
  </div>
</template>

<style scoped>
.auth-container {
  min-height: 100vh;
  display: flex;
  align-items: center;
  justify-content: center;
  background: linear-gradient(135deg, #0f172a 0%, #1e3a5f 100%);
  padding: 1rem;
}

.auth-card {
  background: #1e293b;
  border-radius: 16px;
  padding: 2.5rem;
  width: 100%;
  max-width: 420px;
  box-shadow: 0 20px 60px rgba(0, 0, 0, 0.3);
}

.auth-header {
  text-align: center;
  margin-bottom: 2rem;
}

.auth-header h1 {
  font-size: 1.8rem;
  color: #38bdf8;
  margin-bottom: 0.5rem;
}

.auth-header p {
  color: #94a3b8;
  font-size: 1rem;
}

.auth-form {
  display: flex;
  flex-direction: column;
  gap: 1rem;
}

.form-group {
  display: flex;
  flex-direction: column;
  gap: 0.4rem;
}

.form-group label {
  color: #cbd5e1;
  font-size: 0.9rem;
  font-weight: 500;
}

.form-group input {
  padding: 0.75rem 1rem;
  border: 1px solid #334155;
  border-radius: 8px;
  background: #0f172a;
  color: #e2e8f0;
  font-size: 1rem;
  transition: border-color 0.2s;
}

.form-group input:focus {
  outline: none;
  border-color: #38bdf8;
}

.error-text {
  color: #f87171;
  font-size: 0.85rem;
  text-align: center;
}

.btn-primary {
  padding: 0.75rem;
  border: none;
  border-radius: 8px;
  background: linear-gradient(135deg, #38bdf8, #818cf8);
  color: #fff;
  font-size: 1rem;
  font-weight: 600;
  cursor: pointer;
  transition: opacity 0.2s;
}

.btn-primary:hover {
  opacity: 0.9;
}

.btn-primary:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.auth-footer {
  text-align: center;
  margin-top: 1.5rem;
  color: #94a3b8;
  font-size: 0.9rem;
}

.auth-footer a {
  color: #38bdf8;
  text-decoration: none;
  font-weight: 600;
}

.auth-footer a:hover {
  text-decoration: underline;
}
</style>
