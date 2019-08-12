package main

import (
	"fmt"
	"strings"
	"flag"
	"time"
	"regexp"
	"encoding/json"
	"encoding/xml"
	"strconv"
	"errors"
	"bytes"
	"math/rand"
	"net/http"
	"io/ioutil"
	"net/url"
	"io"
	"os"
	"bufio"
	"github.com/rylio/ytdl"
	"github.com/valyala/fasthttp"
	"github.com/Time6628/OpenTDB-Go"
	"github.com/mvdan/xurls"
	"github.com/Wubsy/dgvoice"
	"github.com/bwmarrin/discordgo"
	"github.com/garyburd/redigo/redis"
	"github.com/knspriggs/go-twitch"
	"github.com/Wubsy/gowikia-b"
)



var (
	rateLimitTime = 5 // In seconds
	announcementChannel string
	twitchCheckEnable = true
	client_id string
	redisAddr string
	redisConnected bool
	token string
	BotID string
	userTime = make(map[string]string)
	client = fasthttp.Client{ReadTimeout: time.Second * 10, WriteTimeout: time.Second * 10}
	trivia = OpenTDB_Go.New(client)
	nofilter []string
	Folder = "download/"
	prefixChar string // Don't use  # and @ because it might mess with channels
	Qreplacer = strings.NewReplacer("&quot;", "\"", "&#039;", "'") //TODO: Legit I can just use the HTML thing that automatically does this. No idea why I haven't done this yet
	Lreplacer = strings.NewReplacer(" ", "+")
	version = "0.8.0"
	isVConnected = false
	APlaylist = "autoplaylist.txt"
	triviaStatus = false
	playSkip = true
	Bot *discordgo.User
	articleName string
	articleUrl string
	articleId int
	totalItems int
	twitchUsers twitchUserSlice //This was more difficult to get working than it needed to be
	commands = []string{
		prefixChar + "removefilter",
		prefixChar + "enablefilter",
		prefixChar + "dogbot",
		prefixChar + "mute",
		prefixChar + "allmute",
		prefixChar + "cat",
		prefixChar + "doge",
		prefixChar + "leaveserver",
		prefixChar + "fplay",
		prefixChar + "csay",
		prefixChar + "play",
		prefixChar + "skip",
		prefixChar + "disconnect or "+prefixChar+"dc",
		prefixChar + "streaming",
		prefixChar + "simpask",
		prefixChar + "lmgtfy",
		prefixChar + "gay",
		prefixChar + "clean",
		prefixChar + "info",
		prefixChar + "playskip",
		prefixChar + "skiplist",
		prefixChar + "trivia",
		prefixChar + "setcredits",
		prefixChar + "credits",
		prefixChar + "flip",
		prefixChar + "slots",
		prefixChar + "daily",
		prefixChar + "srsearch",}
	queue = []string{}
	nowPlaying string
	logging bool
	firstpasstwitch = true
)

func init() {
	flag.Var(&twitchUsers, "tu", "Twitch users that for which to notify")
	flag.StringVar(&prefixChar, "pfx", ".", "Prefix that the bot uses for commands")
	flag.StringVar(&redisAddr, "r", "localhost:6379", "Address of Redis database.") //TODO: Add support for passworded databases. This is a huge security flaw that should be fixed ASAP. That being said, I'll probably not do it for awhile
	flag.StringVar(&client_id, "cl", "", "Twitch API client ID")
	flag.StringVar(&announcementChannel, "c", "", "Channel where all the twitch message announcements go")
	flag.StringVar(&token, "t", "", "Bot Token")
	flag.BoolVar(&logging, "l", false, "Enables/Disables Printing Messages to CMD")
	flag.Parse()
}

func main()  {
	go forever()

	url := "https://wubsy.github.io/DogBot"
	resp, err := http.Get(url)
	if err != nil {
		fmt.Println("Could not reach version check server.")
	} else {
		defer resp.Body.Close()
	}

	if err == nil {
		html, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			fmt.Println(err)
		}
		fmt.Printf("%s\n", html)
	}
	fmt.Println("Starting Dogbot: "+version)

	if token == "" {
		fmt.Println("No token provided. Please run: dogbot -t <bot token>")
		return
	}
	dg, err := discordgo.New("Bot " + token)
	if err != nil {
		fmt.Println("Error creating Discord session: ", err)
		return
	}

	u, err := dg.User("@me")
	if err != nil {
		fmt.Println("Error obtaining account details,", err)
	}

	if redisAddr == "" {
		fmt.Println("Redis address is not set. None of the economy commands will work until enabled.")
	}

	if announcementChannel == "" { //TODO: More checks will eventually need to be run. If the bot does not have access to the channel given, the bot may hang everytime, or even crash
		fmt.Println("No Twitch announcement channel set. Twitch functions will not work.\nUse flag \"c\" to set one." )
		twitchCheckEnable = false
		firstpasstwitch = false
	}

	if client_id == "" { //Same thing as above
		fmt.Println("Client id for Twitch not set. Twitch functions will not work.\nUse flag \"cl\" to set a client id.")
		twitchCheckEnable = false
		firstpasstwitch = false
	}

	dg.AddHandler(messageCreate)

	BotID = u.ID
	Bot = u

	err = dg.Open()

	if err != nil {
		fmt.Println("Could not open Discord session: ", err)
	}
	fmt.Println("DogBot is now running.  Press CTRL-C to exit. Bot Prefix is \""+prefixChar+"\"")
	select {}

}

func forever() {}

func twitchChecker(s *discordgo.Session, tUser string) {
	var isStreaming bool

	for twitchCheckEnable {
		doLater(func() {
			twitchSession, err := twitch.NewSession(twitch.NewSessionInput{ClientID: client_id})
			if err != nil {
				fmt.Println(err)
			}

			searchChannelsInput := twitch.SearchChannelsInputType{
				Query: tUser,
			}

			getChannelsInput := twitch.GetChannelInputType{
				Channel: searchChannelsInput.Query,
			}
			channelData, err := twitchSession.GetChannel(&getChannelsInput)
			if err != nil {
				fmt.Println(err)
			}

			Stream := twitch.GetStreamsInputType{
				Channel: searchChannelsInput.Query,
			}
			test, err := twitchSession.GetStream(&Stream)

			streamEmbedPrimerTeb := []*discordgo.MessageEmbedField{}


			if channelData != nil {
				streamEmbedPrimer := []*discordgo.MessageEmbedField{
					{Name: "Now Playing", Value: channelData.Game, Inline: true, },
					{Name: "Title", Value: channelData.Status, Inline: false, },
					{Name: "Followers", Value: strconv.Itoa(channelData.Followers), Inline: true, },
					{Name: "Total Views", Value: strconv.Itoa(channelData.Views), Inline: true, },
				}

				streamEmbedThumbnail := []*discordgo.MessageEmbedThumbnail{
					{URL: channelData.Logo, Width: 300, Height: 300, },
				}

				streamEmbedImage := []*discordgo.MessageEmbedImage{
					{URL: channelData.ProfileBanner, },
				}
				streamEmbedPrimerTeb = append(streamEmbedPrimerTeb, streamEmbedPrimer[0], streamEmbedPrimer[1], streamEmbedPrimer[2], streamEmbedPrimer[3], )
				embed := discordgo.MessageEmbed{
					Title:     channelData.DisplayName + " is streaming!",
					Color:     10181046,
					URL:       channelData.URL,
					Thumbnail: streamEmbedThumbnail[0],

					Fields: streamEmbedPrimerTeb,

					Image: streamEmbedImage[0],
				}
				if err != nil {
					fmt.Println(err)
				}
				if test != nil {
					if test.Total == 1 && !isStreaming {

						isStreaming = true
						message, err := s.ChannelMessageSendEmbed(announcementChannel, &embed)
						if err != nil {
							fmt.Print(err)
						}

						for isStreaming {
							doLater(func() {
								var checkOnlineStatus, err = twitchSession.GetStream(&Stream)
								if err != nil {
									fmt.Println(err)
								}
								if checkOnlineStatus != nil {
									if checkOnlineStatus.Total == 0 {
										isStreaming = false
										s.ChannelMessageEdit(announcementChannel, message.ID, channelData.DisplayName+"'s stream has ended.")
										return
									}
								}
							})
						}
					}
				}
			}
		})
	}
}

func messageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {
	if m.Author.ID == BotID {
		return
	}

	d, err := s.Channel(m.ChannelID)
	if err != nil {
		return
	}

	g, err := s.Guild(d.GuildID)
	if err != nil {
		return
	}

	member, err := s.GuildMember(g.ID, m.Author.ID)
	if err != nil {
		return
	}

	roles := member.Roles

	c := strings.ToLower(m.Content)

	filters := []*regexp.Regexp{regexp.MustCompile("traps aren't gay"), regexp.MustCompile("traps are not gay"), regexp.MustCompile("traps arent gay")}
	filter := false

	admin := false
	for i := 0; i < len(roles); i++ {
		role, _ := s.State.Role(g.ID, roles[i])
		if (role.Permissions & discordgo.PermissionAdministrator) == discordgo.PermissionAdministrator {
			admin = true
		}
	}

	t := time.Now()
	layout := "2006-01-02 15:04:05"
	timenow := t.Format(layout)

	if logging {
		if m.Content == "" && len(m.Embeds) > 0 {
			if m.Embeds[0].Image != nil {
				fmt.Println(timenow, m.Author.Username, "<#" + d.ID + ">: Embed: \nDesc: " + m.Embeds[0].Description + "\nImage: " + m.Embeds[0].Image.URL + "\nURL: " + m.Embeds[0].URL +"\nImage: " + m.Embeds[0].Image.ProxyURL+"\n")
			} else {
				fmt.Println(timenow, m.Author.Username, "<#"+d.ID+">: Embed: \nDesc: "+m.Embeds[0].Description)
			}

			if len(m.Embeds[0].Fields) > 0 {
				se := m.Embeds[0].Fields[0]
				fmt.Println(timenow, m.Author.Username, "<#" + d.ID + ">:\n RichEmbed: \nName: " + se.Name + "\n" + se.Value)
			}
		} else if len(m.Embeds) < 1 {
			fmt.Println(timenow, m.Author.Username, "<#" + d.ID + ">: " + m.Content)
		} else if len(m.Attachments) > 0 {
			fmt.Println(timenow, m.Author.Username, "<#" + d.ID + ">: \nProxyURL: " + m.Attachments[0].ProxyURL + "\nURL: " + m.Attachments[0].URL)
		}
	}

	if firstpasstwitch {
		for i := 0; i < len(twitchUsers); i++ {
			go twitchChecker(s, twitchUsers[i])
		}
		firstpasstwitch = false
	}

	if filterChannel(d.ID) {
		for i := 0; i < len(filters); i++ {
			filt := filters[i]
			if filt.MatchString(c) {
				filter = true
			}
		}
	}

	if filter {
		s.ChannelMessageDelete(m.ChannelID, m.ID)
		rm, _ := s.ChannelMessageSend(m.ChannelID, "Messaged removed from <@"+m.Author.ID+">.")
		removeLater(s, rm)
		return
	} else if strings.HasPrefix(c, prefixChar+"removefilter") {
		if filterChannel(d.ID) == false {
			e, _ := s.ChannelMessageSend(d.ID, "Channel already unfiltered.")
			removeLaterBulk(s, []*discordgo.Message{e, m.Message})
		} else {
			nofilter = append(nofilter, d.ID)
			e, _ := s.ChannelMessageSend(d.ID, "Channel is no longer filtered.")
			removeLaterBulk(s, []*discordgo.Message{e, m.Message})
		}
	}
	if strings.HasPrefix(c, prefixChar+"help") && !isRateLimited(s, d, m){ //TODO: HOLY GOD USE CASE/BREAK INSTEAD OF ELSEIF WHAT AM I DOING
		var comment string

		commandsEmbedPrimerTeb := []*discordgo.MessageEmbedField{}

		channel, err := s.UserChannelCreate(m.Author.ID)
		if err != nil {
			s.ChannelMessageSend(d.ID, err.Error())
		}
		for i := 0; i >= 0 && i < len(commands); i++ {
			//This is a pretty bad way to do this but eh
			if commands[i] == "removefilter" {
				comment = "Disables chat filter in the channel it is run in"
			}
			if commands[i] == "enablefilter" {
				comment = "Enables chat filter in the channel it is run in"
			}
			if commands[i] == "dogbot" {
				comment = "Displays bot version"
			}
			if commands[i] == "mute" {
				comment = "Mutes the user in the text channel the command is run in"
			}
			if commands[i] == "allmute" {
				comment = "Mutes the user in all text channels"
			}
			if commands[i] == "cat" {
				comment = "Displays up to 15 pictures of cats. No argument will result in a single image"
			}
			if commands[i] == "doge" {
				comment = "Displays up to 15 pictures of dogs. No argument will result in a single image/gif/mp4"
			}
			if commands[i] == "leaveserver" {
				comment = "Leaves the discord server until re-invited"
			}
			if commands[i] == "fplay" {
				comment = "Plays a file from download folder without the extension"
			}
			if commands[i] == "csay" {
				comment = "<channel id> <string>"
			}
			if commands[i] == "play" {
				comment = "YouTube link starting with https://www.youtube.com/ "
			}
			if commands[i] == "skip" {
				comment = "Skips current song"
			}
			if commands[i] == "disconnect or dc" {
				comment = "Disconnects from the current channel"
			}
			if commands[i] == "streaming" {
				comment = "Changes status to streaming"
			}
			if commands[i] == "simpask" {
				comment = "Responds with an answer to a simple, yes/no question"
			}
			if commands[i] == "lmgtfy" {
				comment = "Generates a Let Me Google That For You link and shortens it"
			}
			if commands[i] == "gay" {
				comment = "Determines how gay a user is"
			}
			if commands[i] == "clean" {
				comment = "Removes messages by argument number"
			}
			if commands[i] == "info" {
				comment = "Displays info of the bot"
			}
			if commands[i] == "playskip" {
				comment = "Toggles the ability to run a play command and skip it with out running the skip command"
			}
			if commands[i] == "skiplist" {
				comment = "Skips a full playlist"
			}
			if commands[i] == "trivia" {
				comment = "Starts trivia"
			}
			if commands[i] == "setcredits" {
				comment = "Sets the credits of the user that runs the command"
			}
			if commands[i] == "credits" {
				comment = "Shows the credits of the user that runs the command"
			}
			if commands[i] == "flip" {
				comment = "Runs a coinflip that can potentially double the bet"
			}
			if commands[i] == "slots" {
				comment = "It's like using a real slot machine except it sucks a lot more. You *must* not spam this command. If you go negative, blame yourself"
			}
			if commands[i] == "daily" {
				comment = "Every 24 hours, you can receive your daily 200 credits."
			}
			if commands[i] == "srsearch" {
				comment = "Give it a parameter to search with, and it will return something from the Slime Rancher Wikia"
			}

			if comment == ""{
				comment = "Error reading from slice."
			}


			commandsEmbedPrimer := []*discordgo.MessageEmbedField{
				{Name: commands[i], Value: comment, Inline: false},
			}
			commandsEmbedLink := []*discordgo.MessageEmbedField{
				{Name: "Github docs", Value: "For a more full list, go to https://github.com/Wubsy/DogBot", Inline: false},
			}
			commandsEmbedPrimerTeb = append(commandsEmbedPrimerTeb, commandsEmbedPrimer[0])
			embed := discordgo.MessageEmbed{
				Title:       "Commands",
				Color:       10181046,
				Description: "DogBot Commands",
				URL:         "https://twitter.com/DogBot4Discord",
				Fields:      commandsEmbedPrimerTeb,
			}
			embedLink := discordgo.MessageEmbed{
				Title:       "",
				Color:       10181046,
				Description: "More info",
				URL:         "https://dogbot.site.nfoservers.com",
				Fields:      commandsEmbedLink,
			}

			if i == len(commands)-1 {
				s.ChannelMessageSendEmbed(channel.ID, &embed)
				s.ChannelMessageSendEmbed(channel.ID, &embedLink)
			} else {
				i = i
			}
		}

	} else if strings.HasPrefix(c, prefixChar+"twitchcheck") {
		if twitchCheckEnable {
			twitchCheckEnable = false
		} else {
			twitchCheckEnable = true
		}
		s.ChannelMessageSend(m.ChannelID, "Twitch Check Status: "+strconv.FormatBool(twitchCheckEnable))
	} else if strings.HasPrefix(c, prefixChar+"enablefilter") {
		if filterChannel(d.ID) == false {
			toremove := -1
			for i := range nofilter {
				if nofilter[i] == d.ID {
					toremove = i
				}
			}
			nofilter = append(nofilter[:toremove], nofilter[toremove+1:]...)
			e, _ := s.ChannelMessageSend(d.ID, "Channel is now filtered.")
			removeLaterBulk(s, []*discordgo.Message{e, m.Message})
		} else {
			e, _ := s.ChannelMessageSend(d.ID, "Channel is already filtered.")
			removeLaterBulk(s, []*discordgo.Message{e, m.Message})
		}
	} else if strings.HasPrefix(c, prefixChar+"dogbot") && !isRateLimited(s, d, m){
		s.ChannelMessageSend(m.ChannelID, "bork bork beep boop! I am DogBot "+version+"!")
		return
	} else if strings.HasPrefix(c, prefixChar+"mute") && admin {
		cc := strings.TrimPrefix(c, prefixChar+"mute ")
		if !strings.Contains(cc, "@") {
			s.ChannelMessageSend(d.ID, "Please provide a user to mute!")
			return
		}
		arg := strings.Split(cc, " ")
		//fmt.Println(cc)
		user_id := strings.TrimPrefix(strings.TrimSuffix(arg[0], ">"), "<@")
		if !alreadyMuted(user_id, d) {
			s.ChannelPermissionSet(d.ID, user_id, "member", 0, discordgo.PermissionSendMessages)
			rm, _ := s.ChannelMessageSend(m.ChannelID, "Muted user "+arg[0]+"!")
			fmt.Println(m.Author.Username + " muted " + user_id)
			b := []*discordgo.Message{rm, m.Message, }
			removeLaterBulk(s, b)
		} else {
			rm, _ := s.ChannelMessageSend(m.ChannelID, "User already muted!")
			b := []*discordgo.Message{rm, m.Message, }
			removeLaterBulk(s, b)
		}
	} else if strings.HasPrefix(c, prefixChar+"allmute") && admin {
		cc := strings.TrimPrefix(c, prefixChar+"allmute ")
		arg := strings.Split(cc, " ")
		if !strings.Contains(cc, "@") {
			s.ChannelMessageSend(d.ID, "Please provide a user to mute!")
			return
		}
		user_id := strings.TrimPrefix(strings.TrimSuffix(arg[0], ">"), "<@")
		channels := g.Channels
		for i := 0; i < len(channels); i++ {
			channel := channels[i]
			if !alreadyMuted(user_id, channel) {
				s.ChannelPermissionSet(channel.ID, user_id, "member", 0, discordgo.PermissionSendMessages)
			}
		}
		rm, _ := s.ChannelMessageSend(m.ChannelID, "Muted user "+arg[0]+" in all channels!")
		b := []*discordgo.Message{rm, m.Message, }
		removeLaterBulk(s, b)
		fmt.Println(m.Author.Username + " muted " + user_id + " in all channels.")
	} else if strings.HasPrefix(c, prefixChar+"cat") {
		j := CatResponse{}
		cc := strings.TrimPrefix(c, prefixChar+"cat ")
		if i, err := strconv.ParseInt(cc, 10, 64); err != nil {
			getJson("http://aws.random.cat/meow", &j)
			s.ChannelMessageSend(d.ID, j.URL)
		} else {
			if i > 15 {
				i = 15
			}
			if i < 0 {
				i = 1
			}
			e := ""
			for b := int64(0); b < i; b++ {
				getJson("http://aws.random.cat/meow", &j)
				e = e + j.URL + " "
			}
			s.ChannelMessageSend(d.ID, e)
		}
	} else if strings.HasPrefix(c, prefixChar+"cat") && !isRateLimited(s, d, m) {
		j := CatResponse{}
		cc := strings.TrimPrefix(c, prefixChar+"cat ")
		if i, err := strconv.ParseInt(cc, 10, 64); err != nil {
			getJson("http://aws.random.cat/meow", &j)
			s.ChannelMessageSend(d.ID, j.URL)
		} else {
			if i > 15 || i < 0 {
				i = 15
			}
			e := ""
			for b := int64(0); b < i; b++ {
				getJson("http://aws.random.cat/meow", &j)
				e = e + j.URL + " "
			}
			s.ChannelMessageSend(d.ID, e)
		}
	} else if strings.HasPrefix(c, prefixChar+"doge") && !isRateLimited(s, d, m) {
		j := DogResponse{}
		cc := strings.TrimPrefix(c, prefixChar+"doge ")
		if i, err := strconv.ParseInt(cc, 10, 64); err != nil {
			getJson("https://random.dog/woof.json", &j)
			if strings.Contains(j.URL, ".mp4") {
				getJson("https://random.dog/woof.json", &j)
			} else {
				s.ChannelMessageSend(d.ID, j.URL)
			}
		} else {
			if i > 15 {
				i = 15
			}
			if i < 0 {
				i = 1
			}
			e := ""
			for b := int64(0); b < i; b++ {
				getJson("https://random.dog/woof.json", &j)
				if strings.Contains(j.URL, ".mp4") || strings.Contains(e, j.URL) {
					b--
				} else {
					e = e + j.URL + " "
				}
			}
			s.ChannelMessageSend(d.ID, e)

		}
	} else if strings.HasPrefix(c, "who's a good boy") && !isRateLimited(s, d, m)|| strings.HasPrefix(c, "whos a good boy") && !isRateLimited(s, d, m) {
		s.ChannelMessageSend(d.ID, "ME ME ME <@"+m.Author.ID+">")
	} else if strings.HasPrefix(c, "who's a bad boy")  && !isRateLimited(s, d, m)  || strings.HasPrefix(c, "whos a bad boy") && !isRateLimited(s, d, m) {
		s.ChannelMessageSend(d.ID, "YOU YOU YOU <@"+m.Author.ID+">")
	} else if strings.HasPrefix(c, prefixChar+"leaveserver") && admin {
		s.ChannelMessageSend(d.ID, "Bye :crying_cat_face: :wave: ")
		s.GuildLeave(d.GuildID)
		fmt.Println("Left", d.GuildID)
	} else if strings.HasPrefix(c, prefixChar+"fplay") && admin {
		pp := strings.TrimPrefix(m.Content, prefixChar+"fplay ")
		arg := strings.SplitAfterN(pp, " ", 1)
		guild, _ := s.Guild(d.GuildID)

		channel := getCurrentVoiceChannel(m.Author, s, guild)
		if channel != nil {
			dgv, err := s.ChannelVoiceJoin(d.GuildID, channel.ID, false, true)
			isVConnected = true
			if err != nil {
				s.ChannelMessageSend(m.ChannelID, "Must be in voice channel to play music")
				return
			} else {

				if !isVConnected {
					rate := time.Second * 2
					throttle := time.Tick(rate)
					for req := range "   " {
						<-throttle
						if req != 0 {
						}
					}
				}

				if !dgvoice.IsThisThingOn() {
					url := "https://www.youtube.com/watch?v=" + arg[0]
					vid, err := ytdl.GetVideoInfo(url)
					if err != nil {
						fmt.Println(err)
					}
					if vid.Title == "" {
						fmt.Println("Streaming " + arg[0] + ".mp3")
						//s.UpdateStatus(0, "Streaming "+arg[0]+".mp3") //This is bad for a public bot
					} else {
						//s.UpdateStatus(0, vid.Title)
					}

					nowPlaying = arg[0]
					err = dgvoice.PlayAudioFile(dgv, "download\\"+arg[0]+".mp3", s)
					nowPlaying = ""
					if err != nil {
						nowPlaying = ""
						return
					}
				} else {
					fmt.Println("Error playing file")
				}
			}

		} else {
			s.ChannelMessageSend(m.ChannelID, "File does not exist or "+arg[0]+".mp3 is not valid")
		}
	} else if strings.HasPrefix(c, prefixChar+"csay") {
		if m.Author.ID == "157630049644707840" || m.Author.ID == "155481695167053824" {
			removeNow(s, m.Message)
			cc := strings.TrimPrefix(m.Content, prefixChar+"csay ")
			chann := strings.SplitAfter(cc, " ")
			trimChann := strings.TrimPrefix(m.Content, prefixChar+"csay "+chann[0])
			fmt.Println(m.Author.Username, ": "+trimChann+" in "+chann[0], d.ID)
			if strings.Contains(chann[0], "#") {
				chann[0] = strings.TrimPrefix(chann[0], "<#")
				chann[0] = strings.TrimSuffix(chann[0], "> ")
			}
			_, err := s.ChannelMessageSend(chann[0], trimChann)
			if err != nil {
				fmt.Println(err)
			}
		}
	} else if strings.HasPrefix(c, prefixChar+"play ") && admin {
		//_ = dgvoice.IsThisThingOn()
		pp := strings.TrimPrefix(m.Content, prefixChar+"play ")
		if !strings.Contains(pp, "https://www.youtube.com/") && !strings.Contains(pp, "https://youtu.be/") {
			s.ChannelMessageSend(m.ChannelID, "Must be from`https://www.youtube.com/` or `https://youtu.be/`")
		} else {
			arg := strings.Split(pp, " ")
			if dgvoice.IsThisThingOn() {
				queue = append(queue, arg[0])
				s.ChannelMessageSend(d.ID, "Added `"+arg[0]+"` to the queue")
				return
			} else {
				queue = append(queue, arg[0])
			}

			url := xurls.Strict().FindString(m.Content)
			youtubeDl(url, m.Message, s)
			if err != nil {
				fmt.Println(err)
			}
			vid, _ := ytdl.GetVideoInfo(url)

			for j := 0; j < len(queue); j++ {
				var fileName string
				if strings.Contains(url, "https://youtu.be/") {
					fileName = strings.TrimPrefix(queue[0], "https://youtu.be/")
				} else {
					fileName = strings.TrimPrefix(queue[0], "https://www.youtube.com/watch?v=")
				}
				file := Folder + fileName + ".mp3"

				guild, _ := s.Guild(d.GuildID)
				channel := getCurrentVoiceChannel(m.Author, s, guild)
				if channel != nil {
					dgv, err := s.ChannelVoiceJoin(d.GuildID, channel.ID, false, true)
					isVConnected = true
					if err != nil {
						s.ChannelMessageSend(m.ChannelID, "Must be in voice channel to play music")
					}
					if vid == nil {
						fmt.Println("Error getting video")
						return
					}

					/*err = s.UpdateStatus(0, vid.Title)
					if err != nil {
						fmt.Println(err)		//Again, bad for public bot
						return
					}
					*/

					if !isVConnected {
						rate := time.Second * 2
						throttle := time.Tick(rate)
						for req := range "   " {
							<-throttle
							if req != 0 {
							}
						}
					}

					if !dgvoice.IsThisThingOn() {
						nowPlaying = vid.Title
						err := dgvoice.PlayAudioFile(dgv, file, s)
						defer removeFromQueue()
						nowPlaying = ""
						continue
						if err != nil {
							nowPlaying = ""
							fmt.Println(err)
							return
						}

					}
					if err != nil {
						fmt.Println(err)
					}

				} else {
					fmt.Print(err)
					s.ChannelMessageSend(m.ChannelID, "Must be in voice channel to play music")
				}
			}
		}
	} else if c == prefixChar+"skip" && admin {
		s.ChannelMessageSend(m.ChannelID, "Skipping...")
		dgvoice.KillPlayer()
		//s.UpdateStatus(1, "")
		return
	} else if strings.HasPrefix(c, prefixChar+"disconnect") || strings.HasPrefix(c, prefixChar+"dc") && !isRateLimited(s, d, m){
		if dgvoice.IsThisThingOn() {
			dgvoice.KillPlayer()
		}
		vDisconnect(s, d)
		isVConnected = false
	} else if strings.HasPrefix(c, prefixChar+"streaming") {
		s.UpdateStreamingStatus(3, "Doing dog things", "https://www.twitch.tv/DogBot4Discord")
	} else if strings.HasPrefix(c, prefixChar+"autoplay") && admin {
		if c == prefixChar+"autoplay" {
			dgvoice.ListReady = false

			guild, _ := s.Guild(d.GuildID)
			channel := getCurrentVoiceChannel(m.Author, s, guild)
			if channel != nil {
				vc, err := s.ChannelVoiceJoin(d.GuildID, channel.ID, false, true)
				isVConnected = true
				if err != nil {
					fmt.Println(vc, err)
					isVConnected = false
					return
				}
			} else {
				s.ChannelMessageSend(d.ID, "Must be in a voice channel to use the autoplay feature.")
				return
			}

			lines, err := readLines(APlaylist)
			if err != nil {
				s.ChannelMessageSend(m.ChannelID, err.Error())
				return
			}

			dgv, err := s.ChannelVoiceJoin(d.GuildID, channel.ID, false, true)
			if err != nil {
				fmt.Println(err)
				return
			}

			if !isVConnected {
				rate := time.Second * 2
				throttle := time.Tick(rate)
				for req := range "   " {
					<-throttle
					if req != 0 {
					}
				}
			}
			if !dgvoice.IsThisThingOn() {
				dgvoice.ListReady = true
			} else {
				s.ChannelMessageSend(m.ChannelID, "Not ready to start playlist")
				dgvoice.ListReady = false
				dgvoice.KillPlayer()
			}

			var firstpass= true
			var songmessage *discordgo.Message

			for i := 0; i < len(lines); i++ {
				if dgvoice.ListReady && !dgvoice.IsThisThingOn() {
					url := xurls.Strict().FindString(lines[i])
					fileName := strings.TrimPrefix(url, "https://www.youtube.com/watch?v=")
					file := Folder + fileName + ".mp3"
					youtubeDl(url, m.Message, s)

					vid, err := ytdl.GetVideoInfo(url)
					if err != nil {
						fmt.Println(err, lines)
					}

					if firstpass {
						songmessage, err = s.ChannelMessageSend(d.ID, "Now Playing: `"+vid.Title+"`")
						firstpass = false
					} else if songmessage.ID != "" {
						s.ChannelMessageEdit(songmessage.ChannelID, songmessage.ID, "Now Playing: `"+vid.Title+"`")
					}

					//s.UpdateStatus(0, vid.Title)
					nowPlaying = vid.Title
					err = dgvoice.PlayAudioFile(dgv, file, s)
					nowPlaying = ""
					lineSize := len(lines)
					if i+1 == lineSize {
						i = -1
						continue
						return
					}
					if err != nil {
						continue
						return
					}

				} else {
					return
				}
			}
		}
	} else if strings.HasPrefix(c, prefixChar+"join") && !isRateLimited(s, d, m) {
		if c == prefixChar+"join" {
			guild, _ := s.Guild(d.GuildID)
			channel := getCurrentVoiceChannel(m.Author, s, guild)
			vc, err := s.ChannelVoiceJoin(d.GuildID, channel.ID, false, true)
			isVConnected = true
			if err != nil {
				fmt.Println(vc, err)
				isVConnected = false
			}
		} else {
			cc := strings.TrimPrefix(c, prefixChar+"join ")
			arg := strings.Split(cc, " ")
			vc, err := s.ChannelVoiceJoin(d.GuildID, arg[0], false, false)
			isVConnected = true
			if err != nil {
				isVConnected = false
				fmt.Println(err)
			}
			if vc == nil {
				isVConnected = false
				fmt.Println(vc)
			}

		}
	} else if c == prefixChar+"skiplist" {
		if dgvoice.IsThisThingOn() && isVConnected {
			dgvoice.ListReady = false
			dgvoice.KillPlayer()
			dgvoice.KillPlayer()
			s.ChannelMessageSend(d.ID, "List cleared.")
		}
	} else if strings.HasPrefix(c, prefixChar+"simpask") && !isRateLimited(s, d, m) {
		quest := strings.TrimPrefix(m.Content, prefixChar+"simpask ")
		arg := strings.SplitAfterN(quest, " ", 1)
		fmt.Println(member.User.Username+" simpasked:", arg[0])
		if strings.Contains(strings.ToLower(arg[0]), "gay") || strings.Contains(strings.ToLower(arg[0]), "homo") || strings.Contains(strings.ToLower(arg[0]), "homosexual") || strings.Contains(strings.ToLower(arg[0]), "lesbian") || strings.Contains(strings.ToLower(arg[0]), "fag") || strings.Contains(strings.ToLower(arg[0]), "queer") {
			embed := discordgo.MessageEmbed{
				Title: "",
				Color: 10181046,
				Fields: []*discordgo.MessageEmbedField{
					{Name: member.User.Username + " asked:", Value: arg[0]},
					{Name: "Response:", Value: "Use " + prefixChar + "gay instead"},
				},
			}
			_, err := s.ChannelMessageSendEmbed(d.ID, &embed)
			if err != nil {
				s.ChannelMessageSend(d.ID, formatError(err))
			}
		} else {
			i := rand.Intn(5)
			if strings.Contains(strings.ToLower(arg[0]), "how many") {
				a := rand.Intn(1000)
				embed := discordgo.MessageEmbed{
					Title: "",
					Color: 10181046,
					Fields: []*discordgo.MessageEmbedField{
						{Name: member.User.Username + " asked:", Value: arg[0]},
						{Name: "Response:", Value: strconv.Itoa(a)},
					},
				}
				_, err := s.ChannelMessageSendEmbed(d.ID, &embed)
				if err != nil {
					s.ChannelMessageSend(d.ID, formatError(err))
				}
			} else {
				a := [6]string{"Yes", "No", "Maybe", "Absolutely not", "Impossible", "Perhaps"}
				embed := discordgo.MessageEmbed{
					Title: "",
					Color: 10181046,
					Fields: []*discordgo.MessageEmbedField{
						{Name: member.User.Username + " asked:", Value: arg[0]},
						{Name: "Response:", Value: a[i]},
					},
				}
				_, err := s.ChannelMessageSendEmbed(d.ID, &embed)
				if err != nil {
					s.ChannelMessageSend(d.ID, formatError(err))
				}
			}
		}

	} else if strings.HasPrefix(c, prefixChar+"broom") || strings.HasPrefix(c, prefixChar+"dontbeabroom") && !isRateLimited(s, d, m) {
		s.ChannelMessageSend(d.ID, "https://youtu.be/sSPIMgtcQnU")
	} else if strings.HasPrefix(c, prefixChar+"rick") && !isRateLimited(s, d, m) {
		s.ChannelMessageSend(d.ID, "http://kkmc.info/1LWYru2")
	} else if strings.HasPrefix(c, prefixChar+"vktrs") && !isRateLimited(s, d, m) {
		s.ChannelMessageSend(d.ID, "https://www.youtube.com/watch?v=Iwuy4hHO3YQ")
	} else if strings.HasPrefix(c, prefixChar+"woop") && !isRateLimited(s, d, m) {
		s.ChannelMessageSend(d.ID, "https://www.youtube.com/watch?v=k1Oom5r-cWY")
	} else if strings.HasPrefix(c, prefixChar+"playskip") && !isRateLimited(s, d, m) {
		if !playSkip {
			playSkip = true
		} else {
			playSkip = false
		}

		s.ChannelMessageSend(d.ID, "Allow skipping a video and playing a new one: "+strconv.FormatBool(playSkip)+". Do ***NOT*** use this with playlists. Guaranteed a bad time.")
	} else if strings.HasPrefix(c, prefixChar+"setgame") && m.Author.ID == "157630049644707840" {
		cc := strings.TrimPrefix(m.Content, prefixChar+"setgame ")
		arg := strings.SplitAfterN(cc, " ", 1)
		if arg[0] != prefixChar+"setgame" {
			s.UpdateStatus(0, arg[0])
		} else {
			s.UpdateStatus(1, "")
		}

	} else if strings.HasPrefix(c, prefixChar+"lmgtfy") && !isRateLimited(s, d, m) {
		cc := strings.TrimPrefix(c, prefixChar+"lmgtfy ")
		arg := strings.SplitAfterN(cc, " ", 1)
		if len(arg) == 0 {
			s.ChannelMessageSend(d.ID, "Query is empty.")
		} else if arg[0] == prefixChar+"lmgtfy" {
			s.ChannelMessageSend(d.ID, "Query is empty.")
		} else {
			str := Lreplacer.Replace(arg[0])
			oldUrl := "http://lmgtfy.com/?q=" + str + ""
			url := UrlShortener{}
			url.short(oldUrl, TINY_URL)
			em, _ := s.ChannelMessageSend(m.ChannelID, "<"+url.ShortUrl+">")
			if em == nil {
				fmt.Println("Error shortening URL")
			}
		}
		removeNow(s, m.Message)
	} else if strings.HasPrefix(c, prefixChar+"gay") && !isRateLimited(s, d, m) {
		cc := strings.TrimPrefix(c, prefixChar+"gay ")
		arg := strings.Split(cc, " ")
		i := rand.Intn(100)
		j := strconv.Itoa(i)
		output := ""
		if arg[0] == " " {
			if len(arg) >= 1 {
				output = arg[1]
			}
		} else {
			output = arg[0]
		}
		if len(arg) >= 1 {
			output = arg[len(arg)-1]
			//fmt.Println(len(arg), arg[len(arg)-1])
		}
		if !strings.Contains(cc, "<@") {
			s.ChannelMessageSend(d.ID, "Not sure who test for the gay gene.")
		} else {
			bot_id := strings.TrimPrefix(strings.TrimSuffix(BotID, ">"), "<@")
			if strings.Contains(output, "157630049644707840") {
				s.ChannelMessageSend(m.ChannelID, output+" is 0% gay!")
			} else {
				if strings.Contains(output, bot_id) {
					s.ChannelMessageSend(m.ChannelID, "<@"+BotID+"> is 0% gay!")
				} else {
					if strings.Contains(output, "155481695167053824") {
						s.ChannelMessageSend(m.ChannelID, "<@!155481695167053824> is at least 300% gay!")
					} else {
						s.ChannelMessageSend(m.ChannelID, ""+output+"is "+j+"% gay!")
					}
				}
			}
		}
	} else if strings.HasPrefix(c, prefixChar+"createaccount") && !isRateLimited(s, d, m) {
		err := createAccount(m.Author.ID, 200, d, s)
		if err != nil {
			fmt.Println(err)
		}
	} else if strings.HasPrefix(c, prefixChar+"flip") && !isRateLimited(s, d, m) {
		bet := strings.TrimPrefix(c, prefixChar+"flip ")
		if bet == "" {
			bet = strconv.Itoa(1)
		}
		betInt, err := strconv.Atoi(bet)
		if err != nil {
			betInt = 1
		}
		bal, err := removeCredsBet(m.Author.ID, betInt, d, s)
		if bal == 0 {
			return
		}
	} else if strings.HasPrefix(c, prefixChar+"credits") && !isRateLimited(s, d, m) {
		err := getCredits(m.Author.ID, d, s)
		if err != nil {
			fmt.Println(err)
		}
	} else if strings.HasPrefix(c, prefixChar+"clear") {
		if len(c) < 7 || !canManageMessage(s, m.Author, d) {
		}
		fmt.Println("Clearing messages...")
		args := strings.Split(strings.Replace(c, prefixChar+"clear ", "", -1), " ")
		if len(args) == 0 {
			s.ChannelMessageSend(d.ID, "Invalid parameters")
			fmt.Println("Invalid clear paramters...")
			return
		} else if len(args) == 2 {
			fmt.Println("Clearing messages from " + d.Name + " for user " + member.User.Username)
			if i, err := strconv.ParseInt(args[1], 10, 64); err == nil {
				clearUserChat(int(i+1), d, s, args[0])
				removeLater(s, m.Message)
				return
			}
		} else if len(args) == 1 {
			fmt.Println("Clearing " + args[0] + " messages from " + d.Name + " for user " + member.User.Username)
			if i, err := strconv.ParseInt(args[0], 10, 64); err == nil {
				clearChannelChat(int(i+1), d, s) //+1 because not including the message sent would be silly
				removeLater(s, m.Message)
				return
			}
		}
	} else if strings.HasPrefix(c, prefixChar+"clean") {
		if len(c) < 7 || !canManageMessage(s, m.Author, d) {
		}
		fmt.Println("Clearing messages...")
		args := strings.Split(strings.Replace(c, prefixChar+"clean ", "", -1), " ")
		if len(args) == 0 {
			s.ChannelMessageSend(d.ID, "Invalid parameters")
			fmt.Println("Invalid clear paramters...")
			return
		} else if len(args) == 2 {
			fmt.Println("Cleaning messages from " + d.Name + " for user " + member.User.Username)
			if i, err := strconv.ParseInt(args[1], 10, 64); err == nil {
				clearUserChat(int(i+1), d, s, args[0])
				removeLater(s, m.Message)
				return
			}
		} else if len(args) == 1 {
			fmt.Println("Cleaning " + args[0] + " messages from " + d.Name + " for user " + member.User.Username)
			if i, err := strconv.ParseInt(args[0], 10, 64); err == nil {
				clearChannelChat(int(i+1), d, s)
				removeLater(s, m.Message)
				return
			}
		}
	} else if strings.HasPrefix(c, prefixChar+"info") && !isRateLimited(s, d, m) {
		fmt.Println("Sending info...")
		embed := discordgo.MessageEmbed{
			Title:       "Info",
			Color:       10181046,
			Description: "A rewrite of a rewrite of KookyKraftMC's discord bot, written in Go.",
			URL:         "https://github.com/Time6628/CatBotDiscordGo",
			Fields: []*discordgo.MessageEmbedField{
				{Name: "Servers", Value: strconv.Itoa(len(s.State.Guilds)), Inline: true},
				{Name: "Users", Value: strconv.Itoa(countUsers(s.State.Guilds)), Inline: true},
				{Name: "Channels", Value: strconv.Itoa(countChannels(s.State.Guilds)), Inline: true},
			},
		}
		_, err := s.ChannelMessageSendEmbed(d.ID, &embed)
		if err != nil {
			s.ChannelMessageSend(d.ID, formatError(err))
		}
	} else if strings.HasPrefix(c, prefixChar+"trivia") && admin {
		if triviaStatus {
			s.ChannelMessageSend(d.ID, "Trivia already running.")
		} else {
			fmt.Println("Getting trivia")

			if question, err := trivia.Getter.GetTrivia(1); err == nil {
				triviaStatus = true
				a := append(question.Results[0].IncorrectAnswer, question.Results[0].CorrectAnswer)
				for i := range a {
					j := rand.Intn(i + 1)
					a[i], a[j] = a[j], a[i]
				}
				question.Results[0].Question = Qreplacer.Replace(question.Results[0].Question)
				embedanswers := []*discordgo.MessageEmbedField{}
				if len(a) == 2 {
					embedanswers = []*discordgo.MessageEmbedField{
						{Name: "A", Value: a[0], Inline: true},
						{Name: "B", Value: a[1], Inline: true},
					}
				} else if len(a) == 4 {
					a[0] = Qreplacer.Replace(a[0])
					a[1] = Qreplacer.Replace(a[1])
					a[2] = Qreplacer.Replace(a[2])
					a[3] = Qreplacer.Replace(a[3])

					embedanswers = []*discordgo.MessageEmbedField{
						{Name: "A", Value: a[0], Inline: true},
						{Name: "B", Value: a[1], Inline: true},
						{Name: "C", Value: a[2], Inline: true},
						{Name: "D", Value: a[3], Inline: true},
					}
				}
				embed := discordgo.MessageEmbed{
					Title:       "Trivia",
					Color:       10181046,
					Description: question.Results[0].Question,
					URL:         "https://opentdb.com/",
					Fields:      embedanswers,
				}
				_, err := s.ChannelMessageSendEmbed(d.ID, &embed)
				if err != nil {
					triviaStatus = false
					s.ChannelMessageSend(d.ID, formatError(err))
				}
				fmt.Println(question.Results[0].CorrectAnswer)
				if question.Results[0].CorrectAnswer == "0" {
					fmt.Println("A")
				}
				if question.Results[0].CorrectAnswer == "1" {
					fmt.Println("B")
				}
				if question.Results[0].CorrectAnswer == "2" {
					fmt.Println("C")
				}
				if question.Results[0].CorrectAnswer == "3" {
					fmt.Println("D")
				}
				sendLater(s, d.ID, "The correct answer was: "+question.Results[0].CorrectAnswer)
			} else if err != nil {
				s.ChannelMessageSend(d.ID, formatError(err))
				triviaStatus = false
			}
			doLater(func() {
				triviaStatus = false
			})
		}
	} else if strings.HasPrefix(c, prefixChar+"setcredits") && m.Author.ID == "157630049644707840" {
		credits, err := strconv.Atoi(strings.TrimPrefix(c, prefixChar+"setcredits "))
		if err != nil {
			credits = 200
			return
		}
		setCredits(m.Author.ID, credits)
	} else if strings.HasPrefix(c, prefixChar+"slots") && !isRateLimited(s, d, m) {
		bet := strings.TrimPrefix(c, prefixChar+"slots ")
		if bet == "" {
			bet = strconv.Itoa(1)
		}
		betInt, err := strconv.Atoi(bet)
		if betInt > 500 {
			bet = "500"
			betInt = 500
		}
		if betInt < 1 {
			bet = "1"
			betInt = 1
		}
		if err != nil {
			betInt = 1
		}
		status, err := removeCredsSpin(m.Author.ID, betInt, d, s)
		if err != nil {
			fmt.Println(err)
		}
		if status == 0 {
			fmt.Println(m.Author.Username + " has run the slots command but has no credits. If they are able to still do slots\n this is an issue")
			//Returns that the account is empty
			return
		}

		token := []string{
			":banana:",
			":grapes:",
			":apple:",
			":melon:",
			":moneybag:",
		}
		var result= "LOST"
		pos := [9]string{}
		for i := 0; i >= 0 && i < len(pos); i++ {
			var n int = rand.Intn(len(token))

			pos[i] = token[n]
			if i == len(pos)-1 {
				message, err := s.ChannelMessageSend(d.ID, "**[  :slot_machine: l SLOTS ] **"+
					"\n ------------------ "+
					"\n"+ pos[0]+ " : "+ pos[1]+ " : "+ pos[2]+ "\n"+
					"\n"+ pos[3]+ " : "+ pos[4]+ " : "+ pos[5]+ " <"+ "\n"+
					"\n"+ pos[6]+ " : "+ pos[7]+ " : "+ pos[8]+
					"\n ------------------") //Keeping it organized :)
				if err != nil {
					fmt.Println(err)
				}
				doSoon(func() {
					for i := 0; i >= 0 && i < len(pos); i++ {
						var n int = rand.Intn(len(token))

						pos[i] = token[n]
						if i == len(pos)-1 {
							s.ChannelMessageEdit(d.ID, message.ID, "**[  :slot_machine: l SLOTS ] **"+
								"\n ------------------ "+
								"\n"+ pos[0]+ " : "+ pos[1]+ " : "+ pos[2]+ "\n"+
								"\n"+ pos[3]+ " : "+ pos[4]+ " : "+ pos[5]+ " <"+ "\n"+
								"\n"+ pos[6]+ " : "+ pos[7]+ " : "+ pos[8]+
								"\n ------------------")
						}
					}
					doSoon(func() {
						for i := 0; i >= 0 && i < len(pos); i++ {
							var n int = rand.Intn(len(token))

							pos[i] = token[n]
							if i == len(pos)-1 {
								s.ChannelMessageEdit(d.ID, message.ID, "**[  :slot_machine: l SLOTS ] **"+
									"\n ------------------ "+
									"\n"+ pos[0]+ " : "+ pos[1]+ " : "+ pos[2]+ "\n"+
									"\n"+ pos[3]+ " : "+ pos[4]+ " : "+ pos[5]+ " <"+ "\n"+
									"\n"+ pos[6]+ " : "+ pos[7]+ " : "+ pos[8]+
									"\n ------------------"+
									"\n | : : : **"+ result+ "**  : : : |")
								if pos[3] == pos[4] && pos[4] == pos[5] {
									//Only way to win 3x3x3
									result = "WIN"
									var win int = 2 * betInt
									if pos[3] == ":banana:" {
										win = 2 * betInt
									}
									if pos[3] == ":grapes:" {
										win = 4 * betInt
									}
									if pos[3] == ":apple:" {
										win = 6 * betInt
									}
									if pos[3] == ":melon:" {
										win = 8 * betInt
									}
									if pos[3] == ":moneybag:" {
										win = 10 * betInt
									}
									addCredsSpin(m.Author.ID, win, d, s)
								} else {
									s.ChannelMessageSend(d.ID, m.Author.Username+" used **"+strconv.Itoa(betInt)+"** credit(s) and lost everything.")
								}
							}
						}
					})
				})
			}
		}
	} else if c == prefixChar+"daily" {
		t := time.Now()
		if !redisConnected {
			s.ChannelMessageSend(d.ID, "Redis address has not been given. As a result, economy commands will not work.")
			return
		}

		c, err := redis.Dial("tcp", redisAddr)
		if err != nil {
			return
		}
		bal, err := c.Do("GET", m.Author.ID)
		if bal == nil {
			createAccount(m.Author.ID, 200, d, s)
			c.Do("SET", m.Author.ID+":daily", t.Format("2006-01-02 15:04:05"))
			c.Do("SET", m.Author.ID, 400)
			s.ChannelMessageSend(d.ID, "**"+m.Author.Username+"** has claimed their daily credits.")
			return
		}
		var newBalUint []uint8
		if bal != nil {
			newBalUint = bal.([]uint8)
		} else {
			return
		}
		var newBal = string(newBalUint)
		var newBalInt, convertError = strconv.Atoi(newBal)
		newBal = strconv.Itoa(newBalInt + 200)
		if convertError != nil {
			fmt.Print(convertError)
		}

		lastDailyGet, err := c.Do("GET", m.Author.ID+":daily")
		if lastDailyGet == nil {
			c.Do("SET", m.Author.ID+":daily", t.Format("2006-01-02 15:04:05"))
			c.Do("SET", m.Author.ID, newBal)
			s.ChannelMessageSend(d.ID, "**"+m.Author.Username+"** has claimed their daily credits.")
			return
		}
		layout := "2006-01-02 15:04:05"
		timeLastDaily, err := time.Parse(layout, string(lastDailyGet.([]uint8)))
		if err != nil {
			fmt.Println(err)
		}
		since := time.Since(timeLastDaily)

		sinceString := since.String()

		var sinceInt, err1 = strconv.Atoi(sinceString[:len(sinceString)-13])
		if err1 != nil {
			sinceInt, err1 = strconv.Atoi(sinceString[:len(sinceString)-14])
			if err1 != nil {
				sinceInt, err1 = strconv.Atoi(sinceString[:len(sinceString)-15])
			}
		}
		if strings.ContainsAny(strconv.Itoa(sinceInt), "h") {
			sinceInt, err1 = strconv.Atoi(sinceString[:len(sinceString)-1])
		}
		if err != nil {
			fmt.Println(err)
		}
		sinceInt = sinceInt - 5 //For some reason, it starts with 5 hours already on the clock so they have to be removed
		timeUntil := strconv.Itoa(24 - sinceInt)
		if sinceInt >= 24 {
			fmt.Println(bal)
			fmt.Println(newBal)
			if err != nil {
				fmt.Println(err)
			}
			c.Do("SET", m.Author.ID, newBal)
			c.Do("SET", m.Author.ID+":daily", t.Format("2006-01-02 15:04:05"))
			s.ChannelMessageSend(d.ID, "**"+m.Author.Username+"** has claimed their daily credits.")
		} else {
			s.ChannelMessageSend(d.ID, "Your daily is not ready. Your daily will be ready in **"+timeUntil+"** hour(s).")
		}
	} else if strings.HasPrefix(c, prefixChar+"srsearch") && !isRateLimited(s, d, m) {
		cc := strings.TrimPrefix(c, prefixChar+"srsearch ")
		w, err := GOWikia_Wubsy.NewClient("http://slimerancher.wikia.com/")

		if err != nil {
			fmt.Println(err)
			return
		}

		searchParams := GOWikia_Wubsy.QueryParams{
			Query: cc,
			Lang:  "en",
			Limit: 1,
		}

		results, err := w.SearchList(searchParams)
		if err != nil {
			fmt.Println("Fatal error: " + err.Error())
			return
		}

		if results.Total <= 25 && searchParams.Limit <= 25 {
			totalItems = searchParams.Limit
		} else if results.Total > 25 {
			totalItems = 25
		}

		if totalItems == 1 {
			s.ChannelMessageSend(d.ID, "No results for \""+cc+"\" found.")
			return
		}
		var embed discordgo.MessageEmbed
		if len(results.Items) > 0 {

			articleName = results.Items[0].Title
			articleUrl = results.Items[0].Url
			articleId = results.Items[0].Id

			fmt.Println(articleName, articleId)

			articleParams := GOWikia_Wubsy.GetAsSimpleJsonParams{
				IDs: results.Items[0].Id,
			}

			article, err := w.GetArticleSimpleInfoByID(articleParams)
			if err != nil {
				fmt.Println("Fatal error: " + err.Error())
				return
			}
			resp, erro := http.Get(articleUrl)
			if erro != nil {
				//Do nothing
			}
			//Check table for image, if image cannot be found
			htmlRead, eroo := ioutil.ReadAll(resp.Body)

			if eroo != nil {
				fmt.Println(eroo)
			}

			h := html{}

			erroo := xml.NewDecoder(bytes.NewBuffer(htmlRead)).Decode(&h)
			if erroo != nil {
				fmt.Println("Error: ", erroo)
			}

			fmt.Println(h.Body.Content)
			var articleThumbnail string

			fmt.Println("Article Sections: ", len(article.Sections))
			fmt.Println("Article Images: ", len(article.Sections[0].Images))
			if len(article.Sections[0].Images) > 0 {
				for i := 0; i < len(article.Sections[0].Images); i++ {

					if article.Sections[0].Images[i].Src != "" {
						articleThumbnail = article.Sections[0].Images[i].Src
					} else {
						continue
						return
					}
				}
			} else {
				articleThumbnail = ""
			}

			embedImage := []*discordgo.MessageEmbedImage{
				{URL: articleThumbnail, Width: 250, Height: 250, },
			}

			var articleText string
			if len(article.Sections[0].Content) >= 1 {
				articleText = article.Sections[0].Content[0].Text

				if articleText == "\u00a0" && len(article.Sections[0].Content) >= 1 {
					articleText = article.Sections[0].Content[1].Text
				} else if articleText == "\u00a0" {
					articleText = "Error retrieving article info"
				}
			} else {
				articleText = "Error retrieving article info."
			}

			embed = discordgo.MessageEmbed{
				Title: articleName,
				Color: 10181046,
				URL:   articleUrl,
				Image: embedImage[0],
				Fields: []*discordgo.MessageEmbedField{
					{Name: "Info", Value: articleText, Inline: false, },
				},
			}

		}
		s.ChannelMessageSendEmbed(d.ID, &embed)

	} else if strings.HasPrefix(c, prefixChar+"yikes") && m.Author.ID == "157630049644707840" {
		removeNow(s, m.Message)
		s.ChannelMessageSend(d.ID, "🇾 🇮 🇰 🇪 🇸")
	} else if strings.HasPrefix(c, prefixChar+"volume ") && admin {
		cc := strings.TrimPrefix(c, prefixChar+"volume ")
		newVol, err := strconv.Atoi(cc)
		if err != nil {
			s.ChannelMessageSend(d.ID, "Error setting volume.")
			return
		}

		if newVol > 128 {
			newVol = 128
		}

		if newVol < 0 {
			newVol = 0
		}
		fmt.Println(newVol)
		dgvoice.Volume = newVol
	} else if strings.HasPrefix(c, prefixChar+"playing") && !isRateLimited(s, d, m) {
		if nowPlaying != "" {
			s.ChannelMessageSend(d.ID, nowPlaying)
		} else {
			s.ChannelMessageSend(d.ID, "Not currently playing")
		}
	} else if strings.HasPrefix(c, prefixChar+"queue") {

		if len(queue) != 0 {
			lineText := strings.Join(queue, "\n")
			s.ChannelMessageSend(d.ID, "Songs Currently in queue:\n"+lineText+"")
		} else {
			s.ChannelMessageSend(d.ID, "No songs are currently in the queue")
		}
	} else if strings.HasPrefix(c, prefixChar+"resetpl") && m.Author.ID == "157630049644707840" {
		s.ChannelMessageSend(d.ID, "The queue has been cleared")
		queue = []string{}
	} else if strings.HasPrefix(c, prefixChar+"pause") && admin {
		if !dgvoice.Paused && dgvoice.IsThisThingOn() {
			s.ChannelMessageSend(d.ID, "Pausing...")
			dgvoice.Paused = true
		} else {
			s.ChannelMessageSend(d.ID, "Already paused or not playing")
		}
	} else if strings.HasPrefix(c, prefixChar+"resume") && admin {
		if dgvoice.Paused {
			s.ChannelMessageSend(d.ID, "Resuming...")
			dgvoice.Paused = false
		} else {
			s.ChannelMessageSend(d.ID, "Not paused or already playing")
		}
	} else if strings.HasPrefix(c, prefixChar+"dumpvars") && m.Author.ID == "157630049644707840" {
		var isVConnectedString= strconv.FormatBool(isVConnected)
		var IsSpeakingString= strconv.FormatBool(dgvoice.IsThisThingOn())
		var ListReadyString= strconv.FormatBool(dgvoice.ListReady)

		if err != nil {
			s.ChannelMessageSend(d.ID, "Problem with one or more variable")
			return
		}
		s.ChannelMessageSend(d.ID, "```\n nowPlaying: \""+nowPlaying+"\"\n isVConnected: "+isVConnectedString+"\n dgvoice.IsSpeaking: "+IsSpeakingString+"\n dgvoice.ListReady: "+ListReadyString+"\n dgvoice.Volume: "+strconv.Itoa(dgvoice.Volume)+"```")
	}
}




