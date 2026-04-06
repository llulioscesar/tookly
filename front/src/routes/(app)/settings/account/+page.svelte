<!-- Copyright (c) 2025 Start Codex SAS. All rights reserved. -->
<!-- SPDX-License-Identifier: BUSL-1.1 -->

<script lang="ts">
	import * as Card from '$lib/components/ui/card/index.js';
	import ChangePasswordForm from '$lib/components/change-password-form.svelte';
	import { currentUser } from '$lib/stores/auth';
	import * as m from '$lib/paraglide/messages';
	import { i18n } from '$lib/i18n.svelte';

	const title = $derived.by(() => { i18n.locale; return m.settings_account_title(); });
	const t = $derived.by(() => {
		i18n.locale;
		return {
			noPassword: m.settings_no_password(),
			setPassword: m.settings_set_password()
		};
	});
</script>

<svelte:head><title>Account — Tookly</title></svelte:head>

<div class="space-y-6">
	<h2 class="font-heading text-lg font-bold uppercase tracking-wider">{title}</h2>

	{#if $currentUser?.has_password === false}
		<Card.Root>
			<Card.Content class="pt-6">
				<p class="text-sm text-muted-foreground">{t.noPassword}</p>
				<a href="/forgot-password" class="mt-2 inline-block text-sm font-bold text-primary underline-offset-4 hover:underline">
					{t.setPassword}
				</a>
			</Card.Content>
		</Card.Root>
	{:else}
		<ChangePasswordForm />
	{/if}
</div>
