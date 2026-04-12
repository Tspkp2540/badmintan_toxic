<script setup lang="ts">
import { computed } from 'vue'
import { useAuthStore } from '@/stores/auth'
import { getRankByLevel, getExpForLevel, PlayerRank } from '@/models/User'
import LevelBadge from '@/components/LevelBadge.vue'
import RankBadge from '@/components/RankBadge.vue'

const authStore = useAuthStore()

const user = computed(() => authStore.user)
const level = computed(() => user.value?.level ?? 1)
const rank = computed(() => user.value?.rank ?? getRankByLevel(1))
const exp = computed(() => user.value?.exp ?? 0)
const expToNext = computed(() => user.value?.expToNextLevel ?? getExpForLevel(1))
const expPercent = computed(() => Math.min((exp.value / expToNext.value) * 100, 100))

const stats = computed(() => [
  { label: 'ชนะ', value: user.value?.wins ?? 0, color: '#4ade80' },
  { label: 'แพ้', value: user.value?.losses ?? 0, color: '#f87171' },
  { label: 'แมตช์ทั้งหมด', value: user.value?.totalMatches ?? 0, color: '#38bdf8' },
  { label: 'อัตราชนะ', value: `${(user.value?.winRate ?? 0).toFixed(1)}%`, color: '#fbbf24' },
])

const rankColors: Record<string, string> = {
  [PlayerRank.BRONZE]: '#cd7f32',
  [PlayerRank.SILVER]: '#c0c0c0',
  [PlayerRank.GOLD]: '#ffd700',
  [PlayerRank.PLATINUM]: '#00cec9',
  [PlayerRank.DIAMOND]: '#a29bfe',
  [PlayerRank.MASTER]: '#fd79a8',
  [PlayerRank.GRANDMASTER]: '#e17055',
}
</script>

<template>
  <div class="dashboard">
    <header class="dashboard-header">
      <h1>🏸 Badminton Hub</h1>
      <nav>
        <router-link to="/dashboard">แดชบอร์ด</router-link>
        <router-link to="/court">สนาม</router-link>
        <router-link to="/ranking">อันดับ</router-link>
        <router-link to="/profile">โปรไฟล์</router-link>
        <button @click="authStore.logout()" class="btn-logout">ออกจากระบบ</button>
      </nav>
    </header>

    <main class="dashboard-content">
      <!-- Player Card -->
      <section class="player-card">
        <div class="player-info">
          <div class="avatar">
            {{ user?.fullName?.charAt(0)?.toUpperCase() ?? '?' }}
          </div>
          <div>
            <h2>{{ user?.fullName ?? 'นักแบดมินตัน' }}</h2>
            <p class="username">@{{ user?.username ?? 'player' }}</p>
          </div>
        </div>

        <div class="level-rank">
          <LevelBadge :level="level" />
          <RankBadge :rank="rank" />
        </div>

        <!-- EXP Bar -->
        <div class="exp-section">
          <div class="exp-label">
            <span>EXP</span>
            <span>{{ exp }} / {{ expToNext }}</span>
          </div>
          <div class="exp-bar">
            <div
              class="exp-fill"
              :style="{
                width: `${expPercent}%`,
                background: rankColors[rank] ?? '#38bdf8',
              }"
            ></div>
          </div>
        </div>
      </section>

      <!-- Stats -->
      <section class="stats-grid">
        <div v-for="stat in stats" :key="stat.label" class="stat-card">
          <span class="stat-value" :style="{ color: stat.color }">{{ stat.value }}</span>
          <span class="stat-label">{{ stat.label }}</span>
        </div>
      </section>

      <!-- Quick Actions -->
      <section class="quick-actions">
        <h3>เมนูด่วน</h3>
        <div class="actions-grid">
          <router-link to="/court" class="action-card">
            <span class="action-icon">🏸</span>
            <span>เข้าสนาม</span>
          </router-link>
          <router-link to="/ranking" class="action-card">
            <span class="action-icon">🏆</span>
            <span>ดูอันดับ</span>
          </router-link>
          <router-link to="/profile" class="action-card">
            <span class="action-icon">👤</span>
            <span>โปรไฟล์</span>
          </router-link>
        </div>
      </section>
    </main>
  </div>
</template>

<style scoped>
.dashboard {
  min-height: 100vh;
  background: #0f172a;
  color: #e2e8f0;
}

.dashboard-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 1rem 2rem;
  background: #1e293b;
  border-bottom: 1px solid #334155;
}

.dashboard-header h1 {
  font-size: 1.4rem;
  color: #38bdf8;
}

.dashboard-header nav {
  display: flex;
  align-items: center;
  gap: 1.5rem;
}

.dashboard-header nav a {
  color: #94a3b8;
  text-decoration: none;
  font-weight: 500;
  transition: color 0.2s;
}

.dashboard-header nav a:hover,
.dashboard-header nav a.router-link-active {
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

.dashboard-content {
  max-width: 800px;
  margin: 2rem auto;
  padding: 0 1rem;
  display: flex;
  flex-direction: column;
  gap: 1.5rem;
}

.player-card {
  background: #1e293b;
  border-radius: 16px;
  padding: 2rem;
}

.player-info {
  display: flex;
  align-items: center;
  gap: 1rem;
  margin-bottom: 1.5rem;
}

.avatar {
  width: 60px;
  height: 60px;
  border-radius: 50%;
  background: linear-gradient(135deg, #38bdf8, #818cf8);
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 1.5rem;
  font-weight: 700;
  color: #fff;
}

.player-info h2 {
  font-size: 1.3rem;
  margin-bottom: 0.2rem;
}

.username {
  color: #64748b;
  font-size: 0.9rem;
}

.level-rank {
  display: flex;
  gap: 1rem;
  margin-bottom: 1.5rem;
}

.exp-section {
  margin-top: 0.5rem;
}

.exp-label {
  display: flex;
  justify-content: space-between;
  font-size: 0.85rem;
  color: #94a3b8;
  margin-bottom: 0.4rem;
}

.exp-bar {
  height: 10px;
  background: #334155;
  border-radius: 5px;
  overflow: hidden;
}

.exp-fill {
  height: 100%;
  border-radius: 5px;
  transition: width 0.5s ease;
}

.stats-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(160px, 1fr));
  gap: 1rem;
}

.stat-card {
  background: #1e293b;
  border-radius: 12px;
  padding: 1.5rem;
  text-align: center;
  display: flex;
  flex-direction: column;
  gap: 0.5rem;
}

.stat-value {
  font-size: 1.8rem;
  font-weight: 700;
}

.stat-label {
  color: #94a3b8;
  font-size: 0.85rem;
}

.quick-actions h3 {
  margin-bottom: 1rem;
  color: #cbd5e1;
}

.actions-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(150px, 1fr));
  gap: 1rem;
}

.action-card {
  background: #1e293b;
  border-radius: 12px;
  padding: 1.5rem;
  text-align: center;
  text-decoration: none;
  color: #e2e8f0;
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 0.5rem;
  transition: transform 0.2s, background 0.2s;
}

.action-card:hover {
  transform: translateY(-2px);
  background: #334155;
}

.action-icon {
  font-size: 2rem;
}
</style>
