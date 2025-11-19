<script lang="ts">
	import Logo from '$lib/components/logo/logo.svelte';
	import { getIsScrolled } from './navbar-scroll.svelte';
	import { cn } from '$lib/utils';
	import UserMenu from './nav-top-menu.svelte';
	import NavbarBreadcrumb from './navbar-breadcrumb.svelte';
</script>

<header class="fixed top-0 left-0 right-0 z-50 bg-card">
	<div class="px-6 relative">
		<Logo
			class={cn(
				'h-[24px] absolute left-6 hidden md:flex items-center gap-4 transition-all duration-300 ease-in-out z-10',
				getIsScrolled() ? 'translate-y-[12px] scale-[0.8]' : 'translate-y-[18px]'
			)}
		/>

		<div
			class={cn(
				'flex items-center gap-4 transition-all duration-300 ease-in-out overflow-hidden',
				getIsScrolled() ? 'h-0 opacity-0' : 'h-14 opacity-100'
			)}
		>
			<div class="w-[25px] flex-shrink-0 hidden md:block"></div>

			<NavbarBreadcrumb />

			<div class="ml-auto">
				<UserMenu />
			</div>
		</div>
	</div>
</header>
