<script setup lang="ts">
import { ref, computed, onMounted, onUnmounted } from 'vue'
import { useRoute } from 'vue-router'
import { useCourtStore, type MatchRoom, type MatchSet } from '@/stores/court'
import { useAuthStore } from '@/stores/auth'
import { courtService, type Court } from '@/services/courtService'
import { createCourtSSE, type SSEConnection } from '@/services/sseService'
import RankBadge from '@/components/RankBadge.vue'
import LevelBadge from '@/components/LevelBadge.vue'
import SkillBadge from '@/components/SkillBadge.vue'
import ConfirmModal from '@/components/ConfirmModal.vue'
import ErrorAlert from '@/components/ErrorAlert.vue'

const route = useRoute()
const courtStore = useCourtStore()
const authStore = useAuthStore()

const courtId = computed(() => route.params.id as string)
const courtInfo = ref<Court | null>(null)
const currentUser = computed(() => authStore.user)

// === Load court info & rooms on mount ===
let sseConn: SSEConnection | null = null

onMounted(async () => {
  try {
    courtInfo.value = await courtService.getCourt(courtId.value)
  } catch {
    // court not found
  }
  courtStore.fetchRoomsByCourt(courtId.value)

  // Connect to SSE for real-time updates
  sseConn = createCourtSSE(courtId.value)

  sseConn.on('room_created', () => {
    courtStore.fetchRoomsByCourt(courtId.value)
  })

  sseConn.on('room_updated', (data: { roomId: string }) => {
    courtStore.fetchRoom(data.roomId)
    courtStore.fetchRoomsByCourt(courtId.value)
  })

  sseConn.on('scores_submitted', (data: { roomId: string }) => {
    courtStore.fetchRoom(data.roomId)
    courtStore.fetchRoomsByCourt(courtId.value)
    // Refresh profile since EXP/rank may have changed
    authStore.fetchProfile()
  })

  sseConn.connect()

  // Fallback poll every 30s in case SSE disconnects
  pollInterval = setInterval(() => {
    courtStore.fetchRoomsByCourt(courtId.value)
  }, 30000)
})

let pollInterval: ReturnType<typeof setInterval> | undefined
onUnmounted(() => {
  if (pollInterval) clearInterval(pollInterval)
  if (sseConn) {
    sseConn.disconnect()
    sseConn = null
  }
})

// === Create Room Dialog ===
const showCreateDialog = ref(false)
const newRoomName = ref('')
const newMatchType = ref<'singles' | 'doubles'>('singles')
const newMatchMode = ref<'casual' | 'ranked' | 'skill_test'>('casual')
const newMaxSets = ref(3)

const isCourtLeader = computed(() => {
  if (!currentUser.value || !courtInfo.value) return false
  if (authStore.isAdmin) return true
  return courtInfo.value.leaders?.some(l => l.userId === currentUser.value!.id) ?? false
})

function openCreateDialog() {
  newRoomName.value = ''
  newMatchType.value = 'singles'
  newMatchMode.value = 'casual'
  newMaxSets.value = 3
  showCreateDialog.value = true
}

async function createRoom() {
  if (!newRoomName.value.trim()) return
  await courtStore.createRoom({
    courtId: courtId.value,
    name: newRoomName.value.trim(),
    matchType: newMatchType.value,
    matchMode: newMatchMode.value,
    maxSets: newMaxSets.value,
  })
  showCreateDialog.value = false
}

// === Room View ===
const activeRoom = computed(() => courtStore.activeRoom)

function isCurrentUserInRoom(room: MatchRoom): boolean {
  if (!currentUser.value) return false
  return (
    room.players.some((p) => p.userId === currentUser.value!.id) ||
    room.referee?.id === currentUser.value.id
  )
}

function teamPlayers(room: MatchRoom, team: 'A' | 'B') {
  return room.players.filter((p) => p.team === team)
}

function canStart(room: MatchRoom): boolean {
  const teamA = room.players.filter((p) => p.team === 'A')
  const teamB = room.players.filter((p) => p.team === 'B')
  if (teamA.length === 0 || teamB.length === 0) return false
  if (room.matchType === 'doubles') {
    if (teamA.length < 2 || teamB.length < 2) return false
  }
  return true
}

function maxPerTeam(room: MatchRoom): number {
  return room.matchType === 'singles' ? 1 : 2
}

