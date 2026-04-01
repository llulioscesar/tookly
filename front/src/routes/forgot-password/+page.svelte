<!-- Copyright (c) 2025 Start Codex SAS. All rights reserved. -->
<!-- SPDX-License-Identifier: BUSL-1.1 -->

<script lang="ts">
	import { auth } from '$lib/api';
	import { Button } from '$lib/components/ui/button/index.js';
	import * as Card from '$lib/components/ui/card/index.js';
	import { Field, FieldGroup, FieldLabel } from '$lib/components/ui/field/index.js';
	import { Input } from '$lib/components/ui/input/index.js';
	import * as m from '$lib/paraglide/messages';
	import { i18n } from '$lib/i18n.svelte';

	const t = $derived.by(() => {
		i18n.locale;
		return {
			title: m.forgot_title(),
			description: m.forgot_description(),
			email: m.forgot_email(),
			submit: m.forgot_submit(),
			sending: m.forgot_sending(),
			sent: m.forgot_sent(),
			backToLogin: m.forgot_back_to_login()
		};
	});

	let email = $state('');
	let loading = $state(false);
	let sent = $state(false);

	async function handleSubmit(e: SubmitEvent) {
		e.preventDefault();
		loading = true;
		try {
			await auth.forgotPassword({ email });
		} catch {
			// Always show success — no enumeration
		} finally {
			loading = false;
			sent = true;
		}
	}
</script>

<div class="bg-muted flex min-h-svh flex-col items-center justify-center gap-6 p-6 md:p-10">
	<div class="flex w-full max-w-sm flex-col gap-6">
		<a href="/login" class="flex items-center gap-2 self-center font-medium">
			<div class="bg-primary text-primary-foreground flex size-6 items-center justify-center rounded-md">
				<svg xmlns="http://www.w3.org/2000/svg" class="size-4" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
					<rect width="8" height="8" x="2" y="2" rx="2" /><rect width="8" height="8" x="14" y="2" rx="2" />
					<rect width="8" height="8" x="2" y="14" rx="2" /><rect width="8" height="8" x="14" y="14" rx="2" />
				</svg>
			</div>
			Tookly
		</a>
		<Card.Root>
			<Card.Header class="text-center">
				<Card.Title class="text-xl">{t.title}</Card.Title>
				<Card.Description>{t.description}</Card.Description>
			</Card.Header>
			<Card.Content>
				{#if sent}
					<div class="space-y-4">
						<p class="text-sm text-muted-foreground">{t.sent}</p>
						<a href="/login" class="text-sm underline-offset-4 hover:underline text-primary">{t.backToLogin}</a>
					</div>
				{:else}
					<form onsubmit={handleSubmit}>
						<FieldGroup>
							<Field>
								<FieldLabel for="forgot-email">{t.email}</FieldLabel>
								<Input id="forgot-email" type="email" placeholder="m@example.com" required bind:value={email} />
							</Field>
							<Field>
								<Button type="submit" class="w-full" disabled={loading || !email}>
									{loading ? t.sending : t.submit}
								</Button>
							</Field>
						</FieldGroup>
					</form>
					<div class="mt-4 text-center">
						<a href="/login" class="text-sm underline-offset-4 hover:underline text-muted-foreground">{t.backToLogin}</a>
					</div>
				{/if}
			</Card.Content>
		</Card.Root>
	</div>
</div>
