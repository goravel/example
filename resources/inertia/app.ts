import { createApp, h, type DefineComponent } from 'vue'
import { createInertiaApp } from '@inertiajs/vue3'

const pages = import.meta.glob<{ default: DefineComponent }>('./pages/**/*.vue')

createInertiaApp({
  progress: {
    // The delay after which the progress bar will appear, in milliseconds...
    delay: 250,
    // The color of the progress bar...
    color: '#29d',
    // Whether to include the default NProgress styles...
    includeCSS: true,
    // Whether the NProgress spinner will be shown...
    showSpinner: false,
  },
  title: (title) => (title ? `${title} · Goravel Inertia` : 'Goravel Inertia'),
  resolve: (name) => {
    const page = pages[`./pages/${name}.vue`]
    if (!page) {
      throw new Error(`Page not found: ${name}`)
    }
    // import.meta.glob (non-eager) yields lazy loaders, so invoke the loader and
    // resolve the component's default export for Inertia.
    return page().then((module) => module.default)
  },
  setup({ el, App, props, plugin }) {
    createApp({ render: () => h(App, props) })
      .use(plugin)
      .mount(el)
  },
})
