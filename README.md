# Apéro

_Note_: readme is not yet complete.

*Apéro* is a staging server for pieces of arbitrary data.

The primary use case is transferring data between two devices when direct communication between the devices is inconvenient,
for example between a Macbook laptop and a Windows PC or between a PC and your Android phone.

Data is end-to-end encrypted and therefore never in the clear for the staging server.

## Design

The server is essentially a list of _entries_ where an _entry_ is a unique ID and a slice of bytes.

It provides these APIs:

* `POST /list` to retrieve the entries IDs
* `DELETE /move` to move a piece of data out of the staging server. This removes the entry.
* `POST /paste` to copy a piece of data from the staging server. This doesn't remove the entry.
* `POST /copy` to send a piece of data to the staging server.

## Encryption

Each piece of data is end-to-end encrypted using a key only the different devices know.

The different requests for the APIs described above are signed using a private key only the different devices know.

Finally, the payload (signature + request) is encrypted using a pre-shared key known by both the staging server and the devices.

## Data storage

TODO
