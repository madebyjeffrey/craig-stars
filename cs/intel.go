package cs

import (
	"fmt"
	"slices"

	"github.com/rs/zerolog"
)

const ReportAgeUnexplored = -1
const Unowned = 0

// Each player has intel about planet, fleets, and other map objects. The discoverer is used to update the player
// intel with knowledge about the universe.
type discoverer interface {
	discoverPlayer(player *Player)
	discoverPlayerScores(player *Player)
	discoverPlanet(rules *Rules, planet *Planet, penScanned bool) error
	clearPlanetOwnerIntel(planet *Planet) error
	discoverPlanetStarbase(planet *Planet) error
	discoverPlanetCargo(planet *Planet) error
	discoverPlanetScanner(planet *Planet) error
	discoverPlanetTerraformability(planetNum int) error
	discoverFleet(fleet *Fleet, discoverName bool)
	discoverFleetCargo(fleet *Fleet)
	discoverFleetScanner(fleet *Fleet)
	discoverMineField(mineField *MineField)
	discoverMineralPacket(rules *Rules, mineralPacket *MineralPacket, packetPlayer *Player, target *Planet)
	discoverMineralPacketScanner(mineralPacket *MineralPacket)
	discoverMineralPacketCargo(mineralPacket *MineralPacket)
	discoverSalvage(salvage *Salvage)
	discoverWormhole(wormhole *Wormhole)
	discoverWormholeLink(wormhole1, wormhole2 *Wormhole)
	forgetWormhole(num int)
	discoverMysteryTrader(mineField *MysteryTrader)
	discoverDesign(design *ShipDesign, discoverSlots bool)
}

type discover struct {
	log    zerolog.Logger
	player *Player
}

// A discoverer of discoverers. This implements the discover interface and
// is used to support discovering for a player and all of their allies when scanning, invading, etc
type discovererWithAllies struct {
	playerDiscoverer discover
	allyDiscoverers  []discover
}

func newDiscoverer(log zerolog.Logger, player *Player) discoverer {
	discoverLogger := log.With().Int("Player", player.Num).Logger()
	return &discover{discoverLogger, player}
}

func newDiscovererWithAllies(log zerolog.Logger, player *Player, players []*Player) discoverer {
	// find any players we share maps with
	mapSharePlayers := make([]discover, 0, len(players))
	for i, relation := range player.Relations {
		if i == player.Num-1 {
			continue
		}
		if relation.Relation == PlayerRelationFriend && relation.ShareMap {
			discoverLogger := log.With().Int("Player", players[i].Num).Logger()
			mapSharePlayers = append(mapSharePlayers, discover{discoverLogger, players[i]})
		}
	}

	return &discovererWithAllies{
		playerDiscoverer: discover{log, player},
		allyDiscoverers:  mapSharePlayers,
	}
}

type Intel struct {
	Name      string `json:"name"`
	Num       int    `json:"num"`
	PlayerNum int    `json:"playerNum"`
	ReportAge int    `json:"reportAge"`
}

type MapObjectIntel struct {
	Intel
	Type     MapObjectType `json:"type"`
	Position Vector        `json:"position"`
}

func (intel *Intel) String() string {
	return fmt.Sprintf("Num: %3d %s", intel.Num, intel.Name)
}

func (intel *Intel) Owned() bool {
	return intel.PlayerNum != Unowned
}

type PlanetIntel struct {
	MapObjectIntel
	Hab                           Hab         `json:"hab,omitempty"`
	BaseHab                       Hab         `json:"baseHab,omitempty"`
	MineralConcentration          Mineral     `json:"mineralConcentration,omitempty"`
	Starbase                      *FleetIntel `json:"starbase,omitempty"`
	Cargo                         Cargo       `json:"cargo,omitempty"`
	CargoDiscovered               bool        `json:"cargoDiscovered,omitempty"`
	PlanetHabitability            int         `json:"planetHabitability,omitempty"`
	PlanetHabitabilityTerraformed int         `json:"planetHabitabilityTerraformed,omitempty"`
	Homeworld                     bool        `json:"homeworld,omitempty"`
	Spec                          PlanetSpec  `json:"spec,omitempty"`
}

type ShipDesignIntel struct {
	Intel
	Hull          string           `json:"hull,omitempty"`
	HullSetNumber int              `json:"hullSetNumber,omitempty"`
	Version       int              `json:"version,omitempty"`
	Slots         []ShipDesignSlot `json:"slots,omitempty"`
	Spec          ShipDesignSpec   `json:"spec,omitempty"`
}

type FleetIntel struct {
	MapObjectIntel
	BaseName          string      `json:"baseName,omitempty"`
	Heading           Vector      `json:"heading,omitempty"`
	OrbitingPlanetNum int         `json:"orbitingPlanetNum,omitempty"`
	WarpSpeed         int         `json:"warpSpeed,omitempty"`
	Mass              int         `json:"mass,omitempty"`
	Cargo             Cargo       `json:"cargo,omitempty"`
	CargoDiscovered   bool        `json:"cargoDiscovered,omitempty"`
	Freighter         bool        `json:"freighter,omitempty"`
	ScanRange         int         `json:"scanRange,omitempty"`
	ScanRangePen      int         `json:"scanRangePen,omitempty"`
	Tokens            []ShipToken `json:"tokens,omitempty"`
}

