import React from 'react'
import { Drawer, DrawerProps } from 'antd'

interface DetailDrawerProps extends DrawerProps {}

export const DetailDrawer: React.FC<DetailDrawerProps> = ({
  children,
  ...props
}) => {
  return (
    <Drawer
      width={600}
      placement="right"
      closable
      destroyOnClose
      {...props}
    >
      {children}
    </Drawer>
  )
}
