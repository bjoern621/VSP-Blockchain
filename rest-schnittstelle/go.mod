module s3b/vsp-blockchain/rest-api

go 1.25.3

require bjoernblessin.de/go-utils v1.0.1

require (
	golang.org/x/net v0.42.0 // indirect
	golang.org/x/sys v0.34.0 // indirect
	golang.org/x/text v0.27.0 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20250804133106-a7a43d27e69b // indirect
	google.golang.org/grpc v1.76.0
	google.golang.org/protobuf v1.36.6 // indirect
	s3b/vsp-blockchain/p2p-blockchain v0.0.0-00010101000000-000000000000
)

replace s3b/vsp-blockchain/p2p-blockchain => ../p2p-blockchain
