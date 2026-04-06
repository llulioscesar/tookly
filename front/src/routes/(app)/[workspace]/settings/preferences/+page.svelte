<!-- Copyright (c) 2025 Start Codex SAS. All rights reserved. -->
<!-- SPDX-License-Identifier: BUSL-1.1 -->

<script lang="ts">
	import { tick } from 'svelte';
	import CheckIcon from '@lucide/svelte/icons/check';
	import ChevronsUpDownIcon from '@lucide/svelte/icons/chevrons-up-down';
	import * as Item from '$lib/components/ui/item/index.js';
	import * as Command from '$lib/components/ui/command/index.js';
	import * as Popover from '$lib/components/ui/popover/index.js';
	import { Button } from '$lib/components/ui/button/index.js';
	import { cn } from '$lib/utils.js';
	import * as m from '$lib/paraglide/messages';
	import { i18n, switchLocale } from '$lib/i18n.svelte';
	import type { Locale } from '$lib/paraglide/runtime';

	const t = $derived.by(() => {
		i18n.locale;
		return {
			title: m.settings_preferences(),
			languageTitle: m.settings_language_title(),
			languageDesc: m.settings_language_desc(),
			themeTitle: m.settings_theme_title(),
			themeDesc: m.settings_theme_desc(),
			themeLight: m.settings_theme_light(),
			themeDark: m.settings_theme_dark(),
			themeAuto: m.settings_theme_auto()
		};
	});

	const languages: { value: string; label: string }[] = [
		{ value: 'en', label: 'English' },
		{ value: 'es', label: 'Español' }
	];

	let langOpen = $state(false);
	let langTriggerRef = $state<HTMLButtonElement>(null!);

	const selectedLangLabel = $derived(
		languages.find((l) => l.value === i18n.locale)?.label ?? ''
	);

	function closeLang() {
		langOpen = false;
		tick().then(() => langTriggerRef.focus());
	}

	// Theme
	type Theme = 'light' | 'dark' | 'auto';

	function getStoredTheme(): Theme {
		if (typeof localStorage === 'undefined') return 'auto';
		return (localStorage.getItem('theme') as Theme) || 'auto';
	}

	function applyTheme(theme: Theme) {
		if (theme === 'dark') {
			document.documentElement.classList.add('dark');
		} else if (theme === 'light') {
			document.documentElement.classList.remove('dark');
		} else {
			if (window.matchMedia('(prefers-color-scheme: dark)').matches) {
				document.documentElement.classList.add('dark');
			} else {
				document.documentElement.classList.remove('dark');
			}
		}
		localStorage.setItem('theme', theme);
	}

	let currentTheme = $state<Theme>(getStoredTheme());

	const themes: { value: Theme; label: string }[] = $derived([
		{ value: 'light', label: t.themeLight },
		{ value: 'dark', label: t.themeDark },
		{ value: 'auto', label: t.themeAuto }
	]);

	let themeOpen = $state(false);
	let themeTriggerRef = $state<HTMLButtonElement>(null!);

	const selectedThemeLabel = $derived(
		themes.find((th) => th.value === currentTheme)?.label ?? ''
	);

	function closeTheme() {
		themeOpen = false;
		tick().then(() => themeTriggerRef.focus());
	}
</script>

<svelte:head><title>Preferences — Tookly</title></svelte:head>

<div class="space-y-6">
	<h2 class="font-heading text-lg font-bold uppercase tracking-wider">{t.title}</h2>

	<Item.Group>
		<Item.Root variant="outline">
			<Item.Content>
				<Item.Title>{t.languageTitle}</Item.Title>
				<Item.Description>{t.languageDesc}</Item.Description>
			</Item.Content>
			<Item.Actions>
				<Popover.Root bind:open={langOpen}>
					<Popover.Trigger bind:ref={langTriggerRef}>
						{#snippet child({ props })}
							<Button
								{...props}
								class="w-[160px] justify-between border-2 border-transparent bg-[var(--input)] font-normal hover:border-border hover:bg-white focus-visible:bg-white"
								role="combobox"
								aria-expanded={langOpen}
							>
								{selectedLangLabel}
								<ChevronsUpDownIcon class="opacity-50" />
							</Button>
						{/snippet}
					</Popover.Trigger>
					<Popover.Content class="w-[160px] p-0" align="end">
						<Command.Root>
							<Command.List>
								<Command.Group>
									{#each languages as lang (lang.value)}
										<Command.Item
											value={lang.value}
											onSelect={() => {
												switchLocale(lang.value as Locale);
												closeLang();
											}}
										>
											<CheckIcon
												class={cn(i18n.locale !== lang.value && "text-transparent")}
											/>
											{lang.label}
										</Command.Item>
									{/each}
								</Command.Group>
							</Command.List>
						</Command.Root>
					</Popover.Content>
				</Popover.Root>
			</Item.Actions>
		</Item.Root>

		<Item.Root variant="outline">
			<Item.Content>
				<Item.Title>{t.themeTitle}</Item.Title>
				<Item.Description>{t.themeDesc}</Item.Description>
			</Item.Content>
			<Item.Actions>
				<Popover.Root bind:open={themeOpen}>
					<Popover.Trigger bind:ref={themeTriggerRef}>
						{#snippet child({ props })}
							<Button
								{...props}
								class="w-[160px] justify-between border-2 border-transparent bg-[var(--input)] font-normal hover:border-border hover:bg-white focus-visible:bg-white"
								role="combobox"
								aria-expanded={themeOpen}
							>
								{selectedThemeLabel}
								<ChevronsUpDownIcon class="opacity-50" />
							</Button>
						{/snippet}
					</Popover.Trigger>
					<Popover.Content class="w-[160px] p-0" align="end">
						<Command.Root>
							<Command.List>
								<Command.Group>
									{#each themes as theme (theme.value)}
										<Command.Item
											value={theme.value}
											onSelect={() => {
												currentTheme = theme.value;
												applyTheme(theme.value);
												closeTheme();
											}}
										>
											<CheckIcon
												class={cn(currentTheme !== theme.value && "text-transparent")}
											/>
											{theme.label}
										</Command.Item>
									{/each}
								</Command.Group>
							</Command.List>
						</Command.Root>
					</Popover.Content>
				</Popover.Root>
			</Item.Actions>
		</Item.Root>
	</Item.Group>
</div>
