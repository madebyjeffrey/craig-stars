import { clamp } from '$lib/services/Math';
import type { Cost } from './Cost';
import type { MineFieldType } from './MineField';
import type { Player } from './Player';
import type { PRT } from './Race';
import type { TechLevel } from './TechLevel';
import type { Vector } from './Vector';

export const UnlimitedSpaceDock = -1;
export const NoScanner = -1;
export const NoGate = -1;
export const InfinteGate = 2147483647;
export const OriginMysteryTrader = 'MysteryTrader';
export const GenesisDevice = 'Genesis Device';

export type TechStore = {
	engines: TechEngine[];
	planetaryScanners: TechPlanetaryScanner[];
	terraforms: TechTerraform[];
	defenses: TechDefense[];
	planetaries: TechPlanetary[];
	hullComponents: TechHullComponent[];
	hulls: TechHull[];
};

export type Tech = {
	id?: number;
	createdAt?: string;
	updatedAt?: string;
	techStoreId?: number;

	name: string;
	cost: Cost;
	requirements: TechRequirements;
	ranking?: number;
	category: TechCategory;
	origin?: string;
};

export type TechPlanetary = {
	resetPlanet?: boolean;
} & Tech;

export type TechPlanetaryScanner = {
	scanRange: number;
	scanRangePen?: number;
} & Tech;

export type TechTerraform = {
	ability: number;
	habType: TerraformHabType;
} & Tech;

export type TerraformHabType = (typeof TerraformHabTypes)[keyof typeof TerraformHabTypes];
export const TerraformHabTypes = {
	None: '',
	Gravity: 'Grav',
	Temperature: 'Temp',
	Radiation: 'Rad',
	All: 'All'
} as const;

export type TechDefense = {
	defenseCoverage: number;
} & Tech;

export type TechHullComponent = {
	hullSlotType: HullSlotType;
	mass: number;
	scanner?: boolean;
	scanRange?: number;
	scanRangePen?: number;
	safeHullMass?: number;
	safeRange?: number;
	maxHullMass?: number;
	maxRange?: number;
	packetSpeed?: number;
	miningRate?: number;
	cloakUnits?: number;
	terraformRate?: number;
	killRate?: number;
	minKillRate?: number;
	structureDestroyRate?: number;
	unterraformRate?: number;
	radiating?: boolean;
	smart?: boolean;
	canStealFleetCargo?: boolean;
	canStealPlanetCargo?: boolean;
	armor?: number;
	shield?: number;
	cloakUnarmedOnly?: boolean;
	torpedoBonus?: number;
	initiativeBonus?: number;
	torpedoJamming?: number;
	beamBonus?: number;
	reduceMovement?: number;
	reduceCloaking?: boolean;
	fuelBonus?: number;
	fuelRegenerationRate?: number;
	mineFieldType?: MineFieldType;
	mineLayingRate?: number;
	colonizationModule?: boolean;
	orbitalConstructionModule?: boolean;
	cargoBonus?: number;
	movementBonus?: number;
	beamDefense?: number;
	power?: number;
	range?: number;
	initiative?: number;
	gattling?: boolean;
	hitsAllTargets?: boolean;
	damageShieldsOnly?: boolean;
	accuracy?: number;
	capitalShipMissile?: boolean;
} & Tech;

export type TechHull = {
	armor: number;
	builtInScanner?: boolean;
	cargoCapacity?: number;
	cargoSlotCircle?: boolean;
	cargoSlotPosition?: Vector;
	cargoSlotSize?: Vector;
	fuelCapacity?: number;
	fuelGeneration?: number;
	immuneToOwnDetonation?: boolean;
	initiative?: number;
	innateScanRangePenFactor?: number;
	mass?: number;
	mineLayingBonus?: number;
	orbitalConstructionHull?: boolean;
	rangeBonus?: number;
	repairBonus?: number;
	slots: HullSlot[];
	spaceDock?: number;
	spaceDockSlotCircle?: boolean;
	spaceDockSlotPosition?: Vector;
	spaceDockSlotSize?: Vector;
	starbase?: boolean;
} & Tech;

export type HullSlot = {
	type: HullSlotType;
	capacity: number;
	required?: boolean;
	position: Vector;
};

export enum HullSlotType {
	None = 0,
	Engine = 1 << 1,
	Scanner = 1 << 2,
	Mechanical = 1 << 3,
	Bomb = 1 << 4,
	Mining = 1 << 5,
	Electrical = 1 << 6,
	Shield = 1 << 7,
	Armor = 1 << 8,
	Cargo = 1 << 9,
	SpaceDock = 1 << 10,
	Weapon = 1 << 11,
	Orbital = 1 << 12,
	MineLayer = 1 << 13,
	OrbitalElectrical = HullSlotType.Orbital | HullSlotType.Electrical,
	ShieldElectricalMechanical = HullSlotType.Shield |
		HullSlotType.Electrical |
		HullSlotType.Mechanical,
	ScannerElectricalMechanical = HullSlotType.Scanner |
		HullSlotType.Electrical |
		HullSlotType.Mechanical,
	ArmorScannerElectricalMechanical = HullSlotType.Armor |
		HullSlotType.Scanner |
		HullSlotType.Electrical |
		HullSlotType.Mechanical,
	MineElectricalMechanical = HullSlotType.MineLayer |
		HullSlotType.Electrical |
		HullSlotType.Mechanical,
	ShieldArmor = HullSlotType.Shield | HullSlotType.Armor,
	WeaponShield = HullSlotType.Shield | HullSlotType.Weapon,
	General = HullSlotType.Scanner |
		HullSlotType.Mechanical |
		HullSlotType.Electrical |
		HullSlotType.Shield |
		HullSlotType.Armor |
		HullSlotType.Weapon |
		HullSlotType.MineLayer
}

