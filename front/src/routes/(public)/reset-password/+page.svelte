<!-- Copyright (c) 2025 Start Codex SAS. All rights reserved. -->
<!-- SPDX-License-Identifier: BUSL-1.1 -->

<script lang="ts">
	import { goto } from '$app/navigation';
	import type { PageData } from './$types';
	import { auth, ApiError } from '$lib/api';
	import { Button } from '$lib/components/ui/button/index.js';
	import * as Card from '$lib/components/ui/card/index.js';
	import { Field, FieldGroup, FieldLabel } from '$lib/components/ui/field/index.js';
	import { Input } from '$lib/components/ui/input/index.js';
	import tooklyLogo from '$lib/assets/tookly-logo.svg';
	import * as m from '$lib/paraglide/messages';
	import { i18n } from '$lib/i18n.svelte';

	let { data }: { data: PageData } = $props();

	const t = $derived.by(() => {
		i18n.locale;
		return {
			title: m.reset_title(),
			description: m.reset_description(),
			newPassword: m.reset_new_password(),
			confirmPassword: m.reset_confirm_password(),
			submit: m.reset_submit(),
			submitting: m.reset_submitting(),
			mismatch: m.reset_mismatch(),
			tooShort: m.reset_too_short(),
			invalidToken: m.reset_invalid_token(),
			backToLogin: m.forgot_back_to_login()
		};
	});

	let newPassword = $state('');
	let confirmPassword = $state('');
	let loading = $state(false);
	let error = $state('');

	const hasToken = $derived(!!data.token);

	async function handleSubmit(e: SubmitEvent) {
		e.preventDefault();
		error = '';

		if (newPassword !== confirmPassword) {
			error = t.mismatch;
			return;
		}
		if (newPassword.length < 8) {
			error = t.tooShort;
			return;
		}

		loading = true;
		try {
			await auth.resetPassword({ token: data.token, new_password: newPassword });
			goto('/login?reset=success');
		} catch (err) {
			if (err instanceof ApiError && err.status === 400) {
				error = t.invalidToken;
			} else if (err instanceof ApiError && err.status === 422) {
				error = t.tooShort;
			} else {
				error = err instanceof Error ? err.message : 'Failed to reset password';
			}
		} finally {
			loading = false;
		}
	}
</script>

<svelte:head><title>Reset Password — Tookly</title></svelte:head>

<div class="flex min-h-svh flex-col items-center justify-center gap-6 p-6 md:p-10">
	<div class="flex w-full max-w-sm flex-col gap-6">
		<a href="/login" class="flex items-center gap-3 self-center font-heading text-2xl font-black">
			<img src={tooklyLogo} alt="Tookly" class="size-11" />
			Tookly
		</a>
		<Card.Root>
			<Card.Header class="text-center">
				<Card.Title class="text-xl">{t.title}</Card.Title>
				<Card.Description>{t.description}</Card.Description>
			</Card.Header>
			<Card.Content>
				{#if !hasToken}
					<div class="space-y-4">
						<p class="text-sm text-destructive">{t.invalidToken}</p>
						<a href="/forgot-password" class="text-sm underline-offset-4 hover:underline text-primary">{t.backToLogin}</a>
					</div>
				{:else}
					<form onsubmit={handleSubmit}>
						<FieldGroup>
							<Field>
								<FieldLabel for="new-pw">{t.newPassword}</FieldLabel>
								<Input id="new-pw" type="password" required bind:value={newPassword} />
							</Field>
							<Field>
								<FieldLabel for="confirm-pw">{t.confirmPassword}</FieldLabel>
								<Input id="confirm-pw" type="password" required bind:value={confirmPassword} />
							</Field>
							{#if error}
								<p class="text-destructive text-sm">{error}</p>
							{/if}
							<Field>
								<Button type="submit" class="w-full" disabled={loading || !newPassword || !confirmPassword}>
									{loading ? t.submitting : t.submit}
								</Button>
							</Field>
						</FieldGroup>
					</form>
				{/if}
			</Card.Content>
		</Card.Root>
	</div>
</div>
