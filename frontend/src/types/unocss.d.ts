declare module 'unocss/vite' {
  import { Plugin } from 'vite'
  const plugin: () => Plugin
  export default plugin
}

declare module 'unocss' {
  export { defineConfig } from '@unocss/core'
  export const presetUno: any
  export const presetAttributify: any
  export const presetTypography: any
  export const presetIcons: any
  export const presetWebFonts: any
  export const presetTagify: any
}

declare module 'uno.css' {
  const content: string
  export default content
}

declare module '@unocss/reset/tailwind.css' {
  const content: string
  export default content
} 