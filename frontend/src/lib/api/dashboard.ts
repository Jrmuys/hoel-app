import type {
	AlertLevel,
	DailyOperationsModel,
	IntegrationHealth,
	StatusAlert,
	StatusBarModel,
} from '$lib/types/dashboard';

const defaultBaseURL = 'http://127.0.0.1:8080';
type FetchFn = typeof fetch;

function apiBaseURL(): string {
	const configured = import.meta.env.PUBLIC_API_BASE_URL;
	if (typeof configured !== 'string' || configured.trim() === '') {
		return defaultBaseURL;
	}

	return configured.trim().replace(/\/$/, '');
}

async function fetchJSON<T>(path: string, fetchFn: FetchFn): Promise<T> {
	const response = await fetchFn(`${apiBaseURL()}${path}`, {
		headers: {
			Accept: 'application/json',
		},
	});

	if (!response.ok) {
		const responseText = await response.text();
		const message = responseText.trim() || response.statusText;
		throw new Error(`${path} failed (${response.status}): ${message}`);
	}

	return (await response.json()) as T;
}

interface RawStatusBarResponse {
	system_health?: string;
	alerts?: RawStatusAlert[];
	integrations?: RawIntegrationStatus[];
}

interface RawStatusAlert {
	id?: number | string;
	source?: string;
	severity?: string;
	message?: string;
}

interface RawIntegrationStatus {
	service?: string;
	state?: string;
	last_success_at?: string | null;
	consecutive_failures?: number;
}

interface RawDailyOperationsResponse {
	tasks?: RawDailyTask[];
	garbage?: RawGarbage;
}

interface RawDailyTask {
	id?: string;
	title?: string;
	dueAt?: string;
	completed?: boolean;
	source?: string;
}

interface RawGarbage {
	nextPickupDate?: string;
	isRecyclingWeek?: boolean;
	showIndicator?: boolean;
}

export async function loadStatusBar(fetchFn: FetchFn = fetch): Promise<StatusBarModel> {
	const payload = await fetchJSON<RawStatusBarResponse>('/api/status-bar', fetchFn);

	return {
		systemHealth: toSystemHealth(payload.system_health),
		alerts: (payload.alerts ?? []).map(toStatusAlert),
		integrations: (payload.integrations ?? []).map(toIntegrationHealth),
	};
}

export async function loadDailyOperations(fetchFn: FetchFn = fetch): Promise<DailyOperationsModel> {
	const payload = await fetchJSON<RawDailyOperationsResponse>('/api/daily-operations', fetchFn);

	return {
		tasks: (payload.tasks ?? []).map((task) => ({
			id: task.id ?? '',
			title: task.title ?? 'Untitled task',
			dueAt: task.dueAt ?? new Date().toISOString(),
			completed: Boolean(task.completed),
			source: 'ticktick',
		})),
		garbage: {
			nextPickupDate: payload.garbage?.nextPickupDate ?? '',
			isRecyclingWeek: Boolean(payload.garbage?.isRecyclingWeek),
			showIndicator: Boolean(payload.garbage?.showIndicator),
		},
	};
}

function toStatusAlert(alert: RawStatusAlert): StatusAlert {
	return {
		id: String(alert.id ?? ''),
		source: toAlertSource(alert.source),
		level: toAlertLevel(alert.severity),
		message: alert.message?.trim() || 'Integration alert',
	};
}

function toIntegrationHealth(integration: RawIntegrationStatus): IntegrationHealth {
	return {
		service: toIntegrationService(integration.service),
		state: toIntegrationState(integration.state),
		lastSuccessAt: integration.last_success_at ?? null,
		consecutiveFailures: Number(integration.consecutive_failures ?? 0),
	};
}

function toSystemHealth(value: string | undefined): StatusBarModel['systemHealth'] {
	if (value === 'degraded' || value === 'down') {
		return value;
	}

	return 'ok';
}

function toAlertSource(value: string | undefined): StatusAlert['source'] {
	if (value === 'ticktick' || value === 'pghst') {
		return value;
	}

	return 'system';
}

function toAlertLevel(value: string | undefined): AlertLevel {
	if (value === 'critical' || value === 'info') {
		return value;
	}

	return 'warning';
}

function toIntegrationService(value: string | undefined): IntegrationHealth['service'] {
	if (value === 'ticktick') {
		return 'ticktick';
	}

	return 'pghst';
}

function toIntegrationState(value: string | undefined): IntegrationHealth['state'] {
	if (value === 'degraded' || value === 'down') {
		return value;
	}

	return 'healthy';
}
