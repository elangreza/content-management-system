package service

type MapCommand[K comparable, V any] struct {
	Op      string
	Key     K
	Value   V
	ReplyCh chan interface{}
}

type SafeMap[K comparable, V any] struct {
	cmdCh chan MapCommand[K, V]
}

func NewSafeMap[K comparable, V any]() *SafeMap[K, V] {
	m := make(map[K]V)
	cmdCh := make(chan MapCommand[K, V])
	go func() {
		for cmd := range cmdCh {
			switch cmd.Op {
			case "get":
				cmd.ReplyCh <- m[cmd.Key]
			case "exist":
				_, ok := m[cmd.Key]
				cmd.ReplyCh <- ok
			case "set":
				m[cmd.Key] = cmd.Value
				cmd.ReplyCh <- struct{}{}
			case "delete":
				delete(m, cmd.Key)
				cmd.ReplyCh <- struct{}{}
			case "getAll":
				copyMap := make(map[K]V, len(m))
				for k, v := range m {
					copyMap[k] = v
				}
				cmd.ReplyCh <- copyMap
			}
		}
	}()
	return &SafeMap[K, V]{cmdCh: cmdCh}
}

func (s *SafeMap[K, V]) Get(key K) V {
	reply := make(chan interface{})
	s.cmdCh <- MapCommand[K, V]{Op: "get", Key: key, ReplyCh: reply}
	return (<-reply).(V)
}

func (s *SafeMap[K, V]) Exist(key K) bool {
	reply := make(chan interface{})
	s.cmdCh <- MapCommand[K, V]{Op: "exist", Key: key, ReplyCh: reply}
	return (<-reply).(bool)
}

func (s *SafeMap[K, V]) Set(key K, value V) {
	reply := make(chan interface{})
	s.cmdCh <- MapCommand[K, V]{Op: "set", Key: key, Value: value, ReplyCh: reply}
	<-reply
}

func (s *SafeMap[K, V]) Delete(key K) {
	reply := make(chan interface{})
	s.cmdCh <- MapCommand[K, V]{Op: "delete", Key: key, ReplyCh: reply}
	<-reply
}

func (s *SafeMap[K, V]) GetAll() map[K]V {
	reply := make(chan interface{})
	s.cmdCh <- MapCommand[K, V]{Op: "getAll", ReplyCh: reply}
	return (<-reply).(map[K]V)
}
