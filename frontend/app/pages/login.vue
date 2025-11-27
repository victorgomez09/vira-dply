<script setup lang="ts">
import * as z from 'zod'
import type { FormSubmitEvent, AuthFormField } from '@nuxt/ui'

const {login} = useAuth();

const fields: AuthFormField[] = [
  {
    name: 'username',
    type: 'text',
    label: 'Username',
    placeholder: 'Enter your username',
    required: true
  }, {
    name: 'password',
    label: 'Password',
    type: 'password',
    placeholder: 'Enter your password',
    required: true
  }
]

const schema = z.object({
  username: z.string('Invalid username'),
  password: z.string('Password is required').min(8, 'Must be at least 8 characters')
})

type Schema = z.output<typeof schema>

async function onSubmit(payload: FormSubmitEvent<Schema>) {
  await login({username: payload.data.username, password: payload.data.password});
}
</script>

<template>
  <div class="flex flex-col items-center justify-center gap-4 p-4">
    <UPageCard variant="soft" class="w-full max-w-md">
      <UAuthForm :schema="schema" title="Login" description="Enter your credentials to access your account."
        icon="i-lucide-user" :fields="fields" @submit="onSubmit" />
    </UPageCard>
  </div>
</template>