import React, { useState } from 'react'
import {
  Card,
  List,
  Tabs,
  Button,
  Typography,
  Space,
  Badge,
  Empty,
  Spin,
} from 'antd'
import { CheckOutlined } from '@ant-design/icons'
import dayjs from 'dayjs'
import { useNotifications } from '@/features/notification/hooks/useNotification'

const { Text } = Typography

const NOTIFICATION_TYPE_MAP: Record<string, string> = {
  refund_pending: '退款告警',
  low_stock: '库存告警',
  payment_failed: '支付失败',
  daily_report: '每日报告',
  weekly_report: '每周报告',
}

const NotificationItem: React.FC<{
  item: any
  onMarkRead: (id: number) => void
}> = ({ item, onMarkRead }) => (
  <List.Item
    style={{
      padding: '16px 20px',
      opacity: item.is_read ? 0.7 : 1,
      background: item.is_read ? '#fafafa' : '#fff',
      transition: 'all 0.3s',
    }}
    actions={
      !item.is_read
        ? [
            <Button
              key="read"
              type="link"
              size="small"
              icon={<CheckOutlined />}
              onClick={() => onMarkRead(item.id)}
            >
              标记已读
            </Button>,
          ]
        : []
    }
  >
    <List.Item.Meta
      avatar={
        <Badge
          status={item.is_read ? 'default' : 'processing'}
          text={
            <Text strong={!item.is_read} style={{ fontSize: 15 }}>
              {item.title}
            </Text>
          }
        />
      }
      title={
        <Space>
          <Text type="secondary" style={{ fontSize: 12 }}>
            {NOTIFICATION_TYPE_MAP[item.type] || item.type}
          </Text>
          <Text type="secondary" style={{ fontSize: 12 }}>
            {dayjs(item.created_at).format('YYYY-MM-DD HH:mm')}
          </Text>
        </Space>
      }
      description={
        <div style={{ marginTop: 8 }}>
          <Text style={{ fontSize: 14 }}>{item.content}</Text>
        </div>
      }
    />
  </List.Item>
)

export const NotificationsPage: React.FC = () => {
  const [activeTab, setActiveTab] = useState('all')
  const {
    notifications,
    unreadCount,
    loading,
    markRead,
    markAllRead,
  } = useNotifications()

  const filteredNotifications =
    activeTab === 'all'
      ? notifications
      : activeTab === 'unread'
      ? notifications.filter((n) => !n.is_read)
      : notifications.filter((n) => n.type === activeTab)

  const tabItems = [
    {
      key: 'all',
      label: (
        <Space>
          全部
          <Badge count={notifications.length} size="small" />
        </Space>
      ),
    },
    {
      key: 'unread',
      label: (
        <Space>
          未读
          <Badge count={unreadCount} size="small" />
        </Space>
      ),
    },
    {
      key: 'refund_pending',
      label: '退款告警',
    },
    {
      key: 'low_stock',
      label: '库存告警',
    },
    {
      key: 'daily_report',
      label: '报告',
    },
  ]

  return (
    <div style={{ padding: '24px 28px' }}>
      <div
        style={{
          marginBottom: 24,
          display: 'flex',
          justifyContent: 'space-between',
          alignItems: 'center',
        }}
      >
        <div>
          <h1
            style={{
              fontSize: 22,
              fontWeight: 700,
              color: '#1f2937',
              marginBottom: 4,
            }}
          >
            通知中心
          </h1>
          <Text type="secondary">
            共 {notifications.length} 条通知，{unreadCount} 条未读
          </Text>
        </div>
        {unreadCount > 0 && (
          <Button icon={<CheckOutlined />} onClick={markAllRead}>
            全部已读
          </Button>
        )}
      </div>

      <Card
        style={{ borderRadius: 12, border: '1px solid #e5e7eb' }}
        bodyStyle={{ padding: 0 }}
      >
        <Tabs
          activeKey={activeTab}
          onChange={setActiveTab}
          items={tabItems}
          style={{ padding: '0 20px' }}
        />

        <Spin spinning={loading}>
          {filteredNotifications.length === 0 ? (
            <div style={{ padding: 60, textAlign: 'center' }}>
              <Empty
                description={activeTab === 'unread' ? '暂无未读通知' : '暂无通知'}
                image={Empty.PRESENTED_IMAGE_SIMPLE}
              />
            </div>
          ) : (
            <List
              dataSource={filteredNotifications}
              renderItem={(item) => (
                <NotificationItem item={item} onMarkRead={markRead} />
              )}
            />
          )}
        </Spin>
      </Card>
    </div>
  )
}

export default NotificationsPage
