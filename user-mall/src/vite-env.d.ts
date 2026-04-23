/// <reference types="vite/client" />

interface ImportMetaEnv {
  readonly VITE_WECHAT_APP_ID: string
  readonly VITE_GOOGLE_CLIENT_ID: string
  readonly VITE_APPLE_CLIENT_ID: string
}

interface ImportMeta {
  readonly env: ImportMetaEnv
}
