# 🏢 Ymmo — Plateforme Immobilière & Infrastructure Sécurisée

## 📌 Description

**Ymmo** est une plateforme web immobilière développée dans le cadre d’un projet fil rouge.
Elle permet de centraliser la gestion des biens immobiliers, faciliter les interactions entre clients et agents, et exploiter les données pour améliorer la prise de décision.

Le projet inclut également la mise en place d’une **infrastructure réseau sécurisée multi-sites**, reliant un siège social et plusieurs agences.

---

## 🎯 Objectifs

* Digitaliser la gestion des biens immobiliers
* Centraliser les données entre le siège et les agences
* Sécuriser les échanges via une architecture réseau adaptée
* Offrir une interface moderne et accessible
* Exploiter les données pour analyser le marché immobilier

---

## ⚙️ Stack technique

### 💻 Développement

* **Backend** : Go (Gin)
* **Frontend** : HTML / TailwindCSS
* **Base de données** : MySQL

### 🌐 Infrastructure

* Windows Server (Active Directory)
* DNS / DHCP
* VPN (IPSec)
* Virtualisation (VirtualBox / Hyper-V)

### 📊 Data & Analyse

* Python (analyse de données, statistiques)

---

## 👥 Utilisateurs

* **Clients** : consultation des biens, prise de contact
* **Agents immobiliers** : gestion des annonces
* **Administrateurs** : supervision globale
* **IT** : gestion des accès et de l’infrastructure

---

## 🚀 Fonctionnalités principales

### 🔐 Authentification

* Inscription / Connexion
* Gestion des rôles (client, agent, admin)

### 🏠 Gestion des biens

* Création, modification, suppression
* Ajout d’images
* Consultation des annonces

### 🔎 Recherche

* Filtrage (prix, localisation, type)
* Tri des résultats

### 📊 Analyse

* Statistiques sur les biens
* Analyse des tendances

### 🌐 Infrastructure

* Réseau multi-sites
* VPN sécurisé
* Active Directory
* Gestion des droits d’accès

---

## 🗂️ Structure du projet (exemple)

```
/cmd            # Point d'entrée de l'application
/internal       # Logique métier (services, controllers)
/pkg            # Packages réutilisables
/web            # Frontend (HTML, CSS)
/scripts        # Scripts (DB, data, etc.)
/docs           # Documentation
```

---

## 🛠️ Installation

### 1. Cloner le projet

```
git clone https://github.com/ton-repo/ymmo.git
cd ymmo
```

### 2. Configurer la base de données

* Installer MySQL
* Créer une base de données
* Importer le schéma SQL

### 3. Lancer le backend

```
go mod tidy
go run cmd/main.go
```

### 4. Accéder à l’application

```
http://localhost:8080
```

---

## 🧪 Tests

* Tests API (Postman / Insomnia)
* Tests fonctionnels (navigation, formulaires)
* Tests infrastructure (VM, réseau, AD)

---

## 🔐 Sécurité

* Hash des mots de passe (bcrypt)
* Authentification sécurisée (JWT)
* Gestion des rôles et permissions
* Protection contre injections SQL

---

## 📚 Documentation

* Documentation fonctionnelle
* Documentation technique
* Schéma réseau
* Guide d’installation

---

## 🎤 Démonstration

Le projet inclut :

* Une application web fonctionnelle
* Une infrastructure simulée via machines virtuelles
* Une démonstration complète (DEV + INFRA)

---

## 📈 Améliorations possibles

* Dashboard administrateur
* Système de favoris
* Messagerie interne
* Déploiement cloud (AWS / Azure)

---

## 👨‍💻 Auteurs

Projet réalisé dans le cadre du Bachelor 2 — Ynov Informatique

---

## 📄 Licence

Projet pédagogique — usage académique uniquement
