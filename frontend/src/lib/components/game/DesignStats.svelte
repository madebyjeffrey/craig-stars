<script lang="ts">
	import { Infinite } from '$lib/types/MapObject';
	import { MineFieldTypes } from '$lib/types/MineField';
	import type { Spec } from '$lib/types/ShipDesign';
	import { NoScanner } from '$lib/types/Tech';

	export let spec: Spec;

	function scanRange(range: number | undefined) {
		return !range || range === NoScanner ? '-' : range;
	}
</script>

<div class="flex flex-col min-w-[8rem]">
	<div class="flex justify-between">
		<div class="font-semibold mr-5">Mass</div>
		<div>{spec.mass ?? 0}kT</div>
	</div>
	<div class="flex justify-between">
		<div class="font-semibold mr-5">Max Fuel</div>
		<div>{spec.fuelCapacity ?? 0}mg</div>
	</div>
	{#if spec.fuelGeneration && spec.fuelGeneration > 0}
		<div class="flex justify-between">
			<div class="font-semibold mr-5">Fuel Generation</div>
			<div>{spec.fuelGeneration ?? 0}mg</div>
		</div>
	{/if}
	{#if spec.estimatedRange}
		<div class="flex justify-between">
			{#if spec.estimatedRange == Infinite}
				<div class="font-semibold mr-5">Est Range</div>
				<div>Infinite</div>
			{:else if spec.cargoCapacity}
				<div class="font-semibold mr-5">Est Range (w/cargo)</div>
				<div class="text-right">
					{spec.estimatedRange ?? 0}ly ({spec.estimatedRangeFull ?? 0}ly)
				</div>
			{:else}
				<div class="font-semibold mr-5">Est Range</div>
				<div>{spec.estimatedRange ?? 0}ly</div>
			{/if}
		</div>
	{/if}

	{#if spec.cargoCapacity}
		<div class="flex justify-between">
			<div class="font-semibold mr-5">Cargo Capacity</div>
			<div>{spec.cargoCapacity}kT</div>
		</div>
	{/if}
	<div class="flex justify-between">
		<div class="font-semibold mr-5">Armor</div>
		<div>{spec.armor ?? 0}dp</div>
	</div>
	{#if spec.shields}
		<div class="flex justify-between">
			<div class="font-semibold mr-5">Shields</div>
			<div>{spec.shields}dp</div>
		</div>
	{/if}

	{#if spec.powerRating}
		<div class="flex justify-between">
			<div class="font-semibold mr-5">Rating</div>
			<div>{spec.powerRating}</div>
		</div>
	{/if}
	{#if spec.cloakPercent}
		{#if spec.cargoCapacity}
			<div class="flex justify-between">
				<div class="font-semibold mr-5">Cloak (with cargo)</div>
				<div class="">{spec.cloakPercent ?? 0}% ({spec.cloakPercentFullCargo ?? 0}%)</div>
			</div>
		{:else}
			<div class="flex justify-between">
				<div class="font-semibold mr-5">Cloak</div>
				<div>{spec.cloakPercent ?? 0}%</div>
			</div>
		{/if}
	{/if}
	{#if spec.torpedoBonus || spec.torpedoJamming}
		<div class="flex justify-between">
			<div class="font-semibold mr-5">Torpedo Bonus/Jamming</div>
			<div>
				{((spec.torpedoBonus ?? 0) * 100).toFixed()}%/{(
					(spec.torpedoJamming ?? 0) * 100
				).toFixed()}%
			</div>
		</div>
	{/if}
	{#if (spec.beamBonus && spec.beamBonus !== 1) || (spec.beamDefense && spec.beamDefense !== 1)}
		<div class="flex justify-between">
			<div class="font-semibold mr-5">Beam Bonus/Defense</div>
			<div>
				x{(spec.beamBonus ?? 1).toFixed(1)}/x{(spec.beamDefense ?? 1).toFixed(1)}
			</div>
		</div>
	{/if}
	<div class="flex justify-between">
		<div class="font-semibold mr-5">Initiative/Moves</div>
		<div>{spec.initiative ?? 0}/{spec.movement ?? 0}</div>
	</div>
	{#if spec.scanRange != NoScanner || spec.scanRangePen != NoScanner}
		<div class="flex justify-between">
			<div class="font-semibold mr-5">Scanner Range</div>
			<div>{scanRange(spec.scanRange)}/{scanRange(spec.scanRangePen)}</div>
		</div>
	{/if}

	{#if spec.mineLayingRateByMineType && spec.mineLayingRateByMineType[MineFieldTypes.Standard]}
		<div class="flex justify-between">
			<div class="font-semibold mr-5">Mine Laying</div>
			<div>{spec.mineLayingRateByMineType[MineFieldTypes.Standard]} std/yr</div>
		</div>
	{/if}
	{#if spec.mineLayingRateByMineType && spec.mineLayingRateByMineType[MineFieldTypes.Heavy]}
		<div class="flex justify-between">
			<div class="font-semibold mr-5">Mine Laying</div>
			<div>{spec.mineLayingRateByMineType[MineFieldTypes.Heavy]} hvy/yr</div>
		</div>
	{/if}
	{#if spec.mineLayingRateByMineType && spec.mineLayingRateByMineType[MineFieldTypes.SpeedBump]}
		<div class="flex justify-between">
			<div class="font-semibold mr-5">Mine Laying</div>
			<div>{spec.mineLayingRateByMineType[MineFieldTypes.SpeedBump]} spd/yr</div>
		</div>
	{/if}
	{#if spec.miningRate}
		<div class="flex justify-between">
			<div class="font-semibold mr-5">Remote Mining</div>
			<div>{spec.miningRate}kT/yr</div>
		</div>
	{/if}
</div>
