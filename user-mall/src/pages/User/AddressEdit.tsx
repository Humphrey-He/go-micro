import { useState } from 'react'
import { useNavigate, useParams } from 'react-router-dom'
import { Card, Form, Input, Button, Switch, Toast } from 'antd-mobile'

export default function AddressEdit() {
  const navigate = useNavigate()
  const { id } = useParams<{ id: string }>()
  const [loading, setLoading] = useState(false)
  const isNew = id === 'new'
  const [isDefault, setIsDefault] = useState(false)

  const handleSubmit = async (_values: Record<string, string>) => {
    try {
      setLoading(true)
      // 保存地址
      Toast.show(isNew ? '添加成功' : '保存成功')
      navigate(-1)
    } catch (error) {
      Toast.show('保存失败')
    } finally {
      setLoading(false)
    }
  }

  return (
    <div className="min-h-screen bg-gray-50 pb-20">
      <Card className="m-2">
        <Form layout="vertical" onFinish={handleSubmit}>
          <Form.Item
            name="name"
            label="收货人"
            rules={[{ required: true, message: '请输入收货人姓名' }]}
          >
            <Input placeholder="请输入收货人姓名" />
          </Form.Item>

          <Form.Item
            name="phone"
            label="手机号"
            rules={[
              { required: true, message: '请输入手机号' },
              { pattern: /^1[3-9]\d{9}$/, message: '手机号格式不正确' },
            ]}
          >
            <Input placeholder="请输入手机号" />
          </Form.Item>

          <Form.Item
            name="province"
            label="省份"
            rules={[{ required: true, message: '请选择省份' }]}
          >
            <Input placeholder="请输入省份" />
          </Form.Item>

          <Form.Item
            name="city"
            label="城市"
            rules={[{ required: true, message: '请选择城市' }]}
          >
            <Input placeholder="请输入城市" />
          </Form.Item>

          <Form.Item
            name="district"
            label="区县"
            rules={[{ required: true, message: '请选择区县' }]}
          >
            <Input placeholder="请输入区县" />
          </Form.Item>

          <Form.Item
            name="detail"
            label="详细地址"
            rules={[{ required: true, message: '请输入详细地址' }]}
          >
            <Input placeholder="请输入详细地址" />
          </Form.Item>

          <Form.Item label="设为默认地址">
            <Switch checked={isDefault} onChange={setIsDefault} />
          </Form.Item>

          <div className="p-4">
            <Button
              block
              color="primary"
              size="large"
              loading={loading}
              type="submit"
            >
              保存
            </Button>
          </div>
        </Form>
      </Card>
    </div>
  )
}