func removeCredsSpin(id string, bet int, channel *discordgo.Channel, session *discordgo.Session) (status int, err error){
	if !redisConnected {
		session.ChannelMessageSend(channel.ID, "Redis address has not been given. As a result, economy commands will not work.")
		return
	}
	c, err := redis.Dial("tcp", redisAddr)
	if err != nil {
		return
	}
	curCredsGet, err := c.Do("GET", id)
	if curCredsGet == nil{
		createAccount(id, 200, channel, session)
		removeCredsSpin(id, bet, channel, session)
		return
	}
	curCredsByte, err := strconv.Atoi(string(curCredsGet.([]byte)))
	if curCredsByte <= 0 {
		c.Do("SET", id, 0)
		session.ChannelMessageSend(channel.ID, "Insufficient balance.")
		return 0, nil
	}

	if curCredsByte != 0 && curCredsGet != nil && curCredsByte > 0{
		credsNew := curCredsByte - bet
		c.Do("SET", id, credsNew)
		return 1, nil
		//0 is empty 1 is not
	}
	return
	//0 is empty 1 is not
}

func addCredsSpin(id string, win int, channel *discordgo.Channel, session *discordgo.Session) (err error) {
	if !redisConnected {
		session.ChannelMessageSend(channel.ID, "Redis address has not been given. As a result, economy commands will not work.")
		return
	}
	c, err := redis.Dial("tcp", redisAddr)
	if err != nil {
		return err
	}

	curCredsGet, err := c.Do("GET", id)
	curCredsByte, err := strconv.Atoi(string(curCredsGet.([]byte)))
	if win < 0{
		session.ChannelMessageSend(channel.ID, "There was an error setting your balance.\nYour balance has been set to 200")
		win = 200
	}
	c.Do("SET", id, win)
	session.ChannelMessageSend(channel.ID, "<@"+id+"> won "+strconv.Itoa(win)+" credit(s) and now have "+strconv.Itoa(curCredsByte)+" credit(s).")


	return err
}