function canJoinTeam(room: MatchRoom, team: 'A' | 'B'): boolean {
  if (!currentUser.value) return false
  if (isCurrentUserInRoom(room)) return false
  if (room.status !== 'waiting') return false
  const count = room.players.filter((p) => p.team === team).length
  return count < maxPerTeam(room)
}

// === Timer ===
const now = ref(Date.now())
let timerInterval: ReturnType<typeof setInterval> | undefined

onMounted(() => {
  timerInterval = setInterval(() => {
    now.value = Date.now()
  }, 1000)
})

onUnmounted(() => {
  if (timerInterval) clearInterval(timerInterval)
})

function formatTime(ms: number): string {
  const totalSec = Math.floor(ms / 1000)
  const min = Math.floor(totalSec / 60)
  const sec = totalSec % 60
  return `${min.toString().padStart(2, '0')}:${sec.toString().padStart(2, '0')}`
}

function elapsedTime(room: MatchRoom): number {
  if (!room.startedAt || room.status !== 'playing') return 0
  const started = new Date(room.startedAt).getTime()
  return Math.max(0, now.value - started)
}

// === Scoring ===
const scoreSets = ref<MatchSet[]>([])

function initScoring() {
  if (!activeRoom.value) return
  scoreSets.value = Array.from({ length: activeRoom.value.maxSets }, (_, i) => ({
    setNumber: i + 1,
    teamA: 0,
    teamB: 0,
  }))
}

function submitScores() {
  if (!activeRoom.value) return
  const validSets = scoreSets.value.filter((s) => s.teamA > 0 || s.teamB > 0)
  if (validSets.length === 0) return
  courtStore.submitScores(activeRoom.value.id, validSets)
}

// === Confirm Modals ===
const confirmAction = ref<'start' | 'end' | 'submit' | 'leave' | null>(null)
const confirmLoading = ref(false)

const confirmConfig = computed(() => {
  switch (confirmAction.value) {
    case 'start':
      return { title: '🚀 เริ่มเกม', message: 'ต้องการเริ่มเกมนี้หรือไม่?', variant: 'info' as const, confirmText: 'เริ่มเกม' }
    case 'end':
      return { title: '🏁 จบเกม', message: 'ต้องการจบเกมนี้และเข้าสู่การกรอกคะแนนหรือไม่?', variant: 'warning' as const, confirmText: 'จบเกม' }
    case 'submit':
      return { title: '📊 ส่งคะแนน', message: 'ยืนยันการส่งคะแนน? ผลนี้จะถูกบันทึกถาวร', variant: 'warning' as const, confirmText: 'ส่งคะแนน' }
    case 'leave':
      return { title: 'ออกจากห้อง', message: 'ต้องการออกจากห้องนี้หรือไม่?', variant: 'danger' as const, confirmText: 'ออกจากห้อง' }
    default:
      return { title: '', message: '', variant: 'info' as const, confirmText: '' }
  }
})

async function handleConfirm() {
  if (!activeRoom.value) return
  confirmLoading.value = true
  try {
    switch (confirmAction.value) {
      case 'start':
        await courtStore.startGame(activeRoom.value.id)
        break
      case 'end':
        await courtStore.endGame(activeRoom.value.id)
        break
      case 'submit':
        submitScores()
        break
      case 'leave':
        await courtStore.leaveRoom(activeRoom.value.id)
        break
    }
  } finally {
    confirmLoading.value = false
    confirmAction.value = null
  }
}

import { watch } from 'vue'
watch(
  () => activeRoom.value?.status,
  (status) => {
    if (status === 'scoring') {
      initScoring()
    }
  },
)
</script>

