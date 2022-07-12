# This is the dev version
I use this to test scripts. If you choose to add this one, don't tell me that it's crashing a lot and whatever. It's designed to be tested which requires me to restart it often. 

# DogBot
Horrible rip-off of CatBot that I add a ton of commands to just for the meme
## Inviting the bot
If you're too tired to mess with stuff or don't know how, invite the dev version of my bot using this [invite link.](https://discordapp.com/oauth2/authorize?client_id=309143062288793600&scope=bot&permissions=268446782) You have to be admin on the discord server you're trying to invite it to. I don't often put it up but ¯\_(ツ)_/¯

###### Me
I'm pretty new to this whole github and Golang thing so don't use big words :stuck_out_tongue:

# Libraries
Requires:
```
github.com/Time6628/OpenTDB-Go
github.com/valyala/fasthttp
github.com/bwmarrin/discordgo
```

# Setup
## Downloading Libraries
Implying you already have Golang installed correctly you should be able to do the following command in command prompt
```
>go get github.com/User/Repo
```
for example, the first one would be
```
>go get github.com/Time6628/OpenTDB-Go
```
It will seem to be doing nothing for awhile but until it asks for a new command, it's best just to wait.

If you get an error about gopath/goroot, you have likely not installed Go correctly.

### Getting a bot token
You can get your bot token from a created bot at the [developer site.](https://discordapp.com/developers/applications/me)
It is important that you get this info because the current state only runs from a discord bot API
#### Notices
There might be an issue with capitalization in your IDE. The `github.com/Time6628/OpenTDB-Go` may act weird so you just have to lowercase the letters into `github.com/time6628/opentdb-go` in the ```import()```

#### Running the bot
The only argument required to run the bot is `-t` which should be followed by the bot token.

Running it without an IDE would require you to still have Golang installed. If installed correctly, open cmd in the folder you have it downloaded to and `shift+right click>Open command window here` then type `go run DogBot.go -t` and your bot token.

If you have an IDE, you can build it into an executable and only have to do `shift+right click>Open command window here` and then type `DogBot.exe -t` and the token.

## Current commands
This version is different because I'm constantly changing how things work. At the time of editing this file, these are the commands.
- .cat < Non-zero positive num up to 15 >
- .doge < Non-zero positive num up to 15 >
- .snek < Non-zero positive num up to 15 >
- .trivia (Only does one question | Answer to question will be a letter TODO: Admin only)
- .mute < @user-id > (Mutes only per text channel) (Admin Only)
- .muteall < @user-id > (Mutes in all text channels) (Admin Only) 
- .enablefilter (Enables chat filter | By default only does 'traps arent gay' and the like. TODO: Admin set filters)
- .removefilter (Disables chat filter)
- .clear < Non-zero positive num >
- .info (Shows bot info. Currently working on displaying both links) (BROKEN)
- .broom (Sends message containing [video](https://youtu.be/sSPIMgtcQnU). Same as .dontbeabroom)
- .rick (Sends message containing [video](https://www.youtube.com/watch?v=dQw4w9WgXcQ))
- .vktrs (Sends message containing [video](https://www.youtube.com/watch?v=Iwuy4hHO3YQ))
- .woop (Removed from current build)
- .lmgtfy < Strings with pluses as spaces or actual spaces as spaces > (Removed embed and now shortens link)
