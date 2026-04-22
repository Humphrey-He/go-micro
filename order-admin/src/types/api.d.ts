export interface ApiResponse<T = unknown> {
  code: number
  message: string
  data?: T
}

export interface PageResult<T> {
  list: T[]
  total: number
  page: number
  pageSize: number
}

export interface PaginationParams {
  page: number
  pageSize: number
}

export interface PaginationResult {
  total: number
  page: number
  pageSize: number
}
