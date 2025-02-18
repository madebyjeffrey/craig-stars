package cs

import (
	"fmt"
	"math"
)

// Starbases with Packet Throwers can build mineral packets and fling them at other planets.
type MineralPacket struct {
	MapObject
	TargetPlanetNum   int    `json:"targetPlanetNum"`
	Cargo             Cargo  `json:"cargo,omitempty"`
	WarpSpeed         int    `json:"warpSpeed"`
	SafeWarpSpeed     int    `json:"safeWarpSpeed,omitempty"`
	Heading           Vector `json:"heading"`
	ScanRange         int    `json:"scanRange"`
	ScanRangePen      int    `json:"scanRangePen"`
	distanceTravelled float64
	builtThisTurn     bool
}

type MineralPacketDamage struct {
	Killed            int `json:"killed,omitempty"`
	DefensesDestroyed int `json:"defensesDestroyed,omitempty"`
	Uncaught          int `json:"uncaught,omitempty"`
}

// this mineral packet will decay to nothing before reaching its target
const MineralPacketDecayToNothing = -1

func newMineralPacket(player *Player, num int, warpSpeed int, safeWarpSpeed int, cargo Cargo, position Vector, targetPlanetNum int) *MineralPacket {
	packet := MineralPacket{
		MapObject: MapObject{
			Type:      MapObjectTypeMineralPacket,
			PlayerNum: player.Num,
			Num:       num,
			Name:      fmt.Sprintf("%s Mineral Packet", player.Race.PluralName),
			Position:  position,
		},
		WarpSpeed:       warpSpeed,
		SafeWarpSpeed:   safeWarpSpeed,
		Cargo:           cargo,
		TargetPlanetNum: targetPlanetNum,
		ScanRange:       NoScanner,
		ScanRangePen:    NoScanner,
	}

	// PP packets have built in scanners
	if player.Race.Spec.PacketBuiltInScanner {
		packet.ScanRangePen = warpSpeed * warpSpeed
	}

	return &packet
}

// get the rate of decay for a packet between 0 and 1
//
// Depending on how fast a packet is thrown compared to its safe speed, it decays
// Source: https://wiki.starsautohost.org/wiki/%22Mass_Packet_FAQ%22_by_Barry_Kearns_1997-02-07_v2.6b
func (packet *MineralPacket) getPacketDecayRate(rules *Rules, race *Race) float64 {

	// we only care about packets thrown up to 3 warps over the limit
	overSafeWarp := MinInt(packet.WarpSpeed-packet.SafeWarpSpeed, 3)

	// IT is always counted as being 1 more over the safe warp
	overSafeWarp = MinInt(race.Spec.PacketOverSafeWarpPenalty+overSafeWarp, 3)

	packetDecayRate := 0.0
	if overSafeWarp > 0 {
		packetDecayRate = rules.PacketDecayRate[overSafeWarp]
	}

	// PP have half the decay rate
	packetDecayRate *= race.Spec.PacketDecayFactor

	return packetDecayRate
}

// move this packet through space
func (packet *MineralPacket) movePacket(rules *Rules, player *Player, target *Planet, planetPlayer *Player) {
	dist := float64(packet.WarpSpeed * packet.WarpSpeed)
	totalDist := packet.Position.DistanceTo(target.Position)

	// move at half distance if this packet was created this turn
	if packet.builtThisTurn {
		dist /= 2
	}

	// round up, if we are <1 LY away, i.e. the target is 81.9 ly away, warp 9 (81 ly travel) should be able to make it there
	if dist < totalDist && totalDist-dist < 1 {
		dist = math.Ceil(totalDist)
	}

	vectorTravelled := target.Position.Subtract(packet.Position).Normalized().Scale(dist)
	dist = vectorTravelled.Length()

	// don't overshoot
	dist = math.Min(totalDist, dist)

	if totalDist == dist {
		packet.completeMove(rules, player, target, planetPlayer)
	} else {
		// move this packet closer to the next planet
		packet.distanceTravelled = dist
		packet.Heading = target.Position.Subtract(packet.Position).Normalized()
		packet.Position = packet.Position.Add(packet.Heading.Scale(dist))
		packet.Position = packet.Position.Round()
	}
}

