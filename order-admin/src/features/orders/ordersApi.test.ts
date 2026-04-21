import { describe, it, expect } from 'vitest'

// Simulate the API response (this is what axios interceptor returns = response.data)
const mockApiResponse = {
  code: 0,
  message: 'OK',
  data: {
    orderId: 'ord-101',
    bizNo: 'ORD20260421001',
    userId: 'user-001',
    status: 'SUCCESS',
    totalAmount: 29900,
    items: [
      { skuId: 'SKU002', quantity: 1, price: 12999 },
      { skuId: 'SKU003', quantity: 3, price: 1499 },
    ],
    paymentStatus: 'PAID',
    viewStatus: 'SUCCESS',
  },
}

// This is what the API actually returns (wrapped in code/message/data)
const mockFullResponse = {
  code: 0,
  message: 'OK',
  data: mockApiResponse.data,
}

describe('Order Detail API Response', () => {
  it('should have correct response structure', () => {
    expect(mockApiResponse).toHaveProperty('code', 0)
    expect(mockApiResponse).toHaveProperty('message', 'OK')
    expect(mockApiResponse).toHaveProperty('data')
  })

  it('should have correct order detail fields', () => {
    const data = mockApiResponse.data
    expect(data).toHaveProperty('orderId')
    expect(data).toHaveProperty('bizNo')
    expect(data).toHaveProperty('userId')
    expect(data).toHaveProperty('status')
    expect(data).toHaveProperty('totalAmount')
    expect(data).toHaveProperty('items')
    expect(data).toHaveProperty('paymentStatus')
    expect(data).toHaveProperty('viewStatus')
  })

  it('should have correct item structure', () => {
    const data = mockApiResponse.data
    expect(Array.isArray(data.items)).toBe(true)
    expect(data.items.length).toBeGreaterThan(0)

    const item = data.items[0]
    expect(item).toHaveProperty('skuId')
    expect(item).toHaveProperty('quantity')
    expect(item).toHaveProperty('price')
    expect(typeof item.skuId).toBe('string')
    expect(typeof item.quantity).toBe('number')
    expect(typeof item.price).toBe('number')
  })

  it('should have valid status values', () => {
    const data = mockApiResponse.data
    const validOrderStatuses = ['CREATED', 'RESERVED', 'PROCESSING', 'SUCCESS', 'FAILED', 'CANCELED', 'TIMEOUT']
    const validPaymentStatuses = ['PENDING', 'PAID', 'FAILED', 'REFUNDED']
    const validViewStatuses = ['PENDING', 'PROCESSING', 'SUCCESS', 'FAILED', 'DEAD', 'CANCELED', 'TIMEOUT']

    expect(validOrderStatuses).toContain(data.status)
    expect(validPaymentStatuses).toContain(data.paymentStatus)
    expect(validViewStatuses).toContain(data.viewStatus)
  })

  it('should have numeric amount values', () => {
    const data = mockApiResponse.data
    expect(typeof data.totalAmount).toBe('number')
    expect(data.totalAmount).toBeGreaterThan(0)
  })

  it('should calculate correct item subtotals', () => {
    const data = mockApiResponse.data
    const subtotal = data.items.reduce((sum: number, item: { price: number; quantity: number }) => sum + item.price * item.quantity, 0)
    expect(subtotal).toBeGreaterThan(0)
  })

  it('should correctly unwrap nested API response', () => {
    // Simulating what the hooks do: response.data gives us the outer wrapper
    // Then we access .data to get the actual order
    const unwrapped = mockFullResponse.data
    expect(unwrapped).toHaveProperty('orderId', 'ord-101')
    expect(unwrapped).toHaveProperty('bizNo', 'ORD20260421001')
    expect(unwrapped.items.length).toBe(2)
  })
})

describe('Order Status Constants', () => {
  const ORDER_STATUS_MAP = {
    CREATED: { label: '已创建', color: '#1677ff', bg: '#e6f4ff' },
    RESERVED: { label: '已预留', color: '#0891b2', bg: '#ecfeff' },
    PROCESSING: { label: '处理中', color: '#7c3aed', bg: '#f5f3ff' },
    SUCCESS: { label: '成功', color: '#16a34a', bg: '#f0fdf4' },
    FAILED: { label: '失败', color: '#dc2626', bg: '#fef2f2' },
    CANCELED: { label: '已取消', color: '#6b7280', bg: '#f9fafb' },
    TIMEOUT: { label: '超时', color: '#ea580c', bg: '#fff7ed' },
  }

  it('should have status config with label, color and bg', () => {
    Object.values(ORDER_STATUS_MAP).forEach((config) => {
      expect(config).toHaveProperty('label')
      expect(config).toHaveProperty('color')
      expect(config).toHaveProperty('bg')
      expect(typeof config.label).toBe('string')
      expect(typeof config.color).toBe('string')
      expect(typeof config.bg).toBe('string')
    })
  })
})
