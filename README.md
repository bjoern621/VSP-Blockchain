# VSP-Blockchain

## Übersicht

VSP-Blockchain ist ein verteiltes Blockchain-System zur Abwicklung imaginärer Transaktionen. Das System ermöglicht die Überweisung von Beträgen zwischen Konten in einem dezentralen Netzwerk mit garantierter Konsistenz und Nachvollziehbarkeit.

## Systemarchitektur

### Komponenten

Das System besteht aus folgenden Hauptkomponenten:

#### 1. Blockchain-Node
- Verwaltet die lokale Kopie der Blockchain
- Validiert eingehende Transaktionen
- Führt Konsensus-Mechanismus aus
- Synchronisiert mit anderen Nodes

#### 2. REST API Gateway
- Externes Interface für Client-Anwendungen
- Authentifizierung und Autorisierung
- Request-Validierung
- Rate-Limiting

#### 3. RPC Communication Layer
- Interne Kommunikation zwischen Nodes
- Synchronisation der Blockchain
- Propagierung neuer Blöcke und Transaktionen
- Konsensus-Nachrichten

#### 4. Transaction Pool
- Speichert noch nicht verarbeitete Transaktionen
- Priorisierung nach Zeitstempel
- Duplikatserkennung

#### 5. Persistenz-Schicht
- Speicherung der Blockchain-Daten
- Kontostände (Account Ledger)
- Transaktionshistorie

## Transaktionsmodell

### Transaktionsstruktur

```json
{
  "id": "tx_1234567890",
  "timestamp": 1697718899000,
  "from": "account_A",
  "to": "account_B",
  "amount": 100.50,
  "signature": "digital_signature_hash",
  "nonce": 42
}
```

### Transaktionsablauf

1. **Transaktionserstellung**: Client erstellt eine Transaktion über REST API
2. **Validierung**: Node prüft Gültigkeit (Signatur, Kontostand, Format)
3. **Broadcasting**: Transaktion wird via RPC an alle Nodes verteilt
4. **Pool-Aufnahme**: Transaktion wird in den Transaction Pool aufgenommen
5. **Mining/Blockbildung**: Transaktionen werden zu einem Block zusammengefasst
6. **Konsensus**: Nodes einigen sich auf den neuen Block
7. **Blockchain-Update**: Block wird zur Blockchain hinzugefügt
8. **Kontoupdate**: Kontostände werden aktualisiert

## REST API

### Endpunkte

#### Account Management

**Konto erstellen**
```
POST /api/v1/accounts
Content-Type: application/json

{
  "name": "Alice",
  "initial_balance": 1000.0
}

Response: 201 Created
{
  "account_id": "account_A",
  "balance": 1000.0,
  "created_at": "2023-10-19T14:28:19Z"
}
```

**Kontostand abfragen**
```
GET /api/v1/accounts/{account_id}/balance

Response: 200 OK
{
  "account_id": "account_A",
  "balance": 1000.0,
  "last_updated": "2023-10-19T14:28:19Z"
}
```

#### Transaktionen

**Transaktion erstellen**
```
POST /api/v1/transactions
Content-Type: application/json

{
  "from": "account_A",
  "to": "account_B",
  "amount": 100.50
}

Response: 201 Created
{
  "transaction_id": "tx_1234567890",
  "status": "pending",
  "timestamp": "2023-10-19T14:28:19Z"
}
```

**Transaktionsstatus abfragen**
```
GET /api/v1/transactions/{transaction_id}

Response: 200 OK
{
  "transaction_id": "tx_1234567890",
  "from": "account_A",
  "to": "account_B",
  "amount": 100.50,
  "status": "confirmed",
  "block_number": 1234,
  "confirmations": 6
}
```

