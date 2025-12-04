package notify


type NotifyCMD struct {
	Discord DiscordCMD `cmd:"" help:"Notify users using Discord webhook."`
}
