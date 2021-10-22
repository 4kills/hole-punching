package server

// AddressStore stores addresses with domain ids and allows to process those.
type AddressStore interface {

    // ProcessAddress takes a domain id (of peer connections) and returns all addresses registered to that id except addr.
    // Furthermore, this method associates addr to id.
    // All occurrences of addr should be removed in the returned slice, i.e. the return slice should not contain addr.
    // An empty return slice is not an error case. A non-existent identifier should return an empty slice.
    ProcessAddress(id string, exceptAddr string) ([]string, error)
}

// TODO: add timestamp / max size functionality to remove old/too many connections
type domainAddrMap struct {
    m map[string][]string
}

func (idm domainAddrMap) ProcessAddress(id, addr string) ([]string, error) {
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
