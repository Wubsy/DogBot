package main

import (
	"fmt"
	"strings"
	"flag"
	"time"
	"regexp"
	"encoding/json"
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
	//"github.com/bwmarrin/discordgo" //Disabled because my fork has more functionality
	"github.com/rylio/ytdl"
	"github.com/valyala/fasthttp"
	"github.com/Time6628/OpenTDB-Go"
	"github.com/mvdan/xurls"
	"github.com/Wubsy/dgvoice"
	"github.com/bwmarrin/discordgo"
	//"github.com/gorilla/websocket"
)

func init() {
	flag.StringVar(&token, "t", "", "Bot Token")
	flag.Parse()
}

var (
	token string
	BotID string
	client = fasthttp.Client{ReadTimeout: time.Second * 10, WriteTimeout: time.Second * 10}
	trivia = OpenTDB_Go.New(client)
	nofilter []string
	dgv *discordgo.VoiceConnection
	Folder = "download/"
	prefixChar = "." // Don't use  # and @ because it might mess with channels
	Qreplacer = strings.NewReplacer("&quot;", "\"", "&#039;", "'")
	Lreplacer = strings.NewReplacer(" ", "+")
	version = "0.6.5"
	isVConnected = false
	APlaylist = "autoplaylist.txt"
	triviaStatus = false
	playSkip = true
	listReady = true
	Bot *discordgo.User
)

