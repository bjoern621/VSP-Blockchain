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
- *User Stories, abgelegt in den [GitHub Issues](https://github.com/users/bjoern621/projects/5)

---

### Motivation

Ziel ist es, Nutzern eine intuitive Möglichkeit zu bieten, digitale Währungen untereinander zu übertragen.  
Das System verbessert die Nachvollziehbarkeit und Transparenz von Transaktionen, reduziert manuelle Fehler und schafft Vertrauen zwischen den Beteiligten.

Aus fachlicher Sicht wird damit die grundlegende Aufgabe der sicheren Kontoführung und Transaktionsverwaltung erfüllt, was den zentralen Mehrwert des Systems darstellt.

---

### Form

| **Use Case / Aufgabe** | **Beschreibung**                                                                                                                                                         |
|-------------------------|--------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| Währung senden | Ein Nutzer kann einem anderen Nutzer einen beliebigen Betrag seiner verfügbaren Währung übertragen. Das System prüft, ob der Sender über ausreichendes Guthaben verfügt. |
| Kontostand anzeigen | Ein Nutzer kann seinen aktuellen Kontostand einsehen.                                                                                                                    |
| Transaktionsverlauf anzeigen | Ein Nutzer kann alle vergangenen Transaktionen seines Kontos einsehen, inklusive gesendeter und empfangener Beträge.                                                     |

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

| Rolle         | Kontakt               | Erwartungshaltung                                                                                                                  |
|---------------|-----------------------|------------------------------------------------------------------------------------------------------------------------------------|
| **Product Owner** | wqx847@haw-hamburg.de | Erwartet, dass das System alle Kernfunktionen wie Überweisungen und Transaktionshistorie zuverlässig bereitstellt. |
| **Endnutzer / Konto-Inhaber** | n/a                   | Erwartet eine sichere, transparente und bedienbare Plattform, um Währung zu senden, zu empfangen und Transaktionen einzusehen.     |
| **Entwicklungsteam** | TBD                   | Erwartet eine klare technische Architektur, testbare Anforderungen und stabile Entwicklungs- und Deployment-Workflows.             |

# Randbedingungen

## Technische Randbedingungen

| **Randbedingung** | **Erläuterung**                                                                                                                                                                          |
|--------------------|------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| **Blockchain-Anbindung** | Das System spiegelt reale Währungswerte über eine bestehende **Blockchain-Infrastruktur** wider. Transaktionen im System müssen mit der entsprechenden Blockchain synchronisiert werden. |
| **Blockchain-Protokoll** | Es wird die **V$Goin-Blockchain** verwendet.                                                                                                                                             |
| **Technologiestack** | Das System wird als Webanwendung auf Basis von **GO** entwickelt.                                                                                                                        |
| **API-Kommunikation** | Alle externen Schnittstellen kommunizieren über **RESTful APIs** mit **JSON** als Austauschformat.                                                                                       |
| **Deployment-Umgebung** | Das System wird in einer **Docker-basierten Cloud-Umgebung** betrieben.                                                                                                                  |
| **Versionierung** | Der Quellcode wird in **GitHub** verwaltet, mit **Git Flow** als Branching-Strategie.                                                                                                    |

---

## Organisatorische Randbedingungen

| **Randbedingung** | **Erläuterung**                                                                                                            |
|--------------------|----------------------------------------------------------------------------------------------------------------------------|
| **Entwicklungsteam** | Das Projekt wird von einem Entwicklungsteam mit 4 Entwicklern mit nur lose definierten Rollen (Dev, QA, DevOps) umgesetzt. |
| **Release-Zyklen** | Neue Releases erfolgen im **3-Wochen-Zyklus**.                                                                             |
| **Dokumentationsstandard** | Architektur und Anforderungen werden nach dem **arc42-Template** gepflegt und versioniert.                                 |
| **Code Review Pflicht** | Jeder Merge in den Hauptbranch erfordert mindestens **eine Freigabe (Code Review)**.                                       |

---

## Rechtliche und sicherheitsrelevante Randbedingungen

| **Randbedingung** | **Erläuterung** |
|--------------------|-----------------|
| **Verschlüsselung** | Sämtliche Datenübertragungen erfolgen ausschließlich über **HTTPS/TLS**. |
| **Backup & Recovery** | Die Transaktionsdaten sind durch die zugrunde liegende Blockchain dezentral und revisionssicher gespeichert. |

---


# Kontextabgrenzung

## Fachlicher Kontext


| **Kommunikationspartner**                      | **Eingabe an das System**                             | **Ausgabe vom System** |
|------------------------------------------------|-------------------------------------------------------|--------------------------|
| **Endnutzer**                   | Signaturen, Transaktionsaufträge, Kontostand-Anfragen | Bestätigungsmeldungen, Transaktionshistorie, aktueller Kontostand |
| **V$Goin Blockchain-System**                   | Transaktionsdaten, Signaturen  | Transaktionsbestätigungen, Block-Hashes, Event-Logs |
| **Monitoring- oder Logging-Systeme**   | Statusabfragen, Metriken                              | Logs, Health-Check-Responses |

## Technischer Kontext
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
| **Kommunikationspartner** | **Technische Schnittstelle / Kanal** | **Protokoll / Datenformat** | **Beschreibung / Bemerkung** |
|----------------------------|-------------------------------------|------------------------------|-------------------------------|
| **Frontend (Web-Client)** | HTTPS REST-API | JSON | Zugriff über Weboberfläche auf Konto- und Transaktionsfunktionen |
| **V$Goin Blockchain-System** | RPC | JSON | Kommunikation mit Blockchain-Knoten zum Senden und Prüfen von Transaktionen |

# Lösungsstrategie
## 2.1 Technologieentscheidungen
- **Programmiersprache:** Go  
  Go wurde gewählt, da es eine hoch performante Sprache ist und eingebaute Nebenläufigkeit (Goroutines) besitzt.

- **Kommunikationsprotokolle:**
    - **gRPC:** Für performante Kommunikation zu Miner-Network.
    - **REST:** Für einfache externe Anbindungen (Frontend).  
      → gRPC bietet hohe Effizienz und Typsicherheit, REST bleibt für Interoperabilität und menschliche Lesbarkeit bestehen.

- **Frameworks / Tools:**
    - `Docker` zur Containerisierung der Nodes

---

## 2.2 Top-Level-Architekturentscheidungen
- **Architekturmuster:**  
  Mehrschichtige Client-Server Architektur mit einem Frontend für die Nutzer, einem Backend, welches die Anfragen der Nutzer bearbeitet und mit einem weiteren Blockchain-Backend, welches die Währung repräsentiert, interagiert.


---

## 2.3 Entscheidungen zur Erreichung der wichtigsten Qualitätsziele

| **Priorität** | **Qualitätsziel** | **Maßnahmen / Architekturentscheidungen** |
|---------------|------------------|--------------------------------------------|
| **1** | **Sicherheit** | Verwendung von Signaturen für Transaktionsauthentizität. Alle gRPC-Verbindungen laufen über TLS. REST-Endpunkte sind authentifiziert und autorisiert. |
| **2** | **Zuverlässigkeit** | Validierung jeder Transaktion durch Mehrheitskonsens der Nodes. Unveränderliche Hash-Ketten sichern Datenintegrität. Fehlerbehandlung und Wiederholungsmechanismen für Netzwerkausfälle. |
| **3** | **Wartbarkeit** | Strikte Modultrennung. |
| **4** | **Performance** | gRPC für effiziente Node-Kommunikation. Nebenläufige Verarbeitung (Goroutines) für Transaktionsvalidierung. |

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
blocks -> blockChain: holt sich Blocks
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

# Qualitätsanforderungen

## Übersicht der Qualitätsanforderungen

## Qualitätsszenarien

# Risiken und technische Schulden

# Glossar

| Begriff         | Definition         |
|-----------------|--------------------|
| *\<Begriff-1\>* | *\<Definition-1\>* |
| *\<Begriff-2*   | *\<Definition-2\>* |
