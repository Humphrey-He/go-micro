import { Table, TableProps } from 'antd'
import { TablePaginationConfig } from 'antd/es/table/interface'

interface BusinessTableProps<T> extends TableProps<T> {
  pagination?: false | TablePaginationConfig
  onPageChange?: (page: number, pageSize: number) => void
}

export function BusinessTable<T extends object>({
  pagination,
  onPageChange,
  ...props
}: BusinessTableProps<T>) {
  const handleTableChange: TableProps<T>['onChange'] = (pag) => {
    if (pagination !== false && onPageChange && 'current' in pag) {
      onPageChange(pag.current as number, pag.pageSize as number)
    }
  }

  return (
    <Table<T>
      {...props}
      pagination={
        pagination === false
          ? false
          : {
              showSizeChanger: true,
              showQuickJumper: true,
              showTotal: (total) => `共 ${total} 条`,
              ...pagination,
            }
      }
      onChange={handleTableChange}
    />
  )
}
