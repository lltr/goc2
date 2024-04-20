package main

import (
	"fmt"
	"github.com/bwmarrin/discordgo"
)

var (
	commands = []*discordgo.ApplicationCommand{
		{
			Name:        "basic-command",
			Description: "Basic command",
		},
		{
			Name:        "agents",
			Description: "Returns list of all connected agents",
		},
		{
			Name:        "cmd",
			Description: "Send command to remote execute",
			Options: []*discordgo.ApplicationCommandOption{

				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "agent-id",
					Description: "Identifier of agent",
					Required:    true,
				},
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "command",
					Description: "Command to execute",
					Required:    true,
				},
			},
		},
		{
			Name:        "wcmd",
			Description: "Send waitable command to remote execute",
			Options: []*discordgo.ApplicationCommandOption{

				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "agent-id",
					Description: "Identifier of agent",
					Required:    true,
				},
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "command",
					Description: "Command to execute",
					Required:    true,
				},
			},
		},
	}

	commandHandlers = map[string]func(s *discordgo.Session, i *discordgo.InteractionCreate, ah *ApiHandler){
		"basic-command": func(s *discordgo.Session, i *discordgo.InteractionCreate, ah *ApiHandler) {
			s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: "Hey there! Congratulations, you just executed your first slash command",
				},
			})
		},
		"agents": func(s *discordgo.Session, i *discordgo.InteractionCreate, ah *ApiHandler) {
			s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: ah.getAgents(),
				},
			})
		},
		"cmd": func(s *discordgo.Session, i *discordgo.InteractionCreate, ah *ApiHandler) {
			options := i.ApplicationCommandData().Options
			optionMap := make(map[string]*discordgo.ApplicationCommandInteractionDataOption, len(options))
			for _, opt := range options {
				optionMap[opt.Name] = opt
			}

			input := Input{}

			if opt, ok := optionMap["agent-id"]; ok {
				input.AgentId = opt.StringValue()
			}
			if opt, ok := optionMap["command"]; ok {
				input.Input = opt.StringValue()
			}

			encodedTransferPacket := encodeTransferPacket("command", input.Input)
			ah.Hub.sendTarget <- Message{clientId: input.AgentId, data: encodedTransferPacket}

			s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				// Ignore type for now, they will be discussed in "responses"
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: fmt.Sprintf("Command sent."),
				},
			})
		},
		"wcmd": func(s *discordgo.Session, i *discordgo.InteractionCreate, ah *ApiHandler) {
			options := i.ApplicationCommandData().Options
			optionMap := make(map[string]*discordgo.ApplicationCommandInteractionDataOption, len(options))
			for _, opt := range options {
				optionMap[opt.Name] = opt
			}

			input := Input{}

			if opt, ok := optionMap["agent-id"]; ok {
				input.AgentId = opt.StringValue()
			}
			if opt, ok := optionMap["command"]; ok {
				input.Input = opt.StringValue()
			}

			encodedTransferPacket := encodeTransferPacket("waitable_command", input.Input)
			ah.Hub.sendTarget <- Message{clientId: input.AgentId, data: encodedTransferPacket}

			s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				// Ignore type for now, they will be discussed in "responses"
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: fmt.Sprintf("Waitable command sent."),
				},
			})
		},
	}
)
