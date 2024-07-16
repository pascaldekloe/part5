[![Go Reference](https://pkg.go.dev/badge/github.com/pascaldekloe/part5.svg)](https://pkg.go.dev/github.com/pascaldekloe/part5)
[![Build Status](https://github.com/pascaldekloe/part5/actions/workflows/go.yml/badge.svg)](https://github.com/pascaldekloe/part5/actions/workflows/go.yml)

# Part 5 With The Go Programming Language

The International Electrotechnical Commission standard 870 part 5 (IEC 870-5) is
a set of transmission procedures intended for SCADA systems. The publication got
reissued with a designation in the 60000 series as of the year 1997. Refer to
IEC 60870-5-101 for serial communication, and IEC 60870-5-104 for the TCP-based
evolution. The two are commonly abbreviated as IEC 101 and IEC 104 respectively.

The project consists of a Go library and a command-line tool called iecat(8).

This is free and unencumbered software released into the
[public domain](http://creativecommons.org/publicdomain/zero/1.0).


## Definitions

At its essence, part 5 formalizes reliable means to exchange measurements and
commands. Commands are used by controlling stations to cause a change of state
in operational equipment. Controlled stations may then optionally either
[confirm](http://godoc.org/github.com/pascaldekloe/part5/info#Actcon) or
or [reject](http://godoc.org/github.com/pascaldekloe/part5/info#NegFlag) the
execution. Controlled stations can also optionally indicate completion with a
[terminate](http://godoc.org/github.com/pascaldekloe/part5/info#Actterm).

Controlling stations are called *primary* and the controlled stations are called
*secondary*. With *unbalanced* transmission one station is primary and the other
is secodary, a.k.a. master/slave, and with *balanced* transmission stations can
both act as primary and as secondory, a.k.a. peer to peer.

Every *information object address* resides in a *common address*. Primaries may
acquire the information present in a common address with an
[interrogation command](http://godoc.org/github.com/pascaldekloe/part5/info#C_IC_NA_1).
E.g., run `iecat -host station1.example.com -inro 42` to aquire a listing of
common address 42 on the terminal.


## To Do

The following is not implement yet.

* file transfer, type identifier F_*
* parameter of measured value, type identifeir P_*


## iecat

Run `go install github.com/pascaldekloe/part5/cmd/iecat@latest` to build the
command into the bin directory of `go env GOPATH`.
