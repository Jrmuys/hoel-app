import { loadDailyOperations, loadStatusBar } from '$lib/api/dashboard';
import type { PageLoad } from './$types';

export const load: PageLoad = ({ fetch }) => {
    const statusPromise = loadStatusBar(fetch);
    const dailyPromise = loadDailyOperations(fetch);

    const todayLabel = new Intl.DateTimeFormat(undefined, {
        weekday: 'long',
        month: 'long',
        day: 'numeric',
    }).format(new Date());

    return {
        todayLabel,
        statusPromise,
        dailyPromise,
    };
};
