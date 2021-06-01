package container

type StructArgs struct{}

func (StructArgs) isStructArgs() {}

type isStructArgs interface{ isStructArgs() }