func removeCredsBet(id string, toRemove int, channel *discordgo.Channel, session *discordgo.Session) (creds int, err error) {
	if !redisConnected {
		session.ChannelMessageSend(channel.ID, "Redis address has not been given. As a result, economy commands will not work.")
		return
	}
	c, err := redis.Dial("tcp", redisAddr)
	if err != nil {
		return
	}
	curCredsGet, err := c.Do("GET", id)
	if curCredsGet == nil{
		createAccount(id, 200, channel, session)
		removeCredsBet(id, toRemove, channel, session)
		return
	}
	curCredsByte, err := strconv.Atoi(string(curCredsGet.([]byte)))
	if curCredsByte == 0 {
		session.ChannelMessageSend(channel.ID, "Insufficient balance.")
		return curCredsByte, nil
	}
	if toRemove <= 0 {
		session.ChannelMessageSend(channel.ID, "Invalid ***___BET___***.")
		return
	}
	curCreds := int(curCredsByte) - toRemove
	if curCreds < 0 {
		curCreds = 0
	}

	c.Do("SET", id, curCreds)

	i := rand.Intn(2)
	if i == 1 {
		curCredsNew := curCreds + toRemove*2
		c.Do("SET", id, curCredsNew)
		session.ChannelMessageSend(channel.ID, "You won the flip and now have "+strconv.Itoa(curCredsNew)+" credits!")
	} else {
		session.ChannelMessageSend(channel.ID, "You lost the flip and now have "+strconv.Itoa(curCreds)+ " credits.")
	}


	return curCredsByte, err
}

