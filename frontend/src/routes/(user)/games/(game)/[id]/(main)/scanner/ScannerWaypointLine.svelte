<script lang="ts">
	import { getGameContext } from '$lib/services/GameContext';
	import type { Fleet, Waypoint } from '$lib/types/Fleet';
	import { StargateWarpSpeed } from '$lib/types/MapObject';
	import { distance } from '$lib/types/Vector';
	import type { LayerCake } from 'layercake';
	import { getContext } from 'svelte';
	import type { Writable } from 'svelte/store';

	const scale = getContext<Writable<number>>('scale');
	const { data, xGet, yGet, xScale, yScale, width, height } = getContext<LayerCake>('LayerCake');

	export let fleet: Fleet;
	export let commanded = false;
	export let selectedWaypoint: Waypoint | undefined;

	type WaypointLineSegment = {
		path: string;
		props: any;
	};

	let segments: WaypointLineSegment[] = [];

	$: {
		segments = [];

		if (fleet.waypoints) {
			const heading = fleet.heading ?? { x: 0, y: 0 };
			for (let i = 1; i < fleet.waypoints.length; i++) {
				const wp0 = fleet.waypoints[i - 1];
				const wp1 = fleet.waypoints[i];
				const distancePerYear = wp1.warpSpeed * wp1.warpSpeed;
				const dist = Math.floor(distance(wp0.position, wp1.position));
				let [x1, y1, x2, y2] = [$xGet(wp0), $yGet(wp0), $xGet(wp1), $yGet(wp1)];
				let dashOffset = 0;

				if (i === 1) {
					// move the first coord along the heading a bit so the line starts after our icon
					x1 += (heading.x * 5) / $scale;
					y1 += (heading.y * 5) / $scale;
				}
				const strokeWidth = selectedWaypoint === wp0 ? 6 / $scale : (commanded ? 5 : 3) / $scale;

				segments.push({
					path: `M${x1},${y1}L${x2},${y2}`,
					props: {
						class: commanded ? 'waypoint-line-commanded' : 'waypoint-line',
						'stroke-width': strokeWidth,
						'stroke-dasharray':
							commanded && wp1.warpSpeed != StargateWarpSpeed && distancePerYear < dist
								? `${$xScale(distancePerYear) - $xScale(5)} ${$xScale(5)}`
								: 0
					}
				});
			}
		}
	}
</script>

{#each segments as segment}
	<path d={segment.path} {...segment.props} />
{/each}
