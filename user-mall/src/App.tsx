import { Routes, Route } from 'react-router-dom'
import { lazy, Suspense } from 'react'
import { BasicLayout } from './layouts/BasicLayout'
import { BlankLayout } from './layouts/BlankLayout'
import Loading from './components/LoadingSkeleton'

const Home = lazy(() => import('./pages/Home'))
const ProductList = lazy(() => import('./pages/ProductList'))
const ProductDetail = lazy(() => import('./pages/ProductDetail'))
const ProductReviews = lazy(() => import('./pages/ProductReviews'))
const Search = lazy(() => import('./pages/Search'))
const Cart = lazy(() => import('./pages/Cart'))
const Checkout = lazy(() => import('./pages/Checkout'))
const Payment = lazy(() => import('./pages/Payment'))
const PaymentResult = lazy(() => import('./pages/Payment/Result'))
const OrderList = lazy(() => import('./pages/Orders/OrderList'))
const OrderDetail = lazy(() => import('./pages/OrderDetail'))
const Login = lazy(() => import('./pages/Auth/Login'))
const Register = lazy(() => import('./pages/Auth/Register'))
const UserProfile = lazy(() => import('./pages/User/Profile'))
const AddressList = lazy(() => import('./pages/User/AddressList'))
const AddressEdit = lazy(() => import('./pages/User/AddressEdit'))
const CouponList = lazy(() => import('./pages/User/CouponList'))
const FavoriteList = lazy(() => import('./pages/User/FavoriteList'))
const Footprint = lazy(() => import('./pages/User/Footprint'))
const Messages = lazy(() => import('./pages/User/Messages'))
const Membership = lazy(() => import('./pages/User/Membership'))
const Settings = lazy(() => import('./pages/User/Settings'))
const Seckill = lazy(() => import('./pages/Activity/Seckill'))
const CouponCenter = lazy(() => import('./pages/Activity/CouponCenter'))
const CheckIn = lazy(() => import('./pages/Activity/CheckIn'))
const GroupBuy = lazy(() => import('./pages/Activity/GroupBuy'))

const routes = [
  // 首页 & 商品
  { path: '/', component: Home, layout: BasicLayout },
  { path: '/product/list', component: ProductList, layout: BasicLayout },
  { path: '/product/:skuId', component: ProductDetail, layout: BasicLayout },
  { path: '/product/:skuId/reviews', component: ProductReviews, layout: BasicLayout },
  { path: '/search', component: Search, layout: BasicLayout },

  // 购物
  { path: '/cart', component: Cart, layout: BasicLayout },
  { path: '/checkout', component: Checkout, layout: BasicLayout, requiresAuth: true },

  // 支付
  { path: '/payment/:orderNo', component: Payment, layout: BasicLayout, requiresAuth: true },
  { path: '/payment/result/:orderNo', component: PaymentResult, layout: BasicLayout },

  // 订单
  { path: '/orders', component: OrderList, layout: BasicLayout, requiresAuth: true },
  { path: '/order/:orderNo', component: OrderDetail, layout: BasicLayout, requiresAuth: true },

  // 认证
  { path: '/login', component: Login, layout: BlankLayout },
  { path: '/register', component: Register, layout: BlankLayout },

  // 个人中心
  { path: '/user', component: UserProfile, layout: BasicLayout, requiresAuth: true },
  { path: '/address', component: AddressList, layout: BasicLayout, requiresAuth: true },
  { path: '/address/edit/:id', component: AddressEdit, layout: BasicLayout, requiresAuth: true },
  { path: '/coupons', component: CouponList, layout: BasicLayout, requiresAuth: true },
  { path: '/favorites', component: FavoriteList, layout: BasicLayout, requiresAuth: true },
  { path: '/footprint', component: Footprint, layout: BasicLayout, requiresAuth: true },
  { path: '/messages', component: Messages, layout: BasicLayout, requiresAuth: true },
  { path: '/membership', component: Membership, layout: BasicLayout, requiresAuth: true },
  { path: '/settings', component: Settings, layout: BasicLayout, requiresAuth: true },

  // 活动
  { path: '/seckill', component: Seckill, layout: BasicLayout },
  { path: '/coupon-center', component: CouponCenter, layout: BasicLayout },
  { path: '/checkin', component: CheckIn, layout: BasicLayout, requiresAuth: true },
  { path: '/groupbuy', component: GroupBuy, layout: BasicLayout },
]

function App() {
  return (
    <Suspense fallback={<Loading />}>
      <Routes>
        {routes.map((route) => (
          <Route
            key={route.path}
            path={route.path}
            element={
              <route.layout>
                <route.component />
              </route.layout>
            }
          />
        ))}
      </Routes>
    </Suspense>
  )
}

export default App