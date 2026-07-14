<p align="center">
  <img src="./assets/verislabs.ico" width="75" height="75" alt="Veris Labs Logo" align="middle" />
  <img height="50" src="https://readme-typing-svg.herokuapp.com?font=Inter&weight=700&size=36&color=FFFFFF&center=true&vCenter=true&width=220&lines=verislabs.io&repeat=false&duration=2500" align="middle" />
</p>

<p align="center">
Official <b>Golang SDK</b> for Veris Labs.<br>
It provides a convenient way to interact with the Veris Labs API and perform various operations.<br>
For more information, navigate to <a href="https://docs.verislabs.io">docs.verislabs.io</a>.
</p>

# Install

Simply run the following command to install the SDK:\
`go get github.com/verislabs/veris-go`

# Usage

The main idea of this SDK is not to expose the API endpoints directly, but instead provide a set of *programmatic primitives* to reduce the boilerplate and make it **easier to consume the API**.\
The SDK is designed around the concept of a **"session"**, which covers entire flow - from getting homepage to posting the checkout. Single session primitive ensures **consistency** and makes it easier to use the API **without having to worry about the underlying details**.

## Client

In order to interact with the SDK you must first create a client.

```go
client := veris.NewClient("veris_apikey")
```

## Akamai BMP iOS

`AkamaiBMPIOSSession` returns a builder and does not make a request. `Create`
makes the initialization request and returns the stateful session. Creating an
unused session is still charged at the normal sensor-generation rate.

```go
session, err := client.AkamaiBMPIOSSession(
    "4.2.0",
    "com.xyz.app",
    "1.0.0",
).
    WithIOSVersionRange("17.0.0", "18.9.9").
    WithModel("iPhone14,2").
    Create(ctx)
if err != nil {
    return err
}
```

`Sensor` also only returns a builder. `Generate` makes one charged request and,
on success, automatically advances the session state.

```go
sensor, reportData, err := session.Sensor().
    WithParams(paramsResponse).
    WithDCIScript(dciScript).
    Generate(ctx)
if err != nil {
    return err
}
```

## Akamai Web V3

`AkamaiWebV3Session` returns a builder and makes no request. `Create` only
creates local SDK state; the first `Sensor` call initializes the remote session
and is the first charged request.

```go
session := client.AkamaiWebV3Session(
    userAgent,
    scriptURL,
    scriptBody,
    "en-US",
).
    WithIP(proxyIP).
    Create()

sensor, reportData, err := session.Sensor(ctx, pageURL, abck, bmsz)
if err != nil {
    return err
}
```

The session automatically sends the script on its first request and the latest
opaque encrypted session state on subsequent requests.

## Akamai Web SBSD

`AkamaiWebSBSDSession` returns a builder and makes no request. `Create` only
creates local SDK state; each `Sensor` call is a charged request.

```go
session := client.AkamaiWebSBSDSession(
    userAgent,
    scriptURL,
    scriptBody,
    "en-US",
).Create()

sensor, reportData, err := session.Sensor(ctx, pageURL, bmSo)
if err != nil {
    return err
}
```

The session automatically preserves the encrypted state required for the first
and follow-up SBSD sensor stages.