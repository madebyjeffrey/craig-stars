package cs

import (
	"fmt"

	"slices"

	"github.com/rs/zerolog"
)

// The Universe represents all the objects that occupy space in the game universe.
// The Universe is used by the turn generator to lookup objects in space
type Universe struct {
	Planets              []*Planet        `json:"planets,omitempty"`
	Fleets               []*Fleet         `json:"fleets,omitempty"`
	Starbases            []*Fleet         `json:"starbases,omitempty"`
	Wormholes            []*Wormhole      `json:"wormholes,omitempty"`
	MineralPackets       []*MineralPacket `json:"mineralPackets,omitempty"`
	MineFields           []*MineField     `json:"mineFields,omitempty"`
	MysteryTraders       []*MysteryTrader `json:"mysteryTraders,omitempty"`
	Salvages             []*Salvage       `json:"salvage,omitempty"`
	battlePlansByNum     map[playerBattlePlanNum]*BattlePlan
	mapObjectsByPosition map[Vector][]interface{}
	fleetsByNum          map[playerObject]*Fleet
	designsByNum         map[playerObject]*ShipDesign
	mineFieldsByNum      map[playerObject]*MineField
	mineralPacketsByNum  map[playerObject]*MineralPacket
	salvagesByNum        map[int]*Salvage
	salvagesByPosition   map[Vector]*Salvage
	mysteryTradersByNum  map[int]*MysteryTrader
	wormholesByNum       map[int]*Wormhole
	log                  zerolog.Logger
}

func NewUniverse(log zerolog.Logger, rules *Rules) Universe {

	return Universe{
		battlePlansByNum:     make(map[playerBattlePlanNum]*BattlePlan),
		mapObjectsByPosition: make(map[Vector][]interface{}),
		designsByNum:         make(map[playerObject]*ShipDesign),
		fleetsByNum:          make(map[playerObject]*Fleet),
		mineFieldsByNum:      make(map[playerObject]*MineField),
		mineralPacketsByNum:  make(map[playerObject]*MineralPacket),
		salvagesByNum:        make(map[int]*Salvage),
		salvagesByPosition:   make(map[Vector]*Salvage),
		wormholesByNum:       make(map[int]*Wormhole),
		mysteryTradersByNum:  make(map[int]*MysteryTrader),
		log:                  log,
	}
}

type fleetGetter interface {
	getFleet(playerNum int, num int) *Fleet
}

type mapObjectGetter interface {
	getShipDesign(playerNum int, num int) *ShipDesign
	getMapObject(mapObjectType MapObjectType, num int, playerNum int) *MapObject
	getPlanet(num int) *Planet
	getOrbitingPlanet(fleet *Fleet) *Planet
	getFleet(playerNum int, num int) *Fleet
	getMineField(playerNum int, num int) *MineField
	getAllMineFields() []*MineField
	getMysteryTrader(num int) *MysteryTrader
	getWormhole(num int) *Wormhole
	getSalvage(num int) *Salvage
	getCargoHolder(mapObjectType MapObjectType, num int, playerNum int) (cargoHolder, bool)
	getMapObjectsAtPosition(position Vector) []interface{}
	isPositionValid(pos Vector, occupiedLocations *[]Vector, minDistance float64) bool
	updateMapObjectAtPosition(mo interface{}, originalPosition, newPosition Vector)
}

type playerObject struct {
	PlayerNum int
	Num       int
}

func playerObjectKey(playerNum int, num int) playerObject { return playerObject{playerNum, num} }

type playerBattlePlanNum struct {
	PlayerNum int
	Num       int
}

