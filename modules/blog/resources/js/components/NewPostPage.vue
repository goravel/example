<script setup lang="ts">
import { reactive, ref } from 'vue'

import { navigate } from '@admin/index'

import { createPost, type ValidationErrors } from '../api'

const form = reactive({ title: '', body: '' })
const errors = ref<ValidationErrors>({})

async function submit() {
  const refused = await createPost(form)
  if (refused) {
    errors.value = refused
    return
  }

  navigate('/blog')
}
</script>

<template>
  <form class="admin-form" @submit.prevent="submit">
    <label>
      Title
      <input v-model="form.title" name="title" />
    </label>
    <span v-for="message in errors.title" :key="message" class="admin-error">{{ message }}</span>

    <label>
      Body
      <textarea v-model="form.body" name="body" rows="8" />
    </label>
    <span v-for="message in errors.body" :key="message" class="admin-error">{{ message }}</span>

    <div><button class="admin-button" type="submit">Publish</button></div>
  </form>
</template>