**Transaktionshistorie abrufen**
```
GET /api/v1/accounts/{account_id}/transactions?limit=10&offset=0

Response: 200 OK
{
  "transactions": [
    {
      "transaction_id": "tx_1234567890",
      "type": "debit",
      "amount": 100.50,
      "counterparty": "account_B",
      "timestamp": "2023-10-19T14:28:19Z"
    }
  ],
  "total": 42,
  "limit": 10,
  "offset": 0
}
```

#### Blockchain

**Blockchain-Info abrufen**
```
GET /api/v1/blockchain/info

Response: 200 OK
{
  "chain_length": 1234,
  "latest_block_hash": "00000abc123...",
  "total_transactions": 5678,
  "pending_transactions": 12,
  "connected_nodes": 5
}
```

**Block abrufen**
```
GET /api/v1/blockchain/blocks/{block_number}

Response: 200 OK
{
  "block_number": 1234,
  "timestamp": "2023-10-19T14:28:19Z",
  "previous_hash": "00000def456...",
  "hash": "00000abc123...",
  "transactions": [...],
  "miner": "node_1",
  "nonce": 987654
}
```

## RPC Kommunikationsprotokoll

### Node-zu-Node Kommunikation

Das interne RPC-Protokoll verwendet ein binäres Format für effiziente Kommunikation.

#### Nachrichtentypen

**1. HELLO - Node-Registrierung**
```
Type: HELLO
Payload:
  - node_id: string
  - version: string
  - capabilities: list
  - blockchain_height: int
```

**2. SYNC_REQUEST - Blockchain-Synchronisation**
```
Type: SYNC_REQUEST
Payload:
  - from_block: int
  - to_block: int
```

**3. SYNC_RESPONSE - Blockchain-Daten**
```
Type: SYNC_RESPONSE
Payload:
  - blocks: list[Block]
```

**4. NEW_TRANSACTION - Neue Transaktion propagieren**
```
Type: NEW_TRANSACTION
Payload:
  - transaction: Transaction
```

**5. NEW_BLOCK - Neuer Block propagieren**
```
Type: NEW_BLOCK
Payload:
  - block: Block
```

**6. CONSENSUS_REQUEST - Konsensus-Anfrage**
```
Type: CONSENSUS_REQUEST
Payload:
  - proposed_block: Block
  - proposer_id: string
```

**7. CONSENSUS_VOTE - Konsensus-Stimme**
```
Type: CONSENSUS_VOTE
Payload:
  - block_hash: string
  - vote: bool
  - voter_id: string
```

### Konsensus-Mechanismus

Das System verwendet einen **Proof-of-Authority (PoA)** Konsensus:

1. Autorisierte Nodes können neue Blöcke vorschlagen
2. Mindestens 2/3 der Nodes müssen zustimmen
3. Bei Zustimmung wird Block zur Blockchain hinzugefügt
4. Konflikte werden durch Timestamp und Node-Priorität aufgelöst

## Deployment

### Systemanforderungen

- **CPU**: Mindestens 2 Cores
- **RAM**: Mindestens 4 GB
- **Speicher**: 50 GB SSD
- **Netzwerk**: Stabile Internetverbindung
- **OS**: Linux (Ubuntu 20.04+), macOS, Windows 10+

### Installation

#### 1. Node-Setup

```bash
# Repository klonen
git clone https://github.com/bjoern621/VSP-Blockchain.git
cd VSP-Blockchain

# Dependencies installieren
npm install
# oder
pip install -r requirements.txt

# Konfiguration erstellen
cp config.example.yaml config.yaml
```

#### 2. Konfiguration

**config.yaml**
```yaml
node:
  id: node_1
  role: validator  # validator oder observer
  
rest_api:
  host: 0.0.0.0
  port: 8080
  enable_cors: true
  auth_enabled: true
  
rpc:
  host: 0.0.0.0
  port: 9090
  max_connections: 100
  
blockchain:
  genesis_accounts:
    - id: account_genesis
      balance: 1000000.0
  block_time: 10  # Sekunden
  max_block_size: 1000  # Transaktionen
  
network:
  peers:
    - node_2:9090
    - node_3:9090
  discovery_enabled: true
  
storage:
  type: sqlite  # sqlite, postgresql, mongodb
  path: ./data/blockchain.db
```

