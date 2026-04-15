<script setup lang="ts">
import { ref, onMounted, computed } from 'vue'
import { useAuthStore } from '@/stores/auth'
import { userService } from '@/services/userService'
import { UserRole, UserRoleLabels } from '@/models/User'
import type { User } from '@/models/User'

const authStore = useAuthStore()
const users = ref<User[]>([])
const loading = ref(false)
const errorMsg = ref('')
const successMsg = ref('')

const assignableRoles = computed(() => {
  if (authStore.isAdmin) {
    return [UserRole.ADMIN, UserRole.LEADER, UserRole.VICE_LEADER, UserRole.PLAYER]
  }
  if (authStore.isLeader) {
    return [UserRole.VICE_LEADER, UserRole.PLAYER]
  }
  return []
})

async function loadUsers() {
  loading.value = true
  errorMsg.value = ''
  try {
    users.value = await userService.getUsers()
  } catch {
    errorMsg.value = 'โหลดรายชื่อผู้ใช้ไม่สำเร็จ'
  } finally {
    loading.value = false
  }
}

async function changeRole(userId: string, role: string) {
  errorMsg.value = ''
  successMsg.value = ''
  try {
    await userService.updateUserRole({ userId, role: role as any })
    successMsg.value = 'เปลี่ยนบทบาทสำเร็จ'
    await loadUsers()
  } catch (e: any) {
    errorMsg.value = e.response?.data?.message || 'เปลี่ยนบทบาทไม่สำเร็จ'
  }
}

onMounted(loadUsers)
</script>

<template>
  <div class="manage-users">
    <header class="page-header">
      <h1>🏸 Badminton Hub</h1>
      <nav>
        <router-link to="/dashboard">แดชบอร์ด</router-link>
        <router-link to="/court">สนาม</router-link>
        <router-link to="/ranking">อันดับ</router-link>
        <router-link to="/manage-users">จัดการผู้ใช้</router-link>
        <router-link to="/profile">โปรไฟล์</router-link>
        <button @click="authStore.logout()" class="btn-logout">ออกจากระบบ</button>
      </nav>
    </header>

    <main class="content">
      <h2>จัดการผู้ใช้</h2>
      <p class="subtitle">บทบาทของคุณ: <strong>{{ UserRoleLabels[authStore.userRole] }}</strong></p>

      <p v-if="errorMsg" class="error-text">{{ errorMsg }}</p>
      <p v-if="successMsg" class="success-text">{{ successMsg }}</p>
      <p v-if="loading" class="loading-text">กำลังโหลด...</p>

      <div class="users-table-wrap" v-if="!loading">
        <table class="users-table">
          <thead>
            <tr>
              <th>ชื่อผู้ใช้</th>
              <th>ชื่อ-นามสกุล</th>
              <th>อีเมล</th>
              <th>บทบาท</th>
              <th>เลเวล</th>
              <th v-if="authStore.canManageRoles">เปลี่ยนบทบาท</th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="u in users" :key="u.id" :class="{ 'is-self': u.id === authStore.user?.id }">
              <td>@{{ u.username }}</td>
              <td>{{ u.fullName }}</td>
              <td>{{ u.email }}</td>
              <td>
                <span class="role-badge" :class="'role-' + u.role">
                  {{ UserRoleLabels[u.role] || u.role }}
                </span>
              </td>
              <td>{{ u.level }}</td>
              <td v-if="authStore.canManageRoles">
                <select
                  v-if="u.id !== authStore.user?.id"
                  :value="u.role"
                  @change="changeRole(u.id, ($event.target as HTMLSelectElement).value)"
                  class="role-select"
                >
                  <option
                    v-for="role in assignableRoles"
                    :key="role"
                    :value="role"
                  >
                    {{ UserRoleLabels[role] }}
                  </option>
                </select>
                <span v-else class="self-label">คุณ</span>
              </td>
            </tr>
          </tbody>
        </table>
      </div>
    </main>
  </div>
</template>

<style scoped>
.manage-users {
  min-height: 100vh;
  background: #0f172a;
  color: #e2e8f0;
}

.page-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 1rem 2rem;
  background: #1e293b;
  border-bottom: 1px solid #334155;
}

.page-header h1 {
  font-size: 1.4rem;
  color: #38bdf8;
}

.page-header nav {
  display: flex;
  align-items: center;
  gap: 1.5rem;
}

.page-header nav a {
  color: #94a3b8;
  text-decoration: none;
  font-weight: 500;
  transition: color 0.2s;
}

.page-header nav a:hover,
.page-header nav a.router-link-active {
  color: #38bdf8;
}

.btn-logout {
  background: none;
  border: 1px solid #475569;
  color: #94a3b8;
  padding: 0.4rem 1rem;
  border-radius: 6px;
  cursor: pointer;
  transition: all 0.2s;
}

.btn-logout:hover {
  border-color: #f87171;
  color: #f87171;
}

.content {
  max-width: 1000px;
  margin: 2rem auto;
  padding: 0 1rem;
}

.content h2 {
  font-size: 1.5rem;
  margin-bottom: 0.5rem;
}

.subtitle {
  color: #94a3b8;
  margin-bottom: 1.5rem;
}

.error-text {
  color: #f87171;
  margin-bottom: 1rem;
}

.success-text {
  color: #4ade80;
  margin-bottom: 1rem;
}

.loading-text {
  color: #94a3b8;
}

.users-table-wrap {
  overflow-x: auto;
}

.users-table {
  width: 100%;
  border-collapse: collapse;
  background: #1e293b;
  border-radius: 12px;
  overflow: hidden;
}

.users-table th,
.users-table td {
  padding: 0.75rem 1rem;
  text-align: left;
  border-bottom: 1px solid #334155;
}

.users-table th {
  background: #334155;
  color: #94a3b8;
  font-weight: 600;
  font-size: 0.85rem;
  text-transform: uppercase;
}

.users-table tr.is-self {
  background: rgba(56, 189, 248, 0.05);
}

.role-badge {
  display: inline-block;
  padding: 0.2rem 0.6rem;
  border-radius: 12px;
  font-size: 0.8rem;
  font-weight: 600;
}

.role-admin {
  background: rgba(239, 68, 68, 0.2);
  color: #f87171;
}

.role-leader {
  background: rgba(251, 191, 36, 0.2);
  color: #fbbf24;
}

.role-vice_leader {
  background: rgba(129, 140, 248, 0.2);
  color: #818cf8;
}

.role-player {
  background: rgba(74, 222, 128, 0.2);
  color: #4ade80;
}

.role-select {
  padding: 0.4rem 0.6rem;
  border: 1px solid #475569;
  border-radius: 6px;
  background: #0f172a;
  color: #e2e8f0;
  font-size: 0.85rem;
  cursor: pointer;
}

.role-select:focus {
  outline: none;
  border-color: #38bdf8;
}

.self-label {
  color: #64748b;
  font-style: italic;
}
</style>