// build the maps used for the Get functions
func (u *Universe) buildMaps(players []*Player) error {

	// make a big map to hold all of our universe objects by position
	u.mapObjectsByPosition = make(map[Vector][]interface{}, len(u.Planets))

	// build a map of designs by num
	// so we can inject the design into each token
	numDesigns := 0
	numBattlePlans := 0
	for _, p := range players {
		numDesigns += len(p.Designs)
		numBattlePlans += len(p.BattlePlans)
	}
	u.designsByNum = make(map[playerObject]*ShipDesign, numDesigns)
	u.battlePlansByNum = make(map[playerBattlePlanNum]*BattlePlan, numBattlePlans)

	for _, p := range players {
		for _, design := range p.Designs {
			u.addDesign(design)
		}

		for i := range p.BattlePlans {
			plan := &p.BattlePlans[i]
			u.battlePlansByNum[playerBattlePlanNum{PlayerNum: p.Num, Num: plan.Num}] = plan
		}

		// create a discoverer for this player and any allies they share maps with
		p.discoverer = newDiscovererWithAllies(u.log, p, players)
	}

	u.fleetsByNum = make(map[playerObject]*Fleet, len(u.Fleets))
	for _, fleet := range u.Fleets {
		if fleet.Delete {
			continue
		}
		if err := u.addFleet(fleet); err != nil {
			return err
		}
	}

	for _, starbase := range u.Starbases {
		if starbase.Delete {
			continue
		}
		if err := u.addStarbase(starbase); err != nil {
			return err
		}
	}

	for _, planet := range u.Planets {
		u.addMapObjectByPosition(planet, planet.Position)
		// for owned planets, populate their designs
		if planet.Owned() {
			if planet.PlayerNum > 0 && planet.PlayerNum < len(players)+1 {
				player := players[planet.PlayerNum-1]
				if err := planet.PopulateProductionQueueDesigns(player); err != nil {
					return fmt.Errorf("planet %s unable to populate queue designs %w", planet.Name, err)
				}
			} else {
				return fmt.Errorf("planet %s owner %d out of range", planet.Name, planet.PlayerNum)
			}
		}

	}

	for _, mineralPacket := range u.MineralPackets {
		if mineralPacket.Delete {
			continue
		}
		u.mineralPacketsByNum[playerObjectKey(mineralPacket.PlayerNum, mineralPacket.Num)] = mineralPacket
		u.addMapObjectByPosition(mineralPacket, mineralPacket.Position)
	}
	for _, mineField := range u.MineFields {
		if mineField.Delete {
			continue
		}
		u.mineFieldsByNum[playerObjectKey(mineField.PlayerNum, mineField.Num)] = mineField
		u.addMapObjectByPosition(mineField, mineField.Position)
	}

	u.salvagesByPosition = make(map[Vector]*Salvage, len(u.Salvages))
	u.salvagesByNum = make(map[int]*Salvage, len(u.Salvages))
	for _, salvage := range u.Salvages {
		if salvage.Delete {
			continue
		}
		u.salvagesByNum[salvage.Num] = salvage
		u.addMapObjectByPosition(salvage, salvage.Position)
		u.salvagesByPosition[salvage.Position] = salvage
	}

	u.wormholesByNum = make(map[int]*Wormhole, len(u.Wormholes))
	for _, wormhole := range u.Wormholes {
		if wormhole.Delete {
			continue
		}
		u.wormholesByNum[wormhole.Num] = wormhole
		u.addMapObjectByPosition(wormhole, wormhole.Position)
	}

	u.mysteryTradersByNum = make(map[int]*MysteryTrader, len(u.MysteryTraders))
	for _, mysteryTrader := range u.MysteryTraders {
		if mysteryTrader.Delete {
			continue
		}
		u.mysteryTradersByNum[mysteryTrader.Num] = mysteryTrader
		u.addMapObjectByPosition(mysteryTrader, mysteryTrader.Position)
	}

	return nil
}

func (u *Universe) addMapObjectByPosition(mo interface{}, position Vector) {
	mos, found := u.mapObjectsByPosition[position]
	if !found {
		mos = []interface{}{}
		u.mapObjectsByPosition[position] = mos
	}
	mos = append(mos, mo)
	u.mapObjectsByPosition[position] = mos
}

// Check if a position vector is a mininum distance away from all other points
func (u *Universe) isPositionValid(pos Vector, occupiedLocations *[]Vector, minDistance float64) bool {
	minDistanceSquared := minDistance * minDistance

	for _, to := range *occupiedLocations {
		if pos.DistanceSquaredTo(to) <= minDistanceSquared {
			return false
		}
	}
	return true
}

// get all commandable map objects for a player
func (u *Universe) GetPlayerMapObjects(playerNum int) PlayerMapObjects {
	pmo := PlayerMapObjects{}

	pmo.Planets = u.getPlanets(playerNum)
	pmo.Fleets = u.getFleets(playerNum)
	pmo.Starbases = u.getStarbases(playerNum)
	pmo.MineFields = u.getMineFields(playerNum)
	pmo.MineralPackets = u.getMineralPackets(playerNum)

	return pmo
}