type MineralPacketIntel struct {
	MapObjectIntel
	WarpSpeed       int    `json:"warpSpeed"`
	Heading         Vector `json:"heading"`
	Cargo           Cargo  `json:"cargo,omitempty"`
	TargetPlanetNum int    `json:"targetPlanetNum,omitempty"`
	ScanRange       int    `json:"scanRange,omitempty"`
	ScanRangePen    int    `json:"scanRangePen,omitempty"`
}

type SalvageIntel struct {
	MapObjectIntel
	Cargo Cargo `json:"cargo,omitempty"`
}

type MineFieldIntel struct {
	MapObjectIntel
	NumMines      int           `json:"numMines"`
	MineFieldType MineFieldType `json:"mineFieldType"`
	Spec          MineFieldSpec `json:"spec"`
}

type WormholeIntel struct {
	MapObjectIntel
	DestinationNum int               `json:"destinationNum,omitempty"`
	Stability      WormholeStability `json:"stability,omitempty"`
}

type MysteryTraderIntel struct {
	MapObjectIntel
	WarpSpeed     int    `json:"warpSpeed,omitempty"`
	Heading       Vector `json:"heading"`
	RequestedBoon int    `json:"requestedBoon"`
}

type PlayerIntel struct {
	Name           string `json:"name,omitempty"`
	Num            int    `json:"num,omitempty"`
	Color          string `json:"color,omitempty"`
	Seen           bool   `json:"seen,omitempty"`
	RaceName       string `json:"raceName,omitempty"`
	RacePluralName string `json:"racePluralName,omitempty"`
}

type ScoreIntel struct {
	ScoreHistory []PlayerScore `json:"scoreHistory"`
}

func (p *PlanetIntel) String() string {
	return fmt.Sprintf("Planet %s", &p.MapObjectIntel)
}

func (f *FleetIntel) String() string {
	return fmt.Sprintf("Player: %d, Fleet: %s", f.PlayerNum, f.Name)
}

func (f *SalvageIntel) String() string {
	return fmt.Sprintf("Player: %d, Salvage: %s", f.PlayerNum, f.Name)
}

func (f *MineFieldIntel) String() string {
	return fmt.Sprintf("Player: %d, MineField: %s", f.PlayerNum, f.Name)
}

func (f *MineralPacketIntel) String() string {
	return fmt.Sprintf("Player: %d, MineralPacket: %s", f.PlayerNum, f.Name)
}

func (d *ShipDesignIntel) String() string {
	return fmt.Sprintf("Player: %d, Fleet: %s", d.PlayerNum, d.Name)
}

// create a new FleetIntel object by key
func newFleetIntel(playerNum int, num int) *FleetIntel {
	return &FleetIntel{
		MapObjectIntel: MapObjectIntel{
			Type: MapObjectTypeFleet,
			Intel: Intel{
				PlayerNum: playerNum,
				Num:       num,
			},
		},
	}
}

// create a new WormholeIntel object by key
func newWormholeIntel(num int) *WormholeIntel {
	return &WormholeIntel{
		MapObjectIntel: MapObjectIntel{
			Type: MapObjectTypeWormhole,
			Intel: Intel{
				Num: num,
			},
		},
	}
}

// create a new SalvageIntel object by key
func newSalvageIntel(playerNum int, num int) *SalvageIntel {
	return &SalvageIntel{
		MapObjectIntel: MapObjectIntel{
			Type: MapObjectTypeSalvage,
			Intel: Intel{
				PlayerNum: playerNum,
				Num:       num,
			},
		},
	}
}

// create a new MineFieldIntel object by key
func newMineFieldIntel(playerNum int, num int) *MineFieldIntel {
	return &MineFieldIntel{
		MapObjectIntel: MapObjectIntel{
			Type: MapObjectTypeMineField,
			Intel: Intel{
				PlayerNum: playerNum,
				Num:       num,
			},
		},
	}
}

// create a new MineralPacketIntel object by key
func newMineralPacketIntel(playerNum int, num int) *MineralPacketIntel {
	return &MineralPacketIntel{
		MapObjectIntel: MapObjectIntel{
			Type: MapObjectTypeMineralPacket,
			Intel: Intel{
				PlayerNum: playerNum,
				Num:       num,
			},
		},
	}
}

// create a new MysteryTraderIntel object by key
func newMysteryTraderIntel(num int) *MysteryTraderIntel {
	return &MysteryTraderIntel{
		MapObjectIntel: MapObjectIntel{
			Type: MapObjectTypeMysteryTrader,
			Intel: Intel{
				Num: num,
			},
		},
	}
}

// true if we haven't explored this planet
func (intel *PlanetIntel) Unexplored() bool {
	return intel.ReportAge == ReportAgeUnexplored
}

// true if we have explored this planet
func (intel *PlanetIntel) Explored() bool {
	return intel.ReportAge != ReportAgeUnexplored
}

