/*
Copyright IBM Corp. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package protos

//go:generate -command protoc-events protoc --proto_path=. --go_out=../pkg-blockchain/pb --go_opt=paths=source_relative --plugin=../codegen/protoc-plugin/protoc-gen-mir --mir_out=../pkg-blockchain/pb --mir_opt=paths=source_relative

// Generate the protoc-generated code for events and messages.
//go:generate protoc-events messagepb/messagepb.proto
//go:generate protoc-events eventpb/eventpb.proto
//go:generate protoc-events blockchainpb/blockchainpb.proto
//go:generate protoc-events bcmpb/bcmpb.proto
//go:generate protoc-events minerpb/minerpb.proto
//go:generate protoc-events broadcastpb/broadcastpb.proto
//go:generate protoc-events synchronizerpb/synchronizerpb.proto
//go:generate protoc-events applicationpb/applicationpb.proto
//go:generate protoc-events payloadpb/payloadpb.proto
//go:generate protoc-events statepb/statepb.proto
//go:generate protoc-events interceptorpb/interceptorpb.proto

// Build the custom code generators.
//go:generate -command std-gen ../codegen/generators/mir-std-gen/mir-std-gen.bin

// Generate the Mir-generated code for events and messages.
//go:generate std-gen "github.com/filecoin-project/mir/pkg-blockchain/pb/eventpb"
//go:generate std-gen "github.com/filecoin-project/mir/pkg-blockchain/pb/messagepb"
//go:generate std-gen "github.com/filecoin-project/mir/pkg-blockchain/pb/blockchainpb"
//go:generate std-gen "github.com/filecoin-project/mir/pkg-blockchain/pb/bcmpb"
//go:generate std-gen "github.com/filecoin-project/mir/pkg-blockchain/pb/minerpb"
//go:generate std-gen "github.com/filecoin-project/mir/pkg-blockchain/pb/broadcastpb"
//go:generate std-gen "github.com/filecoin-project/mir/pkg-blockchain/pb/synchronizerpb"
//go:generate std-gen "github.com/filecoin-project/mir/pkg-blockchain/pb/payloadpb"
//go:generate std-gen "github.com/filecoin-project/mir/pkg-blockchain/pb/statepb"
//go:generate std-gen "github.com/filecoin-project/mir/pkg-blockchain/pb/applicationpb"
//go:generate std-gen "github.com/filecoin-project/mir/pkg-blockchain/pb/interceptorpb"

//go:generate cp eventinterface.go.template pkg-blockchain/pb/eventpb/eventinterface.go
