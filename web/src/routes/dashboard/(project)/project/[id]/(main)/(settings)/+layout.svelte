<script lang="ts">
	let { children } = $props();

	import { NavBarBottom } from '$lib/components/ui/navbar';

	const tabs = [
		{ name: 'Overview', href: 'dashboard' },
		{ name: 'Shared Variables', href: 'shared-vars' },
		{ name: 'Architecture', href: 'schema' },
		{ name: 'Settings', href: 'settings' }
	];
</script>

<NavBarBottom {tabs} />
<div class="p-6 pt-[104px]">
	{@render children()}
</div>
