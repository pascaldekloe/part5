[![GoDoc](https://godoc.org/github.com/pascaldekloe/part5?status.svg)](https://godoc.org/github.com/pascaldekloe/part5)
[![Build Status](https://travis-ci.org/pascaldekloe/part5.svg?branch=master)](https://travis-ci.org/pascaldekloe/part5)

# Part5 With The Go Programming Language

The International Electrotechnical Commission standard 870 part 5 (IEC 870-5) is
a set of transmission procedures intended for SCADA systems. Prefix 60 was added
later as in IEC 60870. For serial communication please refer to **60870-5-101**
and **60870-5-104** is the TCP-based evolution.

The project consists of a high-level framework, including a low-level library,
and tooling for network exploration and automated testing.
**Incomplete**! About 95% of the work has been published; see issue #1.

This is free and unencumbered software released into the
[public domain](http://creativecommons.org/publicdomain/zero/1.0).


## Definitions

At its essence, part 5 formalizes reliable means to exchange data and commands.

The initiating stations are called *primary* and the responding stations are
*secondary*. With *unbalanced* transmission stations are either primary or
secodary [master/slave] and with *balanced* transmission stations can act both
as primary and as secondory [peer-to-peer].

Addressing follows this initiate–respond context; an origin address remains that
of the primary and a destination address points to (something on) the secondary,
regardless of the message direction.

A *common address* can be seen as a data set wherein each data element has its
own predefined *information object address*. The set may be downloaded with an
[interrogation command](http://godoc.org/github.com/pascaldekloe/part5/info#C_IC_NA_1).

The standard provides commands that operate on data and they may trigger remote
procceses by agreement. Again, each command is applied to a single information
object address. Secondaries either
[confirm](http://godoc.org/github.com/pascaldekloe/part5/info#Actcon) or
[reject](http://godoc.org/github.com/pascaldekloe/part5/info#NegFlag) execution.
Optionaly the command may indicate completion with a *terminate*
[message](http://godoc.org/github.com/pascaldekloe/part5/info#Actterm).
Some commands can be preceded with a *select*
[directive](http://godoc.org/github.com/pascaldekloe/part5/info#Cmd.Exec)
which locks down processing to one at a time.

To get started please see the API documentation of
[Dial](http://godoc.org/github.com/pascaldekloe/part5#Dial) [serial],
[DialTCP](http://godoc.org/github.com/pascaldekloe/part5#DialTCP) or
[ListenTCP](http://godoc.org/github.com/pascaldekloe/part5#ListenTCP).


## TODO

* file transfer
* parameters
* complete ASDU encoding types


## iechbin

The project comes with a test tool. Download a
[prebuilt binary](https://github.com/pascaldekloe/part5/releases)
or run `go get -u github.com/pascaldekloe/part5/cmd/iechbin`
to make one yourself.
Without arguments the command opens a browser UI.
