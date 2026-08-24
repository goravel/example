import { defineConfig, type PluginOption, type ViteDevServer } from 'vite'
import vue from '@vitejs/plugin-vue'
import { resolve } from 'path'
import { writeFileSync, rmSync } from 'fs'

// goravelHot writes public/hot with the dev-server URL while `vite` runs (and
// removes it on exit), Laravel-style. The Go backend reads that file and loads
// assets from the dev server (HMR) instead of the production manifest — so
// `go run .` needs no VITE_DEV_URL env var during development.
function goravelHot(): PluginOption {
  const hotFile = resolve(__dirname, 'public/hot')
  const clean = () => {
    try {
      rmSync(hotFile)
    } catch {
      // hot file already gone
    }
  }

  return {
    name: 'goravel-hot',
    apply: 'serve',
    configureServer(server: ViteDevServer) {
      const write = () => {
        const { https, host, port } = server.config.server
        const proto = https ? 'https' : 'http'
        const hostname = typeof host === 'string' ? host : 'localhost'
        writeFileSync(hotFile, `${proto}://${hostname}:${port ?? 5173}`)
      }

      server.httpServer?.once('listening', write)
      // Remove public/hot when the dev server stops. Deliberate improvement over
      // the installer stub, which called process.exit() from SIGINT/SIGTERM
      // handlers and bypassed Vite's graceful shutdown: Vite closes the HTTP
      // server on shutdown (firing the 'close' event below), and 'exit' covers
      // every other shutdown path.
      server.httpServer?.once('close', clean)
      process.on('exit', clean)
      // Vite installs a graceful SIGTERM handler but none for SIGINT, so Ctrl+C
      // would terminate without firing 'close'. Route SIGINT through Vite's
      // close() — which fires the httpServer 'close' event above — instead of
      // exiting directly.
      process.on('SIGINT', () => {
        server.close().finally(() => process.exit(0))
      })
    },
  }
}

export default defineConfig(({ command }) => ({
  // Assets are served from /build in production (see routes static mapping) but
  // from the dev server root in development.
  base: command === 'build' ? '/build/' : '/',
  plugins: [vue(), goravelHot()],
  root: '.',
  // Goravel serves ./public itself; outDir lives inside it, so disable Vite's
  // public dir copy to avoid the overlap warning and recursive copies.
  publicDir: false,
  build: {
    outDir: 'public/build',
    manifest: true,
    rollupOptions: {
      input: 'resources/inertia/app.ts',
    },
  },
  server: {
    port: 5173,
    strictPort: true,
    // The page is served by Goravel (:3000) while assets come from the Vite dev
    // server (:5173). Without an explicit origin, imported assets get
    // root-relative URLs that the browser resolves against :3000 → 404. Set
    // origin so Vite emits absolute dev-server URLs in development.
    origin: 'http://localhost:5173',
    cors: true,
  },
  resolve: {
    alias: {
      '@': resolve(__dirname, 'resources/inertia'),
    },
  },
}))
