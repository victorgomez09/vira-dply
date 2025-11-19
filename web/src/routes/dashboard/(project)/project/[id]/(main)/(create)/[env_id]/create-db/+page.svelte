<script lang="ts">
	import { goto } from '$app/navigation';
	import { page } from '$app/state';
	import { Button } from '$lib/components/ui/button';
	import { Card, CardContent, CardHeader, CardTitle } from '$lib/components/ui/card';
	import { Input } from '$lib/components/ui/input';
	import { Label } from '$lib/components/ui/label';
	import { Textarea } from '$lib/components/ui/textarea';
	import DatabaseTypeSelector from '$lib/components/databases/database-type-selector.svelte';
	import PostgresqlForm from '$lib/components/databases/postgresql-form.svelte';
	import MysqlForm from '$lib/components/databases/mysql-form.svelte';
	import MariadbForm from '$lib/components/databases/mariadb-form.svelte';
	import RedisForm from '$lib/components/databases/redis-form.svelte';
	import KeydbForm from '$lib/components/databases/keydb-form.svelte';
	import DragonflyForm from '$lib/components/databases/dragonfly-form.svelte';
	import MongodbForm from '$lib/components/databases/mongodb-form.svelte';
	import ClickhouseForm from '$lib/components/databases/clickhouse-form.svelte';

	import { createMutation } from '@tanstack/svelte-query';
	import {
		databasesApi,
		type DatabaseType,
		type CreateDatabaseRequest,
		type PostgreSQLConfig,
		type MySQLConfig,
		type MariaDBConfig,
		type RedisConfig,
		type KeyDBConfig,
		type DragonflyConfig,
		type MongoDBConfig,
		type ClickHouseConfig
	} from '$lib/api/databases';

	const projectId = $derived(page.params.id);
	const envId = $derived(page.params.env_id);

	let step = $state<'type' | 'config'>('type');
	let selectedType = $state<DatabaseType | null>(null);
	let databaseName = $state('');
	let description = $state('');
	let selectedEnvironmentId = $state<string>(envId);

	let postgresqlConfig = $state<PostgreSQLConfig>({
		version: '17',
		database_name: '',
		username: '',
		password: '',
		port: 5432
	});

	let mysqlConfig = $state<MySQLConfig>({
		version: '8.4',
		database_name: '',
		username: '',
		password: '',
		root_password: '',
		port: 3306
	});

	let mariadbConfig = $state<MariaDBConfig>({
		version: '11.5',
		database_name: '',
		username: '',
		password: '',
		root_password: '',
		port: 3306
	});

	let redisConfig = $state<RedisConfig>({
		version: '7.4',
		password: '',
		port: 6379,
		database: 0
	});

	let keydbConfig = $state<KeyDBConfig>({
		version: '6.3',
		password: '',
		port: 6379,
		database: 0
	});

	let dragonflyConfig = $state<DragonflyConfig>({
		version: '1.23',
		password: '',
		port: 6379,
		persistence: false
	});

	let mongodbConfig = $state<MongoDBConfig>({
		version: '8.0',
		database_name: '',
		username: '',
		password: '',
		port: 27017
	});

	let clickhouseConfig = $state<ClickHouseConfig>({
		version: '24.10',
		database_name: '',
		username: '',
		password: '',
		port: 9000,
		http_port: 8123
	});

	const createDatabaseMutation = createMutation(() => ({
		mutationFn: (data: CreateDatabaseRequest) => databasesApi.create(projectId, data),
		onSuccess: () => {
			goto(`/dashboard/project/${projectId}`);
		}
	}));

	function handleTypeSelect(type: DatabaseType) {
		selectedType = type;
		step = 'config';
	}

	function handleBack() {
		if (step === 'config') {
			step = 'type';
			selectedType = null;
		} else {
			goto(`/dashboard/project/${projectId}`);
		}
	}

	function handleSubmit() {
		if (!selectedType || !selectedEnvironmentId || !databaseName) {
			return;
		}

		const typeConfigMap = {
			postgresql: postgresqlConfig,
			mysql: mysqlConfig,
			mariadb: mariadbConfig,
			redis: redisConfig,
			keydb: keydbConfig,
			dragonfly: dragonflyConfig,
			mongodb: mongodbConfig,
			clickhouse: clickhouseConfig
		};

		const selectedConfig = typeConfigMap[selectedType];

		const config: CreateDatabaseRequest['config'] = {
			type: selectedType,
			[selectedType]: selectedConfig
		};

		createDatabaseMutation.mutate({
			name: databaseName,
			description: description || undefined,
			type: selectedType,
			environment_id: selectedEnvironmentId,
			config
		});
	}
