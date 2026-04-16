<script setup lang="ts">
defineProps<{
  show: boolean
  title?: string
  message?: string
  confirmText?: string
  cancelText?: string
  variant?: 'danger' | 'warning' | 'info'
  loading?: boolean
}>()

defineEmits<{
  confirm: []
  cancel: []
}>()
</script>

<template>
  <Teleport to="body">
    <Transition name="modal">
      <div v-if="show" class="confirm-overlay" @click.self="$emit('cancel')">
        <div class="confirm-modal" :class="variant ?? 'info'">
          <h3 class="confirm-title">{{ title ?? 'ยืนยัน' }}</h3>
          <p class="confirm-message">{{ message ?? 'คุณต้องการดำเนินการต่อหรือไม่?' }}</p>
          <div class="confirm-actions">
            <button class="btn-modal-cancel" :disabled="loading" @click="$emit('cancel')">
              {{ cancelText ?? 'ยกเลิก' }}
            </button>
            <button
              class="btn-modal-confirm"
              :class="variant ?? 'info'"
              :disabled="loading"
              @click="$emit('confirm')"
            >
              <span v-if="loading" class="spinner"></span>
              {{ confirmText ?? 'ยืนยัน' }}
            </button>
          </div>
        </div>
      </div>
    </Transition>
  </Teleport>
</template>

<style scoped>
.confirm-overlay {
  position: fixed;
  inset: 0;
  background: rgba(0, 0, 0, 0.6);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 9999;
  backdrop-filter: blur(4px);
}

.confirm-modal {
  background: #1e293b;
  border: 1px solid #334155;
  border-radius: 16px;
  padding: 2rem;
  min-width: 360px;
  max-width: 440px;
  box-shadow: 0 20px 60px rgba(0, 0, 0, 0.5);
}

.confirm-modal.danger {
  border-color: #991b1b;
}

.confirm-modal.warning {
  border-color: #92400e;
}

.confirm-modal.info {
  border-color: #1e40af;
}

.confirm-title {
  font-size: 1.2rem;
  color: #f1f5f9;
  margin-bottom: 0.75rem;
}

.confirm-message {
  color: #94a3b8;
  font-size: 0.95rem;
  line-height: 1.5;
  margin-bottom: 1.5rem;
}

.confirm-actions {
  display: flex;
  justify-content: flex-end;
  gap: 0.75rem;
}

.btn-modal-cancel {
  background: #334155;
  color: #cbd5e1;
  border: none;
  padding: 0.6rem 1.2rem;
  border-radius: 8px;
  font-weight: 500;
  cursor: pointer;
  transition: background 0.2s;
}

.btn-modal-cancel:hover:not(:disabled) {
  background: #475569;
}

.btn-modal-confirm {
  border: none;
  padding: 0.6rem 1.4rem;
  border-radius: 8px;
  font-weight: 600;
  cursor: pointer;
  color: #fff;
  transition: opacity 0.2s;
  display: flex;
  align-items: center;
  gap: 0.5rem;
}

.btn-modal-confirm.danger {
  background: linear-gradient(135deg, #dc2626, #b91c1c);
}

.btn-modal-confirm.warning {
  background: linear-gradient(135deg, #d97706, #b45309);
}

.btn-modal-confirm.info {
  background: linear-gradient(135deg, #2563eb, #1d4ed8);
}

.btn-modal-confirm:hover:not(:disabled) {
  opacity: 0.9;
}

.btn-modal-confirm:disabled,
.btn-modal-cancel:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.spinner {
  width: 16px;
  height: 16px;
  border: 2px solid rgba(255, 255, 255, 0.3);
  border-top-color: #fff;
  border-radius: 50%;
  animation: spin 0.6s linear infinite;
}

@keyframes spin {
  to { transform: rotate(360deg); }
}

/* Transition */
.modal-enter-active,
.modal-leave-active {
  transition: opacity 0.2s ease;
}

.modal-enter-active .confirm-modal,
.modal-leave-active .confirm-modal {
  transition: transform 0.2s ease;
}

.modal-enter-from,
.modal-leave-to {
  opacity: 0;
}

.modal-enter-from .confirm-modal {
  transform: scale(0.95) translateY(-10px);
}

.modal-leave-to .confirm-modal {
  transform: scale(0.95) translateY(10px);
}
</style>