// discover a planet and add it to the player's intel
func (d *discover) discoverPlanet(rules *Rules, planet *Planet, penScanned bool) error {

	player := d.player
	var intel *PlanetIntel
	planetIndex := planet.Num - 1

	if planetIndex < 0 || planetIndex >= len(player.PlanetIntels) {
		return fmt.Errorf("player %s cannot discover planet %s, planetIndex %d out of range", player, planet, planetIndex)
	}

	intel = &player.PlanetIntels[planetIndex]

	// everyone knows these about planets
	intel.Position = planet.Position
	intel.Name = planet.Name
	intel.Num = planet.Num

	// scanning a planet tells you who owns it, and whether it has a starbase
	// but you don't get hab/pop unlesss you own it
	intel.PlayerNum = planet.PlayerNum
	intel.Spec.HasStarbase = planet.Spec.HasStarbase
	intel.Spec.HasStargate = planet.Spec.HasStargate
	intel.Spec.HasMassDriver = planet.Spec.HasMassDriver
	intel.Spec.DockCapacity = planet.Spec.DockCapacity

	ownedByPlayer := planet.PlayerNum != Unowned && player.Num == planet.PlayerNum

	if penScanned || ownedByPlayer {
		if !ownedByPlayer && intel.ReportAge == ReportAgeUnexplored {
			// let the player know we discovered a new planet
			messager.planetDiscovered(player, planet)
			d.log.Debug().
				Int("Planet", planet.Num).
				Msgf("player discovered planet")
		}

		// if we pen scanned the planet, we learn some things
		intel.ReportAge = 0
		intel.Hab = planet.Hab
		intel.BaseHab = planet.BaseHab
		intel.MineralConcentration = planet.MineralConcentration
		intel.Homeworld = planet.Homeworld
		intel.Spec.Habitability = player.Race.GetPlanetHabitability(intel.Hab)

		// terraforming
		terraformer := NewTerraformer()
		intel.Spec.TerraformAmount = terraformer.getTerraformAmount(intel.Hab, intel.BaseHab, player, player)
		intel.Spec.MinTerraformAmount = terraformer.getMinTerraformAmount(intel.Hab, intel.BaseHab, player, player)
		intel.Spec.CanTerraform = intel.Spec.TerraformAmount.absSum() > 0
		intel.Spec.TerraformedHabitability = player.Race.GetPlanetHabitability(planet.Hab.Add(intel.Spec.TerraformAmount))
		intel.Spec.MaxPopulation = planet.getMaxPopulation(rules, player, intel.Spec.Habitability)

		// discover starbases on scan, but don't discover designs
		intel.Spec.HasStarbase = planet.Spec.HasStarbase
		intel.Spec.HasMassDriver = planet.Spec.HasMassDriver
		intel.Spec.HasStargate = planet.Spec.HasStargate
		intel.Spec.DockCapacity = planet.Spec.DockCapacity

		// discover defense coverage
		intel.Spec.DefenseCoverage = planet.Spec.DefenseCoverage
		intel.Spec.DefenseCoverageSmart = planet.Spec.DefenseCoverageSmart

		// these should never be nil...
		if !ownedByPlayer && planet.Spec.HasStarbase && planet.Starbase != nil && planet.Starbase.Tokens[0].design != nil {
			design := planet.Starbase.Tokens[0].design
			intel.Spec.StarbaseDesignName = design.Name
			intel.Spec.StarbaseDesignNum = design.Num
			d.discoverDesign(design, false)
		}

		// players know their planet pops, but other planets are slightly off
		if ownedByPlayer {
			intel.Spec.Population = planet.population()
		} else {
			var randomPopulationError = rules.random.Float64()*(rules.PopulationScannerError*2) - rules.PopulationScannerError
			intel.Spec.Population = MaxInt(0, int(float64(planet.population())*(1-randomPopulationError)))
		}
	}
	return nil
}

// discover only the planet owner, but nothing else about a planet
func (d *discover) clearPlanetOwnerIntel(planet *Planet) error {

	player := d.player
	var intel *PlanetIntel
	planetIndex := planet.Num - 1

	if planetIndex < 0 || planetIndex >= len(player.PlanetIntels) {
		return fmt.Errorf("player %s cannot discover planet %s, planetIndex %d out of range", player, planet, planetIndex)
	}

	intel = &player.PlanetIntels[planetIndex]

	// if we've been invaded, reset our planet knowledge as if it was
	// unowned, but we maintain knowledge of hab
	intel.PlayerNum = Unowned
	intel.Spec.Population = 0
	intel.Spec.HasStarbase = false
	intel.Spec.HasStargate = false
	intel.Spec.DockCapacity = None
	intel.Spec.HasMassDriver = false
	intel.Spec.StarbaseDesignName = ""
	intel.Spec.StarbaseDesignNum = 0
	intel.ReportAge = 0

	d.log.Debug().
		Int("Planet", planet.Num).
		Msgf("player cleared planet owner intel")

	return nil
}

