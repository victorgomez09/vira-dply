
import { redirect } from '@sveltejs/kit';
import type { PageLoad } from './$types';

export const load: PageLoad = async ({ params }) => {
  redirect(302, `/dashboard/project/${params.id}/${params.env_id}/app/${params.res_id}/overview`);
};
