# Important
Read everything in the readme before asking for help, please
# DogBot
Horrible rip-off of CatBot that I add a ton of commands to just for the meme
## Inviting the bot
If you're too tired to mess with stuff or don't know how, invite my bot using this [invite link.](https://discordapp.com/oauth2/authorize?client_id=269321947278606336&scope=bot&permissions=268446782) You have to be admin on the discord server you're trying to invite it to. I don't often put it up but ¯\\_(ツ)_/¯



# Libraries
Requires:
```
github.com/Time6628/OpenTDB-Go
github.com/rylio/ytdl
github.com/valyala/fasthttp
github.com/bwmarrin/discordgo
github.com/mvdan/xurls
github.com/Wubsy/dgvoice
```
It now also requires you to have ffmpeg in the directory of the program.
# The audio plays at 100% as I have yet to find a way to change this. You MUST turn it down client-side or say goodbye to your ears
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
 - cat < Non-zero positive num up to 15 >
 - doge < Non-zero positive num up to 15 >
 - snek < Non-zero positive num up to 15 >
 - gay < @user-id > (Still in development)
 - trivia (Only does one question | Answer to question will be a letter ~~TODO: Admin only~~)
 - mute < @user-id > (Mutes only per text channel) (Admin Only)
 - allmute < @user-id > (Mutes in all text channels) (Admin Only) 
 - enablefilter (Enables chat filter | By default only does 'traps arent gay' and the like. TODO: Admin set filters)
 - removefilter (Disables chat filter)
 - clear < Non-zero positive num >
 - info (Shows bot info. Currently working on displaying both links)
 - broom (Sends message containing [video](https://youtu.be/sSPIMgtcQnU). Same as .dontbeabroom)
 - rick (Sends message containing [video](https://www.youtube.com/watch?v=dQw4w9WgXcQ))
 - vktrs (Sends message containing [video](https://www.youtube.com/watch?v=Iwuy4hHO3YQ))
 - woop (Sends message containing [video](https://www.youtube.com/watch?v=k1Oom5r-cWY))
 - simpask < Yes/No question >
 - lmgtfy < string >
 - setgame < string > (Sets game in bot's profile)
 - streaming (Currently only goes to my stream TODO: Custom streams) (TODO: Admin only)
 - setcredits < int > (TODO: Remove this command and add a daily collection system)
 - flip < int > (TODO: Add better support for nil arguments)
 - credits (Displays the credits associated with a user's id)
 
 --Voice Chat--
 
 - .play < YouTube link starting with https://www.youtube.com/ > (Downloads and plays a video in format https://youtube.com/watch?v-) 
 - .skip (Skips currently playing video)
 - .disconnect (Leaves the user's voice channel)
 - .fplay <file name without .mp3> (Only supports mp3 at the moment)
  
# Using the Dev version
[You can find the dev version here](https://github.com/Wubsy/DogBot/tree/dev)
It's probably going to be broken most of the time. Make sure you read all of that one's readme

## Launching from executable
As of version 0.5.7, I have started building my go files into executables. Because the executables are already built, you will have very little freedom until future updates. 
#
You can use a .bat file and put something like `BotApp-0.0.1.exe -t` followed by your bot token in it and run it in the same directory as the executable. 

[Download Generic Executable or Premade DogBot](http://willbusby.us/downloads/)
