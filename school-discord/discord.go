package schooldiscord

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/bwmarrin/discordgo"
)

func (s *Service) handleDefaultMsg(m *discordgo.MessageCreate) {

	switch m.Type {
	case discordgo.MessageTypeDefault:
		if m.GuildID == "" {
			s.handleDefaultDirectMsg(m)
		} else {
			s.handleDefaultGuildMsg(m)
		}
	}

}

// Handles a Default message sent to a Guild
func (s *Service) handleDefaultDirectMsg(m *discordgo.MessageCreate) {

	t, ok := s.terminals[m.ChannelID]
	if ok { // If a terminal exists on the receiving channel
		t.in <- m
	} else { // If there is no terminal on the receiving channel
		s.ds.ChannelMessageSend(m.ChannelID,
			fmt.Sprintf("There is currently no active terminal on this channel. Please go to your %s administered server to start a new terminal",
				s.ds.State.User.Username))
	}
}

// Handles a Default message sent to a Guild
func (s *Service) handleDefaultGuildMsg(m *discordgo.MessageCreate) {
	g := s.guilds[m.GuildID]

	//check for command prefix
	if strings.HasPrefix(m.Content, g.cmdPrefix) {
		g.cmds.Run(context.TODO(), strings.Replace(m.Content, g.cmdPrefix, "", 1), s, g, m)
	}
}

func (s *Service) joinFinal(g *guild, final *modelFinal, userID string) (err error) {

	chanName := fmt.Sprintf("%s [%d]", final.name, final.id)

	// check to see if the final already has a channel/role
	if final.channelID == "" {

		var (
			r *discordgo.Role
			c *discordgo.Channel
		)

		r, err = s.makeRole(g, chanName)
		if err != nil {
			return err
		}

		c, err = s.makeTextChan(g, chanName, g.finalsCatID, r.ID)
		if err != nil {
			delerr := s.ds.GuildRoleDelete(g.id(), r.ID)
			if delerr != nil {
				return delerr
			}
			return err
		}

		err = InsertRole(g, r.ID)
		if err != nil {
			derr := s.ds.GuildRoleDelete(g.id(), r.ID)
			if derr != nil {
				return derr
			}
			_, derr = s.ds.ChannelDelete(c.ID)
			if derr != nil {
				return derr
			}
			return err
		}

		err = insertChannel(g, c.ID, r.ID)
		if err != nil {
			derr := s.ds.GuildRoleDelete(g.id(), r.ID)
			if derr != nil {
				return derr
			}
			_, derr = s.ds.ChannelDelete(c.ID)
			if derr != nil {
				return derr
			}
			return err
		}

		err = setFinalChannel(g, final.id, c.ID)
		if err != nil {
			derr := s.ds.GuildRoleDelete(g.id(), r.ID)
			if derr != nil {
				return derr
			}
			_, derr = s.ds.ChannelDelete(c.ID)
			if derr != nil {
				return derr
			}
			return err
		}

		final.roleID = r.ID
		final.channelID = c.ID
	}

	usr, err := getUser(g, userID)
	if err != nil {
		return err
	}

	if usr == nil {
		err = newUser(g, userID)
		if err != nil {
			return err
		}
	}

	ok, err := UserHasFinal(g, userID, final.id)
	if err != nil {
		return err
	}

	if ok {
		return AlreadyJoinedError{errors.New("already joined")}
	}

	err = s.ds.GuildMemberRoleAdd(g.id(), userID, final.roleID)
	if err != nil {
		return err
	}

	err = AddUserToFinal(g, userID, final.id)
	if err != nil {
		derr := s.ds.GuildMemberRoleRemove(g.id(), userID, final.roleID)
		if derr != err {
			return err
		}
		return err
	}

	return
}

func (s *Service) leaveFinal(g *guild, final *modelFinal, userID string) error {

	ok, err := UserHasFinal(g, userID, final.id)
	if err != nil {
		return err
	}

	if !ok {
		return NotJoinedError{errors.New("not joined")}
	}

	err = RemoveUserFromFinal(g, userID, final.id)
	if err != nil {
		return err
	}

	err = s.ds.GuildMemberRoleRemove(g.id(), userID, final.roleID)
	if err != nil {
		derr := AddUserToFinal(g, userID, final.id)
		if derr != nil {
			return derr
		}
		return err
	}

	return nil
}

func newPermViewChan(roleID string, guildID string, canView bool) (perm *discordgo.PermissionOverwrite) {
	perm = &discordgo.PermissionOverwrite{}
	perm.ID = guildID //@everyone role ID is the guild's ID (this is used as the default)
	perm.Type = "role"

	if roleID != "" {
		perm.ID = roleID
	}

	if canView {
		perm.Allow = 0x00000400 // permission for view channel -- source: https://discord.com/developers/docs/topics/permissions
	} else {
		perm.Deny = 0x00000400
	}
	return
}

func (s *Service) makeTextChan(g *guild, name string, parentChanID string, roleID string) (*discordgo.Channel, error) {

	perm := []*discordgo.PermissionOverwrite{
		newPermViewChan("", g.dgGuild.ID, false),
		newPermViewChan(roleID, g.dgGuild.ID, true),
	}

	data := discordgo.GuildChannelCreateData{
		Name:                 name,
		Type:                 discordgo.ChannelTypeGuildText,
		PermissionOverwrites: perm,
		ParentID:             parentChanID,
	}

	return s.ds.GuildChannelCreateComplex(g.id(), data)
}

func (s *Service) makeRole(g *guild, name string) (*discordgo.Role, error) {

	role, err := s.ds.GuildRoleCreate(g.dgGuild.ID)
	if err != nil {
		return nil, err
	}

	roles := g.dgGuild.Roles
	basePerm := roles[0].Permissions //roles[0] is the @everyone role

	r, err := s.ds.GuildRoleEdit(g.dgGuild.ID, role.ID, name, role.Color, false, basePerm, true)
	if err != nil {
		return nil, err
	}

	return r, nil
}

type AlreadyJoinedError struct {
	error
}
type NotJoinedError struct {
	error
}