#### 3. Node starten

```bash
# Node starten
npm start
# oder
python main.py

# Im Hintergrund (Production)
pm2 start main.py --name vsp-blockchain-node
```

### Multi-Node Setup

Für ein verteiltes Netzwerk mit mehreren Nodes:

```bash
# Terminal 1 - Node 1
export NODE_ID=node_1
export REST_PORT=8080
export RPC_PORT=9090
npm start

# Terminal 2 - Node 2
export NODE_ID=node_2
export REST_PORT=8081
export RPC_PORT=9091
export PEER_NODES=localhost:9090
npm start

# Terminal 3 - Node 3
export NODE_ID=node_3
export REST_PORT=8082
export RPC_PORT=9092
export PEER_NODES=localhost:9090,localhost:9091
npm start
```

## Verwendungsbeispiele

### Beispiel 1: Einfache Transaktion

```bash
# Konto A erstellen
curl -X POST http://localhost:8080/api/v1/accounts \
  -H "Content-Type: application/json" \
  -d '{"name": "Alice", "initial_balance": 1000.0}'

# Response: {"account_id": "account_A", "balance": 1000.0}

# Konto B erstellen
curl -X POST http://localhost:8080/api/v1/accounts \
  -H "Content-Type: application/json" \
  -d '{"name": "Bob", "initial_balance": 500.0}'

# Response: {"account_id": "account_B", "balance": 500.0}

# Transaktion: 100 von A nach B
curl -X POST http://localhost:8080/api/v1/transactions \
  -H "Content-Type: application/json" \
  -d '{"from": "account_A", "to": "account_B", "amount": 100.0}'

# Response: {"transaction_id": "tx_123", "status": "pending"}

# Status prüfen
curl http://localhost:8080/api/v1/transactions/tx_123

# Kontostände prüfen
curl http://localhost:8080/api/v1/accounts/account_A/balance
# Response: {"account_id": "account_A", "balance": 900.0}

curl http://localhost:8080/api/v1/accounts/account_B/balance
# Response: {"account_id": "account_B", "balance": 600.0}
```

### Beispiel 2: Python Client

```python
import requests

BASE_URL = "http://localhost:8080/api/v1"

# Konto erstellen
response = requests.post(
    f"{BASE_URL}/accounts",
    json={"name": "Charlie", "initial_balance": 2000.0}
)
account_id = response.json()["account_id"]

# Transaktion senden
response = requests.post(
    f"{BASE_URL}/transactions",
    json={
        "from": account_id,
        "to": "account_B",
        "amount": 250.0
    }
)
tx_id = response.json()["transaction_id"]

# Auf Bestätigung warten
import time
while True:
    response = requests.get(f"{BASE_URL}/transactions/{tx_id}")
    status = response.json()["status"]
    if status == "confirmed":
        print("Transaktion bestätigt!")
        break
    time.sleep(2)
```

## Monitoring und Logging

### Logs

Logs werden in folgende Kategorien unterteilt:

- **transaction.log**: Alle Transaktionsereignisse
- **consensus.log**: Konsensus-Aktivitäten
- **rpc.log**: RPC-Kommunikation
- **rest.log**: REST API Zugriffe
- **error.log**: Fehler und Exceptions

### Metriken

Verfügbare Prometheus-Metriken:

- `blockchain_height`: Aktuelle Blockchain-Höhe
- `pending_transactions`: Anzahl wartender Transaktionen
- `transactions_per_second`: Durchsatz
- `block_time_seconds`: Zeit zwischen Blöcken
- `connected_peers`: Anzahl verbundener Nodes
- `api_request_duration`: REST API Antwortzeiten

