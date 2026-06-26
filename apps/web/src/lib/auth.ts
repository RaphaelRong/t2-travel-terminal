const TOKEN_KEY = 't2_access_token'
const SUPERADMIN_KEY = 't2_is_superadmin'
const ROLE_KEY = 't2_role'

export type UserRole =
  | 'superadmin'
  | 'free_user'
  | 'paid_user'
  | 'premium_paid_user'

export function getToken(): string | null {
  return localStorage.getItem(TOKEN_KEY)
}

export function setToken(token: string): void {
  localStorage.setItem(TOKEN_KEY, token)
}

export function removeToken(): void {
  localStorage.removeItem(TOKEN_KEY)
}

export function getSuperAdmin(): boolean {
  return localStorage.getItem(SUPERADMIN_KEY) === 'true'
}

export function setSuperAdmin(isSuperAdmin: boolean): void {
  localStorage.setItem(SUPERADMIN_KEY, isSuperAdmin ? 'true' : 'false')
}

export function removeSuperAdmin(): void {
  localStorage.removeItem(SUPERADMIN_KEY)
}

export function getRole(): UserRole | null {
  const role = localStorage.getItem(ROLE_KEY)
  if (
    role === 'superadmin' ||
    role === 'free_user' ||
    role === 'paid_user' ||
    role === 'premium_paid_user'
  ) {
    return role
  }
  return null
}

export function setRole(role: UserRole): void {
  localStorage.setItem(ROLE_KEY, role)
}

export function removeRole(): void {
  localStorage.removeItem(ROLE_KEY)
}