func (u *Universe) getMapObject(mapObjectType MapObjectType, num int, playerNum int) *MapObject {
	switch mapObjectType {
	case MapObjectTypePlanet:
		planet := u.getPlanet(num)
		if planet != nil {
			return &planet.MapObject
		}
	case MapObjectTypeFleet:
		fleet := u.getFleet(playerNum, num)
		if fleet != nil {
			return &fleet.MapObject
		}
	case MapObjectTypeWormhole:
		wormhole := u.getWormhole(num)
		if wormhole != nil {
			return &wormhole.MapObject
		}
	case MapObjectTypeMineField:
		mineField := u.getMineField(playerNum, num)
		if mineField != nil {
			return &mineField.MapObject
		}
	case MapObjectTypeMysteryTrader:
		mysteryTrader := u.getMysteryTrader(num)
		if mysteryTrader != nil {
			return &mysteryTrader.MapObject
		}
	case MapObjectTypeSalvage:
		salvage := u.getSalvage(num)
		if salvage != nil {
			return &salvage.MapObject
		}
	case MapObjectTypeMineralPacket:
		mineralPacket := u.getMineralPacket(playerNum, num)
		if mineralPacket != nil {
			return &mineralPacket.MapObject
		}
	}
	return nil
}

// get a ship design by num
func (u *Universe) getShipDesign(playerNum int, num int) *ShipDesign {
	return u.designsByNum[playerObjectKey(playerNum, num)]
}

// Get a planet by num
func (u *Universe) getPlanet(num int) *Planet {
	if num < 1 || num > len(u.Planets) {
		return nil
	}
	return u.Planets[num-1]
}

// get the planet this fleet is orbiting, or nil if none
func (u *Universe) getOrbitingPlanet(fleet *Fleet) *Planet {
	if fleet.OrbitingPlanetNum == None {
		return nil
	}
	return u.getPlanet(fleet.OrbitingPlanetNum)
}

// Get a fleet by player num and fleet num
func (u *Universe) getFleet(playerNum int, num int) *Fleet {
	return u.fleetsByNum[playerObjectKey(playerNum, num)]
}

// Get a planet by num
func (u *Universe) getWormhole(num int) *Wormhole {
	return u.wormholesByNum[num]
}

// Get a salvage by num
func (u *Universe) getSalvage(num int) *Salvage {
	return u.salvagesByNum[num]
}

func (u *Universe) getMineField(playerNum int, num int) *MineField {
	return u.mineFieldsByNum[playerObjectKey(playerNum, num)]
}

func (u *Universe) getAllMineFields() []*MineField {
	return u.MineFields
}

// get a minefield that is close to a position
func (u *Universe) getMineFieldNearPosition(playerNum int, position Vector, mineFieldType MineFieldType) *MineField {
	for _, mineField := range u.MineFields {
		if mineField.PlayerNum == playerNum && mineField.MineFieldType == mineFieldType && isPointInCircle(position, mineField.Position, mineField.Spec.Radius) {
			return mineField
		}
	}

	return nil
}

func (u *Universe) getMineralPacket(playerNum int, num int) *MineralPacket {
	return u.mineralPacketsByNum[playerObjectKey(playerNum, num)]
}

func (u *Universe) getMysteryTrader(num int) *MysteryTrader {
	return u.mysteryTradersByNum[num]
}

// get a cargo holder by natural key (num, playerNum, etc)
func (u *Universe) getCargoHolder(mapObjectType MapObjectType, num int, playerNum int) (cargoHolder, bool) {
	switch mapObjectType {
	case MapObjectTypePlanet:
		mo := u.getPlanet(num)
		return mo, mo != nil
	case MapObjectTypeFleet:
		mo := u.getFleet(playerNum, num)
		return mo, mo != nil
	case MapObjectTypeMineralPacket:
		mo := u.getMineralPacket(playerNum, num)
		return mo, mo != nil
	case MapObjectTypeSalvage:
		mo := u.getSalvage(num)
		return mo, mo != nil
	}
	return nil, false
}

