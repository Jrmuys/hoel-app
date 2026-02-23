export type AlertLevel = 'critical' | 'warning' | 'info';

export interface StatusAlert {
    id: string;
    message: string;
    level: AlertLevel;
    source: 'ticktick' | 'pghst' | 'system';
}

export interface IntegrationHealth {
    service: 'ticktick' | 'pghst';
    state: 'healthy' | 'degraded' | 'down';
    lastSuccessAt: string | null;
    consecutiveFailures: number;
}

export interface StatusBarModel {
    systemHealth: 'ok' | 'degraded' | 'down';
    alerts: StatusAlert[];
    integrations: IntegrationHealth[];
}

export interface DailyTask {
    id: string;
    title: string;
    dueAt: string;
    completed: boolean;
    source: 'ticktick' | 'system';
}

export interface GarbageWindow {
    nextPickupDate: string;
    nextTrashPickupDate: string;
    nextRecyclingPickupDate: string;
    isRecyclingWeek: boolean;
    showIndicator: boolean;
    showTrashTakeOutReminder: boolean;
    showRecyclingTakeOutReminder: boolean;
}

export interface DailyOperationsModel {
    tasks: DailyTask[];
    garbage: GarbageWindow;
}