// Damage calcs as per the Stars! Manual
//
// Example:
// You fling a 1000kT packet at Warp 10 at a planet with a Warp 5 driver, a population of 250,000 and 50 defenses preventing 60% of incoming damage.
// spdPacket = 100
// spdReceiver = 25
// %CaughtSafely = 25%
// minerals recovered = 1000kT x 25% + 1000kT x 75% x 1/3 = 250 + 250 = 500kT
// dmgRaw = 75 x 1000 / 160 = 469
// dmgRaw2 = 469 x 40% = 188
// #colonists killed = Max. of ( 188 x 250,000 / 1000, 188 x 100)
// = Max. of ( 47,000, 18800) = 47,000 colonists
// #defenses destroyed = 50 * 188 / 1000 = 9 (rounded down)
//
// If, however, the receiving planet had no mass driver or defenses, the damage is far greater:
// minerals recovered = 1000kT x 0% + 1000kT x 100% x 1/3 = only 333kT dmgRaw = 100 x 1000 / 160 = 625
// dmgRaw2 = 625 x 100% = 625
// #colonists killed = Max. of (625 x 250,000 / 1000, 625 x 100)
// = Max.of(156,250, 62500) = 156,250.
// If the packet increased speed up to Warp 13, then:
// dmgRaw2 = dmgRaw = 169 x 1000 / 160 = 1056
// #colonists killed = Max. of (1056 x 250,000 / 1000, 1056 x 100)
// = Max. of(264,000, 105600) destroying the colony

// Complete movement of an incoming packet about to impact the planet
func (packet *MineralPacket) completeMove(rules *Rules, player *Player, planet *Planet, planetPlayer *Player) {
	damage := packet.getDamage(planet, planetPlayer)

	if damage == (MineralPacketDamage{}) {
		// caught packet successfully, transfer cargo
		messager.planetPacketCaught(planetPlayer, planet, packet)
	} else if planetPlayer != nil {
		// kill off colonists and defenses
		// note, for AR races, this will be 0 colonists killed or structures destroyed
		planet.setPopulation(roundToNearest100(Clamp(planet.population()-damage.Killed, 0, planet.population())))
		planet.Defenses = Clamp(planet.Defenses-damage.DefensesDestroyed, 0, planet.Defenses)

		messager.planetPacketDamage(planetPlayer, planet, packet, damage.Killed, damage.DefensesDestroyed)
		if planet.population() == 0 {
			planet.emptyPlanet()
			messager.planetDiedOff(planetPlayer, planet)
		}
	}

	mineralsRecovered := 1.0
	if damage.Uncaught > 0 {
		var percentCaughtSafely float64
		if planet.Spec.HasStarbase && planet.Spec.SafePacketSpeed > 0 {
			percentCaughtSafely = float64((packet.WarpSpeed * packet.WarpSpeed) / (planet.Spec.SafePacketSpeed * planet.Spec.SafePacketSpeed))
		}
		packet.checkTerraform(rules, player, planet, 1-percentCaughtSafely)
		packet.checkPermaform(rules, player, planet, 1-percentCaughtSafely)

		// only 1/3 of uncaught minerals will be recovered
		mineralsRecovered = percentCaughtSafely + (1-percentCaughtSafely)/3
	}

	// one way or another, these minerals are ending up on the planet
	planet.Cargo = planet.Cargo.Add(packet.Cargo.Multiply(mineralsRecovered))

	// if we didn't receive this planet, notify the sender
	if planet.PlayerNum != packet.PlayerNum {
		if player.Race.Spec.DetectPacketDestinationStarbases && planet.Spec.HasStarbase {
			// discover the receiving planet's starbase design
			player.discoverer.discoverDesign(planet.Starbase.Tokens[0].design, true)
		}

		messager.planetPacketArrived(player, planet, packet)
	}

	// delete the packet
	packet.Delete = true
}

