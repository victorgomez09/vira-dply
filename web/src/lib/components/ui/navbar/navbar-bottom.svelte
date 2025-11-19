<script lang="ts">
	import { onMount, tick } from 'svelte';
	import { page } from '$app/state';
	import { cn } from '$lib/utils';
	import { getIsScrolled } from './navbar-scroll.svelte';

	interface Tab {
		name: string;
		href: string;
	}

	let { tabs }: { tabs: Tab[] } = $props();

	let tabsRef = $state<(HTMLAnchorElement | null)[]>([]);
	let tabsContainerRef = $state<HTMLDivElement | null>(null);
	let indicatorStyle = $state({ left: 0, width: 0 });
	let direction = $state<'left' | 'right'>('right');

	const pathname = $derived(page.url.pathname);

	const activeTab = $derived(
		tabs.find((tab) => pathname === tab.href || pathname.startsWith(`${tab.href}/`))?.name ||
			tabs[0]?.name
	);

	async function updateIndicator() {
		await tick();
		const activeIndex = tabs.findIndex((tab) => tab.name === activeTab);
		const activeElement = tabsRef[activeIndex];

		console.log('active index', activeIndex);

		if (activeElement && tabsContainerRef) {
			const containerRect = tabsContainerRef.getBoundingClientRect();
			const elementRect = activeElement.getBoundingClientRect();

			indicatorStyle.left = elementRect.left - containerRect.left + tabsContainerRef.scrollLeft;
			indicatorStyle.width = activeElement.offsetWidth;
		}
	}

	function handleTabClick(tabName: string) {
		const currentIndex = tabs.findIndex((tab) => tab.name === activeTab);
		const newIndex = tabs.findIndex((tab) => tab.name === tabName);
		direction = newIndex > currentIndex ? 'right' : 'left';
	}

	$effect(() => {
		activeTab;
		updateIndicator();
	});

	onMount(() => {
		updateIndicator();
	});
</script>

<div
	class={cn(
		'sticky left-0 right-0 z-40 bg-card border-b border-input transition-all duration-300',
		getIsScrolled() ? 'top-0' : 'top-14'
	)}
>
	<div class="mx-auto px-4 md:px-2">
		<div class={cn('flex items-center justify-between transition-all duration-300 h-12')}>
			<div class="flex items-center gap-4 flex-1 min-w-0 h-full">
				<div
					class={cn(
						'hidden md:flex items-center gap-4 transition-all duration-300 ease-in-out flex-shrink-0',
						getIsScrolled() ? 'w-[35px]' : 'md:w-0 w-5'
					)}
				></div>
				<div
					bind:this={tabsContainerRef}
					class="relative flex-1 overflow-x-auto scrollbar-hide min-w-0 h-full"
					style="scrollbar-width: none; -ms-overflow-style: none;"
				>
					<div class="relative flex items-center gap-1 w-max h-full">
						{#each tabs as tab, index}
							<a
								href={tab.href}
								bind:this={tabsRef[index]}
								onclick={() => handleTabClick(tab.name)}
								class={cn(
									'relative px-3 py-2 text-sm font-medium transition-colors rounded-md whitespace-nowrap',
									activeTab === tab.name
										? 'text-foreground'
										: 'text-muted-foreground hover:text-foreground hover:bg-muted/50'
								)}
							>
								{tab.name}
							</a>
						{/each}
						<div
							class="absolute bottom-[-1px] h-[3px] bg-foreground transition-all duration-300 ease-out"
							style="left: {indicatorStyle.left}px; width: {indicatorStyle.width}px; transform-origin: {direction ===
							'right'
								? 'left'
								: 'right'};"
						></div>
					</div>
				</div>
			</div>
		</div>
	</div>
</div>

<style>
	.scrollbar-hide::-webkit-scrollbar {
		display: none;
	}
</style>