## Sicherheit

### Authentifizierung

REST API verwendet JWT (JSON Web Tokens):

```bash
# Login
curl -X POST http://localhost:8080/api/v1/auth/login \
  -d '{"username": "admin", "password": "secret"}'

# Response: {"token": "eyJhbGc..."}

# API-Aufruf mit Token
curl -H "Authorization: Bearer eyJhbGc..." \
  http://localhost:8080/api/v1/accounts
```

### Transaktionssignierung

Alle Transaktionen werden mit privaten Schlüsseln signiert:

- **Algorithmus**: ECDSA (secp256k1)
- **Hash-Funktion**: SHA-256
- **Signatur-Format**: DER-encoded

### Netzwerksicherheit

- **RPC-Verschlüsselung**: TLS 1.3
- **Peer-Authentifizierung**: Mutual TLS (mTLS)
- **DDoS-Schutz**: Rate-Limiting, Connection-Limits

## Troubleshooting

### Häufige Probleme

**Problem**: Node kann sich nicht mit Peers verbinden

```bash
# Prüfe Netzwerkkonnektivität
telnet peer_node_ip 9090

# Prüfe Firewall-Regeln
sudo ufw status

# Prüfe Node-Logs
tail -f logs/rpc.log
```

**Problem**: Transaktion bleibt im Status "pending"

- Prüfe ob genügend Nodes aktiv sind (Konsensus benötigt 2/3)
- Prüfe ob Kontostand ausreichend ist
- Prüfe Transaction Pool: `curl http://localhost:8080/api/v1/blockchain/info`

**Problem**: Blockchain-Synchronisation schlägt fehl

```bash
# Stoppe Node
npm stop

# Lösche lokale Blockchain (Achtung: Datenverlust!)
rm -rf data/blockchain.db

# Starte Node neu (lädt Blockchain von Peers)
npm start
```

## Performance

### Benchmarks

Typische Performance auf Standard-Hardware:

- **Transaktions-Durchsatz**: 100-500 TPS
- **Block-Zeit**: 10 Sekunden (konfigurierbar)
- **REST API Latenz**: < 50ms (lokale Anfragen)
- **RPC Latenz**: < 10ms (internes Netzwerk)
- **Maximale Nodes**: 100+ (Proof-of-Authority)

### Optimierungen

- Verwende SSD-Speicher für Datenbank
- Aktiviere Connection-Pooling
- Nutze Redis für Transaction Pool
- Implementiere Caching für häufige Abfragen

## Entwicklung und Tests

### Tests ausführen

```bash
# Unit Tests
npm test

# Integration Tests
npm run test:integration

# E2E Tests
npm run test:e2e

# Coverage Report
npm run test:coverage
```

### Entwicklungsumgebung

```bash
# Entwicklungsserver mit Hot-Reload
npm run dev

# Code-Formatierung
npm run format

# Linting
npm run lint

# Build
npm run build
```

## Lizenz

MIT License - siehe LICENSE Datei

## Kontakt und Support

- **GitHub**: https://github.com/bjoern621/VSP-Blockchain
- **Issues**: https://github.com/bjoern621/VSP-Blockchain/issues
- **Dokumentation**: https://github.com/bjoern621/VSP-Blockchain/wiki

## Roadmap

### Version 1.0 (Aktuell)
- [x] Grundlegende Blockchain-Funktionalität
- [x] REST API
- [x] RPC-Kommunikation
- [x] Proof-of-Authority Konsensus

### Version 2.0 (Geplant)
- [ ] Smart Contracts
- [ ] Web-Dashboard
- [ ] Mobile Apps
- [ ] Verbesserte Performance (1000+ TPS)
- [ ] Shard-basierte Skalierung

### Version 3.0 (Zukunft)
- [ ] Cross-Chain Interoperabilität
- [ ] Zero-Knowledge Proofs
- [ ] Quantenresistente Kryptographie