import { get, post } from '@/api/request'

// 获取收货地址
export interface Address {
  id: string
  receiver: string
  phone: string
  province: string
  city: string
  district: string
  detail: string
  tag: 'home' | 'company' | 'school' | 'other'
  is_default: boolean
}

export const getAddressList = () =>
  get<{ addresses: Address[] }>('/user/address')

// 新增地址
export interface AddAddressParams {
  receiver: string
  phone: string
  province: string
  city: string
  district: string
  detail: string
  tag?: 'home' | 'company' | 'school' | 'other'
  is_default?: boolean
}

export const addAddress = (params: AddAddressParams) =>
  post<{ id: string } & AddAddressParams>('/user/address', params)

// 更新地址
export const updateAddress = (id: string, params: AddAddressParams) =>
  post<Address>(`/user/address/${id}`, params)

// 删除地址
export const deleteAddress = (id: string) =>
  post<void>(`/user/address/${id}/delete`)

// 设置默认地址
export const setDefaultAddress = (id: string) =>
  post<void>(`/user/address/${id}/default`)