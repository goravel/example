<script setup lang="ts">
import { DashboardWidgets } from '@admin/index'

import PostsStatCard from './PostsStatCard.vue'
</script>

<template>
  <DashboardWidgets.Fill>
    <PostsStatCard />
  </DashboardWidgets.Fill>
</template>
