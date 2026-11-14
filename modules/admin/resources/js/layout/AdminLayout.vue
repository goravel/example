<script setup lang="ts">
import { computed } from 'vue'

import { appName } from '../lib/boot'
import { menu, modules, routes } from '../lib/registry'
import { activePath, currentPath, match } from '../lib/router'
import AdminLink from './AdminLink.vue'
import { DashboardWidgets, HeaderActions } from './slots'

// The registry is sealed before mount, so these are read once.
const groups = menu()
const fills = modules().flatMap((module) => (module.fills ? [module.fills] : []))
const allRoutes = routes()

const matched = computed(() => match(allRoutes, currentPath.value))
const active = computed(() => activePath(currentPath.value, groups))
const title = computed(() => groups.flatMap((group) => group.items).find((item) => item.path === active.value)?.label ?? appName)
</script>

<template>
  <DashboardWidgets.Provider>
    <HeaderActions.Provider>
      <div class="admin">
        <aside class="admin-sidebar">
          <strong class="admin-brand">{{ appName }}</strong>
          <nav v-for="group in groups" :key="group.label">
            <h2>{{ group.label }}</h2>
            <AdminLink v-for="item in group.items" :key="item.path" :to="item.path" :class="{ active: item.path === active }">
              {{ item.label }}
            </AdminLink>
          </nav>
        </aside>

        <div class="admin-main">
          <header class="admin-header">
            <h1>{{ title }}</h1>
            <HeaderActions.Outlet class="admin-header-actions" />
          </header>

          <main>
            <component :is="matched.route.component" v-if="matched" :key="currentPath" v-bind="matched.params" />
            <p v-else>Page not found.</p>
          </main>
        </div>
      </div>

      <component :is="fill" v-for="(fill, i) in fills" :key="i" />
    </HeaderActions.Provider>
  </DashboardWidgets.Provider>
</template>
