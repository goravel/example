<template>
  <Head title="Welcome" />
  <div class="welcome">
    <section class="hero">
      <span class="badge">Goravel + Inertia</span>
      <h1 class="hero-title">{{ message }} + Vue 3</h1>
      <p class="lead">
        A Goravel adapter for Inertia.js — server-driven SPAs with Vue 3, no API
        layer. This page demos a <strong>deferred prop</strong>: the stats below
        stream in after the initial render.
      </p>
    </section>

    <!-- Deferred prop demo: "stats" arrives after the first paint. -->
    <Deferred data="stats">
      <template #fallback>
        <div class="stats">
          <div v-for="n in 3" :key="n" class="stat skeleton"><span class="bar" /></div>
        </div>
      </template>

      <div class="stats">
        <div v-for="(value, key) in stats" :key="key" class="stat">
          <span class="stat-value">{{ value }}</span>
          <span class="stat-label">{{ key }}</span>
        </div>
      </div>
    </Deferred>
  </div>
</template>

<script setup lang="ts">
import { Deferred, Head } from '@inertiajs/vue3'

defineProps<{
  message: string
  stats?: Record<string, string | number>
}>()
</script>

<style scoped>
.welcome {
  display: flex;
  flex-direction: column;
  gap: 1.5rem;
  max-width: 720px;
  margin: 0 auto;
  padding: 2rem;
}
.hero {
  display: flex;
  flex-direction: column;
  align-items: flex-start;
  gap: 1rem;
}
.badge {
  font-size: 0.72rem;
  font-weight: 700;
  letter-spacing: 0.04em;
  text-transform: uppercase;
  color: #6741d9;
  background: #f1ecfe;
  padding: 0.3rem 0.6rem;
  border-radius: 999px;
}
.hero-title {
  margin: 0;
  font-size: 2.4rem;
  font-weight: 800;
  letter-spacing: -0.03em;
}
.lead {
  margin: 0;
  line-height: 1.6;
  color: #555;
}
.stats {
  display: grid;
  grid-template-columns: repeat(3, 1fr);
  gap: 1rem;
}
.stat {
  border: 1px solid #e3e8f0;
  border-radius: 14px;
  padding: 1.3rem;
  display: flex;
  flex-direction: column;
  gap: 0.35rem;
  min-height: 92px;
  justify-content: center;
}
.stat-value {
  font-size: 1.8rem;
  font-weight: 800;
}
.stat-label {
  font-size: 0.75rem;
  text-transform: uppercase;
  letter-spacing: 0.05em;
  color: #888;
}
.skeleton .bar {
  display: block;
  height: 1.8rem;
  width: 70%;
  border-radius: 6px;
  background: linear-gradient(90deg, #eef1f7 25%, #e3e8f0 37%, #eef1f7 63%);
  background-size: 400% 100%;
  animation: shimmer 1.3s ease infinite;
}
@keyframes shimmer {
  0% {
    background-position: 100% 0;
  }
  100% {
    background-position: -100% 0;
  }
}
@media (max-width: 640px) {
  .stats {
    grid-template-columns: 1fr;
  }
  .hero-title {
    font-size: 1.9rem;
  }
}
</style>
