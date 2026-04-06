<!-- Copyright (c) 2025 Start Codex SAS. All rights reserved. -->
<!-- SPDX-License-Identifier: BUSL-1.1 -->

<script lang="ts">
	import './layout.css';
	import favicon from '$lib/assets/favicon.svg';
	import { Toaster } from '$lib/components/ui/sonner/index.js';
	import * as Tooltip from '$lib/components/ui/tooltip/index.js';
	import { onMount } from 'svelte';
	import { initLocale } from '$lib/i18n.svelte';

	let { children } = $props();

	onMount(() => {
		initLocale();
	});
</script>

<svelte:head><link rel="icon" href={favicon} /></svelte:head>
<Toaster />
<Tooltip.Provider>
	{@render children()}
</Tooltip.Provider>
