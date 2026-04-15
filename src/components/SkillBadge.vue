<script setup lang="ts">
import { computed } from 'vue'
import { SkillLevelLabels } from '@/models/User'

const props = defineProps<{
  skillLevel: string
  skillStars: number
}>()

const skillColorMap: Record<string, { bg: string; text: string }> = {
  BG1: { bg: '#374151', text: '#9ca3af' },
  BG2: { bg: '#3f6212', text: '#a3e635' },
  S: { bg: '#1e40af', text: '#93c5fd' },
  N: { bg: '#7e22ce', text: '#d8b4fe' },
  'P-': { bg: '#b45309', text: '#fcd34d' },
  P: { bg: '#dc2626', text: '#fca5a5' },
  'P+': { bg: '#7c2d12', text: '#fdba74' },
}

const colors = computed(() => skillColorMap[props.skillLevel] ?? { bg: '#334155', text: '#94a3b8' })
const label = computed(() => SkillLevelLabels[props.skillLevel] ?? props.skillLevel)
const stars = computed(() => '★'.repeat(props.skillStars) + '☆'.repeat(5 - props.skillStars))
</script>

<template>
  <span
    class="skill-badge"
    :style="{ background: colors.bg, color: colors.text }"
    :title="label"
  >
    {{ skillLevel }} <span class="stars">{{ stars }}</span>
  </span>
</template>

<style scoped>
.skill-badge {
  display: inline-flex;
  align-items: center;
  gap: 0.3rem;
  padding: 0.3rem 0.75rem;
  border-radius: 20px;
  font-size: 0.85rem;
  font-weight: 700;
  letter-spacing: 0.5px;
}
.stars {
  font-size: 0.75rem;
  letter-spacing: 1px;
}
</style>