func setCredits(id string, setCreds int) (err error){
	if !redisConnected {
		return
	}
	c, err := redis.Dial("tcp", redisAddr)
	if err != nil {
		return err
	}
	_, errDo := c.Do("SET", id, setCreds)
	return errDo
}

func createAccount(id string, defCreds int, channel *discordgo.Channel, session *discordgo.Session) (err error){
	if !redisConnected {
		session.ChannelMessageSend(channel.ID, "Redis address has not been given. As a result, economy commands will not work.")
		return
	}
	c, err := redis.Dial("tcp", redisAddr)
	if err != nil {
		return err
	}
	curCredsGet, err := c.Do("GET", id)
	if curCredsGet == nil {
		_, err := c.Do("SET", id, defCreds)
		if err != nil{
			fmt.Println(err)
		}
		curCreds, err := c.Do("GET", id)
		session.ChannelMessageSend(channel.ID, "Account Created! <@"+id+">, you now have "+string(curCreds.([]byte))+" credits!")
		defer c.Close()
		return err
	} else {
		curCreds, err := c.Do("GET", id)
		curCreds, ok := curCreds.([]byte)
		if !ok{
			fmt.Println(err)
			return err
		}
		session.ChannelMessageSend(channel.ID, "You already have an account, <@" + id + ">. You have "+string(curCreds.([]byte))+" credits.")
		return err
	}
}

