<script lang="ts">
	import { Label } from '$lib/components/ui/label';
	import { Textarea } from '$lib/components/ui/textarea';
	import { Button } from '$lib/components/ui/button';
	import { Tabs, TabsContent, TabsList, TabsTrigger } from '$lib/components/ui/tabs';
	import { Upload } from 'lucide-svelte';

	interface Props {
		content: string;
		method: 'paste' | 'upload';
		fileType: 'dockerfile' | 'compose';
		onUpdate: (config: { content: string; method: string; fileType: string }) => void;
	}

	let {
		content = $bindable(),
		method = $bindable(),
		fileType = $bindable(),
		onUpdate
	}: Props = $props();

	let fileInput: HTMLInputElement;

	function handleFileSelect(event: Event) {
		const target = event.target as HTMLInputElement;
		const file = target.files?.[0];
		if (file) {
			const reader = new FileReader();
			reader.onload = (e) => {
				content = e.target?.result as string;
			};
			reader.readAsText(file);
		}
	}
</script>

<div class="space-y-6">
	<Tabs value={fileType} onValueChange={(v) => (fileType = v)}>
		<TabsList class="grid w-full grid-cols-2">
			<TabsTrigger value="dockerfile">Dockerfile</TabsTrigger>
			<TabsTrigger value="compose">Docker Compose</TabsTrigger>
		</TabsList>
	</Tabs>

	<Tabs value={method} onValueChange={(v) => (method = v)}>
		<TabsList class="grid w-full grid-cols-2">
			<TabsTrigger value="paste">Paste content</TabsTrigger>
			<TabsTrigger value="upload">Upload file</TabsTrigger>
		</TabsList>

		<TabsContent value="paste" class="space-y-2 mt-4">
			<Label for="docker-content">
				{fileType === 'dockerfile' ? 'Dockerfile' : 'Docker Compose'} content
			</Label>
			<Textarea
				id="docker-content"
				placeholder={fileType === 'dockerfile'
					? 'FROM node:18\nWORKDIR /app\nCOPY . .\nRUN npm install\nCMD ["npm", "start"]'
					: 'version: "3.8"\nservices:\n  app:\n    build: .\n    ports:\n      - "3000:3000"'}
				bind:value={content}
				rows={12}
				class="font-mono text-sm"
			/>
		</TabsContent>

		<TabsContent value="upload" class="space-y-4 mt-4">
			<div
				class="border-2 border-dashed rounded-lg p-8 text-center hover:border-primary/50 transition-colors"
			>
				<Upload class="h-8 w-8 mx-auto mb-4 text-muted-foreground" />
				<p class="text-sm text-muted-foreground mb-4">
					Click to upload or drag and drop your {fileType === 'dockerfile'
						? 'Dockerfile'
						: 'docker-compose.yml'}
				</p>
				<input
					type="file"
					bind:this={fileInput}
					onchange={handleFileSelect}
					accept={fileType === 'dockerfile' ? '*' : '.yml,.yaml'}
					class="hidden"
				/>
				<Button type="button" variant="outline" onclick={() => fileInput?.click()}>
					Select File
				</Button>
			</div>
			{#if content}
				<div class="bg-muted p-4 rounded-lg">
					<p class="text-sm font-medium mb-2">File content preview:</p>
					<pre
						class="text-xs overflow-auto max-h-48 bg-background p-3 rounded border">{content}</pre>
				</div>
			{/if}
		</TabsContent>
	</Tabs>
</div>
