<!-- Copyright (c) 2025 Start Codex SAS. All rights reserved. -->
<!-- SPDX-License-Identifier: BUSL-1.1 -->

<script lang="ts">
	import { goto } from '$app/navigation';
	import { Button } from '$lib/components/ui/button/index.js';
	import * as Alert from '$lib/components/ui/alert/index.js';
	import * as Card from '$lib/components/ui/card/index.js';
	import { Input } from '$lib/components/ui/input/index.js';
	import XIcon from '@lucide/svelte/icons/x';
	import { instance } from '$lib/api';
	import { login } from '$lib/stores/auth';
	import * as m from '$lib/paraglide/messages';
	import { i18n } from '$lib/i18n.svelte';
	import tooklyLogo from '$lib/assets/tookly-logo.svg';

	const t = $derived.by(() => {
		i18n.locale;
		return {
			title: m.setup_title(),
			description: m.setup_description(),
			name: m.setup_name(),
			email: m.setup_email(),
			password: m.setup_password(),
			confirmPassword: m.setup_confirm_password(),
			submit: m.setup_submit(),
			creating: m.setup_creating(),
			mismatch: m.setup_passwords_mismatch(),
			error: m.setup_error()
		};
	});

	let name = $state('');
	let email = $state('');
	let password = $state('');
	let confirmPassword = $state('');
	let errorMessage = $state('');
	let loading = $state(false);

	async function handleSubmit(e: SubmitEvent) {
		e.preventDefault();
		errorMessage = '';

		if (password !== confirmPassword) {
			errorMessage = t.mismatch;
			return;
		}

		loading = true;
		try {
			const user = await instance.bootstrap({ email, name, password });
			login(user);
			goto('/');
		} catch (err) {
			errorMessage = err instanceof Error ? err.message : t.error;
		} finally {
			loading = false;
		}
	}
</script>

<svelte:head><title>{t.title} — Tookly</title></svelte:head>

<div class="flex min-h-svh flex-col items-center justify-center gap-6 p-6 md:p-10">
	<div class="flex w-full max-w-sm flex-col gap-6">
		<a href="/setup" class="flex items-center gap-3 self-center font-heading text-2xl font-black">
			<img src={tooklyLogo} alt="Tookly" class="size-11" />
			Tookly
		</a>
		<Card.Root>
			<Card.Header class="text-center">
				<Card.Title class="font-heading text-xl">{t.title}</Card.Title>
				<Card.Description>{t.description}</Card.Description>
			</Card.Header>
			<Card.Content>
				<form onsubmit={handleSubmit} class="flex flex-col gap-4">
					<div class="flex flex-col gap-1.5">
						<label for="setup-name" class="font-heading text-sm font-bold uppercase tracking-wider">{t.name}</label>
						<Input id="setup-name" type="text" placeholder="Admin" required bind:value={name} />
					</div>
					<div class="flex flex-col gap-1.5">
						<label for="setup-email" class="font-heading text-sm font-bold uppercase tracking-wider">{t.email}</label>
						<Input id="setup-email" type="email" placeholder="admin@example.com" required bind:value={email} />
					</div>
					<div class="flex flex-col gap-1.5">
						<label for="setup-password" class="font-heading text-sm font-bold uppercase tracking-wider">{t.password}</label>
						<Input id="setup-password" type="password" required bind:value={password} />
					</div>
					<div class="flex flex-col gap-1.5">
						<label for="setup-confirm" class="font-heading text-sm font-bold uppercase tracking-wider">{t.confirmPassword}</label>
						<Input id="setup-confirm" type="password" required bind:value={confirmPassword} />
					</div>
					{#if errorMessage}
						<Alert.Root variant="destructive">
							<XIcon />
							<Alert.Description>{errorMessage}</Alert.Description>
						</Alert.Root>
					{/if}
					<Button type="submit" class="w-full" size="lg"  disabled={loading}>
						{loading ? t.creating : t.submit}
					</Button>
				</form>
			</Card.Content>
		</Card.Root>
	</div>
</div>