func getCredits(id string, channel *discordgo.Channel, session *discordgo.Session) (err error){
	if !redisConnected {
		session.ChannelMessageSend(channel.ID, "Redis address has not been given. As a result, economy commands will not work.")
		return
	}
	c, err := redis.Dial("tcp", redisAddr)
	if err != nil {
		session.ChannelMessageSend(channel.ID, "Could not connect to database.")
		return err
	}
	curCreds, err := c.Do("GET", id)
	if err != nil  {
		fmt.Println(err)
		return err

	} else {
		curCreds, ok := curCreds.([]byte)
		if !ok{
			return err
		}
		session.ChannelMessageSend(channel.ID, "<@" + id + ">, you have " + string(curCreds) + " credits.")
	}
	defer c.Close()
	return err
}

func doLater(i func()) {
	timer := time.NewTimer(time.Minute * 1)
	<- timer.C
	i()
}

func doSoon(i func()) {
	timer := time.NewTimer(time.Second * 2)
	<- timer.C
	i()
}

func countChannels(guilds []*discordgo.Guild) (channels int) {
	for i := 0; i < len(guilds); i++ {
		channels = len(guilds[i].Channels) + channels
	}
	return
}


func filterChannel(id string) (b bool) {
	b = true
	//for all the channels without filters,
	for i := 0; i < len(nofilter); i++ {
		//see if nofilter contains the channel id
		if nofilter[i] == id {
			b = false
			return
		}
	}
	return
}

