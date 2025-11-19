<script lang="ts">
	import { goto } from '$app/navigation';
	import Logo from '$lib/components/logo/logo.svelte';
	import { authApi, type ApiError } from '$lib/api';
	import { authStore } from '$lib/stores/auth.svelte';

	let name = $state('');
	let email = $state('');
	let password = $state('');
	let confirmPassword = $state('');
	let isLoading = $state(false);
	let error = $state('');

	async function handleSubmit(event: SubmitEvent) {
		event.preventDefault();

		if (!name || !email || !password || !confirmPassword) {
			error = 'Please fill in all fields';
			return;
		}

		if (password !== confirmPassword) {
			error = 'Passwords do not match';
			return;
		}

		isLoading = true;
		error = '';

		try {
			const response = await authApi.register({ name, email, password });
			authStore.setUser(response.user);
			goto('/dashboard');
		} catch (err) {
			const apiError = err as ApiError;
			error = apiError.message || 'Registration failed. Please try again.';
		} finally {
			isLoading = false;
		}
	}

	$effect(() => {
		if (name || email || password || confirmPassword) {
			error = '';
		}
	});
</script>

<svelte:head>
	<title>Sign Up - mikrocloud</title>
	<meta name="description" content="Create your mikrocloud account" />
</svelte:head>

