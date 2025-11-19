<script lang="ts">
	import type { Application } from '$lib/api';
	import { Card } from '$lib/components/ui/card';
	import { Badge } from '$lib/components/ui/badge';
	import { Circle, Globe } from 'lucide-svelte';

	interface Props {
		application: Application;
		onclick?: () => void;
	}

	let { application, onclick }: Props = $props();

	const statusColors: Record<string, string> = {
		running: 'bg-green-500',
		stopped: 'bg-gray-500',
		failed: 'bg-red-500',
		pending: 'bg-yellow-500',
		deploying: 'bg-blue-500'
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

<button type="button" class="w-full text-left" onclick={onclick}>
	<Card class="group relative p-4 transition-shadow hover:shadow-md">
		<div class="flex items-start justify-between">
			<div class="flex-1">
				<div class="mb-2 flex items-center gap-2">
					<h3 class="font-semibold">{application.name}</h3>
					<Badge
						variant="outline"
						class={`gap-1 ${application.status === 'running' ? 'border-green-500/50' : ''}`}
					>
						<Circle
							class={`size-2 ${statusColors[application.status] || 'bg-gray-500'} rounded-full`}
						/>
						{application.status}
					</Badge>
				</div>

				{#if application.description}
					<p class="text-muted-foreground mb-2 text-sm">{application.description}</p>
				{/if}

				{#if application.domain}
					<div class="flex items-center gap-1 text-sm">
						<Globe class="text-muted-foreground size-3" />
						<a
							href={`https://${application.domain}`}
							target="_blank"
							rel="noopener noreferrer"
							class="text-primary hover:underline"
						>
							{application.domain}
						</a>
					</div>
				{/if}
			</div>
		</div>

		<div class="text-muted-foreground mt-3 text-xs">Created: {formatDate(application.created_at)}</div>
	</Card>
</button>
