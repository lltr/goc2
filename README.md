# goc2

Simple proof of concept project written in Go.

Run agent: `./agent.bin <C2 IP>`
Run C2: `./c2-server.bin <Bot Token> <Guild ID> <Channel ID>`

Bot commands summary:
* `/cmd` - Returns immediate command
* `/wcmd` - Returns waitable command line by line
* `/agents` - Returns list of connected agent identifiers