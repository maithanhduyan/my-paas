import { Routes, Route, Navigate } from 'react-router-dom'
import { Layout } from './components/Layout'
import { Dashboard } from './pages/Dashboard'
import { ProjectDetail } from './pages/ProjectDetail'
import { NewProject } from './pages/NewProject'
import { Services } from './pages/Services'
import { Login } from './pages/Login'
import { Register } from './pages/Register'
import { Backups } from './pages/Backups'
import { Users } from './pages/Users'
import { Marketplace } from './pages/Marketplace'
import { Swarm } from './pages/Swarm'

export function App() {
  return (
    <Routes>
      <Route path="/login" element={<Login />} />
      <Route path="/register/:token" element={<Register />} />
      <Route element={<Layout />}>
        <Route path="/" element={<Dashboard />} />
        <Route path="/projects/new" element={<NewProject />} />
        <Route path="/projects/:id" element={<ProjectDetail />} />
        <Route path="/services" element={<Services />} />
        <Route path="/backups" element={<Backups />} />
        <Route path="/users" element={<Users />} />
        <Route path="/marketplace" element={<Marketplace />} />
        <Route path="/swarm" element={<Swarm />} />
        <Route path="*" element={<Navigate to="/" replace />} />
      </Route>
    </Routes>
  )
}
