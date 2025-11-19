<script lang="ts">
	import { onMount } from 'svelte';
	import { goto } from '$app/navigation';
	import { authStore } from '$lib/stores/auth.svelte';
	import { NavBarTop } from '$lib/components/ui/navbar';

	let { children } = $props();

	onMount(() => {
		authStore.initialize();
		if (!authStore.isAuthenticated) {
			goto('/login');
		}
	});
</script>

<div class="min-h-screen">
	<NavBarTop />

	{@render children()}
</div>
