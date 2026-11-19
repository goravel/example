// The blog's dashboard script: listed in DashboardScripts() (and ViteEntries())
// of modules/blog/module.go, loaded by the admin page before the admin mounts.
import { registerModule } from '@admin/index'

import BlogFills from './components/BlogFills.vue'
import NewPostPage from './components/NewPostPage.vue'
import PostsPage from './components/PostsPage.vue'

registerModule({
  id: 'blog',
  menu: (menu) => [
    ...menu,
    {
      label: 'Blog',
      items: [
        { label: 'Posts', path: '/blog' },
        { label: 'New post', path: '/blog/new' },
      ],
    },
  ],
  routes: [
    { path: '/blog', component: PostsPage },
    { path: '/blog/new', component: NewPostPage },
  ],
  // The stat card on the dashboard: rendered on every page, shown only where the outlet is.
  fills: BlogFills,
})
