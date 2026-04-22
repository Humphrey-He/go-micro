import { ComponentType } from 'react'

export interface AppRoutes {
  path: string
  component: ComponentType
  layout: ComponentType<{ children: React.ReactNode }>
  requiresAuth?: boolean
}