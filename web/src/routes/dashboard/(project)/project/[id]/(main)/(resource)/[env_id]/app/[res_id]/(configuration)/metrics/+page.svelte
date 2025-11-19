<script lang="ts">
	import { page } from '$app/state';
	import { goto } from '$app/navigation';
	import { Button } from '$lib/components/ui/button';
	import { Badge } from '$lib/components/ui/badge';
	import { Card, CardContent, CardHeader, CardTitle } from '$lib/components/ui/card';
	import { TrendingUp, Users, Globe, Clock, Activity, BarChart3, Download } from 'lucide-svelte';

	const projectId = page.params.id;

	let project = $state({
		name: 'Focalpoint Dashboard',
		workspace: 'Focalpoint',
		category: 'Applications'
	});

	let selectedTimeRange = $state('7d');

	// Mock analytics data
	let analytics = $state({
		overview: {
			totalRequests: 1247832,
			uniqueVisitors: 45621,
			avgResponseTime: 245,
			uptime: 99.97
		},
		traffic: [
			{ date: '2024-01-25', requests: 12450, visitors: 3421 },
			{ date: '2024-01-26', requests: 15230, visitors: 4123 },
			{ date: '2024-01-27', requests: 11890, visitors: 3876 },
			{ date: '2024-01-28', requests: 18340, visitors: 5234 },
			{ date: '2024-01-29', requests: 16720, visitors: 4891 },
			{ date: '2024-01-30', requests: 14560, visitors: 4234 },
			{ date: '2024-01-31', requests: 13890, visitors: 3987 }
		],
		topPages: [
			{ path: '/dashboard', views: 45231, percentage: 32.1 },
			{ path: '/login', views: 23456, percentage: 16.7 },
			{ path: '/profile', views: 18934, percentage: 13.4 },
			{ path: '/settings', views: 12876, percentage: 9.1 },
			{ path: '/api/users', views: 9876, percentage: 7.0 }
		],
		countries: [
			{ country: 'United States', visitors: 18234, percentage: 39.9 },
			{ country: 'United Kingdom', visitors: 8765, percentage: 19.2 },
			{ country: 'Germany', visitors: 5432, percentage: 11.9 },
			{ country: 'France', visitors: 4321, percentage: 9.5 },
			{ country: 'Canada', visitors: 3210, percentage: 7.0 }
		],
		errors: [
			{ code: '404', count: 1234, percentage: 45.2 },
			{ code: '500', count: 876, percentage: 32.1 },
			{ code: '403', count: 432, percentage: 15.8 },
			{ code: '502', count: 189, percentage: 6.9 }
		]
	});

	function formatNumber(num: number) {
		return new Intl.NumberFormat().format(num);
	}

	function exportData() {
		console.log('Exporting analytics data...');
	}
</script>

<svelte:head>
	<title>Analytics - {project.name}</title>
</svelte:head>

