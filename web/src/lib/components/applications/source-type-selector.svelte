<script lang="ts">
	import { Card, CardContent } from '$lib/components/ui/card';
	import { Github, Container, FileArchive } from 'lucide-svelte';
	import { cn } from '$lib/utils';

	interface Props {
		selected: 'git' | 'docker' | 'zip';
		onSelect: (type: 'git' | 'docker' | 'zip') => void;
	}

	let { selected, onSelect }: Props = $props();
</script>

<div class="grid grid-cols-1 md:grid-cols-3 gap-4">
	<Card
		class={cn(
			'cursor-pointer transition-all hover:border-primary/50',
			selected === 'git' && 'border-primary border-2'
		)}
		onclick={() => onSelect('git')}
	>
		<CardContent class="flex flex-col items-center justify-center py-8 space-y-4">
			<div
				class={cn(
					'w-12 h-12 rounded-full flex items-center justify-center',
					selected === 'git' ? 'bg-primary text-primary-foreground' : 'bg-muted'
				)}
			>
				<Github class="h-6 w-6" />
			</div>
			<div class="text-center">
				<h3 class="font-semibold text-lg">Git repository</h3>
				<p class="text-sm text-muted-foreground mt-1">
					Deploy from GitHub, GitLab, or Bitbucket
				</p>
			</div>
		</CardContent>
	</Card>

	<Card
		class={cn(
			'cursor-pointer transition-all hover:border-primary/50',
			selected === 'docker' && 'border-primary border-2'
		)}
		onclick={() => onSelect('docker')}
	>
		<CardContent class="flex flex-col items-center justify-center py-8 space-y-4">
			<div
				class={cn(
					'w-12 h-12 rounded-full flex items-center justify-center',
					selected === 'docker' ? 'bg-primary text-primary-foreground' : 'bg-muted'
				)}
			>
				<Container class="h-6 w-6" />
			</div>
			<div class="text-center">
				<h3 class="font-semibold text-lg">Docker image</h3>
				<p class="text-sm text-muted-foreground mt-1">Use a Dockerfile or Docker Compose</p>
			</div>
		</CardContent>
	</Card>

	<Card
		class={cn(
			'cursor-pointer transition-all hover:border-primary/50',
			selected === 'zip' && 'border-primary border-2'
		)}
		onclick={() => onSelect('zip')}
	>
		<CardContent class="flex flex-col items-center justify-center py-8 space-y-4">
			<div
				class={cn(
					'w-12 h-12 rounded-full flex items-center justify-center',
					selected === 'zip' ? 'bg-primary text-primary-foreground' : 'bg-muted'
				)}
			>
				<FileArchive class="h-6 w-6" />
			</div>
			<div class="text-center">
				<h3 class="font-semibold text-lg">Zipped file</h3>
				<p class="text-sm text-muted-foreground mt-1">Upload a compressed archive</p>
			</div>
		</CardContent>
	</Card>
</div>
