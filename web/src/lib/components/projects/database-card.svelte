<script lang="ts">
	import type { Database } from '$lib/api';
	import { Card } from '$lib/components/ui/card';
	import { Badge } from '$lib/components/ui/badge';
	import { Circle, Database as DatabaseIcon } from 'lucide-svelte';

	interface Props {
		database: Database;
		onclick?: () => void;
	}

	let { database, onclick }: Props = $props();

	const statusColors: Record<string, string> = {
		running: 'bg-green-500',
		stopped: 'bg-gray-500',
		failed: 'bg-red-500',
		pending: 'bg-yellow-500'
	};

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

<button type="button" class="w-full text-left" {onclick}>
	<Card class="group relative p-4 transition-shadow hover:shadow-md">
		<div class="flex items-start justify-between">
			<div class="flex-1">
				<div class="mb-2 flex items-center gap-2">
					<DatabaseIcon class="text-muted-foreground size-4" />
					<h3 class="font-semibold">{database.name}</h3>
					<Badge
						variant="outline"
						class={`gap-1 ${database.status === 'running' ? 'border-green-500/50' : ''}`}
					>
						<Circle
							class={`size-2 ${statusColors[database.status] || 'bg-gray-500'} rounded-full`}
						/>
						{database.status}
					</Badge>
				</div>

				{#if database.description}
					<p class="text-muted-foreground mb-2 text-sm">{database.description}</p>
				{/if}

				<div class="flex items-center gap-2 text-sm">
					<Badge variant="secondary" class="text-xs">{database.type.toUpperCase()}</Badge>
				</div>
			</div>
		</div>

		<div class="text-muted-foreground mt-3 text-xs">Created: {formatDate(database.created_at)}</div>
	</Card>
</button>
