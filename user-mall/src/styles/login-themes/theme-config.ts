// user-mall/src/styles/login-themes/theme-config.ts

export type LoginTheme = 'vibrant' | 'stable' | 'luxury'

export interface ThemeColors {
  primary: string
  secondary: string
  accent: string
  background: string
  backgroundGradient: string
  cardBg: string
  text: string
  textSecondary: string
  border: string
}

export interface ThemeConfig {
  id: LoginTheme
  name: string
  nameCn: string
  colors: ThemeColors
  cardRadius: number
  socialStyle: 'colorful' | 'subtle' | 'outline'
}

export const themes: Record<LoginTheme, ThemeConfig> = {
  vibrant: {
    id: 'vibrant',
    name: 'Youth/Vibrant',
    nameCn: '兴趣电商',
    colors: {
      primary: '#FF6B35',
      secondary: '#F7C948',
      accent: '#00D9C0',
      background: '#FFFFFF',
      backgroundGradient: 'linear-gradient(135deg, #FF6B35 0%, #FF4D94 100%)',
      cardBg: '#FFFFFF',
      text: '#1F2937',
      textSecondary: '#6B7280',
      border: '#E5E7EB',
    },
    cardRadius: 24,
    socialStyle: 'colorful',
  },
  stable: {
    id: 'stable',
    name: 'Amazon Stable',
    nameCn: '大厂稳重',
    colors: {
      primary: '#007185',
      secondary: '#FFA41C',
      accent: '#00A4A4',
      background: '#FFFFFF',
      backgroundGradient: 'linear-gradient(180deg, #F7F7F7 0%, #FFFFFF 100%)',
      cardBg: '#FFFFFF',
      text: '#0F1111',
      textSecondary: '#565959',
      border: '#E7E7E7',
    },
    cardRadius: 8,
    socialStyle: 'subtle',
  },
  luxury: {
    id: 'luxury',
    name: 'Luxury/Minimalist',
    nameCn: '轻奢简约',
    colors: {
      primary: '#1A1A1A',
      secondary: '#C9A96E',
      accent: '#E8E8E8',
      background: '#FAFAFA',
      backgroundGradient: 'linear-gradient(180deg, #FAFAFA 0%, #F5F5F5 100%)',
      cardBg: '#FFFFFF',
      text: '#1A1A1A',
      textSecondary: '#6B6B6B',
      border: '#E8E8E8',
    },
    cardRadius: 4,
    socialStyle: 'outline',
  },
}
