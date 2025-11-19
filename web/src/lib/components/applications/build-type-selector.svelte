<script lang="ts">
	import { Label } from '$lib/components/ui/label';
	import { Input } from '$lib/components/ui/input';
	import { Card, CardContent } from '$lib/components/ui/card';
	import { cn } from '$lib/utils';

	interface Props {
		selected: 'nixpacks' | 'heroku' | 'paketo' | 'static' | 'dockerfile' | 'compose';
		onSelect: (
			type: 'nixpacks' | 'heroku' | 'paketo' | 'static' | 'dockerfile' | 'compose'
		) => void;
		publishDirectory?: string;
		onPublishDirectoryChange?: (value: string) => void;
	}

	let { selected, onSelect, publishDirectory = '', onPublishDirectoryChange }: Props = $props();

	const options = [
		{
			value: 'nixpacks' as const,
			title: 'Nixpacks',
			description: 'Auto-detect and build using Nixpacks (recommended)'
		},
		{
			value: 'heroku' as const,
			title: 'Heroku Buildpacks',
			description: 'Use Heroku-compatible buildpacks'
		},
		{
			value: 'paketo' as const,
			title: 'Paketo Buildpacks',
			description: 'Use Cloud Native Buildpacks (Paketo)'
		},
		{
			value: 'static' as const,
			title: 'Static site',
			description: 'Serve static files with nginx'
		},
		{
			value: 'dockerfile' as const,
			title: 'Dockerfile',
			description: 'Build following a dockerfile'
		},

		{
			value: 'compose' as const,
			title: 'Compose',
			description: 'Build following a composefile'
		}
	];
</script>

<div class="space-y-4">
	<div>
		<Label class="text-base font-semibold">Build method</Label>
		<p class="text-sm text-muted-foreground mt-1">Choose how your application should be built</p>
	</div>

	<div class="space-y-2">
		{#each options as option}
			<Card
				class={cn(
					'cursor-pointer transition-all hover:border-primary/50',
					selected === option.value && 'border-primary border-2'
				)}
				onclick={() => onSelect(option.value)}
			>
				<CardContent class="p-4">
					<div class="flex items-start gap-3">
						<div
							class={cn(
								'w-4 h-4 rounded-full border-2 mt-0.5 flex-shrink-0',
								selected === option.value ? 'border-primary bg-primary' : 'border-muted-foreground'
							)}
						>
							{#if selected === option.value}
								<div class="w-full h-full rounded-full bg-background scale-50"></div>
							{/if}
						</div>
						<div>
							<div class="font-medium">{option.title}</div>
							<div class="text-sm text-muted-foreground">{option.description}</div>
						</div>
					</div>
				</CardContent>
			</Card>
		{/each}
	</div>

	{#if selected === 'static'}
		<div class="pl-7 space-y-2">
			<Label for="publish-dir">Publish directory</Label>
			<Input
				id="publish-dir"
				placeholder="dist"
				value={publishDirectory}
				oninput={(e) => onPublishDirectoryChange?.(e.currentTarget.value)}
			/>
			<p class="text-xs text-muted-foreground">The directory containing your built static files</p>
		</div>
	{/if}
</div>