// discover a planet's starbase specs after a battle
func (d *discover) discoverPlanetStarbase(planet *Planet) error {
	player := d.player
	var intel *PlanetIntel
	planetIndex := planet.Num - 1

	if planetIndex < 0 || planetIndex >= len(player.PlanetIntels) {
		return fmt.Errorf("player %s cannot discover planet %s, planetIndex %d out of range", player, planet, planetIndex)
	}

	intel = &player.PlanetIntels[planetIndex]

	// discover starbases on scan, but don't discover designs
	intel.Spec.HasStarbase = planet.Spec.HasStarbase
	intel.Spec.HasStargate = planet.Spec.HasStargate
	intel.Spec.HasMassDriver = planet.Spec.HasMassDriver
	intel.Spec.DockCapacity = planet.Spec.DockCapacity

	return nil
}

// discover the cargo of a planet
func (d *discover) discoverPlanetCargo(planet *Planet) error {

	player := d.player
	var intel *PlanetIntel
	planetIndex := planet.Num - 1

	if planetIndex < 0 || planetIndex >= len(player.PlanetIntels) {
		return fmt.Errorf("player %s cannot discover planet %s, planetIndex %d out of range", player, planet, planetIndex)
	}

	intel = &player.PlanetIntels[planetIndex]

	intel.CargoDiscovered = true
	intel.Cargo = Cargo{
		Ironium:   planet.Cargo.Ironium,
		Boranium:  planet.Cargo.Boranium,
		Germanium: planet.Cargo.Germanium,
	}

	return nil
}

func (d *discover) discoverPlanetScanner(planet *Planet) error {

	player := d.player
	var intel *PlanetIntel
	planetIndex := planet.Num - 1

	if planetIndex < 0 || planetIndex >= len(player.PlanetIntels) {
		return fmt.Errorf("player %s cannot discover planet %s, planetIndex %d out of range", player, planet, planetIndex)
	}

	intel = &player.PlanetIntels[planetIndex]

	intel.Spec.Scanner = planet.Spec.Scanner
	intel.Spec.ScanRange = planet.Spec.ScanRange
	intel.Spec.ScanRangePen = planet.Spec.ScanRangePen

	return nil
}

func (d *discover) discoverPlanetTerraformability(planetNum int) error {
	player := d.player
	var intel *PlanetIntel
	planetIndex := planetNum - 1

	if planetIndex < 0 || planetIndex >= len(player.PlanetIntels) {
		return fmt.Errorf("planetIndex %d out of range", planetIndex)
	}

	intel = &player.PlanetIntels[planetIndex]

	// if we've discovered this planet before, update the terraform stats
	if intel.ReportAge != ReportAgeUnexplored {
		// terraforming
		terraformer := NewTerraformer()
		intel.Spec.TerraformAmount = terraformer.getTerraformAmount(intel.Hab, intel.BaseHab, player, player)
		intel.Spec.MinTerraformAmount = terraformer.getMinTerraformAmount(intel.Hab, intel.BaseHab, player, player)
		intel.Spec.CanTerraform = intel.Spec.TerraformAmount.absSum() > 0
		intel.Spec.TerraformedHabitability = player.Race.GetPlanetHabitability(intel.Hab.Add(intel.Spec.TerraformAmount))
	}
	return nil
}

// discover a fleet and add it to the player's fleet intel
func (d *discover) discoverFleet(fleet *Fleet, discoverName bool) {
	player := d.player
	intel := player.getFleetIntel(fleet.PlayerNum, fleet.Num)
	if intel == nil {
		// discover this new mineField
		intel = newFleetIntel(fleet.PlayerNum, fleet.Num)
		player.FleetIntels = append(player.FleetIntels, *intel)
		intel = &player.FleetIntels[len(player.FleetIntels)-1]
		d.log.Debug().
			Int("FleetPlayer", fleet.PlayerNum).
			Int("Fleet", fleet.Num).
			Msgf("player discovered fleet")
	}

	intel.BaseName = fleet.Tokens[0].design.Hull
	intel.Name = fmt.Sprintf("%s #%d", fleet.Tokens[0].design.Hull, fleet.Num)

	// we don't learn the fleet name, just the name of the first design in the fleet
	designIntel := d.player.getShipDesignIntel(fleet.PlayerNum, fleet.Tokens[0].DesignNum)
	if designIntel != nil {
		intel.BaseName = designIntel.Name
		intel.Name = fmt.Sprintf("%s #%d", designIntel.Name, fleet.Num)
	}

	if discoverName {
		// ally's tell us the names of their fleets
		intel.BaseName = fleet.BaseName
		intel.Name = fleet.Name
	}
	intel.Position = fleet.Position
	intel.OrbitingPlanetNum = fleet.OrbitingPlanetNum
	intel.Heading = fleet.Heading
	intel.WarpSpeed = fleet.WarpSpeed
	intel.Mass = fleet.Spec.Mass
	intel.Freighter = fleet.Spec.CargoCapacity > 0
	intel.Tokens = fleet.Tokens

}

// discover cargo for an existing fleet
func (d *discover) discoverFleetCargo(fleet *Fleet) {
	player := d.player
	existingIntel := player.getFleetIntel(fleet.PlayerNum, fleet.Num)
	if existingIntel != nil {
		existingIntel.Cargo = fleet.Cargo
		existingIntel.CargoDiscovered = true
	}
}

