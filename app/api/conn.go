package api

import (
	"io"

	"github.com/genzai-io/sliced/common/evio"
)

type ConnKind byte

const (
	ConnCommand   ConnKind = 0
	ConnPubSub    ConnKind = 1
	ConnRaft      ConnKind = 2
	ConnQueue     ConnKind = 3
	ConnInstall   ConnKind = 4
	ConnHTTP      ConnKind = 5
	ConnWebSocket ConnKind = 6
)

type EvConn interface {
	Conn() evio.Conn
}

type EvCloser interface {
	Close() error
	OnClosed()
}

type EvData interface {
	OnData(in []byte) (out []byte, action evio.Action)
}

type EvDetacher interface {
	Detach() error

	OnDetach(rwc io.ReadWriteCloser)
}

type CommandConn interface {
	EvConn
	EvCloser
	EvData
	EvDetacher

	GetKind() ConnKind

	SetKind(kind ConnKind)

	//
	GetDurability() Durability

	//
	GetRaft() RaftService

	//
	SetRaft(raft RaftService)
}
