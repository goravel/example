<template>
  <Head title="Timeline" />
  <div class="timeline">
    <header>
      <h1 class="page-title">Timeline</h1>
      <p class="lead">
        <strong>Merge prop</strong> demo — page {{ page }}. The client appends new
        items from a partial reload instead of replacing the list.
      </p>
    </header>

    <ul class="feed">
      <li v-for="item in items" :key="item.id" class="feed-item">
        <span class="dot">{{ item.id }}</span>
        <strong>{{ item.title }}</strong>
      </li>
    </ul>
  </div>
</template>

<script setup lang="ts">
import { Head } from '@inertiajs/vue3'

interface TimelineItem {
  id: number
  title: string
}

defineProps<{
  page: number
  items: TimelineItem[]
}>()
</script>

<style scoped>
.timeline {
  max-width: 720px;
  margin: 0 auto;
  padding: 2rem;
}
.page-title {
  margin: 0 0 0.5rem;
}
.lead {
  margin: 0 0 1.5rem;
  color: #555;
}
.feed {
  list-style: none;
  margin: 0;
  padding: 0;
  display: flex;
  flex-direction: column;
  gap: 0.75rem;
}
.feed-item {
  display: flex;
  align-items: center;
  gap: 1rem;
  border: 1px solid #e3e8f0;
  border-radius: 12px;
  padding: 1rem 1.2rem;
}
.dot {
  flex: none;
  width: 2.2rem;
  height: 2.2rem;
  display: grid;
  place-items: center;
  border-radius: 10px;
  font-weight: 700;
  color: #fff;
  background: linear-gradient(135deg, #3aa675, #6741d9);
}
</style>
