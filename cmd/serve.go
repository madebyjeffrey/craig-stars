package cmd

import (
	"github.com/rs/zerolog/log"
	"github.com/sirgwain/craig-stars/appcontext"
	"github.com/sirgwain/craig-stars/game"
	"github.com/sirgwain/craig-stars/server"

	"github.com/spf13/cobra"
)

func newServeCmd() *cobra.Command {
	var generateUniverse bool
	serveCmd := &cobra.Command{
		Use:   "serve",
		Short: "Start the webserver",
		Long:  `Start a local gin-gonic webserver and serve requests.`,
		Run: func(cmd *cobra.Command, args []string) {
			ctx := appcontext.Initialize()
			ctx.DB.EnableDebugLogging()

			if generateUniverse {
				generateTestGame(ctx)
			}
			server.Start(ctx)
		},
	}
	serveCmd.Flags().BoolVar(&generateUniverse, "generate-test-game", false, "Generate a test user and game")

	return serveCmd
}

func generateTestGame(ctx *appcontext.AppContext) {
	ctx.DB.MigrateAll()

	techs := game.StaticTechStore
	if err := ctx.DB.CreateTechStore(&techs); err != nil {
		panic(err)
	}

	user, err := ctx.DB.FindUserById(1)
	if err != nil {
		log.Error().Err(err).Msg("Failed to load user for test game")
		panic(err)
	}
	if user == nil {
		user = game.NewUser("admin", "admin", game.RoleAdmin)
		err := ctx.DB.SaveUser(user)
		if err != nil {
			log.Error().Err(err).Msg("Failed to create new user for test game")
		}

	}

	races, err := ctx.DB.GetRaces(user.ID)
	if err != nil {
		panic(err)
	}

	var race *game.Race
	if len(races) == 0 {
		race = game.NewRace()
		race.UserID = user.ID

		if err := ctx.DB.CreateRace(race); err != nil {
			panic(err)
		}
	} else {
		race = &races[0]
	}

	g := game.NewGame()
	g.HostID = user.ID
	g.AddPlayer(game.NewPlayer(1, race))
	// g.Size = game.SizeSmall
	// g.Density = game.DensityNormal
	if err := ctx.DB.CreateGame(g); err != nil {
		panic(err)
	}

	g.GenerateUniverse()
	ctx.DB.SaveGame(g)
}

func init() {
	rootCmd.AddCommand(newServeCmd())
}
