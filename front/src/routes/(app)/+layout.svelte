<!-- Copyright (c) 2025 Start Codex SAS. All rights reserved. -->
<!-- SPDX-License-Identifier: BUSL-1.1 -->

<script lang="ts">
	import { page } from '$app/state';
	import type { LayoutData } from './$types';
	import { currentUser, logout } from '$lib/stores/auth';
	import { auth } from '$lib/api';
	import { goto } from '$app/navigation';
	import { Button } from '$lib/components/ui/button/index.js';
	import * as DropdownMenu from '$lib/components/ui/dropdown-menu/index.js';
	import LogOutIcon from '@lucide/svelte/icons/log-out';
	import SettingsIcon from '@lucide/svelte/icons/settings';
	import ShieldIcon from '@lucide/svelte/icons/shield';
	import EllipsisVerticalIcon from '@lucide/svelte/icons/ellipsis-vertical';
	import tooklyLogo from '$lib/assets/tookly-logo.svg';
	import * as m from '$lib/paraglide/messages';
	import { i18n } from '$lib/i18n.svelte';

	let { children, data }: { children: any; data: LayoutData } = $props();

	const isWorkspaceRoute = $derived(!!page.params.workspace);
	const isSelfContainedRoute = $derived(
		page.url.pathname.startsWith('/admin') || page.url.pathname.startsWith('/settings')
	);

	const showVerifyBanner = $derived(
		!!$currentUser &&
		!$currentUser.email_verified_at &&
		data.emailVerificationRequired
	);

	const t = $derived.by(() => {
		i18n.locale;
		return {
			verifyBanner: m.verify_banner(),
			resend: m.verify_resend(),
			resent: m.verify_resent(),
			resendError: m.verify_resend_error(),
			admin: m.nav_admin(),
			settings: m.settings_nav(),
			logout: m.nav_logout()
		};
	});

	let resending = $state(false);
	let resendSuccess = $state(false);
	let resendError = $state('');

	async function handleResend() {
		resending = true;
		resendSuccess = false;
		resendError = '';
		try {
			await auth.resendVerification();
			resendSuccess = true;
			setTimeout(() => { resendSuccess = false; }, 5000);
		} catch {
			resendError = t.resendError;
		} finally {
			resending = false;
		}
	}
</script>

<!-- Verification banner (above everything) -->
{#if showVerifyBanner}
	<div class="bg-yellow-50 border-b border-yellow-200 px-4 py-2 text-center text-sm text-yellow-800">
		{t.verifyBanner}
		<button
			class="ml-2 underline underline-offset-4 hover:text-yellow-900"
			onclick={handleResend}
			disabled={resending}
		>
			{resending ? '...' : t.resend}
		</button>
		{#if resendSuccess}
			<span class="ml-2 text-green-700">{t.resent}</span>
		{/if}
		{#if resendError}
			<span class="ml-2 text-red-700">{resendError}</span>
		{/if}
	</div>
{/if}

{#if isWorkspaceRoute || isSelfContainedRoute}
	{@render children()}
{:else}
	<div class="bg-background min-h-screen">
		<header class="border-b-[3px] border-border bg-input">
			<div class="mx-auto flex h-14 max-w-screen-xl items-center justify-between px-6">
				<a href="/" class="flex items-center gap-2 font-heading text-lg font-black">
					<img src={tooklyLogo} alt="Tookly" class="size-8" />
					Tookly
				</a>
				{#if $currentUser}
					<div class="flex items-center gap-2">
						<span class="font-heading text-sm font-bold">{$currentUser.email}</span>
						<DropdownMenu.Root>
							<DropdownMenu.Trigger>
								{#snippet child({ props })}
									<button {...props} class="rounded-lg border-2 border-border p-1.5 hover:bg-accent">
										<EllipsisVerticalIcon class="size-4" />
									</button>
								{/snippet}
							</DropdownMenu.Trigger>
						<DropdownMenu.Content class="w-56" align="end">
							<DropdownMenu.Group>
								{#if $currentUser.is_instance_admin}
									<DropdownMenu.Item onSelect={() => goto('/admin')}>
										<ShieldIcon />
										{t.admin}
									</DropdownMenu.Item>
								{/if}
								<DropdownMenu.Item onSelect={() => goto('/settings')}>
									<SettingsIcon />
									{t.settings}
								</DropdownMenu.Item>
							</DropdownMenu.Group>
							<DropdownMenu.Separator />
							<DropdownMenu.Item onSelect={() => logout()}>
								<LogOutIcon />
								{t.logout}
							</DropdownMenu.Item>
						</DropdownMenu.Content>
						</DropdownMenu.Root>
					</div>
				{/if}
			</div>
		</header>
		{@render children()}
	</div>
{/if}
