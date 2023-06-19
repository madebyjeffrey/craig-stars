package cs

import (
	"encoding/json"
	"io/ioutil"
	"testing"

	"github.com/stretchr/testify/assert"
)

func testStalwartDefender(player *Player) *Fleet {
	fleet := &Fleet{
		MapObject: MapObject{
			PlayerNum: player.Num,
		},
		BaseName: "Stalwart Defender",
		Tokens: []ShipToken{
			{
				DesignNum: 1,
				Quantity:  1,
				design: NewShipDesign(player, 1).
					WithHull(Destroyer.Name).
					WithSlots([]ShipDesignSlot{
						{HullComponent: LongHump6.Name, HullSlotIndex: 1, Quantity: 1},
						{HullComponent: BetaTorpedo.Name, HullSlotIndex: 2, Quantity: 1},
						{HullComponent: XRayLaser.Name, HullSlotIndex: 3, Quantity: 1},
						{HullComponent: RhinoScanner.Name, HullSlotIndex: 4, Quantity: 1},
						{HullComponent: Crobmnium.Name, HullSlotIndex: 5, Quantity: 1},
						{HullComponent: Overthruster.Name, HullSlotIndex: 6, Quantity: 1},
						{HullComponent: BattleComputer.Name, HullSlotIndex: 7, Quantity: 1},
					}).
					WithSpec(&rules, player)},
		},
		battlePlan:        &player.BattlePlans[0],
		OrbitingPlanetNum: NotOrbitingPlanet,
	}
	fleet.Spec = ComputeFleetSpec(&rules, player, fleet)
	fleet.Fuel = fleet.Spec.FuelCapacity
	return fleet
}

// create a new small freighter (with cargo pod) fleet for testing
func testTeamster(player *Player) *Fleet {
	fleet := &Fleet{
		MapObject: MapObject{
			PlayerNum: player.Num,
		},
		BaseName: "Teamster",
		Tokens: []ShipToken{
			{
				Quantity:  1,
				DesignNum: 1,
				design: NewShipDesign(player, 1).
					WithHull(SmallFreighter.Name).
					WithSlots([]ShipDesignSlot{
						{HullComponent: LongHump6.Name, HullSlotIndex: 1, Quantity: 1},
						{HullComponent: Crobmnium.Name, HullSlotIndex: 2, Quantity: 1},
						{HullComponent: RhinoScanner.Name, HullSlotIndex: 3, Quantity: 1},
					}).
					WithSpec(&rules, player)},
		},
		battlePlan:        &player.BattlePlans[0],
		OrbitingPlanetNum: NotOrbitingPlanet,
	}

	fleet.Spec = ComputeFleetSpec(&rules, player, fleet)
	fleet.Fuel = fleet.Spec.FuelCapacity
	return fleet

}

func Test_battle_regenerateShields(t *testing.T) {
	type args struct {
		player *Player
		token  battleToken
	}
	tests := []struct {
		name        string
		args        args
		wantShields int
	}{
		{
			name: "no regen",
			args: args{
				player: testPlayer().WithNum(1),
				token: battleToken{
					BattleRecordToken: BattleRecordToken{
						PlayerNum: 1,
					},
					stackShields:      50,
					totalStackShields: 100,
				},
			},
			wantShields: 50,
		},
		{
			name: "regen",
			args: args{
				player: NewPlayer(1, NewRace().WithLRT(RS).WithSpec(&rules)).WithNum(1),
				token: battleToken{
					BattleRecordToken: BattleRecordToken{
						PlayerNum: 1,
					},
					stackShields:      50,
					totalStackShields: 100,
				},
			},
			wantShields: 60,
		},
		{
			name: "no regen when shields gone",
			args: args{
				player: NewPlayer(1, NewRace().WithLRT(RS).WithSpec(&rules)).WithNum(1),
				token: battleToken{
					BattleRecordToken: BattleRecordToken{
						PlayerNum: 1,
					},
					stackShields:      0,
					totalStackShields: 100,
				},
			},
			wantShields: 0,
		},
	}
	for _, tt := range tests {

		t.Run(tt.name, func(t *testing.T) {
			battle := battle{
				players: map[int]*Player{tt.args.player.Num: tt.args.player},
			}
			battle.regenerateShields(&tt.args.token)

			got := tt.args.token.stackShields
			if got != tt.wantShields {
				t.Errorf("battle.regenerateShields() = %v, want %v", got, tt.wantShields)
			}

		})
	}
}

