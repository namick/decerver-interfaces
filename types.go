package types

// struct used for communication over channels
// passed to subscribe functions
type Update struct{
    Address string
    Event string
    Resource interface{}
}

// Ordered map for storage in an account
type Storage struct{
    Storage map[string]string
    Order []string
}

// Ordered map for all accounts
type State struct{
    State map[string]Storage// map addrs to map of storage to value
    Order []string // ordered addrs and ordered storage inside
}