// update the num instances for all tokens
func (u *Universe) updateTokenCounts() {
	for _, design := range u.designsByNum {
		design.Spec.NumInstances = 0
	}
	for _, fleet := range append(u.Fleets, u.Starbases...) {
		if fleet.Delete {
			continue
		}
		for _, token := range fleet.Tokens {
			design := u.designsByNum[playerObjectKey(fleet.PlayerNum, token.DesignNum)]
			design.Spec.NumInstances += token.Quantity
		}
	}
}

// add a design to the universe maps
func (u *Universe) addDesign(design *ShipDesign) {
	u.designsByNum[playerObjectKey(design.PlayerNum, design.Num)] = design
}

// mark a fleet as deleted and remove it from the universe
func (u *Universe) deleteFleet(fleet *Fleet) {
	fleet.Delete = true

	delete(u.fleetsByNum, playerObjectKey(fleet.PlayerNum, fleet.Num))

	u.removeMapObjectAtPosition(fleet, fleet.Position)

	u.log.Debug().
		Int("Player", fleet.PlayerNum).
		Str("Fleet", fleet.Name).
		Msgf("deleted fleet")

}

// mark a starbase as deleted and remove it from the universe
func (u *Universe) deleteStarbase(starbase *Fleet) {
	starbase.Delete = true

	u.removeMapObjectAtPosition(starbase, starbase.Position)

	u.log.Debug().
		Int("Player", starbase.PlayerNum).
		Str("Starbase", starbase.Name).
		Msgf("deleted starbase")

}

// move a fleet from one position to another
func (u *Universe) moveFleet(fleet *Fleet, originalPosition Vector) {
	// upadte mapobjects position
	u.updateMapObjectAtPosition(fleet, originalPosition, fleet.Position)
}

func (u *Universe) addFleet(fleet *Fleet) error {
	u.addMapObjectByPosition(fleet, fleet.Position)

	u.fleetsByNum[playerObjectKey(fleet.PlayerNum, fleet.Num)] = fleet

	if battlePlan, ok := u.battlePlansByNum[playerBattlePlanNum{fleet.PlayerNum, fleet.BattlePlanNum}]; ok {
		fleet.battlePlan = battlePlan
	} else {
		// use the default battle plan if we couldn't find one for some reason, but log a warning
		u.log.Warn().
			Int("Player", fleet.PlayerNum).
			Msgf("Unable to find battle plan %d for fleet %v", fleet.BattlePlanNum, fleet)
		fleet.battlePlan = u.battlePlansByNum[playerBattlePlanNum{fleet.PlayerNum, 0}]
	}

	// inject the design into this
	for i := range fleet.Tokens {
		token := &fleet.Tokens[i]
		token.design = u.designsByNum[playerObjectKey(fleet.PlayerNum, token.DesignNum)]
		if token.design == nil {
			return fmt.Errorf("unable to find design %d for fleet %s", token.DesignNum, fleet.Name)
		}
	}
	return nil
}

func (u *Universe) addStarbase(starbase *Fleet) error {
	u.Planets[starbase.PlanetNum-1].Starbase = starbase

	u.addMapObjectByPosition(starbase, starbase.Position)

	starbase.battlePlan = u.battlePlansByNum[playerBattlePlanNum{starbase.PlayerNum, starbase.BattlePlanNum}]

	// inject the design into this
	for i := range starbase.Tokens {
		token := &starbase.Tokens[i]
		token.design = u.designsByNum[playerObjectKey(starbase.PlayerNum, token.DesignNum)]
		if token.design == nil {
			return fmt.Errorf("unable to find design %d for fleet %s", token.DesignNum, starbase.Name)
		}
	}

	return nil
}

// move a wormhole from one position to another
func (u *Universe) moveWormhole(wormhole *Wormhole, originalPosition Vector) {
	u.updateMapObjectAtPosition(wormhole, originalPosition, wormhole.Position)
}

// delete a wormhole from the universe
func (u *Universe) deleteWormhole(wormhole *Wormhole) {
	wormhole.Delete = true

	delete(u.wormholesByNum, wormhole.Num)
	u.removeMapObjectAtPosition(wormhole, wormhole.Position)

	u.log.Debug().
		Int("Player", wormhole.PlayerNum).
		Str("Wormhole", wormhole.Name).
		Msgf("deleted wormhole")

}

