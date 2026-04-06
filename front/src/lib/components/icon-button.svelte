<!-- Copyright (c) 2025 Start Codex SAS. All rights reserved. -->
<!-- SPDX-License-Identifier: BUSL-1.1 -->

<script lang="ts">
	import * as Tooltip from '$lib/components/ui/tooltip/index.js';
	import type { Snippet } from 'svelte';

	let {
		href,
		onclick,
		tooltip,
		tooltipSide = 'right',
		children
	}: {
		href?: string;
		onclick?: () => void;
		tooltip?: string;
		tooltipSide?: 'top' | 'right' | 'bottom' | 'left';
		children: Snippet;
	} = $props();
</script>

{#if tooltip}
	<Tooltip.Root>
		<Tooltip.Trigger>
			{#if href}
				<a {href} class="flex items-center justify-center p-1 rounded-md border-2 border-border bg-white shadow-[2px_3px_0px_#000] hover:bg-accent">
					{@render children()}
				</a>
			{:else}
				<button type="button" {onclick} class="flex items-center justify-center p-1 rounded-md border-2 border-border bg-white shadow-[2px_3px_0px_#000] hover:bg-accent cursor-pointer">
					{@render children()}
				</button>
			{/if}
		</Tooltip.Trigger>
		<Tooltip.Content side={tooltipSide}>
			<p>{tooltip}</p>
		</Tooltip.Content>
	</Tooltip.Root>
{:else if href}
	<a {href} class="flex items-center justify-center p-1 rounded-md border-2 border-border bg-white shadow-[2px_3px_0px_#000] hover:bg-accent">
		{@render children()}
	</a>
{:else}
	<button type="button" {onclick} class="flex items-center justify-center p-1 rounded-md border-2 border-border bg-white shadow-[2px_3px_0px_#000] hover:bg-accent cursor-pointer">
		{@render children()}
	</button>
{/if}
