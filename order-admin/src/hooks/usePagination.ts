import { useState, useCallback } from 'react'
import { useSearchParams } from 'react-router-dom'

export interface UsePaginationOptions {
  defaultPage?: number
  defaultPageSize?: number
  pageSizeOptions?: number[]
}

export interface UsePaginationResult {
  page: number
  pageSize: number
  total: number
  setTotal: (total: number) => void
  handlePageChange: (page: number, pageSize: number) => void
  searchParams: URLSearchParams
  setSearchParams: (params: Record<string, string | number | undefined>) => void
}

export const usePagination = (
  options: UsePaginationOptions = {}
): UsePaginationResult => {
  const { defaultPage = 1, defaultPageSize = 20 } = options
  const [searchParams, setSearchParams] = useSearchParams()

  const page = parseInt(searchParams.get('page') || String(defaultPage), 10)
  const pageSize = parseInt(
    searchParams.get('pageSize') || String(defaultPageSize),
    10
  )

  const [total, setTotal] = useState(0)

  const handlePageChange = useCallback(
    (newPage: number, newPageSize: number) => {
      const params: Record<string, string | number | undefined> = {}
      searchParams.forEach((value, key) => {
        params[key] = value
      })

      if (newPage === defaultPage) {
        delete params.page
      } else {
        params.page = newPage
      }

      if (newPageSize === defaultPageSize) {
        delete params.pageSize
      } else {
        params.pageSize = newPageSize
      }

      setSearchParams(new URLSearchParams(params as Record<string, string>))
    },
    [searchParams, setSearchParams, defaultPage, defaultPageSize]
  )

  const updateSearchParams = useCallback(
    (params: Record<string, string | number | undefined>) => {
      const newParams: Record<string, string | number | undefined> = {}
      searchParams.forEach((value, key) => {
        newParams[key] = value
      })

      Object.entries(params).forEach(([key, value]) => {
        if (value === undefined || value === '' || value === null) {
          delete newParams[key]
        } else {
          newParams[key] = value
        }
      })

      if (newParams.page === undefined) {
        newParams.page = defaultPage
      }
      if (newParams.pageSize === undefined) {
        newParams.pageSize = defaultPageSize
      }

      setSearchParams(new URLSearchParams(newParams as Record<string, string>))
    },
    [searchParams, setSearchParams, defaultPage, defaultPageSize]
  )

  return {
    page,
    pageSize,
    total,
    setTotal,
    handlePageChange,
    searchParams,
    setSearchParams: updateSearchParams,
  }
}
