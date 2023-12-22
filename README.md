# Not IRC Server

## What is this?

Not an IRC server, let me tell you that much...

This is a bare bones, absolute minimum 'educational' server for proxying messages between some clients. The idea is that multiple clients can login into this server, retrieve the last few messages, and chat.

## Quality

This isn't a project thats meant for production. But, if you need to run a quick thing with a couple of friends, then this works! Its meant to interface with [the client](https://github.com/ShadiestGoat/notIRCClient)

## How to use

1. Run `go install github.com/shadiestgoat/notIRCServer@latest` - installs the executable of this server
2. Create a user config (more on this later)
3. Run `notIRCServer`

### User Config

Since V2, this app requires pre-defined users. An example can be found in `users.example.yaml`. By default, the file read for this is `users.yaml`, although this can be changed with the `-u` flag (eg. `-f ./myEpicUsers.yaml`).

The users file consists of a map of `username` -> a user configuration. The below table defines keys of the user configuration:

|      key      |          default           | info                                                                                                                      |
| :-----------: | :------------------------: | :------------------------------------------------------------------------------------------------------------------------ |
|     token     |     :warning: required     | A password for the user. This is a required key. It could also be in the form of `env:ENV_VAR_NAME` to load from env vars |
|     perms     | empty list (default perms) | A list of permissions. See the below table for reference                                                                  |
|     color     |            #fff            | The user's color. Can be any hex string, with or without prefixes (0x & # are supported)                                  |
| readWhispers  |  A list of just this user  | A list of users who's whispers this user hears                                                                            |
| writeWhispers |             \*             | A list of users who this user can whisper to                                                                              |
|    hidden     |           false            | If true, will not be shown on the /users endpoint                                                                         |

>![INFO]
> If a default permission is included in the perm list, it is interpreted as a 'negative'. So by including the 'write' perm on a user, you'd be making this user 'read-only'

>![INFO]
> readWhispers and writeWhispers accept a special name, '*'. If this name is present, it indicates all users. If this name is present & other users are present, those other users are interpreted as negative values - ie. "everyone except these names"

| permission |      default?      | info                                                       |
| :--------: | :----------------: | :--------------------------------------------------------- |
|    read    | :white_check_mark: | Read past messages                                         |
| read_live  | :white_check_mark: | Connect via WS & read live messages                        |
|   write    | :white_check_mark: | Write new public messages                                  |
|   delete   |        :x:         | delete the last message that was sent (including whispers) |

### Other options

These are env variables that can be used to configure the server. These can be loaded from a `.env` file.

|  env var  | default | info                                                                              |
| :-------: | :-----: | :-------------------------------------------------------------------------------- |
|   PORT    |  3000   | The port to run the server on                                                     |
| RING_SIZE |   700   | The max. amount of messages **clients** can load. Does not affect stored messages |

### Flags

| flag  |  default   | info                                                                                                                                                                                     |
| :---: | :--------: | :--------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
|  -u   | users.yaml | The location of the users configuration                                                                                                                                                  |
|  -s   | db.shitdb  | The location of the messages                                                                                                                                                             |
|  -e   |    .env    | The location of the env file                                                                                                                                                             |
|  -o   |     -      | For export cmd only! Which file to output into. `-` is understood to be stdout.                                                                                                          |
|  -a   |  <empty>   | For export cmd only! Which user to output as. If present, it will include which whispers this user is able to see. Also accepts the special value "*", indicating "only public messages" |

### Commands

The default command is `exec`.

| command | info                                                                                                                      |
| :-----: | :------------------------------------------------------------------------------------------------------------------------ |
|  exec   | Runs the server & hosts                                                                                                   |
| export  | Exports the messages into a more readable log. Syntax is `notIRCServer export <log/json/json-pretty/yaml/yaml-sep>`. Output is STDOUT |
