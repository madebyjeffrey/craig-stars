<script lang="ts">
	import { clickOutside } from '$lib/clickOutside';
	import { onScannerContextPopup } from '$lib/components/game/tooltips/ScannerContextPopup.svelte';
	import { getGameContext } from '$lib/services/GameContext';
	import { clamp } from '$lib/services/Math';
	import { filterFleet } from '$lib/types/Filter';
	import { type Fleet } from '$lib/types/Fleet';
	import {
		MapObjectType,
		None,
		equal as mapObjectEqual,
		type MapObject
	} from '$lib/types/MapObject';
	import { emptyVector, equal, type Vector } from '$lib/types/Vector';
	import type { ScaleLinear } from 'd3-scale';
	import { scaleLinear } from 'd3-scale';
	import { select } from 'd3-selection';
	import { ZoomTransform, zoom, type D3ZoomEvent, type ZoomBehavior } from 'd3-zoom';
	import hotkeys from 'hotkeys-js';
	import { Html, LayerCake, Svg } from 'layercake';
	import { createEventDispatcher, onDestroy, onMount, setContext } from 'svelte';
	import { derived, writable } from 'svelte/store';
	import MapObjectQuadTreeFinder, {
		type FinderEventDetails
	} from './MapObjectQuadTreeFinder.svelte';
	import ScannerFleets from './ScannerFleets.svelte';
	import ScannerMapObjectLocation from './ScannerMapObjectLocation.svelte';
	import ScannerMineFieldPattern from './ScannerMineFieldPattern.svelte';
	import ScannerMineFields from './ScannerMineFields.svelte';
	import ScannerMineralPackets from './ScannerMineralPackets.svelte';
	import ScannerMysteryTraders from './ScannerMysteryTraders.svelte';
	import ScannerNames from './ScannerNames.svelte';
	import ScannerPacketDests from './ScannerPacketDests.svelte';
	import ScannerPlanets from './ScannerPlanets.svelte';
	import ScannerSalvages from './ScannerSalvages.svelte';
	import ScannerScanners from './ScannerScanners.svelte';
	import ScannerWarpLine from './ScannerWarpLine.svelte';
	import ScannerWaypoints from './ScannerWaypoints.svelte';
	import ScannerWormholeLinks from './ScannerWormholeLinks.svelte';
	import ScannerWormholes from './ScannerWormholes.svelte';
	import SelectedMapObject from './SelectedMapObject.svelte';
	import type { DeleteWaypointEvent } from '../command/FleetWaypointsTile.svelte';

	const dispatch = createEventDispatcher<DeleteWaypointEvent>();

	const {
		game,
		player,
		universe,
		settings,
		commandMapObject,
		commandedFleet,
		commandedMapObject,
		commandedPlanet,
		currentSelectedWaypointIndex,
		highlightMapObject,
		mostRecentMapObject,
		selectMapObject,
		selectWaypoint,
		selectedMapObject,
		selectedWaypoint,
		zoomTarget,
		updatePlanetOrders,
		updateFleetOrders
	} = getGameContext();

	const xGetter = (mo: MapObject) => mo?.position?.x;
	const yGetter = (mo: MapObject) => mo?.position?.y;

	let clientWidth = 100;
	let clientHeight = 100;
	let aspectRatio = 1;
	let transform: ZoomTransform;
	let zoomBehavior: ZoomBehavior<HTMLElement, any>;
	let root: HTMLElement;
	let padding = 20; // 20 px, used in zooming
	let scaleX: ScaleLinear<number, number, never>;
	let scaleY: ScaleLinear<number, number, never>;
	let zoomEnabled = true;
	let zooming = false;
	let showLocator = false;
	let shouldAddWaypoint = false;
	let fastestWaypoint = false;

	// our map scales for .75 to 10x, but the icons for the planets and fleets are 2x min
	const minZoom = 0.75;
	const maxZoom = 10;
	const minObjectZoom = 2;
	const scale = writable(3); // default 3x zoom
	const objectScale = derived([scale], ([s]) => clamp(s, minObjectZoom, maxZoom));
	setContext('scale', scale);
	setContext('objectScale', objectScale);

	// $: console.log('scale ', $scale);

	const unsubscribe = zoomTarget.subscribe(() => showTargetLocation());

	onMount(() => {
		hotkeys('v', 'root', showTargetLocation);
	});

	onDestroy(() => {
		hotkeys.unbind('v', 'root', showTargetLocation);
		unsubscribe();
	});

	// handle zoom in/out
	// this behavior controls how the zoom behaves
	// below we handle zooming events by updating a transform
	$: {
		if (root) {
			handleResize();

			zoomBehavior = zoom<HTMLElement, any>()
				.extent([
					[0, 0],
					[clientWidth, clientHeight]
				])
				.scaleExtent([minZoom, maxZoom])
				.translateExtent([
					[-20, -20],
					[clientWidth + padding, clientHeight + padding]
				])
				.on('zoom', handleZoom)
				.on('start', handleZoomStart)
				.on('end', handleZoomEnd);

			enableDragAndZoom();
		}
	}

	$: {
		if ($settings.addWaypoint && zoomEnabled) {
			disableDragAndZoom();
		} else if (!$settings.addWaypoint && !zoomEnabled) {
			enableDragAndZoom();
		}
	}

	// enable drag and zoom, but disable dblclick zoom events
	function enableDragAndZoom() {
		select(root).call(zoomBehavior).on('dblclick.zoom', null);
		dragAndZoomEnabled = true;
	}

	// disable drag and zoom temporarily
	function disableDragAndZoom() {
		select(root).on('.zoom', null);
		dragAndZoomEnabled = false;
		zooming = false;
	}

	const xRange = () => {
		if (aspectRatio > 1 && clientHeight > clientWidth) {
			// tall skinny viewport, wide map, so fully expand on the x
			// but shrink up height
			return [0, clientWidth];
		} else if (aspectRatio > 1 && clientWidth > clientHeight) {
			// wide viewport, wide map, so fully expand on the y
			// but shrink up width
			return [0, clientHeight * aspectRatio];
		}
		return [0, Math.min(clientWidth, clientHeight)];
	};
	const yRange = () => {
		if (aspectRatio > 1 && clientHeight > clientWidth) {
			// tall skinny viewport, wide map, so fully expand on the x
			// but shrink up height
			return [0, clientWidth / aspectRatio];
		} else if (aspectRatio > 1 && clientWidth > clientHeight) {
			// wide viewport, wide map, so fully expand on the y
			return [0, clientHeight];
		}
		return [0, Math.min(clientWidth, clientHeight)];
	};

	function handleResize() {
		clientWidth = root?.clientWidth ?? 100;
		clientHeight = root?.clientHeight ?? 100;
		aspectRatio = $game.area.x / $game.area.y;

		// compute scales
		scaleX = scaleLinear().range(xRange()).domain([0, $game.area.x]);
		scaleY = scaleLinear().range(yRange()).domain([0, $game.area.y]);
	}

	function handleKeyDown(e: KeyboardEvent) {
		// add a waypoint if we are currently commanding a fleet and we didn't just click
		// on the fleet
		shouldAddWaypoint = !!$commandedFleet && (e.shiftKey || e.metaKey);
		fastestWaypoint = !!$commandedFleet && e.metaKey;

		switch (e.key) {
			case '+':
			case '=':
				zoomViewport(clamp($scale + 1, minZoom, maxZoom));
				break;
			case '-':
			case '_':
				zoomViewport(clamp($scale - 1, minZoom, maxZoom));
				break;
		}
	}

	function handleKeyUp(e: KeyboardEvent) {
		// add a waypoint if we are currently commanding a fleet and we didn't just click
		// on the fleet
		shouldAddWaypoint = !!$commandedFleet && (e.shiftKey || e.metaKey);
		fastestWaypoint = !!$commandedFleet && e.metaKey;
	}

	function showTargetLocation() {
		showLocator = true;
		setTimeout(() => (showLocator = false), 500);
	}

	function handleZoom(e: D3ZoomEvent<HTMLElement, any>) {
		transform = e.transform;
		$scale = transform.k;
		// console.log('handleZoom', e, transform);
	}

	function handleZoomStart(e: D3ZoomEvent<HTMLElement, any>) {
		zooming = true;
	}

	function handleZoomEnd(e: D3ZoomEvent<HTMLElement, any>) {
		zooming = false;
	}

	// zoom to the commanded map object every time it changes
	$: if (root && $zoomTarget) {
		translateViewport($zoomTarget.position);
	}

	// zoom the display to a point on the map
	function translateViewport(position: Vector, scaleTo?: number) {
		if (root) {
			select(root).call(zoomBehavior.scaleTo, $scale);
			const scaled: Vector = {
				x: scaleX(position.x),
				y: scaleY(position.y)
			};
			let localScale = $scale;
			if (scaleTo) {
				localScale = scaleTo;
			}
			select(root)
				.call(zoomBehavior.translateTo, scaled.x, scaled.y)
				.call(zoomBehavior.scaleTo, localScale);
		}
	}

	// zoom the viewport to a specific scale
	function zoomViewport(scaleTo: number) {
		if (root) {
			select(root).call(zoomBehavior.scaleTo, scaleTo);
		}
	}

	let pointerDown = false;
	let draggingWaypoint = false;
	let waypointHighlighted = false;
	let dragAndZoomEnabled = true;

	// set to true if we are moving a waypoint to a position rather than a target
	// this is enabled when the shift key is held
	let positionWaypoint = false;

	// if we just added a waypoint, don't drag it around
	let waypointJustAdded = false;

	// turn off dragging
	function onContextMenu(e: CustomEvent<FinderEventDetails>) {
		const { event, found } = e.detail;

		if (found && event instanceof MouseEvent) {
			onScannerContextPopup(event, found.position);
		}
	}

	// as the pointer moves, find the items it is under
	function onPointerMove(e: CustomEvent<FinderEventDetails>) {
		const { event, found, position } = e.detail;

		highlightMapObject(found);

		if (draggingWaypoint && !zooming) {
			positionWaypoint = event.shiftKey;
			dragWaypointMove(position, found);
		}

		// check if we are over the commanded fleet's waypoint
		const fleetWaypoint =
			found &&
			$commandedFleet &&
			$commandedFleet.waypoints.slice(1).find((wp) => equal(wp.position, found.position));
		waypointHighlighted = !!fleetWaypoint;
		if (waypointHighlighted) {
			if (dragAndZoomEnabled) {
				disableDragAndZoom();
			}
		} else {
			if (!draggingWaypoint && !dragAndZoomEnabled) {
				enableDragAndZoom();
			}
		}

		// check if we started a waypoint drag
		// we only
		// * start dragging once
		// * if the pointer is down
		// * if we are over a mapobject
		// * if we have a commanded fleet
		if (!waypointJustAdded && !draggingWaypoint && pointerDown && fleetWaypoint) {
			draggingWaypoint = true;
			selectWaypoint(fleetWaypoint);
		}
	}

	async function onPointerDown(e: CustomEvent<FinderEventDetails>) {
		const { event, found, position } = e.detail;

		if (event instanceof MouseEvent && event.button != 0) {
			// we only care about the first button
			return;
		}

		if (found?.type == MapObjectType.Fleet && !filterFleet($player, found as Fleet, $settings)) {
			// this object we clicked is filtered out, don't do anything
			return;
		}

		pointerDown = true;

		if (found) {
			if ((shouldAddWaypoint || $settings.addWaypoint) && (await addWaypoint(found, position))) {
			} else {
				mapObjectSelected(found);
			}
		} else {
			if (shouldAddWaypoint || $settings.addWaypoint) {
				addWaypoint(found, position);
			}
		}
	}

	// turn off dragging
	function onPointerUp(e: CustomEvent<FinderEventDetails>) {
		const { event, found, position } = e.detail;

		if (event instanceof MouseEvent && event.button != 0) {
			// we only care about the first button
			return;
		}

		if (draggingWaypoint) {
			if (!dragAndZoomEnabled) {
				enableDragAndZoom();
			}

			dragWaypointDone(position, found);
		}
		draggingWaypoint = false;
		pointerDown = false;
		waypointJustAdded = false;
	}

	// move the selected waypoint around snapping to targets
	function dragWaypointMove(position: Vector, mo: MapObject | undefined) {
		if ($selectedWaypoint && $currentSelectedWaypointIndex && $commandedFleet) {
			// don't move the waypoint to any adjacent waypoints
			if (mo && !positionWaypoint) {
				const index = $commandedFleet.waypoints.findIndex((wp) => equal(wp.position, mo.position));
				if (
					index == $currentSelectedWaypointIndex - 1 ||
					index == $currentSelectedWaypointIndex + 1
				) {
					return;
				}
			}

			const dest = mo ? { mo: mo } : { position: position ?? emptyVector };

			// get highest mass of the fleet ships (for stargates)
			const highestShipMass = Math.max(
				...$commandedFleet.tokens.map((t) => $universe.getMyDesign(t.designNum)?.spec.mass ?? 0)
			);

			if (
				$commandedFleet.updateWaypoint(
					$player,
					$universe,
					dest,
					$currentSelectedWaypointIndex,
					highestShipMass,
					$settings.fastestWaypoint || fastestWaypoint
				)
			) {
				// trigger reaction
				$selectedWaypoint = $selectedWaypoint;
			}
		}
	}

	async function dragWaypointDone(position: Vector, mo: MapObject | undefined) {
		// reset waypoint dragging
		if ($selectedWaypoint && $commandedFleet && draggingWaypoint) {
			const dest = mo ? { mo: mo } : { position: position ?? emptyVector };

			// get highest mass of the fleet ships (for stargates)
			const highestShipMass = Math.max(
				...$commandedFleet.tokens.map((t) => $universe.getMyDesign(t.designNum)?.spec.mass ?? 0)
			);

			if (
				$commandedFleet.updateWaypoint(
					$player,
					$universe,
					dest,
					$currentSelectedWaypointIndex,
					highestShipMass,
					$settings.fastestWaypoint || fastestWaypoint
				)
			) {
				await updateFleetOrders($commandedFleet);

				// select the new waypoint
				selectWaypoint($commandedFleet.waypoints[$currentSelectedWaypointIndex]);
				if ($selectedWaypoint && $selectedWaypoint.targetType && $selectedWaypoint.targetNum) {
					const mo = $universe.getMapObject($selectedWaypoint);

					if (mo) {
						selectMapObject(mo);
					}
				}
			} else {
				// we dragged a waypoint to the previous position, delete it
				dispatch('delete-waypoint');
			}
		}
	}

	// disable add waypoint mode when the user clicks outside the scanner
	function disableAddWaypointMode(event: MouseEvent) {
		// ignore clicks on the add-waypoint toolbar button
		const elem = event.target as Element;
		if (elem?.id == 'add-waypoint' || elem?.parentElement?.id == 'add-waypoint') {
			return;
		}
		if ($settings.addWaypoint) {
			$settings.addWaypoint = false;
			$settings.fastestWaypoint = false;
		}
	}

	// if the shift key is held, add a waypoint instead of selecting a mapobject
	async function addWaypoint(mo: MapObject | undefined, position: Vector): Promise<boolean> {
		if (zooming) {
			return false;
		}
		if (!$commandedFleet?.waypoints) {
			return false;
		}

		const dest = mo ? { mo: mo } : { position: position ?? emptyVector };

		// get highest mass of the fleet ships (for stargates)
		const highestShipMass = Math.max(
			...$commandedFleet.tokens.map((t) => $universe.getMyDesign(t.designNum)?.spec.mass ?? 0)
		);

		const newlyAddedWaypointIndex = $commandedFleet.addWaypoint(
			$player,
			$universe,
			dest,
			$currentSelectedWaypointIndex,
			highestShipMass,
			$settings.fastestWaypoint || fastestWaypoint
		);

		if (!newlyAddedWaypointIndex) {
			return false;
		}

		waypointJustAdded = true;

		await updateFleetOrders($commandedFleet);

		// select the new waypoint
		selectWaypoint($commandedFleet.waypoints[newlyAddedWaypointIndex]);
		if ($selectedWaypoint && $selectedWaypoint.targetType && $selectedWaypoint.targetNum) {
			const mo = $universe.getMapObject($selectedWaypoint);

			if (mo) {
				selectMapObject(mo);
			}
		}

		return true;
	}
	/**
	 * When a mapobject is selected we go through a few steps.
	 * - We select it if it's a new selection
	 * - We cycle through our commandable objects at the same location if we own an object there
	 * @param mo
	 */
	function mapObjectSelected(mo: MapObject) {
		if ($settings.setPacketDest) {
			if (mo.type != MapObjectType.Planet) {
				return;
			} else {
				$settings.setPacketDest = false;
				// something went wrong, can't set dest on a planet without a massdriver
				if (!$commandedPlanet?.spec.hasMassDriver) {
					return;
				}

				if (mapObjectEqual(mo, $commandedPlanet)) {
					// clear dest
					$commandedPlanet.packetTargetNum = None;
				} else {
					$commandedPlanet.packetTargetNum = mo.num;
				}

				updatePlanetOrders($commandedPlanet);
				return;
			}
		}

		if ($selectedMapObject !== mo) {
			// we selected a different object, so just select it
			selectMapObject(mo);

			// if we selected a mapobject that is a waypoint, select the waypoint as well
			if ($commandedFleet?.waypoints) {
				const fleetWaypoint = $commandedFleet.waypoints.find((wp) =>
					equal(wp.position, mo.position)
				);
				if (fleetWaypoint) {
					selectWaypoint(fleetWaypoint);
				}
			}
		} else {
			// we selected the same mapobject twice
			const myMapObjectsAtPosition = $universe.getMyMapObjectsByPosition(mo);
			if (myMapObjectsAtPosition?.length > 0) {
				let index = myMapObjectsAtPosition.findIndex((mo) =>
					mapObjectEqual(mo, $commandedMapObject)
				);
				// if our currently commanded map object is not at this location, reset the index
				if (index == -1) {
					index = 0;
				} else {
					// command the next one
					index = index >= myMapObjectsAtPosition.length - 1 ? 0 : index + 1;
				}
				const nextMapObject = myMapObjectsAtPosition[index];

				commandMapObject(nextMapObject);
			}
		}
	}

	let data: MapObject[] = [];
	$: {
		const waypoints: MapObject[] = [];
		if ($commandedFleet?.waypoints) {
			waypoints.push(
				...$commandedFleet.waypoints.map((wp) => {
					const mo = $universe.getMapObject(wp);
					if (mo) {
						return mo;
					} else {
						return {
							position: wp.position,
							type: wp.targetType ?? MapObjectType.PositionWaypoint,
							name: wp.targetName ?? '',
							num: wp.targetNum ?? 0,
							playerNum: wp.targetPlayerNum ?? 0
						} as MapObject;
					}
				})
			);
		}
		data = [
			...waypoints,
			...$universe.fleets.filter(
				(f) => f.orbitingPlanetNum === None || f.orbitingPlanetNum === undefined
			),
			...$universe.mysteryTraders,
			...$universe.mineralPackets,
			...$universe.salvages,
			...$universe.wormholes,
			...$universe.mineFields,
			...$universe.planets
		];
	}