<template>
  <div class="court-page">
    <header class="page-header">
      <div class="header-left">
        <router-link to="/courts" class="back-link">← กลับ</router-link>
        <h1>🏸 {{ courtInfo?.name ?? 'สนามแบดมินตัน' }}</h1>
        <span v-if="courtInfo?.location" class="court-loc">📍 {{ courtInfo.location }}</span>
      </div>
      <button class="btn-create" @click="openCreateDialog">+ สร้างห้อง</button>
    </header>

    <main class="court-content">
      <ErrorAlert :message="courtStore.error" @close="courtStore.error = null" />
      <div class="layout">
        <!-- Room List (left) -->
        <section class="room-list-section">
          <h2>ห้องแข่งขัน</h2>

          <div v-if="courtStore.rooms.length === 0" class="empty-state">
            <p>ยังไม่มีห้อง</p>
            <p class="hint">กด "สร้างห้อง" เพื่อเริ่มแข่งขัน</p>
          </div>

          <div
            v-for="room in courtStore.rooms"
            :key="room.id"
            class="room-card"
            :class="{
              active: activeRoom?.id === room.id,
              playing: room.status === 'playing',
              finished: room.status === 'finished',
            }"
            @click="courtStore.setActiveRoom(room.id)"
          >
            <div class="room-card-header">
              <span class="room-name">{{ room.name }}</span>
              <span class="room-status-badge" :class="room.status">
                {{ room.status === 'waiting' ? 'รอผู้เล่น' : room.status === 'playing' ? 'กำลังแข่ง' : room.status === 'scoring' ? 'กรอกคะแนน' : 'จบแล้ว' }}
              </span>
            </div>
            <div class="room-card-info">
              <span>{{ room.matchType === 'singles' ? '1v1' : '2v2' }}</span>
              <span :class="room.matchMode">{{ room.matchMode === 'ranked' ? '⚔️ แรงค์' : room.matchMode === 'skill_test' ? '🎯 ทดสอบระดับ' : '🎮 แคชชวล' }}</span>
              <span>👥 {{ room.players.length }}/{{ room.matchType === 'singles' ? 2 : 4 }}</span>
            </div>
          </div>
        </section>

        <!-- Active Room Detail (right) -->
        <section v-if="activeRoom" class="room-detail-section">
          <!-- Room Header -->
          <div class="room-header">
            <div>
              <h2>{{ activeRoom.name }}</h2>
              <div class="room-meta">
                <span class="badge" :class="activeRoom.matchType">{{ activeRoom.matchType === 'singles' ? 'เดี่ยว (1v1)' : 'คู่ (2v2)' }}</span>
                <span class="badge" :class="activeRoom.matchMode">{{ activeRoom.matchMode === 'ranked' ? '⚔️ วัดแรงค์' : activeRoom.matchMode === 'skill_test' ? '🎯 ทดสอบระดับ' : '🎮 แคชชวล' }}</span>
                <span class="badge sets">Best of {{ activeRoom.maxSets }}</span>
              </div>
            </div>
            <div v-if="activeRoom.status === 'playing'" class="timer">
              ⏱️ {{ formatTime(elapsedTime(activeRoom)) }}
            </div>
          </div>

          <!-- Referee -->
          <div class="referee-section">
            <span class="referee-label">🧑‍⚖️ กรรมการ:</span>
            <span v-if="activeRoom.referee" class="referee-name">{{ activeRoom.referee.fullName }}</span>
            <span v-else class="no-referee">ไม่มี</span>
            <button
              v-if="!activeRoom.referee && !isCurrentUserInRoom(activeRoom) && activeRoom.status === 'waiting'"
              class="btn-sm btn-referee"
              @click="courtStore.joinAsReferee(activeRoom.id)"
            >
              เป็นกรรมการ
            </button>
          </div>

          <!-- Court Visualization -->
          <div class="court-container">
            <div class="court">
              <!-- Team A Side -->
              <div class="court-side side-a">
                <h3>ทีม A</h3>
                <div class="player-slots">
                  <div
                    v-for="player in teamPlayers(activeRoom, 'A')"
                    :key="player.userId"
                    class="player-slot filled"
                  >
                    <div class="slot-avatar">{{ player.fullName.charAt(0) }}</div>
                    <div class="slot-info">
                      <span class="slot-name">{{ player.fullName }}</span>
                      <div class="slot-badges">
                        <LevelBadge :level="player.level" />
                        <RankBadge :rank="player.rank" />
                        <SkillBadge :skill-level="player.skillLevel" :skill-stars="player.skillStars" />
                      </div>
                    </div>
                  </div>
                  <div
                    v-for="i in (maxPerTeam(activeRoom) - teamPlayers(activeRoom, 'A').length)"
                    :key="'emptyA-' + i"
                    class="player-slot empty"
                  >
                    <button
                      v-if="canJoinTeam(activeRoom, 'A')"
                      class="btn-join"
                      @click="courtStore.joinRoom(activeRoom.id, 'A')"
                    >
                      + เข้าร่วม
                    </button>
                    <span v-else class="waiting-text">รอผู้เล่น...</span>
                  </div>
                </div>
              </div>

              <!-- Net -->
              <div class="net">
                <div class="net-line"></div>
                <span class="vs-text">VS</span>
                <div class="net-line"></div>
              </div>

              <!-- Team B Side -->
              <div class="court-side side-b">
                <h3>ทีม B</h3>
                <div class="player-slots">
                  <div
                    v-for="player in teamPlayers(activeRoom, 'B')"
                    :key="player.userId"
                    class="player-slot filled"
                  >
                    <div class="slot-avatar">{{ player.fullName.charAt(0) }}</div>
                    <div class="slot-info">
                      <span class="slot-name">{{ player.fullName }}</span>
                      <div class="slot-badges">
                        <LevelBadge :level="player.level" />
                        <RankBadge :rank="player.rank" />
                        <SkillBadge :skill-level="player.skillLevel" :skill-stars="player.skillStars" />
                      </div>
                    </div>
                  </div>
                  <div
                    v-for="i in (maxPerTeam(activeRoom) - teamPlayers(activeRoom, 'B').length)"
                    :key="'emptyB-' + i"
                    class="player-slot empty"
                  >
                    <button
                      v-if="canJoinTeam(activeRoom, 'B')"
                      class="btn-join"
                      @click="courtStore.joinRoom(activeRoom.id, 'B')"
                    >
                      + เข้าร่วม
                    </button>
                    <span v-else class="waiting-text">รอผู้เล่น...</span>
                  </div>
                </div>
              </div>
            </div>
          </div>

          <!-- Action Area -->
          <div class="action-area">
            <!-- Waiting state -->
            <template v-if="activeRoom.status === 'waiting'">
              <button
                v-if="isCurrentUserInRoom(activeRoom) && canStart(activeRoom)"
                class="btn-action btn-start"
                @click="confirmAction = 'start'"
              >
                🚀 เริ่มเกม
              </button>
              <button
                v-if="isCurrentUserInRoom(activeRoom)"
                class="btn-action btn-leave"
                @click="confirmAction = 'leave'"
              >
                ออกจากห้อง
              </button>
            </template>

            <!-- Playing state -->
            <template v-if="activeRoom.status === 'playing'">
              <div class="playing-info">
                <p>🏸 กำลังแข่งขัน...</p>
              </div>
              <div class="playing-actions" v-if="isCurrentUserInRoom(activeRoom)">
                <button class="btn-action btn-end" @click="confirmAction = 'end'">
                  🏁 จบเกม
                </button>
              </div>
            </template>

            <!-- Scoring state -->
            <template v-if="activeRoom.status === 'scoring'">
              <div class="scoring-section">
                <h3>📝 กรอกคะแนน</h3>
                <div class="score-table">
                  <div class="score-header">
                    <span>เซ็ต</span>
                    <span>ทีม A</span>
                    <span>ทีม B</span>
                  </div>
                  <div v-for="set in scoreSets" :key="set.setNumber" class="score-row">
                    <span class="set-number">เซ็ตที่ {{ set.setNumber }}</span>
                    <input
                      v-model.number="set.teamA"
                      type="number"
                      min="0"
                      max="30"
                      class="score-input"
                      placeholder="0"
                    />
                    <input
                      v-model.number="set.teamB"
                      type="number"
                      min="0"
                      max="30"
                      class="score-input"
                      placeholder="0"
                    />
                  </div>
                </div>
                <button class="btn-action btn-submit" @click="confirmAction = 'submit'">
                  📊 ส่งคะแนน
                </button>
              </div>
            </template>

            <!-- Finished state -->
            <template v-if="activeRoom.status === 'finished'">
              <div class="result-section">
                <h3>🏆 ผลการแข่งขัน</h3>
                <div class="result-winner" :class="{ draw: activeRoom.winnerTeam === 'draw' }">
                  {{ activeRoom.winnerTeam === 'draw' ? '🤝 เสมอ!' : `🎉 ทีม ${activeRoom.winnerTeam} ชนะ!` }}
                </div>

                <!-- Score Summary -->
                <div class="score-summary">
                  <div v-for="set in activeRoom.sets" :key="set.setNumber" class="score-result-row">
                    <span>เซ็ตที่ {{ set.setNumber }}</span>
                    <span class="score-vs">
                      <span :class="{ winner: set.teamA > set.teamB }">{{ set.teamA }}</span>
                      <span class="dash">-</span>
                      <span :class="{ winner: set.teamB > set.teamA }">{{ set.teamB }}</span>
                    </span>
                  </div>
                </div>

                <!-- Rewards -->
                <div class="rewards-section">
                  <h4>🎁 รางวัล</h4>
                  <div v-for="player in activeRoom.players" :key="player.userId" class="reward-row">
                    <span class="reward-name">{{ player.fullName }}</span>
                    <span class="reward-exp">+{{ player.expGained }} EXP</span>
                    <span
                      v-if="activeRoom.matchMode === 'ranked' || activeRoom.matchMode === 'skill_test'"
                      class="reward-rp"
                      :class="{ negative: player.rankPointsGained < 0 }"
                    >
                      {{ player.rankPointsGained > 0 ? '+' : '' }}{{ player.rankPointsGained }} RP
                    </span>
                  </div>
                </div>

                <button class="btn-action btn-leave" @click="confirmAction = 'leave'">
                  ออกจากห้อง
                </button>
              </div>
            </template>
          </div>
        </section>

        <!-- No room selected -->
        <section v-else class="room-detail-section empty-detail">
          <div class="empty-room-state">
            <span class="big-icon">🏸</span>
            <p>เลือกห้องจากรายการ หรือสร้างห้องใหม่</p>
          </div>
        </section>
      </div>
    </main>

    <!-- Confirm Modal -->
    <ConfirmModal
      :show="!!confirmAction"
      :title="confirmConfig.title"
      :message="confirmConfig.message"
      :variant="confirmConfig.variant"
      :confirm-text="confirmConfig.confirmText"
      :loading="confirmLoading"
      @confirm="handleConfirm"
      @cancel="confirmAction = null"
    />

    <!-- Create Room Modal -->
    <Teleport to="body">
      <div v-if="showCreateDialog" class="modal-overlay" @click.self="showCreateDialog = false">
        <div class="modal">
          <h2>🏸 สร้างห้องแข่งขัน</h2>

          <div class="form-group">
            <label>ชื่อห้อง</label>
            <input v-model="newRoomName" type="text" placeholder="เช่น สนาม 1" class="form-input" maxlength="50" />
          </div>

          <div class="form-group">
            <label>รูปแบบ</label>
            <div class="radio-group">
              <label class="radio-option" :class="{ selected: newMatchType === 'singles' }">
                <input type="radio" v-model="newMatchType" value="singles" />
                <span>🧑 เดี่ยว (1v1)</span>
              </label>
              <label class="radio-option" :class="{ selected: newMatchType === 'doubles' }">
                <input type="radio" v-model="newMatchType" value="doubles" />
                <span>👥 คู่ (2v2)</span>
              </label>
            </div>
          </div>

          <div class="form-group">
            <label>โหมด</label>
            <div class="radio-group">
              <label class="radio-option" :class="{ selected: newMatchMode === 'casual' }">
                <input type="radio" v-model="newMatchMode" value="casual" />
                <span>🎮 แคชชวล</span>
              </label>
              <label class="radio-option" :class="{ selected: newMatchMode === 'ranked' }">
                <input type="radio" v-model="newMatchMode" value="ranked" />
                <span>⚔️ วัดแรงค์</span>
              </label>
              <label v-if="isCourtLeader" class="radio-option" :class="{ selected: newMatchMode === 'skill_test' }">
                <input type="radio" v-model="newMatchMode" value="skill_test" />
                <span>🎯 ทดสอบระดับ</span>
              </label>
            </div>
          </div>

          <div class="form-group">
            <label>จำนวนเซ็ต (Best of)</label>
            <div class="radio-group">
              <label class="radio-option" :class="{ selected: newMaxSets === 1 }">
                <input type="radio" v-model="newMaxSets" :value="1" />
                <span>1</span>
              </label>
              <label class="radio-option" :class="{ selected: newMaxSets === 3 }">
                <input type="radio" v-model="newMaxSets" :value="3" />
                <span>3</span>
              </label>
              <label class="radio-option" :class="{ selected: newMaxSets === 5 }">
                <input type="radio" v-model="newMaxSets" :value="5" />
                <span>5</span>
              </label>
            </div>
          </div>

          <div class="modal-actions">
            <button class="btn-cancel" @click="showCreateDialog = false">ยกเลิก</button>
            <button class="btn-confirm" @click="createRoom" :disabled="!newRoomName.trim()">สร้างห้อง</button>
          </div>
        </div>
      </div>
    </Teleport>
  </div>
