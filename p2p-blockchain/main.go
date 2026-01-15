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
	minerCore "s3b/vsp-blockchain/p2p-blockchain/miner/core"
	"s3b/vsp-blockchain/p2p-blockchain/netzwerkrouting/api"
	networkBlockchain "s3b/vsp-blockchain/p2p-blockchain/netzwerkrouting/core/blockchain"
	"s3b/vsp-blockchain/p2p-blockchain/netzwerkrouting/core/handshake"
	"s3b/vsp-blockchain/p2p-blockchain/netzwerkrouting/core/peer/discovery"
	"s3b/vsp-blockchain/p2p-blockchain/netzwerkrouting/data/peer"
	"s3b/vsp-blockchain/p2p-blockchain/netzwerkrouting/infrastructure/middleware/grpc"
	"s3b/vsp-blockchain/p2p-blockchain/netzwerkrouting/infrastructure/middleware/grpc/networkinfo"
	"s3b/vsp-blockchain/p2p-blockchain/netzwerkrouting/infrastructure/registry"
	walletApi "s3b/vsp-blockchain/p2p-blockchain/wallet/api"
	walletcore "s3b/vsp-blockchain/p2p-blockchain/wallet/core"
	"s3b/vsp-blockchain/p2p-blockchain/wallet/core/keys"

	"bjoernblessin.de/go-utils/util/assert"
	"bjoernblessin.de/go-utils/util/logger"
)

func main() {
	common.Init()

	logger.Infof("Running...")

	logger.Infof("Loglevel set to %v", logger.CurrentLevel())

	peerStore := peer.NewPeerStore()
	networkInfoRegistry := networkinfo.NewNetworkInfoRegistry(peerStore)
	grpcClient := grpc.NewClient(networkInfoRegistry)
	handshakeService := handshake.NewHandshakeService(grpcClient, peerStore)
	handshakeAPI := api.NewHandshakeAPIService(networkInfoRegistry, peerStore, handshakeService)
	networkRegistryAPI := api.NewNetworkRegistryService(networkInfoRegistry, peerStore)
	registryQuerier := registry.NewDNSRegistryQuerier(networkInfoRegistry)
	queryRegistryAPI := api.NewQueryRegistryAPIService(registryQuerier)

	discoveryService := discovery.NewDiscoveryService(registryQuerier, peerStore, grpcClient, peerStore, grpcClient)
	discoveryAPI := api.NewDiscoveryAPIService(discoveryService)

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

	blockchain := core.NewBlockchain(blockchainMsgService, transactionValidator, blockValidator, blockStore, fullNodeUtxoService)

	keyEncodingsImpl := keys.NewKeyEncodingsImpl()
	keyGeneratorImpl := keys.NewKeyGeneratorImpl(keyEncodingsImpl, keyEncodingsImpl)
	keyGeneratorApiImpl := walletApi.NewKeyGeneratorApiImpl(keyGeneratorImpl)

	utxoAPI := blockapi.NewUtxoAPI(fullNodeUtxoService)

	minerImpl := minerCore.NewMinerService(blockchain, fullNodeUtxoService, blockStore)
	blockchain.Attach(minerImpl)

	if common.AppEnabled() {
		logger.Infof("Starting App server...")
		// Intialize Transaction Creation API
		transactionCreationService := walletcore.NewTransactionCreationService(keyGeneratorImpl, keyEncodingsImpl, blockchainMsgService, utxoAPI)
		transactionCreationAPI := walletApi.NewTransactionCreationAPIImpl(transactionCreationService)

		// Initialize transaction service and API
		transactionService := appcore.NewTransactionService(transactionCreationAPI)
		transactionAPI := appapi.NewTransactionAPIImpl(transactionService)

		transactionHandler := adapters.NewTransactionAdapter(transactionAPI)
		connService := appcore.NewConnectionEstablishmentService(handshakeAPI)
		internalViewService := appcore.NewInternsalViewService(networkRegistryAPI)
		queryRegistryService := appcore.NewQueryRegistryService(queryRegistryAPI)
		discoveryAppService := appcore.NewDiscoveryService(discoveryAPI)

		appServer := appgrpc.NewServer(connService, internalViewService, queryRegistryService, keyGeneratorApiImpl, transactionHandler, discoveryAppService)

		err := appServer.Start(common.AppPort())
		if err != nil {
			logger.Warnf("couldn't start App server: %v", err)
		} else {
			addrPort, err := appServer.ListeningEndpoint()
			assert.IsNil(err)
			common.SetAppPort(addrPort.Port())
			logger.Infof("App server started on port %v", addrPort)
		}
	}

	logger.Infof("Starting P2P server...")

	grpcServer := grpc.NewServer(handshakeService, networkInfoRegistry, discoveryService)

	grpcServer.Attach(blockchain)

	err = grpcServer.Start(common.P2PPort())
	if err != nil {
		logger.Warnf("couldn't start P2P server: %v", err)
	} else {
		addrPort, err := grpcServer.ListeningEndpoint()
		assert.IsNil(err)
		common.SetP2PPort(addrPort.Port())
		common.SetP2PListeningIpAddr(addrPort.Addr())
		logger.Infof("P2P server started on port %v", addrPort)
	}

	select {}
}
