package schooldiscord

import (
	"errors"
	"log"

	"github.com/Petrify/simp-core/service"
	"github.com/bwmarrin/discordgo"
)

const typeName = "school-discord"

func init() {
	service.NewSType(typeName, serviceCtor, false)
}

func Start() {
	serv, err := service.NewService(typeName, 1, "discord")
	if err != nil {
		println(err.Error())
		return
	}
	serv.Start()
}

type Service struct {
	//Service specific Members
	token   string
	ds      *discordgo.Session
	running bool

	//guild connections
	guilds map[string]*guild //mapped by guildID

	//terminal connections
	terminals map[string]*terminal //Mapped by channelID

	//Abstract service implementation
	service.AbstractService
}

func serviceCtor(id int64, name string, logger *log.Logger) service.Service {
	s := Service{
		token:     "",
		ds:        nil,
		running:   false,
		guilds:    make(map[string]*guild),
		terminals: make(map[string]*terminal),

		AbstractService: *service.NewAbstractService(name, id, logger),
	}

	return &s
}

func (s *Service) Setup() error {
	err := service.BuildSchema(s, "sd_service_schema.sql")
	return err
}

//Initializes the service to recieve messages from discord
func (s *Service) Init() error {

	err := s.getToken()
	if err != nil {
		return err
	} else if s.token == "" {
		return errors.New("enter a valid bot-token into your database")
	}
	ds, err := discordgo.New("Bot " + s.token)
	if err != nil {
		return err
	}
	s.ds = ds

	ds.Identify.Intents = discordgo.MakeIntent(discordgo.IntentsDirectMessages | discordgo.IntentsGuildMessages | 1)

	s.registerHandlers()

	return nil
}

func (s *Service) Start() error {

	err := s.ds.Open()
	if err != nil {
		s.Log.Println("error opening discord connection")
		return err
	}

	s.Log.Printf("[%d] %s Started Successfully", s.ID(), s.Name())

	return nil
}

func (s *Service) Stop() {
	err := s.ds.Close()
	if err != nil {
		s.Log.Println("Error Closing discord connection", err)
	}
}
