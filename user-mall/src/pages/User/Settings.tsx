import { useState } from 'react'
import { useNavigate } from 'react-router-dom'
import { Card, List, Switch, Button, Dialog, Toast } from 'antd-mobile'
import {
  GlobalOutline,
  BellOutline,
  LockOutline,
  QuestionCircleOutline,
  HandPayCircleOutline,
} from 'antd-mobile-icons'
import { useAuthStore } from '@/stores/authStore'

export default function Settings() {
  const navigate = useNavigate()
  const { logout } = useAuthStore()
  const [notifications, setNotifications] = useState({
    push: true,
    order: true,
    promotion: false,
    sms: true,
  })

  const handleLogout = () => {
    Dialog.confirm({
      title: '退出登录',
      content: '确定要退出登录吗？',
      confirmText: '退出',
      onConfirm: () => {
        logout()
        Toast.show('已退出登录')
        navigate('/login')
      },
    })
  }

  return (
    <div className="min-h-screen bg-gray-50 pb-20">
      {/* 账号与安全 */}
      <Card className="m-2">
        <div className="font-bold mb-3">账号与安全</div>
        <List>
          <List.Item
            prefix={<LockOutline />}
            extra="未绑定"
            onClick={() => Toast.show('功能开发中')}
          >
            登录密码
          </List.Item>
          <List.Item
            prefix={<HandPayCircleOutline />}
            extra="未绑定"
            onClick={() => Toast.show('功能开发中')}
          >
            支付密码
          </List.Item>
          <List.Item
            prefix={<GlobalOutline />}
            extra="中国大陆"
            onClick={() => Toast.show('功能开发中')}
          >
            地区设置
          </List.Item>
        </List>
      </Card>

      {/* 通知设置 */}
      <Card className="m-2">
        <div className="font-bold mb-3">通知设置</div>
        <List>
          <List.Item
            prefix={<BellOutline />}
            extra={
              <Switch
                checked={notifications.push}
                onChange={(checked) =>
                  setNotifications((prev) => ({ ...prev, push: checked }))
                }
              />
            }
          >
            推送通知
          </List.Item>
          <List.Item
            extra={
              <Switch
                checked={notifications.order}
                onChange={(checked) =>
                  setNotifications((prev) => ({ ...prev, order: checked }))
                }
              />
            }
          >
            订单通知
          </List.Item>
          <List.Item
            extra={
              <Switch
                checked={notifications.promotion}
                onChange={(checked) =>
                  setNotifications((prev) => ({ ...prev, promotion: checked }))
                }
              />
            }
          >
            促销通知
          </List.Item>
          <List.Item
            extra={
              <Switch
                checked={notifications.sms}
                onChange={(checked) =>
                  setNotifications((prev) => ({ ...prev, sms: checked }))
                }
              />
            }
          >
            短信通知
          </List.Item>
        </List>
      </Card>

      {/* 通用设置 */}
      <Card className="m-2">
        <div className="font-bold mb-3">通用设置</div>
        <List>
          <List.Item
            prefix={<GlobalOutline />}
            extra="简体中文"
            onClick={() => Toast.show('功能开发中')}
          >
            语言
          </List.Item>
          <List.Item
            prefix={<GlobalOutline />}
            extra="跟随系统"
            onClick={() => Toast.show('功能开发中')}
          >
            主题
          </List.Item>
          <List.Item
            prefix={<QuestionCircleOutline />}
            onClick={() => navigate('/help')}
          >
            帮助与反馈
          </List.Item>
          <List.Item
            prefix={<QuestionCircleOutline />}
            onClick={() => navigate('/about')}
          >
            关于我们
          </List.Item>
        </List>
      </Card>

      {/* 协议 */}
      <Card className="m-2">
        <List>
          <List.Item onClick={() => Toast.show('功能开发中')}>
            用户协议
          </List.Item>
          <List.Item onClick={() => Toast.show('功能开发中')}>
            隐私政策
          </List.Item>
          <List.Item onClick={() => Toast.show('功能开发中')}>
            资质证照
          </List.Item>
        </List>
      </Card>

      {/* 退出登录 */}
      <div className="p-4">
        <Button block color="danger" onClick={handleLogout}>
          退出登录
        </Button>
      </div>

      {/* 版本信息 */}
      <div className="text-center text-xs text-gray-400 pb-4">
        v1.0.0
      </div>
    </div>
  )
}