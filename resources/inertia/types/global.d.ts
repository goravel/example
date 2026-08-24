/// <reference types="vite/client" />

declare module '*.vue' {
  import type { DefineComponent } from 'vue'
  const component: DefineComponent<object, object, unknown>
  export default component
}

// Shared props injected by the backend (provider Share/ShareFunc + middleware
// ShareSession). Augmenting PageProps gives `usePage().props` proper typing.
declare module '@inertiajs/core' {
  interface PageProps {
    appName: string
    timestamp: string
    auth: { user: Record<string, unknown> | null }
    flash: Record<string, string>
  }
}

export {}