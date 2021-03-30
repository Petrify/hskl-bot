package schooldiscord

import "github.com/bwmarrin/discordgo"

func (s *Service) registerHandlers() {
	s.ds.AddHandler(s.defGuildCreate())
	s.ds.AddHandler(s.defMessageCreate())
}

func (s *Service) defGuildCreate() func(ds *discordgo.Session, m *discordgo.GuildCreate) {
	return func(ds *discordgo.Session, m *discordgo.GuildCreate) {
		err := s.newGuild(m.Guild)
		if err != nil {
			s.Log.Print("Error while loading guild: ", err)
		}
	}
}

func (s *Service) defMessageCreate() func(ds *discordgo.Session, m *discordgo.MessageCreate) {

	return func(ds *discordgo.Session, m *discordgo.MessageCreate) {
		if m.Author.Bot {
			return
		}
		s.handleDefaultMsg(m)
	}
}
