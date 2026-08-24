<template>
  <Head title="Message" />
  <div class="message">
    <header>
      <h1 class="page-title">Message</h1>
      <p class="lead">
        <strong>Flash + Validation</strong> demo — POST → validate → redirect back.
        Errors arrive in <code>props.errors</code>, success in
        <code>props.flash.success</code>.
      </p>
    </header>

    <p v-if="flash?.success" class="flash">{{ flash.success }}</p>
    <ul v-if="errors && Object.keys(errors).length" class="errors">
      <li v-for="(msg, field) in errors" :key="field">{{ field }}: {{ msg }}</li>
    </ul>

    <form class="form" method="post" action="/inertia/contact">
      <label>
        <span class="label">Email</span>
        <input type="email" name="email" placeholder="ada@example.com" />
      </label>
      <button type="submit" class="btn">Send message</button>
    </form>
  </div>
</template>

<script setup lang="ts">
import { Head } from '@inertiajs/vue3'

defineProps<{
  flash?: { success?: string }
  errors?: Record<string, string>
}>()
</script>

<style scoped>
.message {
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
.flash {
  color: #1d7a3f;
  font-weight: 600;
}
.errors {
  color: #b02a2a;
}
.form {
  display: flex;
  flex-direction: column;
  gap: 1.1rem;
}
label {
  display: flex;
  flex-direction: column;
  gap: 0.4rem;
}
.label {
  font-size: 0.85rem;
  font-weight: 600;
}
input {
  font: inherit;
  padding: 0.65rem 0.8rem;
  border: 1px solid #e3e8f0;
  border-radius: 10px;
}
.btn {
  align-self: flex-start;
  font: inherit;
  font-weight: 600;
  padding: 0.6rem 1.2rem;
  border: 0;
  border-radius: 10px;
  color: #fff;
  background: #3aa675;
  cursor: pointer;
}
</style>
