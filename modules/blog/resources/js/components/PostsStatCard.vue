<script setup lang="ts">
import { onMounted, ref } from 'vue'

import { AdminLink } from '@admin/index'

import { listPosts } from '../api'

// Mounted only while the dashboard shows (a Fill renders nothing without its outlet), so it loads there only.
const total = ref<number | null>(null)

onMounted(async () => {
  total.value = (await listPosts()).total
})
</script>

<template>
  <div class="admin-card">
    <h3>Blog posts</h3>
    <div class="admin-stat">{{ total ?? '…' }}</div>
    <AdminLink to="/blog">View all</AdminLink>
  </div>
</template>