func (d *discover) discoverFleetScanner(fleet *Fleet) {
	player := d.player
	existingIntel := player.getFleetIntel(fleet.PlayerNum, fleet.Num)
	if existingIntel != nil {
		existingIntel.ScanRange = fleet.Spec.ScanRange
		existingIntel.ScanRangePen = fleet.Spec.ScanRangePen
	}
}

// discover a salvage and add it to the player's salvage intel
func (d *discover) discoverSalvage(salvage *Salvage) {
	player := d.player
	intel := player.getSalvageIntel(salvage.Num)
	if intel == nil {
		// discover this new wormhole
		player.SalvageIntels = append(player.SalvageIntels, *newSalvageIntel(salvage.PlayerNum, salvage.Num))
		intel = &player.SalvageIntels[len(player.SalvageIntels)-1]

		d.log.Debug().
			Int("SalvagePlayer", salvage.PlayerNum).
			Int("Salvage", salvage.Num).
			Msgf("player discovered salvage")
	}

	intel.Name = salvage.Name
	intel.PlayerNum = salvage.PlayerNum
	intel.Position = salvage.Position
	intel.Cargo = salvage.Cargo

}

// discover a mineField and add it to the player's mineField intel
func (d *discover) discoverMineField(mineField *MineField) {
	player := d.player
	intel := player.getMineFieldIntel(mineField.PlayerNum, mineField.Num)
	if intel == nil {
		// discover this new mineField
		intel = newMineFieldIntel(mineField.PlayerNum, mineField.Num)
		player.MineFieldIntels = append(player.MineFieldIntels, *intel)
		intel = &player.MineFieldIntels[len(player.MineFieldIntels)-1]
		d.log.Debug().
			Int("MineFieldPlayer", mineField.PlayerNum).
			Int("MineField", mineField.Num).
			Msgf("player discovered minefield")
	}

	intel.Name = mineField.Name
	intel.Position = mineField.Position
	intel.MineFieldType = mineField.MineFieldType
	intel.NumMines = mineField.NumMines
	intel.Spec.Radius = mineField.Spec.Radius
}

// discover a mineralPacket and add it to the player's mineralPacket intel
func (d *discover) discoverMineralPacket(rules *Rules, mineralPacket *MineralPacket, packetPlayer *Player, target *Planet) {
	player := d.player
	intel := player.getMineralPacketIntel(mineralPacket.PlayerNum, mineralPacket.Num)
	if intel == nil {
		// discover this new mineralPacket
		intel = newMineralPacketIntel(mineralPacket.PlayerNum, mineralPacket.Num)
		player.MineralPacketIntels = append(player.MineralPacketIntels, *intel)
		intel = &player.MineralPacketIntels[len(player.MineralPacketIntels)-1]
		d.log.Debug().
			Int("MineralPacketPlayer", mineralPacket.PlayerNum).
			Int("MineralPacket", mineralPacket.Num).
			Msgf("player discovered mineral packet")
	}

	if player.Num != mineralPacket.PlayerNum {
		if target.PlayerNum == player.Num {
			damage := mineralPacket.estimateDamage(rules, packetPlayer, target, player)
			messager.mineralPacketDiscoveredTargettingPlayer(player, mineralPacket, target, damage)
		} else {
			messager.mineralPacketDiscovered(player, mineralPacket, target)
		}
	}

	intel.Name = mineralPacket.Name
	intel.Position = mineralPacket.Position
	intel.Heading = mineralPacket.Heading
	intel.WarpSpeed = mineralPacket.WarpSpeed
	intel.Cargo = mineralPacket.Cargo
	intel.TargetPlanetNum = mineralPacket.TargetPlanetNum
}

func (d *discover) discoverMineralPacketScanner(mineralPacket *MineralPacket) {
	player := d.player
	existingIntel := player.getMineralPacketIntel(mineralPacket.PlayerNum, mineralPacket.Num)
	if existingIntel != nil {
		existingIntel.ScanRange = mineralPacket.ScanRange
		existingIntel.ScanRangePen = mineralPacket.ScanRangePen
	}
}

func (d *discover) discoverMineralPacketCargo(mineralPacket *MineralPacket) {
	player := d.player
	existingIntel := player.getMineralPacketIntel(mineralPacket.PlayerNum, mineralPacket.Num)
	if existingIntel != nil {
		existingIntel.Cargo = mineralPacket.Cargo
	}
}

