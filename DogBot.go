package main

import (
	"fmt"
	"github.com/bwmarrin/discordgo"
	"strings"
	"flag"
	"time"
	"regexp"
	"encoding/json"
	"strconv"
	"github.com/valyala/fasthttp"
	"errors"
	"bytes"
	"math/rand"
	"github.com/time6628/opentdb-go"
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
)

func main()  {
	go forever()
	fmt.Println("Starting Dogbot 0.1")

	if token == "" {
		fmt.Println("No token provided. Please run: dogbot -t <bot token>")
		return
	}
	dg, err := discordgo.New("Bot " + token)

	u, err := dg.User("@me")
	if err != nil {
		fmt.Println("error obtaining account details,", err)
	}

	if err != nil {
		fmt.Println("Error creating Discord session: ", err)
		return
	}
	dg.AddHandler(messageCreate)

	BotID = u.ID
	err = dg.Open()
	if err != nil {
		fmt.Println("Could not open Discord session: ", err)
	}
	fmt.Println("DogBot is now running.  Press CTRL-C to exit.")
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

	filters := []*regexp.Regexp{regexp.MustCompile("traps aren't gay"), regexp.MustCompile("\\brape"), regexp.MustCompile("traps are not gay"), regexp.MustCompile("traps arent gay")}
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
		rm, _ := s.ChannelMessageSend(m.ChannelID, "Messaged removed from <@" + m.Author.ID + ">.")
		removeLater(s, rm)
		return
	} else if strings.HasPrefix(c, ".removefilter") {
		if filterChannel(d.ID) == false {
			e, _ := s.ChannelMessageSend(d.ID, "Channel already unfiltered.")
			removeLaterBulk(s, []*discordgo.Message{e, m.Message})
		} else {
			nofilter = append(nofilter, d.ID)
			e, _ := s.ChannelMessageSend(d.ID, "Channel is no longer filtered.")
			removeLaterBulk(s, []*discordgo.Message{e, m.Message})
		}
	} else if strings.HasPrefix(c, ".enablefilter") {
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
	} else if strings.HasPrefix(c, ".dogbot") {
		s.ChannelMessageSend(m.ChannelID, "bork bork beep boop! I am dogbot 0.1!")
		return
	} else if strings.HasPrefix(c, ".mute") && admin {
		cc := strings.TrimPrefix(c, ".mute ")
		if !strings.Contains(cc, "@") {
			s.ChannelMessageSend(d.ID, "Please provide a user to mute!")
			return
		}
		arg := strings.Split(cc, " ")
		fmt.Println(arg[0])
		fmt.Println(cc)
		user_id := strings.TrimPrefix(strings.TrimSuffix(arg[0], ">"), "<@")
		if !alreadyMuted(user_id, d) {
			s.ChannelPermissionSet(d.ID, user_id, "member", 0, discordgo.PermissionSendMessages)
			rm, _ := s.ChannelMessageSend(m.ChannelID, "Muted user " + arg[0] + "!")
			fmt.Println(m.Author.Username + " muted " + user_id)
			b := []*discordgo.Message{rm, m.Message,}
			removeLaterBulk(s, b)
		} else {
			rm, _ := s.ChannelMessageSend(m.ChannelID, "User already muted!")
			b := []*discordgo.Message{rm, m.Message,}
			removeLaterBulk(s, b)
		}
	} else if strings.HasPrefix(c, ".allmute") && admin {
		cc := strings.TrimPrefix(c, ".allmute ")
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
		rm, _ := s.ChannelMessageSend(m.ChannelID, "Muted user " + arg[0] + " in all channels!")
		b := []*discordgo.Message{rm, m.Message,}
		removeLaterBulk(s, b)
		fmt.Println(m.Author.Username + " muted " + user_id + " in all channels.")
	} else if strings.HasPrefix(c, ".donationhelp") {
		s.ChannelMessageSend(m.ChannelID,"If you don't have a rank or perk you purchased please make a forum post here: http://kkmc.info/2du3U2l")
		removeLater(s, m.Message)
	} else if strings.HasPrefix(c, ".cat") {
		fmt.Println(time.Now())
		j := CatResponse{}
		cc := strings.TrimPrefix(c, ".cat ")
		if i, err := strconv.ParseInt(cc, 10, 64); err != nil {
			getJson("http://random.cat/meow", &j)
			s.ChannelMessageSend(d.ID, j.URL)
			fmt.Println(time.Now())
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
			fmt.Println(time.Now())
		}
	} else if strings.HasPrefix(c, ".cat") {
		fmt.Println(time.Now())
		j := CatResponse{}
		cc := strings.TrimPrefix(c, ".cat ")
		if i, err := strconv.ParseInt(cc, 10, 64); err != nil {
			getJson("http://random.cat/meow", &j)
			s.ChannelMessageSend(d.ID, j.URL)
			fmt.Println(time.Now())
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
			fmt.Println(time.Now())
		}
	} else if strings.HasPrefix(c, ".doge") {
		j := DogResponse{}

		cc := strings.TrimPrefix(c, ".doge ")
		if i, err := strconv.ParseInt(cc, 9, 64); err != nil {
			getJson("https://random.dog/woof.json", &j)
			s.ChannelMessageSend(d.ID, j.URL)
		} else {
			if i > 15 || i < 0 {
				i = 15
			}
			e := ""
			for b := int64(0); b < i; b++ {
				getJson("https://random.dog/woof.json", &j)
				e = e + j.URL + " "
			}
			s.ChannelMessageSend(d.ID, e)
		}
	} else if strings.HasPrefix(c, ".broom") || strings.HasPrefix(c, ".dontbeabroom") {
		s.ChannelMessageSend(d.ID, "https://youtu.be/sSPIMgtcQnU")
	} else if strings.HasPrefix(c, ".rick") {
		s.ChannelMessageSend(d.ID, "http://kkmc.info/1LWYru2")
	} else if strings.HasPrefix(c, ".vktrs") {
		s.ChannelMessageSend(d.ID, "https://www.youtube.com/watch?v=Iwuy4hHO3YQ")
	} else if strings.HasPrefix(c, ".gay") {
		cc := strings.TrimPrefix(c, ".gay ")
		arg := strings.Split(cc, " ")
		i := rand.Intn(100)
		j := strconv.Itoa(i)
		fmt.Println(j)
		if !strings.Contains(cc, "@") {
			s.ChannelMessageSend(d.ID, "Not sure who test for the gay gene.") } else {
			//		user_id := strings.TrimPrefix(strings.TrimSuffix(arg[0], ">"), "<@")
			if strings.Contains(arg[0], "157630049644707840") {
				rm, _ := s.ChannelMessageSend(m.ChannelID, "<@!157630049644707840> is 0% gay!")
				fmt.Println(rm)
			} else {
				rm, _ := s.ChannelMessageSend(m.ChannelID, ""+arg[0]+"is " +j+ "% gay!")
				fmt.Println(rm)
				fmt.Println(arg[0])
			} }
	} else if strings.HasPrefix(c, ".clear") {
		if len(c) < 7  || !canManageMessage(s, m.Author, d) {
		}
		fmt.Println("Clearing messages...")
		args := strings.Split(strings.Replace(c, ".clear ", "", -1), " ")
		if len(args) == 0 {
			s.ChannelMessageSend(d.ID, "Invalid parameters")
			fmt.Println("Invalid clear paramters...")
			return
		} else if len(args) == 2 {
			fmt.Println("clearing messages from " + d.Name + " for user " + member.User.Username)
			if i, err := strconv.ParseInt(args[1], 10, 64); err == nil {
				clearUserChat(int(i), d, s, args[0])
				removeLater(s, m.Message)
				return
			}
		} else if len(args) == 1 {
			fmt.Println("clearing " + args[0] + " messages from " + d.Name + " for user " + member.User.Username)
			if i, err := strconv.ParseInt(args[0], 10, 64); err == nil {
				clearChannelChat(int(i), d, s)
				removeLater(s, m.Message)
				return
			}
		}
	} else if strings.HasPrefix(c, ".info") {
		fmt.Println("Sending info...")
		embed := discordgo.MessageEmbed{
			Title: "Info",
			Color: 10181046,
			Description: "A rewrite of a rewrite KookyKraftMC discord bot, written in Go.",
			URL: "https://github.com/Time6628/CatBotDiscordGo https://github.com/Wubsy/DogBot",
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
	} else if strings.HasPrefix(c, ".trivia") && admin {
		fmt.Println("Getting trivia")
		if question, err := trivia.Getter.GetTrivia(1); err == nil {
			a := append(question.Results[0].IncorrectAnswer, question.Results[0].CorrectAnswer)
			for i := range a {
				j := rand.Intn(i + 1)
				a[i], a[j] = a[j], a[i]
				fmt.Println(j)
			}
			embedanswers := []*discordgo.MessageEmbedField{}
			if len(a) == 2 {
				embedanswers = []*discordgo.MessageEmbedField{
					{Name: "A", Value: a[0], Inline: true},
					{Name: "B", Value: a[1], Inline: true},
				}
			} else if len(a) == 4 {
				embedanswers = []*discordgo.MessageEmbedField{
					{Name: "A", Value: a[0], Inline: true},
					{Name: "B", Value: a[1], Inline: true},
					{Name: "C", Value: a[2], Inline: true},
					{Name: "D", Value: a[3], Inline: true},
				}
			}
			embed := discordgo.MessageEmbed{
				Title: "Trivia",
				Color: 10181046,
				Description: question.Results[0].Question,
				URL: "https://opentdb.com/",
				Fields: embedanswers,
			}
			_, err := s.ChannelMessageSendEmbed(d.ID, &embed)
			if err != nil {
				s.ChannelMessageSend(d.ID, formatError(err))
			}
			sendLater(s, d.ID, "The correct answer was: " + question.Results[0].CorrectAnswer)
		} else if err != nil {
			s.ChannelMessageSend(d.ID, formatError(err))
		}
	}  else if strings.HasPrefix(c, ".restart") && admin {
		s.ChannelMessageSend(d.ID, "Restarting   :wave:")
		main()
	}
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
	messages, err := session.ChannelMessages(channel.ID, i, "", "")
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
	messages, err := session.ChannelMessages(channel.ID, i, "", "")
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

//func ParseInt(s string, base int, bitSize int) (i int64, err error)

func getJson(url string, target interface{}) error {
	stat, body, err := client.Get(nil, url)
	if err != nil || stat != 200 {
		return errors.New("Could not obtain json response")
	}
	return json.NewDecoder(bytes.NewReader(body)).Decode(target)

	/*
	resp, err := httpClient.Get(url)
	if err != nil {
		panic(err.Error())
	}
	defer resp.Body.Close()
	return json.NewDecoder(resp.Body).Decode(target)
	*/
}
