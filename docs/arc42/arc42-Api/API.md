# 

**Über arc42**

arc42, das Template zur Dokumentation von Software- und
Systemarchitekturen.

Template Version 9.0-DE. (basiert auf der AsciiDoc Version), Juli 2025

Created, maintained and © by Dr. Peter Hruschka, Dr. Gernot Starke and
contributors. Siehe <https://arc42.org>.

# Einführung und Ziele

## Aufgabenstellung
### Inhalt

Das System dient der Verwaltung und Abwicklung von digitalen Währungstransaktionen zwischen Nutzern und erleichtert den Nutzern die Interaktion mit der in der Blockchain repräsentierte Währung.  
Kernfunktionalität ist das Anzeigen von Kontoständen und Transaktionsverläufen sowie die Durchführung und Nachverfolgung von Überweisungen zwischen Nutzern.

Treibende Kräfte sind die Notwendigkeit einer einfachen, transparenten und zuverlässigen Plattform für Transaktionen sowie die Nachvollziehbarkeit aller Bewegungen im System unserers V$Goins.

**Verweise auf Anforderungsdokumente:**
- *User Stories, abgelegt in den [GitHub Issues](https://github.com/users/bjoern621/projects/5/views/1?filterQuery=-status%3ABacklog+label%3A%22rest-api%22)

---

### Motivation

Ziel ist es, Nutzern eine intuitive Möglichkeit zu bieten, digitale Währungen untereinander zu übertragen.  
Das System verbessert die Nachvollziehbarkeit und Transparenz von Transaktionen, reduziert manuelle Fehler und schafft Vertrauen zwischen den Beteiligten.

Aus fachlicher Sicht wird damit die grundlegende Aufgabe der sicheren Kontoführung und Transaktionsverwaltung erfüllt, was den zentralen Mehrwert des Systems darstellt.

---

### Form

| **Use Case / Aufgabe** | **Beschreibung**   | User Stories                                                                                |
|-------------------------|-------------------|---------------------------------------------------------------------------------------------|
| Währung senden | Ein Nutzer kann einem anderen Nutzer einen beliebigen Betrag seiner verfügbaren Währung übertragen. Das System prüft, ob der Sender über ausreichendes Guthaben verfügt.| [US-23 Transaktion](https://github.com/bjoern621/VSP-Blockchain/issues/23)                  |
| Kontostand anzeigen | Ein Nutzer kann seinen aktuellen Kontostand einsehen. | [US-26 Kontostand einsehen](https://github.com/bjoern621/VSP-Blockchain/issues/26)          |
| Transaktionsverlauf anzeigen | Ein Nutzer kann alle vergangenen Transaktionen seines Kontos einsehen, inklusive gesendeter und empfangener Beträge.  | [US-27 Transaktionsverlauf einsehen](https://github.com/bjoern621/VSP-Blockchain/issues/27) |

Alle genannten Anforderungen basieren auf den oben referenzierten User Stories und Akzeptanzkriterien.

## Qualitätsziele

### Inhalt

Die folgenden Qualitätsziele repräsentieren die für die Architektur maßgeblichen Anforderungen gemäß den wichtigsten Stakeholdern.  
Sie orientieren sich an den Qualitätsmerkmalen des ISO/IEC 25010 Standards.

---

### Qualitätsziele

| **Priorität** | **Qualitätsziel**           | **Beschreibung**                                                                                                                                                                    |
|---------------|-----------------------------|-------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| 1             | **Sicherheit** | Alle Transaktionen müssen vor unbefugtem Zugriff geschützt sein. Authentifizierung und Autorisierung sind zentrale Punkte des Systems.                                              |
| 2             | **Zuverlässigkeit** | Das System muss Transaktionen konsistent und fehlerfrei verarbeiten. Datenintegrität hat höchste Priorität, insbesondere bei Transaktionen.                                         |
| 3             | **Wartbarkeit** | Der Quellcode und die Systemarchitektur sollen modular aufgebaut sein, um zukünftige Änderungen (z. B. neue Währungsarten oder Sicherheitsfunktionen) leicht integrieren zu können. |
| 4             | **Performance** | Transaktionen und Kontostandsabfragen sollen ohne merkliche Verzögerung ausgeführt werden, um ein reaktionsschnelles Nutzererlebnis zu gewährleisten.                               |

---
## Stakeholder

| Rolle         | Kontakt               | Erwartungshaltung                                                                                                                      |
|---------------|-----------------------|----------------------------------------------------------------------------------------------------------------------------------------|
| **Product Owner** | wqx847@haw-hamburg.de | Erwartet, dass das System alle Kernfunktionen wie Überweisungen und Transaktionshistorie zuverlässig bereitstellt.                     |
| **Endnutzer / Konto-Inhaber** | n/a                   | Erwartet eine sichere, transparente und einfach bedienbare Plattform, um Währung zu senden, zu empfangen und Transaktionen einzusehen. |
| **Entwicklungsteam** | TBD                   | Erwartet eine stabile und verfügbare Versionsverwaltung (GitHub) und einen stabilen Main branch (durch Code Reviews gesichert)         |

# Randbedingungen

## Technische Randbedingungen

| **Randbedingung** | **Erläuterung**                                                                                                                                                                          |
|--------------------|------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| **Blockchain-Anbindung** | Das System spiegelt reale Währungswerte über eine bestehende **Blockchain-Infrastruktur** wider. Transaktionen im System müssen mit der entsprechenden Blockchain synchronisiert werden. |
| **Blockchain-Protokoll** | Es wird die **V$Goin-Blockchain** verwendet.                                                                                                                                             |
| **Deployment-Umgebung** | Das System wird in der **Cloud-Umgebung** der HAW (ICC) betrieben.                                                                                                                       |

---

## Organisatorische Randbedingungen

| **Randbedingung** | **Erläuterung**                                                                            |
|--------------------|--------------------------------------------------------------------------------------------|
| **Entwicklungsteam** | Das Projekt wird von einem Entwicklungsteam mit 4 Entwicklern umgesetzt.                   |
| **Dokumentationsstandard** | Architektur und Anforderungen werden nach dem **arc42-Template** gepflegt und versioniert. |

---

## Rechtliche und sicherheitsrelevante Randbedingungen

| **Randbedingung** | **Erläuterung** |
|--------------------|-----------------|
| **Verschlüsselung** | Sämtliche Datenübertragungen erfolgen ausschließlich über **HTTPS/TLS**. |
| **Backup & Recovery** | Die Transaktionsdaten sind durch die zugrunde liegende Blockchain dezentral und revisionssicher gespeichert. |

---


# Kontextabgrenzung

## Fachlicher Kontext


| **Kommunikationspartner**                      | **Eingabe an das System**                               | **Ausgabe vom System**                                                                         |
|------------------------------------------------|---------------------------------------------------------|------------------------------------------------------------------------------------------------|
| **Endnutzer**                   | Private Keys, Transaktionsaufträge, Kontostand-Anfragen | Bestätigungsmeldungen, Transaktionshistorie, aktueller Kontostand                              |
| **V$Goin Blockchain-System**                   | Transaktionsdaten, Public Key Hash                      | Transaktionsbestätigungen, UTXOs (verfügbare Zahlungsmittel), Transaktionsverläufe, Event-Logs |
| **Monitoring- oder Logging-Systeme**   | Statusabfragen, Metriken                                | Logs, Health-Check-Responses                                                                   |

## Technischer Kontext
![Diagram](https://www.plantuml.com/plantuml/png/TL91QiCm4BmN-eV159ABFf13ILAQ4WYbrALtMLPZ4NbbP2MbVKz_qezrRTMkPQWEXXtFpiwEX7KRf0_dsbvVWG-vKYFRUlVUQe-TTnGqbHbaYoA2aHU_ojMD8qq1sVDz_eBDqnwvzXUZTDyY6pEbH_63EqchyNhpu0pXaR6UQvsIjgkct4WIM_vvKfKq59rqvLrNJjKNE3XhJUCQaQkAJ0Xjq9OdoHfpTx73y7B-JIeUXC7lVi0YPOf0YFb62mnHqJby1fH68naUQR_HiS0oLLoX_I1LSSofwkYt-lwYOy354Vv2W2p-MQ0OEPl1Q09rAyo2bZswdF5Mh9Ou6xiWl3bMGTnEhY6bh_d5y8Fw0G00)

<details>
    <summary>Code</summary>
    
    ````plantuml
    @startuml
    node "Browser Frontend" as browser
    
    component "REST API Service" as api
    
    node "Lokale V$Goin Node" as localNode
    node "V$Goin-Blockchain" as blockChain
    
    ' Lollipop-Schnittstelle am REST API Service
    interface " " as apiInterface
    apiInterface -- api
    
    ' Browser nutzt die Schnittstelle
    browser --( apiInterface : synchron
    
    ' REST API hängt von Blockchain ab
    interface " " as blockchainApi
    blockchainApi -- localNode
    api --( blockchainApi : asynchron
    localNode -right-> blockChain : asynchron
@enduml
    ````
</details>


| **Kommunikationspartner**    | **Technische Schnittstelle / Kanal** | **Protokoll / Datenformat** | **Beschreibung / Bemerkung**                                                                                      |
|------------------------------|--------------------------------------|-----------------------------|-------------------------------------------------------------------------------------------------------------------|
| **Frontend (Web-Client)**    | HTTPS REST-API                       | JSON                        | Zugriff über Weboberfläche auf Konto- und Transaktionsfunktionen                                                  |
| **Lokale V$Goin Node**       | RPC                                  | Byte                        | Kommunikation mit Blockchain-Knoten zum Senden und Prüfen von Transaktionen, sowie erhalten von Transaktionsdaten |
| **V$Goin Blockchain-System** | RPC                                  | Byte                        | Weiterleiten der erstellten Transaktion an alle weiteren Knoten                                                   |

# Lösungsstrategie
## Technologieentscheidungen
- **Programmiersprache:** Go  
  Go wurde gewählt, da es eine hoch performante Sprache ist und eingebaute Nebenläufigkeit (Goroutines) besitzt.

- **Kommunikationsprotokolle:**
    - **gRPC:** Für performante Kommunikation zu Miner-Network.
    - **REST:** Für einfache externe Anbindungen (Frontend).  
      → gRPC bietet hohe Effizienz und Typsicherheit, REST bleibt für Interoperabilität und menschliche Lesbarkeit bestehen.

- **Frameworks / Tools:**
    - `Docker` zur Containerisierung der Nodes

---

## Top-Level-Architekturentscheidungen
- **Architekturmuster:**  
  Mehrschichtige Client-Server Architektur mit einem Frontend für die Nutzer (in diesem Fall wird nur eine REST Schnittstelle angeboten, es könnte aber eine Frontend Anwendung angeboten werden), einem Backend, welches die Anfragen der Nutzer bearbeitet und an eine lokalen V$Goin Node weiterleitet, welches die Anfragen bearbeitet und an die restlichen Knoten im System weiterleitet.


---

## Entscheidungen zur Erreichung der wichtigsten Qualitätsziele

| **Priorität** | **Qualitätsziel** | **Maßnahmen / Architekturentscheidungen**                                                                                                                                                |
|---------------|------------------|------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| **1** | **Sicherheit** | Verwendung von Signaturen für Transaktionsauthentizität. Alle gRPC-Verbindungen laufen über TLS. REST-Endpunkte sind authentifiziert und autorisiert.                                    |
| **2** | **Zuverlässigkeit** | Validierung jeder Transaktion durch Mehrheitskonsens der Nodes. Unveränderliche Hash-Ketten sichern Datenintegrität. Fehlerbehandlung und Wiederholungsmechanismen für Netzwerkausfälle. |
| **3** | **Wartbarkeit** | Strikte Modultrennung. Pipeline Stages für Code Qualität                                                                                                                                 |
| **4** | **Performance** | gRPC für effiziente Node-Kommunikation. Nebenläufige Verarbeitung (Goroutines) für Transaktionsvalidierung.                                                                              |

# Bausteinsicht

## Whitebox Gesamtsystem
![Diagramm](https://www.plantuml.com/plantuml/png/hLJBQW8n5DqN-WyNANGbxiMAKx0FMefqn5KtSUQgmPX84hLIkkq7z7kwwv_qItgJp4XyGBMQXIJZEPTpppr9orYcxMmYpi-0bbGvGkLQguL13JTQIOiohm0pq0yV0oxyNiBLhbN-2P0kZSK9NAkPp9bUxi7Ar6Ig94eBbUTsseMaSmyfwZdFqAjWKmvliOODKbSpQTZOSYKztlfpviv_uSqSjM2pWUSL-vsS1t95UTJOxNPYUXUtYilg4_bPJN8sjQY3_h0FdFU3p6o_4b4o0O-yh_TpCuopqK1FRJPVPD05QQS7JflN97WVOYNhggg7h99K9kZduxC8mH7bYkGHjHdF4-g0caeBWH2DSPjJp9Bm0ys65dh5cVMtiNwYXFGpPj8HK9x0aE8cSEa6SKIbk7-djyWJAHxI5ECqvuokBYoGh-9M-k3MEdUaoCCxRgpI70Cu6B4D9JkGG1eXpKRY-yiO550B5H8wM7C2jueRu-EpaVP_r2l5kqPSrkjKNoDfhKL-GR8sx4sE8_dkW9wobLKK8KygMuvRRz73IU_gBm00)
<details>
    <summary> Code </summary>

    ````plantuml
    @startuml
    node "Browser Frontend" as browser
    
    ' =====================
    '   System Boundary
    ' =====================
    component "REST API Service" as api {
    
        component "Transaktion" as transaction
        component "Transaktionsverlauf" as verlauf
        component "Konto" as konto
        component "V$Goin-Node-Adapter" as adapter
    }
    
    ' =====================
    '   External Library
    ' =====================
    node "<<extern>>\nV$Goin SPV Node" as lib 
    
    
    
    ' ---------------------------------------------
    ' Browser → System
    ' ---------------------------------------------
    browser --> transaction : erstelle Transaktion
    browser --> verlauf : fragt Verlauf ab
    browser --> konto : Kontoanfragen
    
    ' ---------------------------------------------
    ' System intern
    ' ---------------------------------------------
    transaction --> adapter : gib Transaktionsdaten weiter
    verlauf     --> adapter : hole Historie
    konto     --> adapter : generiere Schlüssel / hole Assets
    
    ' ---------------------------------------------
    ' Adapter → Library
    ' ---------------------------------------------
    adapter --> lib : Adresse/Transaktion Anfragen
    adapter --> lib  : Assets und Historie abfrage
    

    
    @enduml
    ````
</details>


## Blackboxes Ebene 1
### Inhaltsverzeichnis
1. [Transaktion](#transaktion-blackbox)
2. [Transaktionsverlauf](#transaktionsverlauf-blackbox)
3. [Konto](#konto-blackbox)
4. [V$Goin-Node-Adapter](#vgoin-node-adapter-blackbox)

---

### Transaktion (Blackbox)

#### Zweck / Verantwortung
- Entgegennahme und Umwandlung von Transaktionsanfragen
- Weitergabe der Signaturerstellung und verbreiten im Netzwerk durch den Adapter

#### Schnittstelle
- REST-Endpunkt post /transaction
- [OpenAPI Spezifikation](../../../rest-schnittstelle/openapi.yaml)

#### Eingaben / Ausgaben
- Eingaben: Transaktionsdaten vom Client
- Ausgaben: Erfolgs- oder Fehlermeldungen

#### Abhängigkeiten
- V$Goin-Lib-Adapter (Signatur, Weiterleitung)

#### Erfüllte Anforderungen
- [US-23 Transaktion](https://github.com/bjoern621/VSP-Blockchain/issues/23)

#### Qualitätsanforderungen
- Zuverlässigkeit: Ungültige Transaktionen werden abgelehnt

---

### Transaktionsverlauf (Blackbox)

#### Zweck / Verantwortung
- Bereitstellung der Transaktionshistorie für einen bestimmte Wallet Adresse (Public Key Hash)

#### Schnittstelle
- REST-Endpunkt get /history
- [OpenAPI Spezifikation](../../../rest-schnittstelle/openapi.yaml)

#### Eingaben / Ausgaben
- Eingaben: Wallet Adresse (Public Key Hash) base58 encoded
- Ausgaben: Liste von Transaktionen

#### Abhängigkeiten
- V$Goin-Lib-Adapter (History-Abfrage)

#### Erfüllte Anforderungen
- [US-27 Transaktionsverlauf einsehen](https://github.com/bjoern621/VSP-Blockchain/issues/27)

#### Qualitätsanforderungen
- Performance: 99% der Antworten in unter 2s

---

### Konto (Blackbox)

#### Zweck / Verantwortung
- Bereitstellung des Kontostands für eine Wallet Adresse (Public Key Hash)
- Generierung privater Schlüssel
- Ableitung der Public Key Adresse aus einem Private Key

#### Schnittstelle
- REST-Endpunkte /balance und /adress
- [OpenAPI Spezifikation](../../../rest-schnittstelle/openapi.yaml)

#### Eingaben / Ausgaben
- Eingaben: Walled Adresse base58 encoded, Private Key base58 encoded oder Seed für key generierung 
- Ausgaben: Balance, Private Key, Wallet Adresse

#### Abhängigkeiten
- V$Goin-Lib-Adapter (Key-Funktionen, Balance, Key-Ableitung)

#### Erfüllte Anforderungen
- [US-26 Kontostand einsehen](https://github.com/bjoern621/VSP-Blockchain/issues/26)
- [EPIC-24 Konto erstellen](https://github.com/bjoern621/VSP-Blockchain/issues/24)
- [EPIC-94 V$Adresse erhalten](https://github.com/bjoern621/VSP-Blockchain/issues/94)

#### Qualitätsanforderungen
- Performance: 99% der Antworten in unter 2s

---

### V$Goin-Node-Adapter (Blackbox)

#### Zweck / Verantwortung
- Einzige Schnittstelle zur lokalen SPV-Node
- Übersetzung der internen Systemaufrufe zur V$Goin RPC Schnittstelle
- Entkopplung des Systems von V$Goin-Änderungen

#### Schnittstelle
- Funktionen: Signatur, Key-Generierung, Key-Ableitung, Historie, Balance, Broadcast
- [Schnittstellen P2P Netzwerk Wiki](https://github.com/bjoern621/VSP-Blockchain/wiki/Externe-Schnittstelle-Mining-Network)

#### Eingaben / Ausgaben
- Eingaben: Transaktionendaten, Wallet Adressen, Private Keys
- Ausgaben: normalisierte Ergebnisse aus der Node

## Ebene 2

### Whitebox *\<Konto\>*
![Diagramm](https://www.plantuml.com/plantuml/png/VP312i8m38RlWkyGUlBYnQECCNUJUT8d25tKigxTsauO8lWElg5FObtdDb3C8JJD_tzDcbY7nZMbdC_0XnDE4kpeGX9MyBm_8DDbfHKfHy0ohPncGHcoBOIgq609mYlC4JaTNEiH0p7a2dc1fm41rsdp7Vpp3B1DRiXQOe0M-lCVTGVqIwYyCunbKD-crc56OAdKlE1d5AgpRLEKg3ZjgGxIaGFBvMBQXpL4aQ6w4NwqEFuYPzJQC0gr0wxVesE5-v-OX5JkV-u5)
<details>
    <summary>Code</summary>
    
    ````plantuml
    @startuml
    title Level 2 – Komponente "Konto"
    
    skinparam interfaceStyle uml
    
    package "Konto" {
    
        component "Adresse" as Adresse
    
        component "Kontostand" as Kontostand
    }
    
    interface "Blockchain" as IBalanceReq
    Kontostand --( IBalanceReq : <<requires>>
    interface "Keys" as KeyReq
    Adresse --( KeyReq : <<requires>>
    @enduml
    ````
</details>

### Whitebox *\<V$Goin-Node-Adapter\>*
![Diagramm](https://www.plantuml.com/plantuml/png/ZPB1IWCn48RlWkymB1wyvBB7KahjHQHI2egUX-oeORD9JB8jHGJVmJVqIKmJkvijQpM7x2Hy_pz_afqxZzQtZJm_Wp2yy9BWbZOaeOIlZqzOwiPeHSeJ50yNrreejj8LiQiAZITR95sQNIsKGOiDYC3R9-HqvtV1iFDFiq5Uu_ClXl2Me_l13n6WkBUe778ljEe5wE0HfIJ_itD2lv2Qr_m5nL2-8h_LjXxetzEdEyaXBUpJ9bKe_WMLYHg415RfhMAN4O09JAUMNbjXoSrc2H-60XO5YIz71LcA9UrSR7yJghNLcz44hM4T41rDA4GrwfZTV3JErYVzZxY_slGFbE8lKABYrBSulfLuXemQRJ0dLOL_y1i0)

<details>
    <summary>Code</summary>
    
    ````plantuml
    @startuml
    title Level 2 – Komponente "V$Goin-Node-Adapter"
    
    skinparam interfaceStyle uml
    
    package "V$Goin-Node-Adapter" {
    
        component "Transaction-Adapter" as WalletAdapter
    
        component "Blockchain-Adapter" as NetworkAdapter
    }
    interface "V$Goin Node" as Node
    interface "V$Goin Node" as Node2
    WalletAdapter -down-( Node : <<requires>>
    NetworkAdapter --down( Node2 : <<requires>>
    interface "Keys" as IKeyProv
    WalletAdapter -up- IKeyProv : <<provides>>
    interface "Transaction" as TransactionProv
    WalletAdapter -up- TransactionProv : <<provides>>
    interface "Blockchain" as IBalanceProv
    NetworkAdapter -up- IBalanceProv : <<provides>>
    @enduml
    ````
</details>

# Laufzeitsicht

## *\<Bezeichnung Laufzeitszenario 1\>*

- \<hier Laufzeitdiagramm oder Ablaufbeschreibung einfügen\>

- \<hier Besonderheiten bei dem Zusammenspiel der Bausteine in diesem
  Szenario erläutern\>

## *\<Bezeichnung Laufzeitszenario 2\>*

…​

## *\<Bezeichnung Laufzeitszenario n\>*

…​

# Verteilungssicht

## Infrastruktur Ebene 1

***\<Übersichtsdiagramm\>***

Begründung  
*\<Erläuternder Text\>*

Qualitäts- und/oder Leistungsmerkmale  
*\<Erläuternder Text\>*

Zuordnung von Bausteinen zu Infrastruktur  
*\<Beschreibung der Zuordnung\>*

## Infrastruktur Ebene 2

### *\<Infrastrukturelement 1\>*

*\<Diagramm + Erläuterungen\>*

### *\<Infrastrukturelement 2\>*

*\<Diagramm + Erläuterungen\>*

…​

### *\<Infrastrukturelement n\>*

*\<Diagramm + Erläuterungen\>*

# Querschnittliche Konzepte

## 1. Fachliche Konzepte

### 1.1 Domänenmodell V$Goin
Das System bildet ein vereinfachtes Blockchain-basiertes Zahlungssystem ab.  
Zentrale fachliche Objekte sind:

- Adresse (doppelter SHA-256 Hash eines öffentlicher Schlüssel)
- Privater Schlüssel (zur Signatur)
- UTXO / Assets (nicht ausgegebene Transaktionseinheiten)
- Transaktion (signiertes Transferobjekt, welches den Besitzwechsel von Währung representiert)
- Historie (Liste verifizierter Transaktionen)

### 1.2 Validierungsregeln
- Jede Transaktion muss gültig signiert sein.
- Ausreichende UTXOs müssen für Transaktionen verfügbar sein.
- Adressen und Schlüssel dürfen nur Base58 encoded.
- Unvollständige Eingangsdaten werden frühzeitig im API validiert, aber fehlerhafte Daten können erst durch SPV Node bzw. Miner Network validiert werden.

## 2. Sicherheitskonzept
- Die REST Schnittstellen kommunizieren ausschließlich über TLS.

## 3. Persistenz- und Datenhaltungskonzept

### 3.1 Persistenzstrategien
Das System speichert selbst **keine eigenen Blockchain-Daten**, sondern fragt Assets (UTXOs) und Historien dynamisch über die lokale V$Goin Node ab.  
Temporäre Daten:

- Kurzzeit-Caches im Adapter und Backend
- JSON als API-Format, binäre Formate der V$Goin Blockchain

### 3.2 Formatkonzept
- Adressen → Base58
- Schlüssel → Base58 
- Transaktionen → binäre V$Goin-Formate, API JSON

## 4. Kommunikations- und Integrationskonzept

### 4.1 Architekturprinzip
- Der Adapter kapselt sämtliche Interaktionen mit der SPV-Node.
- Das Backend ist vollständig entkoppelt von Blockchain-gRPC-Details.

### 4.2 Kommunikationsmechanismen
- Browser ↔ API: REST/HTTPS
- API ↔ Adapter: interne Funktionsaufrufe
- Adapter ↔ Node: RPC Aufrufe
- Node ↔ Blockchain: gRPC-Kommunikation

### 4.3 Schnittstellen
- Schnittstellen sind in der [OpenAPI Spezifikation](../../../rest-schnittstelle/openapi.yaml) dokumentiert.

## 5. Code-Qualität
- Automatisierte Tests
- Statische Analyse (Sonar)
- Architekturrichtlinienchecks
- Manuelle Code Reviews vor jedem Merge

## 6. Adapter-Pattern
Das System verwendet ein komponentenweites Adapter-Muster, um die SPV-Node von der fachlichen Logik der API zu entkoppeln.  
Der Adapter kapselt sämtliche Low-Level-Funktionen der Node und stellt eine stabile interne Schnittstelle bereit.  
Dadurch können Änderungen an der Blockchain-Technologie vorgenommen werden, ohne das Backend groß anzupassen.

**Motivation:**
- Keine Abhängigkeiten alle Komponenten direkt zur RPC Schnittstelle
- Austauschbarkeit der Blockchain-Implementierung
- Einheitliche Formate intern

# Architekturentscheidungen
## ADR 1: Verwendung von Go als Backend-Programmiersprache
**Status:** Akzeptiert  
**Datum:** 2025-10

### Entscheidung
Das Backend-System wird in der Programmiersprache **Go** implementiert.
### Kontext
Das System interagiert häufig mit externen APIs der V$-Blockchain und internen Services, benötigt hohe Nebenläufigkeitsleistung und effiziente Ressourcennutzung.
### Begründung
Go bietet exzellente Unterstützung für Nebenläufigkeit (Goroutinen), schnelle Ausführung, geringen Speicherverbrauch und erzeugt statische Binaries, die den containerisierten Betrieb vereinfachen.
### Konsequenzen
\+ Hohe Performance bei gleichzeitigen Operationen <br>
\+ Vereinfachte Bereitstellung in Docker <br>
– Ein Teil des Teams weniger Erfahrung mit dieser Sprache  <br>

## ADR 2: Externe APIs als REST/JSON bereitstellen

**Status:** Akzeptiert  
**Datum:** 2025-10

### Entscheidung
Alle externen APIs werden im **REST-Architekturstil** mit **JSON** als Datenaustauschformat umgesetzt.

### Kontext
Endnutzer brauchen verständliche Schnittstelle und externer Client benötigen einfache Schnittstellen.

### Begründung
REST/JSON ist leicht verständlich und dokumentierbar und funktioniert ohne spezielle Tools.

### Konsequenzen
\+ Einfache Integration für Partner <br>
\+ Gute Debugging- und Tool-Unterstützung  <br>
– Geringere Typensicherheit als gRPC  <br>

# Qualitätsanforderungen

## Übersicht der Qualitätsanforderungen
|**Qualitätsziel**           | **Beschreibung**                                                                                                                                                                    | Messkriterium                                                                                                                            |
|-----------------------------|-------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|------------------------------------------------------------------------------------------------------------------------------------------|
|**Sicherheit** | Alle Transaktionen müssen vor unbefugtem Zugriff geschützt sein. Authentifizierung und Autorisierung sind zentrale Punkte des Systems.                                              | 100% der Transaktionen werden vom V$-Blockchain System verfiziert                                                                        |
|**Zuverlässigkeit** | Das System muss Transaktionen konsistent und fehlerfrei verarbeiten. Datenintegrität hat höchste Priorität, insbesondere bei Transaktionen.                                         | 100% der falschen Transaktion werden vom V$-Blockchain System abgelehnt                                                                  |
|**Wartbarkeit** | Der Quellcode und die Systemarchitektur sollen modular aufgebaut sein, um zukünftige Änderungen (z. B. neue Währungsarten oder Sicherheitsfunktionen) leicht integrieren zu können. | Jeder Merge in den Main absolviert alle Codequalitätsstages der Pipeline (Test, Sonar, Architecture-Checking) und wurde manuell reviewed |
|**Performance** | Transaktionen und Kontostandsabfragen sollen ohne merkliche Verzögerung ausgeführt werden, um ein reaktionsschnelles Nutzererlebnis zu gewährleisten.                               | Nutzer erhalten in 99% der Fällen eine Antwort innerhalb 2s                                                                              |

## Qualitätsszenarien

# Risiken und technische Schulden

# Glossar

| Begriff         | Definition         |
|-----------------|--------------------|
| *\<Begriff-1\>* | *\<Definition-1\>* |
| *\<Begriff-2*   | *\<Definition-2\>* |
