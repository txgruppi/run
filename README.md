# run

`run` replaces *tokens* in a config file template by the values of *environment variables* with the same as as the tokens, saves everything in a new config file and it executes a command.

It was designed to be used in *docker containers* where a config file should receive values from the *environment variables* before running the container's command.
