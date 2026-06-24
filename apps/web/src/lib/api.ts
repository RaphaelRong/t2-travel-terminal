import axios from 'axios'
import { getToken, removeToken } from './auth'
import { useTenantStore } from '../store/tenantStore'

export const api = axios.create({
  baseURL: '/api/v1',
  headers: {
    'Content-Type': 'application/json',
  },
})

api.interceptors.request.use((config) => {
  const token = getToken()
  if (token) {
    config.headers.Authorization = `Bearer ${token}`
  }
  const currentTenant = useTenantStore.getState().currentTenant
  if (currentTenant) {
    config.headers['X-Tenant-ID'] = currentTenant.id
  }
  return config
})

api.interceptors.response.use(
  (response) => response,
  (error) => {
    if (error.response?.status === 401) {
      removeToken()
      window.location.href = '/login'
    }
    return Promise.reject(error)
  },
)
