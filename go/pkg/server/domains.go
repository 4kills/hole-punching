package server

import (
    "math"
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
    // This method should be safe for concurrent use.
    //
    // ProcessAddress deletes addr from domain id after timeout if timeout is non-negative. If timeout is negative,
    // nothing should be deleted.
    ProcessAddress(id string, addr string, timeout time.Duration) ([]string, error)

    // FetchAllAddresses returns all addresses currently in the address store as []string or an error.
    // This method must be safe for concurrent use with AddressStore.ProcessAddress.
    FetchAllAddresses() ([]string, error)
}

type domainAddrMap struct {
    m map[string][]string
    mutex *sync.Mutex
    allAddr []string
}

func (idm domainAddrMap) FetchAllAddresses() ([]string, error) {
    idm.mutex.Lock()
    defer idm.mutex.Unlock()

    c := 0
    for _, v := range idm.m {
        c += len(v)
    }

    // reallocates array if it is nil, 4 times too large and larger than 1024 or if it doesn't suffice to store all addresses
    if idm.allAddr == nil || (4 * c <= len(idm.allAddr) && c > 1024 || c > len(idm.allAddr)) {
        n := math.Max(float64(c), 1024)
        idm.allAddr = make([]string, 2 * int(n))
    }

    i := 0
    for _, v := range idm.m {
        for _, addr := range v {
            idm.allAddr[i] = addr
            i++
        }
    }

    return idm.allAddr[:c], nil
}

func (idm domainAddrMap) ProcessAddress(id, addr string, timeout time.Duration) ([]string, error) {
    idm.mutex.Lock()
    defer idm.mutex.Unlock()

    defer func() {go idm.clear(id, addr, timeout)}()

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

// clear removes addr from domain id after timeout. If timeout is negative, clear is not
// executed. If the last addr of a domain id is removed, the map key id is deleted altogether.
func (idm domainAddrMap) clear(id string, addr string, timeout time.Duration) {
    if timeout < 0 {
        return
    }
    time.Sleep(timeout)

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
