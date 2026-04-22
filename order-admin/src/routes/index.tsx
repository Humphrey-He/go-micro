import React, { Suspense, lazy } from 'react'
import { BrowserRouter, Routes, Route, Navigate } from 'react-router-dom'
import { Spin } from 'antd'
import { BasicLayout, BlankLayout } from '@/layouts'
import { ProtectedRoute } from './ProtectedRoute'
import { paths } from './paths'

const LoginPage = lazy(() => import('@/pages/Login'))
const DashboardPage = lazy(() => import('@/pages/Dashboard'))
const OrderListPage = lazy(() => import('@/pages/Orders/OrderList'))
const OrderDetailPage = lazy(() => import('@/pages/Orders/OrderDetail'))
const PaymentListPage = lazy(() => import('@/pages/Payments/PaymentListPage'))
const InventoryPage = lazy(() => import('@/pages/Inventory/InventoryPage'))

const LoadingFallback: React.FC = () => (
  <div
    style={{
      display: 'flex',
      justifyContent: 'center',
      alignItems: 'center',
      height: '100vh',
    }}
  >
    <Spin size="large" />
  </div>
)

export const AppRoutes: React.FC = () => {
  return (
    <BrowserRouter>
      <Suspense fallback={<LoadingFallback />}>
        <Routes>
          <Route element={<BlankLayout />}>
            <Route path={paths.login} element={<LoginPage />} />
          </Route>

          <Route
            element={
              <ProtectedRoute>
                <BasicLayout />
              </ProtectedRoute>
            }
          >
            <Route path={paths.dashboard} element={<DashboardPage />} />
            <Route path={paths.orders} element={<OrderListPage />} />
            <Route path={paths.orderDetail} element={<OrderDetailPage />} />
            <Route path={paths.payments} element={<PaymentListPage />} />
            <Route path={paths.inventory} element={<InventoryPage />} />
          </Route>

          <Route path="/" element={<Navigate to={paths.dashboard} replace />} />
          <Route path="*" element={<Navigate to={paths.orders} replace />} />
        </Routes>
      </Suspense>
    </BrowserRouter>
  )
}
