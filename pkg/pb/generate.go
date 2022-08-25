package pb

//go:generate protoc --go_out=.      --go_opt=paths=source_relative      rpc.proto signaling.proto message.proto config.proto candidate.proto common.proto event.proto interface.proto epice.proto peer.proto
//go:generate protoc --go-grpc_out=. --go-grpc_opt=paths=source_relative rpc.proto signaling.proto
