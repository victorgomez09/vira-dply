<script lang="ts">
	import SearchForm from './ProjectSearch.svelte';
	import * as Breadcrumb from '$lib/components/ui/breadcrumb/index.js';
	import LogoFull from '$lib/components/logo/logo-full.svelte';
</script>

<header class="bg-[#1c1b23] text-white sticky top-0 z-50 flex w-full items-center">
	<div class="h-(--header-height) flex w-full items-center gap-6 px-4">
		<LogoFull fill="#fff" class="h-30 w-30" />
		<Breadcrumb.Root class="hidden sm:block">
			<Breadcrumb.List>
				<Breadcrumb.Item>
					<Breadcrumb.Link class="text-gray-300" href="/dashboard">Home</Breadcrumb.Link>
				</Breadcrumb.Item>
				<Breadcrumb.Separator>
					<span class="text-gray-500">/</span>
				</Breadcrumb.Separator>
				<Breadcrumb.Item>
					<Breadcrumb.Page class="text-gray-300">Data Fetching</Breadcrumb.Page>
				</Breadcrumb.Item>
			</Breadcrumb.List>
		</Breadcrumb.Root>

		<div class="flex items-center space-x-4">
			<!-- Search -->
			<!-- <div class="relative"> -->
			<!-- 	<Search class="absolute left-3 top-1/2 transform -translate-y-1/2 w-4 h-4 text-gray-400" /> -->
			<!-- 	<Input -->
			<!-- 		placeholder="Jump to or search..." -->
			<!-- 		class="pl-10 pr-8 bg-gray-800 border-gray-700 text-white placeholder-gray-400 w-64" -->
			<!-- 	/> -->
			<!-- 	<span class="absolute right-3 top-1/2 transform -translate-y-1/2 text-xs text-gray-500" -->
			<!-- 		>⌘ /</span -->
			<!-- 	> -->
			<!-- </div> -->
			<SearchForm class="w-full sm:ml-auto sm:w-auto" />
		</div>
	</div>
</header>
