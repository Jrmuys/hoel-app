import { loadDailyOperations, loadStatusBar } from '$lib/api/dashboard';
import type { DailyOperationsModel, StatusBarModel } from '$lib/types/dashboard';
import type { PageLoad } from './$types';

const fallbackStatus: StatusBarModel = {
    systemHealth: 'down',
    alerts: [
        {
            id: 'backend-unavailable',
            source: 'system',
            level: 'critical',
            message: 'Backend is unavailable. Showing cached interface state only.',
        },
    ],
    integrations: [],
};

const fallbackDailyOperations: DailyOperationsModel = {
    tasks: [],
    garbage: {
        nextPickupDate: '',
        nextTrashPickupDate: '',
        nextRecyclingPickupDate: '',
        isRecyclingWeek: false,
        showIndicator: false,
        showTrashTakeOutReminder: false,
        showRecyclingTakeOutReminder: false,
    },
};

export const load: PageLoad = ({ fetch }) => {
    const statusPromise = loadStatusBar(fetch).catch(() => fallbackStatus);
    const dailyPromise = loadDailyOperations(fetch).catch(() => fallbackDailyOperations);

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
