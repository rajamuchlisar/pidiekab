import { ViteSSG } from 'vite-ssg'
import App from './App.vue'
import { routes } from './router/routes'

// vite-ssg automatically installs @unhead/vue and provides head context during SSG render
export const createApp = ViteSSG(
  App,
  {
    routes,
    base: '/',
  },
  () => {
    // client/server initialization hook
  }
)
