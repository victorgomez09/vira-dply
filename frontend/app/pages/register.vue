<script setup lang="ts">
import * as z from 'zod'
import type { FormSubmitEvent, AuthFormField } from '@nuxt/ui'

const fields: AuthFormField[] = [
    {
        name: 'username',
        type: 'text',
        label: 'Username',
        placeholder: 'Enter your username',
        required: true
    },
    {
        name: 'email',
        type: 'email',
        label: 'Email',
        placeholder: 'Enter your email',
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
    email: z.email('Invalid email'),
    password: z.string('Password is required').min(6, 'Must be at least 6 characters')
})

type Schema = z.output<typeof schema>

async function onSubmit(payload: FormSubmitEvent<Schema>) {
    payload.preventDefault()
    console.log('Submitted', payload.data)
    const res = await $fetch(`/api/auth/register`, {
        method: 'POST',
        body: payload.data
    })
    console.log('res', res)
}
</script>

<template>
    <div class="flex flex-col items-center justify-center gap-4 p-4">
        <UPageCard variant="soft" class="w-full max-w-md">
            <UAuthForm :schema="schema" title="Register" description="Create new user to continue."
                icon="i-lucide-user" :fields="fields" @submit.prevent="onSubmit" />
        </UPageCard>
    </div>
</template>