// discover a player's design. This is a noop if we already know about
// the design and aren't discovering slots
func (d *discover) discoverDesign(design *ShipDesign, discoverSlots bool) {
	player := d.player
	intel := player.getShipDesignIntel(design.PlayerNum, design.Num)
	if intel == nil {
		// create a new intel for this design
		intel = &ShipDesignIntel{
			Intel: Intel{
				Name:      design.Hull,
				PlayerNum: design.PlayerNum,
				Num:       design.Num,
			},
			Hull:          design.Hull,
			HullSetNumber: design.HullSetNumber,
		}

		// discover mass even without scanning components
		intel.Spec.Mass = design.Spec.Mass

		// save this new design to our intel
		player.ShipDesignIntels = append(player.ShipDesignIntels, *intel)
		intel = &player.ShipDesignIntels[len(player.ShipDesignIntels)-1]

		d.log.Debug().
			Int("ShipDesignPlayer", design.PlayerNum).
			Int("ShipDesign", design.Num).
			Msgf("player discovered design")
	}

	if intel.Hull != design.Hull ||
		intel.HullSetNumber != design.HullSetNumber ||
		intel.Spec.Mass != design.Spec.Mass ||
		(len(intel.Slots) != 0 && !design.SlotsEqual(intel.Slots)) {

		// we discovered a design with this number already, but this is a new design
		// rediscover it
		intel.Name = design.Hull
		intel.PlayerNum = design.PlayerNum
		intel.Num = design.Num
		intel.Hull = design.Hull
		intel.HullSetNumber = design.HullSetNumber
		intel.Spec = ShipDesignSpec{
			Mass: design.Spec.Mass,
		}
		intel.Slots = []ShipDesignSlot{}
		d.log.Debug().
			Int("ShipDesignPlayer", design.PlayerNum).
			Int("ShipDesign", design.Num).
			Msgf("player rediscovered design")

	}

	// discover slots if we haven't already and this scanner discovers them
	if discoverSlots {
		if len(intel.Slots) == 0 {
			// our first time discovering slots
			intel.Slots = make([]ShipDesignSlot, len(design.Slots))
			copy(intel.Slots, design.Slots)
		}

		// if we discovered slots, also discover the name
		intel.Name = design.Name

		// discover stats about the design
		intel.Spec.Mass = design.Spec.Mass
		intel.Spec.Armor = design.Spec.Armor
		intel.Spec.Shields = design.Spec.Shields
		intel.Spec.PowerRating = design.Spec.PowerRating
		intel.Spec.FuelCapacity = design.Spec.FuelCapacity
		intel.Spec.Movement = design.Spec.Movement
		intel.Spec.Initiative = design.Spec.Initiative
		intel.Spec.CloakPercent = design.Spec.CloakPercent
		intel.Spec.BeamBonus = design.Spec.BeamBonus
		intel.Spec.BeamDefense = design.Spec.BeamDefense
		intel.Spec.TorpedoBonus = design.Spec.TorpedoBonus
		intel.Spec.TorpedoJamming = design.Spec.TorpedoJamming
		intel.Spec.ReduceCloaking = design.Spec.ReduceCloaking
		intel.Spec.ReduceMovement = design.Spec.ReduceMovement

		d.log.Debug().
			Int("ShipDesignPlayer", design.PlayerNum).
			Int("ShipDesign", design.Num).
			Msgf("player discovered design slots")
	}
}

// discover a wormhole and add it to the player's wormhole intel
func (d *discover) discoverWormhole(wormhole *Wormhole) {
	player := d.player
	intel := player.getWormholeIntel(wormhole.Num)
	if intel == nil {
		// discover this new wormhole
		player.WormholeIntels = append(player.WormholeIntels, *newWormholeIntel(wormhole.Num))
		intel = &player.WormholeIntels[len(player.WormholeIntels)-1]
		d.log.Debug().
			Int("Wormhole", wormhole.Num).
			Msgf("player discovered wormhole")
	}

	intel.Name = wormhole.Name
	intel.Position = wormhole.Position
	intel.Stability = wormhole.Stability
}

func (d *discover) discoverWormholeLink(wormhole1, wormhole2 *Wormhole) {
	player := d.player
	intel1 := player.getWormholeIntel(wormhole1.Num)
	if intel1 == nil {
		// discover this new wormhole
		player.WormholeIntels = append(player.WormholeIntels, *newWormholeIntel(wormhole1.Num))
		intel1 = &player.WormholeIntels[len(player.WormholeIntels)-1]
		d.log.Debug().
			Int("Wormhole1", wormhole1.Num).
			Msgf("player discovered wormhole1 link")
	}

	intel2 := player.getWormholeIntel(wormhole2.Num)
	if intel2 == nil {
		// discover this new wormhole
		player.WormholeIntels = append(player.WormholeIntels, *newWormholeIntel(wormhole2.Num))
		intel2 = &player.WormholeIntels[len(player.WormholeIntels)-1]
		d.log.Debug().
			Int("Wormhole2", wormhole1.Num).
			Msgf("player discovered wormhole2 link")
	}

	intel1.Name = wormhole1.Name
	intel1.Position = wormhole1.Position
	intel1.Stability = wormhole1.Stability
	intel1.DestinationNum = wormhole1.DestinationNum

	intel2.Name = wormhole2.Name
	intel2.Position = wormhole2.Position
	intel2.Stability = wormhole2.Stability
	intel2.DestinationNum = wormhole2.DestinationNum
}

// forget about a wormhole
func (d *discover) forgetWormhole(num int) {
	player := d.player
	intel := player.getWormholeIntel(num)

	if intel == nil {
		// no wormhole to forget
		return
	}

	// remeber the dest
	dest := intel.DestinationNum

	// forget this wormhole
	player.WormholeIntels = slices.DeleteFunc(player.WormholeIntels, func(w WormholeIntel) bool { return w.Num == num })
	d.log.Debug().
		Int("Wormhole", num).
		Msgf("player forgot wormhole")

		// if we knew the destination, remove the link
	intelLink := player.getWormholeIntel(dest)
	if intelLink != nil {
		intelLink.DestinationNum = None
	}
}

