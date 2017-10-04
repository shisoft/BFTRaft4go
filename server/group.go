package server

import (
	"encoding/binary"
	pb "github.com/PomeloCloud/BFTRaft4go/proto"
	"github.com/dgraph-io/badger"
	"github.com/golang/protobuf/proto"
	"github.com/patrickmn/go-cache"
	"strconv"
)

func (s *BFTRaftServer) GetGroup(groupId uint64) *pb.RaftGroup {
	cacheKey := strconv.Itoa(int(groupId))
	cachedGroup, cacheFound := s.Groups.Get(cacheKey)
	if cacheFound {
		return cachedGroup.(*pb.RaftGroup)
	} else {
		group := &pb.RaftGroup{}
		keyPrefix := ComposeKeyPrefix(groupId, GROUP_META)
		item := badger.KVItem{}
		s.DB.Get(keyPrefix, &item)
		data := ItemValue(&item)
		if data == nil {
			return nil
		} else {
			proto.Unmarshal(*data, group)
			s.Groups.Set(cacheKey, group, cache.DefaultExpiration)
			return group
		}
	}
}

func (s *BFTRaftServer) GetGroupLogMaxIndex(groupId uint64) uint64 {
	key := ComposeKeyPrefix(groupId, GROUP_MAX_IDX)
	item := badger.KVItem{}
	s.DB.Get(key, &item)
	data := ItemValue(&item)
	var idx uint64 = 0
	if data != nil {
		idx = BytesU64(*data, 0)
	}
	return idx
}

func (s *BFTRaftServer) IncrGetGroupLogMaxIndex(groupId uint64) uint64 {
	key := ComposeKeyPrefix(groupId, GROUP_MAX_IDX)
	for true {
		item := badger.KVItem{}
		s.DB.Get(key, &item)
		data := ItemValue(&item)
		var idx uint64 = 0
		if data != nil {
			idx = BytesU64(*data, 0)
		}
		idx += 1
		if s.DB.CompareAndSet(key, U64Bytes(idx), item.Counter()) == nil {
			return idx
		}
	}
	panic("Incr Group IDX Failed")
}
