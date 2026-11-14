import type { Component } from 'vue'

export interface MenuItem {
  label: string
  /** Relative to the admin's base path, starting with "/". */
  path: string
}

export interface MenuGroup {
  label: string
  items: MenuItem[]
}

/** Receives the menu the modules before it built and returns the one it wants. */
export type MenuFilter = (menu: MenuGroup[]) => MenuGroup[]

export interface AdminRoute {
  /** Relative to the admin's base path; ":name" segments become props. */
  path: string
  component: Component
}

export interface ModuleRegistration {
  id: string
  /** Non-negative counts from the start, negative from the end (-1 = last). Default 0. */
  order?: number
  menu?: MenuFilter
  routes?: AdminRoute[]
  /** Rendered for as long as the admin is, on every page: where a module fills slots such as DashboardWidgets. */
  fills?: Component
}

const registrations: ModuleRegistration[] = []
let sealed = false

export function registerModule(registration: ModuleRegistration): void {
  if (sealed) {
    throw new Error(`[admin] module "${registration.id}" registered after the admin mounted: call registerModule() at the top level of a dashboard script`)
  }
  if (registrations.some((module) => module.id === registration.id)) {
    throw new Error(`[admin] module "${registration.id}" is already registered`)
  }

  registrations.push(registration)
}

/** Closes the registry; main.ts calls it right before mounting. */
export function seal(): void {
  sealed = true
}

/** Same rule as modular.Ordered on the Go side: non-negative ascending, then negative ascending, ties keep registration order. */
export function modules(): ModuleRegistration[] {
  const rank = (order: number) => (order < 0 ? [1, order] : [0, order])

  return [...registrations].sort((a, b) => {
    const [ga, oa] = rank(a.order ?? 0)
    const [gb, ob] = rank(b.order ?? 0)

    return ga - gb || oa - ob
  })
}

export function menu(): MenuGroup[] {
  return modules().reduce<MenuGroup[]>((groups, module) => (module.menu ? module.menu(groups) : groups), [])
}

export function routes(): AdminRoute[] {
  return modules().flatMap((module) => module.routes ?? [])
}
