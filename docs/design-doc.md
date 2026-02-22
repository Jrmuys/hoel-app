# System Design Document: Household Operations Dashboard

## 1. System Overview

A cross-platform Progressive Web Application (PWA) designed to aggregate household logistics into a single, user-centric interface. The system acts as a middleware aggregator, offloading state management and scheduling logic to established third-party services (TickTick, Home Assistant) while caching data locally to ensure high availability and low latency on the frontend.

## 2. Architecture Stack

| Component      | Technology               | Rationale                                                    |
| -------------- | ------------------------ | ------------------------------------------------------------ |
| **Frontend**   | SvelteKit + Tailwind CSS | Compiles to highly optimized vanilla JS. Supports PWA installation for mobile/tablet. Capable of granular asynchronous rendering (`{#await}`). |
| **Backend**    | Go (Golang)              | High concurrency for concurrent API polling. Low memory footprint. Strict typing prevents runtime errors when parsing volatile third-party JSON payloads. |
| **Database**   | SQLite                   | Serverless, zero-configuration local storage. Sufficient for caching API payloads and logging system errors. |
| **Deployment** | Docker                   | Containerized deployment on a local server or NAS to minimize latency with local network services (Home Assistant). |

## 3. UI/UX Design System

The interface abstracts API sources into functional, time-based modules.

### 3.1. Color Palette

Theme relies on a forest green primary with distinct light and dark modes to prevent eye strain.

| Role            | Light Mode Hex | Dark Mode Hex | Usage                                                    |
| --------------- | -------------- | ------------- | -------------------------------------------------------- |
| **Primary**     | `#2D5A27`      | `#A3B18A`     | Primary actions, active states, key data points.         |
| **Secondary**   | `#8B9474`      | `#5C6B50`     | Inactive states, secondary buttons, subtle borders.      |
| **Background**  | `#F4F5F0`      | `#1E201E`     | Main application background.                             |
| **Surface**     | `#FFFFFF`      | `#2E302E`     | Component/Card backgrounds for depth.                    |
| **Text**        | `#1A1C19`      | `#E8EAE6`     | Primary reading text.                                    |
| **Alert/Error** | `#991B1B`      | `#F87171`     | System failures, garbage cancellations, critical alerts. |

### 3.2. Layout Structure

Grid-based layout relying on 8px spacing increments. UI is divided into functional contexts rather than data sources. Granular Svelte `{#await}` blocks ensure fast-loading modules (Home Assistant) do not wait for slow modules (TickTick).

## 4. Functional Modules

### 4.1. Status Bar (Header)

- **Purpose:** Immediate situational awareness.
- **Data Points:**
  - System health (API connectivity errors).
  - Critical Home Assistant alerts (e.g., unlocked exterior doors, water leak sensors).
  - Municipal alerts (e.g., holiday garbage cancellation).

### 4.2. Daily Operations

- **Purpose:** Actionable items for the current 24-hour cycle.
- **Data Points:**
  - **TickTick:** Tasks tagged `#daily` or due today.
  - **Calendar:** Today's events (read-only).
  - **PGH.st:** Garbage/Recycling indicator (only visible 24 hours prior to pickup).

### 4.3. Environment (Smart Home)

- **Purpose:** Quick access to high-frequency controls.
- **Data Points (Home Assistant):**
  - Quick toggles (primary lights, locks).
  - Environmental sensors (thermostat, indoor/outdoor temperature).

### 4.4. Logistics & Planning

- **Purpose:** Forward-looking management and procurement.
- **Data Points:**
  - **TickTick (Shopping):** Items on the shared household shopping list.
  - **TickTick (Maintenance):** Upcoming long-term maintenance tasks (furnace filters, vehicle registration).

## 5. External Integrations & Data Flow

### 5.1. TickTick (Task Engine)

- **Role:** Single source of truth for all tasks, shopping, and maintenance schedules.
- **Mechanism:** Go backend manages OAuth2 tokens. Backend polls TickTick API every 5-10 minutes, caching results in SQLite.
- **Write Operations:** Frontend task completions trigger an immediate asynchronous `POST` to the Go backend, which proxies the request to TickTick and invalidates the local cache.

### 5.2. PGH.st (Municipal Services)

- **Role:** Determine trash vs. recycling weeks and holiday shifts.
- **Mechanism:** Go cron job polls `https://pgh.st/locate/...` once every 12 hours.
- **Data Handling:** Backend compares `next_pickup_date` and `next_recycling_date` to determine the boolean state of recycling for the current week. Data is cached in SQLite.

### 5.3. Home Assistant

- **Role:** Smart home telemetry and control.
- **Mechanism:** Direct WebSocket connection or REST API integration via the Go backend to ensure local network security (avoiding exposing Home Assistant directly to the PWA if external).

## 6. Error Handling & Alerting

Silent failures of background integrations must be tracked and escalated if persistent.

### 6.1. SQLite Logging

All external HTTP requests from the Go backend are wrapped in an execution handler. Timeouts, 4xx, and 5xx errors are logged to the `api_errors` table.

```
CREATE TABLE api_errors (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    service_name TEXT NOT NULL,
    endpoint TEXT NOT NULL,
    http_status INTEGER,
    error_message TEXT,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    resolved BOOLEAN DEFAULT 0
);
```

### 6.2. Alert Escalation

- **Mechanism:** Go backend queries the `api_errors` table. If a specific service (e.g., TickTick) fails 3 consecutive times within a 15-minute window, an external alert is triggered.
- **Delivery:** HTTP POST to Pushover or Telegram Bot API.
- **Throttling:** Once an alert is sent, the service is placed on a 60-minute cooldown to prevent notification spam.

## 7. Known System Risks & Mitigation

1. **Undocumented APIs (PGH.st):** High risk of silent structural changes. *Mitigation:* Strict JSON unmarshaling in Go. If fields are missing, the system defaults to the last known cache and triggers an alert.
2. **TickTick Rate Limits:** *Mitigation:* Enforce strict backend polling intervals. The frontend must never query TickTick directly.
3. **UI Desynchronization:** If a user checks off a task, but the TickTick API fails, the UI will revert to the unchecked state on the next cache poll. *Mitigation:* Implement optimistic UI updates in Svelte, coupled with an explicit error toast if the backend mutation fails.

## 8. Implementation Phases

- **Phase 1 (MVP):** Go backend setup, SQLite cache, PGH.st polling, SvelteKit UI scaffolding (Status and Daily Operations), Telegram alerting.
- **Phase 2:** TickTick OAuth integration, bidirectional task sync, Logistics & Planning UI module.
- **Phase 3:** Home Assistant local API integration, Environment UI module, Calendar read-only integration.