<!-- Main Content -->
<div class="flex-1 p-6">
	<!-- Header -->
	<div class="flex items-center justify-between mb-6">
		<div class="flex items-center space-x-4">
			<div>
				<h1 class="text-2xl font-semibold text-gray-900">Analytics</h1>
				<p class="text-sm text-gray-500 mt-1">
					Monitor traffic, performance, and user behavior for {project.name}.
				</p>
			</div>
		</div>
		<div class="flex items-center space-x-3">
			<select
				bind:value={selectedTimeRange}
				class="px-3 py-2 border border-gray-300 rounded-md text-sm"
			>
				<option value="24h">Last 24 hours</option>
				<option value="7d">Last 7 days</option>
				<option value="30d">Last 30 days</option>
				<option value="90d">Last 90 days</option>
			</select>
			<Button variant="outline" onclick={exportData}>
				<Download class="w-4 h-4 mr-2" />
				Export
			</Button>
		</div>
	</div>

	<!-- Overview Cards -->
	<div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-6 mb-8">
		<Card>
			<CardContent class="p-6">
				<div class="flex items-center justify-between">
					<div>
						<p class="text-sm font-medium text-gray-600">Total Requests</p>
						<p class="text-2xl font-bold text-gray-900">
							{formatNumber(analytics.overview.totalRequests)}
						</p>
					</div>
					<BarChart3 class="w-8 h-8 text-blue-500" />
				</div>
				<div class="flex items-center mt-2">
					<TrendingUp class="w-4 h-4 text-green-500 mr-1" />
					<span class="text-sm text-green-600">+12.5% from last week</span>
				</div>
			</CardContent>
		</Card>

		<Card>
			<CardContent class="p-6">
				<div class="flex items-center justify-between">
					<div>
						<p class="text-sm font-medium text-gray-600">Unique Visitors</p>
						<p class="text-2xl font-bold text-gray-900">
							{formatNumber(analytics.overview.uniqueVisitors)}
						</p>
					</div>
					<Users class="w-8 h-8 text-green-500" />
				</div>
				<div class="flex items-center mt-2">
					<TrendingUp class="w-4 h-4 text-green-500 mr-1" />
					<span class="text-sm text-green-600">+8.2% from last week</span>
				</div>
			</CardContent>
		</Card>

		<Card>
			<CardContent class="p-6">
				<div class="flex items-center justify-between">
					<div>
						<p class="text-sm font-medium text-gray-600">Avg Response Time</p>
						<p class="text-2xl font-bold text-gray-900">
							{analytics.overview.avgResponseTime}ms
						</p>
					</div>
					<Clock class="w-8 h-8 text-yellow-500" />
				</div>
				<div class="flex items-center mt-2">
					<TrendingUp class="w-4 h-4 text-red-500 mr-1" />
					<span class="text-sm text-red-600">+5.1% from last week</span>
				</div>
			</CardContent>
		</Card>

		<Card>
			<CardContent class="p-6">
				<div class="flex items-center justify-between">
					<div>
						<p class="text-sm font-medium text-gray-600">Uptime</p>
						<p class="text-2xl font-bold text-gray-900">{analytics.overview.uptime}%</p>
					</div>
					<Activity class="w-8 h-8 text-green-500" />
				</div>
				<div class="flex items-center mt-2">
					<span class="text-sm text-gray-600">99.9% SLA target</span>
				</div>
			</CardContent>
		</Card>
	</div>

	<!-- Charts and Tables -->
	<div class="grid grid-cols-1 lg:grid-cols-2 gap-6 mb-8">
		<!-- Traffic Chart -->
		<Card>
			<CardHeader>
				<CardTitle>Traffic Overview</CardTitle>
			</CardHeader>
			<CardContent>
				<div class="h-64 bg-gray-50 rounded-lg flex items-center justify-center">
					<div class="text-center">
						<BarChart3 class="w-12 h-12 text-gray-400 mx-auto mb-2" />
						<p class="text-gray-500">Traffic chart would be rendered here</p>
					</div>
				</div>
			</CardContent>
		</Card>

		<!-- Top Pages -->
		<Card>
			<CardHeader>
				<CardTitle>Top Pages</CardTitle>
			</CardHeader>
			<CardContent>
				<div class="space-y-4">
					{#each analytics.topPages as page (page.path)}
						<div class="flex items-center justify-between">
							<div class="flex-1">
								<p class="font-mono text-sm text-gray-900">{page.path}</p>
								<div class="w-full bg-gray-200 rounded-full h-2 mt-1">
									<div class="bg-blue-500 h-2 rounded-full" style="width: {page.percentage}%"></div>
								</div>
							</div>
							<div class="ml-4 text-right">
								<p class="text-sm font-medium text-gray-900">{formatNumber(page.views)}</p>
								<p class="text-xs text-gray-500">{page.percentage}%</p>
							</div>
						</div>
					{/each}
				</div>
			</CardContent>
		</Card>
	</div>

	<div class="grid grid-cols-1 lg:grid-cols-2 gap-6">
		<!-- Countries -->
		<Card>
			<CardHeader>
				<CardTitle>Top Countries</CardTitle>
			</CardHeader>
			<CardContent>
				<div class="space-y-4">
					{#each analytics.countries as country (country.country)}
						<div class="flex items-center justify-between">
							<div class="flex items-center space-x-3">
								<Globe class="w-4 h-4 text-gray-500" />
								<span class="text-sm text-gray-900">{country.country}</span>
							</div>
							<div class="text-right">
								<p class="text-sm font-medium text-gray-900">
									{formatNumber(country.visitors)}
								</p>
								<p class="text-xs text-gray-500">{country.percentage}%</p>
							</div>
						</div>
					{/each}
				</div>
			</CardContent>
		</Card>

		<!-- Error Codes -->
		<Card>
			<CardHeader>
				<CardTitle>Error Codes</CardTitle>
			</CardHeader>
			<CardContent>
				<div class="space-y-4">
					{#each analytics.errors as error (error)}
						<div class="flex items-center justify-between">
							<div class="flex items-center space-x-3">
								<Badge variant="outline" class="font-mono">{error.code}</Badge>
							</div>
							<div class="text-right">
								<p class="text-sm font-medium text-gray-900">{formatNumber(error.count)}</p>
								<p class="text-xs text-gray-500">{error.percentage}%</p>
							</div>
						</div>
					{/each}
				</div>
			</CardContent>
		</Card>
	</div>
</div>
