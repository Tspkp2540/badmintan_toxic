<script setup lang="ts">
import { ref } from 'vue'
import { useAuthStore } from '@/stores/auth'
import SkillBadge from '@/components/SkillBadge.vue'
import ConfirmModal from '@/components/ConfirmModal.vue'

const showLogoutConfirm = ref(false)

const authStore = useAuthStore()

interface SkillTier {
  level: string
  label: string
  starsRequired: number
  description: string
  icon: string
}

const tiers: SkillTier[] = [
  { level: 'BG1', label: 'มือใหม่เริ่มหัด', starsRequired: 3, description: 'เพิ่งเริ่มเล่น เรียนรู้กฎกติกาและทักษะพื้นฐาน', icon: '🌱' },
  { level: 'BG2', label: 'มือหน้าบ้าน', starsRequired: 3, description: 'พื้นฐานดี ตีโต้ได้คล่อง เริ่มเข้าใจเกม', icon: '🏠' },
  { level: 'S', label: 'มือกลาง S', starsRequired: 4, description: 'ทักษะรอบด้าน เริ่มอ่านเกมเป็น', icon: '⚡' },
  { level: 'N', label: 'มือกลาง N', starsRequired: 4, description: 'เล่นได้อย่างมีแบบแผน กลยุทธ์หลากหลาย', icon: '🔥' },
  { level: 'P-', label: 'ใกล้เคียงโค้ช', starsRequired: 5, description: 'ทักษะขั้นสูง ควบคุมเกมได้ดี', icon: '💪' },
  { level: 'P', label: 'มือโค้ชทั่วไป', starsRequired: 5, description: 'ระดับโค้ช สอนและนำเกมได้', icon: '🎯' },
  { level: 'P+', label: 'ฟอร์มนักกีฬา', starsRequired: 0, description: 'สุดยอด! ระดับสูงสุด ฟอร์มนักกีฬาตัวจริง', icon: '👑' },
]

const expandedTier = ref<string | null>(null)

function toggleTier(level: string) {
  expandedTier.value = expandedTier.value === level ? null : level
}

function renderStars(filled: number, total: number): string {
  return '★'.repeat(filled) + '☆'.repeat(total - filled)
}
</script>