// discover a mysteryTrader and add it to the player's mysteryTrader intel
func (d *discover) discoverMysteryTrader(mysteryTrader *MysteryTrader) {
	player := d.player
	intel := player.getMysteryTraderIntel(mysteryTrader.Num)
	if intel == nil {
		// discover this new mysteryTrader
		player.MysteryTraderIntels = append(player.MysteryTraderIntels, *newMysteryTraderIntel(mysteryTrader.Num))
		intel = &player.MysteryTraderIntels[len(player.MysteryTraderIntels)-1]
		d.log.Debug().
			Int("MysteryTrader", mysteryTrader.Num).
			Msgf("player discovered mysteryTrader")
	}

	intel.Name = mysteryTrader.Name
	intel.Position = mysteryTrader.Position
	intel.Heading = mysteryTrader.Heading
	intel.WarpSpeed = mysteryTrader.WarpSpeed
	intel.RequestedBoon = mysteryTrader.RequestedBoon
}

// discover a player's race
func (d *discover) discoverPlayer(player *Player) {
	intel := &d.player.PlayerIntels.PlayerIntels[player.Num-1]

	if !intel.Seen {
		d.log.Debug().Msgf("player %s discovered %s", d.player.Name, player.Name)
		messager.playerDiscovered(d.player, player)
		intel.Seen = true
		intel.Name = player.Name
		intel.Color = player.Color
		intel.RaceName = player.Race.Name
		intel.RacePluralName = player.Race.PluralName
	}
}

// discover a player's score
func (d *discover) discoverPlayerScores(player *Player) {
	intel := &d.player.PlayerIntels.ScoreIntels[player.Num-1]

	intel.ScoreHistory = make([]PlayerScore, len(player.ScoreHistory))
	copy(intel.ScoreHistory, player.ScoreHistory)
}

func (d *discovererWithAllies) discoverPlayer(player *Player) {
	d.playerDiscoverer.discoverPlayer(player)
	for _, allyDiscoverer := range d.allyDiscoverers {
		if allyDiscoverer.player.Num != player.Num {
			allyDiscoverer.discoverPlayer(player)
		}
	}
}
func (d *discovererWithAllies) discoverPlayerScores(player *Player) {
	d.playerDiscoverer.discoverPlayerScores(player)
	for _, allyDiscoverer := range d.allyDiscoverers {
		if allyDiscoverer.player.Num != player.Num {
			allyDiscoverer.discoverPlayerScores(player)
		}
	}
}
func (d *discovererWithAllies) discoverPlanet(rules *Rules, planet *Planet, penScanned bool) error {
	if err := d.playerDiscoverer.discoverPlanet(rules, planet, penScanned); err != nil {
		return err
	}
	for _, allyDiscoverer := range d.allyDiscoverers {
		if allyDiscoverer.player.Num != planet.PlayerNum {
			if err := allyDiscoverer.discoverPlanet(rules, planet, penScanned); err != nil {
				return err
			}
		}
	}
	return nil
}

func (d *discovererWithAllies) clearPlanetOwnerIntel(planet *Planet) error {
	if err := d.playerDiscoverer.clearPlanetOwnerIntel(planet); err != nil {
		return err
	}
	for _, allyDiscoverer := range d.allyDiscoverers {
		if allyDiscoverer.player.Num != planet.PlayerNum {
			if err := allyDiscoverer.clearPlanetOwnerIntel(planet); err != nil {
				return err
			}
		}
	}
	return nil
}

func (d *discovererWithAllies) discoverPlanetStarbase(planet *Planet) error {
	if err := d.playerDiscoverer.discoverPlanetStarbase(planet); err != nil {
		return err
	}
	for _, allyDiscoverer := range d.allyDiscoverers {
		if allyDiscoverer.player.Num != planet.PlayerNum {
			if err := allyDiscoverer.discoverPlanetStarbase(planet); err != nil {
				return err
			}
		}
	}
	return nil
}

func (d *discovererWithAllies) discoverPlanetCargo(planet *Planet) error {
	if err := d.playerDiscoverer.discoverPlanetCargo(planet); err != nil {
		return err
	}
	for _, allyDiscoverer := range d.allyDiscoverers {
		if allyDiscoverer.player.Num != planet.PlayerNum {
			if err := allyDiscoverer.discoverPlanetCargo(planet); err != nil {
				return err
			}
		}
	}
	return nil
}

func (d *discovererWithAllies) discoverPlanetScanner(planet *Planet) error {
	if err := d.playerDiscoverer.discoverPlanetScanner(planet); err != nil {
		return err
	}
	for _, allyDiscoverer := range d.allyDiscoverers {
		if allyDiscoverer.player.Num != planet.PlayerNum {
			if err := allyDiscoverer.discoverPlanetScanner(planet); err != nil {
				return err
			}
		}
	}
	return nil
}

