/*
Copyright IBM Corp. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package protos

//go:generate -command protoc-basic protoc --proto_path=.. --go_out=../pkg-blockchain/pb --go_opt=paths=source_relative

// Generate the code for codegen extensions.
//go:generate protoc-basic protos/mir/codegen_extensions.proto
//go:generate protoc-basic protos/net/codegen_extensions.proto

// Build the protoc plugin.
//go:generate go build -o ../codegen/protoc-plugin/protoc-gen-mir ../codegen/protoc-plugin

//go:generate -command protoc-events protoc --proto_path=.. --go_out=../pkg-blockchain/pb --go_opt=paths=source_relative --plugin=../codegen/protoc-plugin/protoc-gen-mir --mir_out=../pkg-blockchain/pb --mir_opt=paths=source_relative

// Generate the protoc-generated code for events and messages.
//go:generate protoc-events protos-blockchain/messagepb/messagepb.proto
//go:generate protoc-events protos-blockchain/eventpb/eventpb.proto
//go:generate protoc-events protos-blockchain/blockchainpb/blockchainpb.proto
//go:generate protoc-events protos-blockchain/bcmpb/bcmpb.proto
//go:generate protoc-events protos-blockchain/minerpb/minerpb.proto
//go:generate protoc-events protos-blockchain/broadcastpb/broadcastpb.proto
//go:generate protoc-events protos-blockchain/synchronizerpb/synchronizerpb.proto
//go:generate protoc-events protos-blockchain/applicationpb/applicationpb.proto
//go:generate protoc-events protos-blockchain/payloadpb/payloadpb.proto
//go:generate protoc-events protos-blockchain/statepb/statepb.proto
//go:generate protoc-events protos-blockchain/interceptorpb/interceptorpb.proto

// Build the custom code generators.
//go:generate go build -o ../codegen/generators/mir-std-gen/mir-std-gen.bin ../codegen/generators/mir-std-gen
//go:generate -command std-gen ../codegen/generators/mir-std-gen/mir-std-gen.bin

// Generate the Mir-generated code for events and messages.
//go:generate std-gen "github.com/filecoin-project/mir/pkg/pb/eventpb"
//go:generate std-gen "github.com/filecoin-project/mir/pkg/pb/messagepb"
//go:generate std-gen "github.com/filecoin-project/mir/pkg-blockchain/pb/blockchainpb"
//go:generate std-gen "github.com/filecoin-project/mir/pkg-blockchain/pb/bcmpb"
//go:generate std-gen "github.com/filecoin-project/mir/pkg-blockchain/pb/minerpb"
//go:generate std-gen "github.com/filecoin-project/mir/pkg-blockchain/pb/broadcastpb"
//go:generate std-gen "github.com/filecoin-project/mir/pkg-blockchain/pb/synchronizerpb"
//go:generate std-gen "github.com/filecoin-project/mir/pkg-blockchain/pb/payloadpb"
//go:generate std-gen "github.com/filecoin-project/mir/pkg-blockchain/pb/statepb"
//go:generate std-gen "github.com/filecoin-project/mir/pkg-blockchain/pb/applicationpb"
//go:generate std-gen "github.com/filecoin-project/mir/pkg-blockchain/pb/interceptorpb"

//go:generate cp eventinterface.go.template pkg-blockchain/pbeventpb/eventinterface.go