export type Engine = {
	idealSpeed?: number;
	freeSpeed?: number;
	maxSafeSpeed?: number;
	fuelUsage?: number[];
};
export type TechEngine = Engine & TechHullComponent;

export enum TechCategory {
	Armor = 'Armor',
	BeamWeapon = 'BeamWeapon',
	Bomb = 'Bomb',
	Electrical = 'Electrical',
	Engine = 'Engine',
	Mechanical = 'Mechanical',
	MineLayer = 'MineLayer',
	MineRobot = 'MineRobot',
	Orbital = 'Orbital',
	Planetary = 'Planetary',
	PlanetaryScanner = 'PlanetaryScanner',
	PlanetaryDefense = 'PlanetaryDefense',
	Scanner = 'Scanner',
	Shield = 'Shield',
	ShipHull = 'ShipHull',
	StarbaseHull = 'StarbaseHull',
	Terraforming = 'Terraforming',
	Torpedo = 'Torpedo'
}

export type TechRequirements = {
	lrtsRequired?: number;
	lrtsDenied?: number;
	prtsRequired?: PRT[];
	prtsDenied?: PRT[];
	hullsAllowed?: string[];
	hullsDenied?: string[];
	acquirable?: boolean;
} & TechLevel;

/**
 * Determine if a tech is a hull component
 * @param category The category to check
 * @returns
 */
export const isHullComponent = (category: TechCategory): boolean => {
	switch (category) {
		case TechCategory.Armor:
		case TechCategory.BeamWeapon:
		case TechCategory.Bomb:
		case TechCategory.Electrical:
		case TechCategory.Engine:
		case TechCategory.Mechanical:
		case TechCategory.MineLayer:
		case TechCategory.MineRobot:
		case TechCategory.Orbital:
		case TechCategory.Scanner:
		case TechCategory.Torpedo:
		case TechCategory.Shield:
			return true;
		case TechCategory.Planetary:
		case TechCategory.PlanetaryScanner:
		case TechCategory.PlanetaryDefense:
		case TechCategory.ShipHull:
		case TechCategory.StarbaseHull:
		case TechCategory.Terraforming:
			return false;
		default:
			return false;
	}
};

export function canFillSlot(hcType: HullSlotType, type: HullSlotType): boolean {
	return (hcType & type) > 0;
}

// true if this hull is allowed to mount this component
export function hullAllowed(hull: string, tech: Tech): boolean {
	const hullAllowed = tech.requirements.hullsAllowed
		? tech.requirements.hullsAllowed.indexOf(hull) != -1
		: true;
	const hullDenied = tech.requirements.hullsDenied
		? tech.requirements.hullsDenied.indexOf(hull) != -1
		: false;
	return hullAllowed && !hullDenied;
}

export function getDefenseCoverage(defense: TechDefense, defenses: number): number {
	return 1.0 - Math.pow(1 - defense.defenseCoverage / 100, clamp(defenses, 0, 100));
}

export function getSmartDefenseCoverage(
	defense: TechDefense,
	defenses: number,
	smartDefenseCoverageFactor?: number
): number {
	smartDefenseCoverageFactor ??= 0.5;
	return (
		1.0 -
		Math.pow(
			1 - (defense.defenseCoverage / 100) * smartDefenseCoverageFactor,
			clamp(defenses, 0, 100)
		)
	);
}

export function getCloakPercentForCloakUnits(cloakUnits: number): number {
	if (cloakUnits <= 100) {
		return cloakUnits / 2;
	} else {
		cloakUnits = cloakUnits - 100;
		if (cloakUnits <= 200) {
			return 50 + cloakUnits / 8;
		} else {
			cloakUnits = cloakUnits - 200;
			if (cloakUnits < 312) {
				return 75 + cloakUnits / 24;
			} else {
				cloakUnits = cloakUnits - 312;
				if (cloakUnits <= 512) {
					return 88 + cloakUnits / 64;
				} else if (cloakUnits < 768) {
					return 96;
				} else if (cloakUnits < 1000) {
					return 97;
				} else {
					return 99;
				}
			}
		}
	}
}

export function getBestTerraform(
	techStore: TechStore,
	player: Player,
	habType: TerraformHabType
): TechTerraform | undefined {
	// get the best terraform for a given type, sorted largest to smallest ranking
	return techStore.terraforms
		.filter((t) => player.hasTech(t) && t.habType == habType)
		.sort((a, b) => (b.ranking ?? 0) - (a.ranking ?? 0))[0];
}