func (d *discovererWithAllies) discoverPlanetTerraformability(planetNum int) error {
	if err := d.playerDiscoverer.discoverPlanetTerraformability(planetNum); err != nil {
		return err
	}
	for _, allyDiscoverer := range d.allyDiscoverers {
		if err := allyDiscoverer.discoverPlanetTerraformability(planetNum); err != nil {
			return err
		}
	}
	return nil
}

func (d *discovererWithAllies) discoverFleet(fleet *Fleet, discoverName bool) {
	d.playerDiscoverer.discoverFleet(fleet, discoverName)
	for _, allyDiscoverer := range d.allyDiscoverers {
		if allyDiscoverer.player.Num != fleet.PlayerNum {
			allyDiscoverer.discoverFleet(fleet, discoverName)
		}
	}
}

func (d *discovererWithAllies) discoverFleetCargo(fleet *Fleet) {
	d.playerDiscoverer.discoverFleetCargo(fleet)
	for _, allyDiscoverer := range d.allyDiscoverers {
		if allyDiscoverer.player.Num != fleet.PlayerNum {
			allyDiscoverer.discoverFleetCargo(fleet)
		}
	}
}

func (d *discovererWithAllies) discoverFleetScanner(fleet *Fleet) {
	d.playerDiscoverer.discoverFleetScanner(fleet)
	for _, allyDiscoverer := range d.allyDiscoverers {
		if allyDiscoverer.player.Num != fleet.PlayerNum {
			allyDiscoverer.discoverFleetScanner(fleet)
		}
	}
}

func (d *discovererWithAllies) discoverMineField(mineField *MineField) {
	d.playerDiscoverer.discoverMineField(mineField)
	for _, allyDiscoverer := range d.allyDiscoverers {
		if allyDiscoverer.player.Num != mineField.PlayerNum {
			allyDiscoverer.discoverMineField(mineField)
		}
	}
}

func (d *discovererWithAllies) discoverMineralPacket(rules *Rules, mineralPacket *MineralPacket, packetPlayer *Player, target *Planet) {
	d.playerDiscoverer.discoverMineralPacket(rules, mineralPacket, packetPlayer, target)
	for _, allyDiscoverer := range d.allyDiscoverers {
		if allyDiscoverer.player.Num != mineralPacket.PlayerNum {
			allyDiscoverer.discoverMineralPacket(rules, mineralPacket, packetPlayer, target)
		}
	}
}

func (d *discovererWithAllies) discoverMineralPacketScanner(mineralPacket *MineralPacket) {
	d.playerDiscoverer.discoverMineralPacketScanner(mineralPacket)
	for _, allyDiscoverer := range d.allyDiscoverers {
		if allyDiscoverer.player.Num != mineralPacket.PlayerNum {
			allyDiscoverer.discoverMineralPacketScanner(mineralPacket)
		}
	}
}

func (d *discovererWithAllies) discoverMineralPacketCargo(mineralPacket *MineralPacket) {
	d.playerDiscoverer.discoverMineralPacketCargo(mineralPacket)
	for _, allyDiscoverer := range d.allyDiscoverers {
		if allyDiscoverer.player.Num != mineralPacket.PlayerNum {
			allyDiscoverer.discoverMineralPacketCargo(mineralPacket)
		}
	}
}

func (d *discovererWithAllies) discoverSalvage(salvage *Salvage) {
	d.playerDiscoverer.discoverSalvage(salvage)
	for _, allyDiscoverer := range d.allyDiscoverers {
		if allyDiscoverer.player.Num != salvage.PlayerNum {
			allyDiscoverer.discoverSalvage(salvage)
		}
	}
}

func (d *discovererWithAllies) discoverWormhole(wormhole *Wormhole) {
	d.playerDiscoverer.discoverWormhole(wormhole)
	for _, allyDiscoverer := range d.allyDiscoverers {
		allyDiscoverer.discoverWormhole(wormhole)
	}
}

func (d *discovererWithAllies) discoverWormholeLink(wormhole1, wormhole2 *Wormhole) {
	d.playerDiscoverer.discoverWormholeLink(wormhole1, wormhole2)
	for _, allyDiscoverer := range d.allyDiscoverers {
		allyDiscoverer.discoverWormholeLink(wormhole1, wormhole2)
	}
}

func (d *discovererWithAllies) forgetWormhole(num int) {
	d.playerDiscoverer.forgetWormhole(num)
	for _, allyDiscoverer := range d.allyDiscoverers {
		allyDiscoverer.forgetWormhole(num)
	}
}

func (d *discovererWithAllies) discoverMysteryTrader(mysteryTrader *MysteryTrader) {
	d.playerDiscoverer.discoverMysteryTrader(mysteryTrader)
	for _, allyDiscoverer := range d.allyDiscoverers {
		allyDiscoverer.discoverMysteryTrader(mysteryTrader)
	}
}

func (d *discovererWithAllies) discoverDesign(design *ShipDesign, discoverSlots bool) {
	d.playerDiscoverer.discoverDesign(design, discoverSlots)
	for _, allyDiscoverer := range d.allyDiscoverers {
		if allyDiscoverer.player.Num != design.PlayerNum {
			allyDiscoverer.discoverDesign(design, discoverSlots)
		}
	}
}
