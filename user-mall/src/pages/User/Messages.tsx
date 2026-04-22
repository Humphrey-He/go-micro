import { useState } from 'react'
import { Badge, Empty, Tabs } from 'antd-mobile'
import {
  BellOutline,
  MessageOutline,
  GiftOutline,
  TruckOutline,
  ExclamationCircleOutline,
} from 'antd-mobile-icons'

interface Message {
  id: string
  type: 'order' | 'promotion' | 'system' | 'activity'
  title: string
  content: string
  time: string
  read: boolean
  link?: string
}

const mockMessages: Message[] = [
  {
    id: '1',
    type: 'order',
    title: '订单已发货',
    content: '您的订单 #ORDER123456 已发货，快递单号：SF1234567890',
    time: '10:30',
    read: false,
    link: '/order/ORDER123456',
  },
  {
    id: '2',
    type: 'promotion',
    title: '限时优惠来袭',
    content: '新人专享券来了！立即领取享受首单优惠',
    time: '昨天',
    read: false,
  },
  {
    id: '3',
    type: 'system',
    title: '账号安全提醒',
    content: '您的账号在新设备登录，如非本人操作请及时修改密码',
    time: '昨天',
    read: true,
  },
  {
    id: '4',
    type: 'activity',
    title: '拼团成功通知',
    content: '您参与的iPhone 15拼团已成功，商品即将发货',
    time: '前天',
    read: true,
  },
]

const getMessageIcon = (type: string) => {
  switch (type) {
    case 'order':
      return <TruckOutline className="text-blue-500" />
    case 'promotion':
      return <GiftOutline className="text-red-500" />
    case 'system':
      return <ExclamationCircleOutline className="text-orange-500" />
    case 'activity':
      return <MessageOutline className="text-green-500" />
    default:
      return <BellOutline className="text-gray-500" />
  }
}

export default function Messages() {
  const [messages, setMessages] = useState<Message[]>(mockMessages)
  const [activeTab, setActiveTab] = useState('all')

  const unreadCount = messages.filter((m) => !m.read).length

  const filteredMessages = messages.filter((m) => {
    if (activeTab === 'all') return true
    if (activeTab === 'unread') return !m.read
    return m.type === activeTab
  })

  const handleRead = (id: string) => {
    setMessages((prev) =>
      prev.map((m) => (m.id === id ? { ...m, read: true } : m))
    )
  }

  const handleReadAll = () => {
    setMessages((prev) => prev.map((m) => ({ ...m, read: true })))
  }

  return (
    <div className="min-h-screen bg-gray-50">
      {/* 顶部 */}
      <div className="bg-white p-4 flex items-center justify-between">
        <div className="flex items-center gap-2">
          <span className="font-bold text-lg">消息通知</span>
          {unreadCount > 0 && (
            <Badge content={unreadCount} color="var(--adm-color-danger)" />
          )}
        </div>
        {unreadCount > 0 && (
          <span
            className="text-sm text-primary-500"
            onClick={handleReadAll}
          >
            全部已读
          </span>
        )}
      </div>

      <Tabs activeKey={activeTab} onChange={(key) => setActiveTab(key)}>
        <Tabs.Tab title="全部" key="all" />
        <Tabs.Tab
          title={
            <span>
              未读
              {unreadCount > 0 && (
                <Badge content={unreadCount} color="var(--adm-color-danger)" />
              )}
            </span>
          }
          key="unread"
        />
        <Tabs.Tab title="订单" key="order" />
        <Tabs.Tab title="优惠" key="promotion" />
        <Tabs.Tab title="活动" key="activity" />
      </Tabs>

      {/* 消息列表 */}
      <div className="bg-white">
        {filteredMessages.length === 0 ? (
          <Empty description="暂无消息" className="py-12" />
        ) : (
          filteredMessages.map((message) => (
            <div
              key={message.id}
              className={`p-4 border-b border-gray-100 ${
                !message.read ? 'bg-blue-50' : ''
              }`}
              onClick={() => handleRead(message.id)}
            >
              <div className="flex items-start gap-3">
                <div className="text-2xl">{getMessageIcon(message.type)}</div>
                <div className="flex-1 min-w-0">
                  <div className="flex items-center justify-between">
                    <span className="font-medium">{message.title}</span>
                    <span className="text-xs text-gray-400">{message.time}</span>
                  </div>
                  <div className="text-sm text-gray-500 mt-1 line-clamp-2">
                    {message.content}
                  </div>
                  {!message.read && (
                    <div className="mt-2">
                      <span className="w-2 h-2 rounded-full bg-primary-500 inline-block" />
                    </div>
                  )}
                </div>
              </div>
            </div>
          ))
        )}
      </div>
    </div>
  )
}