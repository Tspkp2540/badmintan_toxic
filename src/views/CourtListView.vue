<script setup lang="ts">
import { ref, onMounted, onUnmounted } from 'vue'
import { useAuthStore } from '@/stores/auth'
import { courtService, type Court } from '@/services/courtService'

const authStore = useAuthStore()
const courts = ref<Court[]>([])
const loading = ref(false)
const errorMsg = ref('')

// Create Court Dialog (admin only)
const showCreateDialog = ref(false)
const newCourtName = ref('')
const newDescription = ref('')
const newLocation = ref('')
const newMaxRooms = ref(10)

// Edit Court Dialog
const showEditDialog = ref(false)
const editCourt = ref<Court | null>(null)
const editName = ref('')
const editDescription = ref('')
const editLocation = ref('')
const editMaxRooms = ref(10)
const editStatus = ref('open')

async function loadCourts() {
  loading.value = true
  errorMsg.value = ''
  try {
    courts.value = await courtService.getCourts()
  } catch {
    errorMsg.value = 'โหลดรายการสนามไม่สำเร็จ'
  } finally {
    loading.value = false
  }
}

function openCreateDialog() {
  newCourtName.value = ''
  newDescription.value = ''
  newLocation.value = ''
  newMaxRooms.value = 10
  showCreateDialog.value = true
}

async function createCourt() {
  if (!newCourtName.value.trim()) return
  errorMsg.value = ''
  try {
    await courtService.createCourt({
      name: newCourtName.value.trim(),
      description: newDescription.value.trim(),
      location: newLocation.value.trim(),
      maxRooms: newMaxRooms.value,
    })
    showCreateDialog.value = false
    await loadCourts()
  } catch (e: any) {
    errorMsg.value = e.response?.data?.message || 'สร้างสนามไม่สำเร็จ'
  }
}

function openEditDialog(court: Court) {
  editCourt.value = court
  editName.value = court.name
  editDescription.value = court.description
  editLocation.value = court.location
  editMaxRooms.value = court.maxRooms
  editStatus.value = court.status
  showEditDialog.value = true
}

async function updateCourt() {
  if (!editCourt.value) return
  errorMsg.value = ''
  try {
    await courtService.updateCourt(editCourt.value.id, {
      name: editName.value.trim(),
      description: editDescription.value.trim(),
      location: editLocation.value.trim(),
      maxRooms: editMaxRooms.value,
      status: editStatus.value,
    })
    showEditDialog.value = false
    await loadCourts()
  } catch (e: any) {
    errorMsg.value = e.response?.data?.message || 'แก้ไขสนามไม่สำเร็จ'
  }
}

async function deleteCourt(court: Court) {
  if (!confirm(`ต้องการปิดสนาม "${court.name}" หรือไม่?`)) return
  errorMsg.value = ''
  try {
    await courtService.deleteCourt(court.id)
    await loadCourts()
  } catch (e: any) {
    errorMsg.value = e.response?.data?.message || 'ปิดสนามไม่สำเร็จ'
  }
}

function statusLabel(s: string) {
  if (s === 'open') return 'เปิด'
  if (s === 'closed') return 'ปิด'
  if (s === 'maintenance') return 'ซ่อมบำรุง'
  return s
}

function statusIcon(s: string) {
  if (s === 'open') return '🟢'
  if (s === 'closed') return '🔴'
  return '🟡'
}

let pollInterval: ReturnType<typeof setInterval> | undefined
onMounted(() => {
  loadCourts()
  pollInterval = setInterval(loadCourts, 8000)
})
onUnmounted(() => {
  if (pollInterval) clearInterval(pollInterval)
})
</script>

