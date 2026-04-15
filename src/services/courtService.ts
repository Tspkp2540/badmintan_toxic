import apiClient from './api'

export interface Court {
  id: string
  name: string
  description: string
  location: string
  maxRooms: number
  bonusExpPercent: number
  status: 'open' | 'closed' | 'maintenance'
  createdBy: string
  creatorName: string
  activeRooms: number
  totalPlayers: number
  leaders: CourtLeader[]
  createdAt: string
  updatedAt: string
}

export interface CourtLeader {
  userId: string
  username: string
  fullName: string
}

export interface CreateCourtRequest {
  name: string
  description?: string
  location?: string
  maxRooms?: number
}

export interface UpdateCourtRequest {
  name?: string
  description?: string
  location?: string
  maxRooms?: number
  status?: string
}

export const courtService = {
  async getCourts(status?: string): Promise<Court[]> {
    const params = status ? { status } : {}
    const response = await apiClient.get<Court[]>('/courts', { params })
    return response.data
  },

  async getCourt(id: string): Promise<Court> {
    const response = await apiClient.get<Court>(`/courts/${id}`)
    return response.data
  },

  async createCourt(data: CreateCourtRequest): Promise<Court> {
    const response = await apiClient.post<Court>('/courts', data)
    return response.data
  },

  async updateCourt(id: string, data: UpdateCourtRequest): Promise<Court> {
    const response = await apiClient.put<Court>(`/courts/${id}`, data)
    return response.data
  },

  async deleteCourt(id: string): Promise<void> {
    await apiClient.delete(`/courts/${id}`)
  },

  async getCourtMatches(courtId: string, status?: string) {
    const params = status ? { status } : {}
    const response = await apiClient.get(`/courts/${courtId}/matches`, { params })
    return response.data
  },

  async assignCourtLeader(courtId: string, userId: string): Promise<Court> {
    const response = await apiClient.post<Court>(`/courts/${courtId}/leaders`, { userId })
    return response.data
  },

  async removeCourtLeader(courtId: string, userId: string): Promise<Court> {
    const response = await apiClient.delete<Court>(`/courts/${courtId}/leaders`, { data: { userId } })
    return response.data
  },

  async updateCourtBonus(courtId: string, bonusExpPercent: number): Promise<Court> {
    const response = await apiClient.put<Court>(`/courts/${courtId}/bonus`, { bonusExpPercent })
    return response.data
  },

  async getMyCourts(): Promise<Court[]> {
    const response = await apiClient.get<Court[]>('/my-courts')
    return response.data
  },
}
