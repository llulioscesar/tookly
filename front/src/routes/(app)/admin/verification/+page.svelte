<!-- Copyright (c) 2025 Start Codex SAS. All rights reserved. -->
<!-- SPDX-License-Identifier: BUSL-1.1 -->

<script lang="ts">
	import type { PageData } from './$types';
	import { toast } from 'svelte-sonner';
	import { instance } from '$lib/api';
	import * as Item from '$lib/components/ui/item/index.js';
	import { Switch } from '$lib/components/ui/switch/index.js';
	import * as m from '$lib/paraglide/messages';
	import { i18n } from '$lib/i18n.svelte';

	let { data }: { data: PageData } = $props();

	const t = $derived.by(() => {
		i18n.locale;
		return {
			title: m.admin_verification_title(),
			toggle: m.admin_verification_toggle(),
			enabled: m.admin_verification_enabled(),
			disabled: m.admin_verification_disabled()
		};
	});

	let required = $state(false);
	let saving = $state(false);

	$effect(() => { required = data.verificationRequired; });

	async function handleToggle() {
		saving = true;
		try {
			required = !required;
			await instance.verification.save({ required });
			toast.success(m.toast_saved());
		} catch {
			required = !required;
			toast.error(m.toast_error());
		} finally {
			saving = false;
		}
	}
</script>

<svelte:head><title>Verification — Tookly</title></svelte:head>

<div class="space-y-6">
	<h2 class="font-heading text-lg font-bold uppercase tracking-wider">{t.title}</h2>

	<Item.Group>
		<Item.Root variant="outline">
			<Item.Content>
				<Item.Title>{t.toggle}</Item.Title>
				<Item.Description>{required ? t.enabled : t.disabled}</Item.Description>
			</Item.Content>
			<Item.Actions>
				<Switch checked={required} onCheckedChange={handleToggle} disabled={saving} />
			</Item.Actions>
		</Item.Root>
	</Item.Group>
</div>