<template>
  <div class="court-list-page">
    <header class="page-header">
      <div class="header-left">
        <router-link to="/dashboard" class="back-link">← กลับ</router-link>
        <h1>🏸 เลือกสนาม</h1>
      </div>
      <button v-if="authStore.isAdmin" class="btn-create" @click="openCreateDialog">
        + สร้างสนาม
      </button>
    </header>

    <main class="content">
      <p class="page-subtitle">เลือกสนามที่ต้องการเข้าร่วม เหมือนเลือก Server ในเกม!</p>

      <p v-if="errorMsg" class="error-text">{{ errorMsg }}</p>
      <p v-if="loading && courts.length === 0" class="loading-text">กำลังโหลด...</p>

      <div v-if="courts.length === 0 && !loading" class="empty-state">
        <span class="empty-icon">🏟️</span>
        <p>ยังไม่มีสนาม</p>
        <p class="hint" v-if="authStore.isAdmin">กด "สร้างสนาม" เพื่อเปิดสนามใหม่</p>
        <p class="hint" v-else>รอ Admin สร้างสนามก่อนนะ</p>
      </div>

      <div class="courts-grid">
        <div
          v-for="court in courts"
          :key="court.id"
          class="court-card"
          :class="{ closed: court.status !== 'open' }"
        >
          <div class="court-card-top">
            <div class="court-status">
              {{ statusIcon(court.status) }} {{ statusLabel(court.status) }}
            </div>
            <div v-if="authStore.isAdmin" class="court-admin-actions">
              <button class="btn-icon" @click.stop="openEditDialog(court)" title="แก้ไข">✏️</button>
              <button
                v-if="court.status === 'open'"
                class="btn-icon btn-icon-danger"
                @click.stop="deleteCourt(court)"
                title="ปิดสนาม"
              >🔒</button>
            </div>
          </div>

          <div class="court-card-body">
            <h3>{{ court.name }}</h3>
            <p v-if="court.description" class="court-desc">{{ court.description }}</p>
            <p v-if="court.location" class="court-location">📍 {{ court.location }}</p>
            <p v-if="court.bonusExpPercent > 0" class="court-bonus">🎁 EXP โบนัส +{{ court.bonusExpPercent }}%</p>
            <p v-if="court.leaders && court.leaders.length > 0" class="court-leaders">
              👑 หัวก๊วน: {{ court.leaders.map(l => l.fullName).join(', ') }}
            </p>
          </div>

          <div class="court-card-stats">
            <div class="stat">
              <span class="stat-value">{{ court.activeRooms }}/{{ court.maxRooms }}</span>
              <span class="stat-label">ห้อง</span>
            </div>
            <div class="stat">
              <span class="stat-value">{{ court.totalPlayers }}</span>
              <span class="stat-label">ผู้เล่น</span>
            </div>
          </div>

          <div class="court-card-footer">
            <router-link
              v-if="court.status === 'open'"
              :to="`/court/${court.id}`"
              class="btn-enter"
            >
              🚀 เข้าสนาม
            </router-link>
            <span v-else class="closed-text">สนามปิดอยู่</span>
          </div>
        </div>
      </div>
    </main>

    <!-- Create Court Modal -->
    <Teleport to="body">
      <div v-if="showCreateDialog" class="modal-overlay" @click.self="showCreateDialog = false">
        <div class="modal">
          <h2>🏟️ สร้างสนามใหม่</h2>

          <div class="form-group">
            <label>ชื่อสนาม *</label>
            <input v-model="newCourtName" type="text" placeholder="เช่น สนามแบดมินตัน A" class="form-input" maxlength="50" />
          </div>

          <div class="form-group">
            <label>คำอธิบาย</label>
            <input v-model="newDescription" type="text" placeholder="เช่น สนามสำหรับก๊วนวันอังคาร" class="form-input" maxlength="200" />
          </div>

          <div class="form-group">
            <label>สถานที่</label>
            <input v-model="newLocation" type="text" placeholder="เช่น อาคาร B ชั้น 3" class="form-input" maxlength="100" />
          </div>

          <div class="form-group">
            <label>จำนวนห้องสูงสุด</label>
            <input v-model.number="newMaxRooms" type="number" min="1" max="50" class="form-input" />
          </div>

          <div class="modal-actions">
            <button class="btn-cancel" @click="showCreateDialog = false">ยกเลิก</button>
            <button class="btn-confirm" @click="createCourt" :disabled="!newCourtName.trim()">สร้างสนาม</button>
          </div>
        </div>
      </div>
    </Teleport>

    <!-- Edit Court Modal -->
    <Teleport to="body">
      <div v-if="showEditDialog" class="modal-overlay" @click.self="showEditDialog = false">
        <div class="modal">
          <h2>✏️ แก้ไขสนาม</h2>

          <div class="form-group">
            <label>ชื่อสนาม *</label>
            <input v-model="editName" type="text" class="form-input" maxlength="50" />
          </div>

          <div class="form-group">
            <label>คำอธิบาย</label>
            <input v-model="editDescription" type="text" class="form-input" maxlength="200" />
          </div>

          <div class="form-group">
            <label>สถานที่</label>
            <input v-model="editLocation" type="text" class="form-input" maxlength="100" />
          </div>

          <div class="form-group">
            <label>จำนวนห้องสูงสุด</label>
            <input v-model.number="editMaxRooms" type="number" min="1" max="50" class="form-input" />
          </div>

          <div class="form-group">
            <label>สถานะ</label>
            <div class="radio-group">
              <label class="radio-option" :class="{ selected: editStatus === 'open' }">
                <input type="radio" v-model="editStatus" value="open" />
                <span>🟢 เปิด</span>
              </label>
              <label class="radio-option" :class="{ selected: editStatus === 'maintenance' }">
                <input type="radio" v-model="editStatus" value="maintenance" />
                <span>🟡 ซ่อมบำรุง</span>
              </label>
              <label class="radio-option" :class="{ selected: editStatus === 'closed' }">
                <input type="radio" v-model="editStatus" value="closed" />
                <span>🔴 ปิด</span>
              </label>
            </div>
          </div>

          <div class="modal-actions">
            <button class="btn-cancel" @click="showEditDialog = false">ยกเลิก</button>
            <button class="btn-confirm" @click="updateCourt" :disabled="!editName.trim()">บันทึก</button>
          </div>
        </div>
      </div>
    </Teleport>
  </div>
