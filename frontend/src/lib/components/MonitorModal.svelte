<script lang="ts">
	import { createEventDispatcher } from 'svelte';
	import { fetchAPI } from '$lib/api/client';

	export let isOpen = false;
	export let monitor: any = null;

	const dispatch = createEventDispatcher();

	interface FormData {
		name: string;
		url: string;
		type: string;
		keyword: string;
		check_interval: number;
		timeout: number;
		enabled: boolean;
		check_ssl: boolean;
		ssl_alert_days: number;
	}

	let formData: FormData = {
		name: '',
		url: '',
		type: 'http',
		keyword: '',
		check_interval: 60,
		timeout: 30,
		enabled: true,
		check_ssl: true,
		ssl_alert_days: 30
	};

	let errors: Record<string, string> = {};
	let isSubmitting = false;
	let lastMonitorId: string | null = null;

	// Only update formData when monitor actually changes (different monitor or switching between create/edit)
	$: if (monitor && monitor.id !== lastMonitorId) {
		formData = {
			name: monitor.name || '',
			url: monitor.url || '',
			type: monitor.type || 'http',
			keyword: monitor.keyword || '',
			check_interval: monitor.check_interval || 60,
			timeout: monitor.timeout || 30,
			enabled: monitor.enabled !== undefined ? monitor.enabled : true,
			check_ssl: monitor.check_ssl !== undefined ? monitor.check_ssl : true,
			ssl_alert_days: monitor.ssl_alert_days || 30
		};
		lastMonitorId = monitor?.id || null;
	} else if (!monitor && lastMonitorId !== null) {
		// Switching from edit to create mode
		formData = {
			name: '',
			url: '',
			type: 'http',
			keyword: '',
			check_interval: 60,
			timeout: 30,
			enabled: true,
			check_ssl: true,
			ssl_alert_days: 30
		};
		lastMonitorId = null;
	}

	$: isEditMode = !!monitor;

	// Reset SSL settings/Keyword when switching to TCP or Ping
	$: if (formData.type === 'tcp' || formData.type === 'icmp') {
		formData.check_ssl = false;
        formData.keyword = '';
	}

	function validateForm(): boolean {
		errors = {};

		if (!formData.name.trim()) {
			errors.name = 'Name is required';
		}

		if (!formData.url.trim()) {
			errors.url = formData.type === 'tcp' ? 'Host:Port is required' : 'Address is required';
		} else if (formData.type === 'http') {
			try {
				new URL(formData.url);
			} catch {
				errors.url = 'Invalid URL format';
			}
		} else if (formData.type === 'tcp') {
			// Basic validation for host:port format
			const tcpPattern = /^([^:]+):(\d+)$/;
			if (!tcpPattern.test(formData.url)) {
				errors.url = 'Invalid Host:Port format. Use format: host:port';
			}
		} else if (formData.type === 'icmp') {
            if (formData.url.includes('://')) {
                errors.url = 'Enter a hostname or IP address (no protocol)';
            }
        }

		if (formData.check_interval < 60) {
			errors.check_interval = 'Check interval must be at least 60 seconds';
		}

		if (formData.timeout < 5 || formData.timeout > 120) {
			errors.timeout = 'Timeout must be between 5 and 120 seconds';
		}

		if (formData.check_ssl && formData.ssl_alert_days < 1) {
			errors.ssl_alert_days = 'SSL alert days must be at least 1';
		}

		return Object.keys(errors).length === 0;
	}

	async function handleSubmit() {
		if (!validateForm()) {
			return;
		}

		isSubmitting = true;

		try {
			const url = isEditMode ? `/api/v1/monitors/${monitor.id}` : '/api/v1/monitors';
			const method = isEditMode ? 'PUT' : 'POST';

			const response = await fetchAPI(url, {
				method,
				body: JSON.stringify(formData)
			});

			if (!response.ok) {
				const data = await response.json();
				throw new Error(data.error || 'Failed to save monitor');
			}

			const savedMonitor = await response.json();
			dispatch('save', savedMonitor);
			handleClose();
		} catch (err: any) {
			console.error('Error saving monitor:', err);
			alert(err.message || 'Failed to save monitor');
		} finally {
			isSubmitting = false;
		}
	}

	function handleClose() {
		dispatch('close');
	}

    function getPlaceholderForType(type: string): string {
        switch (type) {
            case 'http': return 'https://example.com';
            case 'tcp': return 'example.com:80';
            case 'icmp': return 'example.com';
            default: return '';
        }
    }

    function getLabelForType(type: string): string {
        switch (type) {
            case 'http': return 'URL';
            case 'tcp': return 'Host:Port';
            case 'icmp': return 'Hostname / IP';
            default: return 'Address';
        }
    }
</script>

