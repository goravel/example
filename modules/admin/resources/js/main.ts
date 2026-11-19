// The admin's own entry. The page loads every module's dashboard script before
// this one, so all modules have registered by now.
import { createApp } from 'vue'

import './index'
import './admin.css'
import AdminLayout from './layout/AdminLayout.vue'
import { seal } from './lib/registry'

seal()

createApp(AdminLayout).mount('#admin')
