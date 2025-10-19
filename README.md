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

### Transaktionsablauf

1. **Transaktionserstellung**: Client erstellt eine Transaktion über REST API
2. **Validierung**: Node prüft Gültigkeit (Signatur, Kontostand, Format)
3. **Broadcasting**: Transaktion wird via RPC an alle Nodes verteilt
4. **Pool-Aufnahme**: Transaktion wird in den Transaction Pool aufgenommen
5. **Mining/Blockbildung**: Transaktionen werden zu einem Block zusammengefasst
6. **Konsensus**: Nodes einigen sich auf den neuen Block
7. **Blockchain-Update**: Block wird zur Blockchain hinzugefügt
8. **Kontoupdate**: Kontostände werden aktualisiert

## RPC Kommunikationsprotokoll

### Node-zu-Node Kommunikation

Das interne RPC-Protokoll verwendet ein binäres Format für effiziente Kommunikation.

### Konsensus-Mechanismus

Das System verwendet einen **Proof-of-Authority (PoA)** Konsensus:

1. Autorisierte Nodes können neue Blöcke vorschlagen
2. Mindestens 2/3 der Nodes müssen zustimmen
3. Bei Zustimmung wird Block zur Blockchain hinzugefügt
4. Konflikte werden durch Timestamp und Node-Priorität aufgelöst

## Monitoring und Logging

### Logs

Logs werden in folgende Kategorien unterteilt:

- **transaction.log**: Alle Transaktionsereignisse
- **consensus.log**: Konsensus-Aktivitäten
- **rpc.log**: RPC-Kommunikation
- **rest.log**: REST API Zugriffe
- **error.log**: Fehler und Exceptions

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
