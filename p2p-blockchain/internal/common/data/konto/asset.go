package konto

// Asset represents a single unspent output value belonging to an address.
type Asset struct {
	Value uint64
}

// AssetsResult represents the outcome of an assets query.
type AssetsResult struct {
	Success      bool
	ErrorMessage string
	Assets       []Asset
}
