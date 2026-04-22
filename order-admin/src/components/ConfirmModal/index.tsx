import React from 'react'
import { Modal, ModalProps } from 'antd'
import { ExclamationCircleOutlined } from '@ant-design/icons'

interface ConfirmModalProps extends ModalProps {
  title?: string
  content?: string
  onConfirm?: () => void
  onCancel?: () => void
}

export const ConfirmModal: React.FC<ConfirmModalProps> = ({
  title = '确认操作',
  content,
  onConfirm,
  onCancel,
  ...props
}) => {
  return (
    <Modal
      title={
        <span>
          <ExclamationCircleOutlined style={{ color: '#faad14', marginRight: 8 }} />
          {title}
        </span>
      }
      okText="确认"
      cancelText="取消"
      okButtonProps={{ danger: true }}
      onOk={onConfirm}
      onCancel={onCancel}
      {...props}
    >
      {content && <p>{content}</p>}
    </Modal>
  )
}
