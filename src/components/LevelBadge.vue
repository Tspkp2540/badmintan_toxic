<script setup lang="ts">
defineProps<{
  level: number
}>()
</script>

<template>
  <span class="level-badge">Lv.{{ level }}</span>
</template>

<style scoped>
.level-badge {
  display: inline-flex;
  align-items: center;
  padding: 0.3rem 0.75rem;
  border-radius: 20px;
  background: linear-gradient(135deg, #38bdf8, #818cf8);
  color: #fff;
  font-size: 0.85rem;
  font-weight: 700;
  letter-spacing: 0.5px;
}
</style>
