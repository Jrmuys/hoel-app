<script lang="ts">
    import {
        Calendar,
        ChevronLeft,
        ChevronRight,
        Recycle,
        Trash2,
    } from 'lucide-svelte';
    import {
        completeTickTickTask,
        createTickTickTask,
        forceRefreshTickTick,
        loadDailyOperations,
        updateTickTickTask,
    } from '$lib/api/dashboard';
    import type { DailyOperationsModel, DailyTask } from '$lib/types/dashboard';

    export let data: DailyOperationsModel;

    const EMPTY_DAILY_OPERATIONS: DailyOperationsModel = {
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

    let viewData: DailyOperationsModel = EMPTY_DAILY_OPERATIONS;
    let completingTaskId = '';
    let completionError = '';
    let isRefreshing = false;
    let mutatingTaskId = '';
    let mutationError = '';
    let newTaskTitle = '';
    let newTaskDate = defaultDateInputValue();
    let newTaskTime = defaultTimeInputValue();
    let editingTaskId = '';
    let editTaskTitle = '';
    let editTaskDate = '';
    let editTaskTime = '';
    let pickerTarget: 'new' | 'edit' | '' = '';
    let pickerMonth = new Date();
    const weekdayLabels = ['Su', 'Mo', 'Tu', 'We', 'Th', 'Fr', 'Sa'];
    const monthLabels = [
        'Jan',
        'Feb',
        'Mar',
        'Apr',
        'May',
        'Jun',
        'Jul',
        'Aug',
        'Sep',
        'Oct',
        'Nov',
        'Dec',
    ];
    const quickDateOptions: Array<{ label: string; offset: number }> = [
        { label: 'Today', offset: 0 },
        { label: 'Tomorrow', offset: 1 },
        { label: 'Next week', offset: 7 },
    ];
    let newPickerContainer: HTMLDivElement | undefined;
    let editPickerContainer: HTMLDivElement | undefined;

    function cloneDailyOperations(
        source: DailyOperationsModel,
    ): DailyOperationsModel {
        return {
            tasks: [...source.tasks],
            garbage: { ...source.garbage },
        };
    }

    async function handleForceRefresh() {
        if (isRefreshing || mutatingTaskId !== '' || completingTaskId !== '') {
            return;
        }

        isRefreshing = true;
        mutationError = '';

        try {
            await forceRefreshTickTick();
            const latest = await loadDailyOperations();
            viewData = cloneDailyOperations(latest);
        } catch (error) {
            mutationError =
                error instanceof Error
                    ? error.message
                    : 'Unable to force refresh tasks right now.';
        } finally {
            isRefreshing = false;
        }
    }

    $: if (data) {
        viewData = cloneDailyOperations(data);
    }

    function startOfDay(value: Date): Date {
        return new Date(value.getFullYear(), value.getMonth(), value.getDate());
    }

    function dayDifferenceFromToday(value: Date): number {
        const todayStart = startOfDay(new Date());
        const valueStart = startOfDay(value);
        const millisecondsPerDay = 24 * 60 * 60 * 1000;
        return Math.round(
            (valueStart.getTime() - todayStart.getTime()) / millisecondsPerDay,
        );
    }

    function startOfWeek(value: Date): Date {
        const start = startOfDay(value);
        start.setDate(start.getDate() - start.getDay());
        return start;
    }

    function hasMeaningfulTime(value: Date): boolean {
        return (
            value.getHours() !== 0 ||
            value.getMinutes() !== 0 ||
            value.getSeconds() !== 0 ||
            value.getMilliseconds() !== 0
        );
    }

    function isTaskOverdue(task: DailyTask): boolean {
        const parsed = new Date(task.dueAt);
        if (Number.isNaN(parsed.getTime())) {
            return false;
        }

        const now = new Date();
        if (task.hasTime) {
            return parsed.getTime() < now.getTime();
        }

        return dayDifferenceFromToday(parsed) < 0;
    }

    function isPickupSoon(dateIso: string): boolean {
        if (dateIso.trim() === '') {
            return false;
        }

        const parsed = new Date(dateIso);
        if (Number.isNaN(parsed.getTime())) {
            return false;
        }

        const dayDifference = dayDifferenceFromToday(parsed);
        return dayDifference >= 0 && dayDifference <= 1;
    }

    function formatDate(
        dateIso: string,
        options: { includeTimeForToday?: boolean; hasTime?: boolean } = {},
    ): string {
        const parsed = new Date(dateIso);
        if (Number.isNaN(parsed.getTime())) {
            return 'Invalid date';
        }

        const dayDifference = dayDifferenceFromToday(parsed);
        if (dayDifference === 0) {
            const shouldShowTime = options.hasTime ?? hasMeaningfulTime(parsed);
            if (options.includeTimeForToday && shouldShowTime) {
                return `Today at ${parsed.toLocaleTimeString(undefined, {
                    hour: 'numeric',
                    minute: '2-digit',
                })}`;
            }

            return 'Today';
        }
        if (dayDifference === 1) {
            return 'Tomorrow';
        }
        if (dayDifference === -1) {
            return 'Yesterday';
        }

        const parsedStart = startOfDay(parsed);
        const today = new Date();
        const currentWeekStart = startOfWeek(today);
        const nextWeekStart = new Date(currentWeekStart);
        nextWeekStart.setDate(nextWeekStart.getDate() + 7);
        const weekAfterNextStart = new Date(nextWeekStart);
        weekAfterNextStart.setDate(weekAfterNextStart.getDate() + 7);

        const weekdayName = parsed.toLocaleDateString(undefined, {
            weekday: 'long',
        });

        if (parsedStart >= currentWeekStart && parsedStart < nextWeekStart) {
            return weekdayName;
        }

        if (parsedStart >= nextWeekStart && parsedStart < weekAfterNextStart) {
            return `next ${weekdayName}`;
        }

        return parsed.toLocaleDateString(undefined, {
            weekday: 'short',
            month: 'short',
            day: 'numeric',
        });
    }

    function formatOptionalDate(dateIso: string): string {
        if (!dateIso) {
            return 'Not available';
        }

        return formatDate(dateIso);
    }

    function defaultDateInputValue(): string {
        const now = new Date();
        return toDateInputValue(now);
    }

    function defaultTimeInputValue(): string {
        return '';
    }

    function toDateInputValue(value: Date): string {
        const year = value.getFullYear();
        const month = String(value.getMonth() + 1).padStart(2, '0');
        const day = String(value.getDate()).padStart(2, '0');
        return `${year}-${month}-${day}`;
    }

    function toDueAtPayload(dateValue: string, timeValue: string): string {
        const normalizedDate = dateValue.trim();
        if (normalizedDate === '') {
            return '';
        }

        const normalizedTime = timeValue.trim();
        if (normalizedTime === '') {
            return normalizedDate;
        }

        const dateTimeValue = `${normalizedDate}T${normalizedTime}`;

        const localDate = new Date(dateTimeValue);
        if (Number.isNaN(localDate.getTime())) {
            return '';
        }

        return localDate.toISOString();
    }

    function formatDateControl(dateValue: string): string {
        if (dateValue.trim() === '') {
            return 'Pick date';
        }

        const localDate = new Date(`${dateValue}T00:00:00`);
        if (Number.isNaN(localDate.getTime())) {
            return 'Pick date';
        }

        return localDate.toLocaleDateString(undefined, {
            weekday: 'short',
            month: 'short',
            day: 'numeric',
        });
    }

    function openPicker(target: 'new' | 'edit'): void {
        pickerTarget = target;
        const dateValue = target === 'new' ? newTaskDate : editTaskDate;
        const baseDate = new Date(
            `${dateValue || defaultDateInputValue()}T00:00:00`,
        );
        pickerMonth = new Date(baseDate.getFullYear(), baseDate.getMonth(), 1);
    }

    function closePicker(): void {
        pickerTarget = '';
    }

    function appliesToOpenPicker(target: EventTarget | null): boolean {
        if (!(target instanceof Node)) {
            return false;
        }

        if (pickerTarget === 'new') {
            return newPickerContainer?.contains(target) ?? false;
        }

        if (pickerTarget === 'edit') {
            return editPickerContainer?.contains(target) ?? false;
        }

        return false;
    }

    function handleWindowMouseDown(event: MouseEvent): void {
        if (pickerTarget === '') {
            return;
        }

        if (!appliesToOpenPicker(event.target)) {
            closePicker();
        }
    }

    function handleWindowFocusIn(event: FocusEvent): void {
        if (pickerTarget === '') {
            return;
        }

        if (!appliesToOpenPicker(event.target)) {
            closePicker();
        }
    }

    function handleWindowKeyDown(event: KeyboardEvent): void {
        if (event.key === 'Escape') {
            closePicker();
        }
    }

    function previousMonth(): void {
        pickerMonth = new Date(
            pickerMonth.getFullYear(),
            pickerMonth.getMonth() - 1,
            1,
        );
    }

    function nextMonth(): void {
        pickerMonth = new Date(
            pickerMonth.getFullYear(),
            pickerMonth.getMonth() + 1,
            1,
        );
    }

    function pickerYearOptions(): number[] {
        const centerYear = pickerMonth.getFullYear();
        return Array.from(
            { length: 21 },
            (_, index) => centerYear - 10 + index,
        );
    }

    function setPickerMonthFromSelect(value: string): void {
        const parsed = Number.parseInt(value, 10);
        if (Number.isNaN(parsed)) {
            return;
        }

        pickerMonth = new Date(pickerMonth.getFullYear(), parsed, 1);
    }

    function setPickerYearFromSelect(value: string): void {
        const parsed = Number.parseInt(value, 10);
        if (Number.isNaN(parsed)) {
            return;
        }

        pickerMonth = new Date(parsed, pickerMonth.getMonth(), 1);
    }

    function setPickedDay(day: number): void {
        const pickedDate = new Date(
            pickerMonth.getFullYear(),
            pickerMonth.getMonth(),
            day,
        );
        const dateValue = toDateInputValue(pickedDate);

        if (pickerTarget === 'new') {
            newTaskDate = dateValue;
        } else if (pickerTarget === 'edit') {
            editTaskDate = dateValue;
        }

        closePicker();
    }

    function applyQuickDate(offsetDays: number): void {
        const date = new Date();
        date.setHours(0, 0, 0, 0);
        date.setDate(date.getDate() + offsetDays);

        const dateValue = toDateInputValue(date);
        if (pickerTarget === 'new') {
            newTaskDate = dateValue;
        } else if (pickerTarget === 'edit') {
            editTaskDate = dateValue;
        }

        pickerMonth = new Date(date.getFullYear(), date.getMonth(), 1);
        closePicker();
    }

    function calendarDays(): Array<number | null> {
        const year = pickerMonth.getFullYear();
        const month = pickerMonth.getMonth();
        const firstDayOffset = new Date(year, month, 1).getDay();
        const totalDays = new Date(year, month + 1, 0).getDate();

        const days: Array<number | null> = [];
        for (let index = 0; index < firstDayOffset; index += 1) {
            days.push(null);
        }
        for (let day = 1; day <= totalDays; day += 1) {
            days.push(day);
        }

        return days;
    }

    function selectedDayValue(): string {
        if (pickerTarget === 'new') {
            return newTaskDate;
        }
        if (pickerTarget === 'edit') {
            return editTaskDate;
        }

        return '';
    }

    function isSelectedDay(day: number): boolean {
        const selectedValue = selectedDayValue();
        if (selectedValue === '') {
            return false;
        }

        const selectedDate = new Date(`${selectedValue}T00:00:00`);
        return (
            selectedDate.getFullYear() === pickerMonth.getFullYear() &&
            selectedDate.getMonth() === pickerMonth.getMonth() &&
            selectedDate.getDate() === day
        );
    }

    async function handleTaskCompletion(
        taskId: string,
        source: 'ticktick' | 'system',
    ): Promise<void> {
        if (source !== 'ticktick') {
            return;
        }

        if (completingTaskId !== '') {
            return;
        }

        if (mutatingTaskId !== '') {
            return;
        }

        const currentSnapshot = cloneDailyOperations(viewData);
        completionError = '';
        completingTaskId = taskId;

        viewData = {
            ...viewData,
            tasks: viewData.tasks.map((task: DailyTask) =>
                task.id === taskId ? { ...task, completed: true } : task,
            ),
        };

        try {
            await completeTickTickTask(taskId);
            viewData = await loadDailyOperations();
        } catch {
            viewData = currentSnapshot;
            completionError = 'Unable to complete task right now.';
        } finally {
            completingTaskId = '';
        }
    }

    async function handleCreateTask(): Promise<void> {
        if (mutatingTaskId !== '') {
            return;
        }

        mutationError = '';
        const dueAtISO = toDueAtPayload(newTaskDate, newTaskTime);
        if (newTaskTitle.trim() === '' || dueAtISO === '') {
            mutationError = 'Provide title and due date.';
            return;
        }

        mutatingTaskId = 'create';
        try {
            await createTickTickTask(newTaskTitle, dueAtISO);
            viewData = await loadDailyOperations();
            newTaskTitle = '';
            newTaskDate = defaultDateInputValue();
            newTaskTime = defaultTimeInputValue();
        } catch {
            mutationError = 'Unable to create task right now.';
        } finally {
            mutatingTaskId = '';
        }
    }

    function startEditingTask(task: DailyTask): void {
        editingTaskId = task.id;
        editTaskTitle = task.title;
        const dueAt = new Date(task.dueAt);
        editTaskDate = toDateInputValue(dueAt);
        editTaskTime = task.hasTime
            ? `${String(dueAt.getHours()).padStart(2, '0')}:${String(dueAt.getMinutes()).padStart(2, '0')}`
            : '';
        mutationError = '';
    }

    function cancelEditingTask(): void {
        editingTaskId = '';
        editTaskTitle = '';
        editTaskDate = '';
        editTaskTime = '';
        closePicker();
    }

    async function saveTaskEdit(taskID: string): Promise<void> {
        if (mutatingTaskId !== '') {
            return;
        }

        mutationError = '';
        const dueAtISO = toDueAtPayload(editTaskDate, editTaskTime);
        if (editTaskTitle.trim() === '' || dueAtISO === '') {
            mutationError = 'Provide title and due date before saving.';
            return;
        }

        mutatingTaskId = taskID;
        try {
            await updateTickTickTask(taskID, editTaskTitle, dueAtISO);
            viewData = await loadDailyOperations();
            cancelEditingTask();
        } catch {
            mutationError = 'Unable to update task right now.';
        } finally {
            mutatingTaskId = '';
        }
    }
</script>

<svelte:window
    onmousedown={handleWindowMouseDown}
    onfocusin={handleWindowFocusIn}
    onkeydown={handleWindowKeyDown}
/>

<section class="panel">
    <header class="flex flex-wrap items-start justify-between gap-3">
        <div>
            <h2 class="text-lg font-semibold">Daily Operations</h2>
            <p class="mt-1 text-sm text-[var(--color-text)]/70">
                Priority tasks for the next 24 hours
            </p>
        </div>
        <div class="flex flex-col items-start gap-2 sm:items-end">
            <button
                type="button"
                class="inline-flex h-9 items-center rounded-lg border border-[var(--color-secondary)]/40 px-3 text-sm"
                onclick={handleForceRefresh}
                disabled={isRefreshing ||
                    mutatingTaskId !== '' ||
                    completingTaskId !== ''}
            >
                {isRefreshing ? 'Refreshing…' : 'Force Refresh'}
            </button>
            <div class="flex flex-wrap gap-2 sm:justify-end">
                {#if isPickupSoon(viewData.garbage.nextTrashPickupDate)}
                    <span
                        class="inline-flex items-center gap-1 rounded-full border border-[var(--color-primary)]/30 bg-[var(--color-primary)]/10 px-3 py-1 text-xs font-semibold text-[var(--color-primary)]"
                    >
                        <Trash2 size={14} strokeWidth={2} aria-hidden="true" />
                        Take out trash tonight
                    </span>
                {/if}
                {#if isPickupSoon(viewData.garbage.nextRecyclingPickupDate)}
                    <span
                        class="inline-flex items-center gap-1 rounded-full border border-[var(--color-primary)]/30 bg-[var(--color-primary)]/10 px-3 py-1 text-xs font-semibold text-[var(--color-primary)]"
                    >
                        <Recycle size={14} strokeWidth={2} aria-hidden="true" />
                        Take out recycling tonight
                    </span>
                {/if}
                {#if !isPickupSoon(viewData.garbage.nextTrashPickupDate) && !isPickupSoon(viewData.garbage.nextRecyclingPickupDate)}
                    <span
                        class="rounded-full border border-[var(--color-secondary)]/30 bg-[var(--color-background)]/35 px-3 py-1 text-xs font-medium text-[var(--color-text)]/70"
                    >
                        No pickup scheduled in the next 24 hours
                    </span>
                {/if}
            </div>
        </div>
    </header>

    <div class="mt-4 grid gap-2 sm:grid-cols-2">
        <div
            class="rounded-xl border border-[var(--color-secondary)]/18 bg-[var(--color-background)]/35 p-3"
        >
            <p
                class="inline-flex items-center gap-1 text-xs font-medium uppercase tracking-wide text-[var(--color-text)]/70"
            >
                <Trash2 size={14} strokeWidth={2} aria-hidden="true" />
                Next Trash
            </p>
            <p class="mt-1 text-sm font-semibold">
                {formatOptionalDate(viewData.garbage.nextTrashPickupDate)}
            </p>
        </div>
        <div
            class="rounded-xl border border-[var(--color-secondary)]/18 bg-[var(--color-background)]/35 p-3"
        >
            <p
                class="inline-flex items-center gap-1 text-xs font-medium uppercase tracking-wide text-[var(--color-text)]/70"
            >
                <Recycle size={14} strokeWidth={2} aria-hidden="true" />
                Next Recycling
            </p>
            <p class="mt-1 text-sm font-semibold">
                {formatOptionalDate(viewData.garbage.nextRecyclingPickupDate)}
            </p>
        </div>
    </div>

    <div
        class="mt-4 rounded-xl border border-[var(--color-secondary)]/18 bg-[var(--color-background)]/35 p-3"
    >
        <p
            class="text-xs font-medium uppercase tracking-wide text-[var(--color-text)]/70"
        >
            Add TickTick Task
        </p>
        <div class="mt-2 grid gap-2 lg:grid-cols-[minmax(0,1fr)_auto_auto]">
            <input
                type="text"
                class="h-10 rounded-lg border border-[var(--color-secondary)]/26 bg-transparent px-3 text-sm"
                placeholder="Task title"
                bind:value={newTaskTitle}
                disabled={mutatingTaskId !== ''}
            />
            <div
                class="relative flex items-center gap-2"
                bind:this={newPickerContainer}
            >
                <button
                    type="button"
                    class="inline-flex h-10 items-center gap-2 rounded-lg border border-[var(--color-secondary)]/26 bg-transparent px-3 text-sm"
                    onclick={() => openPicker('new')}
                    disabled={mutatingTaskId !== ''}
                >
                    <Calendar size={14} class="text-white" />
                    {formatDateControl(newTaskDate)}
                </button>
                <input
                    type="time"
                    class="native-time h-10 w-24 rounded-lg border border-[var(--color-secondary)]/26 bg-transparent pl-2 pr-1.5 text-sm"
                    bind:value={newTaskTime}
                    disabled={mutatingTaskId !== ''}
                />

                {#if pickerTarget === 'new'}
                    <div
                        class="absolute right-0 top-12 z-20 w-64 max-w-[calc(100vw-2rem)] rounded-xl border border-[var(--color-secondary)]/35 bg-[var(--color-background)] p-3 shadow-lg"
                    >
                        <div class="mb-3 flex items-center justify-between">
                            <button
                                type="button"
                                class="rounded p-1 text-[var(--color-text)]/80 hover:bg-[var(--color-primary)]/10"
                                onclick={previousMonth}
                            >
                                <ChevronLeft size={14} />
                            </button>
                            <div class="flex items-center gap-1.5">
                                <select
                                    class="native-select rounded-md border border-[var(--color-secondary)]/30 bg-transparent px-1.5 py-1 pr-6 text-xs"
                                    value={String(pickerMonth.getMonth())}
                                    onchange={(event) =>
                                        setPickerMonthFromSelect(
                                            (
                                                event.currentTarget as HTMLSelectElement
                                            ).value,
                                        )}
                                >
                                    {#each monthLabels as monthLabel, monthIndex}
                                        <option value={String(monthIndex)}>
                                            {monthLabel}
                                        </option>
                                    {/each}
                                </select>
                                <select
                                    class="native-select rounded-md border border-[var(--color-secondary)]/30 bg-transparent px-1.5 py-1 pr-6 text-xs"
                                    value={String(pickerMonth.getFullYear())}
                                    onchange={(event) =>
                                        setPickerYearFromSelect(
                                            (
                                                event.currentTarget as HTMLSelectElement
                                            ).value,
                                        )}
                                >
                                    {#each pickerYearOptions() as yearOption}
                                        <option value={String(yearOption)}>
                                            {yearOption}
                                        </option>
                                    {/each}
                                </select>
                            </div>
                            <button
                                type="button"
                                class="rounded p-1 text-[var(--color-text)]/80 hover:bg-[var(--color-primary)]/10"
                                onclick={nextMonth}
                            >
                                <ChevronRight size={14} />
                            </button>
                        </div>
                        <div
                            class="grid grid-cols-7 gap-1 text-center text-xs text-[var(--color-text)]/60"
                        >
                            {#each weekdayLabels as label}
                                <span>{label}</span>
                            {/each}
                        </div>
                        <div class="mt-2 grid grid-cols-7 gap-1">
                            {#each calendarDays() as day}
                                {#if day === null}
                                    <span class="h-8"></span>
                                {:else}
                                    <button
                                        type="button"
                                        class={`h-8 rounded text-sm ${isSelectedDay(day) ? 'bg-[var(--color-primary)]/25 text-white' : 'hover:bg-[var(--color-primary)]/10'}`}
                                        onclick={() => setPickedDay(day)}
                                    >
                                        {day}
                                    </button>
                                {/if}
                            {/each}
                        </div>
                        <div class="mt-3 flex items-center gap-1">
                            {#each quickDateOptions as option}
                                <button
                                    type="button"
                                    class="rounded-md border border-[var(--color-secondary)]/30 px-2 py-1 text-[11px] hover:bg-[var(--color-primary)]/10"
                                    onclick={() =>
                                        applyQuickDate(option.offset)}
                                >
                                    {option.label}
                                </button>
                            {/each}
                        </div>
                    </div>
                {/if}
            </div>
            <button
                type="button"
                class="inline-flex h-10 items-center rounded-lg border border-[var(--color-primary)]/40 bg-[var(--color-primary)]/10 px-3 text-sm font-medium text-[var(--color-primary)]"
                onclick={handleCreateTask}
                disabled={isRefreshing || mutatingTaskId !== ''}
            >
                Add
            </button>
        </div>
    </div>

    {#if viewData.tasks.length > 0}
        <ul
            class="mt-4 overflow-hidden rounded-xl border border-[var(--color-secondary)]/18 divide-y divide-[var(--color-secondary)]/18"
        >
            {#each viewData.tasks as task}
                {@const typedTask = task as DailyTask}
                <li
                    class={`flex gap-2 bg-[var(--color-background)]/35 px-2.5 py-2.5 transition-colors duration-150 hover:bg-[var(--color-background)]/55 ${editingTaskId === typedTask.id ? 'items-start' : 'items-center'}`}
                >
                    <input
                        type="checkbox"
                        checked={typedTask.completed}
                        class="task-checkbox self-center"
                        disabled={typedTask.source !== 'ticktick' ||
                            completingTaskId !== ''}
                        onchange={() =>
                            handleTaskCompletion(
                                typedTask.id,
                                typedTask.source,
                            )}
                    />
                    <div class="min-w-0 flex-1">
                        {#if editingTaskId === typedTask.id}
                            <div
                                class="grid gap-2 lg:grid-cols-[minmax(0,1fr)_auto_auto_auto_auto]"
                            >
                                <input
                                    type="text"
                                    class="h-8 rounded-lg border border-[var(--color-secondary)]/30 bg-transparent px-2.5 text-sm"
                                    bind:value={editTaskTitle}
                                    disabled={mutatingTaskId !== ''}
                                />
                                <input
                                    type="time"
                                    class="native-time h-8 w-24 rounded-lg border border-[var(--color-secondary)]/30 bg-transparent pl-2 pr-1.5 text-sm"
                                    bind:value={editTaskTime}
                                    disabled={mutatingTaskId !== ''}
                                />
                                <div
                                    class="relative"
                                    bind:this={editPickerContainer}
                                >
                                    <button
                                        type="button"
                                        class="inline-flex h-8 items-center gap-1 rounded-lg border border-[var(--color-secondary)]/30 px-2.5 text-sm"
                                        onclick={() => openPicker('edit')}
                                        disabled={mutatingTaskId !== ''}
                                    >
                                        <Calendar
                                            size={12}
                                            class="text-white"
                                        />
                                        {formatDateControl(editTaskDate)}
                                    </button>

                                    {#if pickerTarget === 'edit'}
                                        <div
                                            class="absolute right-0 top-9 z-20 w-64 max-w-[calc(100vw-2rem)] rounded-xl border border-[var(--color-secondary)]/35 bg-[var(--color-background)] p-3 shadow-lg"
                                        >
                                            <div
                                                class="mb-3 flex items-center justify-between"
                                            >
                                                <button
                                                    type="button"
                                                    class="rounded p-1 text-[var(--color-text)]/80 hover:bg-[var(--color-primary)]/10"
                                                    onclick={previousMonth}
                                                >
                                                    <ChevronLeft size={14} />
                                                </button>
                                                <div
                                                    class="flex items-center gap-1.5"
                                                >
                                                    <select
                                                        class="native-select rounded-md border border-[var(--color-secondary)]/30 bg-transparent px-1.5 py-1 pr-6 text-xs"
                                                        value={String(
                                                            pickerMonth.getMonth(),
                                                        )}
                                                        onchange={(event) =>
                                                            setPickerMonthFromSelect(
                                                                (
                                                                    event.currentTarget as HTMLSelectElement
                                                                ).value,
                                                            )}
                                                    >
                                                        {#each monthLabels as monthLabel, monthIndex}
                                                            <option
                                                                value={String(
                                                                    monthIndex,
                                                                )}
                                                            >
                                                                {monthLabel}
                                                            </option>
                                                        {/each}
                                                    </select>
                                                    <select
                                                        class="native-select rounded-md border border-[var(--color-secondary)]/30 bg-transparent px-1.5 py-1 pr-6 text-xs"
                                                        value={String(
                                                            pickerMonth.getFullYear(),
                                                        )}
                                                        onchange={(event) =>
                                                            setPickerYearFromSelect(
                                                                (
                                                                    event.currentTarget as HTMLSelectElement
                                                                ).value,
                                                            )}
                                                    >
                                                        {#each pickerYearOptions() as yearOption}
                                                            <option
                                                                value={String(
                                                                    yearOption,
                                                                )}
                                                            >
                                                                {yearOption}
                                                            </option>
                                                        {/each}
                                                    </select>
                                                </div>
                                                <button
                                                    type="button"
                                                    class="rounded p-1 text-[var(--color-text)]/80 hover:bg-[var(--color-primary)]/10"
                                                    onclick={nextMonth}
                                                >
                                                    <ChevronRight size={14} />
                                                </button>
                                            </div>
                                            <div
                                                class="grid grid-cols-7 gap-1 text-center text-xs text-[var(--color-text)]/60"
                                            >
                                                {#each weekdayLabels as label}
                                                    <span>{label}</span>
                                                {/each}
                                            </div>
                                            <div
                                                class="mt-2 grid grid-cols-7 gap-1"
                                            >
                                                {#each calendarDays() as day}
                                                    {#if day === null}
                                                        <span class="h-8"
                                                        ></span>
                                                    {:else}
                                                        <button
                                                            type="button"
                                                            class={`h-8 rounded text-sm ${isSelectedDay(day) ? 'bg-[var(--color-primary)]/25 text-white' : 'hover:bg-[var(--color-primary)]/10'}`}
                                                            onclick={() =>
                                                                setPickedDay(
                                                                    day,
                                                                )}
                                                        >
                                                            {day}
                                                        </button>
                                                    {/if}
                                                {/each}
                                            </div>
                                            <div
                                                class="mt-3 flex items-center gap-1"
                                            >
                                                {#each quickDateOptions as option}
                                                    <button
                                                        type="button"
                                                        class="rounded-md border border-[var(--color-secondary)]/30 px-2 py-1 text-[11px] hover:bg-[var(--color-primary)]/10"
                                                        onclick={() =>
                                                            applyQuickDate(
                                                                option.offset,
                                                            )}
                                                    >
                                                        {option.label}
                                                    </button>
                                                {/each}
                                            </div>
                                        </div>
                                    {/if}
                                </div>
                                <button
                                    type="button"
                                    class="inline-flex h-8 items-center rounded-lg border border-[var(--color-primary)]/40 bg-[var(--color-primary)]/10 px-3 text-sm font-medium text-[var(--color-primary)]"
                                    onclick={() => saveTaskEdit(typedTask.id)}
                                    disabled={mutatingTaskId !== ''}
                                >
                                    Save
                                </button>
                                <button
                                    type="button"
                                    class="inline-flex h-8 items-center rounded-lg border border-[var(--color-secondary)]/40 px-3 text-sm"
                                    onclick={cancelEditingTask}
                                    disabled={mutatingTaskId !== ''}
                                >
                                    Cancel
                                </button>
                            </div>
                        {:else}
                            <div
                                class="flex min-h-8 flex-wrap items-center gap-x-2 gap-y-0.5"
                            >
                                {#if typedTask.source === 'ticktick'}
                                    <button
                                        type="button"
                                        class="flex min-h-8 min-w-0 flex-wrap items-center gap-x-2 gap-y-0.5 rounded-md px-1 text-left"
                                        onclick={() =>
                                            startEditingTask(typedTask)}
                                        disabled={mutatingTaskId !== '' ||
                                            completingTaskId !== ''}
                                    >
                                        <span
                                            class="truncate text-sm font-medium leading-tight"
                                        >
                                            {typedTask.title}
                                        </span>
                                        <span
                                            class="text-xs text-[var(--color-text)]/70"
                                        >
                                            •
                                        </span>
                                        <span
                                            class={`text-xs ${isTaskOverdue(typedTask) ? 'text-[var(--color-error)]' : 'text-[var(--color-text)]/70'}`}
                                        >
                                            Due {formatDate(typedTask.dueAt, {
                                                includeTimeForToday: true,
                                                hasTime: typedTask.hasTime,
                                            })}
                                        </span>
                                    </button>
                                {:else}
                                    <p
                                        class="truncate text-sm font-medium leading-tight"
                                    >
                                        {typedTask.title}
                                    </p>
                                    <span
                                        class="text-xs text-[var(--color-text)]/70"
                                    >
                                        •
                                    </span>
                                    <span
                                        class={`text-xs ${isTaskOverdue(typedTask) ? 'text-[var(--color-error)]' : 'text-[var(--color-text)]/70'}`}
                                    >
                                        Due {formatDate(typedTask.dueAt, {
                                            includeTimeForToday: true,
                                            hasTime: typedTask.hasTime,
                                        })}
                                    </span>
                                {/if}
                            </div>
                        {/if}
                    </div>
                </li>
            {/each}
        </ul>
        {#if completionError}
            <p class="mt-3 text-sm text-[var(--color-error)]">
                {completionError}
            </p>
        {/if}
        {#if mutationError}
            <p class="mt-3 text-sm text-[var(--color-error)]">
                {mutationError}
            </p>
        {/if}
    {:else}
        <p
            class="mt-4 rounded-xl border border-[var(--color-secondary)]/18 bg-[var(--color-background)]/35 p-3 text-sm text-[var(--color-text)]/70"
        >
            No daily tasks available yet.
        </p>
        {#if mutationError}
            <p class="mt-3 text-sm text-[var(--color-error)]">
                {mutationError}
            </p>
        {/if}
    {/if}
</section>

<style>
    :global(.native-select),
    :global(.native-time) {
        color-scheme: dark;
        color: var(--color-text);
    }

    :global(.native-time) {
        font-variant-numeric: tabular-nums;
        line-height: 1.1;
    }

    :global(.native-time::-webkit-datetime-edit) {
        padding: 0;
    }

    :global(.native-time::-webkit-datetime-edit-fields-wrapper) {
        padding: 0;
        display: inline-flex;
        align-items: center;
        gap: 0;
    }

    :global(.native-time::-webkit-datetime-edit-hour-field),
    :global(.native-time::-webkit-datetime-edit-minute-field),
    :global(.native-time::-webkit-datetime-edit-ampm-field) {
        padding: 0;
    }

    :global(.native-select) {
        appearance: none;
        -webkit-appearance: none;
        -moz-appearance: none;
        background-image: url("data:image/svg+xml,%3Csvg xmlns='http://www.w3.org/2000/svg' width='12' height='12' viewBox='0 0 24 24' fill='none' stroke='white' stroke-width='2.2' stroke-linecap='round' stroke-linejoin='round'%3E%3Cpolyline points='6 9 12 15 18 9'/%3E%3C/svg%3E");
        background-repeat: no-repeat;
        background-position: right 0.45rem center;
        background-size: 0.7rem;
    }

    :global(.native-select::-ms-expand) {
        display: none;
    }

    :global(.native-select option) {
        background: var(--color-background);
        color: var(--color-text);
    }

    :global(.native-time::-webkit-calendar-picker-indicator) {
        filter: brightness(0) invert(1);
        opacity: 1;
        cursor: pointer;
        margin: 0 0 0 0.1rem;
        padding: 0;
        width: 0.72rem;
    }

    :global(.task-checkbox) {
        appearance: none;
        -webkit-appearance: none;
        width: 1rem;
        height: 1rem;
        border-radius: 0.35rem;
        border: 1.5px solid
            color-mix(in oklab, var(--color-text) 35%, transparent);
        background: color-mix(
            in oklab,
            var(--color-background) 70%,
            transparent
        );
        cursor: pointer;
        position: relative;
        transition:
            border-color 120ms ease,
            background-color 120ms ease,
            box-shadow 120ms ease;
    }

    :global(.task-checkbox:checked) {
        background: var(--color-primary);
        border-color: var(--color-primary);
        box-shadow: 0 0 0 1px
            color-mix(in oklab, var(--color-primary) 35%, transparent);
    }

    :global(.task-checkbox:checked::after) {
        content: '';
        position: absolute;
        left: 0.29rem;
        top: 0.08rem;
        width: 0.22rem;
        height: 0.5rem;
        border: solid #fff;
        border-width: 0 2px 2px 0;
        transform: rotate(45deg);
    }

    :global(.task-checkbox:disabled) {
        opacity: 0.55;
        cursor: not-allowed;
    }

    :global(.task-checkbox:focus-visible) {
        outline: none;
        box-shadow: 0 0 0 2px
            color-mix(in oklab, var(--color-primary) 45%, transparent);
    }
</style>