func main()  {
	go forever()

	url := "http://bots.willbusby.us/DogBotVer"
	resp, err := http.Get(url)
	if err != nil {
		fmt.Println(err)
	} else {
		defer resp.Body.Close()
	}
	html, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Printf("%s\n", html)

	fmt.Println("Starting Dogbot: "+version)

	if token == "" {
		fmt.Println("No token provided. Please run: dogbot -t <bot token>")
		return
	}
	dg, err := discordgo.New("Bot " + token)

	u, err := dg.User("@me")
	if err != nil {
		fmt.Println("Error obtaining account details,", err)
	}

	if err != nil {
		fmt.Println("Error creating Discord session: ", err)
		return
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
	} else if strings.HasPrefix(c, prefixChar+"help") {
		commands := []string{prefixChar + "removefilter",
				     prefixChar + "enabledfilter",
				     prefixChar + "dogbot",
				     prefixChar + "mute",
				     prefixChar + "allmute",
				     prefixChar + "cat",
				     prefixChar + "doge",
				     prefixChar + "leave",
				     prefixChar + "fplay",
				     prefixChar + "csay",
				     prefixChar + "play",
				     prefixChar + "skip",
				     prefixChar + "volume",
				     prefixChar + "disconnect",
				     prefixChar + "streaming",
				     prefixChar + "simpask",
				     prefixChar + "lmgtfy",
				     prefixChar + "gay",
				     prefixChar + "clean",
				     prefixChar + "info",
				     prefixChar + "playskip",
				     prefixChar + "skiplist",
				     prefixChar + "trivia", }

		commandsEmbedPrimerTeb := []*discordgo.MessageEmbedField{}
		for i := 0; i >= 0 && i < len(commands); i++ {
			commandsEmbedPrimer := []*discordgo.MessageEmbedField{
				{Name: strconv.Itoa(i+1) + ".", Value: commands[i], Inline: true},
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
				s.ChannelMessageSendEmbed(d.ID, &embed)
				s.ChannelMessageSendEmbed(d.ID, &embedLink)
			} else {
				i = i
			}
		}

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
	} else if strings.HasPrefix(c, prefixChar+"dogbot") {
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
			getJson("http://random.cat/meow", &j)
			s.ChannelMessageSend(d.ID, j.URL)
		} else {
			if i > 15 || i < 0 {
				i = 15
			}
			e := ""
			for b := int64(0); b < i; b++ {
				getJson("http://random.cat/meow", &j)
				e = e + j.URL + " "
			}
			s.ChannelMessageSend(d.ID, e)
		}
	} else if strings.HasPrefix(c, prefixChar+"cat") {
		j := CatResponse{}
		cc := strings.TrimPrefix(c, prefixChar+"cat ")
		if i, err := strconv.ParseInt(cc, 10, 64); err != nil {
			getJson("http://random.cat/meow", &j)
			s.ChannelMessageSend(d.ID, j.URL)
		} else {
			if i > 15 || i < 0 {
				i = 15
			}
			e := ""
			for b := int64(0); b < i; b++ {
				getJson("http://random.cat/meow", &j)
				e = e + j.URL + " "
			}
			s.ChannelMessageSend(d.ID, e)
		}
	} else if strings.HasPrefix(c, prefixChar+"doge") {
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
			if i > 15 || i < 0 {
				i = 15
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
	} else if strings.HasPrefix(c, "who's a good boy") {
		s.ChannelMessageSend(d.ID, "ME ME ME <@"+m.Author.ID+">")
	} else if strings.HasPrefix(c, prefixChar+"leave") {
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
			if err != nil {
				s.ChannelMessageSend(m.ChannelID, "Must be in voice channel to play music")
			} else {
				if !dgvoice.IsSpeaking {
					url := "https://www.youtube.com/watch?v=" + arg[0]
					vid, err := ytdl.GetVideoInfo(url)
					if err != nil {
						fmt.Println(err)
					}
					if vid.Title == "" {
						fmt.Println("Streaming " + arg[0] + ".mp3")
						s.UpdateStatus(0, "Streaming "+arg[0]+".mp3")
					} else {
						s.UpdateStatus(0, "Streaming "+vid.Title)
					}

					dgvoice.PlayAudioFile(dgv, "download\\"+arg[0]+".mp3", s)
				} else {
					fmt.Println("Error playing file")
				}
			}

		} else {
			s.ChannelMessageSend(m.ChannelID, "File does not exist or "+arg[0]+".mp3 is not valid")
		}
	} else if strings.HasPrefix(c, prefixChar+"csay") && m.Author.ID == "157630049644707840" {
		removeNow(s, m.Message)
		cc := strings.TrimPrefix(m.Content, prefixChar+"csay ")
		chann := strings.SplitAfter(cc, " ")
		trimChann := strings.TrimPrefix(m.Content, prefixChar+"csay "+chann[0])

		_, err := s.ChannelMessageSend(chann[0], trimChann)
		if err != nil {
			fmt.Println(err)
		}

	} else if strings.HasPrefix(c, prefixChar+"play ") && admin {
		pp := strings.TrimPrefix(c, prefixChar+"play ")
		if !strings.Contains(pp, "https://www.youtube.com/") {
			s.ChannelMessageSend(m.ChannelID, "Must be from`https://www.youtube.com/`")
		} else {
			arg := strings.Split(pp, " ")
			url := xurls.Strict.FindString(m.Content)
			//s.ChannelMessageSend(m.ChannelID, "Downloading `"+arg[0]+"`")
			youtubeDl(url, m.Message, s)
			if err != nil {
				fmt.Println(err)
			}
			vid, err := ytdl.GetVideoInfo(url)
			if err != nil {
				fmt.Println(err)
			}

			fileName := strings.TrimPrefix(arg[0], "https://www.youtube.com/watch?v=")
			file := Folder + fileName + ".mp3"
			guild, _ := s.Guild(d.GuildID)

			channel := getCurrentVoiceChannel(m.Author, s, guild)
			if channel != nil {
				dgv, err := s.ChannelVoiceJoin(d.GuildID, channel.ID, false, true)
				if err != nil {
					s.ChannelMessageSend(m.ChannelID, "Must be in voice channel to play music")
				}
				if vid == nil {
					return
				}
				err1 := s.UpdateStatus(0, "Streaming "+vid.Title)
				if err1 != nil {
					fmt.Println(err1)
					return
				}
				if playSkip && dgvoice.IsSpeaking {
					dgvoice.KillPlayer()
					dgvoice.PlayAudioFile(dgv, file, s)
				} else {
					dgvoice.PlayAudioFile(dgv, file, s)
				}
				return

				if err != nil {
					fmt.Println(err)
				}
			} else {
				fmt.Print(err)
				s.ChannelMessageSend(m.ChannelID, "Must be in voice channel to play music")
			}
		}
		if !strings.Contains(pp, " ") {
			/*s.ChannelMessageSend(d.ID, "Starting autoplaylist")*/ }
	} else if c == prefixChar+"skip" && admin {
		if dgvoice.IsSpeaking {
			s.ChannelMessageSend(m.ChannelID, "Skipping...")
			dgvoice.KillPlayer()
			dgvoice.IsSpeaking = false
			s.UpdateStatus(1, "")
			return
		}
		if dgvoice.IsSpeaking == false {
			s.ChannelMessageSend(m.ChannelID, "Not currently playing")
			s.UpdateStatus(1, "")
		}
	} else if strings.HasPrefix(c, prefixChar+"disconnect") {
		if dgvoice.IsSpeaking {
			dgvoice.KillPlayer()
			dgvoice.IsSpeaking = false
		}
		vDisconnect(s, d, m)
	} else if strings.HasPrefix(c, prefixChar+"streaming") {
		s.UpdateStreamingStatus(3, "Doing dog things", "https://www.twitch.tv/DogBot4Discord")
	} else if strings.HasPrefix(c, prefixChar+"autoplay") && admin {
		if c == prefixChar+"autoplay" {
			guild, _ := s.Guild(d.GuildID)
			channel := getCurrentVoiceChannel(m.Author, s, guild)
			vc, err := s.ChannelVoiceJoin(d.GuildID, channel.ID, false, true)
			isVConnected = true
			if err != nil {
				fmt.Println(vc, err)
				isVConnected = false
			}
			lines, err := readLines(APlaylist)
			if err != nil {
				s.ChannelMessageSend(m.ChannelID, err.Error())
			}
			dgv, err := s.ChannelVoiceJoin(d.GuildID, channel.ID, false, true)
			if err != nil {
				fmt.Println(err)
			}
			if !dgvoice.IsSpeaking {
				listReady = true
				s.ChannelMessageSend(m.ChannelID, "Starting playlist")
			} else {
				s.ChannelMessageSend(m.ChannelID, "Not ready to start playlist")
			}
			for i, line := range lines {
				if listReady && !dgvoice.IsSpeaking{
					url := xurls.Strict.FindString(lines[i])
					fileName := strings.TrimPrefix(url, "https://www.youtube.com/watch?v=")
					file := Folder + fileName + ".mp3"
					youtubeDl(url, m.Message, s)
					if err != nil {
						fmt.Println(err)
					}
					vid, err := ytdl.GetVideoInfo(url)
					if err != nil {
						fmt.Println(err, line)
					}
					s.UpdateStatus(0, vid.Title)
					dgvoice.PlayAudioFile(dgv, file, s)
				} else {
					return
				} //Sloppy mess. Did this at 1 AM and can't think but I know it goes here somewhere :P
				dgvoice.IsSpeaking = false
			}
			dgvoice.IsSpeaking = false
		}
	} else if strings.HasPrefix(c, prefixChar+"join") {
		guild, _ := s.Guild(d.GuildID)
		channel := getCurrentVoiceChannel(m.Author, s, guild)
		vc, err := s.ChannelVoiceJoin(d.GuildID, channel.ID, false, true)
		isVConnected = true
		if err != nil {
			fmt.Println(vc, err)
			isVConnected = false
		} else {
		cc := strings.TrimPrefix(c, prefixChar+"join ")
		arg := strings.Split(cc, " ")
		vc, err := s.ChannelVoiceJoin(d.GuildID, arg[0], false, false)
		isVConnected = true
		if err != nil {
			isVConnected = false
			fmt.Println(vc, err)
		}

		}
	} else if c == prefixChar+"skiplist" {
		if dgvoice.IsSpeaking && isVConnected {
			listReady = false
			dgvoice.KillPlayer()
		}
	} else if strings.HasPrefix(c, prefixChar+"simpask") {
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

	} else if strings.HasPrefix(c, prefixChar+"broom") || strings.HasPrefix(c, prefixChar+"dontbeabroom") {
		s.ChannelMessageSend(d.ID, "https://youtu.be/sSPIMgtcQnU")
	} else if strings.HasPrefix(c, prefixChar+"rick") {
		s.ChannelMessageSend(d.ID, "http://kkmc.info/1LWYru2")
	} else if strings.HasPrefix(c, prefixChar+"vktrs") {
		s.ChannelMessageSend(d.ID, "https://www.youtube.com/watch?v=Iwuy4hHO3YQ")
	} else if strings.HasPrefix(c, prefixChar+"woop") {
		s.ChannelMessageSend(d.ID, "https://www.youtube.com/watch?v=k1Oom5r-cWY")
	} else if strings.HasPrefix(c, prefixChar+"playskip") {
		if !playSkip {playSkip = true} else {playSkip = false}

		s.ChannelMessageSend(d.ID, "Allow skipping a video and playing a new one: "+strconv.FormatBool(playSkip)+". Do ***NOT*** use this with playlists. Guaranteed a bad time.")
	} else if strings.HasPrefix(c, prefixChar+"setgame") &&  m.Author.ID == "157630049644707840" {
		cc := strings.TrimPrefix(m.Content, prefixChar+"setgame ")
		arg := strings.SplitAfterN(cc, " ", 1)
		if arg[0] != prefixChar+"setgame" {
		s.UpdateStatus(0, arg[0])
		} else {
		s.UpdateStatus(1, "")
		}

		} else if strings.HasPrefix(c, prefixChar+"lmgtfy") {
			cc := strings.TrimPrefix(c, prefixChar+"lmgtfy ")
			arg := strings.SplitAfterN(cc, " ", 1)
			if len(arg) == 0 {
				s.ChannelMessageSend(d.ID, "Query is empty.")
			} else if arg[0] == prefixChar+"lmgtfy" {
				s.ChannelMessageSend(d.ID, "Query is empty.")
			} else {
				str := Lreplacer.Replace(arg[0])
				oldUrl := "http://lmgtfy.com/?q="+str+""
				url := UrlShortener{}
				url.short(oldUrl, TINY_URL)
				fmt.Println(url.ShortUrl)
				fmt.Println(url.OriginalUrl)
				em, _ := s.ChannelMessageSend(m.ChannelID, "<"+url.ShortUrl+">")
				fmt.Println(em)
			}
			removeNow(s, m.Message)
		}else if strings.HasPrefix(c, prefixChar+"gay") {
			cc := strings.TrimPrefix(c, prefixChar+"gay ")
			arg := strings.Split(cc, " ")
			i := rand.Intn(100)
			j := strconv.Itoa(i)
			if !strings.Contains(cc, "<@") {
				s.ChannelMessageSend(d.ID, "Not sure who test for the gay gene.")
			} else {
				bot_id := strings.TrimPrefix(strings.TrimSuffix(BotID, ">"), "<@")
				if strings.Contains(arg[0], "157630049644707840") {
					s.ChannelMessageSend(m.ChannelID, arg[0]+" is 0% gay!")
				} else {
					if strings.Contains(arg[0], bot_id) {
						s.ChannelMessageSend(m.ChannelID, "<@"+BotID+"> is 0% gay!")
					} else {
						if strings.Contains(arg[0], "155481695167053824") {
							s.ChannelMessageSend(m.ChannelID, "<@!155481695167053824> is at least 300% gay!")
						} else {
							s.ChannelMessageSend(m.ChannelID, ""+arg[0]+"is "+j+"% gay!")
						}
					}
				}
			}
		} else if strings.HasPrefix(c, prefixChar+"clear") {
			if len(c) < 7  || !canManageMessage(s, m.Author, d) {
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
		if len(c) < 7  || !canManageMessage(s, m.Author, d) {
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
	} else if strings.HasPrefix(c, prefixChar+"info") {
			fmt.Println("Sending info...")
			embed := discordgo.MessageEmbed{
				Title: "Info",
				Color: 10181046,
				Description: "A rewrite of a rewrite of KookyKraftMC's discord bot, written in Go.",
				URL: "https://github.com/Time6628/CatBotDiscordGo",
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
			doLater(func (){
				triviaStatus = false
			})
		}
	}

}

func doLater(i func()) {
	timer := time.NewTimer(time.Minute * 1)
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

//structs
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
		s.ChannelMessageSend(m.ChannelID, err.Error())
	}

	fileName := strings.TrimPrefix(url, "https://www.youtube.com/watch?v=")

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

		file, err := os.Create("download\\" + fileName + ".mp3")
				err1 := vid.Download(vid.Formats.Best(ytdl.FormatAudioBitrateKey)[0], file)
		if err1 != nil {
			s.ChannelMessageSend(m.ChannelID, err.Error())
		}
		if err != nil {
			s.ChannelMessageSend(m.ChannelID, err.Error())
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

func vDisconnect(s *discordgo.Session, d *discordgo.Channel, m *discordgo.MessageCreate) {
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
				dgvc.Disconnect()
				dgvoice.IsSpeaking = false
			}
		}
	} else {
		s.ChannelMessageSend(d.ID, "Not in a voice channel")
	}
}