<template>
  <div class="skill-guide">
    <header class="guide-header">
      <h1>🏸 Badminton Hub</h1>
      <nav>
        <router-link to="/dashboard">แดชบอร์ด</router-link>
        <router-link to="/courts">สนาม</router-link>
        <router-link to="/ranking">อันดับ</router-link>
        <router-link to="/skill-guide">คู่มือระดับ</router-link>
        <router-link to="/profile">โปรไฟล์</router-link>
        <button @click="showLogoutConfirm = true" class="btn-logout">ออกจากระบบ</button>
      </nav>
    </header>

    <ConfirmModal
      :show="showLogoutConfirm"
      title="ออกจากระบบ"
      message="ต้องการออกจากระบบหรือไม่?"
      variant="danger"
      confirm-text="ออกจากระบบ"
      @confirm="authStore.logout()"
      @cancel="showLogoutConfirm = false"
    />

    <main class="guide-content">
      <section class="guide-intro">
        <h2>📖 คู่มือระดับฝีมือ</h2>
        <p>ระบบจัดอันดับฝีมือผู้เล่น แบ่งเป็น 7 ระดับ ตั้งแต่มือใหม่จนถึงระดับนักกีฬา</p>
      </section>

      <!-- How it works -->
      <section class="how-it-works">
        <h3>🎮 ระบบทำงานอย่างไร?</h3>
        <div class="steps-grid">
          <div class="step-card">
            <span class="step-number">1</span>
            <span class="step-icon">⭐</span>
            <h4>สะสมดาว</h4>
            <p>ชนะ<strong>ห้องวัดระดับ</strong>เท่านั้น = ได้ 1 ดาว<br>แพ้ = เสีย 1 ดาว</p>
          </div>
          <div class="step-card">
            <span class="step-number">2</span>
            <span class="step-icon">✅</span>
            <h4>ดาวครบ!</h4>
            <p>สะสมดาวครบตามจำนวนที่ระดับกำหนด<br>จะเข้าสู่รอบอัพระดับ</p>
          </div>
          <div class="step-card">
            <span class="step-number">3</span>
            <span class="step-icon">⚔️</span>
            <h4>ศึกอัพระดับ (Bo3)</h4>
            <p>ต้องชนะ <strong>2 ใน 3</strong> แมตช์วัดระดับ<br>เพื่ออัพไประดับถัดไป</p>
          </div>
          <div class="step-card">
            <span class="step-number">4</span>
            <span class="step-icon">🎉</span>
            <h4>อัพระดับสำเร็จ!</h4>
            <p>เลื่อนระดับใหม่ ดาวรีเซ็ตเป็น 1<br>เริ่มสะสมใหม่อีกครั้ง</p>
          </div>
        </div>
      </section>

      <!-- Match mode rewards -->
      <section class="mode-section">
        <h3>🏷️ รางวัลแต่ละประเภทห้อง</h3>
        <div class="mode-grid">
          <div class="mode-card">
            <div class="mode-header casual">🎾 ห้องลำลอง (Casual)</div>
            <ul>
              <li>✅ ได้ EXP</li>
              <li>❌ ไม่ได้แรงค์พอยท์</li>
              <li>❌ ไม่ได้ดาวระดับ</li>
            </ul>
          </div>
          <div class="mode-card">
            <div class="mode-header ranked">🏆 ห้องแรงค์ (Ranked)</div>
            <ul>
              <li>✅ ได้ EXP</li>
              <li>✅ ได้แรงค์พอยท์ (RP)</li>
              <li>❌ ไม่ได้ดาวระดับ</li>
            </ul>
          </div>
          <div class="mode-card">
            <div class="mode-header skilltest">⚔️ ห้องวัดระดับ (Skill Test)</div>
            <ul>
              <li>✅ ได้ EXP</li>
              <li>❌ ไม่ได้แรงค์พอยท์</li>
              <li>✅ <strong>ได้ดาวระดับ</strong></li>
            </ul>
          </div>
        </div>
      </section>

      <!-- Promotion fail -->
      <section class="promo-fail-note">
        <div class="note-card warning">
          <span class="note-icon">⚠️</span>
          <div>
            <strong>ถ้าไม่ผ่านศึกอัพระดับ?</strong>
            <p>ดาวจะลดลง 1 ดาว และต้องสะสมใหม่จนดาวเต็มอีกครั้ง</p>
          </div>
        </div>
      </section>

      <!-- Skill gap penalty -->
      <section class="gap-section">
        <h3>🚫 กฎระดับห่างเกินไป (2+ ขั้น)</h3>
        <div class="note-card danger">
          <span class="note-icon">🔴</span>
          <div>
            <strong>ตัวอย่าง: BG1 vs S (ห่าง 2 ขั้น)</strong>
            <p>ผู้เล่นระดับสูงกว่า → <strong>ชนะหรือแพ้ก็ไม่ได้รางวัลใดๆ</strong></p>
            <ul class="gap-list">
              <li>ชนะ: ไม่ได้ EXP, RP, ดาว</li>
              <li>แพ้ในห้องแรงค์: <strong>โดนลด RP</strong></li>
              <li>แพ้ในห้องวัดระดับ: <strong>โดนลดดาว</strong></li>
            </ul>
            <p class="gap-note">ผู้เล่นระดับต่ำกว่าจะได้รับรางวัลตามปกติ</p>
          </div>
        </div>
      </section>

      <!-- Tier list -->
      <section class="tier-list">
        <h3>🏅 ระดับฝีมือทั้งหมด</h3>

        <div class="tier-progression">
          <div
            v-for="(tier, index) in tiers"
            :key="tier.level"
            class="tier-row"
            :class="{ expanded: expandedTier === tier.level, 'is-max': tier.starsRequired === 0 }"
            @click="toggleTier(tier.level)"
          >
            <div class="tier-main">
              <div class="tier-rank">#{{ index + 1 }}</div>
              <div class="tier-icon">{{ tier.icon }}</div>
              <div class="tier-info">
                <div class="tier-header">
                  <SkillBadge :skill-level="tier.level" :skill-stars="tier.starsRequired || 5" />
                  <span class="tier-label">{{ tier.label }}</span>
                </div>
                <div v-if="tier.starsRequired > 0" class="tier-stars-needed">
                  ดาวที่ต้องสะสม: <span class="stars-display">{{ renderStars(tier.starsRequired, 5) }}</span>
                  <span class="stars-count">({{ tier.starsRequired }} ดาว)</span>
                </div>
                <div v-else class="tier-max-label">🏆 ระดับสูงสุด — ไม่มีรอบอัพ</div>
              </div>
              <div class="tier-arrow">{{ expandedTier === tier.level ? '▲' : '▼' }}</div>
            </div>

            <transition name="expand">
              <div v-if="expandedTier === tier.level" class="tier-detail">
                <p class="tier-desc">{{ tier.description }}</p>

                <div v-if="tier.starsRequired > 0" class="promo-flow">
                  <h5>ขั้นตอนอัพจาก {{ tier.level }} → {{ tiers[index + 1]?.level ?? '?' }}</h5>
                  <div class="flow-steps">
                    <div class="flow-step">
                      <div class="flow-circle">🎮</div>
                      <span>เล่นห้องวัดระดับ</span>
                    </div>
                    <div class="flow-arrow">→</div>
                    <div class="flow-step">
                      <div class="flow-circle">⭐</div>
                      <span>สะสม {{ tier.starsRequired }} ดาว</span>
                    </div>
                    <div class="flow-arrow">→</div>
                    <div class="flow-step">
                      <div class="flow-circle">⚔️</div>
                      <span>ศึกอัพ Bo3</span>
                    </div>
                    <div class="flow-arrow">→</div>
                    <div class="flow-step">
                      <div class="flow-circle success">✅</div>
                      <span>ชนะ 2/3 = อัพ!</span>
                    </div>
                  </div>

                  <div class="estimate-box">
                    <span class="estimate-label">จำนวนแมตช์โดยประมาณ:</span>
                    <span class="estimate-value">~{{ tier.starsRequired + 2 }}-{{ tier.starsRequired * 2 + 3 }} แมตช์</span>
                    <span class="estimate-note">(ขึ้นอยู่กับอัตราชนะ)</span>
                  </div>
                </div>

                <div v-else class="max-tier-info">
                  <p>🎊 ยินดีด้วย! คุณอยู่ที่ระดับสูงสุดแล้ว</p>
                  <p>เพิ่มดาวได้ถึง 5 ดาว เพื่อแสดงความแกร่ง!</p>
                </div>
              </div>
            </transition>

            <!-- Arrow between tiers -->
            <div v-if="index < tiers.length - 1 && tier.starsRequired > 0" class="tier-connector">
              <div class="connector-line"></div>
              <div class="connector-label">ชนะ 2/3</div>
              <div class="connector-line"></div>
            </div>
          </div>
        </div>
      </section>

      <!-- Summary table -->
      <section class="summary-section">
        <h3>📊 สรุปตารางระดับ</h3>
        <div class="table-wrapper">
          <table class="summary-table">
            <thead>
              <tr>
                <th>ระดับ</th>
                <th>ชื่อ</th>
                <th>ดาวสะสม</th>
                <th>รอบอัพ</th>
                <th>แมตช์โดยประมาณ</th>
              </tr>
            </thead>
            <tbody>
              <tr v-for="tier in tiers" :key="'table-' + tier.level" :class="{ 'max-row': tier.starsRequired === 0 }">
                <td>
                  <SkillBadge :skill-level="tier.level" :skill-stars="tier.starsRequired || 5" />
                </td>
                <td>{{ tier.label }}</td>
                <td>
                  <template v-if="tier.starsRequired > 0">
                    <span class="table-stars">{{ renderStars(tier.starsRequired, 5) }}</span>
                    ({{ tier.starsRequired }})
                  </template>
                  <template v-else>
                    <span class="max-text">MAX</span>
                  </template>
                </td>
                <td>
                  <template v-if="tier.starsRequired > 0">
                    Bo3 (ชนะ 2/3)
                  </template>
                  <template v-else>—</template>
                </td>
                <td>
                  <template v-if="tier.starsRequired > 0">
                    ~{{ tier.starsRequired + 2 }}-{{ tier.starsRequired * 2 + 3 }}
                  </template>
                  <template v-else>—</template>
                </td>
              </tr>
            </tbody>
          </table>
        </div>
      </section>

      <!-- EXP Scaling note -->
      <section class="scaling-section">
        <h3>⚖️ ระบบ EXP ตามระดับฝีมือ</h3>
        <div class="scaling-cards">
          <div class="scale-card bonus">
            <span class="scale-icon">📈</span>
            <div>
              <strong>ชนะคนเก่งกว่า</strong>
              <p>ได้ EXP โบนัสสูงสุด <strong>2.0x</strong></p>
            </div>
          </div>
          <div class="scale-card normal">
            <span class="scale-icon">➡️</span>
            <div>
              <strong>ชนะระดับเท่ากัน</strong>
              <p>ได้ EXP ปกติ <strong>1.0x</strong></p>
            </div>
          </div>
          <div class="scale-card reduced">
            <span class="scale-icon">📉</span>
            <div>
              <strong>ชนะคนอ่อนกว่า</strong>
              <p>ได้ EXP ลดลงเหลือ <strong>0.3x</strong></p>
            </div>
          </div>
          <div class="scale-card penalty">
            <span class="scale-icon">💀</span>
            <div>
              <strong>แพ้ทั้งที่เก่งกว่า</strong>
              <p>เสีย RP <strong>2x</strong> เพิ่ม</p>
            </div>
          </div>
        </div>
      </section>
    </main>
  </div>
