// Copyright (c) 2025 Start Codex SAS. All rights reserved.
// SPDX-License-Identifier: BUSL-1.1

import { redirect } from '@sveltejs/kit';
import { instance, auth } from '$lib/api';
import type { PageLoad } from './$types';

export const load: PageLoad = async ({ url }) => {
	const { initialized } = await instance.status();
	if (!initialized) redirect(302, '/setup');

	try {
		const me = await auth.me();
		if (me.authenticated) redirect(302, '/');
	} catch {
		// not logged in — continue
	}

	const token = url.searchParams.get('token') ?? '';
	return { token };
};
