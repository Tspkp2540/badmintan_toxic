import type { LeaderboardResponse, RankingEntry } from '@/models/Ranking'
import apiClient from './api'

export const rankingService = {
  async getLeaderboard(page = 1, limit = 20): Promise<LeaderboardResponse> {
    const response = await apiClient.get<LeaderboardResponse>('/rankings', {
      params: { page, limit },
    })
    return response.data
  },

  async getPlayerRank(userId: string): Promise<RankingEntry> {
    const response = await apiClient.get<RankingEntry>(`/rankings/${userId}`)
    return response.data
  },
}
