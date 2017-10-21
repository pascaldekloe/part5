# part5

The International Electrotechnical Commission standard 870 part 5 (IEC 870-5) is
a set of transmission procedures intended for SCADA systems. Prefix 60 was added
later as in IEC 60870. Companion standards **60870-5-101** and **60870-5-104**
define a protocol for serial communication and TCP.

This is free and unencumbered software released into the
[public domain](http://creativecommons.org/publicdomain/zero/1.0).

To get started please see the API documentation of
[Dial](http://godoc.org/github.com/pascaldekloe/part5#Dial) [serial],
[DialTCP](http://godoc.org/github.com/pascaldekloe/part5#DialTCP) or
[ListenTCP](http://godoc.org/github.com/pascaldekloe/part5#ListenTCP).


## iechbin

The project comes with a test tool. Download a
[prebuilt binary](https://github.com/pascaldekloe/part5/releases)
or run `go get -u github.com/pascaldekloe/part5/cmd/iechbin`
to make one yourself.
Without arguments the command opens a browser UI.
