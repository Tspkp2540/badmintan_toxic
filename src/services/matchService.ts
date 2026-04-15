import apiClient from './api'

export const matchService = {
  async getMatches(status?: string) {
    const params = status ? { status } : {}
    const response = await apiClient.get('/matches', { params })
    return response.data
  },

  async getMatch(id: string) {
    const response = await apiClient.get(`/matches/${id}`)
    return response.data
  },

  async createMatch(data: { courtId: string; name: string; matchType: string; matchMode: string; maxSets: number }) {
    const response = await apiClient.post('/matches', data)
    return response.data
  },

  async joinMatch(id: string, team: string) {
    const response = await apiClient.post(`/matches/${id}/join`, { team })
    return response.data
  },

  async joinAsReferee(id: string) {
    const response = await apiClient.post(`/matches/${id}/referee`)
    return response.data
  },

  async startMatch(id: string) {
    const response = await apiClient.post(`/matches/${id}/start`)
    return response.data
  },

  async endMatch(id: string) {
    const response = await apiClient.post(`/matches/${id}/end`)
    return response.data
  },

  async submitScores(id: string, sets: { setNumber: number; teamA: number; teamB: number }[]) {
    const response = await apiClient.post(`/matches/${id}/score`, { sets })
    return response.data
  },

  async leaveMatch(id: string) {
    const response = await apiClient.post(`/matches/${id}/leave`)
    return response.data
  },

  async updateSkillLevel(userId: string, skillLevel: string, skillStars: number) {
    const response = await apiClient.put('/matches/skill-level', { userId, skillLevel, skillStars })
    return response.data
  },
}
