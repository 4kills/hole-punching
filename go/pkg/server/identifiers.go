package server

// TODO: maybe refactor this to one method later (if using databases [to make transactions])
type addressStore interface {

    // FetchAddresses takes an identifier id (of peer connections) and returns all addresses registered to that id except exceptAddr.
    // All occurrences of exceptAddr should be removed, i.e. the return slice should not contain exceptAddr.
    // An empty return slice is not an error case. A non-existent identifier should return an empty slice.
    FetchAddresses(id string, exceptAddr string) ([]string, error)
    PutAddress(id string, addr string) error
}

// TODO: add timestamp / max size functionality to remove old/too many connections
type identifierAddrMap struct {
    m map[string][]string
}

func (idm identifierAddrMap) FetchAddresses(id, exceptAddr string) ([]string, error) {
    s, ok := idm.m[id]
    if !ok {
        return make([]string, 0), nil
    }

    ret := make([]string, len(s))

    for i, addr := range s {
        if addr != exceptAddr {
            ret[i] = addr
        }
    }
    return ret, nil
}

func (idm identifierAddrMap) PutAddress(id, addr string) error {
    s, ok := idm.m[id]
    if !ok {
        s = make([]string, 0)
    }

    s = append(s, addr)

    idm.m[id] = s
    return nil
}
