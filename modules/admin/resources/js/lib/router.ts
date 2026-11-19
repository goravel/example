import { ref } from 'vue'

import { basePath } from './boot'
import type { AdminRoute, MenuGroup } from './registry'

/** The path below the admin's base path: "/admin/blog" is "/blog". */
function relative(pathname: string): string {
  const rest = pathname === basePath ? '' : pathname.startsWith(`${basePath}/`) ? pathname.slice(basePath.length) : ''

  return rest.replace(/\/+$/, '') || '/'
}

export const currentPath = ref(relative(window.location.pathname))

window.addEventListener('popstate', () => {
  currentPath.value = relative(window.location.pathname)
})

export function href(path: string): string {
  return path === '/' ? basePath : basePath + path
}

export function navigate(path: string): void {
  window.history.pushState(null, '', href(path))
  currentPath.value = path
}

export function match(routes: AdminRoute[], path: string): { route: AdminRoute; params: Record<string, string> } | null {
  const segments = path.split('/').filter(Boolean)

  for (const route of routes) {
    const pattern = route.path.split('/').filter(Boolean)
    if (pattern.length !== segments.length) continue

    const params: Record<string, string> = {}
    const ok = pattern.every((part, i) => {
      if (part.startsWith(':')) {
        params[part.slice(1)] = decodeURIComponent(segments[i])
        return true
      }

      return part === segments[i]
    })
    if (ok) return { route, params }
  }

  return null
}

/** The deepest menu item covering the page, so "/blog/new" lights up "New post", not "Posts". */
export function activePath(path: string, groups: MenuGroup[]): string | null {
  let active: string | null = null
  for (const item of groups.flatMap((group) => group.items)) {
    const covers = item.path === '/' ? path === '/' : path === item.path || path.startsWith(`${item.path}/`)
    if (covers && (!active || item.path.length > active.length)) active = item.path
  }

  return active
}
