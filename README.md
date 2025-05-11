# supmap-users

Microservice de gestion des incidents pour Supmap

## Présentation

**supmap-incidents** est un microservice écrit en Go destiné à la gestion des incidents de navigation pour Supmap.

## Architecture

```mermaid
graph TD
  Client[Client] -->|HTTP Request| Router[Router server.go]
  
  subgraph Middleware
  Router -->|CORS| CORSMiddleware[CORS Middleware]
  Router -->|Authentication Check| AuthMiddleware[Auth Middleware]
  Router -->|Admin Check| AdminMiddleware[Admin Middleware]
  end
  
  Middleware -->|Validated Request| Handlers[Handlers handlers.go]
  Handlers -->|DTO Conversion| DTOs[DTOs]
  DTOs -->|Response| Client
  
  subgraph Services["Services Layer"]
  Handlers -.->|Request / Body validation| Validators[Requests Validators]
  Validators -.-> Handlers
  Handlers -->|Business Logic| IncidentService[Incidents Service]
  Handlers -->|Business Logic| InteractionService[Interactions Service]
  end
  
  subgraph AutoModeration["Auto Moderation"]
  Scheduler[Scheduler] -->|Lifetime Check| IncidentService
  end
  
  subgraph Repositories["Repository Layer"]
  direction LR
  IncidentService --> IncidentRepo[Incidents Repository]
  InteractionService --> InteractionRepo[Interactions Repository]
  end
  
  subgraph Models["Domain Models"]
  IncidentRepo --> IncidentModel[Incident Model]
  IncidentRepo --> TypeModel[Type Model]
  InteractionRepo --> InteractionModel[Interaction Model]
  end
  
  subgraph Database
  Models -->|ORM Bun| DB[(PostgreSQL)]
  end
  
  subgraph PubSub["Redis Pub/Sub"]
  IncidentService -->|Publishes Events| Redis[(Redis)]
  end
  
  subgraph Business
  Config[Config.go] -->|Environment Variables| Services
  Config -->|Environment Variables| Repositories
  Config -->|Environment Variables| PubSub
  end
  
  subgraph DTOs["DTO Layer"]
  IncidentDTO[Incident DTO]
  InteractionDTO[Interaction DTO]
  TypeDTO[Type DTO]
  end
```

```
supmap-incidents/
├── cmd/
│   └── api/
│       └── main.go                         # Point d'entrée du microservice
├── internal/
│   ├── api/            
│   │   ├── handlers.go                     # Gestionnaires de requêtes HTTP
│   │   ├── server.go                       # Configuration du serveur HTTP et routes
│   │   ├── middlewares.go                  # Intercepteurs de requête
│   │   └── validations/       
│   │       └── ...                         # Structures de validation
│   ├── config/
│   │   └── config.go                       # Configuration et variables d'environnement
│   ├── models/         
│   │   ├── dto/                            # DTOs permettant d'exposer les données
│   │   └── ...                             # Structures de données pour l'ORM Bun
│   ├── repository/                         # Repository implémentant les requêtes SQL avec l'ORM Bun
│   │   └── ...
│   └── services/                           # Services implémentant les fonctionnalités métier du service
│       ├── ...
│       ├── redis/                        
│       │   ├── redis.go                    # Configuration du client Redis
│       │   └── messages.go                 # Messages envoyés dans le pub/sub
│       └── scheduler/
│           ├── scheduler.go                # Service appelant une fonction à intervalle régulier
│           └── auto-moderate-incidents.go  # Fonctions d'auto modération
├── docs/                                   # Documentation Swagger auto implémentée avec Swggo
│   └── ...
├── Dockerfile                              # Image Docker du microservice
├── go.mod                                  # Dépendances Go
├── go.sum /                                # Checksums des dépendances (auto généré)
└── README.md                               # Documentation du projet
```

## Prérequis et installation

- Go 1.24
- Base de données postgres (conteneurisée ou non)

### Démarrage rapide 

```sh
# Cloner le repo
git clone https://github.com/4PROJ-Le-projet-d-une-vie/supmap-incidents.git
cd supmap-users

# Démarrer le service (nécessite les variables d'environnement, voir ci-dessous)
go run ./cmd/api
```

### Avec Docker

