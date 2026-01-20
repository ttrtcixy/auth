package domain

type TxOptions struct {
	IsoLevel   isoLevel
	AccessMode accessMode
}

var (
	TxSerializableRW = TxOptions{
		IsoLevel:   Serializable,
		AccessMode: ReadWrite,
	}
	TxRepeatableReadRW = TxOptions{
		IsoLevel:   RepeatableRead,
		AccessMode: ReadWrite,
	}
	TxReadCommittedRW = TxOptions{
		IsoLevel:   ReadCommitted,
		AccessMode: ReadWrite,
	}
	TxReadUncommittedRW = TxOptions{
		IsoLevel:   ReadUncommitted,
		AccessMode: ReadWrite,
	}
)

type isoLevel int8

const (
	Serializable isoLevel = iota
	RepeatableRead
	ReadCommitted
	ReadUncommitted
)

type accessMode int8

const (
	ReadWrite accessMode = iota
	ReadOnly
)