</template>

<style scoped>
.court-list-page {
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

.header-left {
  display: flex;
  align-items: center;
  gap: 1.5rem;
}

.back-link {
  color: #64748b;
  text-decoration: none;
  font-size: 0.9rem;
}

.back-link:hover {
  color: #38bdf8;
}

.page-header h1 {
  font-size: 1.4rem;
  color: #38bdf8;
}

.btn-create {
  background: linear-gradient(135deg, #38bdf8, #818cf8);
  color: #fff;
  border: none;
  padding: 0.6rem 1.2rem;
  border-radius: 8px;
  font-weight: 600;
  cursor: pointer;
  transition: transform 0.2s, opacity 0.2s;
}

.btn-create:hover {
  transform: translateY(-1px);
  opacity: 0.9;
}

.content {
  max-width: 1100px;
  margin: 2rem auto;
  padding: 0 1rem;
}

.page-subtitle {
  color: #94a3b8;
  margin-bottom: 1.5rem;
  font-size: 0.95rem;
}

.error-text {
  color: #f87171;
  margin-bottom: 1rem;
}

.loading-text {
  color: #94a3b8;
  text-align: center;
  padding: 3rem;
}

.empty-state {
  text-align: center;
  padding: 4rem 2rem;
  color: #64748b;
}

.empty-icon {
  font-size: 4rem;
  display: block;
  margin-bottom: 1rem;
}

.hint {
  font-size: 0.85rem;
  margin-top: 0.5rem;
}

/* ===== Court Cards Grid ===== */
.courts-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(320px, 1fr));
  gap: 1.25rem;
}

.court-card {
  background: #1e293b;
  border: 2px solid #334155;
  border-radius: 16px;
  overflow: hidden;
  transition: all 0.2s;
  display: flex;
  flex-direction: column;
}

.court-card:hover {
  border-color: #475569;
  transform: translateY(-2px);
  box-shadow: 0 8px 30px rgba(0, 0, 0, 0.3);
}

.court-card.closed {
  opacity: 0.5;
}

.court-card-top {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 0.75rem 1rem;
  background: rgba(0, 0, 0, 0.2);
}

.court-status {
  font-size: 0.8rem;
  font-weight: 600;
}

.court-admin-actions {
  display: flex;
  gap: 0.4rem;
}

.btn-icon {
  background: none;
  border: none;
  cursor: pointer;
  font-size: 0.9rem;
  padding: 0.2rem 0.4rem;
  border-radius: 4px;
  transition: background 0.2s;
}

