CREATE TABLE IF NOT EXISTS source (
    name VARCHAR(128) NOT NULL PRIMARY KEY,
    index int NOT NULL
);

CREATE TABLE IF NOT EXISTS node (
    node_id UUID NOT NULL PRIMARY KEY,
    geography VARCHAR(128) NOT NULL,
    datacenter VARCHAR(128) NOT NULL,
    address VARCHAR(128) NOT NULL,
    metadata JSONB
);

CREATE TABLE IF NOT EXISTS endpoint (
    endpoint_id UUID NOT NULL  PRIMARY KEY,
    node_id UUID NOT NULL,
    service_name VARCHAR(128) NOT NULL,
    config JSONB
);

CREATE INDEX unq_node_service_name ON endpoint(node_id, service_name);
CREATE INDEX idx_endpoint_service_name ON endpoint(service_name);

CREATE TABLE IF NOT EXISTS service (
    service_name VARCHAR(128) NOT NULL PRIMARY KEY,
    config JSONB
);

INSERT INTO source VALUES ('endpoints', 0);
INSERT INTO source VALUES ('routing', 0);