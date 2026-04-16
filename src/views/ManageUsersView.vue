<script setup lang="ts">
import { ref, onMounted, computed } from 'vue'
import { useAuthStore } from '@/stores/auth'
import { userService } from '@/services/userService'
import { UserRole, UserRoleLabels, SkillLevelLabels, SkillLevelOrder } from '@/models/User'
import type { User, SkillLevel } from '@/models/User'
import SkillBadge from '@/components/SkillBadge.vue'
import LevelBadge from '@/components/LevelBadge.vue'
import ConfirmModal from '@/components/ConfirmModal.vue'

const authStore = useAuthStore()
const users = ref<User[]>([])
const loading = ref(false)
const errorMsg = ref('')
const successMsg = ref('')
const searchQuery = ref('')
const filterRole = ref('all')
const selectedUser = ref<User | null>(null)
const showLogoutConfirm = ref(false)

// Skill edit state
const editSkillLevel = ref<string>('')
const editSkillStars = ref<number>(1)

const assignableRoles = computed(() => {
  if (authStore.isAdmin) {
    return [UserRole.ADMIN, UserRole.LEADER, UserRole.VICE_LEADER, UserRole.PLAYER]
  }
  if (authStore.isLeader) {
    return [UserRole.VICE_LEADER, UserRole.PLAYER]
  }
  return []
})

const filteredUsers = computed(() => {
  let result = users.value
  if (filterRole.value !== 'all') {
    result = result.filter(u => u.role === filterRole.value)
  }
  if (searchQuery.value.trim()) {
    const q = searchQuery.value.trim().toLowerCase()
    result = result.filter(u =>
      u.username.toLowerCase().includes(q) ||
      u.fullName.toLowerCase().includes(q) ||
      u.email.toLowerCase().includes(q)
    )
  }
  return result
})

const userCount = computed(() => ({
  total: users.value.length,
  admin: users.value.filter(u => u.role === 'admin').length,
  leader: users.value.filter(u => u.role === 'leader').length,
  vice_leader: users.value.filter(u => u.role === 'vice_leader').length,
  player: users.value.filter(u => u.role === 'player').length,
}))

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

function showSuccess(msg: string) {
  successMsg.value = msg
  setTimeout(() => { successMsg.value = '' }, 3000)
}

async function changeRole(userId: string, role: string) {
  pendingRoleChange.value = { userId, role }
  confirmAction.value = 'role'
}

// Confirm modal state
const confirmAction = ref<'role' | 'skill' | null>(null)
const confirmLoading = ref(false)
const pendingRoleChange = ref<{ userId: string; role: string } | null>(null)

const confirmConfig = computed(() => {
  if (confirmAction.value === 'role' && pendingRoleChange.value) {
    const targetUser = users.value.find(u => u.id === pendingRoleChange.value!.userId)
    const roleName = UserRoleLabels[pendingRoleChange.value.role as keyof typeof UserRoleLabels] || pendingRoleChange.value.role
    return {
      title: '🛡️ เปลี่ยนบทบาท',
      message: `ต้องการเปลี่ยนบทบาทของ "${targetUser?.fullName}" เป็น "${roleName}" หรือไม่?`,
      variant: 'warning' as const,
      confirmText: 'เปลี่ยนบทบาท',
    }
  }
  if (confirmAction.value === 'skill' && selectedUser.value) {
    return {
      title: '⚔️ อัปเดตระดับฝีมือ',
      message: `ต้องการเปลี่ยนระดับฝีมือของ "${selectedUser.value.fullName}" เป็น ${editSkillLevel.value} ★${editSkillStars.value} หรือไม่?`,
      variant: 'warning' as const,
      confirmText: 'บันทึก',
    }
  }
  return { title: '', message: '', variant: 'info' as const, confirmText: '' }
})

async function handleConfirm() {
  confirmLoading.value = true
  try {
    if (confirmAction.value === 'role' && pendingRoleChange.value) {
      await doChangeRole(pendingRoleChange.value.userId, pendingRoleChange.value.role)
      pendingRoleChange.value = null
    } else if (confirmAction.value === 'skill') {
      await doSaveSkillLevel()
    }
  } finally {
    confirmLoading.value = false
    confirmAction.value = null
  }
}