{#if isOpen}
	<div class="fixed inset-0 z-50 overflow-y-auto" aria-labelledby="modal-title" role="dialog" aria-modal="true">
		<div class="flex items-end justify-center min-h-screen pt-4 px-4 pb-20 text-center sm:block sm:p-0">
			<!-- Backdrop -->
			<div 
                class="fixed inset-0 bg-slate-900/75 transition-opacity backdrop-blur-sm" 
                aria-hidden="true"
                on:click={handleClose}
            ></div>

			<!-- This element is to trick the browser into centering the modal contents. -->
			<span class="hidden sm:inline-block sm:align-middle sm:h-screen" aria-hidden="true">&#8203;</span>

			<div class="inline-block align-bottom bg-white dark:bg-slate-800 rounded-xl text-left overflow-hidden shadow-xl transform transition-all sm:my-8 sm:align-middle sm:max-w-lg w-full border border-slate-200 dark:border-slate-700">
				<div class="bg-white dark:bg-slate-800 px-4 pt-5 pb-4 sm:p-6 sm:pb-4">
					<div class="sm:flex sm:items-start">
						<div class="mx-auto flex-shrink-0 flex items-center justify-center h-12 w-12 rounded-full bg-blue-100 dark:bg-blue-900/30 sm:mx-0 sm:h-10 sm:w-10">
                            {#if isEditMode}
                                <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor" class="w-6 h-6 text-blue-600 dark:text-blue-400">
                                    <path stroke-linecap="round" stroke-linejoin="round" d="M16.862 4.487l1.687-1.688a1.875 1.875 0 112.652 2.652L10.582 16.07a4.5 4.5 0 01-1.897 1.13L6 18l.8-2.685a4.5 4.5 0 011.13-1.897l8.932-8.931zm0 0L19.5 7.125M18 14v4.75A2.25 2.25 0 0115.75 21H5.25A2.25 2.25 0 013 18.75V8.25A2.25 2.25 0 015.25 6H10" />
                                </svg>
                            {:else}
                                <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor" class="w-6 h-6 text-blue-600 dark:text-blue-400">
                                    <path stroke-linecap="round" stroke-linejoin="round" d="M12 4.5v15m7.5-7.5h-15" />
                                </svg>
                            {/if}
						</div>
						<div class="mt-3 text-center sm:mt-0 sm:ml-4 sm:text-left w-full">
							<h3 class="text-lg leading-6 font-medium text-slate-900 dark:text-gray-100" id="modal-title">
								{isEditMode ? 'Edit Monitor' : 'Add New Monitor'}
							</h3>
							<div class="mt-4 space-y-4">
								<!-- Name -->
								<div>
									<label for="name" class="block text-sm font-medium text-slate-700 dark:text-slate-300">Name</label>
									<input
										type="text"
										name="name"
										id="name"
										bind:value={formData.name}
										class="mt-1 block w-full border-slate-300 dark:border-slate-600 dark:bg-slate-900/50 dark:text-gray-100 rounded-lg shadow-sm focus:ring-blue-500 focus:border-blue-500 sm:text-sm px-3 py-2 border"
										placeholder="My Website"
									/>
									{#if errors.name}
										<p class="mt-1 text-sm text-red-600">{errors.name}</p>
									{/if}
								</div>

								<!-- Monitor Type -->
								<div>
									<label for="type" class="block text-sm font-medium text-slate-700 dark:text-slate-300">Monitor Type</label>
									<select
										name="type"
										id="type"
										bind:value={formData.type}
										class="mt-1 block w-full border-slate-300 dark:border-slate-600 dark:bg-slate-900/50 dark:text-gray-100 rounded-lg shadow-sm focus:ring-blue-500 focus:border-blue-500 sm:text-sm px-3 py-2 border"
									>
										<option value="http">HTTP/HTTPS</option>
										<option value="tcp">TCP</option>
                                        <option value="icmp">Ping (ICMP)</option>
									</select>
								</div>

								<!-- Address Input -->
								<div>
									<label for="url" class="block text-sm font-medium text-slate-700 dark:text-slate-300">
										{getLabelForType(formData.type)}
									</label>
									<input
										type="text"
										name="url"
										id="url"
										bind:value={formData.url}
										class="mt-1 block w-full border-slate-300 dark:border-slate-600 dark:bg-slate-900/50 dark:text-gray-100 rounded-lg shadow-sm focus:ring-blue-500 focus:border-blue-500 sm:text-sm px-3 py-2 border"
										placeholder={getPlaceholderForType(formData.type)}
									/>
									{#if errors.url}
										<p class="mt-1 text-sm text-red-600">{errors.url}</p>
									{/if}
								</div>

                                <!-- Keyword Search (HTTP only) -->
                                {#if formData.type === 'http'}
                                    <div>
                                        <label for="keyword" class="block text-sm font-medium text-slate-700 dark:text-slate-300">
                                            Keyword Search (Optional)
                                        </label>
                                        <input
                                            type="text"
                                            name="keyword"
                                            id="keyword"
                                            bind:value={formData.keyword}
                                            class="mt-1 block w-full border-slate-300 dark:border-slate-600 dark:bg-slate-900/50 dark:text-gray-100 rounded-lg shadow-sm focus:ring-blue-500 focus:border-blue-500 sm:text-sm px-3 py-2 border"
                                            placeholder="e.g. 'Welcome' or 'Status: OK'"
                                        />
                                        <p class="mt-1 text-xs text-slate-500 dark:text-slate-400">
                                            If set, the monitor will be considered DOWN if this keyword is not found in the response body.
                                        </p>
                                    </div>
                                {/if}

								<div class="grid grid-cols-2 gap-4">
									<!-- Check Interval -->
									<div>
										<label for="check_interval" class="block text-sm font-medium text-slate-700 dark:text-slate-300">Interval (sec)</label>
										<input
											type="number"
											name="check_interval"
											id="check_interval"
											bind:value={formData.check_interval}
											min="60"
											class="mt-1 block w-full border-slate-300 dark:border-slate-600 dark:bg-slate-900/50 dark:text-gray-100 rounded-lg shadow-sm focus:ring-blue-500 focus:border-blue-500 sm:text-sm px-3 py-2 border"
										/>
										{#if errors.check_interval}
											<p class="mt-1 text-sm text-red-600">{errors.check_interval}</p>
										{/if}
									</div>

									<!-- Timeout -->
									<div>
										<label for="timeout" class="block text-sm font-medium text-slate-700 dark:text-slate-300">Timeout (sec)</label>
										<input
											type="number"
											name="timeout"
											id="timeout"
											bind:value={formData.timeout}
											min="5"
											max="120"
											class="mt-1 block w-full border-slate-300 dark:border-slate-600 dark:bg-slate-900/50 dark:text-gray-100 rounded-lg shadow-sm focus:ring-blue-500 focus:border-blue-500 sm:text-sm px-3 py-2 border"
										/>
										{#if errors.timeout}
											<p class="mt-1 text-sm text-red-600">{errors.timeout}</p>
										{/if}
									</div>
								</div>

								<!-- SSL Settings -->
								{#if formData.type === 'http'}
									<div class="space-y-3 pt-2">
										<div class="flex items-start">
											<div class="flex items-center h-5">
												<input
													id="check_ssl"
													name="check_ssl"
													type="checkbox"
													bind:checked={formData.check_ssl}
													class="focus:ring-blue-500 h-4 w-4 text-blue-600 border-slate-300 rounded"
												/>
											</div>
											<div class="ml-3 text-sm">
												<label for="check_ssl" class="font-medium text-slate-700 dark:text-slate-300">Check SSL Certificate</label>
												<p class="text-slate-500 dark:text-slate-400">Monitor SSL certificate validity and expiration.</p>
											</div>
										</div>

										{#if formData.check_ssl}
											<div>
												<label for="ssl_alert_days" class="block text-sm font-medium text-slate-700 dark:text-slate-300">Alert before expiry (days)</label>
												<input
													type="number"
													name="ssl_alert_days"
													id="ssl_alert_days"
													bind:value={formData.ssl_alert_days}
													min="1"
													class="mt-1 block w-full border-slate-300 dark:border-slate-600 dark:bg-slate-900/50 dark:text-gray-100 rounded-lg shadow-sm focus:ring-blue-500 focus:border-blue-500 sm:text-sm px-3 py-2 border"
												/>
												{#if errors.ssl_alert_days}
													<p class="mt-1 text-sm text-red-600">{errors.ssl_alert_days}</p>
												{/if}
											</div>
										{/if}
									</div>
								{/if}

                                <!-- Enabled -->
                                <div class="flex items-start">
                                    <div class="flex items-center h-5">
                                        <input
                                            id="enabled"
                                            name="enabled"
                                            type="checkbox"
                                            bind:checked={formData.enabled}
                                            class="focus:ring-blue-500 h-4 w-4 text-blue-600 border-slate-300 rounded"
                                        />
                                    </div>
                                    <div class="ml-3 text-sm">
                                        <label for="enabled" class="font-medium text-slate-700 dark:text-slate-300">Enabled</label>
                                        <p class="text-slate-500 dark:text-slate-400">Pause monitoring without deleting the configuration.</p>
                                    </div>
                                </div>
							</div>
						</div>
					</div>
				</div>
				<div class="bg-slate-50 dark:bg-slate-950/40 px-4 py-3 sm:px-6 sm:flex sm:flex-row-reverse border-t border-slate-200 dark:border-slate-700">
					<button
						type="button"
						class="w-full inline-flex justify-center rounded-lg border border-transparent shadow-sm px-4 py-2 bg-blue-600 text-base font-medium text-white hover:bg-blue-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-blue-500 sm:ml-3 sm:w-auto sm:text-sm disabled:opacity-50 disabled:cursor-not-allowed transition-colors"
						on:click={handleSubmit}
						disabled={isSubmitting}
					>
						{isSubmitting ? 'Saving...' : 'Save'}
					</button>
					<button
						type="button"
						class="mt-3 w-full inline-flex justify-center rounded-lg border border-slate-300 dark:border-slate-600 shadow-sm px-4 py-2 bg-white dark:bg-slate-900/50 text-base font-medium text-slate-700 dark:text-slate-300 hover:bg-slate-50 dark:hover:bg-slate-800 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-blue-500 sm:mt-0 sm:ml-3 sm:w-auto sm:text-sm transition-colors"
						on:click={handleClose}
						disabled={isSubmitting}
					>
						Cancel
					</button>
				</div>
			</div>
		</div>
	</div>
{/if}
