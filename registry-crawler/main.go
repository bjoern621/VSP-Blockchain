package main

import (
	"context"
	"crypto/tls"
	"fmt"
	"math/rand"
	"net"
	"net/netip"
	"sort"
	"strconv"
	"strings"
	"time"

	"s3b/vsp-blockchain/registry-crawler/internal/pb"

	"bjoernblessin.de/go-utils/util/logger"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

const (
	defaultSeedNamespace     = "vsp-blockchain"
	defaultSeedEndpointsName = "miner-seed"
	defaultSeedDNSConfigMap  = "seed-dns-hosts"
	defaultSeedDNSHostsKey   = "seed.hosts"
	defaultSeedDNSZone       = "seed.local"
)

type config struct {
	appAddr       string
	p2pPort       uint16
	seedNamespace string
	seedName      string
	seedDNSConfig string
	seedDNSKey    string
	seedDNSZone   string
	seedDNSDebug  dnsDebugConfig
	bootstrap     bootstrapConfig
	updateEvery   time.Duration
	allowedPeerID map[string]struct{}
	overrideIPs   []string
	useTLS        bool
}

type bootstrapConfig struct {
	endpoints []string
}

type dnsDebugConfig struct {
	enabled bool
	count   int
	cidr    netip.Prefix
}

func main() {
	logger.Infof("Running registry crawler...")

	cfg := CurrentConfig()

	clientset, k8sEnabled, err := newKubernetesClientset()
	if err != nil {
		logger.Warnf("kubernetes client not available: %v", err)
	}

	if k8sEnabled {
		go runSeedUpdaterLoop(cfg, clientset)
	} else {
		go runSeedLoggerLoop(cfg)
	}

	select {}
}

// runSeedLoggerLoop periodically fetches seed targets and logs them.
func runSeedLoggerLoop(cfg config) {
	ticker := time.NewTicker(cfg.updateEvery)
	defer ticker.Stop()

	for {
		ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
		seedIPs, seedPort, err := fetchSeedTargets(ctx, cfg)
		cancel()
		if err != nil {
			logger.Warnf("seed targets fetch failed: %v", err)
		} else {
			addresses := make([]string, 0, len(seedIPs))
			for ip := range seedIPs {
				addresses = append(addresses, ip)
			}
			sort.Strings(addresses)
			logger.Infof("seed targets: port=%d addrs=%s", seedPort, strings.Join(addresses, ","))
		}

		<-ticker.C
	}
}

// runSeedUpdaterLoop periodically fetches seed targets and updates the Kubernetes Endpoints and ConfigMap.
func runSeedUpdaterLoop(cfg config, clientset kubernetes.Interface) {
	ticker := time.NewTicker(cfg.updateEvery)
	defer ticker.Stop()

	for {
		ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
		err := updateSeedEndpointsOnce(ctx, cfg, clientset)
		cancel()
		if err != nil {
			logger.Warnf("seed endpoints update failed: %v", err)
		}

		<-ticker.C
	}
}

// updateSeedEndpointsOnce fetches seed targets and updates the Kubernetes Endpoints and ConfigMap once.
func updateSeedEndpointsOnce(ctx context.Context, cfg config, clientset kubernetes.Interface) error {
	seedIPs, seedPort, err := fetchSeedTargets(ctx, cfg)
	if err != nil {
		return err
	}

	addresses := make([]string, 0, len(seedIPs))
	for ip := range seedIPs {
		addresses = append(addresses, ip)
	}
	sort.Strings(addresses)

	ep := buildEndpointsObject(cfg.seedNamespace, cfg.seedName, addresses, seedPort)

	endpointsClient := clientset.CoreV1().Endpoints(cfg.seedNamespace)
	current, err := endpointsClient.Get(ctx, cfg.seedName, metav1.GetOptions{})
	if apierrors.IsNotFound(err) {
		_, err = endpointsClient.Create(ctx, ep, metav1.CreateOptions{})
		return err
	}
	if err != nil {
		return err
	}

	ep.ResourceVersion = current.ResourceVersion
	_, err = endpointsClient.Update(ctx, ep, metav1.UpdateOptions{})
	if err != nil {
		return err
	}

	dnsAddresses := addresses
	source := "registry"
	if len(cfg.overrideIPs) > 0 {
		source = "override"
	}
	if len(cfg.bootstrap.endpoints) > 0 {
		source = source + "+bootstrap"
	}
	if cfg.seedDNSDebug.enabled {
		dnsAddresses = generateRandomIPv4s(cfg.seedDNSDebug.cidr, cfg.seedDNSDebug.count)
		source = "debug-random"
	}

	return updateSeedDNSHostsConfigMap(ctx, cfg, clientset, dnsAddresses, source)
}

// updateSeedDNSHostsConfigMap updates or creates the ConfigMap containing the seed DNS hosts file.
func updateSeedDNSHostsConfigMap(ctx context.Context, cfg config, clientset kubernetes.Interface, ipStrings []string, source string) error {
	hostsBody, err := buildSeedHostsFile(cfg.seedName, cfg.seedNamespace, cfg.seedDNSZone, ipStrings, source)
	if err != nil {
		return err
	}

	cmClient := clientset.CoreV1().ConfigMaps(cfg.seedNamespace)
	current, err := cmClient.Get(ctx, cfg.seedDNSConfig, metav1.GetOptions{})
	if apierrors.IsNotFound(err) {
		cm := &corev1.ConfigMap{
			TypeMeta: metav1.TypeMeta{APIVersion: "v1", Kind: "ConfigMap"},
			ObjectMeta: metav1.ObjectMeta{
				Name:      cfg.seedDNSConfig,
				Namespace: cfg.seedNamespace,
			},
			Data: map[string]string{cfg.seedDNSKey: hostsBody},
		}
		_, err = cmClient.Create(ctx, cm, metav1.CreateOptions{})
		return err
	}
	if err != nil {
		return err
	}

	if current.Data == nil {
		current.Data = map[string]string{}
	}
	current.Data[cfg.seedDNSKey] = hostsBody

	_, err = cmClient.Update(ctx, current, metav1.UpdateOptions{})
	return err
}

func buildSeedHostsFile(seedServiceName, namespace, zone string, ipStrings []string, source string) (string, error) {
	zone = strings.TrimSpace(zone)
	if zone == "" {
		return "", fmt.Errorf("seed dns zone is empty")
	}
	zone = strings.TrimSuffix(zone, ".")

	baseName := fmt.Sprintf("%s.%s", seedServiceName, zone)
	namespacedName := fmt.Sprintf("%s.%s.%s", seedServiceName, namespace, zone)

	lines := []string{
		"# Managed by registry-crawler. One line per IP.",
		fmt.Sprintf("# generated_at=%s source=%s", time.Now().UTC().Format(time.RFC3339Nano), strings.TrimSpace(source)),
	}
	for _, ip := range ipStrings {
		ip = strings.TrimSpace(ip)
		if ip == "" {
			continue
		}
		lines = append(lines, fmt.Sprintf("%s %s. %s.", ip, baseName, namespacedName))
	}
	lines = append(lines, "")
	return strings.Join(lines, "\n"), nil
}

func generateRandomIPv4s(prefix netip.Prefix, count int) []string {
	addr := prefix.Masked().Addr()
	base := addr.As4()
	bits := 32 - prefix.Bits()
	max := uint32(1) << uint32(bits)
	if max <= 2 {
		return []string{addr.String()}
	}

	rng := rand.New(rand.NewSource(time.Now().UnixNano()))
	seen := map[uint32]struct{}{}
	res := make([]string, 0, count)

	for len(res) < count {
		offset := uint32(rng.Intn(int(max-2))) + 1
		if _, ok := seen[offset]; ok {
			continue
		}
		seen[offset] = struct{}{}

		ip := netip.AddrFrom4([4]byte{base[0], base[1], base[2], base[3]}).Next()
		for i := uint32(1); i < offset; i++ {
			ip = ip.Next()
		}
		res = append(res, ip.String())
	}

	sort.Strings(res)
	return res
}

func fetchSeedTargets(ctx context.Context, cfg config) (map[string]struct{}, int32, error) {
	bootstrapTargets := parseBootstrapTargets(cfg)

	if len(cfg.overrideIPs) > 0 {
		ips := map[string]struct{}{}
		for _, ipString := range cfg.overrideIPs {
			ips[ipString] = struct{}{}
		}
		for ip := range bootstrapTargets {
			ips[ip] = struct{}{}
		}
		return ips, int32(cfg.p2pPort), nil
	}

	conn, err := dialAppGRPC(ctx, cfg.appAddr, cfg.useTLS)
	if err != nil {
		return nil, 0, err
	}
	defer conn.Close()

	client := pb.NewAppServiceClient(conn)

	bootstrapErr := bootstrapConnect(ctx, client, cfg)
	if bootstrapErr != nil {
		logger.Warnf("bootstrap connect failed: %v", bootstrapErr)
	}

	resp, err := client.GetPeerRegistry(ctx, &pb.GetPeerRegistryRequest{})
	if err != nil {
		return bootstrapTargets, int32(cfg.p2pPort), nil
	}

	ips := map[string]struct{}{}
	var port int32 = int32(cfg.p2pPort)
	for ip := range bootstrapTargets {
		ips[ip] = struct{}{}
	}

	for _, entry := range resp.GetEntries() {
		if entry == nil || entry.ListeningEndpoint == nil {
			continue
		}

		if len(cfg.allowedPeerID) > 0 {
			if _, ok := cfg.allowedPeerID[entry.PeerId]; !ok {
				continue
			}
		} else {
			if !contains(entry.SupportedServices, "miner") {
				continue
			}
			if entry.ConnectionState != "connected" {
				continue
			}
		}

		addr, ok := netip.AddrFromSlice(entry.ListeningEndpoint.IpAddress)
		if !ok {
			continue
		}

		if p := int32(entry.ListeningEndpoint.ListeningPort); p > 0 {
			port = p
		}

		ips[addr.String()] = struct{}{}
	}

	return ips, port, nil
}

func parseBootstrapTargets(cfg config) map[string]struct{} {
	res := map[string]struct{}{}

	endpoints := make([]string, 0, len(cfg.bootstrap.endpoints))
	endpoints = append(endpoints, cfg.bootstrap.endpoints...)
	if len(endpoints) == 0 {
		for _, ip := range cfg.overrideIPs {
			ip = strings.TrimSpace(ip)
			if ip == "" {
				continue
			}
			endpoints = append(endpoints, net.JoinHostPort(ip, strconv.Itoa(int(cfg.p2pPort))))
		}
	}

	for _, token := range endpoints {
		host, port, err := splitHostPortOrDefault(token, int(cfg.p2pPort))
		if err != nil {
			continue
		}
		_ = port

		ip := netip.Addr{}
		if parsed, err := netip.ParseAddr(host); err == nil {
			ip = parsed
			if ip.Is4() {
				res[ip.String()] = struct{}{}
			}
			continue
		}
	}

	return res
}

func bootstrapConnect(ctx context.Context, client pb.AppServiceClient, cfg config) error {
	endpoints := make([]string, 0, len(cfg.bootstrap.endpoints))
	endpoints = append(endpoints, cfg.bootstrap.endpoints...)
	if len(endpoints) == 0 {
		for _, ip := range cfg.overrideIPs {
			ip = strings.TrimSpace(ip)
			if ip == "" {
				continue
			}
			endpoints = append(endpoints, net.JoinHostPort(ip, strconv.Itoa(int(cfg.p2pPort))))
		}
	}

	var lastErr error
	for _, token := range endpoints {
		host, port, err := splitHostPortOrDefault(token, int(cfg.p2pPort))
		if err != nil {
			lastErr = err
			continue
		}

		ips := []netip.Addr{}
		if parsed, err := netip.ParseAddr(host); err == nil {
			ips = append(ips, parsed)
		} else {
			resolved, err := net.DefaultResolver.LookupIPAddr(ctx, host)
			if err != nil {
				lastErr = err
				continue
			}
			for _, r := range resolved {
				if r.IP == nil {
					continue
				}
				addr, ok := netip.AddrFromSlice(r.IP)
				if !ok {
					continue
				}
				ips = append(ips, addr)
			}
		}

		for _, ip := range ips {
			if !ip.Is4() && !ip.Is6() {
				continue
			}
			resp, err := client.ConnectTo(ctx, &pb.ConnectToRequest{IpAddress: ip.AsSlice(), Port: uint32(port)})
			if err != nil {
				lastErr = err
				continue
			}
			if resp != nil && resp.Success {
				return nil
			}
			if resp != nil && !resp.Success {
				lastErr = fmt.Errorf("connect_to failed: %s", strings.TrimSpace(resp.ErrorMessage))
			}
		}
	}

	return lastErr
}

func splitHostPortOrDefault(token string, defaultPort int) (string, int, error) {
	token = strings.TrimSpace(token)
	if token == "" {
		return "", 0, fmt.Errorf("empty endpoint")
	}

	host, portString, err := net.SplitHostPort(token)
	if err != nil {
		return token, defaultPort, nil
	}
	if host == "" {
		return "", 0, fmt.Errorf("empty host")
	}
	port, err := strconv.Atoi(portString)
	if err != nil {
		return "", 0, err
	}
	return host, port, nil
}

func dialAppGRPC(ctx context.Context, addr string, useTLS bool) (*grpc.ClientConn, error) {
	if useTLS {
		creds := credentials.NewTLS(&tls.Config{MinVersion: tls.VersionTLS12})
		return grpc.DialContext(ctx, addr, grpc.WithTransportCredentials(creds))
	}

	return grpc.DialContext(ctx, addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
}

func buildEndpointsObject(namespace, name string, ipStrings []string, port int32) *corev1.Endpoints {
	addresses := make([]corev1.EndpointAddress, 0, len(ipStrings))
	for _, ipString := range ipStrings {
		ipString = strings.TrimSpace(ipString)
		if ipString == "" {
			continue
		}
		addresses = append(addresses, corev1.EndpointAddress{IP: ipString})
	}

	subsets := []corev1.EndpointSubset{}
	if len(addresses) > 0 {
		subsets = append(subsets, corev1.EndpointSubset{
			Addresses: addresses,
			Ports: []corev1.EndpointPort{{
				Name:     "p2p",
				Port:     port,
				Protocol: corev1.ProtocolTCP,
			}},
		})
	}

	return &corev1.Endpoints{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "v1",
			Kind:       "Endpoints",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Subsets: subsets,
	}
}

func newKubernetesClientset() (kubernetes.Interface, bool, error) {
	if K8SDisabled() {
		return nil, false, nil
	}

	restCfg, err := rest.InClusterConfig()
	if err != nil {
		return nil, false, err
	}

	cs, err := kubernetes.NewForConfig(restCfg)
	if err != nil {
		return nil, false, err
	}

	return cs, true, nil
}

func contains(items []string, needle string) bool {
	for _, item := range items {
		if item == needle {
			return true
		}
	}
	return false
}
