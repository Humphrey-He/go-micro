export interface Coupon {
  id: string
  name: string
  type: 'CASH' | 'DISCOUNT'
  value: number
  min_amount: number
  status: 'AVAILABLE' | 'USED' | 'EXPIRED'
  valid_from: string
  valid_until: string
  received?: boolean
}

export const getCoupons = async (_status?: string): Promise<Coupon[]> => {
  return []
}

export const receiveCoupon = async (_id: string): Promise<void> => {
  return
}