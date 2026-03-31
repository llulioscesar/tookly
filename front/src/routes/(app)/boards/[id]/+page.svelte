<!-- Copyright (c) 2025 Start Codex SAS. All rights reserved. -->
<!-- SPDX-License-Identifier: BUSL-1.1 -->

<script lang="ts">
	import { goto } from '$app/navigation';
	import type { PageData } from './$types';
	import type { Issue, Status } from '$lib/api';
	import { issues as issuesApi, statuses as statusesApi } from '$lib/api';
	import { dndzone } from 'svelte-dnd-action';
	import * as Empty from '$lib/components/ui/empty/index.js';
	import * as Sheet from '$lib/components/ui/sheet/index.js';
	import { Button } from '$lib/components/ui/button/index.js';
	import { Input } from '$lib/components/ui/input/index.js';
	import LayoutIcon from '@lucide/svelte/icons/layout';
	import FilterXIcon from '@lucide/svelte/icons/filter-x';
	import * as m from '$lib/paraglide/messages';
	import { i18n } from '$lib/i18n.svelte';

	let { data }: { data: PageData } = $props();

	const t = $derived.by(() => {
		i18n.locale;
		return {
			noIssues:       m.board_no_issues(),
			noStatuses:     m.board_no_statuses(),
			noStatusesDesc: m.board_no_statuses_desc(),
			createStatus:   m.status_create(),
			name:           m.status_name(),
			category:       m.status_category(),
			creating:       m.status_creating(),
			cancel:         m.workspace_cancel(),
			catTodo:        m.status_category_todo(),
			catInProgress:  m.status_category_in_progress(),
			catDone:        m.status_category_done(),
			filterAssignee: m.board_filter_assignee(),
			filterPriority: m.board_filter_priority(),
			filterType:     m.board_filter_type(),
			filterAll:      m.board_filter_all(),
			filterClear:    m.board_filter_clear()
		};
	});

	const priorityColors: Record<string, string> = {
		urgent: 'bg-red-100 text-red-700',
		high: 'bg-orange-100 text-orange-700',
		medium: 'bg-yellow-100 text-yellow-700',
		low: 'bg-blue-100 text-blue-700'
	};

	// --- Local reactive state ---
	let localStatuses = $state<Status[]>([]);
	$effect(() => { localStatuses = [...data.statuses]; });

	const sortedStatuses = $derived([...localStatuses].sort((a, b) => a.position - b.position));

	// All columns (unfiltered) — source of truth for positions and DnD
	// Initialized from data.issues only when the data reference changes (navigation),
	// not when localStatuses changes (e.g. creating a new status).
	let allColumns = $state<Record<string, Issue[]>>({});
	let persistedColumns = $state<Record<string, Issue[]>>({});

	function buildColumns(issueList: Issue[], statusList: Status[]): Record<string, Issue[]> {
		const cols: Record<string, Issue[]> = {};
		for (const status of statusList) {
			cols[status.id] = issueList
				.filter((i) => i.status_id === status.id)
				.sort((a, b) => a.status_position - b.status_position);
		}
		return cols;
	}

	$effect(() => {
		const cols = buildColumns(data.issues, data.statuses);
		allColumns = cols;
		persistedColumns = JSON.parse(JSON.stringify(cols));
	});

	// --- Filters (reset on board change) ---
	let filterAssignee = $state('');
	let filterPriority = $state('');
	let filterType = $state('');

	$effect(() => {
		data.board.id;
		filterAssignee = '';
		filterPriority = '';
		filterType = '';
	});

	const hasFilters = $derived(filterAssignee !== '' || filterPriority !== '' || filterType !== '');

	function matchesFilters(issue: Issue): boolean {
		if (filterAssignee && issue.assignee_id !== filterAssignee) return false;
		if (filterPriority && issue.priority !== filterPriority) return false;
		if (filterType && issue.issue_type_id !== filterType) return false;
		return true;
	}

	function clearFilters() {
		filterAssignee = '';
		filterPriority = '';
		filterType = '';
	}

	// Filtered columns — what DnD zones render
	let columns = $state<Record<string, Issue[]>>({});
	$effect(() => {
		const cols: Record<string, Issue[]> = {};
		for (const statusId of Object.keys(allColumns)) {
			cols[statusId] = allColumns[statusId].filter(matchesFilters);
		}
		columns = cols;
	});

	// --- Drag and drop ---
	const flipDurationMs = 200;

	function handleConsider(statusId: string, e: CustomEvent<{ items: Issue[] }>) {
		columns[statusId] = e.detail.items;
	}

	async function handleFinalize(statusId: string, e: CustomEvent<{ items: Issue[]; info: { id: string } }>) {
		columns[statusId] = e.detail.items;
		const draggedId = e.detail.info.id;

		const visibleIdx = columns[statusId].findIndex((i) => i.id === draggedId);
		if (visibleIdx === -1) return;

		// Remove dragged item from all columns first, then calculate position
		const draggedIssue = columns[statusId][visibleIdx];
		for (const sid of Object.keys(allColumns)) {
			allColumns[sid] = allColumns[sid].filter((i) => i.id !== draggedId);
		}

		// Calculate target_position against the full column (with dragged item already removed)
		const fullCol = allColumns[statusId];
		let targetPosition: number;

		if (!hasFilters) {
			targetPosition = visibleIdx;
		} else {
			const prevVisible = visibleIdx > 0 ? columns[statusId][visibleIdx - 1] : null;
			const nextVisible = visibleIdx < columns[statusId].length - 1 ? columns[statusId][visibleIdx + 1] : null;

			if (prevVisible) {
				targetPosition = fullCol.findIndex((i) => i.id === prevVisible.id) + 1;
			} else if (nextVisible) {
				targetPosition = fullCol.findIndex((i) => i.id === nextVisible.id);
			} else {
				targetPosition = fullCol.length;
			}
		}

		// Insert at calculated position
		allColumns[statusId].splice(targetPosition, 0, { ...draggedIssue, status_id: statusId });

		try {
			await issuesApi.move(data.board.project_id, draggedId, {
				target_status_id: statusId,
				target_position: targetPosition
			});
			persistedColumns = JSON.parse(JSON.stringify(allColumns));
		} catch {
			allColumns = JSON.parse(JSON.stringify(persistedColumns));
			try {
				const fresh = await issuesApi.list(data.board.project_id);
				if (fresh) {
					const cols: Record<string, Issue[]> = {};
					for (const status of localStatuses) {
						cols[status.id] = fresh
							.filter((i) => i.status_id === status.id)
							.sort((a, b) => a.status_position - b.status_position);
					}
					allColumns = cols;
					persistedColumns = JSON.parse(JSON.stringify(cols));
				}
			} catch {
				// keep persisted state
			}
		}
	}

	function issueHref(issue: Issue): string {
		return `/projects/${data.board.project_id}/issues/${issue.id}`;
	}

	// --- Create status sheet ---
	let sheetOpen    = $state(false);
	let statusName   = $state('');
	let category     = $state<'todo' | 'doing' | 'done'>('todo');
	let saving       = $state(false);
	let error        = $state('');

	function resetForm() { statusName = ''; category = 'todo'; error = ''; saving = false; }

	async function handleCreate(e: SubmitEvent) {
		e.preventDefault();
		error = '';
		saving = true;
		try {
			const created = await statusesApi.create(data.board.project_id, {
				name: statusName.trim(),
				category
			});
			localStatuses = [...localStatuses, created];
			allColumns[created.id] = [];
			persistedColumns[created.id] = [];
			sheetOpen = false;
			resetForm();
		} catch (err) {
			error = err instanceof Error ? err.message : 'Failed to create status';
		} finally {
			saving = false;
		}
	}

	const selectClass = 'flex h-8 rounded-md border border-input bg-background px-2 py-1 text-xs shadow-xs focus-visible:outline-none focus-visible:ring-1 focus-visible:ring-ring';
