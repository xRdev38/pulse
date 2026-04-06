# Guide Windows — Pulse

## Prérequis

### 1. Docker Desktop
Télécharge et installe Docker Desktop :
https://www.docker.com/products/docker-desktop/

Après installation, vérifie dans PowerShell :
```powershell
docker --version
docker compose version
```

### 2. Go (pour les services Go)
https://go.dev/dl/ → Télécharge l'installeur Windows (.msi)

```powershell
go version   # doit afficher go1.22+
```

### 3. Node.js (pour Gateway et Notifier)
https://nodejs.org/ → LTS version

```powershell
node --version   # doit afficher v20+
npm --version
```

### 4. Optionnel — grpcurl (pour les tests gRPC)
Via Scoop (gestionnaire de paquets Windows) :
```powershell
# Installer Scoop si pas encore fait :
Set-ExecutionPolicy RemoteSigned -Scope CurrentUser
irm get.scoop.sh | iex

# Installer grpcurl :
scoop install grpcurl
```

Ou télécharge le binaire directement :
https://github.com/fullstorydev/grpcurl/releases

---

## Première utilisation

### Étape 1 — Autoriser le script PowerShell

Par défaut, Windows bloque les scripts `.ps1`. Ouvre PowerShell **en tant qu'administrateur** et exécute :

```powershell
Set-ExecutionPolicy -ExecutionPolicy RemoteSigned -Scope CurrentUser
```

### Étape 2 — Se placer dans le dossier Pulse

```powershell
cd C:\chemin\vers\pulse
```

### Étape 3 — Démarrer les services

```powershell
.\pulse.ps1 up
```

Attendre ~30 secondes que tous les services démarrent.

### Étape 4 — Vérifier que tout fonctionne

```powershell
.\pulse.ps1 health
```

Tu dois voir `"status": "healthy"` pour le Collector et le Gateway.

### Étape 5 — Tester l'ingestion

```powershell
.\pulse.ps1 ingest-test
```

Résultat attendu : `"status": "accepted"`

---

## Commandes courantes

```powershell
.\pulse.ps1 up                  # démarrer
.\pulse.ps1 down                # arrêter
.\pulse.ps1 health              # vérifier l'état
.\pulse.ps1 ingest-test         # tester l'ingestion
.\pulse.ps1 test-alert-flow     # tester le flux d'alerte complet
.\pulse.ps1 ui grafana          # ouvrir Grafana dans le navigateur
.\pulse.ps1 ui jaeger           # ouvrir Jaeger
.\pulse.ps1 help                # voir toutes les commandes
```

---

## Dépannage Windows

### "Le terme pulse.ps1 n'est pas reconnu"
```powershell
# Assure-toi d'être dans le bon dossier :
cd C:\chemin\vers\pulse
.\pulse.ps1 help
```

### "Accès refusé" sur le script
```powershell
Set-ExecutionPolicy -ExecutionPolicy RemoteSigned -Scope CurrentUser
```

### Docker ne démarre pas
- Vérifie que Docker Desktop est bien lancé (icône dans la barre des tâches)
- Sur Windows 10/11 : activer WSL2 si demandé par Docker Desktop
- Redémarre Docker Desktop si nécessaire

### Port déjà utilisé (ex: port 5432)
```powershell
# Voir quel process utilise le port 5432 :
netstat -ano | findstr :5432

# Arrêter le process (remplace <PID> par le numéro) :
taskkill /PID <PID> /F
```

### "Invoke-RestMethod: La connexion a été refusée"
Les services ne sont pas encore prêts. Attends 30s après `.\pulse.ps1 up` et relance `.\pulse.ps1 health`.

### Volumes corrompus
```powershell
.\pulse.ps1 down        # supprime les volumes
.\pulse.ps1 up          # repart de zéro
```

---

## Interfaces web disponibles

| Interface | URL | Credentials |
|---|---|---|
| Grafana (dashboards) | http://localhost:3001 | admin / admin |
| Jaeger (traces) | http://localhost:16686 | — |
| RabbitMQ | http://localhost:15672 | pulse / pulse_secret |
| Prometheus | http://localhost:9091 | — |
| MailHog (emails de test) | http://localhost:8025 | — |

```powershell
.\pulse.ps1 ui grafana     # ouvre automatiquement dans le navigateur
.\pulse.ps1 ui jaeger
.\pulse.ps1 ui rabbitmq
.\pulse.ps1 ui mailhog
```