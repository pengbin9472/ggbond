import { apiClient } from './client'

export interface ReferralCodeResponse {
  code: string
}

export interface ReferralStats {
  total_rewards: number
  invitee_count: number
  reward_type: string
  reward_rate: number
}

export interface ReferralReward {
  id: number
  invitee_id: number
  invitee_email: string
  reward_amount: number
  reward_type: string
  reward_rate: number
  trigger_code_value: number
  created_at: string
}

export interface ReferralHistoryResponse {
  items: ReferralReward[]
  total: number
  page: number
  page_size: number
}

export async function getInvitationCode(): Promise<ReferralCodeResponse> {
  const { data } = await apiClient.get<ReferralCodeResponse>('/referrals/code')
  return data
}

export async function getStats(): Promise<ReferralStats> {
  const { data } = await apiClient.get<ReferralStats>('/referrals/stats')
  return data
}

export async function getHistory(page = 1, pageSize = 20): Promise<ReferralHistoryResponse> {
  const { data } = await apiClient.get<ReferralHistoryResponse>('/referrals/history', {
    params: { page, page_size: pageSize }
  })
  return data
}

export const referralAPI = {
  getInvitationCode,
  getStats,
  getHistory
}

export default referralAPI
