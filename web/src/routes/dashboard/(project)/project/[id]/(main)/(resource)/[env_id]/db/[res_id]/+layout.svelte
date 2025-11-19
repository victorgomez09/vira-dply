<script lang="ts">
	import { page } from '$app/state';
	let { children } = $props();
	import { NavBarBottom } from '$lib/components/ui/navbar';

	const projectId = $derived(page.params.id);
	const envId = $derived(page.params.env_id);
	const resId = $derived(page.params.res_id);

	const tabs = $derived([
		{
			name: 'Configuration',
			href: `/dashboard/project/${projectId}/${envId}/db/${resId}/overview`
		},
		{ name: 'Logs', href: `/dashboard/project/${projectId}/${envId}/db/${resId}/logs` },
		{ name: 'Terminal', href: `/dashboard/project/${projectId}/${envId}/db/${resId}/terminal` },
		{ name: 'DB Studio', href: `/dashboard/project/${projectId}/${envId}/db/${resId}/studio` }
	]);
</script>

<main>
	<NavBarBottom {tabs} />

	<div class="p-12 py-[80px]">
		{@render children()}
	</div>
</main>