</script>

{#if sortedStatuses.length === 0}
	<Empty.Root class="border border-dashed">
		<Empty.Header>
			<Empty.Media variant="icon"><LayoutIcon /></Empty.Media>
			<Empty.Title>{t.noStatuses}</Empty.Title>
			<Empty.Description>{t.noStatusesDesc}</Empty.Description>
		</Empty.Header>
		<Empty.Content>
			<Button onclick={() => { sheetOpen = true; }}>{t.createStatus}</Button>
		</Empty.Content>
	</Empty.Root>
{:else}
	<!-- Filter bar -->
	<div class="mb-4 flex flex-wrap items-center gap-3">
		<select bind:value={filterAssignee} class={selectClass}>
			<option value="">{t.filterAssignee}: {t.filterAll}</option>
			{#each data.members as member}
				<option value={member.user_id}>{member.user_id.slice(0, 8)}</option>
			{/each}
		</select>

		<select bind:value={filterPriority} class={selectClass}>
			<option value="">{t.filterPriority}: {t.filterAll}</option>
			<option value="critical">Critical</option>
			<option value="high">High</option>
			<option value="medium">Medium</option>
			<option value="low">Low</option>
		</select>

		<select bind:value={filterType} class={selectClass}>
			<option value="">{t.filterType}: {t.filterAll}</option>
			{#each data.issueTypes as type}
				<option value={type.id}>{type.name}</option>
			{/each}
		</select>

		{#if hasFilters}
			<Button variant="ghost" size="sm" onclick={clearFilters}>
				<FilterXIcon class="mr-1 size-3.5" />
				{t.filterClear}
			</Button>
		{/if}
	</div>

	<!-- Board columns -->
	<div class="flex h-full min-h-0 gap-4 overflow-x-auto pb-4">
		{#each sortedStatuses as status (status.id)}
			<div class="flex w-72 shrink-0 flex-col gap-3">
				<div class="flex items-center gap-2 px-1">
					<span class="text-sm font-medium">{status.name}</span>
					<span class="rounded-full bg-muted px-2 py-0.5 text-xs text-muted-foreground">
						{(columns[status.id] ?? []).length}
					</span>
				</div>
				<div
					class="flex flex-1 flex-col gap-2 overflow-y-auto rounded-lg bg-muted/40 p-2"
					use:dndzone={{ items: columns[status.id] ?? [], flipDurationMs, dropTargetStyle: {} }}
					onconsider={(e) => handleConsider(status.id, e)}
					onfinalize={(e) => handleFinalize(status.id, e)}
				>
					{#each columns[status.id] ?? [] as issue (issue.id)}
						<div
							class="flex flex-col gap-2 rounded-md border bg-background p-3 shadow-xs transition-colors hover:bg-muted/50 cursor-grab active:cursor-grabbing"
							onclick={() => goto(issueHref(issue))}
							onkeydown={(e) => { if (e.key === 'Enter') goto(issueHref(issue)); }}
							role="button"
							tabindex="0"
						>
							<div class="flex items-start justify-between gap-2">
								<span class="text-sm leading-snug">{issue.title}</span>
								{#if issue.priority && issue.priority !== 'none'}
									<span
										class="shrink-0 rounded px-1.5 py-0.5 text-xs font-medium {priorityColors[issue.priority] ?? 'bg-muted text-muted-foreground'}"
									>
										{issue.priority}
									</span>
								{/if}
							</div>
							<span class="text-xs text-muted-foreground">#{issue.number}</span>
						</div>
					{/each}
				</div>
			</div>
		{/each}
	</div>
{/if}

<Sheet.Root bind:open={sheetOpen} onOpenChange={(open) => { if (!open) resetForm(); }}>
	<Sheet.Portal>
		<Sheet.Overlay />
		<Sheet.Content side="right" class="w-96">
			<Sheet.Header>
				<Sheet.Title>{t.createStatus}</Sheet.Title>
			</Sheet.Header>
			<form onsubmit={handleCreate} class="flex flex-col gap-4 p-6">
				<div class="flex flex-col gap-1.5">
					<label for="status-name" class="text-sm font-medium">{t.name}</label>
					<Input id="status-name" placeholder="To Do" bind:value={statusName} required />
				</div>
				<div class="flex flex-col gap-1.5">
					<label for="status-cat" class="text-sm font-medium">{t.category}</label>
					<select
						id="status-cat"
						bind:value={category}
						class="flex h-9 w-full rounded-md border border-input bg-background px-3 py-1 text-sm shadow-xs focus-visible:outline-none focus-visible:ring-1 focus-visible:ring-ring"
					>
						<option value="todo">{t.catTodo}</option>
						<option value="doing">{t.catInProgress}</option>
						<option value="done">{t.catDone}</option>
					</select>
				</div>
				{#if error}<p class="text-sm text-destructive">{error}</p>{/if}
				<div class="flex justify-end gap-2 pt-2">
					<Sheet.Close><Button variant="outline" type="button">{t.cancel}</Button></Sheet.Close>
					<Button type="submit" disabled={saving || !statusName.trim()}>
						{saving ? t.creating : t.createStatus}
					</Button>
				</div>
			</form>
		</Sheet.Content>
	</Sheet.Portal>
</Sheet.Root>
