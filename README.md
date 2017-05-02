# DogBot
Horrible rip-off of CatBot that I add a ton of commands to just for the meme

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
