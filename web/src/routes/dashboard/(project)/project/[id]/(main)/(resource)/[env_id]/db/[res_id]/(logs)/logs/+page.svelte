<script lang="ts">
	import { page } from '$app/state';
	import { Button } from '$lib/components/ui/button';
	import { Badge } from '$lib/components/ui/badge';
	import { Card, CardContent } from '$lib/components/ui/card';
	import { createQuery } from '@tanstack/svelte-query';
	import { databasesApi } from '$lib/api/databases';
	import { Download, Pause, Play, RotateCcw } from 'lucide-svelte';
	import { onMount, onDestroy } from 'svelte';
	import { toast } from 'svelte-sonner';

	const projectId = $derived(page.params.id);
	const resId = $derived(page.params.res_id);

	const databaseQuery = createQuery(() => ({
		queryKey: ['database', projectId, resId],
		queryFn: () => databasesApi.get(projectId, resId),
		enabled: !!projectId && !!resId
	}));

	const database = $derived(databaseQuery.data);

	let isStreaming = $state(true);
	let selectedLevel = $state('all');
	let autoScroll = $state(true);
	let logContainer = $state<HTMLDivElement>();
	let stopStreaming: (() => void) | null = null;

	interface LogEntry {
		id: number;
		timestamp: string;
		level: string;
		message: string;
		source: string;
	}

	let logs = $state<LogEntry[]>([]);
	let logId = 0;

	function parseLogLine(line: string): Omit<LogEntry, 'id'> | null {
		const timestampRegex = /^(\d{4}-\d{2}-\d{2}[T\s]\d{2}:\d{2}:\d{2})/;
		const levelRegex = /(ERROR|WARN|INFO|DEBUG)/i;

		const timestamp =
			timestampRegex.exec(line)?.[1] || new Date().toISOString().replace('T', ' ').slice(0, 19);
		const levelMatch = levelRegex.exec(line);
		const level = levelMatch ? levelMatch[1].toLowerCase() : 'info';

		const source = database?.type || 'database';
		const message = line.replace(timestampRegex, '').replace(levelRegex, '').trim() || line;

		return {
			timestamp,
			level,
			message,
			source
		};
	}

	function addLog(line: string) {
		const parsed = parseLogLine(line);
		if (parsed) {
			logs = [...logs, { ...parsed, id: logId++ }];

			if (autoScroll && logContainer) {
				setTimeout(() => {
					logContainer.scrollTop = logContainer.scrollHeight;
				}, 0);
			}
		}
	}

	async function startLogStreaming() {
		if (!projectId || !resId) return;

		stopStreaming = await databasesApi.streamLogs(
			projectId,
			resId,
			true,
			(line) => {
				addLog(line);
			},
			(error) => {
				toast.error('Failed to stream logs', {
					description: error.message
				});
				isStreaming = false;
			}
		);
	}

	function getLevelColor(level: string) {
		switch (level) {
			case 'error':
				return 'text-red-400 bg-red-950/50 dark:text-red-400 dark:bg-red-950/50';
			case 'warn':
				return 'text-yellow-400 bg-yellow-950/50 dark:text-yellow-400 dark:bg-yellow-950/50';
			case 'info':
				return 'text-blue-400 bg-blue-950/50 dark:text-blue-400 dark:bg-blue-950/50';
			case 'debug':
				return 'text-gray-400 bg-gray-800/50 dark:text-gray-400 dark:bg-gray-800/50';
			default:
				return 'text-gray-400 bg-gray-800/50 dark:text-gray-400 dark:bg-gray-800/50';
		}
	}

	function toggleStreaming() {
		isStreaming = !isStreaming;
		if (isStreaming) {
			startLogStreaming();
		} else if (stopStreaming) {
			stopStreaming();
			stopStreaming = null;
		}
	}

	function downloadLogs() {
		const logText = logs
			.map((log) => `${log.timestamp} [${log.level.toUpperCase()}] [${log.source}] ${log.message}`)
			.join('\n');
		const blob = new Blob([logText], { type: 'text/plain' });
		const url = URL.createObjectURL(blob);
		const a = document.createElement('a');
		a.href = url;
		a.download = `${database?.name || 'database'}-logs-${new Date().toISOString()}.txt`;
		document.body.appendChild(a);
		a.click();
		document.body.removeChild(a);
		URL.revokeObjectURL(url);
	}

	function clearLogs() {
		logs = [];
	}

	const filteredLogs = $derived.by(() => {
		if (selectedLevel === 'all') return logs;
		return logs.filter((log) => log.level === selectedLevel);
	});

	onMount(() => {
		if (projectId && resId) {
			startLogStreaming();
		}
	});

	onDestroy(() => {
		if (stopStreaming) {
			stopStreaming();
		}
	});
