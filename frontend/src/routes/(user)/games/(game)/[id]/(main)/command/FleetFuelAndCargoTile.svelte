<script lang="ts">
	import CargoBar from '$lib/components/game/CargoBar.svelte';
	import FuelBar from '$lib/components/game/FuelBar.svelte';
	import { getGameContext } from '$lib/services/GameContext';
	import { canTransferCargo, type CommandedFleet } from '$lib/types/Fleet';
	import { createEventDispatcher } from 'svelte';
	import type { CargoTransferDialogEvent } from '../../dialogs/cargo/CargoTranfserDialog.svelte';
	import CommandTile from './CommandTile.svelte';

	const dispatch = createEventDispatcher<CargoTransferDialogEvent>();
	const { game, player, universe } = getGameContext();

	export let fleet: CommandedFleet;

	const transfer = () => {
		dispatch('cargo-transfer-dialog', {
			src: fleet,
			dest: fleet.getCargoTransferTarget($universe)
		});
	};
</script>

{#if fleet?.spec}
	<CommandTile title="Fuel & Cargo">
		<div class="flex justify-between my-1">
			<div class="w-12 text-tile-item-title">Fuel</div>
			<div class="ml-1 h-full w-full">
				<FuelBar value={fleet.fuel} capacity={fleet.spec.fuelCapacity} />
			</div>
		</div>

		<div class="flex justify-between my-1">
			<div class="w-12 text-tile-item-title">Cargo</div>
			<div class="ml-1 h-full w-full">
				<CargoBar
					on:cargo-transfer-dialog={() => transfer()}
					canTransferCargo={canTransferCargo(fleet, $universe)}
					value={fleet.cargo}
					capacity={fleet.spec.cargoCapacity}
				/>
			</div>
		</div>
		<div class="flex justify-between">
			<div class="text-tile-item-title text-ironium">Ironium</div>
			<div>{fleet.cargo.ironium ?? 0}kT</div>
		</div>
		<div class="flex justify-between">
			<div class="text-tile-item-title text-boranium">Boranium</div>
			<div>{fleet.cargo.boranium ?? 0}kT</div>
		</div>
		<div class="flex justify-between">
			<div class="text-tile-item-title text-germanium">Germanium</div>
			<div>{fleet.cargo.germanium ?? 0}kT</div>
		</div>
		<div class="flex justify-between">
			<div class="text-tile-item-title text-colonists">Colonists</div>
			<div>{fleet.cargo.colonists ?? 0}kT</div>
		</div>
	</CommandTile>
{/if}
