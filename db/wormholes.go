package db

import (
	"database/sql"
	"database/sql/driver"
	"fmt"
	"time"

	"github.com/sirgwain/craig-stars/cs"
)

type Wormhole struct {
	ID               int64                `json:"id,omitempty"`
	GameID           int64                `json:"gameId,omitempty"`
	CreatedAt        time.Time            `json:"createdAt,omitempty"`
	UpdatedAt        time.Time            `json:"updatedAt,omitempty"`
	X                float64              `json:"x,omitempty"`
	Y                float64              `json:"y,omitempty"`
	Name             string               `json:"name,omitempty"`
	Num              int                  `json:"num,omitempty"`
	DestinationNum   int                  `json:"destinationNum,omitempty"`
	Stability        cs.WormholeStability `json:"stability,omitempty"`
	YearsAtStability int                  `json:"yearsAtStability,omitempty"`
	Spec             *WormholeSpec        `json:"spec,omitempty"`
}

// we json serialize these types with custom Scan/Value methods
type WormholeSpec cs.WormholeSpec

// db serializer to serialize this to JSON
func (item *WormholeSpec) Value() (driver.Value, error) {
	return valueJSON(item)
}

// db deserializer to read this from JSON
func (item *WormholeSpec) Scan(src interface{}) error {
	return scanJSON(src, item)
}

// get a wormhole by id
func (c *client) GetWormhole(id int64) (*cs.Wormhole, error) {
	item := Wormhole{}
	if err := c.db.Get(&item, "SELECT * FROM wormholes WHERE id = ?", id); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	wormhole := c.converter.ConvertWormhole(&item)
	return wormhole, nil
}

func (c *client) getWormholesForGame(db SQLSelector, gameID int64) ([]*cs.Wormhole, error) {

	items := []Wormhole{}
	if err := db.Select(&items, `SELECT * FROM wormholes WHERE gameId = ?`, gameID); err != nil {
		if err == sql.ErrNoRows {
			return []*cs.Wormhole{}, nil
		}
		return nil, err
	}

	results := make([]*cs.Wormhole, len(items))
	for i := range items {
		results[i] = c.converter.ConvertWormhole(&items[i])
	}

	return results, nil
}

// create a new game
func (c *client) createWormhole(wormhole *cs.Wormhole, tx SQLExecer) error {
	item := c.converter.ConvertGameWormhole(wormhole)
	result, err := tx.NamedExec(`
	INSERT INTO wormholes (
		createdAt,
		updatedAt,
		gameId,
		x,
		y,
		name,
		num,
		destinationNum,
		stability,
		yearsAtStability,
		spec
	)
	VALUES (
		CURRENT_TIMESTAMP,
		CURRENT_TIMESTAMP,
		:gameId,
		:x,
		:y,
		:name,
		:num,
		:destinationNum,
		:stability,
		:yearsAtStability,
		:spec
	)
	`, item)

	if err != nil {
		return err
	}

	// update the id of our passed in game
	wormhole.ID, err = result.LastInsertId()
	if err != nil {
		return err
	}

	return nil
}

func (c *client) UpdateWormhole(wormhole *cs.Wormhole) error {
	return c.updateWormhole(wormhole, c.db)
}

// update an existing wormhole
func (c *client) updateWormhole(wormhole *cs.Wormhole, tx SQLExecer) error {

	item := c.converter.ConvertGameWormhole(wormhole)

	if _, err := tx.NamedExec(`
	UPDATE wormholes SET
		updatedAt = CURRENT_TIMESTAMP,
		gameId = :gameId,
		x = :x,
		y = :y,
		name = :name,
		num = :num,
		destinationNum = :destinationNum,
		stability = :stability,
		yearsAtStability = :yearsAtStability,
		spec = :spec
	WHERE id = :id
	`, item); err != nil {
		return err
	}

	return nil
}

func (c *client) deleteWormhole(wormholeID int64, tx SQLExecer) error {
	if _, err := tx.Exec("DELETE FROM wormholes where id = ?", wormholeID); err != nil {
		return fmt.Errorf("delete wormhole %d %w", wormholeID, err)
	}
	return nil
}
