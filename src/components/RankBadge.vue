<script setup lang="ts">
import { computed } from 'vue'
import { PlayerRank } from '@/models/User'

const props = defineProps<{
  rank: string
}>()

const rankColorMap: Record<string, { bg: string; text: string }> = {
  [PlayerRank.BRONZE]: { bg: '#92400e', text: '#fbbf24' },
  [PlayerRank.SILVER]: { bg: '#374151', text: '#d1d5db' },
  [PlayerRank.GOLD]: { bg: '#78350f', text: '#fde047' },
  [PlayerRank.PLATINUM]: { bg: '#164e63', text: '#22d3ee' },
  [PlayerRank.DIAMOND]: { bg: '#312e81', text: '#a5b4fc' },
  [PlayerRank.MASTER]: { bg: '#831843', text: '#f9a8d4' },
  [PlayerRank.GRANDMASTER]: { bg: '#7c2d12', text: '#fdba74' },
}

const colors = computed(() => rankColorMap[props.rank] ?? { bg: '#334155', text: '#94a3b8' })

const rankEmojis: Record<string, string> = {
  [PlayerRank.BRONZE]: '🥉',
  [PlayerRank.SILVER]: '🥈',
  [PlayerRank.GOLD]: '🥇',
  [PlayerRank.PLATINUM]: '💎',
  [PlayerRank.DIAMOND]: '💠',
  [PlayerRank.MASTER]: '🔥',
  [PlayerRank.GRANDMASTER]: '👑',
}

const emoji = computed(() => rankEmojis[props.rank] ?? '🏅')
</script>

<template>
  <span
    class="rank-badge"
    :style="{ background: colors.bg, color: colors.text }"
  >
    {{ emoji }} {{ rank }}
  </span>
</template>

<style scoped>
.rank-badge {
  display: inline-flex;
  align-items: center;
  gap: 0.3rem;
  padding: 0.3rem 0.75rem;
  border-radius: 20px;
  font-size: 0.85rem;
  font-weight: 700;
  letter-spacing: 0.5px;
}
</style>