</template>

<style scoped>
.court-page {
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

.court-loc {
  color: #64748b;
  font-size: 0.85rem;
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

.court-content {
  padding: 1.5rem;
}

.layout {
  display: grid;
  grid-template-columns: 320px 1fr;
  gap: 1.5rem;
  max-width: 1200px;
  margin: 0 auto;
}

/* ===== Room List ===== */
.room-list-section {
  display: flex;
  flex-direction: column;
  gap: 0.75rem;
}

.room-list-section h2 {
  font-size: 1.1rem;
  color: #cbd5e1;
  margin-bottom: 0.5rem;
}

.empty-state {
  text-align: center;
  padding: 2rem;
  color: #64748b;
}

.hint {
  font-size: 0.85rem;
  margin-top: 0.5rem;
}

.room-card {
  background: #1e293b;
  border: 2px solid transparent;
  border-radius: 12px;
  padding: 1rem;
  cursor: pointer;
  transition: all 0.2s;
}

.room-card:hover {
  border-color: #334155;
}

.room-card.active {
  border-color: #38bdf8;
  background: #1a2a42;
}

.room-card.playing {
  border-left: 4px solid #4ade80;
}

.room-card.finished {
  opacity: 0.6;
}

.room-card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 0.5rem;
}

.room-name {
  font-weight: 600;
  font-size: 1rem;
}

.room-status-badge {
  font-size: 0.75rem;
  padding: 0.2rem 0.5rem;
  border-radius: 12px;
  font-weight: 600;
}

.room-status-badge.waiting {
  background: #1e3a5f;
  color: #38bdf8;
}

.room-status-badge.playing {
  background: #14532d;
  color: #4ade80;
}

.room-status-badge.scoring {
  background: #422006;
  color: #fbbf24;
}

.room-status-badge.finished {
  background: #334155;
  color: #94a3b8;
}

.room-card-info {
  display: flex;
  gap: 0.75rem;
  font-size: 0.85rem;
  color: #94a3b8;
}

.room-card-info .ranked {
  color: #f87171;
}

.room-card-info .casual {
  color: #4ade80;
}

/* ===== Room Detail ===== */
.room-detail-section {
  background: #1e293b;
  border-radius: 16px;
  padding: 1.5rem;
  min-height: 500px;
}

.empty-detail {
  display: flex;
  align-items: center;
  justify-content: center;
}

.empty-room-state {
  text-align: center;
  color: #64748b;
}

.big-icon {
  font-size: 4rem;
  display: block;
  margin-bottom: 1rem;
}

.room-header {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  margin-bottom: 1rem;
}

.room-header h2 {
  font-size: 1.3rem;
  margin-bottom: 0.5rem;
}

.room-meta {
  display: flex;
  gap: 0.5rem;
}

.badge {
  font-size: 0.75rem;
  padding: 0.25rem 0.6rem;
  border-radius: 12px;
  font-weight: 600;
}

.badge.singles {
  background: #164e63;
  color: #22d3ee;
}

.badge.doubles {
  background: #312e81;
  color: #a5b4fc;
}

.badge.ranked {
  background: #7f1d1d;
  color: #fca5a5;
}

.badge.casual {
  background: #14532d;
  color: #86efac;
}

.badge.sets {
  background: #422006;
  color: #fbbf24;
}

.timer {
  font-size: 1.5rem;
  font-weight: 700;
  color: #4ade80;
  font-variant-numeric: tabular-nums;
}

.timer.urgent {
  color: #f87171;
  animation: pulse 1s ease-in-out infinite;
}

@keyframes pulse {
  0%, 100% { opacity: 1; }
  50% { opacity: 0.5; }
}

/* Referee */
.referee-section {
  display: flex;
  align-items: center;
  gap: 0.75rem;
  padding: 0.75rem 1rem;
  background: #0f172a;
  border-radius: 8px;
  margin-bottom: 1rem;
  font-size: 0.9rem;
}

.referee-label {
  color: #94a3b8;
}

.referee-name {
  color: #fbbf24;
  font-weight: 600;
}

.no-referee {
  color: #475569;
}

.btn-sm {
  padding: 0.3rem 0.7rem;
  border-radius: 6px;
  font-size: 0.8rem;
  border: none;
  cursor: pointer;
  font-weight: 600;
}

.btn-referee {
  background: #422006;
  color: #fbbf24;
}

.btn-referee:hover {
  background: #78350f;
}

/* ===== Court Visualization ===== */
.court-container {
  margin: 1rem 0;
}

.court {
  display: grid;
  grid-template-columns: 1fr auto 1fr;
  gap: 0;
  background: #166534;
  border: 3px solid #fff;
  border-radius: 12px;
  overflow: hidden;
  min-height: 240px;
}

.court-side {
  padding: 1.5rem;
  display: flex;
  flex-direction: column;
  align-items: center;
}

.court-side h3 {
  font-size: 0.9rem;
  margin-bottom: 1rem;
  color: rgba(255, 255, 255, 0.7);
  text-transform: uppercase;
  letter-spacing: 1px;
}

.side-a {
  background: rgba(56, 189, 248, 0.15);
  border-right: 2px solid rgba(255, 255, 255, 0.3);
}

.side-b {
  background: rgba(248, 113, 113, 0.15);
  border-left: 2px solid rgba(255, 255, 255, 0.3);
}

.net {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: 0 0.5rem;
  background: rgba(0, 0, 0, 0.3);
}

.net-line {
  width: 2px;
  flex: 1;
  background: repeating-linear-gradient(
    to bottom,
    #fff 0px,
    #fff 4px,
    transparent 4px,
    transparent 8px
  );
}

.vs-text {
  font-weight: 900;
  font-size: 1.2rem;
  color: #fbbf24;
  padding: 0.5rem 0;
}

.player-slots {
  display: flex;
  flex-direction: column;
  gap: 0.75rem;
  width: 100%;
}

.player-slot {
  background: rgba(0, 0, 0, 0.3);
  border-radius: 10px;
  padding: 0.75rem;
  display: flex;
  align-items: center;
  gap: 0.75rem;
  min-height: 60px;
}

.player-slot.filled {
  background: rgba(255, 255, 255, 0.1);
  backdrop-filter: blur(4px);
}

.player-slot.filled.ready {
  border: 2px solid #4ade80;
}

.player-slot.empty {
  border: 2px dashed rgba(255, 255, 255, 0.2);
  justify-content: center;
}

.slot-avatar {
  width: 40px;
  height: 40px;
  border-radius: 50%;
  background: linear-gradient(135deg, #38bdf8, #818cf8);
  display: flex;
  align-items: center;
  justify-content: center;
  font-weight: 700;
  color: #fff;
  flex-shrink: 0;
}

.slot-info {
  flex: 1;
  min-width: 0;
}

.slot-name {
  font-weight: 600;
  font-size: 0.9rem;
  display: block;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.slot-badges {
  display: flex;
  gap: 0.4rem;
  margin-top: 0.25rem;
  transform: scale(0.85);
  transform-origin: left;
}

.ready-badge {
  font-size: 1.2rem;
}

.btn-join {
  background: rgba(56, 189, 248, 0.2);
  border: 1px solid rgba(56, 189, 248, 0.4);
  color: #38bdf8;
  padding: 0.4rem 1rem;
  border-radius: 6px;
  cursor: pointer;
  font-weight: 600;
  transition: all 0.2s;
}

.btn-join:hover {
  background: rgba(56, 189, 248, 0.3);
}

.waiting-text {
  color: rgba(255, 255, 255, 0.3);
  font-size: 0.85rem;
}

/* ===== Action Area ===== */
.action-area {
  margin-top: 1.5rem;
  display: flex;
  flex-direction: column;
  gap: 1rem;
}

.btn-action {
  padding: 0.7rem 1.5rem;
  border: none;
  border-radius: 8px;
  font-weight: 600;
  font-size: 0.95rem;
  cursor: pointer;
  transition: all 0.2s;
}

.btn-ready {
  background: #14532d;
  color: #4ade80;
}

.btn-ready:hover {
  background: #166534;
}

.btn-start {
  background: linear-gradient(135deg, #38bdf8, #818cf8);
  color: #fff;
}

.btn-start:hover {
  opacity: 0.9;
  transform: translateY(-1px);
}

.btn-leave {
  background: #1e293b;
  color: #94a3b8;
  border: 1px solid #334155;
}

.btn-leave:hover {
  border-color: #f87171;
  color: #f87171;
}

.btn-extend {
  background: #422006;
  color: #fbbf24;
}

.btn-extend:hover {
  background: #78350f;
}

.btn-end {
  background: #7f1d1d;
  color: #fca5a5;
}

.btn-end:hover {
  background: #991b1b;
}

.playing-info {
  text-align: center;
  font-size: 1.1rem;
  color: #4ade80;
}

.playing-actions {
  display: flex;
  gap: 0.75rem;
  justify-content: center;
}

/* ===== Scoring ===== */
.scoring-section {
  background: #0f172a;
  border-radius: 12px;
  padding: 1.5rem;
}

.scoring-section h3 {
  margin-bottom: 1rem;
  color: #fbbf24;
}

.score-table {
  display: flex;
  flex-direction: column;
  gap: 0.5rem;
  margin-bottom: 1rem;
}

.score-header,
.score-row {
  display: grid;
  grid-template-columns: 1fr 1fr 1fr;
  gap: 0.75rem;
  align-items: center;
  text-align: center;
}

.score-header {
  font-size: 0.85rem;
  color: #94a3b8;
  font-weight: 600;
  padding-bottom: 0.5rem;
  border-bottom: 1px solid #334155;
}

.set-number {
  color: #cbd5e1;
  font-weight: 500;
}

.score-input {
  background: #1e293b;
  border: 2px solid #334155;
  color: #e2e8f0;
  padding: 0.5rem;
  border-radius: 8px;
  text-align: center;
  font-size: 1.1rem;
  font-weight: 700;
  width: 100%;
  outline: none;
  transition: border-color 0.2s;
}

.score-input:focus {
  border-color: #38bdf8;
}

.btn-submit {
  background: linear-gradient(135deg, #38bdf8, #818cf8);
  color: #fff;
  width: 100%;
}

.btn-submit:hover {
  opacity: 0.9;
}

/* ===== Result ===== */
.result-section {
  background: #0f172a;
  border-radius: 12px;
  padding: 1.5rem;
  text-align: center;
}

.result-section h3 {
  color: #fbbf24;
  margin-bottom: 1rem;
}

.result-winner {
  font-size: 1.5rem;
  font-weight: 700;
  color: #4ade80;
  margin-bottom: 1.5rem;
}

.result-winner.draw {
  color: #fbbf24;
}

.score-summary {
  display: flex;
  flex-direction: column;
  gap: 0.5rem;
  margin-bottom: 1.5rem;
}

.score-result-row {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 0.5rem 1rem;
  background: #1e293b;
  border-radius: 8px;
}

.score-vs {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  font-size: 1.1rem;
  font-weight: 700;
}

.score-vs .winner {
  color: #4ade80;
}

.dash {
  color: #64748b;
}

.rewards-section {
  margin-bottom: 1.5rem;
}

.rewards-section h4 {
  color: #cbd5e1;
  margin-bottom: 0.75rem;
}

.reward-row {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 0.5rem 1rem;
  background: #1e293b;
  border-radius: 8px;
  margin-bottom: 0.4rem;
}

.reward-name {
  font-weight: 500;
}

.reward-exp {
  color: #38bdf8;
  font-weight: 700;
}

.reward-rp {
  color: #4ade80;
  font-weight: 700;
}

.reward-rp.negative {
  color: #f87171;
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
  font-size: 0.9rem;
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

/* ===== Responsive ===== */
@media (max-width: 768px) {
  .layout {
    grid-template-columns: 1fr;
  }

  .court {
    grid-template-columns: 1fr;
    grid-template-rows: 1fr auto 1fr;
  }

  .side-a {
    border-right: none;
    border-bottom: 2px solid rgba(255, 255, 255, 0.3);
  }

  .side-b {
    border-left: none;
    border-top: 2px solid rgba(255, 255, 255, 0.3);
  }

  .net {
    flex-direction: row;
    padding: 0.5rem;
  }

  .net-line {
    width: auto;
    height: 2px;
    flex: 1;
    background: repeating-linear-gradient(
      to right,
      #fff 0px,
      #fff 4px,
      transparent 4px,
      transparent 8px
    );
  }

  .playing-actions {
    flex-direction: column;
  }

  .page-header {
    flex-direction: column;
    gap: 0.75rem;
    padding: 1rem;
    align-items: stretch;
  }

  .header-left {
    flex-wrap: wrap;
    gap: 0.5rem;
  }

  .page-header h1 {
    font-size: 1.1rem;
  }

  .court-content {
    padding: 0.75rem;
  }

  .room-detail-section {
    padding: 1rem;
    min-height: auto;
  }

  .court-side {
    padding: 1rem;
  }

  .room-header {
    flex-direction: column;
    gap: 0.5rem;
  }

  .room-meta {
    flex-wrap: wrap;
  }

  .scoring-section {
    padding: 1rem;
  }

  .score-input {
    font-size: 1rem;
    padding: 0.4rem;
  }

  .referee-section {
    flex-wrap: wrap;
    font-size: 0.85rem;
  }

  .result-winner {
    font-size: 1.2rem;
  }

  .reward-row {
    font-size: 0.85rem;
    padding: 0.4rem 0.75rem;
  }
}
</style>
