# Coffer

A simple secret storage system using keybase and JSON.

## Installation

`coffer` is written in Go, and requires a working keybase installation to
function. You should be logged in (you can check `keybase status` to verify
this), and be able to use `keybase encrypt` and `keybase decrypt` from the
command line.

Then:

`go get github.com/ter0/coffer`

`cd $GOPATH/src/github.com/ter0/coffer`

`go install`

## Usage

### Subcommands
| command | aliases | description                                |
|---------|---------|--------------------------------------------|
| store   | s       | store a new secret                         |
| list    | ls, l   | list the names of currently stored secrets |
| create  | c       | create an empty coffer                     |
| get     | g       | get a single secret from the coffer        |
| delete  | d       | delete a single secret from the coffer     |

### Getting Started

1. `coffer create` - create an empty coffer. The empty coffer will be a file
called `.coffer`, located in the current users home directory.
2. `coffer store twitter` - store a secret. The first argument given will be the
name of the secret. Subsequent calls to store a secret with the same name will
overwrite any existing secret.
3. `coffer list` - list the currently stored secrets. This will list the names
of all currently stored secrets, but not the secret itself.
4. `coffer get twitter` - get the secret for the given name.

## License

Released under the [MIT license](LICENSE).