// get the damage a mineral packet will do when it collides with a planet
func (packet *MineralPacket) getDamage(planet *Planet, planetPlayer *Player) MineralPacketDamage {
	if !planet.Owned() {
		// unowned planets aren't damaged, but all cargo is uncaught
		return MineralPacketDamage{Uncaught: packet.Cargo.Total()}
	}

	if planet.Spec.HasMassDriver && planet.Spec.SafePacketSpeed >= packet.WarpSpeed {
		// planet successfully caught this packet
		return MineralPacketDamage{}
	}

	if planetPlayer != nil && planetPlayer.Race.Spec.LivesOnStarbases {
		// No damage, but all cargo is uncaught and might impact the planet
		return MineralPacketDamage{Uncaught: packet.Cargo.Total()}
	}

	// uh oh, this packet is going too fast and we'll take damage
	receiverDriverSpeed := 0
	if planet.Spec.HasStarbase {
		receiverDriverSpeed = planet.Spec.SafePacketSpeed
	}

	weight := packet.Cargo.Total()
	speedOfPacket := packet.WarpSpeed * packet.WarpSpeed
	speedOfReceiver := receiverDriverSpeed * receiverDriverSpeed
	percentCaughtSafely := float64(speedOfReceiver) / float64(speedOfPacket)
	uncaught := int((1.0 - percentCaughtSafely) * float64(weight))
	rawDamage := float64((speedOfPacket-speedOfReceiver)*weight) / 160
	damageWithDefenses := rawDamage * (1 - planet.Spec.DefenseCoverage)
	colonistsKilled := roundToNearest100f(math.Max(damageWithDefenses*float64(planet.population())/1000, damageWithDefenses*100))
	defensesDestroyed := int(math.Max(float64(planet.Defenses)*damageWithDefenses/1000, damageWithDefenses/20))

	// kill off colonists and defenses
	return MineralPacketDamage{
		Killed:            roundToNearest100(MinInt(colonistsKilled, planet.population())),
		DefensesDestroyed: MinInt(planet.Defenses, defensesDestroyed),
		Uncaught:          uncaught,
	}

}

// Estimate potential damage of incoming mineral packet
// Simulates decay each turn until impact
func (packet *MineralPacket) estimateDamage(rules *Rules, player *Player, target *Planet, planetPlayer *Player) MineralPacketDamage {
	if target.Spec.HasMassDriver && target.Spec.SafePacketSpeed >= packet.WarpSpeed {
		// planet successfully caught this packet
		return MineralPacketDamage{}
	}
	spd := float64(packet.WarpSpeed * packet.WarpSpeed)
	decayRate := 0.0
	totalDist := packet.Position.DistanceTo(target.Position)
	eta := int(math.Ceil(totalDist / spd))

	// save copy of packet so we don't alter the original
	packetCopy := *packet

	for i := 0; i < eta; i++ {
		if totalDist <= spd {
			// 1 turn until impact - only travels/decays partially
			distTraveled := totalDist / float64(spd)
			decayRate = (packetCopy.getPacketDecayRate(rules, &player.Race) * distTraveled)
		} else {
			decayRate = packetCopy.getPacketDecayRate(rules, &player.Race)
			totalDist -= spd
		}

		// no decay, so we don't need to bother calculating decay amount
		if decayRate == 0 {
			break
		}

		// loop through all 3 mineral types and reduce each one in turn
		for _, minType := range [3]CargoType{Ironium, Boranium, Germanium} {
			mineral := packetCopy.Cargo.GetAmount(minType)

			// subtract either the normal or minimum decay amounts, whichever is higher (rounded DOWN)
			if mineral > 0 {
				decayAmount := MaxInt(int(decayRate*float64(mineral)), int(float64(rules.PacketMinDecay)*float64(player.Race.Spec.PacketDecayFactor)))
				packetCopy.Cargo.SubtractAmount(minType, decayAmount)
				packetCopy.Cargo = packetCopy.Cargo.MinZero()
			}
		}

		// packet out of minerals; return special exit code
		if packetCopy.Cargo.Total() == 0 {
			return MineralPacketDamage{Uncaught: MineralPacketDecayToNothing}
		}
	}

	damage := packetCopy.getDamage(target, planetPlayer)

	// clear packet uncaught statistic as we don't care about it (this is a **damage** test function after all)
	damage.Uncaught = 0

	return damage
}