async function doChangeRole(userId: string, role: string) {
  errorMsg.value = ''
  try {
    await userService.updateUserRole({ userId, role: role as any })
    showSuccess('เปลี่ยนบทบาทสำเร็จ')
    await loadUsers()
    if (selectedUser.value?.id === userId) {
      selectedUser.value = users.value.find(u => u.id === userId) ?? null
    }
  } catch (e: any) {
    errorMsg.value = e.response?.data?.message || 'เปลี่ยนบทบาทไม่สำเร็จ'
  }
}

function openUserDetail(user: User) {
  selectedUser.value = user
  editSkillLevel.value = user.skillLevel
  editSkillStars.value = user.skillStars
}

function closeDetail() {
  selectedUser.value = null
}

async function saveSkillLevel() {
  if (!selectedUser.value) return
  confirmAction.value = 'skill'
}

async function doSaveSkillLevel() {
  if (!selectedUser.value) return
  errorMsg.value = ''
  try {
    await userService.updateSkillLevel({
      userId: selectedUser.value.id,
      skillLevel: editSkillLevel.value as SkillLevel,
      skillStars: editSkillStars.value,
    })
    showSuccess('อัปเดตระดับฝีมือสำเร็จ')
    await loadUsers()
    selectedUser.value = users.value.find(u => u.id === selectedUser.value!.id) ?? null
  } catch (e: any) {
    errorMsg.value = e.response?.data?.message || 'อัปเดตระดับฝีมือไม่สำเร็จ'
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
        <router-link to="/courts">สนาม</router-link>
        <router-link to="/ranking">อันดับ</router-link>
        <router-link to="/skill-guide">คู่มือระดับ</router-link>
        <router-link to="/manage-users">จัดการผู้ใช้</router-link>
        <router-link to="/profile">โปรไฟล์</router-link>
        <button @click="showLogoutConfirm = true" class="btn-logout">ออกจากระบบ</button>
      </nav>
    </header>

    <ConfirmModal
      :show="showLogoutConfirm"
      title="ออกจากระบบ"
      message="ต้องการออกจากระบบหรือไม่?"
      variant="danger"
      confirm-text="ออกจากระบบ"
      @confirm="authStore.logout()"
      @cancel="showLogoutConfirm = false"
    />

    <main class="content">
      <div class="page-title-row">
        <h2>👥 จัดการผู้ใช้</h2>
        <span class="user-count-total">ทั้งหมด {{ userCount.total }} คน</span>
      </div>

      <!-- Role stat cards -->
      <div class="role-stats">
        <div class="role-stat" :class="{ active: filterRole === 'all' }" @click="filterRole = 'all'">
          <span class="stat-num">{{ userCount.total }}</span>
          <span class="stat-label">ทั้งหมด</span>
        </div>
        <div class="role-stat role-stat-admin" :class="{ active: filterRole === 'admin' }" @click="filterRole = 'admin'">
          <span class="stat-num">{{ userCount.admin }}</span>
          <span class="stat-label">แอดมิน</span>
        </div>
        <div class="role-stat role-stat-leader" :class="{ active: filterRole === 'leader' }" @click="filterRole = 'leader'">
          <span class="stat-num">{{ userCount.leader }}</span>
          <span class="stat-label">หัวก๊วน</span>
        </div>
        <div class="role-stat role-stat-vice" :class="{ active: filterRole === 'vice_leader' }" @click="filterRole = 'vice_leader'">
          <span class="stat-num">{{ userCount.vice_leader }}</span>
          <span class="stat-label">รองหัวก๊วน</span>
        </div>
        <div class="role-stat role-stat-player" :class="{ active: filterRole === 'player' }" @click="filterRole = 'player'">
          <span class="stat-num">{{ userCount.player }}</span>
          <span class="stat-label">ผู้เล่น</span>
        </div>
      </div>

      <!-- Search -->
      <div class="search-bar">
        <input
          v-model="searchQuery"
          type="text"
          placeholder="🔍 ค้นหาชื่อผู้ใช้ / ชื่อจริง / อีเมล..."
          class="search-input"
        />
      </div>

      <p v-if="errorMsg" class="error-text">{{ errorMsg }}</p>
      <p v-if="successMsg" class="success-text">{{ successMsg }}</p>
      <p v-if="loading" class="loading-text">กำลังโหลด...</p>

      <!-- User table -->
      <div class="users-table-wrap" v-if="!loading">
        <table class="users-table">
          <thead>
            <tr>
              <th>ผู้ใช้</th>
              <th>บทบาท</th>
              <th>ระดับฝีมือ</th>
              <th>เลเวล</th>
              <th>สถิติ</th>
              <th v-if="authStore.canManageRoles">เปลี่ยนบทบาท</th>
              <th>จัดการ</th>
            </tr>
          </thead>
          <tbody>
            <tr
              v-for="u in filteredUsers"
              :key="u.id"
              :class="{ 'is-self': u.id === authStore.user?.id, 'is-selected': selectedUser?.id === u.id }"
            >
              <td class="user-cell">
                <div class="user-avatar">{{ u.fullName?.charAt(0)?.toUpperCase() ?? '?' }}</div>
                <div>
                  <div class="user-name">{{ u.fullName }}</div>
                  <div class="user-username">@{{ u.username }}</div>
                </div>
              </td>
              <td>
                <span class="role-badge" :class="'role-' + u.role">
                  {{ UserRoleLabels[u.role] || u.role }}
                </span>
              </td>
              <td>
                <SkillBadge :skill-level="u.skillLevel" :skill-stars="u.skillStars" />
              </td>
              <td>
                <LevelBadge :level="u.level" />
              </td>
              <td class="stats-cell">
                <span class="win-stat">{{ u.wins }}W</span>
                <span class="lose-stat">{{ u.losses }}L</span>
                <span class="match-stat">{{ u.totalMatches }}G</span>
              </td>
              <td v-if="authStore.canManageRoles">
                <select
                  v-if="u.id !== authStore.user?.id"
                  :value="u.role"
                  @change="changeRole(u.id, ($event.target as HTMLSelectElement).value)"
                  class="role-select"
                >
                  <option v-for="role in assignableRoles" :key="role" :value="role">
                    {{ UserRoleLabels[role] }}
                  </option>
                </select>
                <span v-else class="self-label">คุณ</span>
              </td>
              <td>
                <button class="btn-detail" @click="openUserDetail(u)">ดูรายละเอียด</button>
              </td>
            </tr>
          </tbody>
        </table>
        <p v-if="filteredUsers.length === 0" class="empty-text">ไม่พบผู้ใช้ที่ตรงกับเงื่อนไข</p>
      </div>
    </main>

    <!-- User Detail Side Panel -->
    <transition name="panel">
      <div v-if="selectedUser" class="detail-overlay" @click.self="closeDetail">
        <div class="detail-panel">
          <button class="btn-close" @click="closeDetail">✕</button>

          <div class="detail-header">
            <div class="detail-avatar">{{ selectedUser.fullName?.charAt(0)?.toUpperCase() ?? '?' }}</div>
            <div>
              <h3>{{ selectedUser.fullName }}</h3>
              <p class="detail-username">@{{ selectedUser.username }}</p>
              <span class="role-badge" :class="'role-' + selectedUser.role">
                {{ UserRoleLabels[selectedUser.role] }}
              </span>
            </div>
          </div>

          <div class="detail-info-grid">
            <div class="info-item">
              <span class="info-label">อีเมล</span>
              <span class="info-value">{{ selectedUser.email }}</span>
            </div>
            <div class="info-item">
              <span class="info-label">เลเวล</span>
              <span class="info-value"><LevelBadge :level="selectedUser.level" /></span>
            </div>
            <div class="info-item">
              <span class="info-label">แรงค์</span>
              <span class="info-value">{{ selectedUser.rank }}</span>
            </div>
            <div class="info-item">
              <span class="info-label">RP</span>
              <span class="info-value">{{ selectedUser.rankPoints }}</span>
            </div>
            <div class="info-item">
              <span class="info-label">ชนะ/แพ้</span>
              <span class="info-value">
                <span class="win-stat">{{ selectedUser.wins }}W</span> /
                <span class="lose-stat">{{ selectedUser.losses }}L</span>
              </span>
            </div>
            <div class="info-item">
              <span class="info-label">แมตช์ทั้งหมด</span>
              <span class="info-value">{{ selectedUser.totalMatches }}</span>
            </div>
            <div class="info-item">
              <span class="info-label">อัตราชนะ</span>
              <span class="info-value">{{ selectedUser.winRate.toFixed(1) }}%</span>
            </div>
            <div class="info-item">
              <span class="info-label">EXP</span>
              <span class="info-value">{{ selectedUser.exp }} / {{ selectedUser.expToNextLevel }}</span>
            </div>
            <div class="info-item">
              <span class="info-label">สมัครเมื่อ</span>
              <span class="info-value">{{ new Date(selectedUser.createdAt).toLocaleDateString('th-TH') }}</span>
            </div>
          </div>

          <!-- Skill Level Edit (admin/leader only) -->
          <div v-if="authStore.canManageRoles && selectedUser.id !== authStore.user?.id" class="detail-section">
            <h4>⚔️ จัดการระดับฝีมือ</h4>
            <div class="skill-current">
              <span>ปัจจุบัน:</span>
              <SkillBadge :skill-level="selectedUser.skillLevel" :skill-stars="selectedUser.skillStars" />
            </div>
            <div class="skill-edit-row">
              <select v-model="editSkillLevel" class="skill-select">
                <option v-for="lv in SkillLevelOrder" :key="lv" :value="lv">
                  {{ lv }} — {{ SkillLevelLabels[lv] }}
                </option>
              </select>
              <div class="star-picker">
                <button
                  v-for="i in 5"
                  :key="i"
                  class="star-btn"
                  :class="{ active: i <= editSkillStars }"
                  @click="editSkillStars = i"
                >★</button>
              </div>
              <button class="btn-save-skill" @click="saveSkillLevel">บันทึก</button>
            </div>
            <div class="skill-preview">
              <span>พรีวิว:</span>
              <SkillBadge :skill-level="editSkillLevel" :skill-stars="editSkillStars" />
            </div>
          </div>

          <!-- Role change (admin/leader only) -->
          <div v-if="authStore.canManageRoles && selectedUser.id !== authStore.user?.id" class="detail-section">
            <h4>🛡️ เปลี่ยนบทบาท</h4>
            <div class="role-change-row">
              <select
                :value="selectedUser.role"
                @change="changeRole(selectedUser!.id, ($event.target as HTMLSelectElement).value)"
                class="role-select-lg"
              >
                <option v-for="role in assignableRoles" :key="role" :value="role">
                  {{ UserRoleLabels[role] }}
                </option>
              </select>
            </div>
          </div>
        </div>
      </div>
    </transition>

    <!-- Confirm Modal -->
    <ConfirmModal
      :show="!!confirmAction"
      :title="confirmConfig.title"
      :message="confirmConfig.message"
      :variant="confirmConfig.variant"
      :confirm-text="confirmConfig.confirmText"
      :loading="confirmLoading"
      @confirm="handleConfirm"
      @cancel="confirmAction = null; pendingRoleChange = null"
    />
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
  max-width: 1200px;
  margin: 2rem auto;
  padding: 0 1rem;
}

.page-title-row {
  display: flex;
  align-items: center;
  gap: 1rem;
  margin-bottom: 1.5rem;
}

.page-title-row h2 {
  font-size: 1.5rem;
}

.user-count-total {
  color: #64748b;
  font-size: 0.9rem;
}

/* Role stat cards */
.role-stats {
  display: flex;
  gap: 0.75rem;
  margin-bottom: 1.5rem;
  flex-wrap: wrap;
}

.role-stat {
  background: #1e293b;
  border: 1px solid #334155;
  border-radius: 10px;
  padding: 0.75rem 1.25rem;
  text-align: center;
  cursor: pointer;
  transition: all 0.2s;
  min-width: 80px;
}

.role-stat:hover {
  border-color: #475569;
}

.role-stat.active {
  border-color: #38bdf8;
  background: #1e3a5f;
}

.role-stat-admin.active { border-color: #f87171; background: rgba(239, 68, 68, 0.1); }
.role-stat-leader.active { border-color: #fbbf24; background: rgba(251, 191, 36, 0.1); }
.role-stat-vice.active { border-color: #818cf8; background: rgba(129, 140, 248, 0.1); }
.role-stat-player.active { border-color: #4ade80; background: rgba(74, 222, 128, 0.1); }

.stat-num {
  display: block;
  font-size: 1.4rem;
  font-weight: 700;
  color: #f1f5f9;
}

.stat-label {
  font-size: 0.75rem;
  color: #94a3b8;
}

/* Search */
.search-bar {
  margin-bottom: 1rem;
}

.search-input {
  width: 100%;
  padding: 0.75rem 1rem;
  border: 1px solid #334155;
  border-radius: 10px;
  background: #1e293b;
  color: #e2e8f0;
  font-size: 0.95rem;
  transition: border-color 0.2s;
}

.search-input:focus {
  outline: none;
  border-color: #38bdf8;
}

.search-input::placeholder {
  color: #64748b;
}

.error-text { color: #f87171; margin-bottom: 1rem; }
.success-text { color: #4ade80; margin-bottom: 1rem; }
.loading-text { color: #94a3b8; }
.empty-text { color: #64748b; text-align: center; padding: 2rem; }

/* Users table */
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
  font-size: 0.8rem;
  text-transform: uppercase;
  white-space: nowrap;
}

.users-table tr.is-self {
  background: rgba(56, 189, 248, 0.05);
}

.users-table tr.is-selected {
  background: rgba(56, 189, 248, 0.1);
}

.users-table tr:hover {
  background: rgba(255, 255, 255, 0.02);
}

/* User cell */
.user-cell {
  display: flex;
  align-items: center;
  gap: 0.75rem;
}

.user-avatar {
  width: 36px;
  height: 36px;
  border-radius: 50%;
  background: linear-gradient(135deg, #38bdf8, #818cf8);
  display: flex;
  align-items: center;
  justify-content: center;
  font-weight: 700;
  font-size: 0.85rem;
  color: #fff;
  flex-shrink: 0;
}

.user-name {
  font-weight: 600;
  font-size: 0.9rem;
}

.user-username {
  color: #64748b;
  font-size: 0.8rem;
}

/* Stats cell */
.stats-cell {
  display: flex;
  gap: 0.4rem;
  font-size: 0.8rem;
  font-weight: 600;
}

.win-stat { color: #4ade80; }
.lose-stat { color: #f87171; }
.match-stat { color: #94a3b8; }

/* Role badges */
.role-badge {
  display: inline-block;
  padding: 0.2rem 0.6rem;
  border-radius: 12px;
  font-size: 0.8rem;
  font-weight: 600;
  white-space: nowrap;
}

.role-admin { background: rgba(239, 68, 68, 0.2); color: #f87171; }
.role-leader { background: rgba(251, 191, 36, 0.2); color: #fbbf24; }
.role-vice_leader { background: rgba(129, 140, 248, 0.2); color: #818cf8; }
.role-player { background: rgba(74, 222, 128, 0.2); color: #4ade80; }

.role-select {
  padding: 0.35rem 0.5rem;
  border: 1px solid #475569;
  border-radius: 6px;
  background: #0f172a;
  color: #e2e8f0;
  font-size: 0.8rem;
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

.btn-detail {
  padding: 0.35rem 0.75rem;
  background: #334155;
  border: none;
  border-radius: 6px;
  color: #38bdf8;
  font-size: 0.8rem;
  font-weight: 600;
  cursor: pointer;
  transition: background 0.2s;
  white-space: nowrap;
}

.btn-detail:hover {
  background: #475569;
}

/* Detail side panel */
.detail-overlay {
  position: fixed;
  inset: 0;
  background: rgba(0, 0, 0, 0.5);
  z-index: 100;
  display: flex;
  justify-content: flex-end;
}

.detail-panel {
  width: 420px;
  max-width: 90vw;
  height: 100vh;
  background: #1e293b;
  overflow-y: auto;
  padding: 1.5rem;
  position: relative;
}

.btn-close {
  position: absolute;
  top: 1rem;
  right: 1rem;
  background: none;
  border: none;
  color: #94a3b8;
  font-size: 1.2rem;
  cursor: pointer;
}

.btn-close:hover {
  color: #f1f5f9;
}

.detail-header {
  display: flex;
  align-items: center;
  gap: 1rem;
  margin-bottom: 1.5rem;
  padding-right: 2rem;
}

.detail-avatar {
  width: 56px;
  height: 56px;
  border-radius: 50%;
  background: linear-gradient(135deg, #38bdf8, #818cf8);
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 1.4rem;
  font-weight: 700;
  color: #fff;
  flex-shrink: 0;
}

.detail-header h3 {
  font-size: 1.1rem;
  margin-bottom: 0.1rem;
}

.detail-username {
  color: #64748b;
  font-size: 0.85rem;
  margin-bottom: 0.3rem;
}

/* Info grid */
.detail-info-grid {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 0.75rem;
  margin-bottom: 1.5rem;
}

.info-item {
  background: #0f172a;
  border-radius: 8px;
  padding: 0.6rem 0.8rem;
}

.info-label {
  display: block;
  font-size: 0.7rem;
  color: #64748b;
  text-transform: uppercase;
  margin-bottom: 0.2rem;
}

.info-value {
  font-size: 0.9rem;
  font-weight: 600;
}

/* Detail sections */
.detail-section {
  background: #0f172a;
  border-radius: 10px;
  padding: 1rem;
  margin-bottom: 1rem;
}

.detail-section h4 {
  font-size: 0.95rem;
  margin-bottom: 0.75rem;
  color: #cbd5e1;
}

.skill-current {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  margin-bottom: 0.75rem;
  font-size: 0.85rem;
  color: #94a3b8;
}

.skill-edit-row {
  display: flex;
  gap: 0.5rem;
  align-items: center;
  flex-wrap: wrap;
  margin-bottom: 0.5rem;
}

.skill-select {
  padding: 0.4rem 0.6rem;
  border: 1px solid #475569;
  border-radius: 6px;
  background: #1e293b;
  color: #e2e8f0;
  font-size: 0.85rem;
  flex: 1;
  min-width: 140px;
}

.skill-select:focus {
  outline: none;
  border-color: #38bdf8;
}

.star-picker {
  display: flex;
  gap: 2px;
}

.star-btn {
  background: none;
  border: none;
  font-size: 1.2rem;
  cursor: pointer;
  color: #475569;
  padding: 0.1rem;
  transition: color 0.15s;
}

.star-btn.active {
  color: #fbbf24;
}

.star-btn:hover {
  color: #fde047;
}

.btn-save-skill {
  padding: 0.4rem 1rem;
  background: #065f46;
  border: none;
  border-radius: 6px;
  color: #4ade80;
  font-size: 0.85rem;
  font-weight: 600;
  cursor: pointer;
  transition: background 0.2s;
}

.btn-save-skill:hover {
  background: #047857;
}

.skill-preview {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  font-size: 0.85rem;
  color: #94a3b8;
}

.role-change-row {
  display: flex;
  gap: 0.5rem;
}

.role-select-lg {
  padding: 0.5rem 0.75rem;
  border: 1px solid #475569;
  border-radius: 8px;
  background: #1e293b;
  color: #e2e8f0;
  font-size: 0.9rem;
  cursor: pointer;
  flex: 1;
}

.role-select-lg:focus {
  outline: none;
  border-color: #38bdf8;
}

/* Panel transition */
.panel-enter-active,
.panel-leave-active {
  transition: all 0.25s ease;
}

.panel-enter-from .detail-panel,
.panel-leave-to .detail-panel {
  transform: translateX(100%);
}

.panel-enter-from,
.panel-leave-to {
  opacity: 0;
}

/* Responsive */
@media (max-width: 768px) {
  .page-header {
    flex-direction: column;
    gap: 0.75rem;
  }

  .page-header nav {
    flex-wrap: wrap;
    justify-content: center;
    gap: 0.75rem;
  }

  .role-stats {
    gap: 0.5rem;
  }

  .role-stat {
    padding: 0.5rem 0.75rem;
    min-width: 60px;
  }

  .stat-num {
    font-size: 1.1rem;
  }

  .detail-info-grid {
    grid-template-columns: 1fr;
  }
}
</style>
