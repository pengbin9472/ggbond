/**
 * Monitoring API endpoints
 * Provides group monitoring statistics and history
 */

import { apiClient } from './client'
import type { GroupMonitoringStat, MonitoringHistoryPoint } from '@/types'

export interface GroupMonitoringResponse {
  groups: GroupMonitoringStat[]
}

export interface GroupMonitoringHistoryResponse {
  history: MonitoringHistoryPoint[]
}

/**
 * Get group monitoring statistics
 */
export async function getGroupMonitoring(): Promise<GroupMonitoringResponse> {
  const { data } = await apiClient.get<GroupMonitoringResponse>('/monitoring/groups')
  return data
}

/**
 * Get monitoring history for a specific group
 */
export async function getGroupMonitoringHistory(groupId: number, limit?: number): Promise<GroupMonitoringHistoryResponse> {
  const params = limit ? { limit } : {}
  const { data } = await apiClient.get<GroupMonitoringHistoryResponse>(`/monitoring/groups/${groupId}/history`, { params })
  return data
}

export const monitoringAPI = {
  getGroupMonitoring,
  getGroupMonitoringHistory
}

export default monitoringAPI
