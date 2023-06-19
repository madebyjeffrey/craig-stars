package dbsqlx

import (
	"database/sql"
	"database/sql/driver"
	"encoding/json"
	"errors"
	"time"

	"github.com/sirgwain/craig-stars/game"
)

type Player struct {
	ID                           int64                  `json:"id,omitempty"`
	CreatedAt                    time.Time              `json:"createdAt,omitempty"`
	UpdatedAt                    time.Time              `json:"updatedAt,omitempty"`
	GameID                       int64                  `json:"gameId,omitempty"`
	UserID                       int64                  `json:"userId,omitempty"`
	Name                         string                 `json:"name,omitempty"`
	Num                          int                    `json:"num,omitempty"`
	Ready                        bool                   `json:"ready,omitempty"`
	AIControlled                 bool                   `json:"aiControlled,omitempty"`
	SubmittedTurn                bool                   `json:"submittedTurn,omitempty"`
	Color                        string                 `json:"color,omitempty"`
	DefaultHullSet               int                    `json:"defaultHullSet,omitempty"`
	TechLevelsEnergy             int                    `json:"techLevelsEnergy,omitempty"`
	TechLevelsWeapons            int                    `json:"techLevelsWeapons,omitempty"`
	TechLevelsPropulsion         int                    `json:"techLevelsPropulsion,omitempty"`
	TechLevelsConstruction       int                    `json:"techLevelsConstruction,omitempty"`
	TechLevelsElectronics        int                    `json:"techLevelsElectronics,omitempty"`
	TechLevelsBiotechnology      int                    `json:"techLevelsBiotechnology,omitempty"`
	TechLevelsSpentEnergy        int                    `json:"techLevelsSpentEnergy,omitempty"`
	TechLevelsSpentWeapons       int                    `json:"techLevelsSpentWeapons,omitempty"`
	TechLevelsSpentPropulsion    int                    `json:"techLevelsSpentPropulsion,omitempty"`
	TechLevelsSpentConstruction  int                    `json:"techLevelsSpentConstruction,omitempty"`
	TechLevelsSpentElectronics   int                    `json:"techLevelsSpentElectronics,omitempty"`
	TechLevelsSpentBiotechnology int                    `json:"techLevelsSpentBiotechnology,omitempty"`
	ResearchAmount               int                    `json:"researchAmount,omitempty"`
	ResearchSpentLastYear        int                    `json:"researchSpentLastYear,omitempty"`
	NextResearchField            game.NextResearchField `json:"nextResearchField,omitempty"`
	Researching                  game.TechField         `json:"researching,omitempty"`
	ProductionPlans              ProductionPlans        `json:"productionPlans,omitempty"`
	TransportPlans               TransportPlans         `json:"transportPlans,omitempty"`
	Stats                        *game.PlayerStats      `json:"stats,omitempty"`
	Spec                         *game.PlayerSpec       `json:"spec,omitempty"`
}

// we json serialize these types with custom Scan/Value methods
type ProductionPlans []game.ProductionPlan
type TransportPlans []game.TransportPlan
type PlayerSpec game.PlayerSpec
type PlayerStats game.PlayerStats

// db serializer to serialize this to JSON
func (item ProductionPlans) Value() (driver.Value, error) {
	if item == nil {
		return nil, nil
	}

	data, err := json.Marshal(item)
	if err != nil {
		return nil, err
	}
	return data, nil
}

// db deserializer to read this from JSON
func (item ProductionPlans) Scan(src interface{}) error {
	if src == nil {
		// leave empty
		return nil
	}

	switch v := src.(type) {
	case []byte:
		return json.Unmarshal(v, &item)
	case string:
		return json.Unmarshal([]byte(v), &item)
	}
	return errors.New("type assertion failed")
}

// db serializer to serialize this to JSON
func (item TransportPlans) Value() (driver.Value, error) {
	if item == nil {
		return nil, nil
	}

	data, err := json.Marshal(item)
	if err != nil {
		return nil, err
	}
	return data, nil
}

// db deserializer to read this from JSON
func (item TransportPlans) Scan(src interface{}) error {
	if src == nil {
		// leave empty
		return nil
	}

	switch v := src.(type) {
	case []byte:
		return json.Unmarshal(v, &item)
	case string:
		return json.Unmarshal([]byte(v), &item)
	}
	return errors.New("type assertion failed")
}

// db serializer to serialize this to JSON
func (item *PlayerSpec) Value() (driver.Value, error) {
	if item == nil {
		return nil, nil
	}

	data, err := json.Marshal(item)
	if err != nil {
		return nil, err
	}
	return data, nil
}

// db deserializer to read this from JSON
func (item *PlayerSpec) Scan(src interface{}) error {
	switch v := src.(type) {
	case []byte:
		return json.Unmarshal(v, item)
	case string:
		return json.Unmarshal([]byte(v), item)
	}
	return errors.New("type assertion failed")
}

// db serializer to serialize this to JSON
func (item *PlayerStats) Value() (driver.Value, error) {
	if item == nil {
		return nil, nil
	}

	data, err := json.Marshal(item)
	if err != nil {
		return nil, err
	}
	return data, nil
}