func countUsers(guilds []*discordgo.Guild) (users int) {
	for i := 0; i < len(guilds); i++ {
		users = guilds[i].MemberCount + users
	}
	return
}

func formatError(err error) string {
	return "```" + err.Error() + "```"
}

func canManageMessage(session *discordgo.Session, user *discordgo.User, channel *discordgo.Channel) bool {
	uPerms, _ := session.UserChannelPermissions(user.ID, channel.ID)
	if (uPerms&discordgo.PermissionManageMessages) == discordgo.PermissionManageMessages {
		return true
	}
	return false
}

func clearChannelChat(i int, channel *discordgo.Channel, session *discordgo.Session) {
	fmt.Println("Clearing channel messages...")
	messages, err := session.ChannelMessages(channel.ID, i, "", "", "")
	if err != nil {
		session.ChannelMessageSend(channel.ID, "Could not get messages.")
		session.ChannelMessageSend(channel.ID, "```" + err.Error() + "```")
		return
	}
	todelete := []string{}
	for i := 0; i < len(messages); i++ {
		message := messages[i]
		todelete = append(todelete, message.ID)
	}
	session.ChannelMessagesBulkDelete(channel.ID, todelete)
	m, err := session.ChannelMessageSend(channel.ID, "Messages removed in channel " + channel.Name)
	if err != nil {
		session.ChannelMessageSend(channel.ID, "```" + err.Error() + "```")
		return
	}
	removeLater(session, m)
}

func clearUserChat(i int, channel *discordgo.Channel, session *discordgo.Session, id string) {
	messages, err := session.ChannelMessages(channel.ID, i, "", "", "")
	if err != nil {
		session.ChannelMessageSend(channel.ID, "Could not get messages.")
		session.ChannelMessageSend(channel.ID, "```" + err.Error() + "```")
		return
	}
	todelete := []string{}
	for i := 0; i < len(messages); i++ {
		message := messages[i]
		if message.Author.ID == id {
			todelete = append(todelete, message.ID)
		}
	}
	session.ChannelMessagesBulkDelete(channel.ID, todelete)
	m, _ := session.ChannelMessageSend(channel.ID, "Messages removed for user <@" + id + "> in channel " + channel.Name)
	removeLater(session, m)
}

func removeLaterBulk(session *discordgo.Session, messages []*discordgo.Message) {
	for _, z := range messages {
		timer := time.NewTimer(time.Second * 5)
		<- timer.C
		session.ChannelMessageDelete(z.ChannelID, z.ID)
	}
}