// Check if an uncaught PP packet will terraform the target planet's environment (50% chance/100kT)
func (packet *MineralPacket) checkTerraform(rules *Rules, player *Player, planet *Planet, uncaught float64) {
	if player.Race.Spec.PacketTerraformChance > 0 {
		terraformer := NewTerraformer()
		t := terraform{}

		// Evaluate each mineral type separately
		for i, minType := range [3]CargoType{Ironium, Boranium, Germanium} {
			mineral := int(math.Ceil(float64(packet.Cargo.GetAmount(minType)) * uncaught))
			habType := HabType(i)
			direction := 0
			result := TerraformResult{}

			// no mineral = no terraforming
			if mineral == 0 {
				continue
			}

			// Loop through the mineral amount and perform terraform checks repeatedly
			for uncaughtCheck := 0; uncaughtCheck < mineral; uncaughtCheck += player.Race.Spec.PacketPermaTerraformSizeUnit {

				// if packet has less minerals remaining than 1 check size unit, reduce the packet chance accordingly
				// this ensures that smaller packets can still terraform planets (albeit at a proportionally reduced rate)
				terraformChance := player.Race.Spec.PacketTerraformChance * math.Min(
					float64(mineral-uncaughtCheck)/float64(player.Race.Spec.PacketPermaTerraformSizeUnit), 1)

				// Example math: 250kT packet with 50% chance per 100kT
				// Loop 1 has chance 0.5 * min((250-0)/100, 1) = 0.5 * min(2.5, 1) = 0.5
				// Loop 2 has chance 0.5 * min((250-100)/100, 1) = 0.5 * min(1.5, 1) = 0.5
				// Loop 3 has chance 0.5 * min((250-200)/100, 1) = 0.5 * min(0.5, 1) = 0.25
				// Loop 4 fails to execute as uncaughtCheck (300) is now larger than mineral (250)

				if rules.random.Float64() <= terraformChance {
					if AbsInt(direction) >= t.getTerraformAbility(player).Get(habType) {
						// if we can't terraform hab any further, skip any remaining checks for brevity
						// TerraformHab already caps the result at the player's terraforming ability anyways; this just saves computing power
						continue
					} else if planet.Hab.Get(habType)+direction < player.Race.HabCenter().Get(habType) {
						// planet hab below ideal; need to raise it
						direction += 1
					} else if planet.Hab.Get(habType)+direction > player.Race.HabCenter().Get(habType) {
						// planet hab above ideal; need to lower it
						direction -= 1
					} else {
						// planet hab already ideal; no further changes needed
						continue
					}
				}
			}

			// Terraform & keep track of result (for messages)
			result = terraformer.TerraformHab(planet, player, habType, direction)
			if result.Terraformed() {
				messager.planetPacketTerraform(player, planet, result.Type, direction)
			}
		}
	}
}

// Check if an uncaught PP packet will permanently alter the target planet's environment (0.1% chance/100kT)
func (packet *MineralPacket) checkPermaform(rules *Rules, player *Player, planet *Planet, uncaught float64) {
	if player.Race.Spec.PacketPermaformChance > 0 && player.Race.Spec.PacketPermaTerraformSizeUnit > 0 {
		terraformer := NewTerraformer()

		// Evaluate each mineral type separately
		for i, minType := range [3]CargoType{Ironium, Boranium, Germanium} {
			mineral := int(math.Ceil(float64(packet.Cargo.GetAmount(minType)) * uncaught))
			habType := HabType(i)
			var result TerraformResult
			direction := 0

			// no mineral = no calcs needed
			if mineral == 0 {
				continue
			}

			// Loop through the mineral amount and perform terraform checks repeatedly
			for uncaughtCheck := 0; uncaughtCheck < mineral; uncaughtCheck += player.Race.Spec.PacketPermaTerraformSizeUnit {

				// if packet has less minerals remaining than 1 check size unit, reduce the packet chance accordingly
				// this ensures that smaller packets can still terraform planets (albeit at a proportionally reduced rate)
				permaformChance := player.Race.Spec.PacketPermaformChance * math.Min(
					float64((mineral-uncaughtCheck)/player.Race.Spec.PacketPermaTerraformSizeUnit), 1)

				if permaformChance >= float64(rules.random.Float64()) {
					// Permaform & keep track of result
					result = terraformer.PermaformOneStep(planet, player, habType)
					direction += result.Direction
					if !result.Terraformed() {
						// BaseHab already perfect; skip remaining checks
						continue
					}
				}
			}

			if result.Terraformed() {
				messager.planetPacketPermaform(player, planet, habType, direction)
			}
		}
	}
}