// create a new wormhole in the universe
func (u *Universe) createWormhole(rules *Rules, position Vector, stability WormholeStability, companion *Wormhole) *Wormhole {
	num := 1
	if len(u.Wormholes) > 0 {
		num = u.Wormholes[len(u.Wormholes)-1].Num + 1
	}

	wormhole := newWormhole(position, num, stability)

	if companion != nil {
		companion.DestinationNum = wormhole.Num
		wormhole.DestinationNum = companion.Num
	}

	// compute the spec for this wormhole
	wormhole.Spec = computeWormholeSpec(wormhole, rules)

	u.Wormholes = append(u.Wormholes, wormhole)
	u.wormholesByNum[wormhole.Num] = wormhole
	u.addMapObjectByPosition(wormhole, wormhole.Position)

	return wormhole
}

// find the salvage at a position, or create a new one
func (u *Universe) createSalvage(position Vector, playerNum int, cargo Cargo) *Salvage {
	salvage, exists := u.salvagesByPosition[position]
	// Check for empty salvage, because they are deleted from database, but not mapObjects
	if exists && (salvage.Cargo != Cargo{}) {
		salvage.Cargo = salvage.Cargo.Add(cargo)
		return salvage
	}
	num := 1
	if len(u.Salvages) > 0 {
		num = u.Salvages[len(u.Salvages)-1].Num + 1
	}
	salvage = newSalvage(position, num, playerNum, cargo)
	u.Salvages = append(u.Salvages, salvage)
	u.salvagesByNum[num] = salvage
	u.addMapObjectByPosition(salvage, salvage.Position)

	return salvage
}

// delete a salvage from the universe
func (u *Universe) deleteSalvage(salvage *Salvage) {
	salvage.Delete = true

	delete(u.salvagesByNum, salvage.Num)
	u.removeMapObjectAtPosition(salvage, salvage.Position)
}

// delete a salvage from the universe
func (u *Universe) deletePacket(packet *MineralPacket) {
	packet.Delete = true

	delete(u.mineralPacketsByNum, playerObjectKey(packet.PlayerNum, packet.Num))
	u.removeMapObjectAtPosition(packet, packet.Position)
}

func (u *Universe) getPlanets(playerNum int) []*Planet {
	planets := []*Planet{}
	for _, planet := range u.Planets {
		if planet.PlayerNum == playerNum {
			planets = append(planets, planet)
		}
	}
	return planets
}

func (u *Universe) getFleets(playerNum int) []*Fleet {
	fleets := []*Fleet{}
	for _, fleet := range u.Fleets {
		if fleet.Delete {
			continue
		}
		if fleet.PlayerNum == playerNum {
			fleets = append(fleets, fleet)
		}
	}
	return fleets
}

func (u *Universe) getStarbases(playerNum int) []*Fleet {
	starbases := []*Fleet{}
	for _, starbase := range u.Starbases {
		if starbase.Delete {
			continue
		}
		if starbase.PlayerNum == playerNum {
			starbases = append(starbases, starbase)
		}
	}
	return starbases
}

func (u *Universe) getMineFields(playerNum int) []*MineField {
	mineFields := []*MineField{}
	for _, mineField := range u.MineFields {
		if mineField.Delete {
			continue
		}

		if mineField.PlayerNum == playerNum {
			mineFields = append(mineFields, mineField)
		}
	}
	return mineFields
}

func (u *Universe) getMineralPackets(playerNum int) []*MineralPacket {
	mineralPackets := []*MineralPacket{}
	for _, mineralPacket := range u.MineralPackets {
		if mineralPacket.Delete {
			continue
		}

		if mineralPacket.PlayerNum == playerNum {
			mineralPackets = append(mineralPackets, mineralPacket)
		}
	}
	return mineralPackets
}

// get a slice of mapobjects at a position, or nil if none
func (u *Universe) getMapObjectsAtPosition(position Vector) []interface{} {
	return u.mapObjectsByPosition[position]
}

