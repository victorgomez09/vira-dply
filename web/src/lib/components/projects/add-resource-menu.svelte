<script lang="ts">
	import { Button } from '$lib/components/ui/button';
	import {
		DropdownMenu,
		DropdownMenuContent,
		DropdownMenuItem,
		DropdownMenuTrigger,
		DropdownMenuSeparator
	} from '$lib/components/ui/dropdown-menu';
	import { Plus, Server, Database, FileStack, LayoutGrid } from 'lucide-svelte';

	interface Props {
		onAddApplication: () => void;
		onAddDatabase: () => void;
		onAddCompose?: () => void;
		onAddTemplate?: () => void;
	}

	let { onAddApplication, onAddDatabase, onAddCompose, onAddTemplate }: Props = $props();
</script>

<DropdownMenu>
	<DropdownMenuTrigger>
		<Button variant="default">
			<Plus class="h-4 w-4" />
			Add Resource
		</Button>
	</DropdownMenuTrigger>
	<DropdownMenuContent align="end" class="w-56">
		<DropdownMenuItem onclick={onAddApplication}>
			<Server class="mr-2 h-4 w-4" />
			Application
		</DropdownMenuItem>
		<DropdownMenuItem onclick={onAddDatabase}>
			<Database class="mr-2 h-4 w-4" />
			Database
		</DropdownMenuItem>
		{#if onAddCompose}
			<DropdownMenuSeparator />
			<DropdownMenuItem onclick={onAddCompose}>
				<FileStack class="mr-2 h-4 w-4" />
				Docker Compose
			</DropdownMenuItem>
		{/if}
		{#if onAddTemplate}
			<DropdownMenuItem onclick={onAddTemplate}>
				<LayoutGrid class="mr-2 h-4 w-4" />
				Template
			</DropdownMenuItem>
		{/if}
	</DropdownMenuContent>
</DropdownMenu>
