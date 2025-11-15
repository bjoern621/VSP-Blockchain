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

| **Use Case / Aufgabe** | **Beschreibung**   | User Stories                     |
|-------------------------|-------------------|----------------------------------|
| Währung senden | Ein Nutzer kann einem anderen Nutzer einen beliebigen Betrag seiner verfügbaren Währung übertragen. Das System prüft, ob der Sender über ausreichendes Guthaben verfügt.| US-1 Transaktion                 |
| Kontostand anzeigen | Ein Nutzer kann seinen aktuellen Kontostand einsehen. | US-2 Kontostand einsehen         |
| Transaktionsverlauf anzeigen | Ein Nutzer kann alle vergangenen Transaktionen seines Kontos einsehen, inklusive gesendeter und empfangener Beträge.  | US-3 Transaktionsverlauf ansehen |

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
![Diagram](https://www.plantuml.com/plantuml/png/
````plantuml
@startuml
node "Browser Frontend" as browser

component "REST API Service" as api


  component "V$Goin-Blockchain" as blockChain

' Lollipop-Schnittstelle am REST API Service
interface " " as apiInterface
apiInterface -- api

' Browser nutzt die Schnittstelle
browser --( apiInterface : HTTPS

' REST API hängt von Blockchain ab
interface " " as blockchainApi
blockchainApi -- blockChain
api --( blockchainApi : gRPC
@enduml
````
)

| **Kommunikationspartner** | **Technische Schnittstelle / Kanal** | **Protokoll / Datenformat** | **Beschreibung / Bemerkung** |
|----------------------------|-------------------------------------|------------------------------|-------------------------------|
| **Frontend (Web-Client)** | HTTPS REST-API | JSON | Zugriff über Weboberfläche auf Konto- und Transaktionsfunktionen |
| **V$Goin Blockchain-System** | RPC | JSON | Kommunikation mit Blockchain-Knoten zum Senden und Prüfen von Transaktionen |

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
  Mehrschichtige Client-Server Architektur mit einem Frontend für die Nutzer, einem Backend, welches die Anfragen der Nutzer bearbeitet und mit einem weiteren Blockchain-Backend, welches die Währung repräsentiert, interagiert.


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

***\<Übersichtsdiagramm\>***
````plantuml
@startuml
node "Browser Frontend" as browser
component "REST API Service" as api {
component "Transaktion" as transaction
component "Transaktionsverlauf" as verlauf
component "Kontostand" as konto
component "Blockchain" as blocks
verlauf --> blocks: liest Transaktionsverlauf aus
konto --> blocks: liest Kontostand aus
}
node "V$Goin-Blockchain" as blockChain


'Browser zu Server
browser -down-> transaction: führt Transaktion aus
browser --> verlauf : fragt Verlauf ab
browser --> konto : fragt Kontostand ab

'Server zu Blockchain
transaction -down--> blockChain: gibt Transaktion weiter
blocks -> blockChain: Blöcke werden weitergegeben
blockChain -> blocks
@enduml
````
Begründung  
*\<Erläuternder Text\>*

Enthaltene Bausteine  
*\<Beschreibung der enthaltenen Bausteine (Blackboxen)\>*

Wichtige Schnittstellen  
*\<Beschreibung wichtiger Schnittstellen\>*

### \<Name Blackbox 1\>

*\<Zweck/Verantwortung\>*

*\<Schnittstelle(n)\>*

*\<(Optional) Qualitäts-/Leistungsmerkmale\>*

*\<(Optional) Ablageort/Datei(en)\>*

*\<(Optional) Erfüllte Anforderungen\>*

*\<(optional) Offene Punkte/Probleme/Risiken\>*

### \<Name Blackbox 2\>

*\<Blackbox-Template\>*

### \<Name Blackbox n\>

*\<Blackbox-Template\>*

### \<Name Schnittstelle 1\>

…​

### \<Name Schnittstelle m\>

## Ebene 2

### Whitebox *\<Baustein 1\>*

*\<Whitebox-Template\>*

### Whitebox *\<Baustein 2\>*

*\<Whitebox-Template\>*

…​

### Whitebox *\<Baustein m\>*

*\<Whitebox-Template\>*

## Ebene 3

### Whitebox \<\_Baustein x.1\_\>

*\<Whitebox-Template\>*

### Whitebox \<\_Baustein x.2\_\>

*\<Whitebox-Template\>*

### Whitebox \<\_Baustein y.1\_\>

*\<Whitebox-Template\>*

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

## *\<Konzept 1\>*

*\<Erklärung\>*

## *\<Konzept 2\>*

*\<Erklärung\>*

…​

## *\<Konzept n\>*

*\<Erklärung\>*

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

## Qualitätsszenarien

# Risiken und technische Schulden

# Glossar

| Begriff         | Definition         |
|-----------------|--------------------|
| *\<Begriff-1\>* | *\<Definition-1\>* |
| *\<Begriff-2*   | *\<Definition-2\>* |