// db deserializer to read this from JSON
func (item *PlayerStats) Scan(src interface{}) error {
	switch v := src.(type) {
	case []byte:
		return json.Unmarshal(v, item)
	case string:
		return json.Unmarshal([]byte(v), item)
	}
	return errors.New("type assertion failed")
}

func (c *client) GetPlayers() ([]game.Player, error) {

	// don't include password in bulk select
	items := []Player{}
	if err := c.db.Select(&items, `SELECT * FROM players`); err != nil {
		if err == sql.ErrNoRows {
			return []game.Player{}, nil
		}
		return nil, err
	}

	return c.converter.ConvertPlayers(items), nil
}

func (c *client) GetPlayersForUser(userID int64) ([]game.Player, error) {

	// don't include password in bulk select
	items := []Player{}
	if err := c.db.Select(&items, `SELECT * FROM players WHERE userId = ?`, userID); err != nil {
		if err == sql.ErrNoRows {
			return []game.Player{}, nil
		}
		return nil, err
	}

	return c.converter.ConvertPlayers(items), nil
}

// get a player by id
func (c *client) GetPlayer(id int64) (*game.Player, error) {
	item := Player{}
	if err := c.db.Get(&item, "SELECT * FROM players WHERE id = ?", id); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	player := c.converter.ConvertPlayer(item)
	return &player, nil
}

// create a new game
func (c *client) CreatePlayer(player *game.Player) error {

	item := c.converter.ConvertGamePlayer(player)
	result, err := c.db.NamedExec(`
	INSERT INTO players (
		createdAt,
		updatedAt,
		gameId,
		userId,
		name,
		num,
		ready,
		aiControlled,
		submittedTurn,
		color,
		defaultHullSet,
		techLevelsEnergy,
		techLevelsWeapons,
		techLevelsPropulsion,
		techLevelsConstruction,
		techLevelsElectronics,
		techLevelsBiotechnology,
		techLevelsSpentEnergy,
		techLevelsSpentWeapons,
		techLevelsSpentPropulsion,
		techLevelsSpentConstruction,
		techLevelsSpentElectronics,
		techLevelsSpentBiotechnology,
		researchAmount,
		researchSpentLastYear,
		nextResearchField,
		researching,
		productionPlans,
		transportPlans,
		stats,
		spec
	)
	VALUES (
		CURRENT_TIMESTAMP,
		CURRENT_TIMESTAMP,
		:gameId,
		:userId,
		:name,
		:num,
		:ready,
		:aiControlled,
		:submittedTurn,
		:color,
		:defaultHullSet,
		:techLevelsEnergy,
		:techLevelsWeapons,
		:techLevelsPropulsion,
		:techLevelsConstruction,
		:techLevelsElectronics,
		:techLevelsBiotechnology,
		:techLevelsSpentEnergy,
		:techLevelsSpentWeapons,
		:techLevelsSpentPropulsion,
		:techLevelsSpentConstruction,
		:techLevelsSpentElectronics,
		:techLevelsSpentBiotechnology,
		:researchAmount,
		:researchSpentLastYear,
		:nextResearchField,
		:researching,
		:productionPlans,
		:transportPlans,
		:stats,
		:spec
	)
	`, item)

	if err != nil {
		return err
	}

	// update the id of our passed in game
	id, err := result.LastInsertId()
	if err != nil {
		return err
	}

	player.ID = int64(id)

	return nil
}

// update an existing player
func (c *client) UpdatePlayer(player *game.Player) error {

	item := c.converter.ConvertGamePlayer(player)

	if _, err := c.db.NamedExec(`
	UPDATE players SET
		updatedAt = CURRENT_TIMESTAMP,
		gameId = :gameId,
		userId = :userId,
		name = :name,
		num = :num,
		ready = :ready,
		aiControlled = :aiControlled,
		submittedTurn = :submittedTurn,
		color = :color,
		defaultHullSet = :defaultHullSet,
		techLevelsEnergy = :techLevelsEnergy,
		techLevelsWeapons = :techLevelsWeapons,
		techLevelsPropulsion = :techLevelsPropulsion,
		techLevelsConstruction = :techLevelsConstruction,
		techLevelsElectronics = :techLevelsElectronics,
		techLevelsBiotechnology = :techLevelsBiotechnology,
		techLevelsSpentEnergy = :techLevelsSpentEnergy,
		techLevelsSpentWeapons = :techLevelsSpentWeapons,
		techLevelsSpentPropulsion = :techLevelsSpentPropulsion,
		techLevelsSpentConstruction = :techLevelsSpentConstruction,
		techLevelsSpentElectronics = :techLevelsSpentElectronics,
		techLevelsSpentBiotechnology = :techLevelsSpentBiotechnology,
		researchAmount = :researchAmount,
		researchSpentLastYear = :researchSpentLastYear,
		nextResearchField = :nextResearchField,
		researching = :researching,
		productionPlans = :productionPlans,
		transportPlans = :transportPlans,
		stats = :stats,
		spec = :spec
	WHERE id = :id
	`, item); err != nil {
		return err
	}

	return nil
}

// delete a player by id
func (c *client) DeletePlayer(id int64) error {
	if _, err := c.db.Exec("DELETE FROM players WHERE id = ?", id); err != nil {
		return err
	}

	return nil
}
