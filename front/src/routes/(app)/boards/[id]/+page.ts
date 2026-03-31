// Copyright (c) 2025 Start Codex SAS. All rights reserved.
// SPDX-License-Identifier: BUSL-1.1

import { boards, statuses, issues, projects, issueTypes, workspaces } from '$lib/api';
import type { PageLoad } from './$types';

export const load: PageLoad = async ({ params }) => {
	const board = await boards.get(params.id);
	const project = await projects.get(board.project_id);

	const [statusList, issueList, typeList, memberList] = await Promise.all([
		statuses.list(board.project_id).then((r) => r ?? []),
		issues.list(board.project_id).then((r) => r ?? []),
		issueTypes.list(board.project_id).then((r) => r ?? []),
		workspaces.members.list(project.workspace_id).then((r) => r ?? [])
	]);

	return {
		board,
		project,
		statuses: statusList,
		issues: issueList,
		issueTypes: typeList,
		members: memberList,
		breadcrumb: [{ label: board.name }]
	};
};
