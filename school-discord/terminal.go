package schooldiscord

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/Petrify/simp-core/commands"
	"github.com/bwmarrin/discordgo"
)

const termTimeout = 600 * time.Second
const closeCommand string = "!close"
const killCommand string = "!kill"

type terminal struct {
	userID string
	chanID string
	origin *guild
	serv   *Service

	in   chan *discordgo.MessageCreate
	stop chan string

	cmds *commands.Interpreter

	tMax  time.Duration
	timer *time.Timer
}

func (s *Service) newTerminal(userID string, cmds *commands.Interpreter, source *guild, timeout time.Duration, message string) error {

	//get DM channel for user
	channel, err := s.ds.UserChannelCreate(userID)
	if err != nil {
		return err
	}

	_, ok := s.terminals[channel.ID]
	if ok {
		s.ds.ChannelMessageSend(channel.ID, fmt.Sprintf("There is already an active terminal on this channel. Please use %s to close this terminal before opening a new one. If the terminal is stuck, use %s (not recommended)", closeCommand, killCommand))
		return errors.New("terminal already exists")
	}

	term := &terminal{
		serv:   s,
		userID: userID,
		chanID: channel.ID,
		origin: source,

		in:   make(chan *discordgo.MessageCreate),
		stop: make(chan string, 1),

		cmds: cmds,

		tMax:  timeout,
		timer: time.NewTimer(timeout),
	}

	term.timer.Stop() //so that the timer only truly starts when the terminal's loop begins
	s.terminals[channel.ID] = term
	term.Print(message)
	go term.loop() //start terminal read loop
	return nil
}

//Removes terminal from the undelying service
func (t *terminal) rmTerm() {
	_, ok := t.serv.terminals[t.chanID]
	if ok {
		delete(t.serv.terminals, t.chanID)
	}
}

// closes a terminal
func (t *terminal) close(reason string) {
	t.stop <- reason
}

func (t *terminal) cleanup(reason string) {
	t.timer.Stop()
	t.rmTerm()
	close(t.in)
	t.Print(fmt.Sprintf("Terminal is now closed.\nReason: %s\n", reason))
}

func (t *terminal) Print(text ...interface{}) (err error) {
	_, err = t.serv.ds.ChannelMessageSend(t.chanID, fmt.Sprint(text...))
	return
}

func (t *terminal) Printf(format string, a ...interface{}) (err error) {
	_, err = t.serv.ds.ChannelMessageSend(t.chanID, fmt.Sprintf(format, a...))
	return
}

func (t *terminal) Read() (text string, ok bool) {
	msg, ok := <-t.in
	if !ok {
		return "", ok
	}
	text = msg.Message.Content
	return
}

func (t *terminal) loop() {
	var inp *discordgo.MessageCreate
	// reset the timer to start it
	t.timer.Reset(t.tMax)

	for {
		//force checking timer & stop signal before allowing checking input
		select {
		case <-t.timer.C:
			t.cleanup("The session has expired")
			return

		case reason := <-t.stop:
			t.cleanup(reason)
			return

		default:

			//Check st
			select {
			case reason := <-t.stop:
				t.cleanup(reason)
				return

			case <-t.timer.C:
				t.cleanup("The session has expired")
				return

			case inp = <-t.in:
				t.timer.Stop() //so that the session does not expire during command execution
				err := t.cmds.Run(context.TODO(), strings.ToLower(inp.Message.Content), t, inp)
				if err != nil {
					t.handleCmdErr(err)
				}
			}
		}
		t.timer.Reset(t.tMax)
	}
}

func (t *terminal) handleCmdErr(err error) {
	switch err.(type) {
	case commands.InvalidCommandError:
		t.Print("Unknown command")
	case commands.InvalidArgsError:
		t.Print("That Command does not support those arguments") //TODO: NYI
	case commands.ExecutionError:
		t.serv.Log.Print("Encountered error while executing a command: ", err)
		t.Print(
			`Uh Oh! An error occurred while executing your command! 
			If this issue persists please file a an error report`)
	default:
		t.serv.Log.Print("Encountered error while executing a command: ", err)
		t.Print("An unknown error has occured")
	}
}

