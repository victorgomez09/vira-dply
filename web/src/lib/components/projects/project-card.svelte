<script lang="ts">
	import type { Project } from '$lib/api';
	import { Card } from '$lib/components/ui/card';
	import { Button } from '$lib/components/ui/button';
	import { Badge } from '$lib/components/ui/badge';
	import {
		DropdownMenu,
		DropdownMenuContent,
		DropdownMenuItem,
		DropdownMenuTrigger
	} from '$lib/components/ui/dropdown-menu';
	import { EllipsisVertical, FolderGit2, Trash2, Settings } from 'lucide-svelte';

	interface Props {
		project: Project;
		onDelete?: (id: string) => void;
		onEdit?: (id: string) => void;
		onclick?: () => void;
	}

	let { project, onDelete, onEdit, onclick }: Props = $props();

	const formatDate = (dateString: string) => {
		const date = new Date(dateString);
		const now = new Date();
		const diffMs = now.getTime() - date.getTime();
		const diffMins = Math.floor(diffMs / 60000);
		const diffHours = Math.floor(diffMs / 3600000);
		const diffDays = Math.floor(diffMs / 86400000);

		if (diffMins < 1) return 'Just now';
		if (diffMins < 60) return `${diffMins}m ago`;
		if (diffHours < 24) return `${diffHours}h ago`;
		if (diffDays < 30) return `${diffDays}d ago`;
		return date.toLocaleDateString();
	};
</script>

<Card
	class="group relative flex flex-col gap-4 p-6 transition-shadow hover:shadow-md {onclick
		? 'cursor-pointer'
		: ''}"
	onclick={onclick}
>
	<div class="flex items-start justify-between">
		<div class="flex items-center gap-3">
			<div
				class="bg-primary/10 text-primary flex size-12 items-center justify-center rounded-lg"
			>
				<FolderGit2 class="size-6" />
			</div>
			<div>
				<h3 class="font-semibold text-lg">{project.name}</h3>
				{#if project.description}
					<p class="text-muted-foreground text-sm">{project.description}</p>
				{/if}
			</div>
		</div>

		<DropdownMenu>
			<DropdownMenuTrigger>
				<Button variant="ghost" size="icon" class="size-8">
					<EllipsisVertical class="size-4" />
					<span class="sr-only">Open menu</span>
				</Button>
			</DropdownMenuTrigger>
			<DropdownMenuContent align="end">
				<DropdownMenuItem onclick={() => onEdit?.(project.id)}>
					<Settings class="mr-2 size-4" />
					<span>Settings</span>
				</DropdownMenuItem>
				<DropdownMenuItem onclick={() => onDelete?.(project.id)} class="text-destructive">
					<Trash2 class="mr-2 size-4" />
					<span>Delete</span>
				</DropdownMenuItem>
			</DropdownMenuContent>
		</DropdownMenu>
	</div>

	<div class="flex items-center justify-between">
		<div class="flex items-center gap-2 text-sm">
			<span class="text-muted-foreground">Last deploy:</span>
			<span>{formatDate(project.created_at)}</span>
		</div>

		<div class="flex items-center gap-2">
			<Badge variant="outline" class="gap-1">
				<div class="size-2 rounded-full bg-green-500"></div>
				Active
			</Badge>
		</div>
	</div>
</Card>
