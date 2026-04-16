<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { rankingService } from '@/services/rankingService'
import type { RankingEntry } from '@/models/Ranking'
import RankBadge from '@/components/RankBadge.vue'
import SkillBadge from '@/components/SkillBadge.vue'
import ErrorAlert from '@/components/ErrorAlert.vue'

const rankings = ref<RankingEntry[]>([])
const loading = ref(true)
const errorMsg = ref<string | null>(null)
const currentPage = ref(1)
const totalPages = ref(1)

async function loadRankings(page = 1) {
  loading.value = true
  errorMsg.value = null
  try {
    const data = await rankingService.getLeaderboard(page)
    rankings.value = data.rankings
    currentPage.value = data.currentPage
    totalPages.value = data.totalPages
  } catch (e: any) {
    errorMsg.value = e.response?.data?.message || 'โหลดอันดับไม่สำเร็จ'
  } finally {
    loading.value = false
  }
}

onMounted(() => loadRankings())

function getMedalEmoji(position: number): string {
  if (position === 1) return '🥇'
  if (position === 2) return '🥈'
  if (position === 3) return '🥉'
  return `#${position}`
}
</script>

<template>
  <div class="ranking-page">
    <header class="page-header">
      <router-link to="/dashboard" class="back-link">← กลับ</router-link>
      <h1>🏆 อันดับนักแบดมินตัน</h1>
    </header>

    <div v-if="loading" class="loading">กำลังโหลด...</div>

    <template v-else>
      <ErrorAlert :message="errorMsg" @close="errorMsg = null" />

      <div class="ranking-table-wrapper">
      <table class="ranking-table">
        <thead>
          <tr>
            <th>อันดับ</th>
            <th>ผู้เล่น</th>
            <th>เลเวล</th>
            <th>แร้งค์</th>
            <th>ระดับฝีมือ</th>
            <th>ชนะ</th>
            <th>แพ้</th>
            <th>อัตราชนะ</th>
            <th>คะแนน</th>
          </tr>
        </thead>
        <tbody>
          <tr
            v-for="entry in rankings"
            :key="entry.userId"
            :class="{ 'top-3': entry.position <= 3 }"
          >
            <td class="position">{{ getMedalEmoji(entry.position) }}</td>
            <td class="player-name">
              <span class="avatar-sm">{{ entry.fullName.charAt(0) }}</span>
              {{ entry.fullName }}
            </td>
            <td>Lv.{{ entry.level }}</td>
            <td><RankBadge :rank="entry.rank" /></td>
            <td><SkillBadge :skill-level="entry.skillLevel" :skill-stars="entry.skillStars" /></td>
            <td class="win">{{ entry.wins }}</td>
            <td class="loss">{{ entry.losses }}</td>
            <td>{{ entry.winRate.toFixed(1) }}%</td>
            <td class="points">{{ entry.points.toLocaleString() }}</td>
          </tr>
        </tbody>
      </table>

      <!-- Pagination -->
      <div v-if="totalPages > 1" class="pagination">
        <button
          :disabled="currentPage <= 1"
          @click="loadRankings(currentPage - 1)"
        >
          ← ก่อนหน้า
        </button>
        <span>หน้า {{ currentPage }} / {{ totalPages }}</span>
        <button
          :disabled="currentPage >= totalPages"
          @click="loadRankings(currentPage + 1)"
        >
          ถัดไป →
        </button>
      </div>
      </div>
    </template>
  </div>
</template>

<style scoped>
.ranking-page {
  min-height: 100vh;
  background: #0f172a;
  color: #e2e8f0;
  padding: 2rem;
}

.page-header {
  max-width: 900px;
  margin: 0 auto 2rem;
}

.back-link {
  color: #64748b;
  text-decoration: none;
  font-size: 0.9rem;
  margin-bottom: 0.5rem;
  display: inline-block;
}

.back-link:hover {
  color: #38bdf8;
}

.page-header h1 {
  font-size: 1.6rem;
  color: #fbbf24;
}

.loading {
  text-align: center;
  color: #94a3b8;
  padding: 3rem;
}

.ranking-table-wrapper {
  max-width: 900px;
  margin: 0 auto;
}

.ranking-table {
  width: 100%;
  border-collapse: collapse;
  background: #1e293b;
  border-radius: 12px;
  overflow: hidden;
}

.ranking-table thead {
  background: #334155;
}

.ranking-table th {
  padding: 0.75rem 1rem;
  text-align: left;
  font-size: 0.85rem;
  color: #94a3b8;
  font-weight: 600;
  text-transform: uppercase;
}

.ranking-table td {
  padding: 0.75rem 1rem;
  border-bottom: 1px solid #1e293b;
}

.ranking-table tr.top-3 {
  background: rgba(251, 191, 36, 0.05);
}

.position {
  font-size: 1.1rem;
  font-weight: 700;
}

.player-name {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  font-weight: 500;
}

.avatar-sm {
  width: 30px;
  height: 30px;
  border-radius: 50%;
  background: linear-gradient(135deg, #38bdf8, #818cf8);
  display: inline-flex;
  align-items: center;
  justify-content: center;
  font-size: 0.8rem;
  font-weight: 700;
  color: #fff;
}

.win {
  color: #4ade80;
}

.loss {
  color: #f87171;
}

.points {
  color: #fbbf24;
  font-weight: 700;
}

.pagination {
  display: flex;
  justify-content: center;
  align-items: center;
  gap: 1rem;
  margin-top: 1.5rem;
}

.pagination button {
  background: #334155;
  color: #e2e8f0;
  border: none;
  padding: 0.5rem 1rem;
  border-radius: 6px;
  cursor: pointer;
  transition: background 0.2s;
}

.pagination button:hover:not(:disabled) {
  background: #475569;
}

.pagination button:disabled {
  opacity: 0.4;
  cursor: not-allowed;
}

.pagination span {
  color: #94a3b8;
  font-size: 0.9rem;
}
</style>