<div class="min-h-screen flex flex-col py-12 px-[85px]">
	<header class="flex items-center justify-between">
		<div class="flex items-center">
			<Logo class="h-[25px]" />
		</div>
		<button onclick={() => goto('/login')} class="btn btn-outline">Login</button>
	</header>

	<div class="flex-1 flex">
		<div class="flex-1 flex flex-col justify-center max-w-md">
			<div class="mb-8">
				<h1 class="text-xl font-bold text-white mb-2">Self Hosting made easy</h1>
				<p class="text-muted-foreground text-base">
					Deploy apps with ease and keep full control of your data and environment.
				</p>
			</div>

			<form onsubmit={handleSubmit} class="space-y-5">
				{#if error}
					<div class="text-red-500 text-sm">{error}</div>
				{/if}

				<fieldset class="fieldset">
					<legend class="fieldset-legend">Your name</legend>
					<input id="name" type="text" bind:value={name} required autocomplete="name" class="input input-bordered w-full" />
				</fieldset>

				<fieldset class="fieldset">
					<legend class="fieldset-legend">Email</legend>
					<input
						id="email"
						type="email"
						bind:value={email}
						placeholder="example@gmail.com"
						required
						autocomplete="email"
						class="input input-bordered w-full"
					/>
				</fieldset>

				<fieldset class="fieldset">
					<legend class="fieldset-legend">Password</legend>
					<input
						id="password"
						type="password"
						bind:value={password}
						required
						autocomplete="new-password"
						class="input input-bordered w-full"
					/>
				</fieldset>

				<fieldset class="fieldset">
					<legend class="fieldset-legend">Confirm Password</legend>
					<input
						id="confirmPassword"
						type="password"
						bind:value={confirmPassword}
						required
						autocomplete="new-password"
						class="input input-bordered w-full"
					/>
				</fieldset>

				<button
					type="submit"
					class="w-full btn"
					disabled={isLoading}
				>
					{#if isLoading}
						<div class="flex items-center">
							<div class="animate-spin rounded-full h-4 w-4 border-b-2 border-black mr-2"></div>
							Creating account...
						</div>
					{:else}
						Sign up
					{/if}
					</button>
			</form>

			<!-- <div class="mt-6 space-y-3"> -->
			<!-- 	<Button -->
			<!-- 		variant="outline" -->
			<!-- 		class="w-full border-white/20 text-gray-400 hover:bg-white/5 hover:text-white" -->
			<!-- 	> -->
			<!-- 		<svg class="w-4 h-4 mr-2" viewBox="0 0 24 24"> -->
			<!-- 			<path -->
			<!-- 				fill="currentColor" -->
			<!-- 				d="M22.56 12.25c0-.78-.07-1.53-.2-2.25H12v4.26h5.92c-.26 1.37-1.04 2.53-2.21 3.31v2.77h3.57c2.08-1.92 3.28-4.74 3.28-8.09z" -->
			<!-- 			/> -->
			<!-- 			<path -->
			<!-- 				fill="currentColor" -->
			<!-- 				d="M12 23c2.97 0 5.46-.98 7.28-2.66l-3.57-2.77c-.98.66-2.23 1.06-3.71 1.06-2.86 0-5.29-1.93-6.16-4.53H2.18v2.84C3.99 20.53 7.7 23 12 23z" -->
			<!-- 			/> -->
			<!-- 			<path -->
			<!-- 				fill="currentColor" -->
			<!-- 				d="M5.84 14.09c-.22-.66-.35-1.36-.35-2.09s.13-1.43.35-2.09V7.07H2.18C1.43 8.55 1 10.22 1 12s.43 3.45 1.18 4.93l2.85-2.22.81-.62z" -->
			<!-- 			/> -->
			<!-- 			<path -->
			<!-- 				fill="currentColor" -->
			<!-- 				d="M12 5.38c1.62 0 3.06.56 4.21 1.64l3.15-3.15C17.45 2.09 14.97 1 12 1 7.7 1 3.99 3.47 2.18 7.07l3.66 2.84c.87-2.6 3.3-4.53 6.16-4.53z" -->
			<!-- 			/> -->
			<!-- 		</svg> -->
			<!-- 		Continue with google -->
			<!-- 	</Button> -->
			<!---->
			<!-- 	<Button -->
			<!-- 		variant="outline" -->
			<!-- 		class="w-full border-white/20 text-gray-400 hover:bg-white/5 hover:text-white" -->
			<!-- 	> -->
			<!-- 		<svg class="w-5 h-5 mr-2" fill="currentColor" viewBox="0 0 24 24"> -->
			<!-- 			<path -->
			<!-- 				d="M12 0c-6.626 0-12 5.373-12 12 0 5.302 3.438 9.8 8.207 11.387.599.111.793-.261.793-.577v-2.234c-3.338.726-4.033-1.416-4.033-1.416-.546-1.387-1.333-1.756-1.333-1.756-1.089-.745.083-.729.083-.729 1.205.084 1.839 1.237 1.839 1.237 1.07 1.834 2.807 1.304 3.492.997.107-.775.418-1.305.762-1.604-2.665-.305-5.467-1.334-5.467-5.931 0-1.311.469-2.381 1.236-3.221-.124-.303-.535-1.524.117-3.176 0 0 1.008-.322 3.301 1.23.957-.266 1.983-.399 3.003-.404 1.02.005 2.047.138 3.006.404 2.291-1.552 3.297-1.23 3.297-1.23.653 1.653.242 2.874.118 3.176.77.84 1.235 1.911 1.235 3.221 0 4.609-2.807 5.624-5.479 5.921.43.372.823 1.102.823 2.222v3.293c0 .319.192.694.801.576 4.765-1.589 8.199-6.086 8.199-11.386 0-6.627-5.373-12-12-12z" -->
			<!-- 			/> -->
			<!-- 		</svg> -->
			<!-- 		Continue with github -->
			<!-- 	</Button> -->
			<!---->
			<!-- 	<div class="grid grid-cols-2 gap-3"> -->
			<!-- 		<Button -->
			<!-- 			variant="outline" -->
			<!-- 			class="border-white/20 text-gray-400 hover:bg-white/5 hover:text-white" -->
			<!-- 		> -->
			<!-- 			<svg class="w-5 h-5" fill="#FF6B35" viewBox="0 0 24 24"> -->
			<!-- 				<path -->
			<!-- 					d="M23.546 10.93L13.067.452c-.604-.603-1.582-.603-2.188 0L8.708 2.627l2.76 2.76c.645-.215 1.379-.07 1.889.441.516.515.658 1.258.438 1.9l2.658 2.66c.645-.223 1.387-.078 1.9.435.721.72.721 1.884 0 2.604-.719.719-1.881.719-2.6 0-.539-.541-.674-1.337-.404-1.996L12.86 8.955v6.525c.176.086.342.203.488.348.713.721.713 1.883 0 2.6-.719.721-1.889.721-2.609 0-.719-.719-.719-1.879 0-2.598.182-.18.387-.316.605-.406V8.835c-.217-.091-.424-.222-.6-.401-.545-.545-.676-1.342-.396-2.009L7.636 3.7.45 10.881c-.6.605-.6 1.584 0 2.189l10.48 10.477c.604.604 1.582.604 2.186 0l10.43-10.43c.605-.603.605-1.582 0-2.187" -->
			<!-- 				/> -->
			<!-- 			</svg> -->
			<!-- 		</Button> -->
			<!-- 		<Button -->
			<!-- 			variant="outline" -->
			<!-- 			class="border-white/20 text-gray-400 hover:bg-white/5 hover:text-white" -->
			<!-- 		> -->
			<!-- 			<svg class="w-5 h-5" fill="#1E8FE1" viewBox="0 0 24 24"> -->
			<!-- 				<path -->
			<!-- 					d="M23.546 10.93L13.067.452c-.604-.603-1.582-.603-2.188 0L8.708 2.627l2.76 2.76c.645-.215 1.379-.07 1.889.441.516.515.658 1.258.438 1.9l2.658 2.66c.645-.223 1.387-.078 1.9.435.721.72.721 1.884 0 2.604-.719.719-1.881.719-2.6 0-.539-.541-.674-1.337-.404-1.996L12.86 8.955v6.525c.176.086.342.203.488.348.713.721.713 1.883 0 2.6-.719.721-1.889.721-2.609 0-.719-.719-.719-1.879 0-2.598.182-.18.387-.316.605-.406V8.835c-.217-.091-.424-.222-.6-.401-.545-.545-.676-1.342-.396-2.009L7.636 3.7.45 10.881c-.6.605-.6 1.584 0 2.189l10.48 10.477c.604.604 1.582.604 2.186 0l10.43-10.43c.605-.603.605-1.582 0-2.187" -->
			<!-- 				/> -->
			<!-- 			</svg> -->
			<!-- 		</Button> -->
			<!-- 	</div> -->
			<!-- </div> -->
		</div>

		<div class="hidden lg:block flex-1 relative -mx-[85px] pointer-events-none">
			<div class="absolute inset-0 flex items-center justify-end">
				<img
					src="/dashboard-preview.png"
					alt="Dashboard Preview"
					class="object-contain w-[750px]"
				/>
			</div>
		</div>
	</div>
</div>
