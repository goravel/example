export interface Post {
  id: number
  title: string
  body: string
  created_at: string
}

export type ValidationErrors = Record<string, Record<string, string>>

export async function listPosts(): Promise<{ data: Post[]; total: number }> {
  const response = await fetch('/api/blog/posts', { headers: { Accept: 'application/json' } })
  if (!response.ok) throw new Error(`GET /api/blog/posts: ${response.status}`)

  return response.json()
}

/** Resolves to the validation errors when the server refuses the post, null when it is created. */
export async function createPost(input: { title: string; body: string }): Promise<ValidationErrors | null> {
  const response = await fetch('/api/blog/posts', {
    method: 'POST',
    headers: { Accept: 'application/json', 'Content-Type': 'application/json' },
    body: JSON.stringify(input),
  })
  if (response.status === 422) return (await response.json()).errors
  if (!response.ok) throw new Error(`POST /api/blog/posts: ${response.status}`)

  return null
}