func Test_battle_willTarget(t *testing.T) {

	type args struct {
		target BattleTarget
		token  battleToken
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		// if our token has armed/starbase attributes, it should only target armed or starbases
		{args: args{BattleTargetAny, battleToken{}}, want: true},
		{args: args{BattleTargetStarbase, battleToken{attributes: battleTokenAttributeArmed | battleTokenAttributeStarbase}}, want: true},
		{args: args{BattleTargetArmedShips, battleToken{attributes: battleTokenAttributeArmed | battleTokenAttributeStarbase}}, want: true},
		{args: args{BattleTargetNone, battleToken{attributes: battleTokenAttributeArmed | battleTokenAttributeStarbase}}, want: false},
		{args: args{BattleTargetBombersFreighters, battleToken{attributes: battleTokenAttributeArmed | battleTokenAttributeStarbase}}, want: false},
		{args: args{BattleTargetUnarmedShips, battleToken{attributes: battleTokenAttributeArmed | battleTokenAttributeStarbase}}, want: false},
		{args: args{BattleTargetFuelTransports, battleToken{attributes: battleTokenAttributeArmed | battleTokenAttributeStarbase}}, want: false},
		{args: args{BattleTargetFreighters, battleToken{attributes: battleTokenAttributeArmed | battleTokenAttributeStarbase}}, want: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := &battle{}
			if got := b.willTarget(tt.args.target, &tt.args.token); got != tt.want {
				t.Errorf("battle.willTarget() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_battleToken_getDistanceAway(t *testing.T) {

	type args struct {
		position Vector
	}
	tests := []struct {
		name string
		bt   battleToken
		args args
		want int
	}{
		{"no distance", battleToken{BattleRecordToken: BattleRecordToken{Position: Vector{0, 0}}}, args{Vector{0, 0}}, 0},
		{"x distance greatest", battleToken{BattleRecordToken: BattleRecordToken{Position: Vector{2, 1}}}, args{Vector{4, 2}}, 2},
		{"y distance greatest", battleToken{BattleRecordToken: BattleRecordToken{Position: Vector{1, 2}}}, args{Vector{2, 5}}, 3},
		{"negative distance (token behind)", battleToken{BattleRecordToken: BattleRecordToken{Position: Vector{1, 1}}}, args{Vector{0, 0}}, 1},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.bt.getDistanceAway(tt.args.position); got != tt.want {
				t.Errorf("battleToken.getDistanceAway() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_battleWeaponSlot_isInRangePosition(t *testing.T) {
	type args struct {
		position Vector
	}
	tests := []struct {
		name   string
		weapon battleWeaponSlot
		args   args
		want   bool
	}{
		{"no distance, in range", battleWeaponSlot{token: &battleToken{BattleRecordToken: BattleRecordToken{Position: Vector{0, 0}}}}, args{Vector{0, 0}}, true},
		{"distance 1, in range", battleWeaponSlot{token: &battleToken{BattleRecordToken: BattleRecordToken{Position: Vector{0, 0}}}, weaponRange: 1}, args{Vector{1, 1}}, true},
		{"distance 2, out of range", battleWeaponSlot{token: &battleToken{BattleRecordToken: BattleRecordToken{Position: Vector{0, 0}}}, weaponRange: 1}, args{Vector{1, 2}}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.weapon.isInRangePosition(tt.args.position); got != tt.want {
				t.Errorf("battleWeaponSlot.isInRangePosition() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_battle_fireBeamWeapon(t *testing.T) {

	type weapon struct {
		weaponSlot   *battleWeaponSlot
		shipQuantity int
		position     Vector
	}
	type args struct {
		weapon  weapon
		targets []*battleToken
	}
	type want struct {
		damage            float64
		quantityDamaged   int
		quantityRemaining int
		stackShields      int
	}
	tests := []struct {
		name string
		args args
		want []want
	}{
		{name: "Single weapon, do 10 damage, no kills",
			args: args{
				weapon: weapon{
					weaponSlot: &battleWeaponSlot{
						slot: ShipDesignSlot{
							Quantity: 1, // 1 beam weapon
						},
						power:       10,
						weaponRange: 1,
					},
					shipQuantity: 1,
				},
				targets: []*battleToken{
					{
						ShipToken: &ShipToken{
							Quantity: 1,
							design:   &ShipDesign{Name: "defender"}, // for logging
						},
						armor: 20,
					},
				},
			},
			want: []want{{damage: 10, quantityDamaged: 1, quantityRemaining: 1}},
		},
		{name: "Single weapon, do 30 damage, to a ship stack with two ships, one damaged",
			args: args{
				weapon: weapon{
					weaponSlot: &battleWeaponSlot{
						slot: ShipDesignSlot{
							Quantity: 1, // 1 beam weapon
						},
						power:       30,
						weaponRange: 1,
					},
					shipQuantity: 1,
				},
				targets: []*battleToken{
					{
						ShipToken: &ShipToken{
							Quantity:        2,
							Damage:          5,
							QuantityDamaged: 1,
							design:          &ShipDesign{Name: "defender"}, // for logging
						},
						armor: 20,
					},
				},
			},
			want: []want{{damage: 15, quantityDamaged: 1, quantityRemaining: 1}},
		},
		{name: "Single weapon, do 10 damage reduced to 9 for range",
			args: args{
				weapon: weapon{
					weaponSlot: &battleWeaponSlot{
						slot: ShipDesignSlot{
							Quantity: 1, // 1 beam weapon
						},
						power:       10,
						weaponRange: 2,
					},
					shipQuantity: 1,
					position:     Vector{2, 0}, // 1 away from target
				},
				targets: []*battleToken{
					{
						ShipToken: &ShipToken{
							Quantity: 1,
							design:   &ShipDesign{Name: "defender"}, // for logging
						},
						BattleRecordToken: BattleRecordToken{
							Position: Vector{0, 0},
						},
						armor: 20,
					},
				},
			},
			want: []want{{damage: 9, quantityDamaged: 1, quantityRemaining: 1}},
		},
		{name: "two weapons, do 30 damage total, one (over)kill",
			args: args{
				weapon: weapon{
					weaponSlot: &battleWeaponSlot{
						slot: ShipDesignSlot{
							Quantity: 2, // 2 beam weapons
						},
						power:       15, // 10 damage per beam
						weaponRange: 2,
					},
					shipQuantity: 1, // one ship in the attacker stack
				},
				targets: []*battleToken{
					{
						ShipToken: &ShipToken{
							Quantity: 1,
							design:   &ShipDesign{Name: "defender"},
						},
						armor: 20, // 20 armor, will be destroyed
					},
				},
			},
			want: []want{{damage: 0, quantityDamaged: 0, quantityRemaining: 0}},
		},
		{name: "two weapons, two ships, do 40 damage total, one kill, one damaged",
			args: args{
				weapon: weapon{
					weaponSlot: &battleWeaponSlot{
						slot: ShipDesignSlot{
							Quantity: 2, // 2 beam weapons
						},
						power: 10,
					},
					shipQuantity: 2, // 2 ships in attacker stack
				},
				targets: []*battleToken{
					{
						ShipToken: &ShipToken{
							Quantity: 2, // two ships in defender
							design:   &ShipDesign{Name: "defender"},
						},
						armor: 30,
					},
				},
			},
			want: []want{{damage: 10, quantityDamaged: 1, quantityRemaining: 1}},
		},
		{name: "one weapon, do 10 damage to shields, no damage",
			args: args{
				weapon: weapon{
					weaponSlot: &battleWeaponSlot{
						slot: ShipDesignSlot{
							Quantity: 1,
						},
						power: 10,
					},
					shipQuantity: 1,
				},
				targets: []*battleToken{
					{
						ShipToken: &ShipToken{
							Quantity: 1,
							design:   &ShipDesign{Name: "defender"},
						},
						armor:        30,
						shields:      20,
						stackShields: 20,
					},
				},
			},
			want: []want{{damage: 0, quantityDamaged: 0, quantityRemaining: 1, stackShields: 10}},
		},
		{name: "one super beam, do 100 damage destroy one stack and damage another",
			args: args{
				weapon: weapon{
					weaponSlot: &battleWeaponSlot{
						slot: ShipDesignSlot{
							Quantity: 1,
						},
						power: 100,
					},
					shipQuantity: 1,
				},
				targets: []*battleToken{
					{
						ShipToken: &ShipToken{
							Quantity: 1,
							design:   &ShipDesign{Name: "defender1"},
						},
						armor: 10,
					},
					{
						ShipToken: &ShipToken{
							Quantity: 1,
							design:   &ShipDesign{Name: "defender2"},
						},
						armor: 100,
					},
				},
			},
			want: []want{
				{damage: 0, quantityDamaged: 0, quantityRemaining: 0},
				{damage: 90, quantityDamaged: 1, quantityRemaining: 1},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := &battle{rules: &rules,
				record: newBattleRecord(1, NotOrbitingPlanet, Vector{}, []BattleRecordToken{})}
			b.record.RecordNewRound()

			// setup this weapon's token based on shipQuantity and position
			tt.args.weapon.weaponSlot.token = &battleToken{
				ShipToken: &ShipToken{
					Quantity: tt.args.weapon.shipQuantity,
					design:   &ShipDesign{Name: "attacker"}, // for logging
				},
				BattleRecordToken: BattleRecordToken{
					Position: tt.args.weapon.position,
				},
			}

			// fire the beam weapon!
			b.fireBeamWeapon(tt.args.weapon.weaponSlot, tt.args.targets)

			for i, target := range tt.args.targets {
				if target.Quantity != tt.want[i].quantityRemaining {
					t.Errorf("battleWeaponSlot.fireBeamWeapon() target: %d quantityRemaining = %v, want %v", i, target.Quantity, tt.want[i].quantityRemaining)
				}
				if target.Damage != tt.want[i].damage {
					t.Errorf("battleWeaponSlot.fireBeamWeapon() target: %d damage = %v, want %v", i, target.Damage, tt.want[i].damage)
				}
				if target.QuantityDamaged != tt.want[i].quantityDamaged {
					t.Errorf("battleWeaponSlot.fireBeamWeapon() target: %d quantityDamaged = %v, want %v", i, target.QuantityDamaged, tt.want[i].quantityDamaged)
				}
				if target.stackShields != tt.want[i].stackShields {
					t.Errorf("battleWeaponSlot.fireBeamWeapon() target: %d stackShields = %v, want %v", i, target.stackShields, tt.want[i].stackShields)
				}
			}

		})
	}
}

func Test_battle_fireTorpedo(t *testing.T) {

	type weapon struct {
		weaponSlot   *battleWeaponSlot
		shipQuantity int
		position     Vector
	}
	type args struct {
		weapon  weapon
		targets []*battleToken
	}
	type want struct {
		damage            float64
		quantityDamaged   int
		quantityRemaining int
		stackShields      int
	}
	tests := []struct {
		name string
		args args
		want []want
	}{
		{name: "Single torpedo, do 10 damage, no kills",
			args: args{
				weapon: weapon{
					weaponSlot: &battleWeaponSlot{
						slot: ShipDesignSlot{
							Quantity: 1, // 1 torpedo
						},
						power:    10,
						accuracy: 1,
					},
					shipQuantity: 1,
				},
				targets: []*battleToken{
					{
						ShipToken: &ShipToken{
							Quantity: 1,
							design:   &ShipDesign{Name: "defender"}, // for logging
						},
						armor: 20,
					},
				},
			},
			want: []want{{damage: 10, quantityDamaged: 1, quantityRemaining: 1}},
		},
		{name: "Single torpedo, do 10 damage to a 2 ship stack with 1@5 damage",
			args: args{
				weapon: weapon{
					weaponSlot: &battleWeaponSlot{
						slot: ShipDesignSlot{
							Quantity: 1, // 1 torpedo
						},
						power:    10,
						accuracy: 1,
					},
					shipQuantity: 1,
				},
				targets: []*battleToken{
					{
						ShipToken: &ShipToken{
							Quantity:        2,
							QuantityDamaged: 1,
							Damage:          5,
							design:          &ShipDesign{Name: "defender"}, // for logging
						},
						armor: 20,
					},
				},
			},
			// TODO: not sure about this. It doesn't make sense for a torpedo to splash damage at the end...
			want: []want{{damage: 15 / 2., quantityDamaged: 2, quantityRemaining: 2}},
		},
		{name: "Single torpedo, do 30 damage to a stack with two ships, destroy one, other undamaged",
			args: args{
				weapon: weapon{
					weaponSlot: &battleWeaponSlot{
						slot: ShipDesignSlot{
							Quantity: 1, // 1 torpedo
						},
						power:    30,
						accuracy: 1,
					},
					shipQuantity: 1,
				},
				targets: []*battleToken{
					{
						ShipToken: &ShipToken{
							Quantity: 2,
							design:   &ShipDesign{Name: "defender"}, // for logging
						},
						armor: 20,
					},
				},
			},
			want: []want{{damage: 0, quantityDamaged: 0, quantityRemaining: 1}},
		},

		{name: "two torpedos, do 15 damage each, kill ship with first hit",
			args: args{
				weapon: weapon{
					weaponSlot: &battleWeaponSlot{
						slot: ShipDesignSlot{
							Quantity: 2, // 2 torpedos
						},
						power:    15, // 15 damage per torpedo
						accuracy: 1,
					},
					shipQuantity: 1, // one ship in the attacker stack
				},
				targets: []*battleToken{
					{
						ShipToken: &ShipToken{
							Quantity: 1,
							design:   &ShipDesign{Name: "defender"},
						},
						armor: 10, // 10 armor, will be destroyed
					},
				},
			},
			want: []want{{damage: 0, quantityDamaged: 0, quantityRemaining: 0}},
		},
		{name: "two torpedos, two attacker ships, 4x torpedos do 40 damage total, one kill, one damaged",
			args: args{
				weapon: weapon{
					weaponSlot: &battleWeaponSlot{
						slot: ShipDesignSlot{
							Quantity: 2, // 2 torpedos
						},
						power:    10,
						accuracy: 1,
					},
					shipQuantity: 2, // 2 ships in attacker stack
				},
				targets: []*battleToken{
					{
						ShipToken: &ShipToken{
							Quantity: 2, // two ships in defender
							design:   &ShipDesign{Name: "defender"},
						},
						armor: 30,
					},
				},
			},
			want: []want{{damage: 10, quantityDamaged: 1, quantityRemaining: 1}},
		},
		{name: "one torpedo, do 5 damage to shields, 5 damage to hull",
			args: args{
				weapon: weapon{
					weaponSlot: &battleWeaponSlot{
						slot: ShipDesignSlot{
							Quantity: 1,
						},
						power:    10,
						accuracy: 1,
					},
					shipQuantity: 1,
				},
				targets: []*battleToken{
					{
						ShipToken: &ShipToken{
							Quantity: 1,
							design:   &ShipDesign{Name: "defender"},
						},
						armor:        20,
						shields:      20,
						stackShields: 20,
					},
				},
			},
			want: []want{{damage: 5, quantityDamaged: 1, quantityRemaining: 1, stackShields: 15}},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := &battle{rules: &rules,
				record: newBattleRecord(1, NotOrbitingPlanet, Vector{}, []BattleRecordToken{})}
			b.record.RecordNewRound()

			// setup this weapon's token based on shipQuantity and position
			tt.args.weapon.weaponSlot.token = &battleToken{
				ShipToken: &ShipToken{
					Quantity: tt.args.weapon.shipQuantity,
					design:   &ShipDesign{Name: "attacker"}, // for logging
				},
				BattleRecordToken: BattleRecordToken{
					Position: tt.args.weapon.position,
				},
			}

			// fire the beam weapon!
			b.fireTorpedo(tt.args.weapon.weaponSlot, tt.args.targets)

			for i, target := range tt.args.targets {
				if target.Quantity != tt.want[i].quantityRemaining {
					t.Errorf("battleWeaponSlot.fireTorpedo() target: %d quantityRemaining = %v, want %v", i, target.Quantity, tt.want[i].quantityRemaining)
				}
				if target.Damage != tt.want[i].damage {
					t.Errorf("battleWeaponSlot.fireTorpedo() target: %d damage = %v, want %v", i, target.Damage, tt.want[i].damage)
				}
				if target.QuantityDamaged != tt.want[i].quantityDamaged {
					t.Errorf("battleWeaponSlot.fireTorpedo() target: %d quantityDamaged = %v, want %v", i, target.QuantityDamaged, tt.want[i].quantityDamaged)
				}
				if target.stackShields != tt.want[i].stackShields {
					t.Errorf("battleWeaponSlot.fireTorpedo() target: %d stackShields = %v, want %v", i, target.stackShields, tt.want[i].stackShields)
				}
			}

		})
	}
}

func Test_battle_runBattle1(t *testing.T) {
	player1 := testPlayer().WithNum(1)
	player2 := testPlayer().WithNum(2)
	player1.Relations = []PlayerRelationship{{Relation: PlayerRelationFriend}, {Relation: PlayerRelationEnemy}}
	player2.Relations = []PlayerRelationship{{Relation: PlayerRelationEnemy}, {Relation: PlayerRelationFriend}}

	fleets := []*Fleet{
		testStalwartDefender(player1),
		testLongRangeScout(player1),
		testTeamster(player2),
	}

	designNum := 1
	for _, fleet := range fleets {
		for _, token := range fleet.Tokens {
			token.design.Num = designNum
			designNum += 1
		}
	}

	battle := newBattler(&rules, &StaticTechStore, 1, map[int]*Player{1: player1, 2: player2}, fleets, nil)

	record := battle.runBattle()

	// ran some number of turns
	assert.Greater(t, len(record.ActionsPerRound), 1)

	gotJson, _ := json.MarshalIndent(record, "", "  ")
	_ = ioutil.WriteFile("../tmp/battle.json", gotJson, 0644)
}
