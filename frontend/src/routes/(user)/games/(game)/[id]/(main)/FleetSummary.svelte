<script lang="ts">
	import CargoBar from '$lib/components/game/CargoBar.svelte';
	import FuelBar from '$lib/components/game/FuelBar.svelte';
	import { onShipDesignTooltip } from '$lib/components/game/tooltips/ShipDesignTooltip.svelte';
	import { getGameContext } from '$lib/services/GameContext';
	import {
		getDamagePercentForToken,
		type Fleet,
		WaypointTask,
		canTransferCargo,
		CommandedFleet
	} from '$lib/types/Fleet';
	import { StargateWarpSpeed, ownedBy } from '$lib/types/MapObject';
	import type { ShipDesign } from '$lib/types/ShipDesign';
	import { kebabCase, startCase } from 'lodash-es';
	import { createEventDispatcher } from 'svelte';
	import type { CargoTransferDialogEvent } from '../dialogs/cargo/CargoTranfserDialog.svelte';

	const dispatch = createEventDispatcher<CargoTransferDialogEvent>();
	const { player, universe } = getGameContext();

	export let fleet: Fleet;

	let design: ShipDesign | undefined;

	function getIcon(fleet: Fleet): string {
		if (fleet.tokens && fleet.tokens.length > 0) {
			const designNum = fleet.tokens[0].designNum;
			design = $universe.getDesign(fleet.playerNum, designNum);
			if (design) {
				return `hull-${kebabCase(design.hull)}-${design.hullSetNumber ?? 0}`;
			}
		}
		return '';
	}

	// get either warpSpeed as a number, or "stargate"
	function getWarpSpeed(fleet: Fleet): string {
		const warpSpeed: number =
			(fleet?.waypoints && fleet.waypoints.length > 1
				? fleet.waypoints[1].warpSpeed
				: fleet.warpSpeed) ?? 0;

		if (warpSpeed == StargateWarpSpeed) {
			return 'Use Stargate';
		}
		return `${warpSpeed}`;
	}

	function transfer() {
		const f = new CommandedFleet(fleet);
		dispatch('cargo-transfer-dialog', { src: f, dest: f.getCargoTransferTarget($universe) });
	}
</script>

<div class="flex flex-row min-h-[11rem]">
	<div class="flex flex-col">
		<div class="avatar mr-2">
			<div
				class="border-2 border-neutral p-2 bg-black"
				style={`border-color: ${$universe.getPlayerColor(fleet.playerNum)};`}
			>
				{#if fleet.tokens && fleet.tokens.reduce((count, t) => count + t.quantity, 0) > 1}
					<div class="absolute -right-2 -top-1 text-xl w-6 h-6">+</div>
				{/if}

				<div class="fleet-avatar {getIcon(fleet)} bg-black">
					<button
						type="button"
						class="w-full h-full cursor-help"
						on:pointerdown|preventDefault={(e) => onShipDesignTooltip(e, design)}
					/>
				</div>
			</div>
		</div>
		<div class="text-center">{$universe.getPlayerPluralName(fleet.playerNum)}</div>
	</div>
	<div class="flex flex-col grow">
		<div class="flex flex-row">
			<div class="w-32 text-tile-item-title">Ship Count:</div>
			<div>
				{fleet.tokens ? fleet.tokens.reduce((count, t) => count + t.quantity, 0) : 'unknown'}
			</div>
		</div>
		<div class="flex flex-row">
			<div class="w-32 text-tile-item-title">Fleet Mass:</div>
			<div>
				{fleet.spec?.mass ?? fleet.mass ?? 0}kT
			</div>
		</div>
		{#if ownedBy(fleet, $player.num)}
			<div class="flex flex-row">
				<div class="w-32 text-tile-item-title">Fuel:</div>
				<div class="grow">
					<FuelBar value={fleet.fuel ?? 0} capacity={fleet.spec?.fuelCapacity ?? 0} />
				</div>
			</div>
			<div class="flex flex-row">
				<div class="w-32 text-tile-item-title">Cargo:</div>
				<div class="grow">
					<CargoBar
						on:cargo-transfer-dialog={() => transfer()}
						canTransferCargo={canTransferCargo(fleet, $universe)}
						value={fleet.cargo}
						capacity={fleet.spec?.cargoCapacity}
					/>
				</div>
			</div>
		{/if}
		{#if fleet.waypoints && fleet.waypoints.length > 1}
			<div class="flex flex-row">
				<div class="w-32 text-tile-item-title">Next Waypoint:</div>
				<div>{$universe.getTargetName(fleet.waypoints[1])}</div>
			</div>
			{#if fleet.waypoints[1].task != WaypointTask.None}
				<div class="flex flex-row">
					<div class="w-32 text-tile-item-title">Task:</div>
					<div>{startCase(fleet.waypoints[1].task)}</div>
				</div>
			{/if}
			<div class="flex flex-row">
				<div class="w-32 text-tile-item-title">Warp Speed:</div>
				<div>{getWarpSpeed(fleet)}</div>
			</div>
		{:else if fleet.warpSpeed}
			<div class="flex flex-row">
				<div class="w-32 text-tile-item-title">Warp Speed:</div>
				<div>{getWarpSpeed(fleet)}</div>
			</div>
		{/if}

		{#if !ownedBy(fleet, $player.num) && fleet.tokens}
			<div class="text-tile-item-title">
				Fleet Composition:
				<div class="bg-base-100 h-16 overflow-y-auto mt-1 w-full md:w-60 font-normal">
					<ul class="w-full h-full">
						{#each fleet.tokens as token, index}
							<li class="pl-1">
								<button
									type="button"
									class="w-full cursor-help"
									on:pointerdown|preventDefault={(e) =>
										onShipDesignTooltip(e, $universe.getDesign(fleet.playerNum, token.designNum))}
								>
									<div class="flex flex-row justify-between relative">
										{#if (token.damage ?? 0) > 0 && (token.quantityDamaged ?? 0) > 0}
											<div
												style={`width: ${getDamagePercentForToken(
													token,
													$universe.getDesign(fleet.playerNum, token.designNum)
												).toFixed()}%`}
												class="damage-bar h-full absolute opacity-50"
											/>
										{/if}

										<div>
											{$universe.getDesign(fleet.playerNum, token.designNum)?.name}
										</div>
										<div>
											{token.quantity}
										</div>
									</div>
								</button>
							</li>
						{/each}
					</ul>
				</div>
			</div>
		{/if}
	</div>
</div>
