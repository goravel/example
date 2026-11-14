<script setup lang="ts">
import { href, navigate } from '../lib/router'

const props = defineProps<{ to: string }>()

// Client-side navigation for a plain click; a new tab or window still gets the real URL.
function onClick(event: MouseEvent) {
  if (event.button !== 0 || event.metaKey || event.ctrlKey || event.shiftKey || event.altKey) return

  event.preventDefault()
  navigate(props.to)
}
</script>

<template>
  <a :href="href(to)" @click="onClick"><slot /></a>
</template>
