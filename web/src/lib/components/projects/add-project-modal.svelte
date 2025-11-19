<script lang="ts">
	import { Button } from '$lib/components/ui/button';
	import { Input } from '$lib/components/ui/input';
	import { Label } from '$lib/components/ui/label';
	import {
		Sheet,
		SheetContent,
		SheetDescription,
		SheetFooter,
		SheetHeader,
		SheetTitle
	} from '$lib/components/ui/sheet';
	import type { CreateProjectRequest } from '$lib/api';

	interface Props {
		open?: boolean;
		onClose?: () => void;
		onSubmit?: (data: CreateProjectRequest) => Promise<void>;
	}

	let { open = $bindable(false), onClose, onSubmit }: Props = $props();

	let name = $state('');
	let description = $state('');
	let loading = $state(false);
	let error = $state('');

	const handleSubmit = async (e: Event) => {
		e.preventDefault();
		error = '';

		if (!name.trim()) {
			error = 'Project name is required';
			return;
		}

		loading = true;
		try {
			await onSubmit?.({ name: name.trim(), description: description.trim() || undefined });
			name = '';
			description = '';
			open = false;
			onClose?.();
		} catch (err) {
			error = err instanceof Error ? err.message : 'Failed to create project';
		} finally {
			loading = false;
		}
	};

	const handleOpenChange = (isOpen: boolean) => {
		open = isOpen;
		if (!isOpen) {
			name = '';
			description = '';
			error = '';
			onClose?.();
		}
	};
</script>

<Sheet bind:open onOpenChange={handleOpenChange}>
	<SheetContent side="right">
		<SheetHeader>
			<SheetTitle>Create New Project</SheetTitle>
			<SheetDescription>Add a new project to your workspace.</SheetDescription>
		</SheetHeader>

		<form onsubmit={handleSubmit} class="mt-6 space-y-4">
			<div class="space-y-2">
				<Label for="name">Project Name</Label>
				<Input
					id="name"
					bind:value={name}
					placeholder="my-awesome-project"
					disabled={loading}
					required
				/>
			</div>

			<div class="space-y-2">
				<Label for="description">Description (optional)</Label>
				<Input
					id="description"
					bind:value={description}
					placeholder="A brief description of your project"
					disabled={loading}
				/>
			</div>

			{#if error}
				<p class="text-destructive text-sm">{error}</p>
			{/if}

			<SheetFooter>
				<Button type="button" variant="outline" onclick={() => (open = false)} disabled={loading}>
					Cancel
				</Button>
				<Button type="submit" disabled={loading}>
					{loading ? 'Creating...' : 'Create Project'}
				</Button>
			</SheetFooter>
		</form>
	</SheetContent>
</Sheet>
