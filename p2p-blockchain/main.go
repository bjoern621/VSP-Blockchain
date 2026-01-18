package main

import (
	appapi "s3b/vsp-blockchain/p2p-blockchain/app/api"
	appcore "s3b/vsp-blockchain/p2p-blockchain/app/core"
	"s3b/vsp-blockchain/p2p-blockchain/app/infrastructure/adapters"
	appgrpc "s3b/vsp-blockchain/p2p-blockchain/app/infrastructure/grpc"
	blockapi "s3b/vsp-blockchain/p2p-blockchain/blockchain/api"
	"s3b/vsp-blockchain/p2p-blockchain/blockchain/core"
	"s3b/vsp-blockchain/p2p-blockchain/blockchain/core/utxo"
	"s3b/vsp-blockchain/p2p-blockchain/blockchain/core/validation"
	blockchainData "s3b/vsp-blockchain/p2p-blockchain/blockchain/data/blockchain"
	"s3b/vsp-blockchain/p2p-blockchain/blockchain/infrastructure"
	"s3b/vsp-blockchain/p2p-blockchain/internal/common"
	"s3b/vsp-blockchain/p2p-blockchain/internal/common/data/transaction"
	minerCore "s3b/vsp-blockchain/p2p-blockchain/miner/core"
	"s3b/vsp-blockchain/p2p-blockchain/netzwerkrouting/api"
	networkBlockchain "s3b/vsp-blockchain/p2p-blockchain/netzwerkrouting/core/blockchain"
	"s3b/vsp-blockchain/p2p-blockchain/netzwerkrouting/core/connectioncheck"
	"s3b/vsp-blockchain/p2p-blockchain/netzwerkrouting/core/disconnect"
	"s3b/vsp-blockchain/p2p-blockchain/netzwerkrouting/core/handshake"
	"s3b/vsp-blockchain/p2p-blockchain/netzwerkrouting/core/keepalive"
	corepeer "s3b/vsp-blockchain/p2p-blockchain/netzwerkrouting/core/peer"
	"s3b/vsp-blockchain/p2p-blockchain/netzwerkrouting/core/peer/discovery"
	"s3b/vsp-blockchain/p2p-blockchain/netzwerkrouting/core/peermanagement"
	"s3b/vsp-blockchain/p2p-blockchain/netzwerkrouting/data/peer"
	"s3b/vsp-blockchain/p2p-blockchain/netzwerkrouting/infrastructure/middleware/grpc"
	"s3b/vsp-blockchain/p2p-blockchain/netzwerkrouting/infrastructure/middleware/grpc/networkinfo"
	"s3b/vsp-blockchain/p2p-blockchain/netzwerkrouting/infrastructure/registry"
	walletApi "s3b/vsp-blockchain/p2p-blockchain/wallet/api"
	walletcore "s3b/vsp-blockchain/p2p-blockchain/wallet/core"
	"s3b/vsp-blockchain/p2p-blockchain/wallet/core/keys"

	"os"
	"os/signal"
	"syscall"

	"bjoernblessin.de/go-utils/util/assert"
	"bjoernblessin.de/go-utils/util/logger"
)

