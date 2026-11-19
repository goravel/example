// What a dashboard script imports as '@admin/index'. Its first import is the
// admin's own registration, so the dashboard comes before every module.
import DashboardPage from './layout/DashboardPage.vue'
import { registerModule } from './lib/registry'

registerModule({
  id: 'dashboard',
  menu: (menu) => [{ label: 'General', items: [{ label: 'Dashboard', path: '/' }] }, ...menu],
  routes: [{ path: '/', component: DashboardPage }],
})

export { default as AdminLink } from './layout/AdminLink.vue'
export { DashboardWidgets, HeaderActions } from './layout/slots'
export { basePath } from './lib/boot'
export { registerModule, type AdminRoute, type MenuFilter, type MenuGroup, type MenuItem, type ModuleRegistration } from './lib/registry'
export { navigate } from './lib/router'
export { createSlot, type Slot } from './lib/slot'
