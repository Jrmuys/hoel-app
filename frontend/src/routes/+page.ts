import {
    loadDailyOperations,
    loadLogisticsPlanning,
    loadStatusBar,
} from '$lib/api/dashboard';
import type {
    DailyOperationsModel,
    LogisticsPlanningModel,
    StatusBarModel,
} from '$lib/types/dashboard';
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

const fallbackLogisticsPlanning: LogisticsPlanningModel = {
    shoppingTasks: [],
    maintenanceTasks: [],
};

export const load: PageLoad = ({ fetch }) => {
    const statusPromise = loadStatusBar(fetch).catch(() => fallbackStatus);
    const dailyPromise = loadDailyOperations(fetch).catch(() => fallbackDailyOperations);
    const logisticsPromise = loadLogisticsPlanning(fetch).catch(() =>
        fallbackLogisticsPlanning,
    );

    const todayLabel = new Intl.DateTimeFormat(undefined, {
        weekday: 'long',
        month: 'long',
        day: 'numeric',
    }).format(new Date());

    return {
        todayLabel,
        statusPromise,
        dailyPromise,
        logisticsPromise,
    };
};
