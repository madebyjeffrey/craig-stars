<script lang="ts" context="module">
	import { type Fleet } from '$lib/types/Fleet';
	import { Unexplored, type Planet } from '$lib/types/Planet';

	export type Results = {
		planets: Planet[];
		fleets: Fleet[];
		mysteryTraders: MysteryTrader[];
	};

	export type SearchResultsEvent = {
		ok: MapObject | undefined;
		cancel: void;
	};
</script>

<script lang="ts">
	import MineralMini from '$lib/components/game/MineralMini.svelte';
	import { getGameContext } from '$lib/services/GameContext';
	import { getMapObjectName, None, owned, ownedBy, type MapObject } from '$lib/types/MapObject';
	import type { MysteryTrader } from '$lib/types/MysteryTrader';
	import { createEventDispatcher, onMount } from 'svelte';

	const { game, player, universe, settings, commandMapObject, selectMapObject, zoomToMapObject } =
		getGameContext();
	const dispatch = createEventDispatcher<SearchResultsEvent>();

	export let maxPlanetResults = 10;
	export let maxFleetResults = 10;
	export let maxMiscResults = 10;

	// the currently selected item
	$: selectedItemIndex = 0;
	$: selectedItem =
		selectedItemIndex < results.planets.length
			? results.planets[selectedItemIndex]
			: selectedItemIndex < results.planets.length + results.fleets.length
				? results.fleets[selectedItemIndex - results.planets.length]
				: selectedItemIndex <
					  results.planets.length + results.fleets.length + results.mysteryTraders.length
					? results.mysteryTraders[
							selectedItemIndex - results.planets.length + results.fleets.length
						]
					: undefined;

	function getResults(search: string): Results {
		if (search == '') {
			return {
				planets: [],
				fleets: [],
				mysteryTraders: []
			};
		}
		const terms = search.split(' ');

		const planets = $universe.getPlanets($settings.sortPlanetsKey, $settings.sortPlanetsDescending);
		const fleets = $universe.getFleets($settings.sortFleetsKey, $settings.sortFleetsDescending);
		const mysteryTraders = $universe.mysteryTraders;

		// return true if a mapboject name or player matches a search term
		const termSearch = (term: string, mo: MapObject): boolean =>
			mo.name.toLowerCase().indexOf(term.toLowerCase()) != -1 ||
			(mo.playerNum != None &&
				$universe.getPlayerPluralName(mo.playerNum).toLowerCase().indexOf(term.toLowerCase()) !=
					-1);

		// reset the selected item when the search is updated
		selectedItemIndex = 0;
		return {
			planets:
				planets
					.filter((i) => terms.every((term) => termSearch(term, i)))
					.slice(0, maxPlanetResults) ?? [],
			fleets:
				fleets
					.filter((i) => terms.every((term) => termSearch(term, i)))
					.slice(0, maxFleetResults) ?? [],

			mysteryTraders:
				mysteryTraders
					.filter((i) => terms.every((term) => termSearch(term, i)))
					.slice(0, maxMiscResults) ?? []
		};
	}

	function ok() {
		dispatch('ok', selectedItem);
	}
	function cancel() {
		dispatch('cancel');
	}

	function selectPrevious() {
		selectedItemIndex = Math.max(0, selectedItemIndex - 1);
	}

	function selectNext() {
		selectedItemIndex = Math.min(
			results.planets.length + results.fleets.length + results.mysteryTraders.length,
			selectedItemIndex + 1
		);
	}

	function onSearchKeyDown(event: KeyboardEvent) {
		switch (event.key) {
			case 'ArrowDown':
				// Do something for "down arrow" key press.
				selectNext();
				event.preventDefault();
				break;
			case 'ArrowUp':
				// Do something for "up arrow" key press.
				selectPrevious();
				event.preventDefault();
				break;
			case 'Enter':
				ok();
				event.preventDefault();
				break;
			case 'Escape':
				if ($settings.searchQuery != '') {
					$settings.searchQuery = '';
				} else {
					cancel();
					event.preventDefault();
				}
				break;
		}
	}

	let searchInput: HTMLInputElement | undefined;
	onMount(() => {
		searchInput?.focus();
	});
	// when search chnages, update our search results
	$: results = getResults($settings.searchQuery);
</script>

