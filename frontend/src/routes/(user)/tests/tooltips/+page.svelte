<script lang="ts">
	import FactoriesTooltip from '$lib/components/game/tooltips/FactoriesTooltip.svelte';
	import MineralTooltip from '$lib/components/game/tooltips/MineralTooltip.svelte';
	import ResourcesTooltip from '$lib/components/game/tooltips/ResourcesTooltip.svelte';
</script>

<h1 class="text-xl">Mineral Tooltip</h1>
<MineralTooltip
	mineralType="Ironium"
	surfaceAmount={1248}
	concentration={114}
	miningRate={52}
	homeworld={false}
/>

<h1 class="text-xl">Factories Tooltip</h1>
<FactoriesTooltip
	planetName="Brin"
	factories={10}
	maxFactories={1000}
	maxPossibleFactories={1200}
	canBuildFactories={true}
/>

<h1 class="text-xl">Factories Tooltip (innate)</h1>
<FactoriesTooltip
	planetName="Brin"
	factories={10}
	maxFactories={1000}
	maxPossibleFactories={1200}
	canBuildFactories={false}
/>

<h1 class="text-xl">Resources Tooltip</h1>
<ResourcesTooltip
	planetName="Brin"
	resourcesPerYear={35}
	resourcesPerYearAvailable={30}
	resourcesPerYearResearch={5}
	resourcesPerYearResearchEstimated={1}
	innateResources={true}
/>
