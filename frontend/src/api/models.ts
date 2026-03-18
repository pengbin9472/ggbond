import apiClient from './client'
import type { ModelCatalogResponse } from '@/types'

export async function getModelCatalog(): Promise<ModelCatalogResponse> {
  const { data } = await apiClient.get<ModelCatalogResponse>('/models/catalog')
  return data
}
