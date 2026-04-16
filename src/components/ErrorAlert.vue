<script setup lang="ts">
defineProps<{
  message: string | null
}>()

defineEmits<{
  close: []
}>()
</script>

<template>
  <Transition name="alert">
    <div v-if="message" class="error-alert">
      <span class="error-icon">⚠️</span>
      <span class="error-msg">{{ message }}</span>
      <button class="error-close" @click="$emit('close')">✕</button>
    </div>
  </Transition>
</template>

<style scoped>
.error-alert {
  display: flex;
  align-items: center;
  gap: 0.75rem;
  background: rgba(239, 68, 68, 0.15);
  border: 1px solid #991b1b;
  border-radius: 10px;
  padding: 0.75rem 1rem;
  margin-bottom: 1rem;
  color: #fca5a5;
  font-size: 0.9rem;
}

.error-icon {
  font-size: 1.1rem;
  flex-shrink: 0;
}

.error-msg {
  flex: 1;
}

.error-close {
  background: none;
  border: none;
  color: #f87171;
  cursor: pointer;
  font-size: 1rem;
  padding: 0.2rem;
  opacity: 0.7;
  transition: opacity 0.2s;
}

.error-close:hover {
  opacity: 1;
}

.alert-enter-active,
.alert-leave-active {
  transition: all 0.3s ease;
}

.alert-enter-from,
.alert-leave-to {
  opacity: 0;
  transform: translateY(-8px);
}
</style>