func main() {
	common.Init()

	logger.Infof("[main] Running...")

	logger.Infof("[main] Loglevel set to %v", logger.CurrentLevel())

	peerStore := peer.NewPeerStore()
	networkInfoRegistry := networkinfo.NewNetworkInfoRegistry(peerStore)
	disconnectService := disconnect.NewDisconnectService(networkInfoRegistry, peerStore)
	grpcClient := grpc.NewClient(networkInfoRegistry, disconnectService)
	handshakeService := handshake.NewHandshakeService(grpcClient, peerStore, grpcClient)
	handshakeAPI := api.NewHandshakeAPIService(networkInfoRegistry, peerStore, handshakeService)
	peerRetrieverAdapter := corepeer.NewPeerRetrieverAdapter(peerStore)
	networkRegistryAPI := api.NewNetworkRegistryService(networkInfoRegistry, peerRetrieverAdapter)
	registryQuerier := registry.NewDNSRegistryQuerier(networkInfoRegistry)
	queryRegistryAPI := api.NewQueryRegistryAPIService(registryQuerier)

	discoveryService := discovery.NewDiscoveryService(registryQuerier, grpcClient, peerStore, grpcClient, grpcClient)
	discoveryAPI := api.NewDiscoveryAPIService(discoveryService)
	periodicDiscoveryService := discovery.NewPeriodicDiscoveryService(peerStore, grpcClient, discoveryService)
	keepaliveService := keepalive.NewKeepaliveService(peerStore, grpcClient, grpcClient)
	connectionCheckService := connectioncheck.NewConnectionCheckService(peerStore, disconnectService, networkInfoRegistry)
	peerManagementService := peermanagement.NewPeerManagementService(peerStore, discoveryService, peerStore, handshakeService, peerStore)

	chainStateConfig := utxo.ChainStateConfig{CacheSize: 1000}
	utxoEntryDAOConfig := infrastructure.UTXOEntryDAOConfig{DBPath: "", InMemory: true}
	dao, err := infrastructure.NewUTXOEntryDAO(utxoEntryDAOConfig)
	assert.IsNil(err, "couldn't create UTXOEntryDAO")
	chainStateService, err := utxo.NewChainStateService(chainStateConfig, dao)
	assert.IsNil(err, "couldn't create chainStateService")
	// Initialize UTXO lookup service and API
	memPoolService := utxo.NewMemUTXOPoolService()
	fullNodeUtxoService := utxo.NewFullNodeUTXOService(memPoolService, chainStateService)

	genesisBlock := blockchainData.GenesisBlock()
	blockStore := blockchainData.NewBlockStore(genesisBlock)

	blockchainMsgService := networkBlockchain.NewBlockchainService(grpcClient, peerStore)

	transactionValidator := validation.NewValidationService(chainStateService)
	blockValidator := validation.NewBlockValidationService()

	blockchain := core.NewBlockchain(
		blockchainMsgService,
		grpcClient,
		grpcClient,
		transactionValidator,
		blockValidator,
		blockStore,
		fullNodeUtxoService,
		peerStore,
	)

	keyEncodingsImpl := keys.NewKeyEncodingsImpl()
	keyGeneratorImpl := keys.NewKeyGeneratorImpl(keyEncodingsImpl, keyEncodingsImpl)
	keyGeneratorApiImpl := walletApi.NewKeyGeneratorApiImpl(keyGeneratorImpl)

	utxoAPI := blockapi.NewUtxoAPI(fullNodeUtxoService)

	minerImpl := minerCore.NewMinerService(blockchain, fullNodeUtxoService, blockStore)
	blockchain.Attach(minerImpl)
	minerImpl.StartMining(make([]transaction.Transaction, 0)) // TODO

	if common.AppEnabled() {
		logger.Infof("[main] Starting App server...")
		// Intialize Transaction Creation API
		transactionCreationService := walletcore.NewTransactionCreationService(keyGeneratorImpl, keyEncodingsImpl, blockchainMsgService, utxoAPI)
		transactionCreationAPI := walletApi.NewTransactionCreationAPIImpl(transactionCreationService)

		// Initialize transaction service and API
		transactionService := appcore.NewTransactionService(transactionCreationAPI)
		transactionAPI := appapi.NewTransactionAPIImpl(transactionService)

		transactionHandler := adapters.NewTransactionAdapter(transactionAPI)

		// Initialize konto API and handler
		kontoAPI := appapi.NewKontoAPIImpl(utxoAPI, keyEncodingsImpl)
		kontoHandler := adapters.NewKontoAdapter(kontoAPI)

		// Initialize visualization service and handler
		visualizationService := appcore.NewVisualizationService(blockStore)
		visualizationHandler := adapters.NewVisualizationAdapter(visualizationService)

		// Initialize mining service
		miningService := appcore.NewMiningService(minerImpl)

		connService := appcore.NewConnectionEstablishmentService(handshakeAPI)
		internalViewService := appcore.NewInternsalViewService(networkRegistryAPI)
		queryRegistryService := appcore.NewQueryRegistryService(queryRegistryAPI)
		discoveryAppService := appcore.NewDiscoveryService(discoveryAPI)
		disconnectAPI := api.NewDisconnectAPIService(networkInfoRegistry, disconnectService)
		disconnectAppService := appcore.NewDisconnectService(disconnectAPI)

		appServer := appgrpc.NewServer(
			connService,
			internalViewService,
			queryRegistryService,
			keyGeneratorApiImpl,
			transactionHandler,
			discoveryAppService,
			kontoHandler,
			visualizationHandler,
			miningService,
			disconnectAppService,
		)

		err := appServer.Start(common.AppPort())
		if err != nil {
			logger.Warnf("[main] couldn't start App server: %v", err)
		} else {
			addrPort, err := appServer.ListeningEndpoint()
			assert.IsNil(err)
			common.SetAppPort(addrPort.Port())
			logger.Infof("[main] App server started on port %v", addrPort)
		}
	}

	logger.Infof("[main] Starting P2P server...")

	grpcServer := grpc.NewServer(handshakeService, networkInfoRegistry, discoveryService, keepaliveService, peerStore)

	grpcServer.Attach(blockchain)

	err = grpcServer.Start(common.P2PPort())
	if err != nil {
		logger.Warnf("[main] couldn't start P2P server: %v", err)
	} else {
		addrPort, err := grpcServer.ListeningEndpoint()
		assert.IsNil(err)
		common.SetP2PPort(addrPort.Port())
		common.SetP2PListeningIpAddr(addrPort.Addr())
		logger.Infof("[main] P2P server started on port %v", addrPort)
	}

	// Start keepalive service
	keepaliveService.Start()

	// Start connection check service
	connectionCheckService.Start()

	// Start periodic discovery service
	periodicDiscoveryService.Start()

	// Start peer management service
	peerManagementService.Start()

	// Handle graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	<-sigChan

	logger.Infof("[main] Shutting down...")
	keepaliveService.Stop()
	connectionCheckService.Stop()
	periodicDiscoveryService.Stop()
	peerManagementService.Stop()
	logger.Infof("[main] Shutdown complete")
}