</script>

<svelte:window on:resize={handleResize} on:keydown={handleKeyDown} on:keyup={handleKeyUp} />

<div
	class:cursor-grab={waypointHighlighted}
	class:cursor-cell={shouldAddWaypoint ||
		(!!$commandedFleet && $settings.addWaypoint) ||
		$settings.setPacketDest}
	class={`grow bg-black overflow-hidden p-[${padding}px] select-none`}
	use:clickOutside={disableAddWaypointMode}
>
	<LayerCake
		{data}
		x={xGetter}
		y={yGetter}
		xDomain={[0, $game.area.x]}
		yDomain={[0, $game.area.y]}
		{xRange}
		{yRange}
		yReverse={true}
		bind:element={root}
	>
		<!-- <Svg viewBox={`0 0 ${game.area.x} ${game.area.y}`}> -->
		<Svg>
			<g transform={transform?.toString()}>
				<ScannerScanners />
				<ScannerMineFieldPattern />
				<ScannerMineFields />
				<ScannerPacketDests />
				<ScannerWaypoints />
				<ScannerPlanets />
				<ScannerMineralPackets />
				<ScannerWormholes />
				<ScannerFleets />
				<ScannerMysteryTraders />
				<ScannerWarpLine />
				<ScannerWormholeLinks />
				<ScannerSalvages />
				<SelectedMapObject />
				{#if showLocator}
					<ScannerMapObjectLocation show={$mostRecentMapObject} />
				{/if}
			</g>
		</Svg>
		<Html>
			{#if transform}
				<ScannerNames {transform} />

				<MapObjectQuadTreeFinder
					on:contextmenu={onContextMenu}
					on:pointermove={onPointerMove}
					on:pointerdown={onPointerDown}
					on:pointerup={onPointerUp}
					on:touchmove={onPointerMove}
					searchRadius={20}
					{transform}
				/>
			{/if}
		</Html>
	</LayerCake>
</div>
