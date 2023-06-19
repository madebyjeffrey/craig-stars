package server

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/sirgwain/craig-stars/game"
)

// Allow a user to update a planet's orders
func (s *server) UpdatePlanetOrders(c *gin.Context) {
	user := s.GetSessionUser(c)

	var id idBind
	if err := c.ShouldBindUri(&id); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// find the player for this user
	player, err := s.db.FindPlayerByGameIdLight(id.ID, user.ID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	planet := game.Planet{}
	if err := c.ShouldBindJSON(&planet); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// find the existing planet by id
	existing, err := s.db.FindPlanetByID(planet.ID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// verify the user actually owns this planet
	if existing.PlayerNum != player.Num {
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Errorf("%s does not own %s", player, existing)})
		return
	}

	rules, err := s.db.FindGameRulesByGameID(planet.GameID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if rules == nil {
		r := game.NewRules()
		rules = &r
	}

	// copy user modifiable things to the existing planet
	existing.ContributesOnlyLeftoverToResearch = planet.ContributesOnlyLeftoverToResearch
	existing.ProductionQueue = planet.ProductionQueue
	existing.Spec = game.ComputePlanetSpec(rules, existing, player)
	s.db.SavePlanet(existing.GameID, existing)

	c.JSON(http.StatusOK, existing)
}
