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
  // 所有需要租户上下文的接口统一由后端根据用户角色授权。
  // SuperAdmin 即使不属于该租户，也能通过 tenant.Middleware 进入任意租户上下文。
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
