import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'
import { resolve } from 'path'
import goravel from 'goravel-vite-plugin'

export default defineConfig({
  // goravel-vite-plugin is the framework's counterpart. It asks the application
  // where to build (config/vite.go) and which entries the views reference
  // (`{{ vite ... }}` through `artisan vite:inputs`), writes the build and its
  // manifest where the framework reads them, and while `vite` runs it writes
  // public/hot with the dev-server URL, removing it again on exit. None of that
  // is repeated here.
  plugins: [
    vue(),
    // Go templates are parsed when the server boots, so reload the page when a
    // view, a language file or a route changes and air rebuilds.
    goravel({ refresh: true }),
  ],
  server: {
    port: 5173,
    strictPort: true,
  },
  resolve: {
    // The example keeps its frontend in resources/inertia; the plugin's default
    // is resources/js.
    alias: {
      '@': resolve(import.meta.dirname, 'resources/inertia'),
    },
  },
})
