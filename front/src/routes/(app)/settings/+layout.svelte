<!-- Copyright (c) 2025 Start Codex SAS. All rights reserved. -->
<!-- SPDX-License-Identifier: BUSL-1.1 -->

<script lang="ts">
	import { page } from '$app/state';
	import SlidersIcon from '@lucide/svelte/icons/sliders-horizontal';
	import UserIcon from '@lucide/svelte/icons/user';
	import ArrowLeftIcon from '@lucide/svelte/icons/arrow-left';
	import IconButton from '$lib/components/icon-button.svelte';
	import * as m from '$lib/paraglide/messages';
	import { i18n } from '$lib/i18n.svelte';

	let { children } = $props();

	const navItems = $derived.by(() => {
		i18n.locale;
		return [
			{ href: '/settings/account',     label: m.settings_nav_account(),     icon: UserIcon     },
			{ href: '/settings/preferences', label: m.settings_nav_preferences(), icon: SlidersIcon }
		];
	});

	const title = $derived.by(() => { i18n.locale; return m.settings_title(); });
	const dashboard = $derived.by(() => { i18n.locale; return m.nav_dashboard(); });
</script>

<div class="flex min-h-screen">
	<!-- Sidebar -->
	<aside class="flex w-64 shrink-0 flex-col text-black" style="background-color: #AAC8FF;">
		<!-- Back -->
		<div class="px-4 pt-4 pb-2">
			<IconButton href="/" tooltip={dashboard}>
				<ArrowLeftIcon class="size-6 shrink-0" />
			</IconButton>
		</div>

		<!-- Nav -->
		<div class="px-3 pt-2 pb-4">
			<p class="mb-2 px-3 text-xs font-bold uppercase tracking-wider text-black">
				{title}
			</p>
			<nav class="flex flex-col gap-1">
				{#each navItems as item}
					{@const active = page.url.pathname === item.href}
					<a
						href={item.href}
						class="font-heading flex items-center gap-2.5 rounded-lg border-2 px-3 py-2 text-sm font-bold transition-colors
							{active
								? 'border-black bg-accent text-black shadow-[2px_3px_0px_#000]'
								: 'border-transparent text-black hover:text-black'}"
					>
						<item.icon class="size-4 shrink-0" />
						{item.label}
					</a>
				{/each}
			</nav>
		</div>

	</aside>

	<!-- Content -->
	<main class="min-w-0 flex-1 bg-background p-8">
		<div class="mx-auto max-w-[760px]">
			{@render children()}
		</div>
	</main>
</div>
