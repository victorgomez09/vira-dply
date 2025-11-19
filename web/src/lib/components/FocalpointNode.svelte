<script lang="ts">
	import { Activity, Earth, Layers, MapPin, Settings, Terminal } from 'lucide-svelte';
	import { Position, Handle, type NodeProps } from '@xyflow/svelte';
	import Button from './ui/button/button.svelte';

	let { data }: NodeProps = $props();
</script>

<div class="space-y-6 bg-gray-300/20 border border-gray-200 rounded-lg p-4 shadow-sm min-w-[280px]">
	<!-- Service Info -->
	<div>
		<div class="flex items-center space-x-2 mb-4">
			<Layers class="w-4 h-4 text-gray-600" />
			<span class="font-medium text-gray-900">{data.name}</span>
		</div>

		<div class="bg-white rounded-lg p-4 space-y-3">
			<Handle type="target" position={Position.Left} />
			<div class="flex items-center space-x-2 bg-white">
				<Earth class="w-4 h-4 text-gray-600" />
				<span class="text-sm font-medium text-gray-900">Web process</span>
			</div>

			<div class="space-y-2 text-sm">
				<div class="flex items-center justify-between">
					<span class="text-gray-600">Start command</span>
					<code class="bg-gray-200 px-2 py-1 rounded text-xs font-mono">{data.startCommand}</code>
				</div>
				<div class="flex items-center justify-between">
					<span class="text-gray-600">Resources</span>
					<span class="font-medium">{data.resources}</span>
				</div>
				<div class="flex items-center justify-between">
					<span class="text-gray-600">Instances</span>
					<span class="font-medium">{data.instances}</span>
				</div>
				<div class="flex items-center justify-between">
					<span class="text-gray-600">Domain</span>
					<span class="font-medium text-blue-600">{data.domain}</span>
				</div>
			</div>

			<div class="flex space-x-2 pt-2">
				<Button variant="outline" size="sm" class="flex-1">
					<Activity class="w-4 h-4 mr-1" />
					View logs
				</Button>
				<Button variant="outline" size="sm" class="flex-1">
					<Terminal class="w-4 h-4 mr-1" />
					Web terminal
				</Button>
				<Button variant="outline" size="sm" class="flex-1">
					<Settings class="w-4 h-4 mr-1" />
					Settings
				</Button>
			</div>
		</div>
	</div>

	<!-- Location -->
	<div class="flex items-center space-x-2 text-sm text-gray-600">
		<MapPin class="w-4 h-4" />
		<span>Council Bluffs, Iowa (us-central1)</span>
	</div>
</div>
