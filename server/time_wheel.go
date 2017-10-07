package server

import (
	"time"
	"math/rand"
	"context"
)

func RandomTimeout(mult float32) int {
	lowRange := 100 * mult
	highRange := 1000 * mult
	return int(lowRange + highRange * rand.Float32())
}

func RefreshTimer(meta *RTGroupMeta, mult float32) {
	meta.Timeout = time.Now().Add(time.Duration(RandomTimeout(mult)) * time.Millisecond)
}

func (s *BFTRaftServer) StartTimingWheel() {
	go func() {
		for true {
			for _, meta := range s.GroupsOnboard {
				meta.Lock.Lock()
				if meta.Timeout.After(time.Now()) {
					if meta.Role == FOLLOWER {
						if meta.Leader != meta.Peer {
							panic("Follower is leader")
						}
						// not leader
						// TODO: request votes
						s.BecomeCandidate(meta)
					} else if meta.Role == LEADER {
						// is leader, send heartbeat
						s.SendFollowersHeartbeat(context.Background(), meta.Peer, meta.Group)
					} else if meta.Role == CANDIDATE {
						// is candidate but vote expired, start a new vote
						s.BecomeCandidate(meta)
					} else if meta.Role == OBSERVER {
						// TODO: update local data
						RefreshTimer(meta, 10)
					}
				}
				meta.Lock.Unlock()
			}
			time.Sleep(100 * time.Millisecond)
		}
	}()
}