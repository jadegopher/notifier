# Task

Write a small program that uses the library above. It should read stdin and
send new messages every interval (should be configurable). Each line should
be interpreted as a new message that needs to be notified about.
The program should implement graceful shutdown on SIGINT.
Example usage information for clarification purposes (the solution doesn’t
have to reproduce this output):

## Usage 

`notify -url=http://localhost:8080/notify -i=5s`

## Flags

`--help` Show context-sensitive help.

`-i` duration

    Notification interval (default 5s)

`-url` string

    Target URL for notifications (default "http://localhost:8080/notify")

## Example call

`$ notify -url=http://localhost:8080/notify -i=5s < test.txt`