.btn-icon:hover {
  background: rgba(255, 255, 255, 0.1);
}

.btn-icon-danger:hover {
  background: rgba(239, 68, 68, 0.2);
}

.court-card-body {
  padding: 1.25rem 1rem;
  flex: 1;
}

.court-card-body h3 {
  font-size: 1.2rem;
  margin-bottom: 0.5rem;
  color: #f1f5f9;
}

.court-desc {
  color: #94a3b8;
  font-size: 0.85rem;
  margin-bottom: 0.4rem;
}

.court-location {
  color: #64748b;
  font-size: 0.8rem;
}

.court-card-stats {
  display: flex;
  gap: 2rem;
  padding: 0.75rem 1rem;
  border-top: 1px solid #334155;
}

.stat {
  display: flex;
  flex-direction: column;
  align-items: center;
}

.stat .stat-value {
  font-size: 1.2rem;
  font-weight: 700;
  color: #38bdf8;
}

.stat .stat-label {
  font-size: 0.75rem;
  color: #64748b;
}

.court-card-footer {
  padding: 0.75rem 1rem;
  border-top: 1px solid #334155;
}

.btn-enter {
  display: block;
  text-align: center;
  padding: 0.65rem;
  background: linear-gradient(135deg, #38bdf8, #818cf8);
  color: #fff;
  text-decoration: none;
  border-radius: 8px;
  font-weight: 600;
  transition: opacity 0.2s;
}

.btn-enter:hover {
  opacity: 0.9;
}

.closed-text {
  display: block;
  text-align: center;
  color: #64748b;
  font-size: 0.9rem;
  padding: 0.65rem;
}

/* ===== Modal ===== */
.modal-overlay {
  position: fixed;
  inset: 0;
  background: rgba(0, 0, 0, 0.7);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 1000;
  backdrop-filter: blur(4px);
}

.modal {
  background: #1e293b;
  border-radius: 16px;
  padding: 2rem;
  width: 90%;
  max-width: 480px;
  max-height: 90vh;
  overflow-y: auto;
}

.modal h2 {
  margin-bottom: 1.5rem;
  color: #e2e8f0;
}

.form-group {
  margin-bottom: 1.25rem;
}

.form-group label {
  display: block;
  font-size: 0.85rem;
  color: #94a3b8;
  margin-bottom: 0.4rem;
  font-weight: 600;
}

.form-input {
  width: 100%;
  padding: 0.6rem 0.75rem;
  background: #0f172a;
  border: 2px solid #334155;
  border-radius: 8px;
  color: #e2e8f0;
  font-size: 1rem;
  outline: none;
  transition: border-color 0.2s;
  box-sizing: border-box;
}

.form-input:focus {
  border-color: #38bdf8;
}

.radio-group {
  display: flex;
  gap: 0.75rem;
}

.radio-option {
  flex: 1;
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 0.6rem;
  background: #0f172a;
  border: 2px solid #334155;
  border-radius: 8px;
  cursor: pointer;
  transition: all 0.2s;
  font-size: 0.85rem;
}

.radio-option input {
  display: none;
}

.radio-option.selected {
  border-color: #38bdf8;
  background: rgba(56, 189, 248, 0.1);
  color: #38bdf8;
}

.modal-actions {
  display: flex;
  gap: 0.75rem;
  margin-top: 1.5rem;
}

.btn-cancel {
  flex: 1;
  padding: 0.65rem;
  background: #334155;
  color: #94a3b8;
  border: none;
  border-radius: 8px;
  cursor: pointer;
  font-weight: 600;
}

.btn-cancel:hover {
  background: #475569;
}

.btn-confirm {
  flex: 1;
  padding: 0.65rem;
  background: linear-gradient(135deg, #38bdf8, #818cf8);
  color: #fff;
  border: none;
  border-radius: 8px;
  cursor: pointer;
  font-weight: 600;
  transition: opacity 0.2s;
}

.btn-confirm:hover {
  opacity: 0.9;
}

.btn-confirm:disabled {
  opacity: 0.4;
  cursor: not-allowed;
}

@media (max-width: 768px) {
  .courts-grid {
    grid-template-columns: 1fr;
  }
}
</style>
