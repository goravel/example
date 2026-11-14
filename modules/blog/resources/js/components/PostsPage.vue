<script setup lang="ts">
import { onMounted, ref } from 'vue'

import { AdminLink, HeaderActions } from '@admin/index'

import { listPosts, type Post } from '../api'

const posts = ref<Post[]>([])

onMounted(async () => {
  posts.value = (await listPosts()).data
})
</script>

<template>
  <!-- Shown in the layout's header for as long as this page is. -->
  <HeaderActions.Fill>
    <AdminLink to="/blog/new" class="admin-button">New post</AdminLink>
  </HeaderActions.Fill>

  <table class="admin-table">
    <thead>
      <tr><th>#</th><th>Title</th><th>Created</th></tr>
    </thead>
    <tbody>
      <tr v-for="post in posts" :key="post.id">
        <td>{{ post.id }}</td>
        <td>{{ post.title }}</td>
        <td>{{ new Date(post.created_at).toLocaleString() }}</td>
      </tr>
    </tbody>
  </table>
</template>
