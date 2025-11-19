<script lang="ts">
	import type { DatabaseType } from '$lib/api/databases';

	import ClickhouseLogo from '$lib/components/logo/clickhouse.svelte';
	import KeyDBLogo from '$lib/components/logo/keydb.svelte';
	import MongoDBLogo from '$lib/components/logo/mongodb.svelte';
	import RedisLogo from '$lib/components/logo/redis.svelte';
	import PostgresLogo from '$lib/components/logo/postgres.svelte';
	import MysqlLogo from '$lib/components/logo/mysql.svelte';
	import MariaDBLogo from '$lib/components/logo/mariadb.svelte';
	import type { Component } from 'svelte';

	interface Props {
		selected?: DatabaseType;
		onSelect: (type: DatabaseType) => void;
	}

	let { selected = $bindable(), onSelect }: Props = $props();

	const databaseTypes: Array<{
		type: DatabaseType;
		label: string;
		icon: Component;
	}> = [
		{ type: 'postgresql', label: 'PostgreSQL', icon: PostgresLogo },
		{ type: 'mysql', label: 'MySQL', icon: MysqlLogo },
		{ type: 'mariadb', label: 'MariaDB', icon: MariaDBLogo },
		{ type: 'redis', label: 'Redis', icon: RedisLogo },
		{ type: 'keydb', label: 'KeyDB', icon: KeyDBLogo },
		{ type: 'dragonfly', label: 'Dragonfly', icon: RedisLogo },
		{ type: 'mongodb', label: 'MongoDB', icon: MongoDBLogo },
		{ type: 'clickhouse', label: 'ClickHouse', icon: ClickhouseLogo }
	];
</script>

<div class="grid grid-cols-3 gap-4">
	{#each databaseTypes as db}
		<button
			type="button"
			class="flex flex-col items-center gap-3 rounded-lg border-2 bg-card p-6 transition-all hover:border-primary/50 {selected ===
			db.type
				? 'border-primary'
				: 'border-border'}"
			onclick={() => {
				selected = db.type;
				onSelect(db.type);
			}}
		>
			<div class="flex h-12 w-12 items-center justify-center">
				<db.icon class="h-10 w-10" />
			</div>
			<span class="text-sm font-medium">{db.label}</span>
			{#if selected === db.type}
				<div
					class="flex h-5 w-5 items-center justify-center rounded-full bg-primary text-primary-foreground"
				>
					<svg class="h-3 w-3" fill="none" stroke="currentColor" viewBox="0 0 24 24">
						<path stroke-linecap="round" stroke-linejoin="round" stroke-width="3" d="M5 13l4 4L19 7"
						></path>
					</svg>
				</div>
			{/if}
		</button>
	{/each}
</div>
