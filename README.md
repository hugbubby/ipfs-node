# ipfs-remote
A server that receives and handles authenticated requests to interact with the IPFS API.


Quick Cooking Instructions:
Assuming binary is in working directory,
1. Place 1 ssh public key at security/master.pub
2. Place 1 TLS RSA public/private keypair at security/server.cert and security/server.key, respectively
3. Add sugar

Voila. You may now send signed requests to the URL specified in config.json.
Read protocol.tex for more information.
