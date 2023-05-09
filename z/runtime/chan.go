package runtime

import "log"

func IsClosedChan(c <-chan struct{}) bool {
	select {
	case v, ok := <-c:
		log.Printf("IsClosedChan(c) = (%v, %v), chan = %v", v, ok, c)
		return !ok
	default:
		log.Printf("IsClosedChan(c) default, chan = %v ", c)
		return false
	}
}
