import React from 'react'
import { Badge, Popover, List, Button, Typography, Space, Empty } from 'antd'
import { BellOutlined } from '@ant-design/icons'
import dayjs from 'dayjs'
import { useNotifications } from '../hooks/useNotification'
import { useUnreadCount } from '../hooks/useUnreadCount'

const { Text, Title } = Typography

export const NotificationBell: React.FC = () => {
  const { count } = useUnreadCount()
  const { notifications, markRead, markAllRead, refresh } = useNotifications()

  const popoverContent = (
    <div style={{ width: 360, maxHeight: 480, overflow: 'auto' }}>
      <div
        style={{
          display: 'flex',
          justifyContent: 'space-between',
          alignItems: 'center',
          marginBottom: 12,
          padding: '8px 0',
          borderBottom: '1px solid #f0f0f0',
        }}
      >
        <Title level={5} style={{ margin: 0 }}>
          通知中心
        </Title>
        {count > 0 && (
          <Button type="link" size="small" onClick={markAllRead}>
            全部已读
          </Button>
        )}
      </div>

      {notifications.length === 0 ? (
        <Empty description="暂无通知" image={Empty.PRESENTED_IMAGE_SIMPLE} />
      ) : (
        <List
          dataSource={notifications.slice(0, 10)}
          renderItem={(item) => (
            <List.Item
              style={{
                padding: '10px 0',
                opacity: item.is_read ? 0.6 : 1,
                cursor: 'pointer',
              }}
              onClick={() => !item.is_read && markRead(item.id)}
            >
              <List.Item.Meta
                title={
                  <Space>
                    {!item.is_read && (
                      <Badge status="processing" color="#1677ff" />
                    )}
                    <Text strong={!item.is_read}>{item.title}</Text>
                  </Space>
                }
                description={
                  <div>
                    <Text type="secondary" style={{ fontSize: 12 }}>
                      {item.content}
                    </Text>
                    <div style={{ marginTop: 4 }}>
                      <Text type="secondary" style={{ fontSize: 11 }}>
                        {dayjs(item.created_at).format('MM-DD HH:mm')}
                      </Text>
                    </div>
                  </div>
                }
              />
            </List.Item>
          )}
        />
      )}

      <div style={{ textAlign: 'center', marginTop: 8 }}>
        <Button type="link" size="small" onClick={refresh}>
          刷新
        </Button>
      </div>
    </div>
  )

  return (
    <Popover
      content={popoverContent}
      trigger="click"
      placement="bottomRight"
      arrow={{ pointAtCenter: true }}
    >
      <Badge count={count} size="small" offset={[-2, 2]}>
        <Button
          type="text"
          icon={<BellOutlined style={{ fontSize: 20 }} />}
          style={{ color: '#fff' }}
        />
      </Badge>
    </Popover>
  )
}
