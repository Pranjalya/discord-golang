# Discord Link Reader Bot
A simplistic Discord bot in Golang to read links and generate PDFs.

## Features
- Read https:// links and check whether it exists
- Get PDF of that link and reply to the message containing that link with that PDF
- Activate or deactivate the bot if the person is having role of Administrator in a channel
- Print GitHub links in console and do not generate its PDF

## Steps
- Update the `config.json` file to provide Discord Bot Token, Administrator usernames and Backend URL (By default it is http://0.0.0.0:3000)
- Generate build of the project using command
```
> go get
> go build -o reader_bot
```
- Initialize the `Gotenberg` backend which is Docker-based API to generate PDFs by running *(It starts backend on port 3000 of local machine, which provides the URL for Backend URL in configuration file)*
```
> docker run --rm -p 3000:3000 thecodingmachine/gotenberg:6
```
**[ NOTE ]** :  Other ways of running the image on a host can be used as well.
- Start the application by executing the file.
    - For windows, start the executable file `reader_bot.exe`
    - For linux, start the exeutable file `./reader_bot`
- If administrator, use commands
    - `!activate` to activate the bot in a channel
    - `!deactivate` to deactivate the bot in a channel
