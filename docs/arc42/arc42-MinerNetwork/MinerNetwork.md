#

**Über arc42**

arc42, das Template zur Dokumentation von Software- und
Systemarchitekturen.

Template Version 9.0-DE. (basiert auf der AsciiDoc Version), Juli 2025

Created, maintained and © by Dr. Peter Hruschka, Dr. Gernot Starke and
contributors. Siehe <https://arc42.org>.

# Einführung und Ziele

Dieses Dokument beschreibt die Architektur des Peer-To-Peer (P2P)-Netzwerk für die Kryptowährung V$Goin. Im Kontext von Kryptowährungen kann dieses Netzwerk als ein öffentliches, dezentrales, Proof-of-Work orientiertes Netz eingeordnet werden. Das heißt, dass jeder teil dieses Netzes sein kann, Transaktionen über mehrere Teilnehmer verteilt gespeichert werden und ein gewissener Rechnenaufwand erforderlich ist, um die Aufgabe eines "Miners" zu erfüllen. Das Netz ist stark an existierenden Blockchains orientiert, wobei Konzepte auf grundlegendes reduziert werden.

Es gibt zwei Hauptakteure im Netzwerk: Miner und Händler. Händler sind nur an der Nutzung des Netzes orientiert. Sie geben hauptsächlich Transaktionen in Auftrag. Miner sind all die Systeme, die zur Erweiterung der Blockchain beitragen. Sie führen bestimmte kryptographische Operationen, die mit Rechenaufwand verbunden sind (Proof-of-Work), aus und ermöglichen so, dass Transaktionen getätigt werden können. Für diese Arbeit werden sie entlohnt. Sowohl Händler als auch Miner können dem Netzwerk jederzeit beitreten und verlassen.

In einem größeren Kontext wird dieses Netzwerk als verteilte Datenbank für den V$Goin genutzt und parallel mit dem System REST-API entwickelt. Die REST-API baut auf dieses Netzwerk auf und soll unseren Kunden einen benutzerfreundlichen Zugang bieten.

## Aufgabenstellung

Das P2P-Netzwerk dient in erster Linie der Ermöglichung von Handel der Kryptowährung V$Goin in einem sicheren und dezentralen Ansatz. Die Grundanforderungen beziehen sich hauptsächlich auf die Erfüllung der Eigenschaften einer Blockchain. Besonders Erzeugung, Verteilung, Validierung und dauerhafte, unveränderliche Speicherung von Transaktionen.

Außerdem entsteht dieses System im Rahmen des Moduls "Verteilte Systeme" im Informatik Studium. Ein wichtiger Teil der Arbeit ist es daher ebenso neue Technologien (Blockchain), Architekturen (der verteilten Systeme) und Programmiersprachen (Go) kennenzulernen.

<div align="center">
    <img src="images/use-cases-network.drawio.svg" alt="Use Case Diagramm mit zentralen Anforderungen"  height="400">
</div>

## Qualitätsziele

| Prioriät | Qualitätsziel                    | Motivation                                                          |
| -------- | -------------------------------- | ------------------------------------------------------------------- |
| 1        | Ease-of-use                      | developer                                                           |
| 2        | Zuverlässigkeit - Fehlertoleranz | Es wird mit Geld gehandelt, ein Fehler kann nicht verkraftet werden |
| 3        | Effizienz - Kapazität            | Ein Ziel von verteilten Systemen ist Skalierbarkeit                 |

Resource Sharing
Openness
Scalability
Distribution Transparency

## Stakeholder

| Rolle      | Erwartungshaltung                                   |
| ---------- | --------------------------------------------------- |
| Entwickler | Lernen der Technologien bei akzeptablem Zeitaufwand |
| Kunde 1    | _\<Erwartung-2\>_                                   |
| Kunde 2    | _\<Erwartung-2\>_                                   |

# Randbedingungen

# Kontextabgrenzung

## Fachlicher Kontext

**\<Diagramm und/oder Tabelle\>**

**\<optional: Erläuterung der externen fachlichen Schnittstellen\>**

## Technischer Kontext

**\<Diagramm oder Tabelle\>**

**\<optional: Erläuterung der externen technischen Schnittstellen\>**

**\<Mapping fachliche auf technische Schnittstellen\>**

# Lösungsstrategie

# Bausteinsicht

## Whitebox Gesamtsystem

**_\<Übersichtsdiagramm\>_**

Begründung  
_\<Erläuternder Text\>_

Enthaltene Bausteine  
_\<Beschreibung der enthaltenen Bausteine (Blackboxen)\>_

Wichtige Schnittstellen  
_\<Beschreibung wichtiger Schnittstellen\>_

### \<Name Blackbox 1\>

_\<Zweck/Verantwortung\>_

_\<Schnittstelle(n)\>_

_\<(Optional) Qualitäts-/Leistungsmerkmale\>_

_\<(Optional) Ablageort/Datei(en)\>_

_\<(Optional) Erfüllte Anforderungen\>_

_\<(optional) Offene Punkte/Probleme/Risiken\>_

### \<Name Blackbox 2\>

_\<Blackbox-Template\>_

### \<Name Blackbox n\>

_\<Blackbox-Template\>_

### \<Name Schnittstelle 1\>

…​

### \<Name Schnittstelle m\>

## Ebene 2

### Whitebox _\<Baustein 1\>_

_\<Whitebox-Template\>_

### Whitebox _\<Baustein 2\>_

_\<Whitebox-Template\>_

…​

### Whitebox _\<Baustein m\>_

_\<Whitebox-Template\>_

## Ebene 3

### Whitebox \<\_Baustein x.1\_\>

_\<Whitebox-Template\>_

### Whitebox \<\_Baustein x.2\_\>

_\<Whitebox-Template\>_

### Whitebox \<\_Baustein y.1\_\>

_\<Whitebox-Template\>_

# Laufzeitsicht

## _\<Bezeichnung Laufzeitszenario 1\>_

-   \<hier Laufzeitdiagramm oder Ablaufbeschreibung einfügen\>

-   \<hier Besonderheiten bei dem Zusammenspiel der Bausteine in diesem
    Szenario erläutern\>

## _\<Bezeichnung Laufzeitszenario 2\>_

…​

## _\<Bezeichnung Laufzeitszenario n\>_

…​

# Verteilungssicht

## Infrastruktur Ebene 1

**_\<Übersichtsdiagramm\>_**

Begründung  
_\<Erläuternder Text\>_

Qualitäts- und/oder Leistungsmerkmale  
_\<Erläuternder Text\>_

Zuordnung von Bausteinen zu Infrastruktur  
_\<Beschreibung der Zuordnung\>_

## Infrastruktur Ebene 2

### _\<Infrastrukturelement 1\>_

_\<Diagramm + Erläuterungen\>_

### _\<Infrastrukturelement 2\>_

_\<Diagramm + Erläuterungen\>_

…​

### _\<Infrastrukturelement n\>_

_\<Diagramm + Erläuterungen\>_

# Querschnittliche Konzepte

## _\<Konzept 1\>_

_\<Erklärung\>_

## _\<Konzept 2\>_

_\<Erklärung\>_

…​

## _\<Konzept n\>_

_\<Erklärung\>_

# Architekturentscheidungen

# Qualitätsanforderungen

## Übersicht der Qualitätsanforderungen

## Qualitätsszenarien

# Risiken und technische Schulden

# Glossar

| Begriff         | Definition         |
| --------------- | ------------------ |
| _\<Begriff-1\>_ | _\<Definition-1\>_ |
| _\<Begriff-2_   | _\<Definition-2\>_ |