</script>

<div class="container max-w-5xl py-8">
	<div class="mb-6">
		<h1 class="text-3xl font-bold">Create Database</h1>
		<p class="text-muted-foreground">Add a new database to your project</p>
	</div>

	{#if step === 'type'}
		<Card>
			<CardHeader>
				<CardTitle>Select Database Type</CardTitle>
			</CardHeader>
			<CardContent>
				<DatabaseTypeSelector onSelect={handleTypeSelect} />
			</CardContent>
		</Card>
	{:else if step === 'config' && selectedType}
		<Card>
			<CardHeader>
				<CardTitle>Configure {selectedType}</CardTitle>
			</CardHeader>
			<CardContent>
				<form
					onsubmit={(e) => {
						e.preventDefault();
						handleSubmit();
					}}
					class="space-y-6"
				>
					<div class="space-y-4">
						<div class="space-y-2">
							<Label for="name">Database Name *</Label>
							<Input id="name" placeholder="my-database" bind:value={databaseName} required />
						</div>

						<div class="space-y-2">
							<Label for="description">Description</Label>
							<Textarea
								id="description"
								placeholder="Enter a description for your database"
								bind:value={description}
							/>
						</div>
					</div>

					<div class="border-t pt-6">
						<h3 class="text-lg font-semibold mb-4">Database Configuration</h3>
						{#if selectedType === 'postgresql'}
							<PostgresqlForm
								bind:config={postgresqlConfig}
								onConfigChange={(c) => (postgresqlConfig = c)}
							/>
						{:else if selectedType === 'mysql'}
							<MysqlForm bind:config={mysqlConfig} onConfigChange={(c) => (mysqlConfig = c)} />
						{:else if selectedType === 'mariadb'}
							<MariadbForm
								bind:config={mariadbConfig}
								onConfigChange={(c) => (mariadbConfig = c)}
							/>
						{:else if selectedType === 'redis'}
							<RedisForm bind:config={redisConfig} onConfigChange={(c) => (redisConfig = c)} />
						{:else if selectedType === 'keydb'}
							<KeydbForm bind:config={keydbConfig} onConfigChange={(c) => (keydbConfig = c)} />
						{:else if selectedType === 'dragonfly'}
							<DragonflyForm
								bind:config={dragonflyConfig}
								onConfigChange={(c) => (dragonflyConfig = c)}
							/>
						{:else if selectedType === 'mongodb'}
							<MongodbForm
								bind:config={mongodbConfig}
								onConfigChange={(c) => (mongodbConfig = c)}
							/>
						{:else if selectedType === 'clickhouse'}
							<ClickhouseForm
								bind:config={clickhouseConfig}
								onConfigChange={(c) => (clickhouseConfig = c)}
							/>
						{/if}
					</div>

					<div class="flex justify-end gap-4 pt-4">
						<Button type="button" variant="outline" onclick={handleBack}>Back</Button>
						<Button type="submit" disabled={createDatabaseMutation.isPending}>
							{createDatabaseMutation.isPending ? 'Creating...' : 'Create Database'}
						</Button>
					</div>
				</form>
			</CardContent>
		</Card>
	{/if}
</div>