func alreadyMuted(id string, channel *discordgo.Channel) (b bool) {
	permissions := channel.PermissionOverwrites
	for i := 0; i < len(permissions); i++ {
		permission := permissions[i]
		if permission.ID == id && permission.Type == "member" {
			b = permission.Deny == discordgo.PermissionSendMessages
		}
	}
	return
}

func removeLater(s *discordgo.Session, m *discordgo.Message) {
	timer := time.NewTimer(time.Second * 5)
	<- timer.C
	s.ChannelMessageDelete(m.ChannelID, m.ID)
}

func removeNow(s *discordgo.Session, m *discordgo.Message) {
	timer := time.NewTimer(time.Second * 1)
	<- timer.C
	s.ChannelMessageDelete(m.ChannelID, m.ID)
}

func sendLater(s *discordgo.Session, cid string, msg string) {
	timer := time.NewTimer(time.Minute * 1)
	<- timer.C
	s.ChannelMessageSend(cid, msg)
}

func removeFromQueue(){
	if len(queue) >= 1 {
		queue = queue[1:]
	}
}

func (s *twitchUserSlice) String() string {
	return fmt.Sprintf("%d", *s)
}

func (s *twitchUserSlice) Set(str string) error {
	fmt.Sprintf("%s\n", str)
	*s = append(*s, str)
	return nil
}

//structs
type TwitchConvert struct{
	Status int `json:"Total"`
	Channel string `json:"DisplayName"`
	ProfileImage string `json:"Logo"`
	Title string `json:"Status"`
	Game string `json:"Game"`
}

type twitchUserSlice []string

type html struct {
	Body body `xml:"<aside class=\"portable-infobox pi-background pi-theme-wikia pi-layout-stacked\">>"`
}

type body struct {
	Content string `xml:",innerxml"`
}

type CatResponse struct {
	URL string `json:"file"`
}

type DogResponse struct {
	URL string `json:"url"`
}

type VersionControl struct {
	ver string `json:"version"`
}


type Guild struct {
	ID string `json:"id"`
}

func getJson(url string, target interface{}) error {
	stat, body, err := client.Get(nil, url)
	if err != nil || stat != 200 {
		return errors.New("Could not obtain json response")
	}
	return json.NewDecoder(bytes.NewReader(body)).Decode(target)


}

//Until I can get the import to work on this, I'm plopping it down here
//This is from https://github.com/mikicaivosevic/golang-url-shortener but it's set up as an application


const (
	TINY_URL = 1
	IS_GD    = 2
)

type UrlShortener struct {
	ShortUrl    string
	OriginalUrl string
}

func getResponseData(urlOrig string) string {
	response, err := http.Get(urlOrig)
	if err != nil {
		fmt.Print(err)
	}
	defer response.Body.Close()
	contents, err := ioutil.ReadAll(response.Body)
	return string(contents)
}

func tinyUrlShortener(urlOrig string) (string, string) {
	escapedUrl := url.QueryEscape(urlOrig)
	tinyUrl := fmt.Sprintf("http://tinyurl.com/api-create.php?url=%s", escapedUrl)
	return getResponseData(tinyUrl), urlOrig
}

func isGdShortener(urlOrig string) (string, string) {
	escapedUrl := url.QueryEscape(urlOrig)
	isGdUrl := fmt.Sprintf("http://is.gd/create.php?url=%s&format=simple", escapedUrl)
	return getResponseData(isGdUrl), urlOrig
}

func (u *UrlShortener) short(urlOrig string, shortener int) *UrlShortener {
	switch shortener {
	case TINY_URL:
		shortUrl, originalUrl := tinyUrlShortener(urlOrig)
		u.ShortUrl = shortUrl
		u.OriginalUrl = originalUrl
		return u
	case IS_GD:
		shortUrl, originalUrl := isGdShortener(urlOrig)
		u.ShortUrl = shortUrl
		u.OriginalUrl = originalUrl
		return u
	}
	return u
}

func youtubeDl(url string, m *discordgo.Message, s *discordgo.Session) (io.Reader, error) {

	vid, err := ytdl.GetVideoInfo(url)
	if err != nil {
		removeFromQueue()
		s.ChannelMessageSend(m.ChannelID, "Error getting video info. Is the video age restricted?")
		return nil, err
	}
	var fileName string
	if strings.Contains(url, "https://youtu.be/"){
		fileName = strings.TrimPrefix(url, "https://youtu.be/")
	} else {
		fileName = strings.TrimPrefix(url, "https://www.youtube.com/watch?v=")
	}
	if _, err := os.Stat("download\\" + fileName + ".mp3"); os.IsNotExist(err) {
		if _, err := os.Stat("download\\" + fileName + ".mp3"); err == nil {
			file, err := os.Open("download\\" + fileName + ".mp3")
			if err != nil {
				fmt.Println(err)
			}
			defer file.Close()

			stat, err := file.Stat()

			var strBytes int64
			strBytes = stat.Size()
			if strBytes == 0 {
				s.ChannelMessageSend(m.ChannelID, "File is empty. Redownloading")
				file, err := os.Create("download\\" + fileName + ".mp3")
				err = vid.Download(vid.Formats.Best(ytdl.FormatAudioBitrateKey)[0], file)

				if err != nil {
					s.ChannelMessageSend(m.ChannelID, err.Error())
				}
			}
			if err != nil {
				fmt.Println(err)
			}
		}

		file, writeErr := os.Create("download\\" + fileName + ".mp3")
		downloadErr := vid.Download(vid.Formats.Best(ytdl.FormatAudioBitrateKey)[0], file)

		if writeErr != nil {
			fmt.Println(writeErr)
		}
		if downloadErr != nil {
			fmt.Println(downloadErr)

			removeErr := os.Remove("download\\" + fileName + ".mp3")
			if removeErr != nil {
				//fmt.Println() //eh if it errors it's perfectly fine
			}
			s.ChannelMessageSend(m.ChannelID, downloadErr.Error())
		}
		return nil, nil
	}
	return nil, nil
}

func getCurrentVoiceChannel(user *discordgo.User, session *discordgo.Session, guild *discordgo.Guild) *discordgo.Channel {
	for _, vs := range guild.VoiceStates {
		if vs.UserID == user.ID {
			channel, _ := session.Channel(vs.ChannelID)
			return channel
		}
	}
	return nil
}

func readLines(path string) ([]string, error) {
	file, err := os.Open(path)
	if err != nil {
		fmt.Println(err)
	}
	defer func() {
		file.Close()
	}()

	var lines []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	return lines, scanner.Err()
}

func vDisconnect(s *discordgo.Session, d *discordgo.Channel) {
	if isVConnected {
		guild, _ := s.Guild(d.GuildID)
		gCVC := getCurrentVoiceChannel(Bot, s, guild)
		if gCVC == nil {
			s.ChannelMessageSend(d.ID, "Must be connected to disconnect")
		} else {
			dgvc, err := s.ChannelVoiceJoin(d.GuildID, gCVC.ID, true, true)
			if err != nil {
				fmt.Println(err)
			} else {
				dgvoice.KillPlayer()
				dgvoice.ListReady = false
				dgvc.Disconnect()
			}
		}
	} else {
		s.ChannelMessageSend(d.ID, "Not in a voice channel")
	}
}

func isRateLimited(s *discordgo.Session, d *discordgo.Channel, m *discordgo.MessageCreate) bool {
	if m.Author.ID == "157630049644707840" {
		return false //hehe
	}

	//A really janky way to rate limit

	tNow := time.Now() 																							//Time of type Time
	//tForm := tNow.Format("04:05")																				//Time of type string
	tForm := tNow.Format("Mon Jan 2 15:04:05 -0700 MST 2006")											//Time of type string
	//var userTime = make(map[string]string)
	//var recentMessageTime string
	if userTime[m.Author.ID] == "" {		 																	//Check to see if last message is nil
		//fmt.Println("userTime[m.Author.ID] is nil")
		userTime[m.Author.ID] = tForm 																			//Pack string into map
		return false
		//recentMessageTime = userTime[m.Author.ID]																//Assign map to recentMessageTime
		//fmt.Println(recentMessageTime)
	} else {
		//fmt.Println("userTime[m.Author.ID] is not nil")
		tExpand, err := time.Parse("Mon Jan 2 15:04:05 -0700 MST 2006", userTime[m.Author.ID])			//Change tForm back into tNow
		if err != nil {
			fmt.Println(err)
		}
		fmt.Println(tExpand)
		tSince := time.Now().Sub(tExpand)
		fmt.Println(tSince)
		noPrevEntrTime, err := time.ParseDuration("2562047h47m16.854775807s")
		if err != nil {
			fmt.Println(err)
		}
		if tSince != noPrevEntrTime {
			//tSec := tSince.Seconds()
			//t := tSec.Format("05")
			//tInt := strconv.Atoi(t)
			timeSec := strconv.Itoa(rateLimitTime) + "s"
			rLimitS, err := time.ParseDuration(timeSec)
			if err != nil {
				fmt.Println("Error occurred parsing duration:")
				fmt.Println(err)
				return false
			}
			tNow := time.Now() 														//Time of type Time
			//tForm := tNow.Format("04:05")											//Time of type string
			tForm := tNow.Format("Mon Jan 2 15:04:05 -0700 MST 2006")		//Time of type string
			userTime[m.Author.ID] = tForm											//Pack string into map

			//recentMessageTime = userTime[m.Author.ID]								//Assign map to recentMessageTime

			if tSince > rLimitS {
				s.ChannelMessageSend(d.ID, "You are being rate-limited. Please try again later.")
				//fmt.Println("Time since is greater than the limit")
				return false
			} else {
				return true
			}
		}
		return false
	}
}

//TODO: Phase out dgvoice.IsSpeaking and replace with dgvoice.Run and its state.
//i.e. if dgvoice.Run == nil { fmt.Println("FFmpeg not running") }