func verifyTerm(ext []interface{}) (*terminal, *discordgo.MessageCreate) {
	e0 := ext[0].(*terminal)
	e1 := ext[1].(*discordgo.MessageCreate)
	return e0, e1
}

func cmdSearch(ctx context.Context, args []string, ext ...interface{}) error {
	t, _ := verifyTerm(ext)

	var key string
	var max int
	if len(args) == 0 {
		t.Print("Bitte Suchbegriff eingeben")
		return nil
	}

	max = 10
	key = strings.Join(args, " ")

	ctlg, err := t.serv.getModuleCatalog(t.origin)
	if err != nil {
		return err
	}

	matches := fuzzySearch(ctlg, key, max)
	if len(matches) == 0 {
		return t.Print(fmt.Sprintf("Keine Ergebnisse Für **%s**", key))
	}
	resp := strings.Builder{}
	resp.WriteString(fmt.Sprintf("Suchergebnisse für **%s**\n", key))
	resp.WriteString("```[-ID-] | Prüfungsfach (Studiengänge)\n")
	for _, m := range matches {
		resp.WriteString(fmt.Sprintf("[%4d] | %s (%s)\n", m.id, m.name, strings.Join(m.majors, ", ")))
	}
	resp.WriteString("```")
	t.Print(resp.String())
	return nil

}

func cmdJoin(ctx context.Context, args []string, ext ...interface{}) error {
	t, m := verifyTerm(ext)

	if len(args) == 0 {
		t.Print("Bitte ID eingeben")
		return nil
	}

	id, err := strconv.Atoi(args[0])
	if err != nil {
		t.Print(args[0], " ist keine ID")
		return nil
	}

	mf, err := t.serv.getFinal(int64(id), t.origin)
	if err != nil {
		return err
	}

	if mf == nil {
		t.Print(args[0], " wurde nicht gefunden")
		return nil
	}

	err = t.serv.joinFinal(t.origin, mf, m.Author.ID)
	if err != nil {
		switch err.(type) {
		case AlreadyJoinedError:
			t.Print("Du bist dieser Prüfung bereits beigetreten")
			return nil
		}
		return err
	}

	t.Printf("**%s** wurde erfolgreich zu Deinen Prüfungen hinzugefügt.", mf.name)
	return nil
}

func cmdLeave(ctx context.Context, args []string, ext ...interface{}) error {
	t, m := verifyTerm(ext)

	if len(args) == 0 {
		t.Print("Bitte ID eingeben")
		return nil
	}

	id, err := strconv.Atoi(args[0])
	if err != nil {
		t.Print(args[0], " ist keine ID >:(")
		return nil
	}

	mf, err := t.serv.getFinal(int64(id), t.origin)
	if err != nil {
		return err
	}

	if mf == nil {
		t.Print(args[0], " wurde nicht gefunden")
		return nil
	}

	err = t.serv.leaveFinal(t.origin, mf, m.Author.ID)
	if err != nil {
		switch err.(type) {
		case NotJoinedError:
			t.Printf("Man kann keine Prüfung varlassen der Man nie beigetreten bist!")
		}
		return err
	}

	t.Printf("**%s** wurde erfolgreich von Deinen Prüfungen gelöscht.", mf.name)
	return nil
}

func cmdList(ctx context.Context, args []string, ext ...interface{}) error {
	t, m := verifyTerm(ext)

	lst, err := getUserFinals(t.origin, m.Author.ID)
	if err != nil {
		return err
	}

	if len(lst) == 0 {
		return t.Print("Keine Prüfungen gefunden")
	}
	resp := strings.Builder{}
	resp.WriteString("Deine Prüfungen:")
	resp.WriteString("```")
	for _, m := range lst {
		resp.WriteString(fmt.Sprintf("[%4d] | %s \n", m.id, m.name))
	}
	resp.WriteString("```")
	t.Print(resp.String())

	return nil
}
