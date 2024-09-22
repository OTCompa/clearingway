package clearingway

import (
	"fmt"
	"strings"

	"github.com/Veraticus/clearingway/internal/discord"
	"github.com/bwmarrin/discordgo"
)

// types of UI elements
type MenuType string

const (
	MenuMain      MenuType = "menuMain"
	MenuVerify    MenuType = "menuVerify"
	MenuRemove    MenuType = "menuRemove"
	MenuEncounter MenuType = "menuEncounter"
)

type CommandType string

const (
	CommandMenu             CommandType = "menu"
	CommandClearsModal      CommandType = "clearsModal"
	CommandRemoveComfy      CommandType = "removeComfy"
	CommandRemoveColor      CommandType = "removeColor"
	CommandRemoveAll        CommandType = "removeAll"
	CommandEncounterProcess CommandType = "encounterProcess"
)

// struct to hold data for all different menu components
type Menus struct {
	Menus map[string]*Menu
}

type Menu struct {
	Name           string              // internal name to uniquely identify menus
	Type           MenuType            // type of menu to differentiate AdditionalData types
	Title          string              // title to show in embed
	Description    string              // optional description to show in embed
	ImageURL       string              // optional image URL
	AdditionalData *MenuAdditionalData // additional data depending on MenuType
}

type MenuRoleHelper struct {
	Role         *Role
	Prerequisite *Role
}

type MenuAdditionalData struct {
	MessageMainMenu  *discordgo.MessageSend
	MessageEphemeral *discordgo.InteractionResponse
	Roles            map[string]*MenuRoleHelper
	ExtraRoles       []*Role
	RoleType         []RoleType
	MultiSelect      bool
	RequireClear     bool
}

func (m *Menu) Init(c *ConfigMenu) {
	m.Name = c.Name
	m.Type = MenuType(c.Type)
	m.Title = c.Title

	if len(c.Description) != 0 {
		m.Description = c.Description
	}

	if len(c.ImageUrl) != 0 {
		m.ImageURL = c.ImageUrl
	}

	switch m.Type {
	case MenuEncounter:
		m.AdditionalData = &MenuAdditionalData{}
		data := m.AdditionalData

		if len(c.RoleType) != 0 {
			for _, roleType := range c.RoleType {
				data.RoleType = append(data.RoleType, RoleType(roleType))
			}
		}

		if c.MultiSelect {
			data.MultiSelect = true
		} else {
			data.MultiSelect = false
		}

		if c.RequireClear {
			data.RequireClear = true
		} else {
			data.RequireClear = false
		}

		for _, configRole := range c.ConfigRoles {
			role := &Role{}
			if len(configRole.Name) != 0 {
				role.Name = configRole.Name
			}
			if len(configRole.Description) != 0 {
				role.Description = configRole.Description
			}
			if configRole.Color != 0 {
				role.Color = configRole.Color
			}
			if configRole.Hoist {
				role.Hoist = true
			}
			if configRole.Mention {
				role.Mention = true
			}
			data.ExtraRoles = append(data.ExtraRoles, role)
		}
	}
}

// Creates the menu component of MenuRemove
func (m *Menu) MenuRemoveInit() {
	type removeButton struct {
		name        string
		commandType CommandType
	}

	removeButtons := []discordgo.MessageComponent{}
	removeButtonsList := []removeButton{
		{name: "Uncomfy", commandType: CommandRemoveComfy},
		{name: "Uncolor", commandType: CommandRemoveColor},
		{name: "Remove All", commandType: CommandRemoveAll},
	}

	for _, button := range removeButtonsList {
		customIDslice := []string{string(MenuRemove), string(button.commandType)}
		removeButtons = append(removeButtons, discordgo.Button{
			Label:    button.name,
			Style:    discordgo.DangerButton,
			Disabled: false,
			CustomID: strings.Join(customIDslice, " "),
		})
	}

	message := &discordgo.InteractionResponseData{
		Flags: discordgo.MessageFlagsEphemeral,
		Embeds: []*discordgo.MessageEmbed{
			{
				Title:       m.Title,
				Description: m.Description,
			},
		},
		Components: []discordgo.MessageComponent{
			discordgo.ActionsRow{
				Components: removeButtons,
			},
		},
	}

	if len(m.ImageURL) > 0 {
		message.Embeds[0].Image = &discordgo.MessageEmbedImage{URL: m.ImageURL}
	}

	m.AdditionalData = &MenuAdditionalData{
		MessageEphemeral: &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: message,
		},
	}

}

func (g *Guild) DefaultMenus() {
	g.Menus.Menus[string(MenuMain)] = &Menu{
		Name:        string(MenuMain),
		Type:        MenuMain,
		Title:       "Welcome to " + g.Name,
		Description: "Use the buttons below to assign roles!",
	}

	g.Menus.Menus[string(MenuVerify)] = &Menu{
		Name:  string(MenuVerify),
		Type:  MenuVerify,
		Title: "Verify Character",
	}

	g.Menus.Menus[string(MenuRemove)] = &Menu{
		Name:        string(MenuRemove),
		Type:        MenuRemove,
		Title:       "Remove Roles",
		Description: "Use the buttons below to remove Clearingway related roles!",
	}
}

// Sends the main menu as an standalone message in the channel it is called in
func (c *Clearingway) MenuMainSend(s *discordgo.Session, i *discordgo.InteractionCreate) {
	g, ok := c.Guilds.Guilds[i.GuildID]
	if !ok {
		fmt.Printf("Interaction received from guild %s with no configuration!\n", i.GuildID)
		return
	}

	menu := g.Menus.Menus[string(MenuMain)]
	additionalData := menu.AdditionalData

	_, err := s.ChannelMessageSendComplex(i.ChannelID, additionalData.MessageMainMenu)
	if err != nil {
		fmt.Printf("Error sending Discord message: %v\n", err)
		return
	}

	err = discord.StartInteraction(s, i.Interaction, "Sent menu message.")
	if err != nil {
		fmt.Printf("Error sending Discord message: %v\n", err)
		return
	}
}

// Creates an response to an interaction with a static menu
func (c *Clearingway) MenuStaticRespond(s *discordgo.Session, i *discordgo.InteractionCreate, menuName string) {
	g, ok := c.Guilds.Guilds[i.GuildID]
	if !ok {
		fmt.Printf("Interaction received from guild %s with no configuration!\n", i.GuildID)
		return
	}

	menu := g.Menus.Menus[menuName]
	additionalData := menu.AdditionalData

	err := s.InteractionRespond(i.Interaction, additionalData.MessageEphemeral)
	if err != nil {
		fmt.Printf("Error sending Discord message: %v\n", err)
		return
	}
}

// Returns with all the additional roles that were specified
// by the yaml under the menu config
func (ms *Menus) Roles() *Roles {
	roles := &Roles{Roles: []*Role{}}

	for _, menu := range ms.Menus {
		if menu.Type == MenuEncounter {
			roles.Roles = append(roles.Roles, menu.AdditionalData.ExtraRoles...)
		}
	}

	return roles
}
