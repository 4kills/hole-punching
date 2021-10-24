package server

import (
    "sync"
    "time"
)

// AddressStore stores addresses with domain ids and allows to process those. AddressStore must be safe for concurrent use.
type AddressStore interface {
    // ProcessAddress takes a domain id (of peer connections) and returns all addresses registered to that id except addr.
    // Furthermore, this method associates addr to id.
    // All occurrences of addr should be removed in the returned slice, i.e. the return slice should not contain addr.
    // An empty return slice is not an error case. A non-existent identifier should return an empty slice.
    //
    // ProcessAddress should be safe for concurrent use.
    //
    // ProcessAddress can also be used for clearing a domain (or individual address) from the AddressStore.
    ProcessAddress(id string, addr string) ([]string, error)
}

func SetAddressStore(store AddressStore) {
    addrStore = store
}

// TODO: add timestamp / max size functionality to remove old/too many connections
type domainAddrMap struct {
    m map[string][]string
    mutex *sync.Mutex
    timeout time.Duration
}

func (idm domainAddrMap) ProcessAddress(id, addr string) ([]string, error) {
    idm.mutex.Lock()
    defer idm.mutex.Unlock()

    defer func() {go idm.clear(id, addr)}()

    var ret []string

    s, ok := idm.m[id]
    if !ok {
        ret = make([]string, 1)
        ret[0] = addr
        idm.m[id] = ret

        return ret[:0], nil
    }

    ret = make([]string, len(s), len(s) + 1)

    i := 0
    for _, v := range s {
        if v != addr {
            ret[i] = v
            i++
        }
    }
    ret = ret[:i+1]

    ret[i] = addr
    idm.m[id] = ret

    return ret[:i], nil
}

func (idm domainAddrMap) clear(id string, addr string) {
    time.Sleep(idm.timeout)

    idm.mutex.Lock()
    defer idm.mutex.Unlock()

    s, ok := idm.m[id]
    if !ok {
        return
    }

    for i := 0; i < len(s); i++ {
        if addr == s[i] {
            // remove addr
            s[i] = s[len(s) - 1]
            s = s[:len(s) - 1]
            i-- // check same spot again in case addr is contained multiple times
        }
    }
    idm.m[id] = s

    if len(s) == 0 {
        delete(idm.m, id)
    }
}
