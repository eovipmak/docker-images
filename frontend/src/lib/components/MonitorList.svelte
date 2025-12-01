<script lang="ts">
  import MonitorTable from './MonitorTable.svelte';
  import MonitorCard from './MonitorCard.svelte';
  import type { Monitor } from '$lib/types';
  export let monitors: Monitor[] = [];
  export let useTable: boolean = false;
  import { createEventDispatcher } from 'svelte';
  const dispatch = createEventDispatcher();

  function handleView(monitor: Monitor) {
    console.debug('[MonitorList] dispatch view', monitor?.id);
    dispatch('view', monitor);
  }
  function handleEdit(monitor: Monitor) {
    dispatch('edit', monitor);
  }
  function handleDelete(monitor: Monitor) {
    dispatch('delete', monitor);
  }
</script>

{#if useTable}
  <MonitorTable {monitors} on:view={(e) => handleView(e.detail)} on:edit={(e) => handleEdit(e.detail)} on:delete={(e) => handleDelete(e.detail)} />
{:else}
  <div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6" data-testid="monitors-grid">
    {#each monitors as monitor (monitor.id)}
      <div class="group">
        <MonitorCard monitor={monitor} on:view={(e) => handleView(e.detail)} on:edit={(e) => handleEdit(e.detail)} on:delete={(e) => handleDelete(e.detail)} />
      </div>
    {/each}
  </div>
{/if}