// get a slice of mapobjects at a position, or nil if none
func (u *Universe) updateMapObjectAtPosition(mo interface{}, originalPosition, newPosition Vector) {
	mos := u.mapObjectsByPosition[originalPosition]
	if mos != nil {
		updatedMos := make([]interface{}, 0, len(mos)-1)
		for _, existingMo := range mos {
			if existingMo == mo {
				continue
			}
			updatedMos = append(updatedMos, existingMo)
		}
		u.mapObjectsByPosition[originalPosition] = updatedMos
	} else {
		u.log.Warn().Msgf("tried to update position of %s from %v to %v, no mapobjects were found at %v", mo, originalPosition, newPosition, originalPosition)
	}

	// add the new object to the list
	u.addMapObjectByPosition(mo, newPosition)
}

// get a slice of mapobjects at a position, or nil if none
func (u *Universe) removeMapObjectAtPosition(mo interface{}, position Vector) {
	mos := u.mapObjectsByPosition[position]
	if mos != nil {
		index := slices.IndexFunc(mos, func(item interface{}) bool { return item == mo })
		if index >= 0 && index < len(mos) {
			u.mapObjectsByPosition[position] = slices.Delete(mos, index, index+1)
		} else {
			u.log.Warn().Msgf("tried to remove mapobject %s at position %v but index %d of position out of range", mo, position, index)
		}
	} else {
		u.log.Warn().Msgf("tried to to remove mapobject %s at position %v, no mapobjects were found at %v", mo, position, position)
	}
}

// get the next mineField number to use
func (u *Universe) getNextMineFieldNum() int {
	num := 0
	for _, mineField := range u.MineFields {
		num = MaxInt(num, mineField.Num)
	}
	return num + 1
}

func (u *Universe) addMineField(mineField *MineField) {
	u.MineFields = append(u.MineFields, mineField)
	u.mineFieldsByNum[playerObjectKey(mineField.PlayerNum, mineField.Num)] = mineField
	u.addMapObjectByPosition(mineField, mineField.Position)
}

// mark a mineField as deleted and remove it from the universe
func (u *Universe) deleteMineField(mineField *MineField) {
	mineField.Delete = true

	delete(u.mineFieldsByNum, playerObjectKey(mineField.PlayerNum, mineField.Num))
	u.removeMapObjectAtPosition(mineField, mineField.Position)

	u.log.Debug().
		Int("Player", mineField.PlayerNum).
		Str("MineField", mineField.Name).
		Msgf("deleted mineField")

}

// get the number of planets within a circle
func (u *Universe) numPlanetsWithin(position Vector, radius float64) (numPlanets int) {
	for _, planet := range u.Planets {
		if isPointInCircle(planet.Position, position, radius) {
			numPlanets++
		}
	}
	return numPlanets
}

// get fleets within a circle
func (u *Universe) fleetsWithin(position Vector, radius float64) []*Fleet {
	fleetsWithin := make([]*Fleet, 0, 10)
	for _, fleet := range u.Fleets {
		if isPointInCircle(fleet.Position, position, radius) {
			fleetsWithin = append(fleetsWithin, fleet)
		}
	}
	return fleetsWithin
}

// get the next mysteryTrader number to use
func (u *Universe) getNextMysteryTraderNum() int {
	num := 0
	for _, mysteryTrader := range u.MysteryTraders {
		num = MaxInt(num, mysteryTrader.Num)
	}
	return num + 1
}

func (u *Universe) addMysteryTrader(mysteryTrader *MysteryTrader) {
	u.MysteryTraders = append(u.MysteryTraders, mysteryTrader)
	u.mysteryTradersByNum[mysteryTrader.Num] = mysteryTrader
	u.addMapObjectByPosition(mysteryTrader, mysteryTrader.Position)
}

// move a fleet from one position to another
func (u *Universe) moveMysteryTrader(mysteryTrader *MysteryTrader, originalPosition Vector) {
	// upadte mapobjects position
	u.updateMapObjectAtPosition(mysteryTrader, originalPosition, mysteryTrader.Position)
}

// mark a mysteryTrader as deleted and remove it from the universe
func (u *Universe) deleteMysteryTrader(mysteryTrader *MysteryTrader) {
	mysteryTrader.Delete = true

	delete(u.mysteryTradersByNum, mysteryTrader.Num)
	u.removeMapObjectAtPosition(mysteryTrader, mysteryTrader.Position)

	u.log.Debug().
		Int("Player", mysteryTrader.PlayerNum).
		Str("MysteryTrader", mysteryTrader.Name).
		Msgf("deleted mysteryTrader")

}
