package schooldiscord

import (
	"context"
	"fmt"

	"github.com/Petrify/simp-core/commands"
	"github.com/Petrify/simp-core/service"
	simpsql "github.com/Petrify/simp-core/sql"
	"github.com/bwmarrin/discordgo"
)

type guild struct {
	cmds *commands.Interpreter

	//settings
	cmdPrefix   string
	finalsCatID string

	dgGuild *discordgo.Guild

	ds       *discordgo.Session
	dbSchema string
}

func (s *Service) newGuild(dgGuild *discordgo.Guild) error {

	s.Log.Println("Attempting to load guild:", dgGuild.Name, dgGuild.ID)

	g := guild{
		cmds:     interpreterGuild(),
		dgGuild:  dgGuild,
		ds:       s.ds,
		dbSchema: fmt.Sprintf("%s_guild%s", service.Schema(s), dgGuild.ID),
	}

	// Verify Database Schema
	ok, err := simpsql.SchemaExists(g.dbSchema)
	if err != nil {
		return err
	} else if !ok {
		s.Log.Print("Guild has no database schema. Attempting fisrt time setup")
		if err = simpsql.MakeSchema(g.dbSchema); err != nil {
			return err
		}
		tx, err := simpsql.UsingSchema(g.dbSchema)
		if err != nil {
			tx.Rollback()
			simpsql.DelSchema(g.dbSchema)
			return err
		}

		sc, _ := simpsql.Open("sd_guild_schema.sql")
		for sc.Next() {
			s.Log.Print(sc.Stmt())
		}

		if err = simpsql.ExecScript(tx, "sd_guild_schema.sql"); err != nil {
			tx.Rollback()
			simpsql.DelSchema(g.dbSchema)
			return err
		}
		tx.Commit()
	}

	err = g.initCommands(s)
	if err != nil {
		return err
	}

	s.loadSettings(&g)

	s.guilds[dgGuild.ID] = &g
	s.Log.Println("Guild connected:", g.dgGuild.Name)
	return nil
}

func (g *guild) id() string {
	return g.dgGuild.ID
}

func (g *guild) initCommands(s *Service) error {

	g.cmds.AddCommand("terminal admin", adminTerminal)
	g.cmds.AddCommand("edit", classTerminal)

	return nil
}

// -----COMMAND FUNCTIONS--------

func verifyGuild(ext []interface{}) (*Service, *guild, *discordgo.MessageCreate) {
	e0 := ext[0].(*Service)
	e1 := ext[1].(*guild)
	e2 := ext[2].(*discordgo.MessageCreate)
	return e0, e1, e2
}

func adminTerminal(ctx context.Context, args []string, ext ...interface{}) error {

	s, g, msg := verifyGuild(ext)

	//TODO: My ID hardcoded as Amdin (bad)
	if msg.Author.ID == "84787975480700928" {
		return s.newTerminal(msg.Author.ID, adminCommands(), g, termTimeout, "Started an Admin Terminal")
	}
	s.ds.ChannelMessageSend(msg.ChannelID, "Access Denied")
	return nil
}

func classTerminal(ctx context.Context, args []string, ext ...interface{}) error {

	s, g, msg := verifyGuild(ext)

	return s.newTerminal(msg.Author.ID, classEditCommands(), g, termTimeout,
		"Hallo! Ich kann dir helfen deine Vorlesungen zu konfigurieren! Ganz einfach diese Commands (ohne `!`) eingeben.\n`search <begriff>` um nach vorlesungen zu suchen\n`join <ID>` um der Volesung beizutreten\n`leave <ID>` um eine Vorlesung zu verlassen")
}

func cmdTest(ctx context.Context, args []string, ext ...interface{}) error {

	s, _, msg := verifyGuild(ext)

	_, err := s.ds.ChannelMessageSend(msg.ChannelID, "Pong!")
	if err != nil {
		s.Log.Print("Error sending message: ", err)
	}

	return nil
}
