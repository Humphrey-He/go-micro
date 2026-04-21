import axios, { AxiosError, InternalAxiosRequestConfig } from 'axios'
import { message } from 'antd'
import { useAuthStore } from '@/stores/authStore'

const request = axios.create({
  baseURL: import.meta.env.VITE_API_BASE_URL,
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
  (error: AxiosError) => {
    return Promise.reject(error)
  }
)

request.interceptors.response.use(
  (response) => response.data,
  (error: AxiosError) => {
    const status = error.response?.status

    if (status === 401) {
      useAuthStore.getState().logout()
      message.error('登录已过期，请重新登录')
      window.location.href = '/login'
      return Promise.reject(error)
    }

    const data = error.response?.data as { message?: string } | undefined
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

export const get = <T>(url: string, config?: InternalAxiosRequestConfig): Promise<T> =>
  request.get<ApiResponse<T>>(url, config).then((res) => {
    if (res.data.code !== 0) {
      return Promise.reject(new Error(res.data.message))
    }
    return res.data.data as T
  })

export const post = <T>(url: string, data?: unknown, config?: InternalAxiosRequestConfig): Promise<T> =>
  request.post<ApiResponse<T>>(url, data, config).then((res) => {
    if (res.data.code !== 0) {
      return Promise.reject(new Error(res.data.message))
    }
    return res.data.data as T
  })
