# go-orchestrator

## Start using the SDK

You need to enable private go modules from entur:

```sh
go env -w GOPRIVATE='github.com/entur/*'
env GIT_TERMINAL_PROMPT=1 go get github.com/entur/go-orchestrator # to fix if you default to https
# git config --global --add url."git@github.com:".insteadOf "https://github.com/" # if you want ssh default always
```

## Minimal example

See `./orchestrator_test.go` for a complete test.
