<script setup lang="ts">
import { computed } from 'vue'
import { useAuthStore } from '@/stores/auth'
import { getRankByLevel, getExpForLevel, RANK_THRESHOLDS } from '@/models/User'
import LevelBadge from '@/components/LevelBadge.vue'
import RankBadge from '@/components/RankBadge.vue'

const authStore = useAuthStore()
const user = computed(() => authStore.user)
const level = computed(() => user.value?.level ?? 1)
const rank = computed(() => user.value?.rank ?? getRankByLevel(1))
const exp = computed(() => user.value?.exp ?? 0)
const expToNext = computed(() => user.value?.expToNextLevel ?? getExpForLevel(1))

const allRanks = Object.entries(RANK_THRESHOLDS).map(([name, minLevel]) => ({
  name,
  minLevel,
  achieved: level.value >= minLevel,
}))
</script>

<template>
  <div class="profile-page">
    <header class="page-header">
      <router-link to="/dashboard" class="back-link">← กลับ</router-link>
      <h1>👤 โปรไฟล์</h1>
    </header>

    <div class="profile-content">
      <!-- Profile Card -->
      <section class="profile-card">
        <div class="avatar-large">
          {{ user?.fullName?.charAt(0)?.toUpperCase() ?? '?' }}
        </div>
        <h2>{{ user?.fullName ?? 'นักแบดมินตัน' }}</h2>
        <p class="username">@{{ user?.username ?? 'player' }}</p>
        <div class="badges">
          <LevelBadge :level="level" />
          <RankBadge :rank="rank" />
        </div>
      </section>

      <!-- EXP Progress -->
      <section class="card">
        <h3>⚡ ความก้าวหน้า</h3>
        <div class="exp-info">
          <span>Level {{ level }}</span>
          <span>{{ exp }} / {{ expToNext }} EXP</span>
        </div>
        <div class="exp-bar">
          <div
            class="exp-fill"
            :style="{ width: `${Math.min((exp / expToNext) * 100, 100)}%` }"
          ></div>
        </div>
      </section>

      <!-- Stats -->
      <section class="card">
        <h3>📊 สถิติการเล่น</h3>
        <div class="stats-list">
          <div class="stat-row">
            <span>ชนะ</span>
            <span class="win">{{ user?.wins ?? 0 }}</span>
          </div>
          <div class="stat-row">
            <span>แพ้</span>
            <span class="loss">{{ user?.losses ?? 0 }}</span>
          </div>
          <div class="stat-row">
            <span>แมตช์ทั้งหมด</span>
            <span>{{ user?.totalMatches ?? 0 }}</span>
          </div>
          <div class="stat-row">
            <span>อัตราชนะ</span>
            <span class="highlight">{{ (user?.winRate ?? 0).toFixed(1) }}%</span>
          </div>
        </div>
      </section>

      <!-- Rank Progression -->
      <section class="card">
        <h3>🎖️ ระดับแร้งค์</h3>
        <div class="rank-progression">
          <div
            v-for="r in allRanks"
            :key="r.name"
            class="rank-item"
            :class="{ achieved: r.achieved, current: r.name === rank }"
          >
            <RankBadge :rank="r.name" />
            <span class="rank-level">Lv.{{ r.minLevel }}+</span>
          </div>
        </div>
      </section>
    </div>
  </div>
</template>

<style scoped>
.profile-page {
  min-height: 100vh;
  background: #0f172a;
  color: #e2e8f0;
  padding: 2rem;
}

.page-header {
  max-width: 600px;
  margin: 0 auto 2rem;
}

.back-link {
  color: #64748b;
  text-decoration: none;
  font-size: 0.9rem;
  display: inline-block;
  margin-bottom: 0.5rem;
}

.back-link:hover {
  color: #38bdf8;
}

.page-header h1 {
  font-size: 1.6rem;
}

.profile-content {
  max-width: 600px;
  margin: 0 auto;
  display: flex;
  flex-direction: column;
  gap: 1.5rem;
}

.profile-card {
  background: #1e293b;
  border-radius: 16px;
  padding: 2rem;
  text-align: center;
}

.avatar-large {
  width: 80px;
  height: 80px;
  border-radius: 50%;
  background: linear-gradient(135deg, #38bdf8, #818cf8);
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 2rem;
  font-weight: 700;
  color: #fff;
  margin: 0 auto 1rem;
}

.profile-card h2 {
  font-size: 1.4rem;
  margin-bottom: 0.3rem;
}

.username {
  color: #64748b;
  margin-bottom: 1rem;
}

.badges {
  display: flex;
  justify-content: center;
  gap: 0.75rem;
}

.card {
  background: #1e293b;
  border-radius: 12px;
  padding: 1.5rem;
}

.card h3 {
  margin-bottom: 1rem;
  font-size: 1.1rem;
  color: #cbd5e1;
}

.exp-info {
  display: flex;
  justify-content: space-between;
  font-size: 0.85rem;
  color: #94a3b8;
  margin-bottom: 0.5rem;
}

.exp-bar {
  height: 12px;
  background: #334155;
  border-radius: 6px;
  overflow: hidden;
}

.exp-fill {
  height: 100%;
  background: linear-gradient(90deg, #38bdf8, #818cf8);
  border-radius: 6px;
  transition: width 0.5s ease;
}

.stats-list {
  display: flex;
  flex-direction: column;
  gap: 0.75rem;
}

.stat-row {
  display: flex;
  justify-content: space-between;
  padding: 0.5rem 0;
  border-bottom: 1px solid #334155;
  font-size: 0.95rem;
}

.stat-row:last-child {
  border-bottom: none;
}

.win {
  color: #4ade80;
  font-weight: 600;
}

.loss {
  color: #f87171;
  font-weight: 600;
}

.highlight {
  color: #fbbf24;
  font-weight: 600;
}

.rank-progression {
  display: flex;
  flex-wrap: wrap;
  gap: 1rem;
  justify-content: center;
}

.rank-item {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 0.3rem;
  opacity: 0.35;
  transition: opacity 0.2s;
}

.rank-item.achieved {
  opacity: 1;
}

.rank-item.current {
  transform: scale(1.15);
}

.rank-level {
  font-size: 0.75rem;
  color: #64748b;
}
</style>
