<!-- Copyright (c) 2025 Start Codex SAS. All rights reserved. -->
<!-- SPDX-License-Identifier: BUSL-1.1 -->

<script lang="ts">
	import type { PageData } from './$types';
	import tooklyLogo from '$lib/assets/tookly-logo.svg';
	import * as Card from '$lib/components/ui/card/index.js';
	import * as m from '$lib/paraglide/messages';
	import { i18n } from '$lib/i18n.svelte';

	let { data }: { data: PageData } = $props();

	const t = $derived.by(() => {
		i18n.locale;
		return {
			success: m.verify_success(),
			invalid: m.verify_invalid(),
			error: m.verify_error()
		};
	});
</script>

<svelte:head><title>Verify Email — Tookly</title></svelte:head>

<div class="flex min-h-svh flex-col items-center justify-center gap-6 p-6 md:p-10">
	<div class="flex w-full max-w-sm flex-col gap-6">
		<a href="/" class="flex items-center gap-3 self-center font-heading text-2xl font-black">
			<img src={tooklyLogo} alt="Tookly" class="size-11" />
			Tookly
		</a>
		<Card.Root>
			<Card.Content class="py-8 text-center">
				{#if data.status === 'success'}
					<p class="text-sm text-green-700">{t.success}</p>
					<a href="/" class="mt-4 inline-block text-sm text-primary underline-offset-4 hover:underline">
						Continue to Tookly
					</a>
				{:else if data.status === 'invalid'}
					<p class="text-sm text-destructive">{t.invalid}</p>
				{:else}
					<p class="text-sm text-destructive">{t.error}</p>
				{/if}
			</Card.Content>
		</Card.Root>
	</div>
</div>
