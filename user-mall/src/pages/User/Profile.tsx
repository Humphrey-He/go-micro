import { useNavigate } from 'react-router-dom'
import { Card, Avatar, Button, Toast, List, Badge } from 'antd-mobile'
import {
  LocationOutline,
  CouponOutline,
  StarOutline,
  MessageOutline,
  TeamOutline,
  ClockCircleOutline,
  SendOutline,
  TruckOutline,
  EditSFill,
  FireFill,
  CalendarOutline,
  UnorderedListOutline,
  GiftOutline,
  SetOutline,
  QuestionCircleOutline,
  PhoneFill,
} from 'antd-mobile-icons'
import { useAuthStore } from '@/stores/authStore'

export default function Profile() {
  const navigate = useNavigate()
  const { userInfo, logout } = useAuthStore()

  const handleLogout = () => {
    logout()
    Toast.show('已退出登录')
    navigate('/login')
  }

  return (
    <div className="min-h-screen bg-gray-100 pb-20">
      {/* 用户信息 */}
      <div className="bg-gradient-to-br from-primary-500 to-primary-600 text-white p-6 pb-8">
        <div className="flex items-center gap-4">
          <div className="relative">
            <Avatar
              src={userInfo?.avatar || ''}
              className="w-16 h-16 border-2 border-white border-opacity-50"
              style={{ '--size': '64px' }}
            />
            {userInfo && (
              <div className="absolute -bottom-1 -right-1 w-5 h-5 bg-yellow-400 rounded-full flex items-center justify-center text-xs text-yellow-900 font-bold">
                V{userInfo.level}
              </div>
            )}
          </div>
          <div className="flex-1">
            <div className="text-xl font-bold">{userInfo?.nickname || '未登录'}</div>
            <div className="text-sm opacity-80 mt-1">
              {userInfo?.phone || '登录后享受更多权益'}
            </div>
          </div>
          {!userInfo && (
            <Button size="small" color="warning" onClick={() => navigate('/login')}>
              登录
            </Button>
          )}
          {userInfo && (
            <Button
              size="small"
              fill="outline"
              onClick={() => navigate('/settings')}
              className="text-white border-white"
            >
              设置
            </Button>
          )}
        </div>

        {userInfo && (
          <div
            className="flex justify-around mt-5 pt-4 border-t border-white border-opacity-20"
            onClick={() => navigate('/membership')}
          >
            <div className="text-center cursor-pointer">
              <div className="text-2xl font-bold">{userInfo.points}</div>
              <div className="text-xs opacity-80 mt-1">积分</div>
            </div>
            <div className="w-px bg-white border-opacity-20" />
            <div className="text-center cursor-pointer">
              <div className="text-2xl font-bold">{userInfo.member_title}</div>
              <div className="text-xs opacity-80 mt-1">会员等级</div>
            </div>
            <div className="w-px bg-white border-opacity-20" />
            <div className="text-center cursor-pointer">
              <div className="text-2xl font-bold">0</div>
              <div className="text-xs opacity-80 mt-1">优惠券</div>
            </div>
          </div>
        )}
      </div>

      {/* 订单入口 */}
      <Card className="mx-2 -mt-4 rounded-xl shadow-sm">
        <div className="flex items-center justify-between mb-4">
          <span className="font-bold text-base">我的订单</span>
          <span
            className="text-sm text-gray-400 flex items-center"
            onClick={() => navigate('/orders')}
          >
            查看全部
            <UnorderedListOutline style={{ fontSize: 14 }} />
          </span>
        </div>
        <div className="flex justify-around py-2">
          {[
            { icon: <ClockCircleOutline className="text-2xl" />, label: '待付款', status: 'PENDING_PAYMENT', count: 0, color: '#fa8c16' },
            { icon: <SendOutline className="text-2xl" />, label: '待发货', status: 'PAID', count: 0, color: '#1890ff' },
            { icon: <TruckOutline className="text-2xl" />, label: '待收货', status: 'SHIPPED', count: 1, color: '#52c41a' },
            { icon: <EditSFill className="text-2xl" />, label: '待评价', status: 'CONFIRMED', count: 0, color: '#722ed1' },
          ].map((item) => (
            <div
              key={item.status}
              className="flex flex-col items-center"
              onClick={() => navigate(`/orders?status=${item.status}`)}
            >
              <div className="relative">
                <div
                  className="w-12 h-12 rounded-full bg-gray-50 flex items-center justify-center"
                  style={{ color: item.color }}
                >
                  {item.icon}
                </div>
                {item.count > 0 && (
                  <Badge
                    content={item.count}
                    className="absolute -top-1 -right-1"
                    style={{ backgroundColor: '#ff4d4f' }}
                  />
                )}
              </div>
              <span className="text-xs text-gray-600 mt-2">{item.label}</span>
            </div>
          ))}
        </div>
      </Card>

      {/* 营销活动入口 */}
      <Card className="mx-2 mt-2 rounded-xl shadow-sm">
        <div className="flex items-center justify-between mb-4">
          <span className="font-bold text-base">专属服务</span>
        </div>
        <div className="grid grid-cols-4 gap-4">
          {[
            { icon: <FireFill className="text-xl" />, label: '秒杀', color: '#ff6a00', bg: '#fff7e6' },
            { icon: <CouponOutline className="text-xl" />, label: '优惠券', color: '#ff4d4f', bg: '#fff1f0' },
            { icon: <GiftOutline className="text-xl" />, label: '礼品卡', color: '#eb2f96', bg: '#fff0f3' },
            { icon: <CalendarOutline className="text-xl" />, label: '签到', color: '#52c41a', bg: '#f6ffed' },
          ].map((item) => (
            <div
              key={item.label}
              className="flex flex-col items-center cursor-pointer"
              onClick={() => navigate('/seckill')}
            >
              <div
                className="w-11 h-11 rounded-full flex items-center justify-center"
                style={{ backgroundColor: item.bg, color: item.color }}
              >
                {item.icon}
              </div>
              <span className="text-xs text-gray-600 mt-1.5">{item.label}</span>
            </div>
          ))}
        </div>
      </Card>

      {/* 功能列表 */}
      <Card className="mx-2 mt-2 rounded-xl shadow-sm">
        <List mode="card" className="bg-transparent">
          <List.Item
            prefix={<LocationOutline className="text-lg" style={{ color: '#1890ff' }} />}
            arrow
            onClick={() => navigate('/address')}
          >
            收货地址
          </List.Item>
          <List.Item
            prefix={<CouponOutline className="text-lg" style={{ color: '#ff4d4f' }} />}
            arrow
            onClick={() => navigate('/coupons')}
          >
            我的优惠券
          </List.Item>
          <List.Item
            prefix={<StarOutline className="text-lg" style={{ color: '#faad14' }} />}
            arrow
            onClick={() => navigate('/favorites')}
          >
            收藏夹
          </List.Item>
          <List.Item
            prefix={<MessageOutline className="text-lg" style={{ color: '#722ed1' }} />}
            arrow
            onClick={() => navigate('/messages')}
          >
            消息通知
            <span className="ml-2 text-xs text-gray-400">3条新消息</span>
          </List.Item>
          <List.Item
            prefix={<TeamOutline className="text-lg" style={{ color: '#eb2f96' }} />}
            arrow
            onClick={() => navigate('/membership')}
          >
            会员中心
          </List.Item>
        </List>
      </Card>

      {/* 其他功能 */}
      <Card className="mx-2 mt-2 rounded-xl shadow-sm">
        <List mode="card" className="bg-transparent">
          <List.Item
            prefix={<PhoneFill className="text-lg" style={{ color: '#52c41a' }} />}
            arrow
            onClick={() => navigate('/support')}
          >
            客服中心
          </List.Item>
          <List.Item
            prefix={<QuestionCircleOutline className="text-lg" style={{ color: '#8c8c8c' }} />}
            arrow
            onClick={() => navigate('/help')}
          >
            帮助与反馈
          </List.Item>
          <List.Item
            prefix={<SetOutline className="text-lg" style={{ color: '#595959' }} />}
            arrow
            onClick={() => navigate('/settings')}
          >
            设置
          </List.Item>
        </List>
      </Card>

      {/* 退出登录 */}
      {userInfo && (
        <div className="p-4 mt-2">
          <Button block color="danger" className="rounded-lg" onClick={handleLogout}>
            退出登录
          </Button>
        </div>
      )}

      {/* 底部版权 */}
      <div className="text-center text-xs text-gray-400 py-4">
        兴趣电商 v1.0.0
      </div>
    </div>
  )
}