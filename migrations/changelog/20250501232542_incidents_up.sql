-- +goose Up
-- +goose StatementBegin
CREATE TABLE incident_types
(
    id                            SERIAL PRIMARY KEY,
    name                          VARCHAR(100) NOT NULL,
    description                   TEXT         NOT NULL,
    lifetime_without_confirmation INTEGER      NOT NULL, -- en secondes
    negative_reports_threshold    INTEGER      NOT NULL,
    global_lifetime               INTEGER      NOT NULL, -- en secondes
    need_recalculation            BOOLEAN      NOT NULL
);

INSERT INTO incident_types (name, description, lifetime_without_confirmation, negative_reports_threshold, global_lifetime, need_recalculation) VALUES
('Accident', 'Collision ou incident impliquant un ou plusieurs véhicules.', 3600, 10, 21600, false),
('Route Fermée', 'Route temporairement inaccessible à la circulation.', 86400, 5, 172800, true),
('Embouteillage', 'Circulation fortement ralentie ou à l’arrêt.', 1800, 5, 7200, true),
('Contrôle de Police', 'Contrôle routier en cours par les forces de l’ordre.', 57600, 10, 172800, false),
('Obstacles', 'Objets ou éléments gênant la circulation sur la chaussée.', 3600, 10, 21600, false);

CREATE TABLE incidents
(
    id         SERIAL PRIMARY KEY,
    type_id    INTEGER        NOT NULL REFERENCES incident_types (id),
    user_id    INTEGER        NOT NULL,
    latitude   DECIMAL(10, 8) NOT NULL,
    longitude  DECIMAL(11, 8) NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP WITH TIME ZONE,
    CONSTRAINT lat_range CHECK (latitude BETWEEN -90 AND 90),
    CONSTRAINT lon_range CHECK (longitude BETWEEN -180 AND 180)
);

CREATE TABLE incident_interactions
(
    id               SERIAL PRIMARY KEY,
    incident_id      INTEGER NOT NULL REFERENCES incidents (id),
    user_id          INTEGER NOT NULL,
    is_still_present BOOLEAN NOT NULL,
    created_at       TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Index pour les recherches géographiques
CREATE INDEX incidents_lat_long_idx ON incidents (latitude, longitude)
    WHERE deleted_at IS NULL;

-- Index pour la gestion des durées de vie
CREATE INDEX incidents_lifetime_idx ON incidents (created_at, type_id)
    WHERE deleted_at IS NULL;

-- Index pour les recherches d'historique par utilisateur
CREATE INDEX incidents_user_idx ON incidents (user_id);
CREATE INDEX interactions_user_idx ON incident_interactions (user_id);

-- Index pour compter rapidement les interactions négatives
CREATE INDEX incident_interactions_negative_idx
    ON incident_interactions (incident_id)
    WHERE NOT is_still_present;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX IF EXISTS incident_interactions_negative_idx;
DROP INDEX IF EXISTS interactions_user_idx;
DROP INDEX IF EXISTS incidents_user_idx;
DROP INDEX IF EXISTS incidents_lifetime_idx;
DROP INDEX IF EXISTS incidents_lat_long_idx;

DROP TABLE IF EXISTS incident_interactions;
DROP TABLE IF EXISTS incidents;
DROP TABLE IF EXISTS incident_types;
-- +goose StatementEnd
