# ipfs-remote
A simple server that receives and handles authenticated requests to interact with the IPFS API for use in other applications.

Its featureset is very small - It takes JWT's (Json Web Tokens), along with a request to the IPFS HTTP API. If the JWT is valid,
it forwards the request. If it is not, it doesn't. The token type it accepts is the default specified in the JWT-go library.

The main use case of this project is for larger IPFS data storage applications that require a secure method of accessing multiple different
IPFS instances, especially use cases that require centralized control to maintain security in the event of the compromise or failure
of one or more ipfs nodes. 

This project does not come standard with a method of generating JWT tokens in the style mentioned above, and instead relies on the
client application(s) to generate such tickets in a manner it feels responsible. An example implemmentation of a ticketmaster that
fits most standard use cases is coming soon.

Specific details on how to use IPFS-Remote and its protocol specs are located in the .tex and .pdf documents under /docs/.
