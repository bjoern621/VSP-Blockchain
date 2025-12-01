# Getting Started Development

Dieses Dokument beschreibt, wie eine lokale Entwicklungsumgebung für VSP-Blockchain eingerichtet wird.

## Voraussetzungen

Stelle sicher, dass die folgenden Programme auf deinem System installiert sind:

-   **Go 1.25.3** (`winget install GoLang.Go`)
-   **Protoc** Buf (Protoc Wrapper) (`winget install -e --id bufbuild.buf`)
-   **Go protoc plugins**

    ```
    go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
    go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
    ```

-   **Docker, Docker Desktop und Kubernetes lokal** Für die Containerisierung und lokales Testen der gesamten Umgebung ([Download Docker](https://www.docker.com/products/docker-desktop/))
-   **Git** Versionskontrolle ([Download Git](https://git-scm.com/downloads) oder `winget install Git.Git`)
-   **Task** Zum Ausführen der Build und Deployment Tasks ([Download](https://taskfile.dev/docs/installation) oder `winget install Task.Task`)
-   **Kubernetes CLI** (`winget install Kubernetes.kubectl`)

## Setup

Klone das Repository, falls noch nicht geschehen:

```bash
git clone https://github.com/bjoern621/VSP-Blockchain.git
```

Stelle sicher, dass Kubernetes in Docker Desktop aktiviert ist:

-   Docker Desktop öffnen → Settings → Kubernetes → "Enable Kubernetes" aktivieren

## 1. Review/Test-Umgebung (Kubernetes Deployment)

Das komplette System kann in einem lokalen Kubernetes-Cluster deployed werden. Diese Umgebung sollte zum testen / reviewen eines Branches genutzt werden.

### Deployment

1. Docker Desktop & lokales Cluster starten
2. `kubectl config use-context docker-desktop` - Zum lokalen Cluster wechseln
3. `task deploy` im Root-Directory

`task deploy`...

1. Baut die Docker Images für REST-API und Miner
2. Löscht alte Deployments (falls vorhanden)
3. Deployed das System in den `vsp-blockchain` Kubernetes Namespace

### Zugriff auf die Services

Nach erfolgreichem Deployment sind folgende Services verfügbar:

-   **REST API**: [`http://localhost:8080`](http://localhost:8080)
-   **Miner Pods**: 3 StatefulSet Pods mit gRPC auf Port 50051

## 2. Lokale Entwicklung

### 2.1 Debuggen

Zum Debuggen in VS Code wird **Delve** benötigt. Falls noch nicht installiert:

```bash
go install github.com/go-delve/delve/cmd/dlv@latest
```

Das Projekt enthält bereits vorkonfigurierte Debug-Konfigurationen in `.vscode/launch.json`:

-   **Launch P2P-Blockchain** - Startet den Miner im Debug-Modus
-   **Launch REST-Schnittstelle** - Startet die REST-API im Debug-Modus

Die Services werden standardmäßig mit `LOG_LEVEL=DEBUG` gestartet, dies kann in der `launch.json` geändert werden.

### 2.2 Nur Starten

Alternativ können die Services auch nur gestartet werden. Dann sind aber keine Breakpoints, etc. möglich.

#### REST-API lokal starten

```bash
cd rest-schnittstelle
go run main.go
```

REST-API läuft auf [`http://localhost:8080`](http://localhost:8080)

#### Miner lokal starten

```bash
cd p2p-blockchain
go run main.go
```

Miner läuft auf Port `50051` (gRPC)

### 2.3. Protocol Buffers neu generieren

Falls `.proto` Dateien geändert wurden oder die Ordner \[p2p-blockchain|rest-schnittstelle\]/internal/pb/ fehlen bzw. nicht aktuell sind:

```bash
cd <root>
task miner:proto
task rest:proto
```
