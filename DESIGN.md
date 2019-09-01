# Goals

These are the major goals:

* shared clipboard between multiple devices, using a staging area
* E2E encryption
* multiple entries in the clipboard

# Security model

One instance of the server is intended for one "entity" where entity is generally a person
wishing to share a clipboard between their devices. It can also be multiple persons if they
chose to setup the clipboard with the same parameters.

The server will identify devices with a public key that the devices will have created.
The device keeps its private key somewhere.

The server has its own key pair which a client uses to verify the content it gets from the server.

Both sides share a key used to encrypt payloads. This also prevents someone registering as a device
without knowing the key in advance.

# Request flow

* decrypt using the pre-shared key.
* verify the signature of the payload
  * using the device's public key on the server
  * using the server's public key on the device
* do the operation

# API Design

# POST /register

Plaintext body:

```json
{
    "device_id": "1sTJhvrvt7wVnRSKYatlNTr/nfs3ye/EOymo8Fvf3cQ=",
    "public_key": "ShrZQ7DTb5J/ksU7WDF0tk3YLNrBpeVE0/hIf8/JTfk="
}
```

Then we use [secretbox.Seal](https://godoc.org/golang.org/x/crypto/nacl/secretbox#Seal) with the PSK.

We prepend the nonce to that and base64 encode the result which is then used as the body of the request.
