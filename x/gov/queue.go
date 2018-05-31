package gov

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

type Queue struct {
	value []int64
	key   []byte
	store sdk.KVStore
	keeper Keeper
}

func GetQueue(keeper Keeper, store sdk.KVStore, key []byte) *Queue {
	queueName, _ := keeper.cdc.MarshalBinary(key)
	bz := store.Get(queueName)

	var data []int64
	if len(bz) != 0 {
		err := keeper.cdc.UnmarshalBinary(bz, &data)
		if err != nil {
			panic(err)
		}
	}

	return &Queue{
		data, key, store, keeper,
	}
}

func (q *Queue) Push(data int64) {
	q.value = append(q.value, data)
	bz, err := q.keeper.cdc.MarshalBinary(q.value)
	if err != nil {
		panic(err)
	}
	key, _ := q.keeper.cdc.MarshalBinary(q.key)
	q.store.Set(key, bz)
}

func (q *Queue) Peek() (*Proposal) {
	if len(q.value) == 0 {
		return nil
	}
	return q.get(q.value[0])
}

func (q *Queue) Pop() (*Proposal) {
	var newQ = []int64{}
	var element int64
	if len(q.value) == 0 {
		return nil
	}else if len(q.value) == 1 {
		element = q.value[0]
	}else {
		element, newQ = q.value[0], q.value[1:]
	}
	bz, err := q.keeper.cdc.MarshalBinary(newQ)
	if err != nil {
		panic(err)
	}
	key, _ := q.keeper.cdc.MarshalBinary(q.key)
	q.store.Set(key, bz)
	return q.get(element)
}

func (q *Queue) Remove(value int64) (removed bool){
	for index, ele := range q.value {
		if ele == value {
			data := q.value[:index]
			q.value = append(data, q.value[index+1:]...)
			removed = true
			break
		}
	}
	bz, err := q.keeper.cdc.MarshalBinary(q.value)
	if err != nil {
		panic(err)
	}
	key, err := q.keeper.cdc.MarshalBinary(q.key)
	if err != nil {
		panic(err)
	}
	q.store.Set(key, bz)

	return removed
}

func (q *Queue) GetAll() (list []*Proposal) {
	for _,id := range q.value {
		list = append(list,q.get(id))
	}
	return list
}

func (q *Queue) get(proposalID int64) *Proposal {
	key, _ := q.keeper.cdc.MarshalBinary(proposalID)
	bz := q.store.Get(key)
	if bz == nil {
		return nil
	}

	proposal := &Proposal{}
	err := q.keeper.cdc.UnmarshalBinary(bz, proposal)
	if err != nil {
		panic(err)
	}

	return proposal
}
