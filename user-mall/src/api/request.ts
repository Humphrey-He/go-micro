import axios, { AxiosError, InternalAxiosRequestConfig } from 'axios'
import { message } from 'antd'
import { useAuthStore } from '@/stores/authStore'

const request = axios.create({
  baseURL: '/api/v1',
  timeout: 10000,
})

request.interceptors.request.use(
  (config: InternalAxiosRequestConfig) => {
    const token = useAuthStore.getState().token
    if (token && config.headers) {
      config.headers.Authorization = `Bearer ${token}`
    }
    return config
  },
  (error: AxiosError) => Promise.reject(error)
)

request.interceptors.response.use(
  (response) => response.data,
  (error: AxiosError) => {
    const status = error.response?.status
    const data = error.response?.data as { message?: string; code?: number } | undefined

    if (status === 401 || data?.code === 40101) {
      useAuthStore.getState().logout()
      message.error('登录已过期，请重新登录')
      window.location.href = '/login'
      return Promise.reject(error)
    }

    if (data?.message) {
      message.error(data.message)
    } else if (status === 500) {
      message.error('服务器错误，请稍后重试')
    } else if (!status) {
      message.error('网络错误，请检查网络连接')
    }

    return Promise.reject(error)
  }
)

export default request

export interface ApiResponse<T> {
  code: number
  message: string
  data?: T
}

export const unwrapResponse = <T>(res: ApiResponse<T>): T => {
  if (res.code !== 0) {
    throw new Error(res.message)
  }
  return res.data as T
}

export const get = <T>(url: string, config?: Record<string, unknown>): Promise<T> =>
  request.get<unknown, ApiResponse<T>>(url, config as unknown as InternalAxiosRequestConfig).then(unwrapResponse)

export const post = <T>(url: string, data?: unknown, config?: Record<string, unknown>): Promise<T> =>
  request.post<unknown, ApiResponse<T>>(url, data, config as unknown as InternalAxiosRequestConfig).then(unwrapResponse)

export const put = <T>(url: string, data?: unknown, config?: Record<string, unknown>): Promise<T> =>
  request.put<unknown, ApiResponse<T>>(url, data, config as unknown as InternalAxiosRequestConfig).then(unwrapResponse)

export const del = <T>(url: string, config?: Record<string, unknown>): Promise<T> =>
  request.delete<unknown, ApiResponse<T>>(url, config as unknown as InternalAxiosRequestConfig).then(unwrapResponse)