package provider

import ()

// A reset is spent by itself only within resetExpiryLead of running out,
// and only when the account's windows have something to start again; an
// account is read again when there may be something to do, not before.
