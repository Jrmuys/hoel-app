import type { DailyOperationsModel, StatusBarModel } from '$lib/types/dashboard';

function delay(durationMs: number): Promise<void> {
    return new Promise((resolve) => {
        setTimeout(resolve, durationMs);
    });
}

export async function loadStatusBar(): Promise<StatusBarModel> {
    await delay(120);

    return {
        systemHealth: 'degraded',
        alerts: [
            {
                id: 'ticktick-sync',
                message: 'TickTick sync degraded. Last refresh is older than 15 minutes.',
                level: 'warning',
                source: 'ticktick'
            }
        ],
        integrations: [
            {
                service: 'ticktick',
                state: 'degraded',
                lastSuccessAt: new Date(Date.now() - 18 * 60_000).toISOString(),
                consecutiveFailures: 2
            },
            {
                service: 'pghst',
                state: 'healthy',
                lastSuccessAt: new Date(Date.now() - 2 * 60_000).toISOString(),
                consecutiveFailures: 0
            }
        ]
    };
}

export async function loadDailyOperations(): Promise<DailyOperationsModel> {
    await delay(180);

    return {
        tasks: [
            {
                id: 'task-1',
                title: 'Take out trash bins',
                dueAt: new Date().toISOString(),
                completed: false,
                source: 'ticktick'
            },
            {
                id: 'task-2',
                title: 'Start dishwasher before bed',
                dueAt: new Date().toISOString(),
                completed: false,
                source: 'ticktick'
            }
        ],
        garbage: {
            nextPickupDate: new Date(Date.now() + 24 * 60 * 60 * 1000).toISOString(),
            nextTrashPickupDate: new Date(Date.now() + 24 * 60 * 60 * 1000).toISOString(),
            nextRecyclingPickupDate: new Date(Date.now() + 3 * 24 * 60 * 60 * 1000).toISOString(),
            isRecyclingWeek: true,
            showIndicator: true,
            showTrashTakeOutReminder: true,
            showRecyclingTakeOutReminder: false,
        }
    };
}