<div class="flex flex-col gap-1 h-full pb-2">
	<input
		type="search"
		name="search"
		placeholder="search"
		class="input input-bordered input-sm sm:w-auto mt-1"
		autocomplete="off"
		autocorrect="off"
		autocapitalize="off"
		spellcheck="false"
		bind:this={searchInput}
		bind:value={$settings.searchQuery}
		on:keydown={onSearchKeyDown}
		on:focus={() => searchInput?.select()}
	/>
	<div class="h-full">
		<div class="mt-2 w-full h-full bg-base-200 border-2 border-base-300 overflow-y-auto pl-2">
			{#if results.planets.length > 0}
				<h3 class="text-2xl font-bold mb-1">Planets</h3>
				<ul class="mx-1">
					{#each results.planets as planet, index}
						<!-- svelte-ignore a11y-mouse-events-have-key-events -->
						<li
							class="rounded-lg px-2"
							class:bg-primary={selectedItemIndex == index}
							on:mouseover={(e) => (selectedItemIndex = index)}
						>
							<button class="text-xl text-left w-full" on:click={ok}>
								<div class="flex flex-row gap-1">
									{#if planet.playerNum != None}
										<span style={`color: ${$universe.getPlayerColor(planet.playerNum)}`}
											>{$universe.getPlayerPluralName(planet.playerNum)}</span
										>
										{planet.name}
									{:else}
										{planet.name}
									{/if}
									{#if planet.reportAge != Unexplored}
										{#if owned(planet)}
											<div>-</div>
											<div class="text-base my-auto">
												{planet.spec.population
													? planet.spec.population.toLocaleString() + ' pop'
													: ''}
											</div>
										{/if}
										<div>-</div>
										<div class="text-base my-auto">
											{#if planet.spec.canTerraform}
												<span
													class:text-habitable={(planet.spec.habitability ?? 0) > 0}
													class:text-uninhabitable={(planet.spec.habitability ?? 0) < 0}
													>{planet.spec.habitability ?? 0}%</span
												>
												/
												<span class="text-terraformable"
													>{planet.spec.terraformedHabitability ?? 0}%</span
												>
											{:else}
												<span
													class:text-habitable={(planet.spec.habitability ?? 0) > 0}
													class:text-uninhabitable={(planet.spec.habitability ?? 0) < 0}
												>
													{planet.spec.habitability ?? 0}%</span
												>
											{/if}
										</div>
										{#if ownedBy(planet, $player.num)}
											<div>-</div>
											<div class="text-base my-auto">
												{planet.spec.resourcesPerYear
													? planet.spec.resourcesPerYear.toLocaleString() + ' res'
													: ''}
											</div>
											<div>-</div>
											<div class="text-base my-auto">
												<MineralMini mineral={planet.cargo} />
											</div>
										{/if}
									{/if}
								</div>
							</button>
						</li>
					{/each}
				</ul>
			{/if}
			{#if results.fleets.length > 0}
				<h3 class="text-2xl font-bold mb-1">Fleets</h3>
				<ul class="mx-1">
					{#each results.fleets as fleet, index}
						<!-- svelte-ignore a11y-mouse-events-have-key-events -->
						<li
							class="rounded-lg px-2"
							class:bg-primary={selectedItemIndex == results.planets.length + index}
							on:mouseover={(e) => (selectedItemIndex = results.planets.length + index)}
						>
							<button class="text-xl text-left w-full" on:click={ok}>
								<span style={`color: ${$universe.getPlayerColor(fleet.playerNum)}`}
									>{$universe.getPlayerPluralName(fleet.playerNum)}</span
								>
								{getMapObjectName(fleet)}
							</button>
						</li>
					{/each}
				</ul>
			{/if}
			{#if results.mysteryTraders.length > 0}
				<h3 class="text-2xl font-bold mb-1">Mystery Traders</h3>
				<ul class="mx-1">
					{#each results.mysteryTraders as mysterytrader, index}
						<!-- svelte-ignore a11y-mouse-events-have-key-events -->
						<li
							class="rounded-lg px-2"
							class:bg-primary={selectedItemIndex ==
								results.planets.length + results.fleets.length + index}
							on:mouseover={(e) =>
								(selectedItemIndex = results.planets.length + results.fleets.length + index)}
						>
							<button class="text-xl text-left w-full" on:click={ok}>
								<span class="text-mystery-trader"> {mysterytrader.name}</span></button
							>
						</li>
					{/each}
				</ul>
			{/if}
		</div>
	</div>
</div>
