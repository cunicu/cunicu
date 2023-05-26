// SPDX-FileCopyrightText: 2023 Steffen Vogel <post@steffenvogel.de>
// SPDX-License-Identifier: Apache-2.0

// Package proto contains auto-generated Go-code based of Protobuf descriptions
package proto

//go:generate protoc --proto_path=../../proto --go_out=.      --go_opt=paths=import,module=github.com/stv0g/cunicu/pkg/proto common.proto
//go:generate protoc --proto_path=../../proto --go_out=.      --go_opt=paths=import,module=github.com/stv0g/cunicu/pkg/proto core/peer.proto core/interface.proto core/net.proto
//go:generate protoc --proto_path=../../proto --go_out=.      --go_opt=paths=import,module=github.com/stv0g/cunicu/pkg/proto signaling/signaling.proto signaling/relay.proto
//go:generate protoc --proto_path=../../proto --go_out=.      --go_opt=paths=import,module=github.com/stv0g/cunicu/pkg/proto rpc/daemon.proto rpc/epdisc.proto rpc/event.proto rpc/signaling.proto rpc/invitation.proto
//go:generate protoc --proto_path=../../proto --go_out=.      --go_opt=paths=import,module=github.com/stv0g/cunicu/pkg/proto feature/epdisc.proto feature/epdisc_candidate.proto feature/pdisc.proto feature/pske.proto feature/hooks.proto

//go:generate protoc --proto_path=../../proto --go-grpc_out=. --go-grpc_opt=paths=import,module=github.com/stv0g/cunicu/pkg/proto rpc/daemon.proto rpc/epdisc.proto rpc/signaling.proto
//go:generate protoc --proto_path=../../proto --go-grpc_out=. --go-grpc_opt=paths=import,module=github.com/stv0g/cunicu/pkg/proto signaling/signaling.proto signaling/relay.proto
