// The blog's dashboard script: listed in DashboardScripts() (and Frontend().Entries)
// of modules/blog/module.go, so the admin page loads it before its main.ts.
const card = document.createElement('div')
card.className = 'admin-card'
card.innerHTML = '<h3>Blog posts</h3><div class="admin-stat">…</div><a href="/blog">View all</a>'
document.getElementById('widgets')!.append(card)

fetch('/api/blog/posts', { headers: { Accept: 'application/json' } })
  .then((response) => response.json() as Promise<{ total: number }>)
  .then(({ total }) => {
    card.querySelector('.admin-stat')!.textContent = String(total)
  })
