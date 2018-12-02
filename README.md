# ipfs-node
A simple server that receives and handles authenticated requests to interact
with the IPFS API for use in other applications.

Its featureset is very small - It takes JWT's (Json Web Tokens), along with a
request to the IPFS HTTP API. If the JWT is valid, it forwards the request. If
it is not, it doesn't. The token type it accepts is the default specified in
the JWT-go library.

The main use case of this project is for larger IPFS data storage applications
that require a secure method of accessing multiple different IPFS instances,
especially use cases that require centralized control to maintain security in
the event of the compromise or failure of one or more ipfs nodes. 

This project does not come standard with a method of generating JWT tokens in
the style mentioned above, nor does it come with a method of managing users.
Instead, a ticketmaster application should be built around ipfs-node to manage
what keys the node considers valid and/or the distribution of JWT Tokens.
Implemmentation of a such a program that fits most standard use cases is coming
soon and has been deliberately isolated from ipfs-node so it can be swapped as
necessary.

Specific details on how to use IPFS-Node and its protocol specs are located in
the .tex and .pdf documents under /docs/.
