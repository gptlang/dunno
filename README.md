# dunno
AI powered command line

## Installation
`go install github.com/gptlang/dunno@latest`

### Requirements
- [Go](https://go.dev/)
- `~/go/bin` in `$PATH`

### Authentication
You only need to set it once. It saves this key to config file. To change key, just do this again before running it.
`export OPENAI_KEY="sk-..."`

## Usage
```
$ dunno
Usage: dunno <prompt>
Example: dunno list tcp network connections
Enter x to interrupt or enter to allow the assistant to run the command
```