```sh
docker pull ghcr.io/4proj-le-projet-d-une-vie/supmap-incidents:latest
docker run --env-file .env -p 8080:80 supmap-incidents
```

#### Authentification

Pour pull l'image, il faut être authentifié par docker login.

- Générer un Personal Access Token sur GitHub :
    - Se rendre sur https://github.com/settings/tokens
    - Cliquer sur "Generate new token"
    - Cocher au minimum la permission read:packages
    - Copier le token
- Connecter Docker à GHCR avec le token :

```sh
echo 'YOUR_GITHUB_TOKEN' | docker login ghcr.io -u YOUR_GITHUB_USERNAME --password-stdin
```

## Configuration

La configuration se fait via des variables d'environnement ou un fichier `.env` :

|   Variable   | Description                                                                          |
|:------------:|:-------------------------------------------------------------------------------------|
|    `ENV`     | Définit l'environnement dans lequel est exécuté le programme (par défaut production) |
|   `DB_URL`   | URL complète vers la base de donnée                                                  |
|    `PORT`    | Port sur lequel écoutera le service pour recevoir les requêtes                       |
| `JWT_SECRET` | Secret permettant de vérifier l'authenticité d'un token JWT pour l'authentification  |

## Swagger

Chaque handler de ce service comprend des commentaires [Swaggo](https://github.com/swaggo/swag) pour créer dynamiquement une page Swagger-ui.
Exécutez les commande suivantes pour générer la documentation :
```sh
# Installez l'interpréteur de commande Swag
go install github.com/swaggo/swag/cmd/swag@latest

# Générez la documentation
swag init -g cmd/api/main.go
```

Maintenant, vous pouvez accèder à l'URL http://localhost:8080/swagger/index.html décrivant la structure attendue pour chaque endpoint de l'application

> **NB:** La documentation n'inclut pas les endpoints /internal destinés à une utilisation exclusivement interne

## Authentification

L'authentification dans ce service est entièrement externalisée vers le microservice `supmap-users`.

Lorsqu'une requête authentifiée arrive sur le service, le middleware d'authentification :
1. Récupère le token JWT depuis le header `Authorization`
2. Effectue une requête HTTP vers le endpoint interne `/internal/users/check-auth` du service utilisateurs
3. Vérifie la réponse :
  - Si le token est valide, la requête continue son traitement
  - Si le token est invalide ou expiré, une erreur 401 ou 403 est retournée

Cette approche permet de :
- Centraliser la logique d'authentification dans un seul service
- Garantir la cohérence des vérifications de sécurité
- Simplifier la maintenance en évitant la duplication de code

> **Note:** Les routes `/internal` ne sont accessibles que depuis le réseau interne et ne nécessitent pas d'authentification supplémentaire

## Migrations de base de données

Les migrations permettent de versionner la structure de la base de données et de suivre son évolution au fil du temps.
Elles garantissent que tous les environnements (développement, production, etc.) partagent le même schéma de base de données.

Ce projet utilise [Goose](https://github.com/pressly/goose) pour gérer les migrations SQL. Les fichiers de migration sont stockés dans le dossier `migrations/changelog/` et sont embarqués dans le binaire grâce à la directive `//go:embed` dans [migrate.go](migrations/migrate.go).

### Création d'une migration

Pour créer une nouvelle migration, installez d'abord le CLI Goose :

```sh
go install github.com/pressly/goose/v3/cmd/goose@latest
```

Puis créez une nouvelle migration (la commande se base sur les variables du fichier .env) :
```shell
# Crée une migration vide
goose -dir migrations/changelog create nom_de_la_migration sql

# La commande génère un fichier horodaté dans migrations/changelog/
# Example: 20240315143211_nom_de_la_migration.sql
```

### Exécution des migrations

Les migrations sont exécutées automatiquement au démarrage du service via le package migrations :
```go
// Dans main.go
if err := migrations.Migrate("pgx", conf.DbUrl, logger); err != nil {
    logger.Error("migration failed", "err", err)
}
```

Le package migrations utilise embed.FS pour embarquer les fichiers SQL dans le binaire :
```go
//go:embed changelog/*.sql
var changelog embed.FS
// Connexion à la base de données
goose.SetBaseFS(changelog)
```