</template>

<style scoped>
.skill-guide {
  min-height: 100vh;
  background: #0f172a;
  color: #e2e8f0;
}

.guide-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 1rem 2rem;
  background: #1e293b;
  border-bottom: 1px solid #334155;
}

.guide-header h1 {
  font-size: 1.4rem;
  color: #38bdf8;
}

.guide-header nav {
  display: flex;
  align-items: center;
  gap: 1.5rem;
}

.guide-header nav a {
  color: #94a3b8;
  text-decoration: none;
  font-weight: 500;
  transition: color 0.2s;
}

.guide-header nav a:hover,
.guide-header nav a.router-link-active {
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

.guide-content {
  max-width: 900px;
  margin: 2rem auto;
  padding: 0 1rem;
  display: flex;
  flex-direction: column;
  gap: 2rem;
}

/* Intro */
.guide-intro {
  text-align: center;
}

.guide-intro h2 {
  font-size: 1.8rem;
  margin-bottom: 0.5rem;
  color: #f1f5f9;
}

.guide-intro p {
  color: #94a3b8;
  font-size: 1rem;
}

/* How it works */
.how-it-works h3 {
  margin-bottom: 1rem;
  color: #cbd5e1;
}

.steps-grid {
  display: grid;
  grid-template-columns: repeat(4, 1fr);
  gap: 1rem;
}

.step-card {
  background: #1e293b;
  border-radius: 12px;
  padding: 1.5rem 1rem;
  text-align: center;
  position: relative;
}

.step-number {
  position: absolute;
  top: 0.5rem;
  left: 0.75rem;
  font-size: 0.7rem;
  font-weight: 700;
  color: #475569;
  background: #0f172a;
  width: 20px;
  height: 20px;
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
}

.step-icon {
  font-size: 2rem;
  display: block;
  margin-bottom: 0.5rem;
}

.step-card h4 {
  font-size: 0.95rem;
  margin-bottom: 0.3rem;
  color: #f1f5f9;
}

.step-card p {
  font-size: 0.8rem;
  color: #94a3b8;
  line-height: 1.5;
}

/* Promo fail note */
.note-card {
  display: flex;
  align-items: flex-start;
  gap: 0.75rem;
  padding: 1rem 1.5rem;
  border-radius: 12px;
}

.note-card.warning {
  background: #422006;
  border: 1px solid #92400e;
}

.note-card.danger {
  background: #2a0a0a;
  border: 1px solid #7f1d1d;
}

.note-card.danger strong {
  color: #fca5a5;
}

.note-card.danger p {
  color: #d4a0a0;
}

.gap-list {
  margin: 0.5rem 0 0.5rem 1.2rem;
  padding: 0;
  font-size: 0.85rem;
  line-height: 1.8;
}

.gap-note {
  color: #4ade80 !important;
  font-size: 0.85rem;
  margin-top: 0.3rem;
}

/* Match mode section */
.mode-section h3 {
  margin-bottom: 1rem;
  color: #cbd5e1;
}

.mode-grid {
  display: grid;
  grid-template-columns: repeat(3, 1fr);
  gap: 1rem;
}

.mode-card {
  background: #1e293b;
  border-radius: 12px;
  overflow: hidden;
}

.mode-header {
  padding: 0.75rem 1rem;
  font-weight: 700;
  font-size: 0.95rem;
  text-align: center;
}

.mode-header.casual {
  background: #1e3a5f;
  color: #93c5fd;
}

.mode-header.ranked {
  background: #3b2a15;
  color: #fbbf24;
}

.mode-header.skilltest {
  background: #1a3320;
  color: #4ade80;
}

.mode-card ul {
  list-style: none;
  padding: 1rem;
  margin: 0;
  display: flex;
  flex-direction: column;
  gap: 0.4rem;
  font-size: 0.85rem;
}

/* Gap section */
.gap-section h3 {
  margin-bottom: 1rem;
  color: #cbd5e1;
}

.note-icon {
  font-size: 1.5rem;
  flex-shrink: 0;
}

.note-card strong {
  color: #fbbf24;
}

.note-card p {
  color: #d4a276;
  font-size: 0.9rem;
  margin-top: 0.2rem;
}

/* Tier list */
.tier-list h3 {
  margin-bottom: 1rem;
  color: #cbd5e1;
}

.tier-progression {
  display: flex;
  flex-direction: column;
  gap: 0;
}

.tier-row {
  cursor: pointer;
}

.tier-main {
  display: flex;
  align-items: center;
  gap: 1rem;
  background: #1e293b;
  border-radius: 12px;
  padding: 1.25rem 1.5rem;
  transition: background 0.2s;
}

.tier-main:hover {
  background: #273548;
}

.tier-row.is-max .tier-main {
  border: 1px solid #fbbf24;
  background: linear-gradient(135deg, #1e293b 0%, #3b2a15 100%);
}

.tier-rank {
  color: #475569;
  font-size: 0.8rem;
  font-weight: 700;
  min-width: 24px;
}

.tier-icon {
  font-size: 1.5rem;
  min-width: 32px;
  text-align: center;
}

.tier-info {
  flex: 1;
}

.tier-header {
  display: flex;
  align-items: center;
  gap: 0.75rem;
  margin-bottom: 0.3rem;
}

.tier-label {
  color: #cbd5e1;
  font-weight: 600;
}

.tier-stars-needed {
  font-size: 0.85rem;
  color: #94a3b8;
}

.stars-display {
  color: #fbbf24;
  letter-spacing: 2px;
  margin: 0 0.3rem;
}

.stars-count {
  color: #64748b;
  font-size: 0.8rem;
}

.tier-max-label {
  font-size: 0.85rem;
  color: #fbbf24;
  font-weight: 600;
}

.tier-arrow {
  color: #475569;
  font-size: 0.7rem;
}

/* Expanded detail */
.tier-detail {
  background: #162032;
  border-radius: 0 0 12px 12px;
  padding: 1.25rem 1.5rem;
  margin-top: -8px;
  border-top: 1px solid #334155;
}

.tier-desc {
  color: #94a3b8;
  font-size: 0.9rem;
  margin-bottom: 1rem;
}

.promo-flow h5 {
  color: #38bdf8;
  font-size: 0.9rem;
  margin-bottom: 0.75rem;
}

.flow-steps {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  flex-wrap: wrap;
  margin-bottom: 1rem;
}

.flow-step {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 0.3rem;
}

.flow-circle {
  width: 40px;
  height: 40px;
  border-radius: 50%;
  background: #334155;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 1.2rem;
}

.flow-circle.success {
  background: #065f46;
}

.flow-step span:last-child {
  font-size: 0.75rem;
  color: #94a3b8;
  text-align: center;
  max-width: 80px;
}

.flow-arrow {
  color: #475569;
  font-size: 1.2rem;
  font-weight: 700;
  margin-top: -16px;
}

.estimate-box {
  background: #1e293b;
  border-radius: 8px;
  padding: 0.75rem 1rem;
  display: flex;
  align-items: center;
  gap: 0.5rem;
  flex-wrap: wrap;
}

.estimate-label {
  color: #94a3b8;
  font-size: 0.85rem;
}

.estimate-value {
  color: #38bdf8;
  font-weight: 700;
  font-size: 0.95rem;
}

.estimate-note {
  color: #64748b;
  font-size: 0.8rem;
}

.max-tier-info {
  text-align: center;
  padding: 1rem;
}

.max-tier-info p {
  color: #fbbf24;
  font-size: 0.9rem;
  line-height: 1.6;
}

/* Connector between tiers */
.tier-connector {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  padding: 0.3rem 0 0.3rem 3.5rem;
}

.connector-line {
  flex: 1;
  height: 1px;
  background: #334155;
}

.connector-label {
  color: #64748b;
  font-size: 0.7rem;
  font-weight: 600;
  white-space: nowrap;
}

/* Summary table */
.summary-section h3 {
  margin-bottom: 1rem;
  color: #cbd5e1;
}

.table-wrapper {
  overflow-x: auto;
}

.summary-table {
  width: 100%;
  border-collapse: separate;
  border-spacing: 0;
  background: #1e293b;
  border-radius: 12px;
  overflow: hidden;
}

.summary-table th {
  background: #334155;
  padding: 0.75rem 1rem;
  text-align: left;
  font-size: 0.85rem;
  color: #94a3b8;
  font-weight: 600;
}

.summary-table td {
  padding: 0.75rem 1rem;
  font-size: 0.9rem;
  border-top: 1px solid #2d3a4e;
}

.summary-table tr.max-row td {
  background: #1a1a0e;
  color: #fbbf24;
}

.table-stars {
  color: #fbbf24;
  letter-spacing: 2px;
  margin-right: 0.3rem;
}

.max-text {
  color: #fbbf24;
  font-weight: 700;
}

/* Scaling section */
.scaling-section h3 {
  margin-bottom: 1rem;
  color: #cbd5e1;
}

.scaling-cards {
  display: grid;
  grid-template-columns: repeat(2, 1fr);
  gap: 1rem;
}

.scale-card {
  display: flex;
  align-items: flex-start;
  gap: 0.75rem;
  padding: 1.25rem;
  border-radius: 12px;
}

.scale-card.bonus {
  background: #052e16;
  border: 1px solid #166534;
}

.scale-card.normal {
  background: #1e293b;
  border: 1px solid #334155;
}

.scale-card.reduced {
  background: #1c1917;
  border: 1px solid #44403c;
}

.scale-card.penalty {
  background: #450a0a;
  border: 1px solid #7f1d1d;
}

.scale-icon {
  font-size: 1.5rem;
  flex-shrink: 0;
}

.scale-card strong {
  display: block;
  margin-bottom: 0.2rem;
  font-size: 0.9rem;
}

.scale-card p {
  color: #94a3b8;
  font-size: 0.85rem;
}

.scale-card.bonus strong { color: #4ade80; }
.scale-card.normal strong { color: #e2e8f0; }
.scale-card.reduced strong { color: #a8a29e; }
.scale-card.penalty strong { color: #fca5a5; }

/* Expand transition */
.expand-enter-active,
.expand-leave-active {
  transition: all 0.25s ease;
  overflow: hidden;
}

.expand-enter-from,
.expand-leave-to {
  opacity: 0;
  max-height: 0;
  padding-top: 0;
  padding-bottom: 0;
}

/* Responsive */
@media (max-width: 768px) {
  .guide-header {
    flex-direction: column;
    gap: 0.75rem;
    padding: 1rem;
  }

  .guide-header h1 {
    font-size: 1.2rem;
  }

  .guide-header nav {
    flex-wrap: wrap;
    justify-content: center;
    gap: 0.75rem;
  }

  .guide-content {
    margin: 1rem auto;
    padding: 0 0.75rem;
  }

  .steps-grid {
    grid-template-columns: repeat(2, 1fr);
  }

  .scaling-cards {
    grid-template-columns: 1fr;
  }

  .mode-grid {
    grid-template-columns: 1fr;
  }

  .flow-steps {
    justify-content: center;
  }

  .tier-connector {
    padding-left: 2rem;
  }
}

@media (max-width: 480px) {
  .steps-grid {
    grid-template-columns: 1fr;
  }

  .tier-main {
    padding: 1rem;
    gap: 0.5rem;
  }
}
</style>