</script>

{#if database}
	<div class="space-y-6">
		<div class="flex items-center justify-between">
			<div>
				<h2 class="text-2xl font-bold tracking-tight">Database Logs</h2>
				<p class="text-muted-foreground">Real-time logs for {database.name}</p>
			</div>
			<div class="flex items-center gap-3">
				<label class="flex items-center gap-2 text-sm">
					<input type="checkbox" bind:checked={autoScroll} class="rounded" />
					Auto-scroll
				</label>

				<select
					bind:value={selectedLevel}
					class="px-3 py-2 border border-input rounded-md text-sm bg-background"
				>
					<option value="all">All Levels</option>
					<option value="error">Error</option>
					<option value="warn">Warning</option>
					<option value="info">Info</option>
					<option value="debug">Debug</option>
				</select>

				<Button variant="outline" onclick={downloadLogs}>
					<Download class="w-4 h-4 mr-2" />
					Download
				</Button>
				<Button variant="outline" onclick={toggleStreaming}>
					{#if isStreaming}
						<Pause class="w-4 h-4 mr-2" />
						Pause
					{:else}
						<Play class="w-4 h-4 mr-2" />
						Resume
					{/if}
				</Button>
				<Button variant="outline" onclick={clearLogs}>
					<RotateCcw class="w-4 h-4 mr-2" />
					Clear
				</Button>
			</div>
		</div>

		<Card class="h-[calc(100vh-300px)] bg-card border-border">
			<CardContent class="p-0 h-full">
				<div bind:this={logContainer} class="h-full overflow-auto bg-background font-mono text-sm">
					<div class="p-4 space-y-1">
						{#if filteredLogs.length === 0}
							<div class="flex items-center justify-center h-full text-muted-foreground">
								<p>
									No logs yet. {isStreaming
										? 'Waiting for logs...'
										: 'Start streaming to see logs.'}
								</p>
							</div>
						{:else}
							{#each filteredLogs as log (log.id)}
								<div
									class="flex items-start space-x-4 py-1 hover:bg-accent/50 px-2 rounded transition-colors"
								>
									<span class="text-muted-foreground text-xs w-32 flex-shrink-0 mt-0.5">
										{log.timestamp}
									</span>
									<Badge
										variant="outline"
										class="text-xs {getLevelColor(log.level)} border-0 w-16 justify-center"
									>
										{log.level.toUpperCase()}
									</Badge>
									<span class="text-primary text-xs w-20 flex-shrink-0 mt-0.5">
										[{log.source}]
									</span>
									<span class="text-foreground flex-1 break-all">
										{log.message}
									</span>
								</div>
							{/each}
						{/if}

						{#if isStreaming}
							<div class="flex items-center space-x-2 py-2 px-2">
								<div class="w-2 h-2 bg-green-500 rounded-full animate-pulse"></div>
								<span class="text-muted-foreground text-xs">Live streaming...</span>
							</div>
						{/if}
					</div>
				</div>
			</CardContent>
		</Card>
	</div>
{/if}
