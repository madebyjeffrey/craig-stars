<script lang="ts">
	import { page } from '$app/stores';
	import { getGameContext } from '$lib/services/GameContext';
	import type { CommandedFleet, Fleet } from '$lib/types/Fleet';
	import MergeFleets from '../../../dialogs/merge/MergeFleets.svelte';

	const { game, player, universe, commandedFleet, commandMapObject, merge } = getGameContext();
	let num = parseInt($page.params.num);

	let fleetsInOrbit: Fleet[] = [];

	$: {
		if ($commandedFleet && $commandedFleet.num === num) {
			fleetsInOrbit = $universe
				.getMyFleetsByPosition($commandedFleet)
				.filter((mo) => mo.num !== $commandedFleet?.num) as Fleet[];
		} else {
			const fleet = $universe.getFleet($player.num, num);
			if (fleet) {
				commandMapObject(fleet);
			}
		}
	}

</script>

{#if $commandedFleet}
	<MergeFleets
		fleet={$commandedFleet}
		otherFleetsHere={fleetsInOrbit}
		on:ok={(e) => merge(e.detail.fleet, e.detail.fleetNums)}
	/>
{/if}
