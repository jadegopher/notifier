# Task

Write a small program that uses the library above. It should read stdin and
send new messages every interval (should be configurable). Each line should
be interpreted as a new message that needs to be notified about.
The program should implement graceful shutdown on SIGINT.
Example usage information for clarification purposes (the solution doesn’t
have to reproduce this output):

## Usage 

`notify --url=URL [<flags>]`

## Flags

`--help` Show context-sensitive help (also try `--help-long` and `--help-man`).

`-i`, `--interval=5s` Notification interval

## Example call

`$ notify --url http://localhost:8080/notify < messages.txt`

