# This example will call the server to retrivev the public key of the server
# And then the client can use the key to encrypt their seed data to avoid the network attack (But middle man attack might happens)

curl  http://localhost:8080/v1/serverPublicKeys
