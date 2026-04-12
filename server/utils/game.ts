export function getExpForLevel(level: number): number {
  return Math.floor(100 * Math.pow(1.2, level - 1))
}

export function getRankByLevel(level: number): string {
  if (level >= 75) return 'Grandmaster'
  if (level >= 50) return 'Master'
  if (level >= 35) return 'Diamond'
  if (level >= 20) return 'Platinum'
  if (level >= 10) return 'Gold'
  if (level >= 5) return 'Silver'
  return 'Bronze'
}

export const EXP_REWARDS = {
  casual: { win: 30, lose: 10, draw: 15 },
  ranked: { win: 50, lose: 15, draw: 25 },
} as const

export const RANK_POINT_REWARDS = {
  win: 25,
  lose: -10,
  draw: 5,
} as const
