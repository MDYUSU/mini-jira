import { Navigate, Route, Routes } from 'react-router-dom'
import { AuthGate } from './components/AuthGate'
import { AppLayout } from './components/AppLayout'
import { LoginPage } from './pages/LoginPage'
import { RegisterPage } from './pages/RegisterPage'
import { HomePage } from './pages/HomePage'
import { ProjectPage } from './pages/project/ProjectPage'
import { TaskPage } from './pages/TaskPage'
import { ProfilePage } from './pages/ProfilePage'
import { isLoggedIn } from './lib/auth'

function PublicOnly({ children }: { children: React.ReactNode }) {
  return isLoggedIn() ? <Navigate to="/" replace /> : <>{children}</>
}

export default function App() {
  return (
    <Routes>
      <Route path="/login" element={<PublicOnly><LoginPage /></PublicOnly>} />
      <Route path="/register" element={<PublicOnly><RegisterPage /></PublicOnly>} />
      <Route element={<AuthGate><AppLayout /></AuthGate>}>
        <Route path="/" element={<HomePage />} />
        <Route path="/orgs/:orgId" element={<HomePage />} />
        <Route path="/projects/:projectId" element={<ProjectPage />} />
        <Route path="/projects/:projectId/:tab" element={<ProjectPage />} />
        <Route path="/tasks/:taskId" element={<TaskPage />} />
        <Route path="/profile" element={<ProfilePage />} />
      </Route>
      <Route path="*" element={<Navigate to="/" replace />} />
    </Routes>
  )
}
