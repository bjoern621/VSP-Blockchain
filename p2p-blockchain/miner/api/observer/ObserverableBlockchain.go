package observer

type ObservableBlockchain interface {
	// Attach is called by the observer to attach itself to the server.
	Attach(o BlockchainObserverAPI)
	// Detach is called by the observer to detach itself from the server.
	Detach(o BlockchainObserverAPI)

	// NotifyStartMining Starts the mining process
	NotifyStartMining()
	// NotifyStopMining Stops the mining process
	NotifyStopMining()
}
