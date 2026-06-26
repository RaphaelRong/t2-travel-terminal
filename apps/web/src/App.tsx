import { useEffect } from 'react'
import { BrowserRouter, Routes, Route, Navigate } from 'react-router-dom'
import { useAuthStore } from './store/authStore'
import { Layout } from './components/Layout'
import { ProtectedRoute } from './components/ProtectedRoute'
import { SubscriptionRoute } from './components/SubscriptionRoute'
import { AdminRoute } from './components/AdminRoute'
import { HomePage } from './pages/HomePage'
import { LoginPage } from './pages/LoginPage'
import { RegisterPage } from './pages/RegisterPage'
import { VerifyEmailPage } from './pages/VerifyEmailPage'
import { PlaygroundPage } from './pages/PlaygroundPage'
import { ProjectsPage } from './pages/ProjectsPage'
import { MembersPage } from './pages/MembersPage'
import { PlansPage } from './pages/PlansPage'
import { ProfilePage } from './pages/ProfilePage'
import { AdminPage } from './pages/AdminPage'
import { AdminPlansPage } from './pages/AdminPlansPage'
import { AdminProjectsPage } from './pages/AdminProjectsPage'
import { AdminUsersPage } from './pages/AdminUsersPage'

function AppInitializer() {
  const init = useAuthStore((s) => s.init)
  useEffect(() => {
    init()
  }, [init])
  return null
}

function App() {
  return (
    <BrowserRouter>
      <AppInitializer />
      <Routes>
        <Route path="/" element={<HomePage />} />
        <Route path="/login" element={<LoginPage />} />
        <Route path="/register" element={<RegisterPage />} />
        <Route path="/verify" element={<VerifyEmailPage />} />
        <Route element={<ProtectedRoute />}>
          <Route element={<Layout />}>
            <Route path="/dashboard" element={<Navigate to="/playground" replace />} />
            <Route path="/playground" element={<PlaygroundPage />} />
            <Route path="/plans" element={<PlansPage />} />
            <Route path="/profile" element={<ProfilePage />} />
          </Route>
        </Route>
        <Route element={<ProtectedRoute />}>
          <Route element={<SubscriptionRoute />}>
            <Route element={<Layout />}>
              <Route path="/projects" element={<ProjectsPage />} />
              <Route path="/members" element={<MembersPage />} />
            </Route>
          </Route>
        </Route>
        <Route element={<AdminRoute />}>
          <Route element={<Layout />}>
            <Route path="/admin" element={<AdminPage />}>
              <Route index element={<Navigate to="/admin/plans" replace />} />
              <Route path="plans" element={<AdminPlansPage />} />
              <Route path="projects" element={<AdminProjectsPage />} />
              <Route path="users" element={<AdminUsersPage />} />
            </Route>
          </Route>
        </Route>
        <Route path="*" element={<Navigate to="/" replace />} />
      </Routes>
    </BrowserRouter>
  )
}

export default App
